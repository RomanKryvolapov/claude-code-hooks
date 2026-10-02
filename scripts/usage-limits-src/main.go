// usage-limits — Claude Code limits awareness (Go port of an earlier Node hook).
//
// Compiled to a static, dependency-free binary per platform (darwin/arm64,
// linux/amd64, windows/amd64), committed under scripts/ and dispatched by the
// `.claude/hooks/usage-limits` sh launcher, which execs the binary matching the
// current OS. No Node runtime required — this is the whole point of the port.
//
// It injects subscription-window usage into the model's context, powers the
// status line, gates subagent spawns, and exposes JSON for planning.
// Model-aware: the zone accounts for the per-model weekly bucket of the ACTIVE
// model (and, in gate mode, of the model being spawned).
//
// Data source: https://api.anthropic.com/api/oauth/usage (account-global,
// covers ALL sessions on this subscription), authenticated with the local
// OAuth token that `claude login` stored: .credentials.json in the config dir
// (~/.claude, or CLAUDE_CONFIG_DIR) on Linux and Windows, the login Keychain on
// macOS. See readAccessToken.
//
// The status line draws a picture for a person. Three lines: CONTEXT (how full this session's
// context window is, one gauge cell per percent), then the limits side by side (5-hour window,
// 7-day window, then each per-model weekly bucket) as LIMIT (budget spent) over the share of the
// window gone — comparing the two says whether the spend is ahead of the clock. That second gauge
// is calendar time on the 5-hour row (captioned TIME) and working time on the weekly rows
// (captioned WORK), weighted per weekday; see workingWeek. It is coloured, fitted to the
// terminal, and carries the prompt cache and token counts beside CONTEXT; see statusline.go.
//
// The model is handed the same figures as data — XML, with no gauge; see modelblock.go — plus the
// token counts and the elements it acts on: burn, zone, and the model at session start.
//
// Modes (CLI flags, case-insensitive; PowerShell-style -Mode also accepted):
//   --mode inject --event SessionStart      <usage_limits> block, with <model> -> model context
//   --mode inject --event UserPromptSubmit  <usage_limits> block                -> model context
//   --mode gate                             PreToolUse Agent|Task|Workflow gate
//   --mode json [--model <id>]              full computed state as JSON
//   --mode statusline                       the picture, painted, with cache and token counts
//
// Never breaks the session: on any error it exits 0 — the injection and gate
// modes stay silent, statusline and JSON print a short unavailable marker.

package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	// The working week is measured in a named zone (see defaultTimeZone), and Windows ships no
	// zone database at all while a Linux container often ships none either. Embedding it costs
	// about 400 KB per binary and is what makes the named zone resolve the same everywhere.
	_ "time/tzdata"
)

// --- arg parsing (accepts --key value and PowerShell -Key value) --------------

type argSet struct {
	vals    map[string]string
	present map[string]bool
}

func parseArgs(argv []string) argSet {
	out := argSet{vals: map[string]string{}, present: map[string]bool{}}
	for i := 0; i < len(argv); i++ {
		a := argv[i]
		if !strings.HasPrefix(a, "-") {
			continue
		}
		key := strings.ToLower(strings.TrimLeft(a, "-"))
		out.present[key] = true
		if i+1 < len(argv) && !strings.HasPrefix(argv[i+1], "-") {
			out.vals[key] = argv[i+1]
			i++
		} else {
			out.vals[key] = "true"
		}
	}
	return out
}

func (a argSet) str(key, def string) string {
	if v, ok := a.vals[key]; ok && v != "true" {
		return v
	}
	if _, ok := a.present[key]; ok {
		if a.vals[key] != "" {
			return a.vals[key]
		}
	}
	return def
}

func (a argSet) numOr(key string, def float64) float64 {
	if !a.present[key] {
		return def
	}
	f, err := strconv.ParseFloat(a.vals[key], 64)
	if err != nil {
		return def
	}
	return f
}

// --- globals (populated in main) ----------------------------------------------

// Paths, resolved in loadConfig. Everything this hook can be tuned by lives in constants.go.
var (
	claudeDir      string
	cachePath      string
	gateAsksPath   string
	tokenCachePath string
	credPath       string
)

func homeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return h
}

// configDir is where Claude Code keeps its user-level files: CLAUDE_CONFIG_DIR when set, ~/.claude
// otherwise. The login, the user settings the model pin is read from and this hook's own cache all
// follow it, so a session started with another config dir reads its own account.
func configDir() string {
	if d := os.Getenv("CLAUDE_CONFIG_DIR"); d != "" {
		return d
	}
	return filepath.Join(homeDir(), ".claude")
}

// --- data structures ----------------------------------------------------------

type usageWindow struct {
	Utilization float64 `json:"utilization"`
	ResetsAt    string  `json:"resets_at"`
}

type usageLimit struct {
	Kind     string  `json:"kind"`
	Percent  float64 `json:"percent"`
	ResetsAt string  `json:"resets_at"`
	Scope    *struct {
		Model *struct {
			DisplayName string `json:"display_name"`
		} `json:"model"`
	} `json:"scope"`
}

type usageRaw struct {
	FiveHour *usageWindow `json:"five_hour"`
	SevenDay *usageWindow `json:"seven_day"`
	Limits   []usageLimit `json:"limits"`
}

type sample struct {
	T  string             `json:"t"`
	S5 float64            `json:"s5"`
	S7 float64            `json:"s7"`
	R5 string             `json:"r5"`
	R7 string             `json:"r7"`
	Bk map[string]float64 `json:"bk"`
}

// cacheFile holds only what the usage API said and what was measured from it. A figure Claude Code
// handed the status line is never written here: the gate and the injected block read this file
// too, and they cannot tell where a number came from.
type cacheFile struct {
	FetchedAt string          `json:"fetched_at"`
	Raw       json.RawMessage `json:"raw"`
	Samples   []sample        `json:"samples"`
	// TriedAt is the last time the API was asked, answered or not. It is asked at most once per
	// fetch period either way: an endpoint that has started refusing — it rate-limits — is not
	// asked again by every status-line refresh of every open session.
	TriedAt string `json:"tried_at,omitempty"`
}

type scopedBucket struct {
	Model    string  `json:"model"`
	Percent  float64 `json:"percent"`
	ResetsAt string  `json:"resets_at,omitempty"`
}

// --- helpers ------------------------------------------------------------------

func readStdin() string {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return ""
	}
	if (fi.Mode() & os.ModeCharDevice) != 0 {
		return "" // TTY, no piped input
	}
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return ""
	}
	return string(b)
}

// stripBOM removes a UTF-8 byte-order mark. The JSON this hook reads whole — its cache, its config,
// the credentials, the settings files it takes the model pin from — sits under ~/.claude or in a
// project and gets opened in editors; one saved with a BOM by a Windows editor would make
// encoding/json fail on the first byte, silently. Hence the four whole-file read sites below; the
// machine-written JSONL transcripts are parsed line by line, never hand-edited, and stay out.
func stripBOM(b []byte) []byte {
	return bytes.TrimPrefix(b, []byte{0xEF, 0xBB, 0xBF})
}

func getCacheObject() *cacheFile {
	b, err := os.ReadFile(cachePath)
	if err != nil {
		return nil
	}
	var c cacheFile
	if json.Unmarshal(stripBOM(b), &c) != nil {
		return nil
	}
	return &c
}

// storedLogin is the part of Claude Code's saved login this hook uses.
type storedLogin struct {
	AccessToken string `json:"accessToken"`
	ExpiresAt   int64  `json:"expiresAt"` // epoch ms
}

// parseLogin reads the claudeAiOauth block of a credentials document, whichever store it came from.
func parseLogin(b []byte) (storedLogin, bool) {
	var doc struct {
		ClaudeAiOauth storedLogin `json:"claudeAiOauth"`
	}
	if json.Unmarshal(stripBOM(b), &doc) != nil || doc.ClaudeAiOauth.AccessToken == "" {
		return storedLogin{}, false
	}
	return doc.ClaudeAiOauth, true
}

// readAccessToken returns the OAuth token `claude login` stored. On Linux and Windows that is
// .credentials.json in the config dir. On macOS Claude Code keeps it in the login Keychain and writes
// the file only when a Keychain write fails (an SSH session, a locked keychain), so there the file can
// be a stale leftover: both stores are read and the login that expires last wins.
func readAccessToken() string {
	var logins []storedLogin
	if runtime.GOOS == "darwin" {
		if l, ok := parseLogin(keychainLogin()); ok {
			logins = append(logins, l)
		}
	}
	if b, err := os.ReadFile(credPath); err == nil {
		if l, ok := parseLogin(b); ok {
			logins = append(logins, l)
		}
	}
	return freshestToken(logins)
}

// freshestToken picks the login that expires last; on a tie the store read first, the Keychain, wins.
func freshestToken(logins []storedLogin) string {
	best := -1
	for i, l := range logins {
		if best < 0 || l.ExpiresAt > logins[best].ExpiresAt {
			best = i
		}
	}
	if best < 0 {
		return ""
	}
	return logins[best].AccessToken
}

// keychainService is the generic-password item Claude Code writes on macOS: "Claude Code-credentials",
// or, when CLAUDE_CONFIG_DIR is set, that name plus "-" and the first 8 hex digits of the SHA-256 of
// the variable's value exactly as exported. The suffix is observed behaviour, not documented — the docs
// only say the item is keyed to the directory. There is deliberately no fallback to the unsuffixed
// item: it holds whichever account the default config dir is signed in to.
func keychainService() string {
	const base = "Claude Code-credentials"
	d := os.Getenv("CLAUDE_CONFIG_DIR")
	if d == "" {
		return base
	}
	sum := sha256.Sum256([]byte(d))
	return base + "-" + hex.EncodeToString(sum[:])[:8]
}

// keychainLogin reads that item through the system `security` tool, which needs no cgo. Bounded by a
// timeout, because a locked keychain must not hold a hook; any failure — no item, no GUI session over
// SSH — returns nil and the file is tried on its own.
func keychainLogin() []byte {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "/usr/bin/security", "find-generic-password", "-s", keychainService(), "-w").Output()
	if err != nil {
		return nil
	}
	return bytes.TrimSpace(out)
}

func fetchUsage() json.RawMessage {
	token := readAccessToken()
	if token == "" {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.anthropic.com/api/oauth/usage", nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("anthropic-beta", "oauth-2025-04-20")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil
	}
	if !json.Valid(body) {
		return nil
	}
	return json.RawMessage(body)
}

func cleanModelString(m string) string {
	if m == "" {
		return ""
	}
	if i := strings.Index(m, "["); i >= 0 {
		if j := strings.Index(m[i:], "]"); j >= 0 {
			m = m[:i] + m[i+j+1:]
		}
	}
	return strings.ToLower(strings.TrimSpace(m))
}

func getActiveModel(args argSet, hookInput map[string]any) string {
	// 1) transcript: model of the last assistant message
	if hookInput != nil {
		if tp, ok := hookInput["transcript_path"].(string); ok && tp != "" {
			if f, err := os.Open(tp); err == nil {
				var lines []string
				sc := bufio.NewScanner(f)
				sc.Buffer(make([]byte, 1024*1024), 8*1024*1024)
				for sc.Scan() {
					if t := sc.Text(); t != "" {
						lines = append(lines, t)
					}
				}
				f.Close()
				start := 0
				if len(lines) > 60 {
					start = len(lines) - 60
				}
				lines = lines[start:]
				for i := len(lines) - 1; i >= 0; i-- {
					var o map[string]any
					if json.Unmarshal([]byte(lines[i]), &o) != nil {
						continue
					}
					if msg, ok := o["message"].(map[string]any); ok {
						if mdl, ok := msg["model"].(string); ok && mdl != "" {
							return cleanModelString(mdl)
						}
					}
				}
			}
		}
	}
	// 2) explicit parameter
	if m := args.str("model", ""); m != "" {
		return cleanModelString(m)
	}
	// 3) env
	if m := os.Getenv("ANTHROPIC_MODEL"); m != "" {
		return cleanModelString(m)
	}
	// 4) settings pins: project, then user
	var candidates []string
	if pd := os.Getenv("CLAUDE_PROJECT_DIR"); pd != "" {
		candidates = append(candidates, filepath.Join(pd, ".claude", "settings.json"))
	}
	candidates = append(candidates, filepath.Join(claudeDir, "settings.json"))
	for _, sp := range candidates {
		b, err := os.ReadFile(sp)
		if err != nil {
			continue
		}
		var o map[string]any
		if json.Unmarshal(stripBOM(b), &o) != nil {
			continue
		}
		if m, ok := o["model"].(string); ok && m != "" {
			return cleanModelString(m)
		}
	}
	return ""
}

func findBucket(scoped []scopedBucket, modelID string) *scopedBucket {
	if modelID == "" {
		return nil
	}
	for i := range scoped {
		name := strings.ToLower(scoped[i].Model)
		if name != "" && strings.Contains(modelID, name) {
			return &scoped[i]
		}
	}
	return nil
}

func zoneOf(v float64, t map[string]float64) string {
	if v >= t["RED"] {
		return "RED"
	}
	if v >= t["ORANGE"] {
		return "ORANGE"
	}
	if v >= t["YELLOW"] {
		return "YELLOW"
	}
	return "GREEN"
}

func pad2(n int) string { return fmt.Sprintf("%02d", n) }
func fmtHHmm(t time.Time) string {
	t = t.Local()
	return pad2(t.Hour()) + ":" + pad2(t.Minute())
}
func fmtDddHHmm(t time.Time) string {
	t = t.Local()
	return weekdays[int(t.Weekday())] + " " + pad2(t.Hour()) + ":" + pad2(t.Minute())
}
func fmtYmdHM(t time.Time) string {
	t = t.Local()
	return fmt.Sprintf("%d-%s-%s %s:%s", t.Year(), pad2(int(t.Month())), pad2(t.Day()), pad2(t.Hour()), pad2(t.Minute()))
}

func round(v float64, digits int) float64 {
	f := math.Pow(10, float64(digits))
	return math.Round(v*f) / f
}

// num formats a rounded float the way JS number->string does (no trailing zeros).
func num(v float64, digits int) string {
	return strconv.FormatFloat(round(v, digits), 'f', -1, 64)
}

func parseTime(s string) (time.Time, bool) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t, err = time.Parse(time.RFC3339Nano, s)
		if err != nil {
			return time.Time{}, false
		}
	}
	return t, true
}

// --- state --------------------------------------------------------------------

type state struct {
	s5, s7                             float64
	r5Local, r7Local                   time.Time
	burn                               float64
	burnSource                         string
	paceRatio                          float64
	minToExhaust, minToReset5          float64
	minToReset7, horizonMin            float64
	zone, zone5, zone7, zoneBucket     string
	limiter                            string
	scoped                             []scopedBucket
	activeModel                        string
	activeBucket                       *scopedBucket
	bucketPct                          float64
	bucketBurn                         float64
	bucketBurnSource                   string
	minToBucketExhaust, minToBucketRst float64
	bucketResetLocal                   *time.Time
	sampleCount                        int
	// fetchedAt is when the usage API gave the figures this state is built on; zero when it never
	// has. stale says it has not answered for longer than two fetch periods, so those figures may
	// no longer hold — except where Claude Code's own figures, which only the status line is handed,
	// stood in for them.
	fetchedAt time.Time
	stale     bool
}

// getState builds everything the outputs draw from the usage API's figures. live is the 5-hour and
// 7-day figures Claude Code handed the status line, or nil everywhere else; they stand in for the
// API's only when the API's are stale or missing, and are never written to the cache.
func getState(args argSet, activeModelID string, live *usageRaw) *state {
	overrideS5 := args.numOr("overrides5", -1)
	overrideS7 := args.numOr("overrides7", -1)
	overrideBucket := args.numOr("overridebucket", -1)

	now := time.Now()
	nowMs := float64(now.UnixMilli())
	cache := getCacheObject()
	var rawBytes json.RawMessage
	var fetchedAt time.Time
	fetchedFresh := false

	if cache != nil && cache.FetchedAt != "" {
		if ft, ok := parseTime(cache.FetchedAt); ok {
			age := (nowMs - float64(ft.UnixMilli())) / 1000
			if age >= 0 && age < float64(fetchTTLSec) {
				rawBytes = cache.Raw
				fetchedAt = ft
			}
		}
	}
	triedAt := ""
	if cache != nil {
		triedAt = cache.TriedAt
	}
	if rawBytes == nil {
		askedRecently := false
		if t, ok := parseTime(triedAt); ok {
			since := now.Sub(t)
			askedRecently = since >= 0 && since < time.Duration(fetchTTLSec)*time.Second
		}
		if !askedRecently {
			rawBytes = fetchUsage()
			triedAt = now.Format(time.RFC3339Nano)
		}
		if rawBytes != nil {
			fetchedFresh = true
			fetchedAt = now
		} else if cache != nil {
			rawBytes = cache.Raw // network down -> stale cache beats nothing
			fetchedAt, _ = parseTime(cache.FetchedAt)
		}
	}
	// Stale means the API did not answer this time and what it said last is older than two fetch
	// periods — one failed refresh is a blip, two in a row are an outage. The injected block tells
	// the model so; the status line swaps in Claude Code's own figures where it has them.
	stale := !fetchedFresh && (fetchedAt.IsZero() || now.Sub(fetchedAt) > 2*time.Duration(fetchTTLSec)*time.Second)

	var raw usageRaw
	if rawBytes != nil {
		_ = json.Unmarshal(rawBytes, &raw)
	}
	// Claude Code's own figure stands in for a window the API's figures cannot vouch for: all of
	// them while they are stale, and one the API left out even when it answered.
	liveFive, liveSeven := false, false
	if live != nil {
		if live.FiveHour != nil && (stale || raw.FiveHour == nil) {
			raw.FiveHour, liveFive = live.FiveHour, true
		}
		if live.SevenDay != nil && (stale || raw.SevenDay == nil) {
			raw.SevenDay, liveSeven = live.SevenDay, true
		}
	}
	// What the API answered on this call, if it did: when there is nothing to draw from, it is still
	// what the cache must hold from now on.
	var fresh json.RawMessage
	if fetchedFresh {
		fresh = rawBytes
	}
	if raw.FiveHour == nil {
		// Nothing to draw from, but the attempt is still written down, so that the API is not
		// asked again before its time.
		rememberTheAttempt(cache, triedAt, fresh)
		return nil
	}

	s5 := raw.FiveHour.Utilization
	var s7 float64
	if raw.SevenDay != nil {
		s7 = raw.SevenDay.Utilization
	}
	if overrideS5 >= 0 {
		s5 = overrideS5
	}
	if overrideS7 >= 0 {
		s7 = overrideS7
	}
	r5, ok5 := parseTime(raw.FiveHour.ResetsAt)
	if !ok5 {
		// A 5-hour window that has not started carries no reset time: nothing to draw from either,
		// and the attempt is written down the same way. Given up before it, every refresh of every
		// open session asked the API again (review of PR 697).
		rememberTheAttempt(cache, triedAt, fresh)
		return nil
	}
	var r7 time.Time
	if raw.SevenDay != nil {
		r7, _ = parseTime(raw.SevenDay.ResetsAt)
	}
	r5Ms := float64(r5.UnixMilli())

	// --- per-model scoped weekly buckets (dynamic, from limits[]) ---------------
	var scoped []scopedBucket
	bkMap := map[string]float64{}
	for _, lim := range raw.Limits {
		if lim.Kind == "weekly_scoped" && lim.Scope != nil && lim.Scope.Model != nil {
			e := scopedBucket{Model: lim.Scope.Model.DisplayName, Percent: lim.Percent}
			if lim.ResetsAt != "" {
				e.ResetsAt = lim.ResetsAt
			}
			scoped = append(scoped, e)
			bkMap[e.Model] = e.Percent
		}
	}

	// --- burn-rate samples (kept per 5h window, pruned on window change) --------
	var srcSamples []sample
	if cache != nil {
		srcSamples = cache.Samples
	}
	var kept []sample
	for _, smp := range srcSamples {
		sr, ok := parseTime(smp.R5)
		if !ok {
			continue
		}
		st, ok := parseTime(smp.T)
		if !ok {
			continue
		}
		sameWindow := math.Abs((float64(sr.UnixMilli())-r5Ms)/1000) < 120
		recent := (nowMs-float64(st.UnixMilli()))/60000 < sampleMaxAgeMin
		notAfterReset := smp.S5 <= s5+1.0
		if sameWindow && recent && notAfterReset {
			kept = append(kept, smp)
		}
	}
	if fetchedFresh {
		kept = append(kept, sample{
			T: now.Format(time.RFC3339Nano), S5: s5, S7: s7,
			R5: r5.Format(time.RFC3339Nano), R7: r7.Format(time.RFC3339Nano), Bk: bkMap,
		})
	}

	// --- burn %/min: sampled if we have a >=3 min span, else pace-based ---------
	winStartMs := r5Ms - sessionWindowMin*60*1000
	elapsedMin := (nowMs - winStartMs) / 60000
	if elapsedMin < 1 {
		elapsedMin = 1
	}
	burn := s5 / elapsedMin
	burnSource := "pace"
	if len(kept) >= 2 {
		first := kept[0]
		last := kept[len(kept)-1]
		ft, _ := parseTime(first.T)
		lt, _ := parseTime(last.T)
		spanMin := float64(lt.UnixMilli()-ft.UnixMilli()) / 60000
		if spanMin >= 3 {
			sampled := (last.S5 - first.S5) / spanMin
			if sampled > 0 {
				burn = sampled
				burnSource = "sampled"
			} else {
				burn = 0
				burnSource = "idle"
			}
		}
	}

	// --- forecast ---------------------------------------------------------------
	minToReset5 := (r5Ms - nowMs) / 60000
	minToReset7 := math.Inf(1)
	if !r7.IsZero() {
		minToReset7 = (float64(r7.UnixMilli()) - nowMs) / 60000
	}
	minToExhaust := math.Inf(1)
	if burn > 0.001 {
		minToExhaust = (100 - s5) / burn
	}
	horizonMin := math.Min(minToExhaust, minToReset5)
	paceRatio := 0.0
	elapsedFrac := (elapsedMin / sessionWindowMin) * 100
	if elapsedFrac > 0.5 {
		paceRatio = s5 / elapsedFrac
	}

	// --- active model + its bucket ----------------------------------------------
	activeBucket := findBucket(scoped, activeModelID)
	bucketPct := -1.0
	if activeBucket != nil {
		bucketPct = activeBucket.Percent
	}
	if overrideBucket >= 0 {
		bucketPct = overrideBucket
		if activeBucket == nil {
			activeBucket = &scopedBucket{Model: "test", Percent: overrideBucket}
		}
	}

	// --- bucket burn + exhaustion forecast for the active model's bucket --------
	bucketBurn := 0.0
	bucketBurnSource := "idle"
	minToBucketExhaust := math.Inf(1)
	minToBucketReset := math.Inf(1)
	var bucketResetLocal *time.Time
	if activeBucket != nil && bucketPct >= 0 {
		if activeBucket.ResetsAt != "" {
			if br, ok := parseTime(activeBucket.ResetsAt); ok {
				bucketResetLocal = &br
				minToBucketReset = (float64(br.UnixMilli()) - nowMs) / 60000
			}
		}
		type pt struct{ t, v float64 }
		var pts []pt
		for _, smp := range kept {
			if smp.Bk == nil {
				continue
			}
			v, ok := smp.Bk[activeBucket.Model]
			if !ok {
				continue
			}
			if v <= bucketPct+1.0 {
				st, _ := parseTime(smp.T)
				pts = append(pts, pt{float64(st.UnixMilli()), v})
			}
		}
		if len(pts) >= 2 {
			spanMin := (pts[len(pts)-1].t - pts[0].t) / 60000
			if spanMin >= 3 {
				bs := (pts[len(pts)-1].v - pts[0].v) / spanMin
				if bs > 0 {
					bucketBurn = bs
					bucketBurnSource = "sampled"
				}
			}
		}
		if bucketBurn <= 0.0001 && burn > 0.001 && burnSource != "idle" {
			bucketBurn = burn * (sessionWindowMin / weeklyWindowMin)
			bucketBurnSource = "estimated"
		}
		if bucketBurn > 0.0001 {
			minToBucketExhaust = (100 - bucketPct) / bucketBurn
		}
		if minToBucketExhaust > minToBucketReset {
			minToBucketExhaust = math.Inf(1)
		}
	}

	// --- zones: worst of session, overall weekly, ACTIVE model's bucket ---------
	zone5 := zoneOf(s5, z5)
	zone7 := zoneOf(s7, z7)
	zoneBucket := "GREEN"
	if bucketPct >= 0 {
		zoneBucket = zoneOf(bucketPct, z7)
	}
	zone := zone5
	for _, cand := range []string{zone7, zoneBucket} {
		if rank[cand] > rank[zone] {
			zone = cand
		}
	}
	limiter := "session"
	if zone == zone7 && rank[zone7] >= rank[zone5] {
		limiter = "weekly"
	}
	if zone == zoneBucket && rank[zoneBucket] >= rank[zone7] && rank[zoneBucket] >= rank[zone5] && bucketPct >= 0 {
		limiter = "model-bucket"
	}

	// --- persist cache ----------------------------------------------------------
	// The cache holds what the API said, when it was last asked, and nothing else. Where a live
	// figure stood in, the samples above were pruned and extended against it, so the ones read are
	// kept instead.
	samples := kept
	if liveFive || liveSeven {
		samples = nil
		if cache != nil {
			samples = cache.Samples
		}
	}
	out := cacheFile{Raw: rawBytes, Samples: samples, TriedAt: triedAt}
	switch {
	case fetchedFresh:
		out.FetchedAt = now.Format(time.RFC3339Nano)
	case cache != nil:
		out.FetchedAt = cache.FetchedAt
	}
	saveCache(out)

	return &state{
		s5: s5, s7: s7,
		r5Local: r5, r7Local: r7,
		burn: burn, burnSource: burnSource, paceRatio: paceRatio,
		minToExhaust: minToExhaust, minToReset5: minToReset5, minToReset7: minToReset7,
		horizonMin: horizonMin,
		zone:       zone, zone5: zone5, zone7: zone7, zoneBucket: zoneBucket,
		limiter:      limiter,
		scoped:       scoped,
		activeModel:  activeModelID,
		activeBucket: activeBucket,
		bucketPct:    bucketPct,
		bucketBurn:   bucketBurn, bucketBurnSource: bucketBurnSource,
		minToBucketExhaust: minToBucketExhaust,
		minToBucketRst:     minToBucketReset,
		bucketResetLocal:   bucketResetLocal,
		sampleCount:        len(kept),
		fetchedAt:          fetchedAt,
		stale:              stale,
	}
}

// rememberTheAttempt writes down that the API was just asked, where there were no figures to draw
// from, so that it is not asked again before its time. An answer the API did give — a window that
// has not started — replaces the cached one: left behind the new attempt, the last window's figures,
// its reset already past, were drawn until the next fetch, and a spent window read RED.
func rememberTheAttempt(cache *cacheFile, triedAt string, fresh json.RawMessage) {
	if cache == nil {
		cache = &cacheFile{}
	}
	if fresh == nil && triedAt == cache.TriedAt {
		return
	}
	kept := *cache
	kept.TriedAt = triedAt
	if fresh != nil {
		kept.Raw = fresh
		kept.FetchedAt = triedAt
	}
	saveCache(kept)
}

// saveCache writes the usage cache, where Claude Code keeps its own files — and nowhere, when that
// folder does not exist.
func saveCache(c cacheFile) {
	if fi, err := os.Stat(claudeDir); err == nil && fi.IsDir() {
		if b, err := json.Marshal(c); err == nil {
			_ = writeFileAtomic(cachePath, b)
		}
	}
}

// writeFileAtomic replaces a file in one step, so a reader running at the same moment — another
// session's status line, the gate — gets the old document or the new one and never half of each. A
// torn read is taken for a missing cache, which throws the burn samples away with it. The one
// exception is the in-place fallback below, taken only when the rename is refused.
func writeFileAtomic(path string, b []byte) error {
	sweepTemps(path)
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	_, werr := tmp.Write(b)
	cerr := tmp.Close()
	if werr == nil && cerr == nil {
		if werr = os.Rename(name, path); werr == nil {
			return nil
		}
		// On Windows a file another session holds open for reading cannot be replaced by a rename
		// ("Access is denied"), and the update was lost (review of PR 697). It is written in place
		// instead. A reader caught mid-write — most likely the very one that held the file open —
		// reads a torn file and takes it for none: for the usage cache that is one fetch out of
		// turn and the burn samples lost, for the token cache one full recount of the transcripts.
		// Rarer and cheaper than every such write lost.
		if werr = os.WriteFile(path, b, 0o600); werr == nil {
			_ = os.Remove(name)
			return nil
		}
	}
	_ = os.Remove(name)
	if werr != nil {
		return werr
	}
	return cerr
}

// sweepTemps removes what an interrupted writeFileAtomic left behind. Claude Code stops a status
// line that is still running when the next redraw is due, so a run can die between writing its
// temporary file and renaming it, and at a redraw every half-minute those would pile up. Only files
// older than a minute are touched: a younger one may be another run's write in progress.
func sweepTemps(path string) {
	leftovers, _ := filepath.Glob(path + ".*.tmp")
	for _, p := range leftovers {
		if info, err := os.Stat(p); err == nil && time.Since(info.ModTime()) > time.Minute {
			_ = os.Remove(p)
		}
	}
}

// --- formatting ---------------------------------------------------------------

// plural spells the unit out, singular only for exactly one.
func plural(value, unit string) string {
	if value == "1" {
		return value + " " + unit
	}
	return value + " " + unit + "S"
}

// formatCountdown renders "time left until this limit resets": minutes up to two hours, then hours
// up to two days, then days.
func formatCountdown(m float64) string {
	if m < 120 {
		return plural(num(m, 0), "MINUTE")
	}
	if m < 2880 {
		return plural(num(m/60, 0), "HOUR")
	}
	return plural(num(m/1440, 1), "DAY")
}

// formatWait renders the span for the AFTER cell as a clock-style duration — `3:49` below a day,
// `6 days 17:25` above one — so the wait reads exactly, at a glance, instead of as a rounded
// "5 HOURS" or a fraction of a day.
func formatWait(m float64) string {
	total := int(math.Round(m))
	if total < 0 {
		total = 0
	}
	if total < 1440 {
		return fmt.Sprintf("%d:%02d", total/60, total%60)
	}
	days, rest := total/1440, total%1440
	word := "days"
	if days == 1 {
		word = "day"
	}
	return fmt.Sprintf("%d %s %02d:%02d", days, word, rest/60, rest%60)
}

// The status line is a small table: one row per limit, columns separated by a light vertical bar
// and padded so the numbers line up under each other. Claude Code renders every printed line as its
// own row of the status area.
// usageBar draws a percentage as a fixed-width gauge, so "how full is it" reads without arithmetic.
func usageBar(pct float64) string { return barOf(pct, barCells) }

// barOf draws a percentage as a gauge of `cells` cells and fills one cell per *started* slice of it —
// with the usual 20 cells that is one cell per started 5 %, so the gauge never reads empty while
// something has already been spent. Computed from the same rounded percentage the row prints, so the
// bar and the number can never disagree.
func barOf(pct float64, cells int) string {
	if cells <= 0 {
		return ""
	}
	filled := int(math.Ceil(math.Round(pct) * float64(cells) / 100))
	if filled < 0 {
		filled = 0
	}
	if filled > cells {
		filled = cells
	}
	return strings.Repeat(string(barFilled), filled) + strings.Repeat(string(barEmpty), cells-filled)
}

// contextTokens is how much context the session is holding right now: the tokens the last
// main-thread request carried — fresh input plus everything written to or read from the prompt
// cache — taken from the session transcript Claude Code hands the hook. Subagent (sidechain) turns
// are skipped: they run in their own context. The transcript is append-only and can be megabytes,
// so only its tail is read.
//
// Reports false until the session's first reply is on disk — on the very first prompt of a session
// there is no usage entry yet, so the CONTEXT line is simply absent rather than showing a zero.
func contextTokens(transcriptPath string) (float64, bool) {
	if transcriptPath == "" {
		return 0, false
	}
	f, err := os.Open(transcriptPath)
	if err != nil {
		return 0, false
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return 0, false
	}
	const tailBytes = int64(2 << 20)
	start := int64(0)
	if fi.Size() > tailBytes {
		start = fi.Size() - tailBytes
	}
	if _, err := f.Seek(start, io.SeekStart); err != nil {
		return 0, false
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return 0, false
	}
	lines := strings.Split(string(b), "\n")
	if start > 0 && len(lines) > 1 {
		lines = lines[1:] // the first line of the tail is cut in half
	}
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" || !strings.Contains(line, `"usage"`) {
			continue
		}
		var entry struct {
			IsSidechain bool `json:"isSidechain"`
			Message     struct {
				Usage *struct {
					Input  float64 `json:"input_tokens"`
					Create float64 `json:"cache_creation_input_tokens"`
					Read   float64 `json:"cache_read_input_tokens"`
				} `json:"usage"`
			} `json:"message"`
		}
		if json.Unmarshal([]byte(line), &entry) != nil || entry.IsSidechain || entry.Message.Usage == nil {
			continue
		}
		u := entry.Message.Usage
		return u.Input + u.Create + u.Read, true
	}
	return 0, false
}

// shortTokens prints a token count the way the counter reads best: 510k, 2m, 648m, 29.9b. Thousands
// and millions are whole; billions keep a tenth, where a whole one would hide a great deal.
func shortTokens(n float64) string {
	switch {
	case n >= 1e9:
		if v := n / 1e9; v < 100 {
			return strings.TrimSuffix(fmt.Sprintf("%.1f", v), ".0") + "b"
		}
		return fmt.Sprintf("%.0fb", n/1e9)
	case n >= 1e6:
		return fmt.Sprintf("%.0fm", n/1e6)
	case n >= 1000:
		return fmt.Sprintf("%.0fk", n/1000)
	default:
		return fmt.Sprintf("%.0f", n)
	}
}

// statusRow is one row of the table. For a limit, `reset` and `after` stay empty when the API
// reported no reset time — those columns are then left off entirely rather than printed
// half-filled. `text` turns the row into a free-form one (BURN / MODEL / ZONE): same label column,
// everything after it is prose, so the extra context lines line up with the limits above them.
type statusRow struct {
	label string
	pct   float64
	// reset is the clock time the limit resets at — "04:00" for today, "MON 18:00" once the cell
	// needs a weekday. Always labelled "RESET AT": it is a point in time, never a wait. The wait
	// itself is the "RESET AFTER" cell on the TIME line below, and labelling this one "RESET IN"
	// made the two read as contradicting each other.
	reset string
	// windowBar / windowPct are the second gauge: the same 20 cells and the same number, but for how
	// much of the window has already run out. Drawn under the usage gauge so the two can be compared.
	windowBar string
	windowPct float64
	// windowCaption names that second gauge. Empty means TIME, the calendar scale. The weekly rows
	// say WORK instead once day weights make the working week shorter than the calendar one — two
	// gauges measuring different things must not carry the same word.
	windowCaption string
	after         string
	text          string
}

// resetCells turns a reset timestamp into its two display columns.
func resetCells(reset *time.Time, minsLeft float64, withWeekday bool) (string, string) {
	if reset == nil {
		return "", ""
	}
	clock := fmtHHmm(*reset)
	if withWeekday {
		clock = strings.ToUpper(fmtDddHHmm(*reset))
	}
	after := ""
	if !math.IsInf(minsLeft, 1) && minsLeft >= 0 {
		after = formatWait(minsLeft)
	}
	return clock, after
}

// Widths are counted in runes: the gauges are drawn with multi-byte block characters, so len()
// would pad them wrong and the columns would drift apart. A colour code takes no column on screen,
// so every escape sequence is skipped rather than counted.
func runeWidth(s string) int {
	n := 0
	for i := 0; i < len(s); {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			// A CSI sequence runs to its final byte, the first one in @..~.
			j := i + 2
			for j < len(s) && (s[j] < 0x40 || s[j] > 0x7e) {
				j++
			}
			i = j + 1
			continue
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		i += size
		n++
	}
	return n
}

func padRunes(s string, width int) string {
	if w := runeWidth(s); w < width {
		return s + strings.Repeat(" ", width-w)
	}
	return s
}

// pctCell prints the percentage in a fixed-width field so the gauges after it start in the same
// column on both stacked lines. What the number counts is spelled out by captionCell next to it.
func pctCell(pct float64) string { return fmt.Sprintf("%3s %%", num(pct, 0)) }

// captionCell names what the gauge next to it measures — LIMIT is the budget spent, TIME the share
// of the window gone. Right-aligned to one width so the gauges start in the same column.
func captionCell(caption string) string { return fmt.Sprintf("%5s", caption) }

// contextReading is how full the session's context window is: the tokens the last request carried
// and the size of the window they are measured against.
type contextReading struct {
	tokens, window float64
	ok             bool
}

// renderOpts is how a picture is drawn. The zero value is the injected block's: plain text, laid out
// as wide as it needs to be.
type renderOpts struct {
	p     painter
	width int // columns available; zero when unknown, and then nothing is fitted
	// contextTail is written after the CONTEXT gauge, on the same line: the status line puts the
	// session's token count there.
	contextTail string
}

// limitBlock is one limit drawn as its two stacked lines. bottom starts after the label's columns,
// which are filled in once the block's place in a row is known — see renderLimits.
type limitBlock struct {
	top, bottom string
	labelWidth  int
	width       int
}

// limitBlocks draws every limit with gauges of the given length.
func limitBlocks(rows []statusRow, cells int, p painter) []limitBlock {
	blocks := make([]limitBlock, 0, len(rows))
	for _, r := range rows {
		// The LIMIT gauge's colour is how far the spend runs ahead of the time gauge under it. A limit
		// the API gave no reset for has no time gauge and is measured as if no time had passed, so a
		// nearly spent one still reads red.
		top := r.label + cellGap + p.paint(pctCell(r.pct), p.figuresColour()) + " " + captionCell("LIMIT") + " " +
			p.paint(barOf(r.pct, cells), p.gaugeColour((r.pct-r.windowPct)/limitAheadRed))
		bottom := cellGap
		// windowBar is only a presence mark here: it was drawn at the default length, and the gauge
		// is drawn again from its percentage at whatever length this layout needs.
		if r.windowBar != "" {
			caption := r.windowCaption
			if caption == "" {
				caption = captionCalendar
			}
			bottom += p.paint(pctCell(r.windowPct), p.figuresColour()) + " " + captionCell(caption) + " " +
				p.paint(barOf(r.windowPct, cells), colourTime)
		}
		if r.reset != "" {
			top += " RESET AT " + p.paint(r.reset, p.figuresColour())
			if r.after != "" && r.windowBar != "" {
				bottom += " RESET AFTER " + paintFigures(p, r.after)
			}
		}
		lw := runeWidth(r.label)
		blocks = append(blocks, limitBlock{top: top, bottom: bottom, labelWidth: lw,
			width: max(runeWidth(top), lw+runeWidth(bottom))})
	}
	return blocks
}

// packBlocks puts the blocks into rows no wider than width, in order and as many to a row as fit. A
// width of zero means the terminal's size is unknown, and every block goes on one row.
func packBlocks(blocks []limitBlock, width int) [][]limitBlock {
	if width <= 0 || len(blocks) == 0 {
		return [][]limitBlock{blocks}
	}
	var rows [][]limitBlock
	var cur []limitBlock
	used := 0
	for _, b := range blocks {
		need := b.width
		if len(cur) > 0 {
			need += runeWidth(blockGap)
			if used+need > width {
				rows = append(rows, cur)
				cur, used, need = nil, 0, b.width
			}
		}
		cur = append(cur, b)
		used += need
	}
	return append(rows, cur)
}

// renderLimits lays the limits out side by side, each as a two-line block: the budget spent on the
// top line, the share of the window already gone on the bottom one, drawn with the same gauge.
// Reading one against the other is the whole point — a fuller top bar than bottom bar means the
// budget is running out faster than the clock, i.e. it will not last to the reset.
//
// Given a width, the blocks wrap onto further rows rather than off the edge, and a terminal too
// narrow for one block at full size gets gauges half as long.
func renderLimits(rows []statusRow, ctx contextReading, o renderOpts) string {
	blocks := limitBlocks(rows, barCells, o.p)
	widest := 0
	for _, b := range blocks {
		widest = max(widest, b.width)
	}
	if o.width > 0 && widest > o.width {
		blocks = limitBlocks(rows, narrowBarCells, o.p)
	}

	lines := make([]string, 0, 3)
	if top := topLine(ctx, o); top != "" {
		lines = append(lines, top)
	}
	for _, group := range packBlocks(blocks, o.width) {
		tops := make([]string, 0, len(group))
		bottoms := make([]string, 0, len(group))
		for i, b := range group {
			// The bottom line carries no label of its own — just the space its block's label takes, so
			// the two lines stay column-for-column. The first column of a row is the exception: Claude
			// Code trims leading whitespace off every status-line row, which would shift the whole line
			// left, so a quiet marker holds that one open.
			area := strings.Repeat(" ", b.labelWidth)
			if i == 0 && b.labelWidth > 0 {
				area = "-" + strings.Repeat(" ", b.labelWidth-1)
			}
			tops = append(tops, padRunes(b.top, b.width))
			bottoms = append(bottoms, padRunes(area+b.bottom, b.width))
		}
		lines = append(lines,
			strings.TrimRight(statusIndent+strings.Join(tops, blockGap), " "),
			strings.TrimRight(statusIndent+strings.Join(bottoms, blockGap), " "),
		)
	}
	return strings.Join(lines, "\n")
}

// contextLine draws the CONTEXT gauge: one cell per percent, so it reads straight as the number
// beside it — unless the terminal is too narrow for a hundred cells, when it shortens to fit.
func contextLine(ctx contextReading, o renderOpts) string {
	pct := ctx.tokens / ctx.window * 100
	if pct > 100 {
		pct = 100
	}
	// The counter and the percentage are in the gauge's own colour, so the line reads as one.
	colour := o.p.gaugeColour(pct / 100)
	head := statusIndent + "CONTEXT" + cellGap + o.p.paint(shortTokens(ctx.tokens)+"/"+shortTokens(ctx.window), colour) +
		cellGap + o.p.paint(pctCell(pct), colour) + cellGap
	tail := ""
	if o.contextTail != "" {
		tail = " " + o.contextTail
	}
	cells := ctxBarCells
	if o.width > 0 {
		if room := o.width - runeWidth(head) - runeWidth(tail); room < cells {
			cells = max(room, minContextCells)
		}
	}
	return head + o.p.paint(barOf(pct, cells), colour) + tail
}

// topLine is the line above the limits: the CONTEXT gauge with whatever the status line writes after
// it, that alone while there is no context reading yet, or nothing.
func topLine(ctx contextReading, o renderOpts) string {
	if ctx.ok && ctx.window > 0 {
		return contextLine(ctx, o)
	}
	return o.contextTail
}

// gaugeColour is the colour at a point of the green-to-red scale, from 0 (green) to 1 (red); see
// gaugeHueStart for what places a gauge on it.
func (p painter) gaugeColour(at float64) string {
	at = math.Max(0, math.Min(1, at))
	if !p.rgb {
		return "38;5;" + strconv.Itoa(gaugePalette256[int(math.Round(at*10))])
	}
	r, g, b := hsvToRGB(gaugeHueStart+(gaugeHueEnd-gaugeHueStart)*at, gaugeSaturated, gaugeBright)
	return "38;2;" + strconv.Itoa(r) + ";" + strconv.Itoa(g) + ";" + strconv.Itoa(b)
}

// paintFigures paints the figures of a wait like "6 days 20:02" and leaves its words alone.
func paintFigures(p painter, s string) string {
	fields := strings.Split(s, " ")
	for i, f := range fields {
		if strings.ContainsAny(f, "0123456789") {
			fields[i] = p.paint(f, p.figuresColour())
		}
	}
	return strings.Join(fields, " ")
}

// figuresColour is the colour of the figures; see figuresRGB.
func (p painter) figuresColour() string {
	if !p.rgb {
		return "38;5;" + strconv.Itoa(figures256)
	}
	return "38;2;" + strconv.Itoa(figuresRGB[0]) + ";" + strconv.Itoa(figuresRGB[1]) + ";" + strconv.Itoa(figuresRGB[2])
}

// hsvToRGB converts a hue in degrees, from red at 0 through green at 120, with a saturation and a
// brightness from 0 to 1, into 8-bit red, green and blue.
func hsvToRGB(h, s, v float64) (int, int, int) {
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	var r, g, b float64
	switch {
	case h < 60:
		r, g = c, x
	case h < 120:
		r, g = x, c
	default:
		g, b = c, x
	}
	m := v - c
	to8 := func(f float64) int { return int(math.Round((f + m) * 255)) }
	return to8(r), to8(g), to8(b)
}

// renderTextRows prints the BURN / MODEL / ZONE lines under the limits, one per line.
func renderTextRows(rows []statusRow) string {
	labelWidth := 0
	for _, r := range rows {
		if runeWidth(r.label) > labelWidth {
			labelWidth = runeWidth(r.label)
		}
	}
	lines := make([]string, 0, len(rows))
	for _, r := range rows {
		lines = append(lines, strings.TrimRight(statusIndent+padRunes(r.label, labelWidth)+cellGap+r.text, " "))
	}
	return strings.Join(lines, "\n")
}

// --- the working-time scale behind the weekly WORK gauge ----------------------
//
// The shape of the working week — which hours of which day count, and for how much — is in
// constants.go. This is what reads it.

var (
	timeZoneName = defaultTimeZone
	weekLoc      *time.Location
)

// weekLocation resolves the configured zone once. A name this build cannot load falls back to the
// machine's own zone instead of failing — a limits helper must never take the session down, and a
// misspelled zone should cost the working week its portability and nothing more.
func weekLocation() *time.Location {
	if weekLoc != nil {
		return weekLoc
	}
	weekLoc = time.Local
	if timeZoneName != "" {
		if loc, err := time.LoadLocation(timeZoneName); err == nil {
			weekLoc = loc
		}
	}
	return weekLoc
}

// configProblems holds everything the config file asked for that this binary could not act on. A
// setting silently ignored is worse than one refused out loud: the person believes the weekend is
// weighted, or the evening excluded, and it is not, and nothing on screen says otherwise. Shown as
// a CONFIG line under the gauges, which is empty on any correct config and so costs nothing until
// something is wrong.
var configProblems []string

// workingRateAt is how fast working time accrues at an instant: the day's percentage while the
// clock is inside that day's working hours, and nothing at all outside them. It is the whole of the
// arithmetic — everything below only finds the moments at which it changes.
func workingRateAt(t time.Time) float64 {
	local := t.In(weekLocation())
	day := workingWeek[int(local.Weekday())]
	if day.ToMin <= day.FromMin {
		return 0 // a day not worked, or a range refused for running backwards
	}
	minute := local.Hour()*60 + local.Minute()
	if minute < day.FromMin || minute >= day.ToMin {
		return 0
	}
	return day.PercentOfDay / 100
}

// searchStepMin is the coarse stride of the search below, cached because it only changes when the
// configured week does.
var searchStepMin int

// searchStep is the shortest stretch the configured week can hold — a working window, or the run of
// night either side of one — so the coarse pass cannot step over a stretch entirely. Taking the
// shortest *piece* rather than the shortest run is deliberately conservative: two pieces that meet
// at midnight merge into a longer run, never a shorter one.
func searchStep() time.Duration {
	if searchStepMin > 0 {
		return time.Duration(searchStepMin) * time.Minute
	}
	shortest := int(searchStepCeilingMin / time.Minute)
	for _, day := range workingWeek {
		if day.ToMin <= day.FromMin {
			continue
		}
		for _, piece := range []int{day.ToMin - day.FromMin, day.FromMin, 1440 - day.ToMin} {
			if piece > 0 && piece < shortest {
				shortest = piece
			}
		}
	}
	if shortest < 1 {
		shortest = 1
	}
	searchStepMin = shortest
	return time.Duration(shortest) * time.Minute
}

// nextRateChange returns the first instant after cur at which the working rate changes.
//
// It is searched rather than constructed, and that is the point. Midnight and the working hours
// look like landmarks you can build, and they are not: Chile and Cuba move their clocks at exactly
// 00:00, so a local day can have no midnight at all, and asking the calendar for one hands back an
// instant on the day before. A walk built on that charged a whole stretch to the wrong day — with
// the shipped weights, 1104 minutes of a week for America/Santiago and America/Havana, and 48 for
// Africa/Cairo and Asia/Beirut.
//
// "What is the rate at this instant" is a question every zone answers correctly, so a search over
// it is right wherever a constructed boundary is not. Coarse pass first, then refined to the minute
// and to the second: working hours and clock changes both land on whole minutes.
func nextRateChange(cur time.Time) time.Time {
	rate := workingRateAt(cur)
	step := searchStep()
	hi := cur
	for i := 0; i < int(30*time.Hour/step)+2 && workingRateAt(hi) == rate; i++ {
		hi = hi.Add(step)
	}
	if workingRateAt(hi) == rate {
		// Thirty hours at one rate is not a week any config can describe; returning hi keeps the
		// walk finite rather than pretending to have found something.
		return hi
	}
	lo := hi.Add(-step)
	for _, fine := range [2]time.Duration{time.Minute, time.Second} {
		if fine >= step {
			continue
		}
		for t := lo.Add(fine); t.Before(hi); t = t.Add(fine) {
			if workingRateAt(t) != rate {
				hi = t
				break
			}
		}
		lo = hi.Add(-fine)
	}
	return hi
}

// weightedMinutes is how much working time lies between two moments: every stretch counted at the
// rate in force across it, so the gauge moves through a working afternoon, stands still overnight,
// and never jumps at a boundary.
//
// Hours and weekdays of the configured zone, not UTC ones: the working day is the user's. Walked
// one stretch at a time rather than divided arithmetically, so a daylight-saving day of 23 or 25
// hours counts as the day it actually was, and the hour a zone repeats when its clock goes back is
// counted under the day it is repeated on.
func weightedMinutes(from, to time.Time) float64 {
	if !to.After(from) {
		return 0
	}
	total := 0.0
	cur := from
	// A week holds two stretches a day, plus whatever a clock change splits; the counter only
	// guarantees the walk ends.
	for i := 0; i < 128 && cur.Before(to); i++ {
		next := nextRateChange(cur)
		if next.After(to) {
			next = to
		}
		total += next.Sub(cur).Minutes() * workingRateAt(cur)
		cur = next
	}
	return total
}

// workingScaleIsCalendar reports whether the configured week measures exactly what the calendar
// does — every day the same percentage and worked round the clock. Only the ratios matter, so seven
// days at 50 from 00:00 to 24:00 measure identically to seven at 100. A week that works no hours at
// all also counts as the calendar, because there is then nothing to measure against and the gauge
// falls back to it. In all of those the two scales are one ruler and the gauge keeps the caption
// that says so.
func workingScaleIsCalendar() bool {
	first := workingWeek[0]
	sameShape, total := true, 0.0
	for _, day := range workingWeek {
		if day.PercentOfDay != first.PercentOfDay || day.FromMin != 0 || day.ToMin != 1440 {
			sameShape = false
		}
		if day.ToMin > day.FromMin {
			total += float64(day.ToMin-day.FromMin) * day.PercentOfDay
		}
	}
	return sameShape || total == 0
}

// clampPercent keeps a configured percentage inside the scale it is written on. A negative day
// would run the gauge backwards and one over 100 would say a day is worth more than a whole day.
func clampPercent(v float64) float64 {
	if math.IsNaN(v) || v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

// weekdayIndex maps a configured key onto a weekday, accepting the full English name or its
// three-letter abbreviation, in any case. Deliberately exact: matching on the first three letters
// took "saturdayy" for Saturday and dropped "weekend" without a word, so one typo bound to a day
// nobody named and another vanished, both in silence.
func weekdayIndex(name string) int {
	key := strings.ToLower(strings.TrimSpace(name))
	for i := 0; i < 7; i++ {
		full := strings.ToLower(time.Weekday(i).String())
		if key == full || key == full[:3] {
			return i
		}
	}
	return -1
}

// parseClock reads a working-hour bound written as "HH:MM" into minutes from midnight. "24:00" is
// accepted and is the only way to say a whole day.
func parseClock(s string) (int, bool) {
	parts := strings.Split(strings.TrimSpace(s), ":")
	if len(parts) != 2 {
		return 0, false
	}
	hour, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	minute, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil {
		return 0, false
	}
	if hour < 0 || hour > 24 || minute < 0 || minute > 59 || (hour == 24 && minute != 0) {
		return 0, false
	}
	return hour*60 + minute, true
}

// workingDayConfig is one day as the config file writes it. Every field is optional: a day that
// gives only a percentage keeps the default hours, and one that gives only hours keeps the default
// percentage.
type workingDayConfig struct {
	Percent *float64 `json:"percent"`
	From    *string  `json:"from"`
	To      *string  `json:"to"`
}

// workingDayFields are the only names a day understands. A day written with none of them —
// {"pct": 50, "start": "09:00"} — decodes into an empty struct without error and changes nothing,
// which is the silence this file exists to prevent, so the names are checked rather than the decode.
var workingDayFields = map[string]bool{"percent": true, "from": true, "to": true}

// applyWorkingWeek folds a configured week onto the defaults. Keyed by weekday and forgiving about
// the spelling — "Monday", "MON" and "mon" all land on the same day — and a day the config leaves
// out keeps its default rather than silently dropping to nothing.
func applyWorkingWeek(cfg map[string]workingDayConfig) {
	// Keys are taken in sorted order, never in map order. Ranging a Go map is unordered, so with
	// both "sat" and "Saturday" present the winner was drawn afresh on every run and the weekly bar
	// jumped between values with nothing on screen and nothing in the file having changed — which is
	// what somebody gets for adding a day beside the shipped three-letter one instead of editing it.
	keys := make([]string, 0, len(cfg))
	for name := range cfg {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	claimed := map[int]string{}
	for _, name := range keys {
		i := weekdayIndex(name)
		if i < 0 {
			configProblems = append(configProblems,
				"working_week: "+strconv.Quote(name)+" names no weekday, so it was ignored")
			continue
		}
		if first, ok := claimed[i]; ok {
			// One day named twice is a contradiction only the author can settle, so say which key
			// won rather than pick one in silence.
			configProblems = append(configProblems, "working_week: "+strconv.Quote(first)+" and "+
				strconv.Quote(name)+" both name "+time.Weekday(i).String()+"; "+strconv.Quote(first)+" is used")
			continue
		}
		claimed[i] = name

		day := workingWeek[i]
		entry := cfg[name]
		if entry.Percent != nil {
			day.PercentOfDay = clampPercent(*entry.Percent)
			if day.PercentOfDay != *entry.Percent {
				// A config asking for 500 is asking for something it will not get; saying so is
				// cheaper than letting somebody wonder why the day did not get heavier.
				configProblems = append(configProblems, "working_week: "+strconv.Quote(name)+" percent "+
					strconv.FormatFloat(*entry.Percent, 'f', -1, 64)+" is outside 0..100 and was read as "+
					strconv.FormatFloat(day.PercentOfDay, 'f', -1, 64))
			}
		}
		for _, bound := range []struct {
			label string
			text  *string
			into  *int
		}{{"from", entry.From, &day.FromMin}, {"to", entry.To, &day.ToMin}} {
			if bound.text == nil {
				continue
			}
			m, ok := parseClock(*bound.text)
			if !ok {
				configProblems = append(configProblems, "working_week: "+strconv.Quote(name)+" "+
					bound.label+" "+strconv.Quote(*bound.text)+" is not a time of day like \"12:00\", so it was ignored")
				continue
			}
			*bound.into = m
		}
		if day.ToMin < day.FromMin {
			// Refused rather than wrapped past midnight: a night shift is a different decision, and
			// quietly accepting one would make the week a shape nobody asked for.
			configProblems = append(configProblems, "working_week: "+strconv.Quote(name)+
				" ends before it starts, and working hours do not run past midnight; the day is not counted")
			day.ToMin = day.FromMin
		}
		workingWeek[i] = day
	}
	searchStepMin = 0 // the week changed; the coarse stride is worked out again on next use
}

// windowBar is the second gauge: it fills as the window runs out, so when it reaches the end the
// limit resets. Same cells and same fill rule as usageBar, and read together with it — the left
// gauge is how much of the budget is gone, this one how much of the time is gone.
func windowGauge(minsLeft, windowMin float64) (string, float64) {
	if math.IsInf(minsLeft, 1) || minsLeft < 0 || windowMin <= 0 {
		return "", 0
	}
	elapsed := windowMin - minsLeft
	if elapsed < 0 {
		elapsed = 0
	}
	pct := elapsed / windowMin * 100
	return usageBar(pct), pct
}

// weeklyWindowGauge is windowGauge on the working-time scale: the same gauge and the same number,
// but measuring how much of the *working* week has gone rather than how much of the calendar one.
// The window itself is unchanged — it still starts seven real days before the reset it is given.
func weeklyWindowGauge(reset, now time.Time) (string, float64) {
	if reset.IsZero() || reset.Before(now) {
		return "", 0
	}
	start := reset.Add(-time.Duration(weeklyWindowMin) * time.Minute)
	whole := weightedMinutes(start, reset)
	if whole <= 0 {
		// Every day weighted zero leaves nothing to measure against. Falling back to the calendar
		// keeps the gauge saying something true rather than reading a permanent zero.
		return windowGauge(reset.Sub(now).Minutes(), weeklyWindowMin)
	}
	pct := weightedMinutes(start, now) / whole * 100
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	return usageBar(pct), pct
}

// limitRows builds one table row per limit the API reports: the session window, the weekly window,
// then every per-model weekly bucket — whether or not it belongs to the model in use.
func limitRows(st *state) []statusRow {
	now := time.Now()
	// The weekly gauges alone move to the working-time scale; the session window is five hours of one
	// day and weighting it would say nothing.
	weeklyCaption := ""
	if !workingScaleIsCalendar() {
		weeklyCaption = captionWorking
	}
	sessionReset, sessionAfter := resetCells(&st.r5Local, st.minToReset5, false)
	sessionBar, sessionPct := windowGauge(st.minToReset5, sessionWindowMin)
	rows := []statusRow{{
		label:     "5 HOURS",
		pct:       st.s5,
		reset:     sessionReset,
		windowBar: sessionBar,
		windowPct: sessionPct,
		after:     sessionAfter,
	}}

	weekly := statusRow{label: "7 DAYS", pct: st.s7}
	if !st.r7Local.IsZero() {
		weekly.reset, weekly.after = resetCells(&st.r7Local, st.minToReset7, true)
		weekly.windowBar, weekly.windowPct = weeklyWindowGauge(st.r7Local, now)
		weekly.windowCaption = weeklyCaption
	}
	rows = append(rows, weekly)

	for _, sc := range st.scoped {
		row := statusRow{label: strings.ToUpper(sc.Model), pct: sc.Percent}
		if t, ok := parseTime(sc.ResetsAt); ok {
			row.reset, row.after = resetCells(&t, t.Sub(now).Minutes(), true)
			row.windowBar, row.windowPct = weeklyWindowGauge(t, now)
			row.windowCaption = weeklyCaption
		}
		rows = append(rows, row)
	}
	return rows
}

// sortedKeys is the order every part of the config is read in. Ranging a Go map is unordered, and
// a config that reads differently between two runs of the same binary is the worst kind of setting.
func sortedKeys(m map[string]json.RawMessage) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// budgetRuleInstalled reports whether this project carries the behaviour rule the injection points
// at. The hook supplies numbers; that rule is what tells the model to act on them, and it is a
// separate file a project may or may not have taken.
func budgetRuleInstalled() bool {
	root := os.Getenv("CLAUDE_PROJECT_DIR")
	if root == "" {
		return false
	}
	fi, err := os.Stat(filepath.Join(root, ".claude", "rules", "session-budget.md"))
	return err == nil && !fi.IsDir()
}

// configRows turns anything the config asked for and did not get into visible lines, in the shape
// the BURN / MODEL / ZONE rows already use. Empty on a correct config.
func configRows() []statusRow {
	rows := make([]statusRow, 0, len(configProblems))
	for _, p := range configProblems {
		rows = append(rows, statusRow{label: "CONFIG", text: p})
	}
	return rows
}

// formatSwitchAdvice names where heavy work can go when the model's own bucket is what binds.
func formatSwitchAdvice(st *state) string {
	if st.limiter != "model-bucket" {
		return ""
	}
	var others []string
	for _, sc := range st.scoped {
		if st.activeBucket != nil && sc.Model != st.activeBucket.Model {
			others = append(others, sc.Model+" "+num(sc.Percent, 0)+"%")
		}
	}
	alt := "models without their own bucket (limited only by overall weekly " + num(st.s7, 0) + "%)"
	if len(others) > 0 {
		alt = strings.Join(others, ", ") + ", or " + alt
	}
	return " Binding limit is the " + st.activeBucket.Model + " bucket - move heavy work to: " + alt + "; or wait for the weekly reset."
}

var zoneAdvice = map[string]string{
	"GREEN":  "spend to task size; subagents only when they clearly save tokens",
	"YELLOW": "be lean: max 2 parallel subagents, no heavy fan-outs, avoid re-reads",
	"ORANGE": "the session's first subagent is asked about and the rest run unprompted, so spawn only what you would defend; essential edits only, offer to defer XL tasks past reset",
	"RED":    "finish + checkpoint only; new tasks after reset",
}

// --- mode dispatch ------------------------------------------------------------

func run() {
	args := parseArgs(os.Args[1:])
	mode := strings.ToLower(args.str("mode", "inject"))
	hookEvent := args.str("event", "UserPromptSubmit")

	var hookInput map[string]any
	rawInput := ""
	if mode == "inject" || mode == "gate" || mode == "statusline" {
		if rawInput = readStdin(); rawInput != "" {
			_ = json.Unmarshal([]byte(rawInput), &hookInput)
		}
	}
	// Claude Code hands the status line its model; every other mode works it out, which can mean
	// reading the whole transcript.
	activeModel := ""
	if mode != "statusline" {
		activeModel = getActiveModel(args, hookInput)
	}

	switch mode {
	case "json":
		st := getState(args, activeModel, nil)
		if st == nil {
			// The config's own complaints do not depend on the API, and a malformed config is most
			// likely to be edited exactly when the hook cannot reach it — before a login, or offline.
			if b, err := json.Marshal(struct {
				Error          string   `json:"error"`
				ConfigProblems []string `json:"config_problems"`
			}{"usage data unavailable", configProblems}); err == nil {
				fmt.Print(string(b))
				return
			}
			fmt.Print(`{"error":"usage data unavailable"}`)
			return
		}
		type bucketJSON struct {
			Model        string   `json:"model"`
			Percent      float64  `json:"percent"`
			BurnPerMin   float64  `json:"burn_pct_per_min"`
			BurnSource   string   `json:"burn_source"`
			MinToExhaust *float64 `json:"min_to_exhaust"`
			MinToReset   *float64 `json:"min_to_reset"`
			ResetsAt     *string  `json:"resets_at"`
		}
		var bucketOut *bucketJSON
		if st.bucketPct >= 0 && st.activeBucket != nil {
			b := &bucketJSON{
				Model:      st.activeBucket.Model,
				Percent:    round(st.bucketPct, 1),
				BurnPerMin: round(st.bucketBurn, 4),
				BurnSource: st.bucketBurnSource,
			}
			if !math.IsInf(st.minToBucketExhaust, 1) {
				v := round(st.minToBucketExhaust, 0)
				b.MinToExhaust = &v
			}
			if !math.IsInf(st.minToBucketRst, 1) {
				v := round(st.minToBucketRst, 0)
				b.MinToReset = &v
			}
			if st.bucketResetLocal != nil {
				s := fmtYmdHM(*st.bucketResetLocal)
				b.ResetsAt = &s
			}
			bucketOut = b
		}
		var minToExhaust *float64
		if !math.IsInf(st.minToExhaust, 1) {
			v := round(st.minToExhaust, 0)
			minToExhaust = &v
		}
		// Infinite means "no answer", not a number: encoding/json refuses non-finite floats and
		// would drop the whole document, so an unknown reset distance goes out as null.
		var minToWeeklyReset *float64
		if !math.IsInf(st.minToReset7, 1) {
			v := round(st.minToReset7, 0)
			minToWeeklyReset = &v
		}
		// Same for the timestamp: an unset weekly window would otherwise print as year 1.
		var weeklyResetsAt *string
		if !st.r7Local.IsZero() {
			s := fmtYmdHM(st.r7Local)
			weeklyResetsAt = &s
		}
		out := struct {
			ActiveModel       string         `json:"active_model"`
			ActiveModelBucket *bucketJSON    `json:"active_model_bucket"`
			SessionPct        float64        `json:"session_pct"`
			WeeklyPct         float64        `json:"weekly_pct"`
			SessionResetsAt   string         `json:"session_resets_at"`
			WeeklyResetsAt    *string        `json:"weekly_resets_at"`
			BurnPerMin        float64        `json:"burn_pct_per_min"`
			BurnSource        string         `json:"burn_source"`
			PaceRatio         float64        `json:"pace_ratio"`
			MinToExhaust      *float64       `json:"min_to_exhaust"`
			MinToSessionReset float64        `json:"min_to_session_reset"`
			MinToWeeklyReset  *float64       `json:"min_to_weekly_reset"`
			HorizonMin        float64        `json:"horizon_min"`
			Zone              string         `json:"zone"`
			ZoneSession       string         `json:"zone_session"`
			ZoneWeekly        string         `json:"zone_weekly"`
			ZoneModelBucket   string         `json:"zone_model_bucket"`
			Limiter           string         `json:"limiter"`
			WeeklyScoped      []scopedBucket `json:"weekly_scoped"`
			BurnSamples       int            `json:"burn_samples"`
			ConfigProblems    []string       `json:"config_problems"`
		}{
			ActiveModel:       st.activeModel,
			ActiveModelBucket: bucketOut,
			SessionPct:        round(st.s5, 1),
			WeeklyPct:         round(st.s7, 1),
			SessionResetsAt:   fmtYmdHM(st.r5Local),
			WeeklyResetsAt:    weeklyResetsAt,
			BurnPerMin:        round(st.burn, 3),
			BurnSource:        st.burnSource,
			PaceRatio:         round(st.paceRatio, 2),
			MinToExhaust:      minToExhaust,
			MinToSessionReset: round(st.minToReset5, 0),
			MinToWeeklyReset:  minToWeeklyReset,
			HorizonMin:        round(st.horizonMin, 0),
			Zone:              st.zone,
			ZoneSession:       st.zone5,
			ZoneWeekly:        st.zone7,
			ZoneModelBucket:   st.zoneBucket,
			Limiter:           st.limiter,
			WeeklyScoped:      st.scoped,
			BurnSamples:       st.sampleCount,
			ConfigProblems:    configProblems,
		}
		b, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			// Never answer with silence: a caller that asked for JSON gets JSON either way.
			fmt.Print(`{"error":"usage state could not be encoded"}`)
			return
		}
		fmt.Print(string(b))

	case "inject":
		st := getState(args, activeModel, nil)
		if st == nil {
			// Nothing to hand over, but a broken config still has to be said out loud.
			if len(configProblems) > 0 {
				fmt.Print(unavailableBlock())
			}
			return
		}
		now := time.Now()
		transcript, _ := hookInput["transcript_path"].(string)
		ctx := contextReading{window: contextWindowTokens}
		ctx.tokens, ctx.ok = contextTokens(transcript)
		fiveStart, weekFrom := limitWindows(st)
		var tokens *tokenStats
		if t, ok := sessionTokenStats(transcript, fiveStart, weekFrom, now); ok {
			tokens = &t
		}
		fmt.Print(modelBlock(st, limitRows(st), ctx, tokens, hookEvent == "SessionStart", now))

	case "statusline":
		fmt.Print(renderStatusline(args, rawInput, hookInput, time.Now()))

	case "gate":
		if !gateEnabled {
			return
		}
		judgeModel := activeModel
		if hookInput != nil {
			if ti, ok := hookInput["tool_input"].(map[string]any); ok {
				// The independent grading pass the finishing-work rule requires on every change is
				// never gated: blocking it does not save the budget, it removes the only reviewer
				// who is not the author. It is one agent per revision, not a fan-out.
				if st, ok := ti["subagent_type"].(string); ok && st == "change-reviewer" {
					return
				}
				if m, ok := ti["model"].(string); ok && m != "" {
					judgeModel = cleanModelString(m)
				}
			}
		}
		st := getState(args, judgeModel, nil)
		if st == nil {
			return
		}
		if st.zone == "RED" || st.zone == "ORANGE" {
			target := "this session"
			if judgeModel != "" {
				target = "model '" + judgeModel + "'"
			}
			detail := "5h " + num(st.s5, 0) + "%, 7d " + num(st.s7, 0) + "%"
			if st.bucketPct >= 0 {
				detail += ", " + st.activeBucket.Model + " bucket " + num(st.bucketPct, 0) + "%"
			}
			switchTip := formatSwitchAdvice(st)
			var reason, decision string
			if st.zone == "RED" {
				reason = "Usage gate: RED for " + target + " (" + detail + "; limiter: " + st.limiter + "; session resets " + fmtHHmm(st.r5Local) + "). Subagent spawn blocked - work inline and compactly, defer heavy work past reset." + switchTip
				decision = "deny"
			} else {
				// Asked once per session, not once per spawn. The first fan-out in ORANGE is worth
				// stopping for; the fifth is nagging, and a gate that nags is a gate people switch
				// off to get work done, which protects nothing at all. RED still stops every spawn,
				// because RED is a wall rather than a warning.
				sid, _ := hookInput["session_id"].(string)
				if sid != "" && gateAlreadyAsked(sid) {
					return
				}
				if sid != "" {
					markGateAsked(sid)
				}
				reason = "Usage gate: ORANGE for " + target + " (" + detail + "; limiter: " + st.limiter + "). Confirm this spawn is worth the budget - asked once per session, not again. Session resets " + fmtHHmm(st.r5Local) + "." + switchTip
				decision = "ask"
			}
			out := map[string]any{
				"hookSpecificOutput": map[string]any{
					"hookEventName":            "PreToolUse",
					"permissionDecision":       decision,
					"permissionDecisionReason": reason,
				},
			}
			b, _ := json.Marshal(out)
			fmt.Print(string(b))
		}
	}
}

// Which sessions have already been asked in ORANGE, so the gate stops at one confirmation instead
// of one per spawn. Keyed by session id and pruned to a day, so the file cannot grow forever.
// Every failure path degrades to asking again rather than to silence: a lost marker costs one
// prompt, a wrongly remembered one costs the confirmation the zone exists for.
func gateAlreadyAsked(sessionID string) bool {
	b, err := os.ReadFile(gateAsksPath)
	if err != nil {
		return false
	}
	var seen map[string]int64
	if json.Unmarshal(b, &seen) != nil {
		return false
	}
	_, ok := seen[sessionID]
	return ok
}

func markGateAsked(sessionID string) {
	seen := map[string]int64{}
	if b, err := os.ReadFile(gateAsksPath); err == nil {
		_ = json.Unmarshal(b, &seen)
	}
	cutoff := time.Now().Add(-24 * time.Hour).Unix()
	for id, at := range seen {
		if at < cutoff {
			delete(seen, id)
		}
	}
	seen[sessionID] = time.Now().Unix()
	if b, err := json.Marshal(seen); err == nil {
		_ = os.WriteFile(gateAsksPath, b, 0o600)
	}
}

func loadConfig() {
	claudeDir = configDir()
	cachePath = filepath.Join(claudeDir, "usage-limits-cache.json")
	gateAsksPath = filepath.Join(claudeDir, "usage-limits-gate-asks.json")
	tokenCachePath = filepath.Join(claudeDir, "usage-limits-tokens.json")
	credPath = filepath.Join(claudeDir, ".credentials.json")

	// Project config first, then the user one. The binary lives in <project>/scripts/, and Claude Code
	// passes CLAUDE_PROJECT_DIR to hooks — try that first so a binary invoked from anywhere still finds
	// the project file.
	candidates := []string{}
	if root := os.Getenv("CLAUDE_PROJECT_DIR"); root != "" {
		candidates = append(candidates, filepath.Join(root, ".claude", "usage-limits-config.json"))
	}
	if exe, err := os.Executable(); err == nil {
		// Walk up from the binary rather than assuming it sits one folder below the project. It is
		// kept in scripts/, scripts/claude-code/ or .claude/bin/ depending on the project, and
		// assuming the first meant a binary in .claude/bin looked for its config one level too deep
		// and silently fell back to the user-level one.
		dir := filepath.Dir(exe)
		for i := 0; i < 4 && dir != ""; i++ {
			dir = filepath.Dir(dir)
			candidates = append(candidates, filepath.Join(dir, ".claude", "usage-limits-config.json"))
		}
	}
	configPath := filepath.Join(claudeDir, "usage-limits-config.json")
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			configPath = p
			break
		}
	}
	b, err := os.ReadFile(configPath)
	if err != nil {
		return
	}
	// Every setting is decoded on its own. Decoding the file as one struct meant a single value of
	// the wrong type — a weight written "20" rather than 20 is the easy slip — silently reverted the
	// zone thresholds, the fetch interval and the gate as well, and the only thing said about it was
	// that the file was not valid JSON, which it was.
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(stripBOM(b), &fields); err != nil {
		configProblems = append(configProblems,
			"config file is not valid JSON and was ignored ("+err.Error()+"): "+configPath)
		return
	}
	field := func(name string, into any) bool {
		raw, ok := fields[name]
		if !ok {
			return false
		}
		// A JSON null decodes into every Go type without complaint and leaves the zero value behind,
		// so taking it as an answer silently turned the gate off, set the cache lifetime to nothing —
		// which made the hook call the API on every prompt — and dropped the named zone. It is a
		// setting written and not filled in, so the default stands and the file is told about it.
		if string(bytes.TrimSpace(raw)) == "null" {
			configProblems = append(configProblems, "config: "+strconv.Quote(name)+
				" is null, so the built-in default is used")
			return false
		}
		if err := json.Unmarshal(raw, into); err != nil {
			configProblems = append(configProblems, "config: "+strconv.Quote(name)+
				" is not the shape this setting takes and was ignored ("+err.Error()+")")
			return false
		}
		return true
	}

	var ctxTokens float64
	if field("context_window_tokens", &ctxTokens) {
		if ctxTokens > 0 {
			contextWindowTokens = ctxTokens
		} else {
			configProblems = append(configProblems, "config: \"context_window_tokens\" must be above zero, so "+
				strconv.FormatFloat(ctxTokens, 'f', -1, 64)+" was ignored")
		}
	}
	var ahead float64
	if field("limit_ahead_red_pct", &ahead) {
		if ahead > 0 && ahead <= 100 {
			limitAheadRed = ahead
		} else {
			// The lead is one percentage less another, so it never passes a hundred points: a threshold
			// above that is never reached. At zero the lead is divided by nothing, and below it the scale
			// runs backwards — red while the spend is behind the clock.
			configProblems = append(configProblems, "config: \"limit_ahead_red_pct\" must be above 0 and at most 100, so "+
				strconv.FormatFloat(ahead, 'f', -1, 64)+" was ignored")
		}
	}
	var ttl int
	if field("fetch_ttl_sec", &ttl) {
		if ttl >= 0 {
			fetchTTLSec = ttl
		} else {
			// A negative lifetime makes every cached document look expired, so the API is called on
			// every prompt and every status-line render. That is a typo, not a setting.
			configProblems = append(configProblems, "config: \"fetch_ttl_sec\" cannot be negative, so "+
				strconv.Itoa(ttl)+" was ignored")
		}
	}
	var gate bool
	if field("gate_enabled", &gate) {
		gateEnabled = gate
	}
	var tz string
	if field("time_zone", &tz) {
		timeZoneName = strings.TrimSpace(tz)
		weekLoc = nil // resolved again below
	}
	// One day at a time, for the same reason: a single bad entry should cost that day and not the
	// whole week.
	var rawDays map[string]json.RawMessage
	if field("working_week", &rawDays) {
		days := map[string]workingDayConfig{}
		for _, name := range sortedKeys(rawDays) {
			var fieldsOfDay map[string]json.RawMessage
			if json.Unmarshal(rawDays[name], &fieldsOfDay) == nil {
				for _, key := range sortedKeys(fieldsOfDay) {
					if !workingDayFields[strings.ToLower(strings.TrimSpace(key))] {
						configProblems = append(configProblems, "working_week: "+strconv.Quote(name)+" has no setting called "+
							strconv.Quote(key)+"; a day takes \"percent\", \"from\" and \"to\"")
					}
				}
			}
			var day workingDayConfig
			if err := json.Unmarshal(rawDays[name], &day); err != nil {
				configProblems = append(configProblems, "working_week: "+strconv.Quote(name)+
					" is not a day like {\"percent\": 100, \"from\": \"12:00\", \"to\": \"20:00\"} and was ignored ("+
					err.Error()+")")
				continue
			}
			days[name] = day
		}
		applyWorkingWeek(days)
	}
	// The key this one replaced. Left unread it would silently take a weekend weighting away from a
	// config that still carries it, which is exactly the kind of silence this file reports on.
	if _, ok := fields["week_day_weights"]; ok {
		configProblems = append(configProblems,
			"config: \"week_day_weights\" is no longer read; it is now \"working_week\", "+
				"which carries the working hours alongside the percentage")
	}
	// Resolved now rather than on first use, so a zone this build cannot load is reported with the
	// rest of the config instead of after the picture has already been drawn.
	if timeZoneName != "" && weekLocation().String() != timeZoneName {
		configProblems = append(configProblems, "time_zone: "+strconv.Quote(timeZoneName)+
			" names no zone this build knows, so day boundaries follow this machine's own zone")
	}
	// One threshold at a time, for the same reason the days are: a single mistyped number should not
	// take all six with it, and a level this hook has never heard of should be named rather than
	// dropped in the silence an unknown weekday is not dropped in.
	var rawZones map[string]json.RawMessage
	if field("zones", &rawZones) {
		scopes := map[string]map[string]float64{"session": z5, "weekly": z7}
		for _, scope := range sortedKeys(rawZones) {
			into, known := scopes[scope]
			if !known {
				configProblems = append(configProblems, "zones: "+strconv.Quote(scope)+
					" is neither \"session\" nor \"weekly\", so it was ignored")
				continue
			}
			var levels map[string]json.RawMessage
			if err := json.Unmarshal(rawZones[scope], &levels); err != nil {
				configProblems = append(configProblems, "zones: "+strconv.Quote(scope)+
					" is not a set of thresholds and was ignored ("+err.Error()+")")
				continue
			}
			for _, level := range sortedKeys(levels) {
				if _, ok := into[level]; !ok {
					configProblems = append(configProblems, "zones."+scope+": "+strconv.Quote(level)+
						" names no zone, so it was ignored")
					continue
				}
				var v float64
				if err := json.Unmarshal(levels[level], &v); err != nil {
					configProblems = append(configProblems, "zones."+scope+": "+strconv.Quote(level)+
						" is not a number and was ignored ("+err.Error()+")")
					continue
				}
				if math.IsNaN(v) || v < 0 || v > 100 {
					// A threshold is a percentage of a budget. One outside the scale is not a strict
					// setting, it is a typo: below zero pins the zone on permanently and denies every
					// spawn from then on, above a hundred can never be reached at all.
					configProblems = append(configProblems, "zones."+scope+": "+strconv.Quote(level)+" is "+
						strconv.FormatFloat(v, 'f', -1, 64)+", which is outside 0..100, so it was ignored")
					continue
				}
				into[level] = v
			}
		}
	}
}

func main() {
	defer func() {
		// a limits helper must never take the session down
		recover()
		os.Exit(0)
	}()
	loadConfig()
	run()
}
