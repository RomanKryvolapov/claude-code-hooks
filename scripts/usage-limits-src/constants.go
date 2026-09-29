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
const defaultTimeZone = "Europe/Sofia"

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
	barCells     = 20  // one cell per 5 %
	barFilled    = '▓' // every gauge, the context one included
	barEmpty     = '░'
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

// --- colour, for the status line alone ------------------------------------------
//
// Only the status line is painted, and in it only the gauges and the figures. The injected blocks go
// into the model's context, where an escape code is noise, and the JSON is read by programs. Setting
// NO_COLOR (https://no-color.org) to anything turns the paint off there too.
//
// A LIMIT or CONTEXT gauge is drawn whole in one colour, a hue turning from green through yellow to
// red, in 24-bit colour at the saturation and brightness below. What turns it differs:
//
//   - a LIMIT gauge turns with how far the spend runs ahead of the time gauge under it — green while
//     it is not ahead at all, red once it is limitAheadRed points ahead. That is the reading the two
//     gauges exist for: spend ahead of the clock will not last to the reset;
//   - the CONTEXT gauge, which has no clock, turns with how full it is — green empty, red full.
const (
	gaugeHueStart  = 120.0 // green
	gaugeHueEnd    = 0.0   // red
	gaugeSaturated = 0.85
	gaugeBright    = 0.95
)

// limitAheadRed is how many points a LIMIT gauge's spend runs ahead of the time gauge under it before
// the gauge is fully red. Override with "limit_ahead_red_pct" in the config.
var limitAheadRed = 20.0

// A terminal that cannot draw 24-bit colour gets the same scale from the 256-colour palette, green to
// red in eleven steps. Apple's Terminal is the one still in wide use: it could not draw 24-bit colour
// for most of its life, and reads the code as other attributes.
var gaugePalette256 = [11]int{46, 82, 118, 154, 190, 226, 220, 214, 208, 202, 196}

// The figures — the percentages of the limits and of the time gauges under them, the reset times, and
// the figures of a wait — are one amber; a terminal without 24-bit colour gets the nearest the
// 256-colour palette has. The CONTEXT counter and percentage take the CONTEXT gauge's own colour, and
// the prompt cache's time left its own green-to-red scale.
var figuresRGB = [3]int{255, 180, 0}

const figures256 = 214

// The TIME/WORK gauges are cyan, so a limit's two gauges never read alike.
const colourTime = "36"

// --- fitting the status line to the terminal ------------------------------------
//
// Claude Code tells the status line how wide the terminal is. The limit blocks are laid side by side
// while they fit and wrapped onto further rows when they do not; a terminal too narrow for even one
// block at full size gets gauges half as long.

// statusMarginCols is the part of the terminal the status line does not get: Claude Code draws it
// inside a margin of its own that the documentation does not size, so this is an allowance, not a
// measurement.
const statusMarginCols = 4

// narrowBarCells is the limit gauge on a narrow terminal: one cell per started 10 %.
const narrowBarCells = 10

// minContextCells is as short as the CONTEXT gauge gets on a narrow terminal.
const minContextCells = 20

// searchStepCeilingMin bounds the coarse pass that finds the next moment the working rate changes.
// The real step is the shortest stretch the configured week can hold — a working window, or the gap
// between two of them — so that no stretch can be stepped over; this only stops the pass from
// taking needlessly large strides on an ordinary week.
const searchStepCeilingMin = 30 * time.Minute
