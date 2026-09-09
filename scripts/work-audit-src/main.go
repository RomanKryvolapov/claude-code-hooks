// work-audit — appends the sequence of work to a per-author audit log.
//
// Registered in .claude/settings.json (async) for the events below. It records WHAT happened, in
// order, with no timestamps (append order is the sequence) and no token counts. Example:
//
//	session start: id=71ba66a1-… source=startup model=claude-opus-4-8
//	prompt:
//	<the user's request>
//	subagent start: Explore
//	subagent done: Explore
//	question: Old bugs
//	<the question Claude asked>
//	user chose: Delete them
//	prompt (typed while the turn was running):
//	<a message queued mid-turn: it never reaches UserPromptSubmit, so it is recovered at Stop>
//	turn:
//	model=claude-opus-4-8 effort=max duration=2m22s subagents=1 tools=18: Edit×7, Read×5, …
//	files:
//	~ scripts/work-audit-src/main.go (+45/-12) +[50-61,200-205] -[120-124]
//	+ management/tasks/TEMPLATE.md (+30)
//	answer:
//	<Claude Code's final "what was done" answer, verbatim>
//	compaction: trigger=auto
//	error: rate_limit — 429 too many requests
//	session end: reason=clear duration=1h2m
//
// Two labels that are easy to confuse, so they are kept apart: `user chose:` is the option the user
// picked in a question Claude asked, `answer:` is Claude's own final reply for the turn.
//
// Each event writes one entry, and every entry is a single append — so entries never interleave,
// and the file reads in the order the events reached the hook. The one thing that arrives late is a
// message typed mid-turn: it is queued rather than submitted, so it surfaces only at Stop and is
// written then, on its own, immediately before the turn it interrupted.
//
// Per-turn stats are derived at Stop by parsing the session transcript for the
// main-thread work since the last typed user prompt; sidechain (subagent) lines
// are excluded from the turn's tool count. Session duration is the span of the
// whole transcript, emitted at SessionEnd.
//
// Subagents are counted from the SubagentStart events rather than from the tool calls: a Workflow
// fan-out is a single `Workflow` tool_use that spawns dozens of agents. The Task/Agent tool count
// stays as a floor, so a missed event can only undercount, never erase. Those fan-out agents get no
// start/done line of their own — dozens of them would bury the turn — only the count in `subagents=`.
// Agents started by hand (Task / Agent) keep their start/done lines.
//
// Logs are filed management/logs/<author-slug>/<date>/<session8>.log — one folder
// per developer, one folder per day inside it, one file per session — so parallel
// sessions cannot merge-conflict (see management/logs/** merge=union). Cross-platform Go,
// one binary per OS (see build.sh). Must never break the session: any error is
// swallowed and the process exits 0.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type hookEvent struct {
	HookEventName  string          `json:"hook_event_name"`
	SessionID      string          `json:"session_id"`
	Prompt         string          `json:"prompt"`
	Source         string          `json:"source"`
	Model          string          `json:"model"`
	TranscriptPath string          `json:"transcript_path"`
	ToolName       string          `json:"tool_name"`
	ToolInput      json.RawMessage `json:"tool_input"`
	ToolResponse   json.RawMessage `json:"tool_response"`
	EndReason      string          `json:"reason"`
	AgentType      string          `json:"agent_type"`
	AgentID        string          `json:"agent_id"`
	StopReason     string          `json:"stop_reason"`
	LastAssistant  string          `json:"last_assistant_message"`
	ErrorType      string          `json:"error"`
	ErrorDetails   json.RawMessage `json:"error_details"`
	Compaction     string          `json:"trigger"`
}

// tLine is the subset of a transcript JSONL record we care about.
type tLine struct {
	Type         string `json:"type"`
	Timestamp    string `json:"timestamp"`
	Effort       string `json:"effort"`
	PromptSource string `json:"promptSource"`
	IsSidechain  bool   `json:"isSidechain"`
	// A message typed while the turn was still running is queued rather than submitted, so it never
	// reaches UserPromptSubmit — it shows up here as {"type":"queue-operation","operation":"enqueue"}
	// with the text at the top level. Without these the audit misses half of what the user asked.
	Operation string `json:"operation"`
	Content   string `json:"content"`
	Message   struct {
		Role    string          `json:"role"`
		Model   string          `json:"model"`
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

var (
	wsRe       = regexp.MustCompile(`\s+`)
	nonAlnumRe = regexp.MustCompile(`[^a-z0-9]+`)
	edgeDashRe = regexp.MustCompile(`^-+|-+$`)
)

func trunc(s string, max int) string {
	s = strings.TrimSpace(wsRe.ReplaceAllString(s, " "))
	r := []rune(s)
	if len(r) > max {
		return string(r[:max]) + "..."
	}
	return s
}

func slugify(s string) string {
	s = nonAlnumRe.ReplaceAllString(strings.ToLower(s), "-")
	return edgeDashRe.ReplaceAllString(s, "")
}

// promptSep visually marks the start of each new prompt / turn in the audit log.
const promptSep = "════════════════════════════════════════════════════════════"

// cleanBlock keeps a string's non-empty lines (dropping blank lines and trailing
// whitespace, preserving indentation) and rune-caps it. Used for the prompt and the
// final report — multi-line text laid out under a label, where blank lines are noise.
func cleanBlock(s string, max int) string {
	var kept []string
	for _, ln := range strings.Split(strings.TrimSpace(s), "\n") {
		if strings.TrimSpace(ln) != "" {
			kept = append(kept, strings.TrimRight(ln, " \t\r"))
		}
	}
	out := strings.Join(kept, "\n")
	r := []rune(out)
	if len(r) > max {
		return string(r[:max]) + "… (truncated)"
	}
	return out
}

// shortID returns the first 8 alphanumeric characters of a session id (the
// leading group of a UUID), used to key one audit file per session.
func shortID(id string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(id) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			if b.Len() >= 8 {
				break
			}
		}
	}
	return b.String()
}

func firstNonEmpty(xs ...string) string {
	for _, x := range xs {
		if x != "" {
			return x
		}
	}
	return ""
}

// fmtDur renders seconds as "58s", "2m22s", "43m", or "1h2m".
func fmtDur(sec int) string {
	if sec < 90 {
		return fmt.Sprintf("%ds", sec)
	}
	m := sec / 60
	if m < 90 {
		if s := sec % 60; s != 0 {
			return fmt.Sprintf("%dm%ds", m, s)
		}
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%dh%dm", m/60, m%60)
}

func gitAuthor(repoRoot string) string {
	out, err := runGit(repoRoot, "config", "user.name")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

func runGit(repoRoot string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	return string(out), err
}

func countLines(path string) int {
	b, err := os.ReadFile(path)
	if err != nil || len(b) == 0 {
		return 0
	}
	n := strings.Count(string(b), "\n")
	if !strings.HasSuffix(string(b), "\n") {
		n++
	}
	return n
}

type fileDiff struct {
	added, removed []int // added = new-file line numbers; removed = old-file line numbers
	isNew, isDel   bool
}

// parseDiff walks `git diff -U0` output and records, per file, the exact line
// numbers added (in the new file) and removed (in the old file).
func parseDiff(diff string) map[string]*fileDiff {
	res := map[string]*fileDiff{}
	var cur *fileDiff
	var oldLine, newLine int
	inHunk := false
	for _, ln := range strings.Split(diff, "\n") {
		switch {
		case strings.HasPrefix(ln, "diff --git "):
			cur, inHunk = &fileDiff{}, false
			if i := strings.Index(ln, " b/"); i >= 0 {
				res[filepath.ToSlash(strings.TrimSpace(ln[i+3:]))] = cur
			}
		case cur == nil:
			// pre-file noise
		case strings.HasPrefix(ln, "new file mode"):
			cur.isNew = true
		case strings.HasPrefix(ln, "deleted file mode"):
			cur.isDel = true
		case strings.HasPrefix(ln, "@@"):
			oldLine, newLine = hunkStarts(ln)
			inHunk = true
		case !inHunk:
			// index / --- / +++ headers
		case strings.HasPrefix(ln, "+"):
			cur.added = append(cur.added, newLine)
			newLine++
		case strings.HasPrefix(ln, "-"):
			cur.removed = append(cur.removed, oldLine)
			oldLine++
		case strings.HasPrefix(ln, "\\"): // "\ No newline at end of file"
		default:
			oldLine++
			newLine++
		}
	}
	return res
}

// hunkStarts parses "@@ -oldStart[,c] +newStart[,c] @@ [context]" → oldStart,
// newStart. Only the part between the two "@@" is parsed, so a lone +/- in the
// trailing context text cannot corrupt the numbers.
func hunkStarts(ln string) (int, int) {
	if i := strings.Index(ln, "@@"); i >= 0 {
		if j := strings.Index(ln[i+2:], "@@"); j >= 0 {
			ln = ln[i+2 : i+2+j]
		}
	}
	var oldS, newS int
	for _, tok := range strings.Fields(ln) {
		if strings.HasPrefix(tok, "-") {
			oldS = atoiHead(tok[1:])
		} else if strings.HasPrefix(tok, "+") {
			newS = atoiHead(tok[1:])
		}
	}
	return oldS, newS
}

func atoiHead(s string) int {
	if i := strings.IndexByte(s, ','); i >= 0 {
		s = s[:i]
	}
	n, _ := strconv.Atoi(s)
	return n
}

// coalesce turns an ascending list of line numbers into "50-61,80,120-124",
// capped so a huge diff cannot blow up the line.
func coalesce(nums []int) string {
	if len(nums) == 0 {
		return ""
	}
	var parts []string
	start, prev := nums[0], nums[0]
	flush := func() {
		if start == prev {
			parts = append(parts, strconv.Itoa(start))
		} else {
			parts = append(parts, strconv.Itoa(start)+"-"+strconv.Itoa(prev))
		}
	}
	for _, n := range nums[1:] {
		if n == prev || n == prev+1 {
			prev = n
			continue
		}
		flush()
		start, prev = n, n
	}
	flush()
	if len(parts) > 20 {
		parts = append(parts[:20], "…")
	}
	return strings.Join(parts, ",")
}

// fileStats annotates each touched file (repo-relative) with a create/modify/delete
// marker, added/removed line counts, and — for modified files — the exact line
// ranges added (new file) and removed (old file). Best-effort: on any git error a
// file is left un-annotated and the caller falls back to the bare path. Examples:
//
//   - docs/new.md (+40)
//     ~ scripts/main.go (+12/-5) +[50-55,80] -[120-124]
//   - old.txt (-30)
func fileStats(repoRoot string, files []string) map[string]string {
	res := map[string]string{}
	if repoRoot == "" || len(files) == 0 {
		return res
	}
	// A path relPath could not make repo-relative (a scratchpad temp file, another checkout) makes
	// git exit 128 — "is outside repository" — and that one path would otherwise cost the whole
	// turn its markers and line counts. Drop them here: they fall back to a bare path on their own.
	inRepo := make([]string, 0, len(files))
	for _, f := range files {
		// filepath.IsAbs alone is not enough on Windows: a POSIX-rooted path like /mnt/d/... — what
		// an IDE integration hands over — carries no volume name, so Go reads it as relative while
		// git still rejects it. Check the leading separator too.
		if filepath.IsAbs(f) || strings.HasPrefix(f, "/") || strings.HasPrefix(f, `\`) ||
			strings.HasPrefix(f, "../") || strings.HasPrefix(f, `..\`) {
			continue
		}
		inRepo = append(inRepo, f)
	}
	if len(inRepo) == 0 {
		return res
	}
	// If git is missing or this is not a repo, the status call errors: bail out with
	// no annotations so the caller falls back to bare file paths — the rest of the
	// turn (prompt, stats, report) is unaffected.
	statusOut, err := runGit(repoRoot, append([]string{"-c", "core.quotepath=false", "status", "--porcelain", "--"}, inRepo...)...)
	if err != nil {
		return res
	}
	status := map[string]byte{}
	for _, ln := range strings.Split(statusOut, "\n") {
		if len(ln) < 4 {
			continue
		}
		code, p := ln[:2], strings.TrimSpace(ln[3:])
		if k := strings.Index(p, " -> "); k >= 0 { // rename: keep the new path
			p = p[k+4:]
		}
		p = filepath.ToSlash(strings.Trim(p, `"`))
		switch {
		case strings.ContainsAny(code, "?A"):
			status[p] = 'A'
		case strings.Contains(code, "D"):
			status[p] = 'D'
		default:
			status[p] = 'M'
		}
	}
	var diffs map[string]*fileDiff
	if out, err := runGit(repoRoot, append([]string{"-c", "core.quotepath=false", "diff", "-U0", "HEAD", "--"}, inRepo...)...); err == nil {
		diffs = parseDiff(out)
	}
	for _, p := range inRepo {
		code := status[p]
		if code == 0 {
			code = 'M'
		}
		d := diffs[p]
		add, del := 0, 0
		if d != nil {
			add, del = len(d.added), len(d.removed)
		}
		switch code {
		case 'A':
			if add == 0 { // untracked new file: not visible to `diff HEAD`
				add = countLines(filepath.Join(repoRoot, filepath.FromSlash(p)))
			}
			res[p] = fmt.Sprintf("+ %s (+%d)", p, add)
		case 'D':
			res[p] = fmt.Sprintf("- %s (-%d)", p, del)
		default:
			line := fmt.Sprintf("~ %s (+%d/-%d)", p, add, del)
			if d != nil {
				if r := coalesce(d.added); r != "" {
					line += " +[" + r + "]"
				}
				if r := coalesce(d.removed); r != "" {
					line += " -[" + r + "]"
				}
			}
			res[p] = line
		}
	}
	return res
}

func main() {
	defer func() {
		_ = recover()
		os.Exit(0)
	}()
	run()
}

func run() {
	raw, err := io.ReadAll(os.Stdin)
	if err != nil || strings.TrimSpace(string(raw)) == "" {
		return
	}
	var e hookEvent
	if err := json.Unmarshal(raw, &e); err != nil {
		return
	}

	repoRoot := os.Getenv("CLAUDE_PROJECT_DIR")
	if repoRoot == "" {
		if cwd, err := os.Getwd(); err == nil {
			repoRoot = cwd
		}
	}
	slug := slugify(gitAuthor(repoRoot))
	if slug == "" {
		slug = "unknown"
	}
	session8 := shortID(e.SessionID)

	if e.HookEventName == "SubagentStart" {
		noteSubagent(session8)
	}

	// Anything the user typed mid-turn goes in first, one entry each, so every message gets its own
	// timestamp and they sit ahead of the turn they interrupted instead of inside it.
	if e.HookEventName == "Stop" || e.HookEventName == "StopFailure" {
		for _, queued := range takeQueuedPrompts(e.TranscriptPath, session8) {
			appendAudit(repoRoot, slug, session8, queued)
		}
	}

	if out := format(e, session8); strings.TrimSpace(out) != "" {
		appendAudit(repoRoot, slug, session8, out)
	}
}

// isSystemNotice reports whether a "prompt" is really a harness notification (a background task
// reporting back), not something the user typed.
func isSystemNotice(prompt string) bool {
	head := strings.TrimSpace(prompt)
	return strings.HasPrefix(head, "<task-notification>") ||
		strings.HasPrefix(head, "[SYSTEM NOTIFICATION")
}

// isFanOutAgent reports whether an agent type comes from a Workflow fan-out. One Workflow tool call
// spawns dozens of them, so a start/done line per agent buries the turn it belongs to; they are
// counted in the turn line instead (see subagentTally / turnStats).
func isFanOutAgent(agentType string) bool {
	return agentType == "workflow-subagent"
}

// Subagents are counted per turn, not per tool call: a Workflow fan-out is one `Workflow` tool_use
// in the transcript but N agents, so the transcript alone cannot answer "how many ran". Every
// SubagentStart appends one byte to a per-session tally next to the OS temp dir, and the turn line
// reads and clears it. Append-only, so parallel agents starting at once cannot lose a count.
//
// This tally and the queued-message marker below are throwaway state in the OS temp dir. A temp
// sweep mid-session costs a turn its subagent count and can re-emit a queued message that was
// already logged; the log still records what happened, only the counting restarts. Nothing that
// belongs in the repository is kept there.
func tallyPath(session8 string) string {
	if session8 == "" {
		session8 = "session"
	}
	return filepath.Join(os.TempDir(), "claude-work-audit", session8+".agents")
}

func noteSubagent(session8 string) {
	p := tallyPath(session8)
	if os.MkdirAll(filepath.Dir(p), 0o700) != nil {
		return
	}
	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = f.WriteString("x")
}

// takeQueuedPrompts returns the messages the user typed while a turn was running, as ready-to-write
// entries, and marks them as logged. They were queued rather than submitted, so UserPromptSubmit
// never saw them and the audit would otherwise show an answer to a question it never recorded.
//
// Counted over the whole transcript against what was already logged, because a queued message
// leaves no marker once it is consumed — scanning only this turn's slice would re-emit it on every
// later turn. Each one is its own entry, so it gets its own timestamp; the label carries the time
// it was actually typed, which is earlier than the entry itself by however long the turn ran.
func takeQueuedPrompts(path, session8 string) []string {
	lines := readTranscript(path)
	if len(lines) == 0 {
		return nil
	}
	var out []string
	seen := queuedSeen(session8)
	found := 0
	for _, l := range lines {
		if l.Type != "queue-operation" || l.Operation != "enqueue" {
			continue
		}
		found++
		if found <= seen {
			continue
		}
		text := cleanBlock(l.Content, 2000)
		if text == "" {
			continue
		}
		out = append(out, "prompt (typed while the turn was running):\n"+text)
	}
	if found != seen {
		setQueuedSeen(session8, found)
	}
	return out
}

// takeSubagents returns how many subagents started since the last turn and resets the tally.
func takeSubagents(session8 string) int {
	p := tallyPath(session8)
	info, err := os.Stat(p)
	if err != nil {
		return 0
	}
	_ = os.Remove(p)
	return int(info.Size())
}

// A queued message stays in the transcript as a queue-operation for good — nothing marks it as
// consumed — so the count already written to the log is remembered next to the subagent tally.
// Otherwise every later turn would repeat the same message.
func queuedSeenPath(session8 string) string {
	return tallyPath(session8) + ".queued"
}

func queuedSeen(session8 string) int {
	raw, err := os.ReadFile(queuedSeenPath(session8))
	if err != nil {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func setQueuedSeen(session8 string, n int) {
	p := queuedSeenPath(session8)
	if os.MkdirAll(filepath.Dir(p), 0o700) != nil {
		return
	}
	_ = os.WriteFile(p, []byte(strconv.Itoa(n)), 0o600)
}

// sessionLogDir finds the date folder a session already started writing in, newest first, so a
// session that crosses midnight is not split into two files with the same name.
func sessionLogDir(base, file string) string {
	entries, err := os.ReadDir(base)
	if err != nil {
		return ""
	}
	for i := len(entries) - 1; i >= 0; i-- {
		if !entries[i].IsDir() {
			continue
		}
		dir := filepath.Join(base, entries[i].Name())
		if _, err := os.Stat(filepath.Join(dir, file)); err == nil {
			return dir
		}
	}
	return ""
}

// appendAudit writes one blank-line-separated entry to the per-session audit log.
func appendAudit(repoRoot, slug, session8, line string) {
	// <author-slug>/<date>/<session8>.log: the developer owns the folder, each day is
	// a folder inside it, and one file per session keeps parallel sessions of the same
	// author from interleaving.
	name := session8
	if name == "" {
		name = "session"
	}
	base := filepath.Join(repoRoot, "management", "logs", slug)
	dir := filepath.Join(base, time.Now().Format("2006-01-02"))
	// A session that runs past midnight keeps writing where it started: one file per session, not
	// one per session per day — otherwise the second half opens mid-turn, with no session header.
	if _, err := os.Stat(filepath.Join(dir, name+".log")); err != nil {
		if started := sessionLogDir(base, name+".log"); started != "" {
			dir = started
		}
	}
	if os.MkdirAll(dir, 0o755) != nil {
		return
	}
	f, err := os.OpenFile(filepath.Join(dir, name+".log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s\n\n", line)
}

// format turns a hook event into the text to append, or "" to skip.
func format(e hookEvent, session8 string) string {
	switch e.HookEventName {
	case "UserPromptSubmit":
		// A finished background task is delivered through the same event, but nobody typed it —
		// recording it as a prompt puts words in the user's mouth.
		if isSystemNotice(e.Prompt) {
			return "background task finished — see the turn below"
		}
		return promptSep + "\n\nprompt:\n" + cleanBlock(e.Prompt, 2000)
	case "SessionStart":
		return joinKV("session start:", kv("id", e.SessionID), kv("source", e.Source), kv("model", e.Model))
	case "SessionEnd":
		return joinKV("session end:", kv("reason", e.EndReason), kv("duration", sessionDuration(e.TranscriptPath)))
	case "Stop":
		block := turnStats(e.TranscriptPath, session8)
		// Capture Claude Code's final answer (the "what was done" report), blank lines dropped.
		if r := cleanBlock(e.LastAssistant, 4000); r != "" {
			if block != "" {
				block += "\n\nanswer:\n" + r
			} else {
				block = "answer:\n" + r
			}
		}
		return block
	case "SubagentStart":
		// Only log a readable agent type — never a raw agent id.
		if e.AgentType == "" || isFanOutAgent(e.AgentType) {
			return ""
		}
		return "subagent start: " + e.AgentType
	case "SubagentStop":
		t := e.AgentType
		if t == "" || isFanOutAgent(t) {
			return "" // no readable type → skip (a bare agent id is noise)
		}
		if e.StopReason != "" {
			return "subagent done: " + t + " (" + e.StopReason + ")"
		}
		return "subagent done: " + t
	case "StopFailure":
		// error_details may be a JSON string or an object — stringify either way.
		details := jsonStr(e.ErrorDetails)
		if details == "null" {
			details = ""
		}
		if e.ErrorType == "" && details == "" {
			return ""
		}
		out := "error: " + firstNonEmpty(e.ErrorType, "unknown")
		if details != "" {
			out += " — " + trunc(details, 200)
		}
		return out
	case "PreCompact":
		if e.Compaction == "" {
			return ""
		}
		return "compaction: trigger=" + e.Compaction
	case "PostToolUse":
		// Only AskUserQuestion is hooked here — capture each question Claude asked
		// and the option the user picked. Never per-tool noise.
		if e.ToolName == "AskUserQuestion" {
			return askBlock(e.ToolInput, e.ToolResponse)
		}
	}
	return ""
}

type askInput struct {
	Questions []struct {
		Question string `json:"question"`
		Header   string `json:"header"`
	} `json:"questions"`
}

// askBlock renders each question and the option the user picked, e.g.:
//
//	question: <header>
//	<question text>
//	user chose: <picked option>
//
// `user chose:` rather than `answer:` on purpose — `answer:` is Claude's own reply at the end of a
// turn, and the two used to be indistinguishable in the log.
func askBlock(input, response json.RawMessage) string {
	var in askInput
	if json.Unmarshal(input, &in) != nil || len(in.Questions) == 0 {
		return ""
	}
	byQuestion, inOrder := answersOf(response)
	var out []string
	for i, q := range in.Questions {
		head := "question:"
		if q.Header != "" {
			head += " " + q.Header
		}
		out = append(out, head, cleanBlock(q.Question, 500), "user chose: "+pickAnswer(byQuestion, inOrder, q.Question, i))
	}
	return strings.Join(out, "\n")
}

var answerPairRe = regexp.MustCompile(`"([^"]{1,400})"="([^"]{1,400})"`)

// answersOf pulls the picked options out of the tool result. Claude Code hands the hook the
// structured result — `{"answers": {"<question>": "<option>"}}` — while the text the model sees is
// the sentence `Your questions have been answered: "<question>"="<option>"`. Both shapes are read
// here, so a change on either side degrades to "(not captured)" instead of to a wrong answer.
//
// The second return value is the picks in order, and only from the shapes that HAVE an order: Go
// map iteration is random, so a positional fallback on the map form could pair a question with
// someone else's answer.
func answersOf(response json.RawMessage) (map[string]string, []string) {
	byQuestion := map[string]string{}
	var inOrder []string

	var structured struct {
		Answers json.RawMessage `json:"answers"`
	}
	if json.Unmarshal(response, &structured) == nil && len(structured.Answers) > 0 {
		var asMap map[string]string
		if json.Unmarshal(structured.Answers, &asMap) == nil && len(asMap) > 0 {
			for q, a := range asMap {
				byQuestion[q] = a
			}
			return byQuestion, nil
		}
		var asList []struct {
			Question string `json:"question"`
			Answer   string `json:"answer"`
		}
		if json.Unmarshal(structured.Answers, &asList) == nil && len(asList) > 0 {
			for _, qa := range asList {
				byQuestion[qa.Question] = qa.Answer
				inOrder = append(inOrder, qa.Answer)
			}
			return byQuestion, inOrder
		}
	}
	for _, m := range answerPairRe.FindAllStringSubmatch(jsonStr(response), -1) {
		byQuestion[m[1]] = m[2]
		inOrder = append(inOrder, m[2])
	}
	return byQuestion, inOrder
}

// pickAnswer matches by question text first; position is only a fallback, and only for the shapes
// that carry an order.
func pickAnswer(byQuestion map[string]string, inOrder []string, question string, idx int) string {
	if a := strings.TrimSpace(byQuestion[question]); a != "" {
		return a
	}
	if idx < len(inOrder) && strings.TrimSpace(inOrder[idx]) != "" {
		return strings.TrimSpace(inOrder[idx])
	}
	return "(not captured)"
}

// jsonStr returns raw as a Go string whether it is a JSON string or already text.
func jsonStr(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	return string(raw)
}

func kv(k, v string) string {
	if v == "" {
		return ""
	}
	return k + "=" + v
}

func joinKV(prefix string, parts ...string) string {
	var nz []string
	for _, p := range parts {
		if p != "" {
			nz = append(nz, p)
		}
	}
	if len(nz) == 0 {
		return ""
	}
	return prefix + " " + strings.Join(nz, " ")
}

// turnStats summarizes the latest turn (main-thread work since the last typed
// user prompt) as a short indented block. Returns "" when nothing to report.
func turnStats(path, session8 string) string {
	// Read the tally even when there is nothing to report, so a turn that produced no stats does not
	// leave its agents to be counted against the next one.
	spawned := takeSubagents(session8)
	lines := readTranscript(path)
	if len(lines) == 0 {
		return ""
	}

	start := 0
	for i := len(lines) - 1; i >= 0; i-- {
		l := lines[i]
		if l.Type == "user" && !l.IsSidechain && l.PromptSource == "typed" && isJSONString(l.Message.Content) {
			start = i
			break
		}
	}

	model, effort := "", ""
	startTs, endTs := lines[start].Timestamp, ""
	tools := map[string]int{}
	touched := map[string]bool{}
	total := 0
	for _, l := range lines[start:] {
		if l.Type != "assistant" || l.IsSidechain { // main-thread assistant work only
			continue
		}
		if l.Message.Model != "" {
			model = l.Message.Model
		}
		if l.Effort != "" {
			effort = l.Effort
		}
		if l.Timestamp != "" {
			endTs = l.Timestamp
		}
		for _, b := range blocks(l.Message.Content) {
			if b.Type != "tool_use" {
				continue
			}
			tools[b.Name]++
			total++
			switch b.Name {
			case "Write", "Edit", "MultiEdit", "NotebookEdit":
				fp := b.Input.FilePath
				if fp == "" {
					fp = b.Input.NotebookPath
				}
				if fp != "" {
					touched[relPath(fp)] = true
				}
			}
		}
	}
	if total == 0 && model == "" {
		return ""
	}

	var f []string
	if model != "" {
		f = append(f, "model="+model)
	}
	if effort != "" {
		f = append(f, "effort="+effort)
	}
	if d := durationSec(startTs, endTs); d >= 0 {
		f = append(f, "duration="+fmtDur(d))
	}
	// SubagentStart fires once per agent, including every agent of a Workflow fan-out, so the tally
	// is the real number. The tool count stays as a floor in case the events did not reach the hook.
	spawns := tools["Task"] + tools["Agent"]
	if spawned > spawns {
		spawns = spawned
	}
	f = append(f, fmt.Sprintf("subagents=%d", spawns))
	tl := fmt.Sprintf("tools=%d", total)
	if bd := breakdown(tools); bd != "" {
		tl += ": " + bd
	}
	f = append(f, tl)
	// turn: label on its own line, all metrics on one line below it.
	res := "turn:\n" + strings.Join(f, " ")
	if len(touched) > 0 {
		files := make([]string, 0, len(touched))
		for p := range touched {
			files = append(files, p)
		}
		sort.Strings(files)
		stats := fileStats(os.Getenv("CLAUDE_PROJECT_DIR"), files)
		lines := make([]string, 0, len(files))
		for _, p := range files {
			if s := stats[p]; s != "" {
				lines = append(lines, s)
			} else {
				lines = append(lines, p) // no git stats (e.g. git unavailable): bare path
			}
		}
		res += "\n\nfiles:\n" + strings.Join(lines, "\n")
	}
	return res
}

// sessionDuration is the span between the first and last transcript records.
func sessionDuration(path string) string {
	lines := readTranscript(path)
	first, last := "", ""
	for _, l := range lines {
		if l.Timestamp == "" {
			continue
		}
		if first == "" {
			first = l.Timestamp
		}
		last = l.Timestamp
	}
	if d := durationSec(first, last); d >= 0 {
		return fmtDur(d)
	}
	return ""
}

func readTranscript(path string) []tLine {
	if path == "" {
		return nil
	}
	fh, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer fh.Close()
	sc := bufio.NewScanner(fh)
	sc.Buffer(make([]byte, 1024*1024), 64*1024*1024) // transcript lines can be large
	var lines []tLine
	for sc.Scan() {
		var l tLine
		if json.Unmarshal(sc.Bytes(), &l) == nil {
			lines = append(lines, l)
		}
	}
	return lines
}

func isJSONString(r json.RawMessage) bool {
	return len(r) > 0 && r[0] == '"'
}

type block struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Input struct {
		FilePath     string `json:"file_path"`
		NotebookPath string `json:"notebook_path"`
	} `json:"input"`
}

// relPath makes an absolute file path repo-relative (forward slashes) via
// CLAUDE_PROJECT_DIR, so the log lists e.g. management/tasks/TEMPLATE.md.
func relPath(p string) string {
	p = filepath.ToSlash(p)
	if root := filepath.ToSlash(os.Getenv("CLAUDE_PROJECT_DIR")); root != "" && strings.HasPrefix(p, root) {
		p = strings.TrimPrefix(strings.TrimPrefix(p, root), "/")
	}
	return p
}

func blocks(r json.RawMessage) []block {
	if !isJSONString(r) {
		var bs []block
		if json.Unmarshal(r, &bs) == nil {
			return bs
		}
	}
	return nil
}

func durationSec(startTs, endTs string) int {
	s, err1 := time.Parse(time.RFC3339, startTs)
	e, err2 := time.Parse(time.RFC3339, endTs)
	if err1 != nil || err2 != nil {
		return -1
	}
	if d := int(e.Sub(s).Seconds()); d >= 0 {
		return d
	}
	return -1
}

// breakdown renders the busiest tools as "Edit×7, Read×5, …" (top 6), excluding
// the Task/Agent subagent tools, which are counted separately.
func breakdown(tools map[string]int) string {
	type kvn struct {
		name string
		n    int
	}
	var xs []kvn
	for name, n := range tools {
		if name == "Task" || name == "Agent" {
			continue
		}
		xs = append(xs, kvn{name, n})
	}
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].n != xs[j].n {
			return xs[i].n > xs[j].n
		}
		return xs[i].name < xs[j].name
	})
	var out []string
	for i, x := range xs {
		if i >= 6 {
			break
		}
		out = append(out, fmt.Sprintf("%s×%d", x.name, x.n))
	}
	return strings.Join(out, ", ")
}
