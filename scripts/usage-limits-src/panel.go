// panel.go — the panel: what the status line says, as data, for the usage-limits-panel plugin.
//
// The plugin draws it above the prompt in the VS Code extension, which shows no status line, with that
// editor's own elements: the gauges as bars, the columns as boxes. A picture made of characters keeps
// its columns only in a monospaced face, and the extension draws a plugin's text in its proportional
// one, so this mode hands over the status line's figures, words and colours and leaves the laying out
// to the plugin.
//
// The input is the status line's, which the plugin builds from what Claude Code hands a plugin. That
// holds no prompt-cache reading, so where the input carries none it is read from the transcript. And
// before a session's first reply, when nothing has measured the context, the plugin adds what /context
// estimates it holds, which the status line, handed no such figure, leaves out.

package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// panelPicture is what --mode panel prints, as JSON. Every percentage is rounded as the status line
// prints it, and every colour is "#rrggbb". Every gauge's empty part is runs of one colour, first to
// last, out of the status line's hundred cells: the colours the status line paints them in, which is
// what makes the filled part as long as the status line's too.
type panelPicture struct {
	// Context is the CONTEXT gauge: measured once the session's first reply is on disk, estimated
	// before that where the input carries an estimate, and null where neither is there.
	Context *panelContext `json:"context"`
	// Cache is how long the prompt cache stays warm, the status line's text after CACHE; null where
	// nothing says.
	Cache *panelText `json:"cache"`
	// Tokens is the session's and the account's token counts as the status line writes them; empty
	// where there is no transcript to count them in.
	Tokens string `json:"tokens"`
	// Limits holds one entry per limit, in the status line's order; null without any figures at all,
	// which the status line writes as LIMITS -> N/A.
	Limits []panelLimit `json:"limits"`
	// Notes are the lines under the limits: the config the hook could not use.
	Notes []panelNote `json:"notes"`
	// FiguresColour paints the limits' figures — their percentages, reset times and the figures of a
	// wait, whose words stay unpainted.
	FiguresColour string `json:"figures_colour"`
}

type panelContext struct {
	Tokens string  `json:"tokens"` // what the last request carried, of the window: "228k/1m"
	Pct    float64 `json:"pct"`
	Colour string  `json:"colour"` // the counter's, the percentage's and the gauge's
	// Estimated says the figure is /context's estimate, before any request measured the context.
	Estimated bool        `json:"estimated"`
	Empty     []colourRun `json:"empty"` // the colours still ahead, along the gauge
}

type panelText struct {
	Text   string `json:"text"`
	Colour string `json:"colour"`
}

// panelLimit is one limit: its LIMIT gauge, and under it the share of its window gone.
type panelLimit struct {
	Label  string      `json:"label"` // "5 HOURS", "7 DAYS", or a model's name
	Pct    float64     `json:"pct"`
	Colour string      `json:"colour"` // the LIMIT gauge's, by how far the spend runs ahead of the clock
	Empty  []colourRun `json:"empty"`  // its whole scale, blue to red, across whatever is left
	Reset  string      `json:"reset"`  // the clock time it resets at; empty where the API gave none
	// Window is the gauge under it; null where the API gave no reset, where the status line draws no
	// second line either.
	Window *panelWindow `json:"window"`
}

type panelWindow struct {
	Caption string      `json:"caption"` // TIME, or WORK on the working-time scale
	Pct     float64     `json:"pct"`
	Colour  string      `json:"colour"`
	Empty   []colourRun `json:"empty"` // in the gauge's own colour
	After   string      `json:"after"` // how long until the reset, "6 days 19:57"; empty where unknown
}

type panelNote struct {
	Label string `json:"label"`
	Text  string `json:"text"`
}

// renderPanel is --mode panel: what the status line would say for the same input.
func renderPanel(args argSet, rawInput string, hookInput map[string]any, now time.Time) string {
	in := decodeStatusInput(rawInput)
	if in.PromptCache == nil {
		in.PromptCache = transcriptPromptCache(in.TranscriptPath, now)
	}
	st := in.limitsState(args, hookInput, now)
	fiveStart, weekFrom := limitWindows(st)

	pic := panelPicture{
		Tokens:        tokensText(in.TranscriptPath, fiveStart, weekFrom, now),
		Notes:         []panelNote{},
		FiguresColour: hexColour(figuresRGB),
	}
	ctx, estimated := in.statusContext(), false
	if !ctx.ok && in.ContextEstimate > 0 {
		ctx.tokens, ctx.ok, estimated = in.ContextEstimate, true, true
	}
	if ctx.ok && ctx.window > 0 {
		pct := ctx.pct()
		pic.Context = &panelContext{
			Tokens:    ctx.counter(),
			Pct:       round(pct, 0),
			Colour:    hexColour(gaugeRGB(pct / 100)),
			Estimated: estimated,
			Empty:     panelEmpty(pct, func(c, first, n int) [3]int { return gaugeRGB(contextShade(c, first, n)) }),
		}
	}
	if text, at, ok := in.cacheReading(now); ok {
		pic.Cache = &panelText{Text: text, Colour: hexColour(gaugeRGB(at))}
	}
	if st != nil {
		pic.Limits = []panelLimit{}
		for _, r := range windowRows(st, now) {
			limit := panelLimit{Label: r.label, Pct: round(r.pct, 0), Colour: hexColour(limitRGB(r.paceAt())), Reset: r.reset,
				Empty: panelEmpty(r.pct, func(c, first, n int) [3]int { return limitRGB(limitShade(c, first, n)) })}
			// windowBar marks that the row has a time gauge at all; see limitBlocks.
			if r.windowBar != "" {
				caption := r.windowCaption
				if caption == "" {
					caption = captionCalendar
				}
				limit.Window = &panelWindow{Caption: caption, Pct: round(r.windowPct, 0), Colour: hexColour(timeRGB), After: r.after,
					Empty: panelEmpty(r.windowPct, func(int, int, int) [3]int { return timeRGB })}
			}
			pic.Limits = append(pic.Limits, limit)
		}
	}
	for _, r := range configRows() {
		pic.Notes = append(pic.Notes, panelNote{Label: r.label, Text: r.text})
	}
	out, err := json.Marshal(pic)
	if err != nil {
		return ""
	}
	return string(out)
}

func hexColour(rgb [3]int) string { return fmt.Sprintf("#%02x%02x%02x", rgb[0], rgb[1], rgb[2]) }

// panelEmpty is the empty part of a gauge reading pct percent, as the status line paints it at its usual
// hundred cells in 24-bit colour: colourOf gives each cell's colour.
func panelEmpty(pct float64, colourOf func(cell, firstEmpty, cells int) [3]int) []colourRun {
	return emptyRuns(filledCells(pct, barCells), barCells, func(c, first, n int) string { return hexColour(colourOf(c, first, n)) })
}

// transcriptPromptCache is how long the prompt cache stays warm, read off the transcript: from the
// newest main-thread reply that read the cache or wrote to it, for the lifetime the newest reply that
// wrote to it was written with — 5 minutes or an hour, which the API reports in every reply's usage —
// or an hour where no reply in reach says. Nil when there is no such reply: no transcript, or no
// reply yet, which leaves the cache out, as the status line does under an older Claude Code.
//
// A reply's line is written once the model has answered, while the cache's lifetime runs from when
// the request read it, so the time left reads high by how long that answer took: seconds, against an
// hour.
func transcriptPromptCache(path string, now time.Time) *promptCache {
	lines := transcriptTail(path)
	var used time.Time
	ttl := ""
	for i := len(lines) - 1; i >= 0 && (used.IsZero() || ttl == ""); i-- {
		line := lines[i]
		if !strings.Contains(line, `"usage"`) {
			continue
		}
		var entry struct {
			Type        string `json:"type"`
			IsSidechain bool   `json:"isSidechain"`
			Timestamp   string `json:"timestamp"`
			Message     struct {
				Usage *struct {
					Write    float64 `json:"cache_creation_input_tokens"`
					Read     float64 `json:"cache_read_input_tokens"`
					Lifetime *struct {
						FiveMinutes float64 `json:"ephemeral_5m_input_tokens"`
						OneHour     float64 `json:"ephemeral_1h_input_tokens"`
					} `json:"cache_creation"`
				} `json:"usage"`
			} `json:"message"`
		}
		if json.Unmarshal([]byte(line), &entry) != nil || entry.Type != "assistant" || entry.IsSidechain ||
			entry.Message.Usage == nil {
			continue
		}
		u := entry.Message.Usage
		if used.IsZero() {
			at, ok := parseTime(entry.Timestamp)
			if !ok || u.Write+u.Read == 0 {
				continue
			}
			used = at
		}
		if l := u.Lifetime; l != nil {
			switch {
			case l.OneHour > 0:
				ttl = "1h"
			case l.FiveMinutes > 0:
				ttl = "5m"
			}
		}
	}
	if used.IsZero() {
		return nil
	}
	lifetime := time.Hour
	if d, err := time.ParseDuration(ttl); err == nil {
		lifetime = d
	}
	expires := used.Add(lifetime)
	at := float64(expires.Unix())
	return &promptCache{Warm: now.Before(expires), CachingObserved: true, TTL: ttl, ExpiresAt: &at}
}
