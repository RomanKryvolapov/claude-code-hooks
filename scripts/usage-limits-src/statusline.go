// statusline.go — the status line: what Claude Code hands it, and how it is drawn.
//
// The status line is the one output a person reads rather than a model or a program, so it is the
// only one painted and the only one fitted to the terminal. It is also the only one Claude Code
// hands the real size of the context window and its own 5-hour and 7-day figures, which stand in
// for the usage API's, without a word, when those have gone stale. Beside the CONTEXT gauge it
// writes how long the prompt cache stays warm, and last, under the limits, how many tokens the
// session and the account have used. Of the limits it draws the account's windows alone; a model's
// weekly bucket of its own is the injected block's.

package main

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"
)

// statusInput is the part of the status line's input this hook reads. Any field can be missing:
// Claude Code sends the limits only to a subscriber and only after the first reply, and an older
// Claude Code sends fewer fields altogether.
type statusInput struct {
	TranscriptPath string `json:"transcript_path"`
	Model          *struct {
		ID string `json:"id"`
	} `json:"model"`
	ContextWindow *struct {
		TotalInputTokens  float64 `json:"total_input_tokens"`
		ContextWindowSize float64 `json:"context_window_size"`
		// Null before the session's first reply and again right after a compaction, which is when
		// the total beside it cannot be taken at its word.
		CurrentUsage json.RawMessage `json:"current_usage"`
	} `json:"context_window"`
	RateLimits *struct {
		FiveHour *liveWindow `json:"five_hour"`
		SevenDay *liveWindow `json:"seven_day"`
	} `json:"rate_limits"`
	PromptCache *promptCache `json:"prompt_cache"`
	// ContextEstimate is the panel plugin's alone: what /context estimates the context holds, sent
	// while no request has measured it — before a session's first reply; see renderPanel.
	ContextEstimate float64 `json:"context_estimate"`
}

// promptCache is how long the prompt cache stays warm, as Claude Code reports it to a status line —
// or, for the panel, which Claude Code hands no such reading, as the transcript shows it; see
// transcriptPromptCache.
type promptCache struct {
	Warm            bool     `json:"warm"`
	CachingObserved bool     `json:"caching_observed"`
	TTL             string   `json:"ttl"`        // the lifetime it was written with: "5m" or "1h"
	ExpiresAt       *float64 `json:"expires_at"` // Unix seconds; null once nothing is cached
}

type liveWindow struct {
	UsedPercentage *float64 `json:"used_percentage"`
	ResetsAt       *float64 `json:"resets_at"` // Unix seconds
}

// decodeStatusInput reads what it can. A field of an unexpected type costs that field and not the
// rest, and input that is not JSON at all leaves every field empty.
func decodeStatusInput(raw string) *statusInput {
	in := &statusInput{}
	if raw != "" {
		_ = json.Unmarshal([]byte(raw), in)
	}
	return in
}

// liveLimits turns the limits Claude Code handed over into the shape the usage API answers in. A
// window whose reset has already passed says nothing about the one running now, and is left out.
func (in *statusInput) liveLimits(now time.Time) *usageRaw {
	if in.RateLimits == nil {
		return nil
	}
	window := func(w *liveWindow) *usageWindow {
		if w == nil || w.UsedPercentage == nil || w.ResetsAt == nil {
			return nil
		}
		reset := time.Unix(int64(*w.ResetsAt), 0)
		if !reset.After(now) {
			return nil
		}
		return &usageWindow{Utilization: *w.UsedPercentage, ResetsAt: reset.UTC().Format(time.RFC3339)}
	}
	five, seven := window(in.RateLimits.FiveHour), window(in.RateLimits.SevenDay)
	if five == nil && seven == nil {
		return nil
	}
	return &usageRaw{FiveHour: five, SevenDay: seven}
}

// statusContext is the CONTEXT reading from Claude Code's own figures: the real size of the window
// the model has, which the configured size can only guess at, and the tokens the last request
// carried. Where Claude Code has no reading yet, the transcript is read instead, as the injected
// block does.
func (in *statusInput) statusContext() contextReading {
	r := contextReading{window: contextWindowTokens}
	if cw := in.ContextWindow; cw != nil {
		if cw.ContextWindowSize > 0 {
			r.window = cw.ContextWindowSize
		}
		if isJSONValue(cw.CurrentUsage) && cw.TotalInputTokens > 0 {
			r.tokens, r.ok = cw.TotalInputTokens, true
			return r
		}
	}
	r.tokens, r.ok = contextTokens(in.TranscriptPath)
	return r
}

// cacheText is how long the prompt cache stays warm, painted on the gauges' scale; see cacheReading.
// Nothing when Claude Code reports no cache: an older version, or no reply yet.
func (in *statusInput) cacheText(now time.Time, p painter) string {
	text, at, ok := in.cacheReading(now)
	if !ok {
		return ""
	}
	return "CACHE " + p.paint(text, p.gaugeColour(at))
}

// cacheReading is how long the prompt cache stays warm — a wait, COLD, or WARM where Claude Code
// gives no end — and where that stands on the gauges' scale: green with the whole cache lifetime
// ahead, red as it runs out, and a cold cache red. Past that, the next request writes the whole
// context into the cache again — the dearest request a long session makes. Not ok when Claude Code
// reports no cache.
func (in *statusInput) cacheReading(now time.Time) (text string, at float64, ok bool) {
	c := in.PromptCache
	if c == nil || !c.CachingObserved {
		return "", 0, false
	}
	if !c.Warm {
		return "COLD", 1, true
	}
	if c.ExpiresAt == nil {
		return "WARM", 0, true
	}
	left := time.Unix(int64(*c.ExpiresAt), 0).Sub(now).Minutes()
	if left <= 0 {
		return "COLD", 1, true
	}
	// The lifetime the cache was written with, "5m" or "1h"; an hour where Claude Code does not say.
	lifetime := 60.0
	if d, err := time.ParseDuration(c.TTL); err == nil && d > 0 {
		lifetime = d.Minutes()
	}
	return formatWait(left), 1 - left/lifetime, true
}

// isJSONValue reports whether a field was present and not null.
func isJSONValue(raw json.RawMessage) bool {
	s := strings.TrimSpace(string(raw))
	return s != "" && s != "null"
}

// painter colours text for a terminal. The zero value paints nothing, which is what every output
// but the status line uses. rgb says the terminal draws 24-bit colour; without it the gauges fall
// back to the 256-colour palette.
type painter struct{ on, rgb bool }

// statusPainter is the status line's painter: on, unless NO_COLOR says otherwise, and in 24-bit
// colour everywhere but Apple's Terminal; see gaugePalette256.
func statusPainter() painter {
	return painter{on: os.Getenv("NO_COLOR") == "", rgb: os.Getenv("TERM_PROGRAM") != "Apple_Terminal"}
}

func (p painter) paint(s, code string) string {
	if !p.on || code == "" || s == "" {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

// statusWidth is how many columns the status line may use, from the terminal size Claude Code
// passes in COLUMNS; zero when it passes none.
func statusWidth() int {
	n, err := strconv.Atoi(strings.TrimSpace(os.Getenv("COLUMNS")))
	if err != nil || n <= 0 {
		return 0
	}
	return max(n-statusMarginCols, 1)
}

// limitWindows is where the limits' own windows opened, counted back from their resets: the account's
// tokens are counted in them. Both are zero without a weekly reset to count back from, and the
// account's tokens are then left out; fiveStart alone is zero where there is no running 5-hour
// window, and the 5-hour count is then left out.
func limitWindows(st *state) (fiveStart, weekFrom time.Time) {
	if st == nil || st.r7Local.IsZero() {
		return time.Time{}, time.Time{}
	}
	if st.hasSession && !st.r5Local.IsZero() {
		fiveStart = st.r5Local.Add(-time.Duration(sessionWindowMin) * time.Minute)
	}
	return fiveStart, st.r7Local.Add(-time.Duration(weeklyWindowMin) * time.Minute)
}

// limitsState is what the limits are now, for the status line and the panel: for the model Claude
// Code names, or the one worked out from the session where it names none, with Claude Code's own
// figures standing in for stale ones of the usage API's. Nil without any figures.
func (in *statusInput) limitsState(args argSet, hookInput map[string]any, now time.Time) *state {
	activeModel := ""
	if in.Model != nil && in.Model.ID != "" {
		activeModel = cleanModelString(in.Model.ID)
	} else {
		activeModel = getActiveModel(args, hookInput)
	}
	return getState(args, activeModel, in.liveLimits(now))
}

// renderStatusline draws the whole status line.
func renderStatusline(args argSet, rawInput string, hookInput map[string]any, now time.Time) string {
	in := decodeStatusInput(rawInput)
	st := in.limitsState(args, hookInput, now)
	fiveStart, weekFrom := limitWindows(st)
	p := statusPainter()
	// The CONTEXT line carries how long the cache stays warm and, two spaces on, the token counts,
	// where the terminal has room for them; see renderOpts.tokens.
	o := renderOpts{p: p, width: statusWidth(), contextTail: in.cacheText(now, p),
		tokens: tokensText(in.TranscriptPath, fiveStart, weekFrom, now)}

	var lines []string
	if st == nil {
		// No limits to line the CONTEXT gauge up with: all of it beside the gauge, which shortens to fit.
		ctx := in.statusContext()
		o.contextTail = o.withTokens()
		if top := topLine(ctx, o, headsFor(nil, ctx), ctxBarCells); top != "" {
			lines = append(lines, top)
		}
		lines = append(lines, "LIMITS -> N/A")
	} else {
		lines = append(lines, renderLimits(windowRows(st, now), in.statusContext(), o))
	}
	if rows := configRows(); len(rows) > 0 {
		lines = append(lines, renderTextRows(rows))
	}
	return strings.Join(lines, "\n")
}
