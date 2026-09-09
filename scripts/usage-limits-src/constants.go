// constants.go — the values somebody actually changes, in one place.
//
// The working week, the zone it is measured in, the zone thresholds, the windows the API reports
// against, and the shape of the picture. Each is a default: the project's own
// `.claude/usage-limits-config.json` overrides the ones that make sense per project, and the loader
// in main.go says out loud when it cannot use something it was given.
//
// **Not everything numeric lives here.** The HTTP timeout, how much of a transcript is read, how
// many samples make a burn rate trustworthy and how long the gate remembers having asked are
// implementation choices rather than settings, and they stay beside the code that makes them.
//
// Nothing here decides how anything is computed; that is main.go's job.

package main

import "time"

// --- the working week ---------------------------------------------------------
//
// The gauge under a weekly limit measures *working* time, not calendar time, and two things decide
// how much of a day counts: which hours of it are worked at all, and how much of those hours count.
// A weekend spends almost none of the budget while eating two of the seven days, and nobody spends
// any of it at four in the morning — so on the calendar scale the bar ran ahead of the spend every
// Monday and behind it every Friday, and the comparison the two bars exist for said nothing.

// clockMin turns a wall-clock time into minutes from local midnight. `clockMin(24, 0)` is the end of
// the day and is the only way to say "all of it".
func clockMin(hour, minute int) int { return hour*60 + minute }

// workingDay is one day of that week.
type workingDay struct {
	// PercentOfDay is how much of this day's working hours counts as working time: 100 counts them
	// in full, 0 takes the day out of the week entirely, 50 counts it as half a day of work. Only
	// the ratios between the seven matter — the week is their sum.
	PercentOfDay float64
	// FromMin, ToMin are those hours, as minutes from local midnight in the zone below, 0 to 1440.
	// Equal means the day is not worked at all. ToMin **below** FromMin is refused rather than
	// wrapped around midnight: a night shift is a different product decision, and silently
	// accepting one would make the week a shape nobody asked for.
	FromMin, ToMin int
}

// workingWeek is indexed by time.Weekday(), Sunday first — the same order the weekdays table below
// uses. Override per day with "working_week" in the config.
var workingWeek = [7]workingDay{
	{PercentOfDay: 20, FromMin: clockMin(12, 0), ToMin: clockMin(20, 0)},  // Sun
	{PercentOfDay: 100, FromMin: clockMin(12, 0), ToMin: clockMin(20, 0)}, // Mon
	{PercentOfDay: 100, FromMin: clockMin(12, 0), ToMin: clockMin(20, 0)}, // Tue
	{PercentOfDay: 100, FromMin: clockMin(12, 0), ToMin: clockMin(20, 0)}, // Wed
	{PercentOfDay: 100, FromMin: clockMin(12, 0), ToMin: clockMin(20, 0)}, // Thu
	{PercentOfDay: 100, FromMin: clockMin(12, 0), ToMin: clockMin(20, 0)}, // Fri
	{PercentOfDay: 20, FromMin: clockMin(12, 0), ToMin: clockMin(20, 0)},  // Sat
}

// defaultTimeZone is the zone that working week is measured in — where a day starts and ends, and
// therefore which hours are the weekend and which are the night. Written down rather than taken
// from the host clock so the same repository answers the same way wherever it runs: a CI box on
// UTC, a container, a laptop that travelled. A working day is a fact about the person, not about
// the machine.
//
// Any IANA name; override with "time_zone" in the config, and an empty value means "use whatever
// zone this machine is set to". It moves the day and the hours only — every clock time on screen
// stays in the machine's own zone, because that is where the person is reading it.
const defaultTimeZone = "Europe/Kyiv"

// --- the windows the API reports against --------------------------------------
//
// Fixed by the subscription, not by us: the session window is five hours and the weekly one is seven
// days. They are here because the gauges measure against them.
const (
	sessionWindowMin = 5 * 60.0
	weeklyWindowMin  = 7 * 24 * 60.0
)

// --- zone thresholds ----------------------------------------------------------
//
// Defaults, for a binary running without a config file or with one that sets no thresholds. A
// project that wants different ones sets them in its own config, and several here do — so these are
// the floor, not a claim about what any particular project runs. The hook always reports the zone it
// actually computed, which is the number to trust over any document.
var (
	z5 = map[string]float64{"YELLOW": 50, "ORANGE": 80, "RED": 90}
	z7 = map[string]float64{"YELLOW": 80, "ORANGE": 90, "RED": 95}
)

// rank orders the zones so the worst of several can be picked.
var rank = map[string]int{"GREEN": 0, "YELLOW": 1, "ORANGE": 2, "RED": 3}

// weekdays are the short names the reset cell prints, indexed by time.Weekday().
var weekdays = []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}

// --- fetching and sampling ----------------------------------------------------
var (
	// How long a fetched usage document is reused before the API is asked again.
	fetchTTLSec = 60
	// Whether the PreToolUse gate refuses subagent spawns in a hot zone at all.
	gateEnabled = true
	// The context window the CONTEXT gauge measures against.
	contextWindowTokens = 1_000_000.0
)

// How far back a burn-rate sample stays useful. Older ones are dropped rather than averaged in,
// because a rate measured two hours ago says nothing about the next ten minutes.
const sampleMaxAgeMin = 180.0

// --- the picture --------------------------------------------------------------
const (
	barCells     = 20 // one cell per 5 %
	barFilled    = '█'
	barEmpty     = '░'
	ctxFilled    = '▓'    // the context gauge, so it never reads as one of the limit gauges
	ctxBarCells  = 100    // the context gauge: one cell per percent
	cellGap      = "  "   // inside a block: label → number → gauge → reset text
	blockGap     = "    " // between two limit blocks
	statusIndent = ""
)

// The two captions under a limit's own gauge. TIME is the calendar scale the session row is on;
// WORK is the working-time scale the weekly rows move to once the working week differs from a plain
// calendar one. Two gauges measuring different things must not carry the same word.
const (
	captionCalendar = "TIME"
	captionWorking  = "WORK"
)

// searchStepCeilingMin bounds the coarse pass that finds the next moment the working rate changes.
// The real step is the shortest stretch the configured week can hold — a working window, or the gap
// between two of them — so that no stretch can be stepped over; this only stops the pass from
// taking needlessly large strides on an ordinary week.
const searchStepCeilingMin = 30 * time.Minute
