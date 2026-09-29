package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// --- the injected block must not change ---------------------------------------

// The status line's picture drawn plain — no colour, no known width — is the layout everything else
// paints over, and it is pinned byte for byte.
func TestPlainPicture(t *testing.T) {
	rows := []statusRow{
		{label: "5 HOURS", pct: 48, reset: "23:49", windowBar: usageBar(34), windowPct: 34, after: "3:17"},
		{label: "7 DAYS", pct: 13, reset: "MON 17:59", windowBar: usageBar(5), windowPct: 5, windowCaption: "WORK", after: "6 days 21:27"},
		{label: "FABLE", pct: 0, reset: "MON 18:00", windowBar: usageBar(5), windowPct: 5, windowCaption: "WORK", after: "6 days 21:27"},
		{label: "OPUS", pct: 61},
	}
	want := "CONTEXT  168k/1m   17 %  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░\n5 HOURS   48 % LIMIT ▓▓▓▓▓▓▓▓▓▓░░░░░░░░░░ RESET AT 23:49      7 DAYS   13 % LIMIT ▓▓▓░░░░░░░░░░░░░░░░░ RESET AT MON 17:59          FABLE    0 % LIMIT ░░░░░░░░░░░░░░░░░░░░ RESET AT MON 18:00          OPUS   61 % LIMIT ▓▓▓▓▓▓▓▓▓▓▓▓▓░░░░░░░\n-         34 %  TIME ▓▓▓▓▓▓▓░░░░░░░░░░░░░ RESET AFTER 3:17              5 %  WORK ▓░░░░░░░░░░░░░░░░░░░ RESET AFTER 6 days 21:27             5 %  WORK ▓░░░░░░░░░░░░░░░░░░░ RESET AFTER 6 days 21:27"
	if got := renderLimits(rows, contextReading{tokens: 168000, window: 1e6, ok: true}, renderOpts{}); got != want {
		t.Errorf("injected picture changed:\n got %q\nwant %q", got, want)
	}
	want = "5 HOURS   48 % LIMIT ▓▓▓▓▓▓▓▓▓▓░░░░░░░░░░ RESET AT 23:49      7 DAYS   13 % LIMIT ▓▓▓░░░░░░░░░░░░░░░░░ RESET AT MON 17:59\n-         34 %  TIME ▓▓▓▓▓▓▓░░░░░░░░░░░░░ RESET AFTER 3:17              5 %  WORK ▓░░░░░░░░░░░░░░░░░░░ RESET AFTER 6 days 21:27"
	if got := renderLimits(rows[:2], contextReading{}, renderOpts{}); got != want {
		t.Errorf("injected picture without context changed:\n got %q\nwant %q", got, want)
	}
	want = "BURN    0.47 %/min (sampled) - ~110 MINUTES to cap, pace x1.4\nCONFIG  x\nZONE    GREEN"
	if got := renderTextRows([]statusRow{{label: "BURN", text: "0.47 %/min (sampled) - ~110 MINUTES to cap, pace x1.4"}, {label: "CONFIG", text: "x"}, {label: "ZONE", text: "GREEN"}}); got != want {
		t.Errorf("text rows changed:\n got %q\nwant %q", got, want)
	}
}

// The model is handed the figures as data: semantic XML, one element per line, the format stated
// once on the root, and no gauge — which would tell it nothing the number does not and cost its
// context tokens on every prompt.
func TestModelBlock(t *testing.T) {
	now := time.Date(2026, 9, 28, 22, 20, 0, 0, time.Local)
	st := &state{
		r5Local: now.Add(91 * time.Minute), minToReset5: 91, minToExhaust: math.Inf(1),
		burn: 0.16, burnSource: "sampled", paceRatio: 0.8,
		zone: "YELLOW", limiter: "session",
		scoped:      []scopedBucket{{Model: "Fable", Percent: 0}, {Model: "Opus", Percent: 12}},
		activeModel: "claude-opus-5-5", bucketPct: -1, minToBucketExhaust: math.Inf(1),
	}
	rows := []statusRow{
		{label: "5 HOURS", pct: 58, reset: "23:51", windowBar: usageBar(70), windowPct: 70, after: "1:31"},
		{label: "7 DAYS", pct: 16, reset: "MON 18:00", windowBar: usageBar(5), windowPct: 5, windowCaption: captionWorking, after: "6 days 19:41"},
		{label: "FABLE", pct: 0, reset: "MON 18:00", windowBar: usageBar(5), windowPct: 5, windowCaption: captionWorking, after: "6 days 19:41"},
		{label: "OPUS", pct: 12}, // the API gave no reset for it
	}
	tokens := &tokenStats{session: tokenSums{Input: 2e6, Read: 164e6}, sessionWhole: true, account: true,
		fiveHour: tokenSums{Input: 19e6, Read: 1.1e9}, weekly: tokenSums{Input: 20e6, Read: 1.1e9}, accountWhole: true}
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	withoutConfigProblems(t, func() {
		got := modelBlock(st, rows, contextReading{tokens: 867000, window: 1e6, ok: true}, tokens, false, now)
		want := strings.Join([]string{
			`<usage_limits scope="account-global: every session on this subscription" format="tokens in k/m/b; clock times local; waits h:mm">`,
			`<context used="867k of 1m (87%)"/>`,
			`<limit name="5-hour" used="58%" gone="70% of the window" resets="23:51, in 1:31"/>`,
			`<limit name="7-day" used="16%" gone="5% of the working week" resets="MON 18:00, in 6 days 19:41"/>`,
			`<limit name="Fable weekly bucket" used="0%" gone="5% of the working week" resets="MON 18:00, in 6 days 19:41"/>`,
			`<limit name="Opus weekly bucket" used="12%"/>`,
			`<tokens session="fresh 2m, cached 164m" account_5h="fresh 19m, cached 1.1b" account_7d="fresh 20m, cached 1.1b"/>`,
			`<burn rate="0.16%/min, sampled" forecast="lasts to the reset at 23:51" pace="0.8"/>`,
			`<zone level="YELLOW" limiter="session">be lean: max 2 parallel subagents, no heavy fan-outs, avoid re-reads.</zone>`,
			`</usage_limits>`,
		}, "\n")
		if got != want {
			t.Errorf("block:\n%s\nwant:\n%s", got, want)
		}
		if strings.ContainsAny(got, "█░▓") {
			t.Error("a gauge reached the model")
		}

		// Session start adds the model; a stale figure, a config problem and the burn running to
		// the cap before the reset each say so — and text is escaped as XML.
		st.stale, st.fetchedAt = true, now.Add(-54*time.Minute)
		st.minToExhaust = 40
		configProblems = []string{`config: "fetch_ttl_sec" & <more>`}
		got = modelBlock(st, rows, contextReading{}, nil, true, now)
		for _, want := range []string{
			`format="tokens in k/m/b; clock times local; waits h:mm" at="session start">`,
			`<stale since="21:26">the usage API has not answered since then; every figure here is as of then</stale>`,
			`<burn rate="0.16%/min, sampled" forecast="reaches the cap in 0:40" pace="0.8"/>`,
			`<config_problem>config: "fetch_ttl_sec" &amp; &lt;more&gt;</config_problem>`,
			`<model id="claude-opus-5-5" own_bucket="none; the 7-day limit applies"/>`,
		} {
			if !strings.Contains(got, want) {
				t.Errorf("session-start block lacks %q:\n%s", want, got)
			}
		}
		if strings.Contains(got, "<context") || strings.Contains(got, "<tokens") {
			t.Errorf("a context or tokens element without a reading:\n%s", got)
		}
		if got := unavailableBlock(); got != "<usage_limits unavailable=\"the usage API gave no figures\">\n"+
			"<config_problem>config: \"fetch_ttl_sec\" &amp; &lt;more&gt;</config_problem>\n</usage_limits>" {
			t.Errorf("unavailable block = %q", got)
		}
	})
}

// --- colour and width ------------------------------------------------------------

func TestRuneWidthSkipsColourCodes(t *testing.T) {
	on := painter{on: true}
	for _, s := range []string{"5 HOURS", "▓▓░░"} {
		if got, want := runeWidth(on.paint(s, "38;5;208")), runeWidth(s); got != want {
			t.Errorf("width of painted %q = %d, want %d", s, got, want)
		}
	}
}

func TestPainter(t *testing.T) {
	if got := (painter{}).paint("x", "32"); got != "x" {
		t.Errorf("painter off painted: %q", got)
	}
	if got := (painter{on: true}).paint("x", "32"); got != "\x1b[32mx\x1b[0m" {
		t.Errorf("painter on: %q", got)
	}
	t.Setenv("NO_COLOR", "1")
	if statusPainter().on {
		t.Error("NO_COLOR did not turn the paint off")
	}
	t.Setenv("NO_COLOR", "")
	if !statusPainter().on {
		t.Error("an empty NO_COLOR turned the paint off; the convention is that only a value does")
	}
}

// The scale runs green at 0, yellow at a half, red at 1 — in 24-bit colour, and in eleven steps of
// the 256-colour palette for a terminal without it.
func TestGaugeColour(t *testing.T) {
	rgb, palette := painter{on: true, rgb: true}, painter{on: true}
	for _, c := range []struct {
		p    painter
		at   float64
		want string
	}{
		{rgb, 0, "38;2;36;242;36"},
		{rgb, 0.5, "38;2;242;242;36"},
		{rgb, 1, "38;2;242;36;36"},
		{rgb, -0.4, "38;2;36;242;36"},
		{rgb, 1.8, "38;2;242;36;36"},
		{palette, 0, "38;5;46"},
		{palette, 0.14, "38;5;82"},
		{palette, 0.5, "38;5;226"},
		{palette, 1, "38;5;196"},
	} {
		if got := c.p.gaugeColour(c.at); got != c.want {
			t.Errorf("gaugeColour(%v) rgb=%v = %q, want %q", c.at, c.p.rgb, got, c.want)
		}
	}
	// In 24-bit colour it turns smoothly: no two of these share a colour.
	seen := map[string]bool{}
	for at := 0.0; at <= 1.0001; at += 0.05 {
		seen[rgb.gaugeColour(at)] = true
	}
	if len(seen) != 21 {
		t.Errorf("21 points share %d colours", len(seen))
	}
}

func sampleRows() []statusRow {
	return []statusRow{
		{label: "5 HOURS", pct: 55, reset: "23:49", windowBar: usageBar(34), windowPct: 34, after: "3:17"},
		{label: "7 DAYS", pct: 13, reset: "MON 17:59", windowBar: usageBar(5), windowPct: 5, windowCaption: "WORK", after: "6 days 21:27"},
		{label: "FABLE", pct: 96, reset: "MON 18:00", windowBar: usageBar(5), windowPct: 5, windowCaption: "WORK", after: "6 days 21:27"},
	}
}

// Each gauge is painted whole in one colour: a LIMIT gauge by how far its spend runs ahead of the
// time gauge under it, the CONTEXT gauge — its figures with it — by how full it is. The limits'
// percentages are a soft yellow, the TIME/WORK gauges cyan, and every word keeps the terminal's
// colour.
func TestGaugesArePaintedWhole(t *testing.T) {
	p := painter{on: true, rgb: true}
	ctx := contextReading{tokens: 150000, window: 200000, ok: true}
	out := renderLimits(sampleRows(), ctx, renderOpts{p: p})
	figure := func(s string) string { return "\x1b[38;2;255;180;0m" + s + "\x1b[0m" }
	cyan := func(s string) string { return "\x1b[36m" + s + "\x1b[0m" }
	gauge := func(at float64, s string) string { return "\x1b[" + p.gaugeColour(at) + "m" + s + "\x1b[0m" }
	for _, want := range []string{
		"CONTEXT  " + gauge(0.75, "150k/200k") + "  " + gauge(0.75, " 75 %") + "  " + gauge(0.75, strings.Repeat("▓", 75)+strings.Repeat("░", 25)),
		// 55 % spent with 34 % of the time gone: 21 points ahead, past the 20 that make it red.
		"5 HOURS  " + figure(" 55 %") + " LIMIT " + gauge(1, strings.Repeat("▓", 11)+strings.Repeat("░", 9)) + " RESET AT " + figure("23:49"),
		// 13 % against 5 %: 8 points ahead, 0.4 of the way to red.
		"7 DAYS  " + figure(" 13 %") + " LIMIT " + gauge(0.4, "▓▓▓"+strings.Repeat("░", 17)) + " RESET AT " + figure("MON 17:59"),
		// 96 % against 5 %: far past it.
		"FABLE  " + figure(" 96 %") + " LIMIT " + gauge(1, strings.Repeat("▓", 20)),
		figure(" 34 %") + "  TIME " + cyan(strings.Repeat("▓", 7)+strings.Repeat("░", 13)) + " RESET AFTER " + figure("3:17"),
		// The figures of a longer wait painted, the word between them not.
		" RESET AFTER " + figure("6") + " days " + figure("21:27"),
	} {
		if !strings.Contains(out, want) {
			t.Errorf("painted picture lacks %q:\n%q", want, out)
		}
	}
	if plain := renderLimits(sampleRows(), ctx, renderOpts{}); stripColour(out) != plain {
		t.Errorf("painting changed the text:\n%q\n%q", stripColour(out), plain)
	}
	if strings.Contains(renderLimits(sampleRows(), ctx, renderOpts{p: painter{on: true}}), "38;2;") {
		t.Error("a terminal without 24-bit colour got 24-bit codes")
	}
	// Spend behind the clock is green however full the gauge; a limit with no time gauge is measured
	// as if no time had passed.
	behind := []statusRow{{label: "5 HOURS", pct: 80, reset: "23:49", windowBar: usageBar(90), windowPct: 90, after: "0:30"},
		{label: "OPUS", pct: 6}}
	out = renderLimits(behind, contextReading{}, renderOpts{p: p})
	if !strings.Contains(out, "LIMIT "+gauge(0, strings.Repeat("▓", 16)+strings.Repeat("░", 4))) {
		t.Errorf("spend behind the clock is not green:\n%q", out)
	}
	if !strings.Contains(out, "OPUS  "+figure("  6 %")+" LIMIT "+gauge(0.3, "▓▓"+strings.Repeat("░", 18))) {
		t.Errorf("a limit without a time gauge is not measured against no time passed:\n%q", out)
	}
}

// The lead that turns a LIMIT gauge red is the config's: the 7-day gauge 8 points ahead is 0.4 of the
// way at the shipped 20, and fully red once the setting is 8.
func TestTheRedLeadIsConfigurable(t *testing.T) {
	prev := limitAheadRed
	defer func() { limitAheadRed = prev }()
	limitAheadRed = 8
	p := painter{on: true, rgb: true}
	out := renderLimits(sampleRows(), contextReading{}, renderOpts{p: p})
	want := "7 DAYS  \x1b[38;2;255;180;0m 13 %\x1b[0m LIMIT \x1b[" + p.gaugeColour(1) + "m▓▓▓" + strings.Repeat("░", 17) + "\x1b[0m"
	if !strings.Contains(out, want) {
		t.Errorf("a gauge as far ahead as the setting is not red:\n%q", out)
	}
}

func TestRenderLimitsFitsTheWidth(t *testing.T) {
	tail := "CACHE 0:59, SESSION: FRESH 1m CACHED 108m, ACCOUNT 5H: FRESH 20m CACHED 1.2b, 7D: FRESH 21m CACHED 1.2b"
	ctx := contextReading{tokens: 150000, window: 200000, ok: true}
	for _, width := range []int{400, 200, 130, 70, 60} {
		out := renderLimits(sampleRows(), ctx, renderOpts{p: painter{on: true}, width: width, contextTail: tail})
		lines := strings.Split(out, "\n")
		for _, l := range lines[1:] {
			if w := runeWidth(l); w > width {
				t.Errorf("width %d: a line is %d wide: %q", width, w, l)
			}
		}
		if !strings.HasSuffix(stripColour(lines[0]), "░ "+tail) {
			t.Errorf("width %d: the tokens are not on the CONTEXT line: %q", width, lines[0])
		}
		// The CONTEXT line shrinks its gauge to fit until the gauge is as short as it gets.
		if w := runeWidth(lines[0]); w > width && strings.Count(lines[0], "▓")+strings.Count(lines[0], "░") > minContextCells {
			t.Errorf("width %d: the CONTEXT line is %d wide with room to shrink", width, w)
		}
		// The limit lines come in pairs after CONTEXT, and every pair's bottom line opens with the
		// marker that stops Claude Code trimming it out of line.
		if (len(lines)-1)%2 != 0 {
			t.Fatalf("width %d: %d limit lines, not pairs", width, len(lines)-1)
		}
		for i := 2; i < len(lines); i += 2 {
			if !strings.HasPrefix(stripColour(lines[i]), "-") {
				t.Errorf("width %d: bottom line %d does not open with the marker: %q", width, i, lines[i])
			}
		}
	}
	// Wide enough for all three: one pair. Narrow: one block per pair, with short gauges below 64.
	if n := strings.Count(renderLimits(sampleRows(), contextReading{}, renderOpts{width: 400}), "\n"); n != 1 {
		t.Errorf("at 400 columns the blocks took %d lines, want 2", n+1)
	}
	if n := strings.Count(renderLimits(sampleRows(), contextReading{}, renderOpts{width: 70}), "\n"); n != 5 {
		t.Errorf("at 70 columns the blocks took %d lines, want 6", n+1)
	}
	narrow := renderLimits(sampleRows(), contextReading{}, renderOpts{width: 50})
	if strings.Contains(narrow, strings.Repeat("░", 11)) || !strings.Contains(narrow, "LIMIT ▓▓▓▓▓▓░░░░") {
		t.Errorf("a narrow terminal did not get ten-cell gauges:\n%s", narrow)
	}
	// Unknown width: everything on one row, as before.
	if n := strings.Count(renderLimits(sampleRows(), contextReading{}, renderOpts{}), "\n"); n != 1 {
		t.Errorf("without a width the blocks took %d lines, want 2", n+1)
	}
	// No context reading yet: the tokens stand on the top line alone.
	if top := strings.Split(renderLimits(sampleRows(), contextReading{}, renderOpts{contextTail: tail}), "\n")[0]; top != tail {
		t.Errorf("top line without a context reading = %q", top)
	}
}

func stripColour(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			for i < len(s) && s[i] != 'm' {
				i++
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// --- what Claude Code hands the status line ----------------------------------------

func TestStatusContextUsesClaudeCodesWindow(t *testing.T) {
	in := decodeStatusInput(`{"context_window":{"total_input_tokens":150000,"context_window_size":200000,"current_usage":{"input_tokens":1}}}`)
	if r := in.statusContext(); !r.ok || r.tokens != 150000 || r.window != 200000 {
		t.Errorf("context = %+v, want 150000 of 200000", r)
	}
	// No reading yet: the window still comes from Claude Code, the tokens from the transcript.
	dir := t.TempDir()
	tp := filepath.Join(dir, "s.jsonl")
	writeLines(t, tp, `{"type":"assistant","message":{"usage":{"input_tokens":10,"cache_creation_input_tokens":20,"cache_read_input_tokens":30}}}`)
	in = decodeStatusInput(fmt.Sprintf(`{"transcript_path":%q,"context_window":{"total_input_tokens":0,"context_window_size":200000,"current_usage":null}}`, tp))
	if r := in.statusContext(); !r.ok || r.tokens != 60 || r.window != 200000 {
		t.Errorf("fallback context = %+v, want 60 of 200000", r)
	}
	// An older Claude Code that sends no window: the configured size, as before.
	in = decodeStatusInput(fmt.Sprintf(`{"transcript_path":%q}`, tp))
	if r := in.statusContext(); r.window != contextWindowTokens {
		t.Errorf("window without Claude Code's figure = %v, want the configured %v", r.window, contextWindowTokens)
	}
}

func TestLiveLimits(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	in := decodeStatusInput(fmt.Sprintf(`{"rate_limits":{"five_hour":{"used_percentage":23.5,"resets_at":%d},"seven_day":{"used_percentage":41.2,"resets_at":%d}}}`,
		now.Unix()+3600, now.Unix()-1))
	live := in.liveLimits(now)
	if live == nil || live.FiveHour == nil || live.FiveHour.Utilization != 23.5 {
		t.Fatalf("five_hour not taken: %+v", live)
	}
	if r, _ := parseTime(live.FiveHour.ResetsAt); !r.Equal(now.Add(time.Hour)) {
		t.Errorf("reset = %v", live.FiveHour.ResetsAt)
	}
	if live.SevenDay != nil {
		t.Error("a window whose reset has passed was taken")
	}
	if decodeStatusInput(`{}`).liveLimits(now) != nil || decodeStatusInput(`not json`).liveLimits(now) != nil {
		t.Error("no limits should give nil")
	}
}

// --- stale figures -------------------------------------------------------------------

// withCache points the hook at a temporary config dir holding no login, so nothing here can reach
// the network, and writes the usage cache with the given age.
func withCache(t *testing.T, age time.Duration, raw string) {
	t.Helper()
	dir := t.TempDir()
	saved := []string{claudeDir, cachePath, credPath, gateAsksPath, tokenCachePath}
	savedTTL := fetchTTLSec
	t.Cleanup(func() {
		claudeDir, cachePath, credPath, gateAsksPath, tokenCachePath = saved[0], saved[1], saved[2], saved[3], saved[4]
		fetchTTLSec = savedTTL
	})
	claudeDir = dir
	cachePath = filepath.Join(dir, "usage-limits-cache.json")
	credPath = filepath.Join(dir, "no-login.json")
	gateAsksPath = filepath.Join(dir, "gate.json")
	tokenCachePath = filepath.Join(dir, "tokens.json")
	fetchTTLSec = 60
	if raw == "" {
		return
	}
	b, _ := json.Marshal(cacheFile{FetchedAt: time.Now().Add(-age).Format(time.RFC3339Nano), Raw: json.RawMessage(raw)})
	if err := os.WriteFile(cachePath, b, 0o600); err != nil {
		t.Fatal(err)
	}
}

func usageDoc(s5, s7 float64, now time.Time) string {
	return fmt.Sprintf(`{"five_hour":{"utilization":%v,"resets_at":%q},"seven_day":{"utilization":%v,"resets_at":%q},`+
		`"limits":[{"kind":"weekly_scoped","percent":12,"resets_at":%q,"scope":{"model":{"display_name":"Fable"}}}]}`,
		s5, now.Add(2*time.Hour).UTC().Format(time.RFC3339), s7, now.Add(100*time.Hour).UTC().Format(time.RFC3339),
		now.Add(100*time.Hour).UTC().Format(time.RFC3339))
}

func runtimeIsDarwin() bool { return runtime.GOOS == "darwin" }

func TestStaleFiguresAreMarkedAndClaudeCodesStandIn(t *testing.T) {
	if runtimeIsDarwin() {
		t.Skip("reads the login Keychain on macOS")
	}
	now := time.Now()
	withCache(t, time.Hour, usageDoc(40, 10, now))
	before := getCacheObject()

	live := &usageRaw{FiveHour: &usageWindow{Utilization: 71, ResetsAt: now.Add(time.Hour).UTC().Format(time.RFC3339)}}
	st := getState(argSet{}, "claude-fable-5-1", live)
	if st == nil || !st.stale {
		t.Fatalf("state = %+v", st)
	}
	if st.s5 != 71 || st.s7 != 10 || st.zone5 != "YELLOW" {
		t.Errorf("s5 %v s7 %v zone5 %v: want Claude Code's 71 and the cached 10", st.s5, st.s7, st.zone5)
	}
	if st.activeBucket == nil || st.activeBucket.Percent != 12 {
		t.Error("the per-model bucket from the cache was lost")
	}
	after := getCacheObject()
	if after == nil || after.FetchedAt != before.FetchedAt || !bytes.Equal(after.Raw, before.Raw) || len(after.Samples) != 0 {
		t.Errorf("a figure from Claude Code reached the cache the gate reads: %+v", after)
	}
	if after.TriedAt == "" {
		t.Error("the failed attempt was not written down")
	}

	// Without Claude Code's figures — the injected block, the gate — the cached ones are used, and
	// the model is told they are stale and since when.
	st = getState(argSet{}, "", nil)
	if st == nil || !st.stale || st.s5 != 40 {
		t.Fatalf("state without live figures = %+v", st)
	}
	since := fmtHHmm(st.fetchedAt)
	if st.fetchedAt.Local().Day() != now.Local().Day() {
		since = strings.ToUpper(fmtDddHHmm(st.fetchedAt))
	}
	want := `<stale since="` + since + `">the usage API has not answered since then; every figure here is as of then</stale>`
	if got := staleElement(st, now); got != want {
		t.Errorf("stale element = %q, want %q", got, want)
	}
}

// The API is asked at most once per fetch period, answered or not: a rate-limited endpoint is not
// asked again by every refresh of every open status line.
func TestAFailedAttemptIsNotRepeatedWithinTheFetchPeriod(t *testing.T) {
	if runtimeIsDarwin() {
		t.Skip("reads the login Keychain on macOS")
	}
	withCache(t, time.Hour, usageDoc(40, 10, time.Now()))
	getState(argSet{}, "", nil)
	first := getCacheObject().TriedAt
	if first == "" {
		t.Fatal("the attempt was not written down")
	}
	getState(argSet{}, "", nil)
	if again := getCacheObject().TriedAt; again != first {
		t.Errorf("asked again within the fetch period: %s then %s", first, again)
	}
	// Once the period has passed, it is asked again.
	c := getCacheObject()
	c.TriedAt = time.Now().Add(-2 * time.Minute).Format(time.RFC3339Nano)
	saveCache(*c)
	getState(argSet{}, "", nil)
	if again := getCacheObject().TriedAt; again == c.TriedAt {
		t.Error("not asked again after the fetch period")
	}
}

func TestFreshFiguresAreNotOverridden(t *testing.T) {
	if runtimeIsDarwin() {
		t.Skip("reads the login Keychain on macOS")
	}
	now := time.Now()
	withCache(t, 10*time.Second, usageDoc(40, 10, now))
	live := &usageRaw{FiveHour: &usageWindow{Utilization: 71, ResetsAt: now.Add(time.Hour).UTC().Format(time.RFC3339)}}
	st := getState(argSet{}, "", live)
	if st == nil || st.stale || st.s5 != 40 {
		t.Fatalf("fresh state = %+v", st)
	}
	// One failed refresh is not an outage: a minute and a half old is not stale at a one-minute TTL.
	withCache(t, 90*time.Second, usageDoc(40, 10, now))
	if st := getState(argSet{}, "", nil); st == nil || st.stale {
		t.Error("a single missed refresh was called stale")
	}
}

func TestNoCacheAndNoAPIFallsBackToClaudeCode(t *testing.T) {
	if runtimeIsDarwin() {
		t.Skip("reads the login Keychain on macOS")
	}
	now := time.Now()
	withCache(t, 0, "")
	if getState(argSet{}, "", nil) != nil {
		t.Fatal("no data at all should give no state")
	}
	live := &usageRaw{FiveHour: &usageWindow{Utilization: 20, ResetsAt: now.Add(time.Hour).UTC().Format(time.RFC3339)}}
	st := getState(argSet{}, "", live)
	if st == nil || st.s5 != 20 || !st.fetchedAt.IsZero() {
		t.Fatalf("state = %+v", st)
	}
	if c := getCacheObject(); c != nil && (isJSONValue(c.Raw) || c.FetchedAt != "" || len(c.Samples) != 0) {
		t.Errorf("Claude Code's figures reached the cache: %+v", c)
	}
}

// --- tokens ------------------------------------------------------------------------

func writeLines(t *testing.T, path string, lines ...string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	for _, l := range lines {
		if _, err := f.WriteString(l + "\n"); err != nil {
			t.Fatal(err)
		}
	}
}

func replyLine(id string, in, out, write, read int) string {
	return fmt.Sprintf(`{"type":"assistant","requestId":"req_%s","message":{"id":"msg_%s","usage":{"input_tokens":%d,"output_tokens":%d,"cache_creation_input_tokens":%d,"cache_read_input_tokens":%d}}}`,
		id, id, in, out, write, read)
}

// replyAt is a reply line written at a given moment.
func replyAt(id string, at time.Time, in, out, write, read int) string {
	return fmt.Sprintf(`{"type":"assistant","timestamp":%q,"requestId":"req_%s","message":{"id":"msg_%s","usage":{"input_tokens":%d,"output_tokens":%d,"cache_creation_input_tokens":%d,"cache_read_input_tokens":%d}}}`,
		at.UTC().Format(time.RFC3339Nano), id, id, in, out, write, read)
}

func TestTranscriptReplyCountsItsLastLine(t *testing.T) {
	p := filepath.Join(t.TempDir(), "s.jsonl")
	// One reply written as three lines whose output count grows, a user line, and the next reply.
	writeLines(t, p,
		replyLine("a", 2, 7, 100, 1000),
		replyLine("a", 2, 7, 100, 1000),
		replyLine("a", 2, 145, 100, 1000),
		`{"type":"user","message":{"content":"ok"}}`,
		replyLine("b", 1, 50, 0, 1100))
	f := &fileTokens{}
	info, _ := os.Stat(p)
	if !f.read(p, info.Size(), time.Now().Add(time.Minute)) {
		t.Fatal("did not reach the end")
	}
	got := f.total()
	if got.Output != 195 || got.Input != 3 || got.Write != 100 || got.Read != 2100 {
		t.Errorf("sums = %+v, want output 145+50", got)
	}
	// The second reply is still open — its count may yet grow — and is counted from there.
	if f.OpenKey == "" || f.Sum.Output != 145 {
		t.Errorf("open reply not held apart: %+v", f)
	}
	// It grows, and a half-written line is left for later.
	writeLines(t, p, replyLine("b", 1, 400, 0, 1100))
	raw, _ := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0o600)
	_, _ = raw.WriteString(`{"type":"assistant","message":{"id":"msg_c",`)
	raw.Close()
	info, _ = os.Stat(p)
	f.read(p, info.Size(), time.Now().Add(time.Minute))
	if got := f.total(); got.Output != 545 {
		t.Errorf("after growth output = %v, want 145+400", got.Output)
	}
	if f.Offset == info.Size() {
		t.Error("a half-written line was consumed")
	}
	// A file rewritten shorter than what was read is read again from the start.
	if err := os.WriteFile(p, []byte(replyLine("z", 1, 1, 1, 1)+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	info, _ = os.Stat(p)
	f.read(p, info.Size(), time.Now().Add(time.Minute))
	if got := f.total().total(); got != 4 {
		t.Errorf("rewritten file total = %v, want 4", got)
	}
	// Out of time: nothing is read, and it says so.
	if (&fileTokens{}).read(p, info.Size(), time.Now().Add(-time.Second)) {
		t.Error("a read past its deadline claimed to have finished")
	}
}

// The session is its own transcript and its subagents'; the account is every transcript written to
// this week, counted in the five-hour window and in the week.
func TestReadTokens(t *testing.T) {
	withCache(t, 0, "")
	now := time.Now()
	fiveStart, weekFrom := now.Add(-3*time.Hour), now.Add(-3*24*time.Hour)
	project := filepath.Join(claudeDir, "projects", "p")
	session := filepath.Join(project, "s1.jsonl")
	sub := filepath.Join(project, "s1", "subagents", "agent-x.jsonl")
	other := filepath.Join(project, "s0.jsonl")
	lastWeek := filepath.Join(project, "s9.jsonl")
	if err := os.MkdirAll(filepath.Dir(sub), 0o700); err != nil {
		t.Fatal(err)
	}
	writeLines(t, session, replyAt("a", now.Add(-time.Hour), 1, 2, 3, 4))
	writeLines(t, sub, replyAt("b", now.Add(-time.Hour), 10, 20, 30, 40))
	// Another session of the account: one reply inside the five hours, one earlier in the week, and
	// one from before the week began, which counts nowhere.
	writeLines(t, other,
		replyAt("c", now.Add(-4*24*time.Hour), 5000, 0, 0, 0),
		replyAt("d", now.Add(-2*24*time.Hour), 100, 200, 300, 400),
		replyAt("e", now.Add(-time.Hour), 1000, 0, 0, 2000))
	// A session last written before the week began, left in the cache by an earlier run.
	writeLines(t, lastWeek, replyAt("f", now.Add(-10*24*time.Hour), 7, 7, 7, 7))
	old := now.Add(-10 * 24 * time.Hour)
	if err := os.Chtimes(lastWeek, old, old); err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(tokenCache{Files: map[string]*fileTokens{lastWeek: {Size: 1}}})
	if err := os.WriteFile(tokenCachePath, b, 0o600); err != nil {
		t.Fatal(err)
	}

	st := readTokens(session, fiveStart, weekFrom, now)
	if !st.sessionWhole || st.session.total() != 110 {
		t.Errorf("session = %+v, want 110 with its subagent", st.session)
	}
	// Five hours: the session's 110 and the other session's latest reply, 3000.
	if st.fiveHour.total() != 3110 || st.fiveHour.fresh() != 1066 || st.fiveHour.Read != 2044 {
		t.Errorf("five hours = %+v, want 3110", st.fiveHour)
	}
	// The week adds the reply from two days ago, 1000; the one from four days ago is before it.
	if st.weekly.total() != 4110 || !st.accountWhole {
		t.Errorf("week = %+v (whole %v), want 4110", st.weekly, st.accountWhole)
	}
	// The session named by a path that differs from the walk's only in its spelling — a redundant
	// "./", and on Windows forward slashes — is the same file, counted once.
	sep := string(filepath.Separator)
	alias := strings.ReplaceAll(project+sep+"."+sep+"s1.jsonl", `\`, "/")
	if runtime.GOOS != "windows" {
		alias = project + sep + "." + sep + "s1.jsonl"
	}
	if again := readTokens(alias, fiveStart, weekFrom, now); again.session.total() != 110 || again.fiveHour.total() != 3110 || again.weekly.total() != 4110 {
		t.Errorf("a differently spelled path counted the session twice: session %v, five hours %v, week %v",
			again.session.total(), again.fiveHour.total(), again.weekly.total())
	}
	var c tokenCache
	b, _ = os.ReadFile(tokenCachePath)
	if json.Unmarshal(b, &c) != nil || c.Files[fileKey(lastWeek)] != nil || c.Files[fileKey(other)] == nil {
		t.Errorf("cache after the run = %s", b)
	}
	for bucket := range c.Files[fileKey(other)].Buckets {
		if bucket+tokenBucketSec <= weekFrom.Unix() {
			t.Errorf("a bucket from before the week was kept: %d", bucket)
		}
	}

	want := "SESSION: FRESH 66 CACHED 44, ACCOUNT 5H: FRESH 1k CACHED 2k, 7D: FRESH 2k CACHED 2k"
	if got := tokensText(session, fiveStart, weekFrom, now); got != want {
		t.Errorf("tokens text = %q, want %q", got, want)
	}
	// No windows known: the session alone.
	if got := tokensText(session, time.Time{}, time.Time{}, now); got != "SESSION: FRESH 66 CACHED 44" {
		t.Errorf("tokens text without windows = %q", got)
	}
	// The same counts as the model is handed them.
	if got := tokensElement(st); got != `<tokens session="fresh 66, cached 44" account_5h="fresh 1k, cached 2k" account_7d="fresh 2k, cached 2k"/>` {
		t.Errorf("tokens element = %q", got)
	}
	if _, ok := sessionTokenStats("", fiveStart, weekFrom, now); ok || tokensText("", fiveStart, weekFrom, now) != "" {
		t.Error("tokens without a session")
	}
}

// --- the whole status line ----------------------------------------------------------

func TestRenderStatusline(t *testing.T) {
	if runtimeIsDarwin() {
		t.Skip("reads the login Keychain on macOS")
	}
	now := time.Now()
	withCache(t, 5*time.Second, usageDoc(55, 10, now))
	tp := filepath.Join(claudeDir, "projects", "p", "s.jsonl")
	if err := os.MkdirAll(filepath.Dir(tp), 0o700); err != nil {
		t.Fatal(err)
	}
	writeLines(t, tp, replyAt("a", now.Add(-time.Minute), 5, 1000, 20000, 130000))
	input := fmt.Sprintf(`{"transcript_path":%q,"model":{"id":"claude-fable-5-1","display_name":"Fable"},`+
		`"context_window":{"total_input_tokens":150005,"context_window_size":200000,"current_usage":{}},`+
		`"prompt_cache":{"warm":true,"caching_observed":true,"ttl":"1h","expires_at":%d}}`, tp, now.Unix()+3600)

	t.Setenv("COLUMNS", "")
	t.Setenv("NO_COLOR", "1")
	out := renderStatusline(argSet{}, input, nil, now)
	lines := strings.Split(out, "\n")
	wantTail := "░ CACHE 1:00, SESSION: FRESH 21k CACHED 130k, ACCOUNT 5H: FRESH 21k CACHED 130k, 7D: FRESH 21k CACHED 130k"
	if !strings.HasPrefix(lines[0], "CONTEXT  150k/200k   75 %") || !strings.HasSuffix(lines[0], wantTail) {
		t.Errorf("top line = %q", lines[0])
	}
	if len(lines) != 3 || !strings.HasPrefix(lines[1], "5 HOURS") || !strings.Contains(lines[1], "FABLE   12 % LIMIT") {
		t.Errorf("want CONTEXT and one pair of limit lines, nothing else:\n%s", out)
	}
	if strings.Contains(out, "\x1b") {
		t.Errorf("NO_COLOR not honoured:\n%s", out)
	}
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM_PROGRAM", "")
	if !strings.Contains(renderStatusline(argSet{}, input, nil, now), "\x1b[38;2;") {
		t.Error("the status line is not painted in 24-bit colour")
	}
	t.Setenv("TERM_PROGRAM", "Apple_Terminal")
	if out := renderStatusline(argSet{}, input, nil, now); !strings.Contains(out, "\x1b[38;5;") || strings.Contains(out, "38;2;") {
		t.Error("Apple's Terminal did not get the 256-colour palette")
	}

	// A silent API: Claude Code's own 5-hour figure stands in, without a word on screen about it.
	t.Setenv("NO_COLOR", "1")
	stale, _ := json.Marshal(cacheFile{FetchedAt: now.Add(-time.Hour).Format(time.RFC3339Nano), Raw: json.RawMessage(usageDoc(55, 10, now))})
	if err := os.WriteFile(cachePath, stale, 0o600); err != nil {
		t.Fatal(err)
	}
	withLive := strings.TrimSuffix(input, "}") + fmt.Sprintf(`,"rate_limits":{"five_hour":{"used_percentage":71,"resets_at":%d}}}`, now.Unix()+3600)
	out = renderStatusline(argSet{}, withLive, nil, now)
	if !strings.Contains(out, "5 HOURS   71 % LIMIT") || strings.Contains(out, "DATA") || len(strings.Split(out, "\n")) != 3 {
		t.Errorf("stale figures not replaced quietly:\n%s", out)
	}
}

// A write leaves no temporary file behind, and sweeps up what an interrupted one left — but not a
// write another run may still have in progress.
func TestWriteFileAtomicSweepsLeftovers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")
	stale, fresh := path+".111.tmp", path+".222.tmp"
	for _, p := range []string{stale, fresh} {
		if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	old := time.Now().Add(-5 * time.Minute)
	if err := os.Chtimes(stale, old, old); err != nil {
		t.Fatal(err)
	}
	if err := writeFileAtomic(path, []byte(`{"a":1}`)); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(path); string(b) != `{"a":1}` {
		t.Errorf("written %q", b)
	}
	if _, err := os.Stat(stale); err == nil {
		t.Error("an interrupted write's leftover was kept")
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Error("a write that may still be in progress was swept away")
	}
	if left, _ := filepath.Glob(path + ".*.tmp"); len(left) != 1 {
		t.Errorf("temporary files left: %v", left)
	}
}

// The cache part says how long the prompt cache stays warm, or that it has gone cold, and is left
// out when Claude Code reports no cache at all.
func TestCacheText(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	cache := func(ttl string, left int64) string {
		return fmt.Sprintf(`{"prompt_cache":{"warm":true,"caching_observed":true,"ttl":%q,"expires_at":%d}}`, ttl, now.Unix()+left)
	}
	for raw, want := range map[string]string{
		cache("1h", 52*60): "CACHE 0:52",
		cache("1h", -60):   "CACHE COLD",
		`{"prompt_cache":{"warm":false,"caching_observed":true,"expires_at":null}}`: "CACHE COLD",
		`{"prompt_cache":{"warm":true,"caching_observed":true}}`:                    "CACHE WARM",
		`{"prompt_cache":{"warm":false,"caching_observed":false}}`:                  "",
		`{}`: "",
	} {
		if got := decodeStatusInput(raw).cacheText(now, painter{}); got != want {
			t.Errorf("%s: %q, want %q", raw, got, want)
		}
	}
	// The time left is green with the whole lifetime ahead and turns red as it runs out, on the
	// lifetime the cache was written with; a cold cache is red.
	p := painter{on: true, rgb: true}
	paint := func(at float64, s string) string { return "CACHE \x1b[" + p.gaugeColour(at) + "m" + s + "\x1b[0m" }
	for _, c := range []struct {
		raw, want string
	}{
		{cache("1h", 60*60), paint(0, "1:00")},
		{cache("1h", 30*60), paint(0.5, "0:30")},
		{cache("5m", 5*60), paint(0, "0:05")},
		{cache("", 15*60), paint(0.75, "0:15")}, // an hour where Claude Code does not say
		{cache("1h", -60), paint(1, "COLD")},
	} {
		if got := decodeStatusInput(c.raw).cacheText(now, p); got != c.want {
			t.Errorf("%s: %q, want %q", c.raw, got, c.want)
		}
	}
	// With no session to count, the cache stands alone after the label; with neither, nothing.
	in := decodeStatusInput(cache("1h", 600))
	if got := in.contextTail(now, painter{}, time.Time{}, time.Time{}); got != "CACHE 0:10" {
		t.Errorf("tail = %q", got)
	}
	if got := decodeStatusInput(`{}`).contextTail(now, painter{}, time.Time{}, time.Time{}); got != "" {
		t.Errorf("tail with nothing to say = %q", got)
	}
}

func TestShortTokens(t *testing.T) {
	for n, want := range map[float64]string{510_000: "510k", 1e6: "1m", 1.4e6: "1m", 1.6e6: "2m", 12e6: "12m", 24.1e6: "24m",
		647.97e6: "648m", 1e9: "1b", 1.24e9: "1.2b", 29.94e9: "29.9b", 950: "950", 5_600: "6k"} {
		if got := shortTokens(n); got != want {
			t.Errorf("shortTokens(%v) = %q, want %q", n, got, want)
		}
	}
}
