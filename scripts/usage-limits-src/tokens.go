// tokens.go — how many tokens this session, and the account on this machine, have used.
//
// Counted from Claude Code's own transcripts, where every reply is written down with the usage the
// API reported for it: the session's and its subagents' for the session, and every transcript
// written to in the current weekly window for the account. That makes the account's figures a count
// for this machine — a session on another computer, or on claude.ai, is in the limits but not in
// these files — and every figure a few percent below what /usage says, since Claude Code does not
// write its own service requests to a transcript.
//
// The transcripts folder runs to gigabytes and the status line is redrawn every half-minute, so each
// file is read on from where the last run stopped, what was read is kept in a cache file of its own,
// and a run spends at most a fixed budget reading: the first pass over the folder is spread across
// several redraws, and the account's figures say so until it is done.

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	// How long one run may spend reading transcripts.
	tokenReadBudget = 700 * time.Millisecond
	// How often the whole transcripts folder is walked again once it has been read through. The
	// session's own files are read on every run.
	tokenRescanEvery = 60 * time.Second
	// Replies are summed in five-minute buckets, by the time they began: as fine as a window's edge
	// needs placing, and coarse enough that a week of them stays small.
	tokenBucketSec = 5 * 60
)

// tokenSums is the usage of a set of requests, in the four parts the API reports it in.
type tokenSums struct {
	Input  float64 `json:"input"`
	Output float64 `json:"output"`
	Write  float64 `json:"write"` // written to the prompt cache
	Read   float64 `json:"read"`  // read back from it
}

func (s *tokenSums) add(o tokenSums) {
	s.Input += o.Input
	s.Output += o.Output
	s.Write += o.Write
	s.Read += o.Read
}

// fresh is everything processed at full price or more: new input, output, and cache writes. What
// was read back from the cache is the rest.
func (s tokenSums) fresh() float64 { return s.Input + s.Output + s.Write }
func (s tokenSums) total() float64 { return s.fresh() + s.Read }

// fileTokens is how far one transcript has been read, and what it held.
type fileTokens struct {
	Size   int64     `json:"size"`   // the size it had when last read to its end
	Offset int64     `json:"offset"` // how far it has been read, in whole lines
	Sum    tokenSums `json:"sum"`
	// The same replies by the five minutes each began in, Unix seconds; only those of the current
	// weekly window are kept.
	Buckets map[int64]tokenSums `json:"buckets,omitempty"`
	// The reply being read when the file ended. Claude Code writes a reply as one line per content
	// block, every line carrying the reply's usage, and only the last line's output count is final
	// — so a reply is added only once a line of the next one appears, and until then it waits here
	// and is counted from here.
	OpenKey string    `json:"open_key,omitempty"`
	OpenAt  int64     `json:"open_at,omitempty"`
	Open    tokenSums `json:"open"`
}

// transcriptLine is the part of a transcript line that carries usage.
type transcriptLine struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	RequestID string `json:"requestId"`
	Message   struct {
		ID    string `json:"id"`
		Usage *struct {
			Input  float64 `json:"input_tokens"`
			Output float64 `json:"output_tokens"`
			Write  float64 `json:"cache_creation_input_tokens"`
			Read   float64 `json:"cache_read_input_tokens"`
		} `json:"usage"`
	} `json:"message"`
}

// bucketOf is the five minutes a moment falls in; zero when the moment cannot be read.
func bucketOf(timestamp string) int64 {
	t, ok := parseTime(timestamp)
	if !ok {
		return 0
	}
	return t.Unix() / tokenBucketSec * tokenBucketSec
}

// take counts one line of a reply.
func (f *fileTokens) take(e *transcriptLine) {
	u := tokenSums{Input: e.Message.Usage.Input, Output: e.Message.Usage.Output,
		Write: e.Message.Usage.Write, Read: e.Message.Usage.Read}
	key := e.Message.ID + "|" + e.RequestID
	if key == "|" {
		// Nothing ties this line to any other, so it is a reply of its own.
		f.settle()
		f.count(bucketOf(e.Timestamp), u)
		return
	}
	if key == f.OpenKey {
		f.Open = u // the same reply again, and this line's count is the later one
		return
	}
	f.settle()
	f.OpenKey, f.OpenAt, f.Open = key, bucketOf(e.Timestamp), u
}

// settle counts the reply that was waiting, now that it is complete.
func (f *fileTokens) settle() {
	if f.OpenKey != "" {
		f.count(f.OpenAt, f.Open)
	}
	f.OpenKey, f.OpenAt, f.Open = "", 0, tokenSums{}
}

func (f *fileTokens) count(bucket int64, u tokenSums) {
	f.Sum.add(u)
	if f.Buckets == nil {
		f.Buckets = map[int64]tokenSums{}
	}
	s := f.Buckets[bucket]
	s.add(u)
	f.Buckets[bucket] = s
}

// total is everything the file holds, the reply still open included.
func (f *fileTokens) total() tokenSums {
	s := f.Sum
	if f.OpenKey != "" {
		s.add(f.Open)
	}
	return s
}

// since is what the file holds from a moment on: every five minutes that end after it, so a reply
// begun in the minutes just before a window opened is counted in it rather than lost.
func (f *fileTokens) since(start int64) tokenSums {
	var s tokenSums
	for b, v := range f.Buckets {
		if b+tokenBucketSec > start {
			s.add(v)
		}
	}
	if f.OpenKey != "" && f.OpenAt+tokenBucketSec > start {
		s.add(f.Open)
	}
	return s
}

// forget drops the five minutes that ended before a moment.
func (f *fileTokens) forget(start int64) {
	for b := range f.Buckets {
		if b+tokenBucketSec <= start {
			delete(f.Buckets, b)
		}
	}
}

// read reads a transcript on from where it was left, whole lines only: a line still being written
// is left for the next run. It stops at the deadline and reports whether it got to the end.
func (f *fileTokens) read(path string, size int64, deadline time.Time) bool {
	if size < f.Offset {
		*f = fileTokens{} // rewritten from the start, so read it again from the start
	}
	file, err := os.Open(path)
	if err != nil {
		return true // gone or unreadable: there is nothing in it to count
	}
	defer file.Close()
	if _, err := file.Seek(f.Offset, io.SeekStart); err != nil {
		return true
	}
	r := bufio.NewReaderSize(file, 256<<10)
	for {
		// Checked on every line: a line runs to megabytes when a reply carries an image, so a check
		// every so many lines could overrun the budget by a long way.
		if time.Now().After(deadline) {
			return false
		}
		line, err := r.ReadBytes('\n')
		if err != nil {
			// The end, or a last line that is not finished yet.
			f.Size = size
			return true
		}
		f.Offset += int64(len(line))
		if bytes.Contains(line, []byte(`"usage"`)) && bytes.Contains(line, []byte(`"assistant"`)) {
			var e transcriptLine
			_ = json.Unmarshal(line, &e)
			if e.Type == "assistant" && e.Message.Usage != nil {
				f.take(&e)
			}
		}
	}
}

// fileKey is how a transcript is named in the cache. One file reached two ways — the path Claude
// Code hands the status line and the one the walk of the folder finds, which can differ in their
// separators or, on Windows, in case — must be one entry, or it is counted twice.
func fileKey(path string) string {
	path = filepath.Clean(path)
	if runtime.GOOS == "windows" {
		path = strings.ToLower(path)
	}
	return path
}

// sessionFiles are the files one session writes: its own transcript, and those of its subagents in
// the folder beside it.
func sessionFiles(transcript string) []string {
	files := []string{transcript}
	sub := filepath.Join(strings.TrimSuffix(transcript, filepath.Ext(transcript)), "subagents")
	_ = filepath.WalkDir(sub, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".jsonl") {
			files = append(files, p)
		}
		return nil
	})
	return files
}

// tokenCache is every transcript of the current weekly window, as far as each has been read.
type tokenCache struct {
	// When the folder was last read through; empty while a pass is still going.
	ReadThroughAt string                 `json:"read_through_at"`
	Files         map[string]*fileTokens `json:"files"`
}

// tokenStats is what the status line writes about tokens. The account's figures are there only
// when the limits say where their windows begin.
type tokenStats struct {
	session             tokenSums
	sessionWhole        bool
	account             bool
	fiveHour, weekly    tokenSums
	accountWhole        bool
	fiveStart, weekFrom int64
}

// readTokens brings the cache up to date and adds it up: the session's files every time, and every
// transcript written to since `from` — the weekly window's start — whenever the last full pass is
// older than tokenRescanEvery or never finished. A zero `from` leaves the account out.
func readTokens(transcript string, fiveStart, from, now time.Time) tokenStats {
	st := tokenStats{sessionWhole: true, accountWhole: true}
	c := &tokenCache{}
	if b, err := os.ReadFile(tokenCachePath); err == nil {
		_ = json.Unmarshal(stripBOM(b), c)
	}
	if c.Files == nil {
		c.Files = map[string]*fileTokens{}
	}
	deadline := now.Add(tokenReadBudget)
	changed := false
	readFile := func(p string, size int64) bool {
		key := fileKey(p)
		f := c.Files[key]
		if f == nil {
			f = &fileTokens{}
			c.Files[key] = f
		}
		if f.Size == size {
			return true
		}
		changed = true
		return f.read(p, size, deadline)
	}

	seen := map[string]bool{}
	session := sessionFiles(transcript)
	for _, p := range session {
		if info, err := os.Stat(p); err == nil {
			seen[fileKey(p)] = true
			if !readFile(p, info.Size()) {
				st.sessionWhole = false
			}
		}
	}

	if !from.IsZero() {
		last, ok := parseTime(c.ReadThroughAt)
		if !ok || now.Sub(last) >= tokenRescanEvery {
			_ = filepath.WalkDir(filepath.Join(claudeDir, "projects"), func(p string, d fs.DirEntry, err error) error {
				if err != nil || d.IsDir() || !strings.HasSuffix(p, ".jsonl") {
					return nil
				}
				info, err := d.Info()
				if err != nil || info.ModTime().Before(from) {
					return nil // last written before the week began, so nothing in it is this week's
				}
				seen[fileKey(p)] = true
				if time.Now().After(deadline) {
					if f := c.Files[fileKey(p)]; f == nil || f.Size != info.Size() {
						st.accountWhole = false
					}
					return nil
				}
				if !readFile(p, info.Size()) {
					st.accountWhole = false
				}
				return nil
			})
			for p, f := range c.Files {
				if !seen[p] {
					delete(c.Files, p)
					changed = true
				} else if len(f.Buckets) > 0 {
					f.forget(from.Unix())
				}
			}
			switch {
			case st.accountWhole:
				c.ReadThroughAt = now.Format(time.RFC3339Nano)
				changed = true
			case c.ReadThroughAt != "":
				c.ReadThroughAt = ""
				changed = true
			}
		}
		st.account = true
		st.fiveStart, st.weekFrom = fiveStart.Unix(), from.Unix()
		if fiveStart.IsZero() {
			st.fiveStart = st.weekFrom
		}
		for _, f := range c.Files {
			st.fiveHour.add(f.since(st.fiveStart))
			st.weekly.add(f.since(st.weekFrom))
		}
	}

	for _, p := range session {
		if f := c.Files[fileKey(p)]; f != nil {
			st.session.add(f.total())
		}
	}
	if changed {
		if b, err := json.Marshal(c); err == nil {
			_ = writeFileAtomic(tokenCachePath, b)
		}
	}
	return st
}

// freshAndCached is how a set of token counts is written: "FRESH 1m CACHED 108m".
func freshAndCached(s tokenSums) string {
	return "FRESH " + shortTokens(s.fresh()) + " CACHED " + shortTokens(s.Read)
}

// sessionTokenStats is readTokens for a session that has something to count: false without a
// transcript — somebody running the hook by hand has no session — and before the first reply.
func sessionTokenStats(transcript string, fiveStart, weekFrom, now time.Time) (tokenStats, bool) {
	if tokenCachePath == "" || transcript == "" {
		return tokenStats{}, false
	}
	st := readTokens(transcript, fiveStart, weekFrom, now)
	return st, st.session.total() > 0 || st.weekly.total() > 0
}

// tokensText is the session's tokens and, where the limits' windows are known, the account's in
// each, as the status line writes them: "SESSION: FRESH 1m CACHED 108m, ACCOUNT 5H: FRESH 20m CACHED
// 1.2b, 7D: FRESH 21m CACHED 1.2b".
func tokensText(transcript string, fiveStart, weekFrom, now time.Time) string {
	st, ok := sessionTokenStats(transcript, fiveStart, weekFrom, now)
	if !ok {
		return ""
	}
	text := "SESSION: " + freshAndCached(st.session)
	if !st.sessionWhole {
		text += " SO FAR"
	}
	if st.account {
		text += ", ACCOUNT 5H: " + freshAndCached(st.fiveHour) + ", 7D: " + freshAndCached(st.weekly)
		if !st.accountWhole {
			text += " SO FAR"
		}
	}
	return text
}
