// statusline.go — the status line: what Claude Code hands it, and how it is drawn.
//
// The status line is the one output a person reads rather than a model or a program, so it is the
// only one painted and the only one fitted to the terminal. It is also the only one Claude Code
// hands the real size of the context window and its own 5-hour and 7-day figures, which stand in
// for the usage API's, without a word, when those have gone stale. Beside the CONTEXT gauge it
// writes how many tokens the session has used; otherwise the picture is the one the injected block
// carries, without the lines under it.

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
	PromptCache *struct {
		Warm            bool     `json:"warm"`
		CachingObserved bool     `json:"caching_observed"`
		TTL             string   `json:"ttl"`        // the lifetime it was written with: "5m" or "1h"
		ExpiresAt       *float64 `json:"expires_at"` // Unix seconds; null once nothing is cached
	} `json:"prompt_cache"`
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

// contextTail is what the status line writes after the CONTEXT gauge: how long the prompt cache
// stays warm, then the tokens of the session and of the account in the limits' windows — whichever
// there is — "CACHE 0:59, SESSION: FRESH 1m CACHED 108m, ACCOUNT 5H: FRESH 20m CACHED 1.2b, 7D: …".
func (in *statusInput) contextTail(now time.Time, p painter, fiveStart, weekFrom time.Time) string {
	var parts []string
	if c := in.cacheText(now, p); c != "" {
		parts = append(parts, c)
	}
	if s := tokensText(in.TranscriptPath, fiveStart, weekFrom, now); s != "" {
		parts = append(parts, s)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, ", ")
}

// cacheText is how long the prompt cache stays warm. Past that, the next request writes the whole
// context into the cache again — the dearest request a long session makes. The time left is painted
// on the gauges' scale: green with the whole cache lifetime ahead, red as it runs out, and a cold
// cache red. Nothing when Claude Code reports no cache: an older version, or no reply yet.
func (in *statusInput) cacheText(now time.Time, p painter) string {
	c := in.PromptCache
	if c == nil || !c.CachingObserved {
		return ""
	}
	cold := "CACHE " + p.paint("COLD", p.gaugeColour(1))
	if !c.Warm {
		return cold
	}
	if c.ExpiresAt == nil {
		return "CACHE " + p.paint("WARM", p.gaugeColour(0))
	}
	left := time.Unix(int64(*c.ExpiresAt), 0).Sub(now).Minutes()
	if left <= 0 {
		return cold
	}
	// The lifetime the cache was written with, "5m" or "1h"; an hour where Claude Code does not say.
	lifetime := 60.0
	if d, err := time.ParseDuration(c.TTL); err == nil && d > 0 {
		lifetime = d.Minutes()
	}
	return "CACHE " + p.paint(formatWait(left), p.gaugeColour(1-left/lifetime))
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
// account's tokens are then left out.
func limitWindows(st *state) (fiveStart, weekFrom time.Time) {
	if st == nil || st.r7Local.IsZero() {
		return time.Time{}, time.Time{}
	}
	return st.r5Local.Add(-time.Duration(sessionWindowMin) * time.Minute),
		st.r7Local.Add(-time.Duration(weeklyWindowMin) * time.Minute)
}

// renderStatusline draws the whole status line.
func renderStatusline(args argSet, rawInput string, hookInput map[string]any, now time.Time) string {
	in := decodeStatusInput(rawInput)
	activeModel := ""
	if in.Model != nil && in.Model.ID != "" {
		activeModel = cleanModelString(in.Model.ID)
	} else {
		activeModel = getActiveModel(args, hookInput)
	}
	st := getState(args, activeModel, in.liveLimits(now))
	fiveStart, weekFrom := limitWindows(st)
	p := statusPainter()
	o := renderOpts{p: p, width: statusWidth(), contextTail: in.contextTail(now, p, fiveStart, weekFrom)}

	if st == nil {
		var lines []string
		if top := topLine(in.statusContext(), o); top != "" {
			lines = append(lines, top)
		}
		lines = append(lines, "LIMITS -> N/A")
		if rows := configRows(); len(rows) > 0 {
			lines = append(lines, renderTextRows(rows))
		}
		return strings.Join(lines, "\n")
	}

	out := renderLimits(limitRows(st), in.statusContext(), o)
	if rows := configRows(); len(rows) > 0 {
		out += "\n" + renderTextRows(rows)
	}
	return out
}
