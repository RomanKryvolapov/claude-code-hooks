// modelblock.go — what the model is handed at session start and on every prompt.
//
// The block is data, not instructions: what to do with the figures is the session-budget rule's,
// which the model already carries. So it is written the way Claude reads data best — semantic XML,
// one element per line, every figure in the one format the root element states — and it holds the
// figures and nothing else. There is no gauge in it: a gauge tells a model nothing the number beside
// it does not, and costs its context tokens on every prompt.

package main

import (
	"math"
	"strings"
	"time"
)

var (
	xmlText = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	xmlAttr = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
)

// openTag writes <name a="v" …>, from attribute names and values in turn.
func openTag(name string, attrs ...string) string {
	var b strings.Builder
	b.WriteString("<" + name)
	for i := 0; i+1 < len(attrs); i += 2 {
		b.WriteString(" " + attrs[i] + `="` + xmlAttr.Replace(attrs[i+1]) + `"`)
	}
	b.WriteString(">")
	return b.String()
}

// element writes <name a="v" …/>, or <name …>content</name> when there is content.
func element(name, content string, attrs ...string) string {
	open := openTag(name, attrs...)
	if content == "" {
		return strings.TrimSuffix(open, ">") + "/>"
	}
	return open + xmlText.Replace(content) + "</" + name + ">"
}

// The root's attributes: what the figures cover, and the one format they are all written in.
var blockAttrs = []string{
	"scope", "account-global: every session on this subscription",
	"format", "tokens in k/m/b; clock times local; waits h:mm",
}

// modelBlock is the whole block for a state the usage API — or Claude Code — gave figures for.
// rows come as limitRows lays them out: the 5-hour window, the weekly one, then the per-model
// buckets in the order st.scoped holds them. tokens is nil where there is no session to count.
func modelBlock(st *state, rows []statusRow, ctx contextReading, tokens *tokenStats, sessionStart bool, now time.Time) string {
	attrs := blockAttrs
	if sessionStart {
		attrs = append(append([]string{}, blockAttrs...), "at", "session start")
	}
	lines := []string{openTag("usage_limits", attrs...)}
	if st.stale {
		lines = append(lines, staleElement(st, now))
	}
	if ctx.ok && ctx.window > 0 {
		pct := math.Min(100, ctx.tokens/ctx.window*100)
		lines = append(lines, element("context", "", "used", shortTokens(ctx.tokens)+" of "+shortTokens(ctx.window)+" ("+num(pct, 0)+"%)"))
	}
	lines = append(lines, limitElements(st, rows)...)
	if tokens != nil {
		lines = append(lines, tokensElement(*tokens))
	}
	lines = append(lines, burnElement(st))
	for _, p := range configProblems {
		lines = append(lines, element("config_problem", p))
	}
	if sessionStart && st.activeModel != "" {
		lines = append(lines, modelElement(st))
	}
	lines = append(lines, zoneElement(st))
	// Named only where it exists. Pointing every session at a file the project never installed is an
	// instruction that cannot be followed, planted in the model's context on every start.
	if sessionStart && budgetRuleInstalled() {
		lines = append(lines, element("rule", "plan M/L/XL tasks by it", "path", ".claude/rules/session-budget.md"))
	}
	return strings.Join(append(lines, "</usage_limits>"), "\n")
}

// unavailableBlock is the block when there are no figures at all: only a broken config, which still
// has to be said out loud.
func unavailableBlock() string {
	lines := []string{openTag("usage_limits", "unavailable", "the usage API gave no figures")}
	for _, p := range configProblems {
		lines = append(lines, element("config_problem", p))
	}
	return strings.Join(append(lines, "</usage_limits>"), "\n")
}

// limitElements is one element per limit: how much is used against how much of its window is gone,
// and when it resets.
func limitElements(st *state, rows []statusRow) []string {
	out := make([]string, 0, len(rows))
	for i, r := range rows {
		name := "5-hour"
		switch {
		case i == 1:
			name = "7-day"
		case i >= 2 && i-2 < len(st.scoped):
			name = st.scoped[i-2].Model + " weekly bucket"
		}
		attrs := []string{"name", name, "used", num(r.pct, 0) + "%"}
		if r.windowBar != "" {
			of := "the window"
			if r.windowCaption == captionWorking {
				of = "the working week"
			}
			attrs = append(attrs, "gone", num(r.windowPct, 0)+"% of "+of)
		}
		if r.reset != "" {
			resets := r.reset
			if r.after != "" {
				resets += ", in " + r.after
			}
			attrs = append(attrs, "resets", resets)
		}
		out = append(out, element("limit", "", attrs...))
	}
	return out
}

// tokensElement is the tokens of the session and, where the limits' windows are known, of the
// account on this machine in each.
func tokensElement(t tokenStats) string {
	counts := func(s tokenSums) string {
		return "fresh " + shortTokens(s.fresh()) + ", cached " + shortTokens(s.Read)
	}
	attrs := []string{"session", counts(t.session)}
	if t.account {
		attrs = append(attrs, "account_5h", counts(t.fiveHour), "account_7d", counts(t.weekly))
	}
	if !t.sessionWhole || (t.account && !t.accountWhole) {
		attrs = append(attrs, "counting", "unfinished")
	}
	return element("tokens", "", attrs...)
}

// burnElement is how fast the 5-hour window is filling, where that lands, and the pace against the
// clock — spend divided by the share of the window gone.
func burnElement(st *state) string {
	forecast := "lasts to the reset at " + fmtHHmm(st.r5Local)
	if !math.IsInf(st.minToExhaust, 1) && st.minToExhaust < st.minToReset5 {
		forecast = "reaches the cap in " + formatWait(st.minToExhaust)
	}
	return element("burn", "", "rate", num(st.burn, 2)+"%/min, "+st.burnSource, "forecast", forecast,
		"pace", num(st.paceRatio, 1))
}

// modelElement names the model in use and whether it has a weekly bucket of its own.
func modelElement(st *state) string {
	if st.bucketPct < 0 || st.activeBucket == nil {
		return element("model", "", "id", st.activeModel, "own_bucket", "none; the 7-day limit applies")
	}
	attrs := []string{"id", st.activeModel, "own_bucket", st.activeBucket.Model + ", " + num(st.bucketPct, 0) + "% used"}
	if !math.IsInf(st.minToBucketExhaust, 1) {
		attrs = append(attrs, "bucket_forecast", "reaches its cap in "+formatWait(st.minToBucketExhaust))
	}
	if st.bucketBurn > 0.0001 {
		attrs = append(attrs, "bucket_burn", num(st.bucketBurn, 3)+"%/min, "+st.bucketBurnSource)
	}
	return element("model", "", attrs...)
}

// zoneElement is the zone, what binds it, and what it allows — the part the session-budget rule
// acts on.
func zoneElement(st *state) string {
	return element("zone", zoneAdvice[st.zone]+"."+formatSwitchAdvice(st), "level", st.zone, "limiter", st.limiter)
}

// staleElement says the figures are old, and since when. Written only when they are, so silence
// means fresh; since spend only grows inside a window, an old figure is a floor.
func staleElement(st *state, now time.Time) string {
	if st.fetchedAt.IsZero() {
		return element("stale", "the usage API has not answered")
	}
	when := strings.ToUpper(fmtDddHHmm(st.fetchedAt))
	if y, m, d := st.fetchedAt.Local().Date(); y == now.Local().Year() && m == now.Local().Month() && d == now.Local().Day() {
		when = fmtHHmm(st.fetchedAt)
	}
	return element("stale", "the usage API has not answered since then; every figure here is as of then", "since", when)
}
