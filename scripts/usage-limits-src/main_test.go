package main

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// withZone runs fn with the working week's zone pinned, so a test asserting on day and hour
// boundaries means the same thing on every machine.
func withZone(t *testing.T, name string, fn func()) {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatalf("load zone %s: %v", name, err)
	}
	prevName, prevLoc := timeZoneName, weekLoc
	timeZoneName, weekLoc = name, loc
	defer func() { timeZoneName, weekLoc = prevName, prevLoc }()
	fn()
}

// withWeek runs fn with a given working week, restoring what was there afterwards and clearing the
// cached search stride both ways.
func withWeek(t *testing.T, w [7]workingDay, fn func()) {
	t.Helper()
	prev, prevStep := workingWeek, searchStepMin
	workingWeek, searchStepMin = w, 0
	defer func() { workingWeek, searchStepMin = prev, prevStep }()
	fn()
}

// withoutConfigProblems clears the collected config complaints around fn, since they are a global
// the binary fills once at startup.
func withoutConfigProblems(t *testing.T, fn func()) {
	t.Helper()
	prev := configProblems
	configProblems = nil
	defer func() { configProblems = prev }()
	fn()
}

// day is shorthand for one entry of a test week, on whole hours.
func day(pct float64, fromH, toH int) workingDay {
	return workingDay{PercentOfDay: pct, FromMin: clockMin(fromH, 0), ToMin: clockMin(toH, 0)}
}

func week(days ...workingDay) [7]workingDay {
	var w [7]workingDay
	copy(w[:], days)
	return w
}

var (
	// The shipped week: Sunday first, weekend at a fifth, noon to eight.
	shippedWeek = week(day(20, 12, 20), day(100, 12, 20), day(100, 12, 20), day(100, 12, 20),
		day(100, 12, 20), day(100, 12, 20), day(20, 12, 20))
	// Round the clock at one rate — the only shape that measures what the calendar does.
	calendarWeek = week(day(100, 0, 24), day(100, 0, 24), day(100, 0, 24), day(100, 0, 24),
		day(100, 0, 24), day(100, 0, 24), day(100, 0, 24))
	halfCalendarWeek = week(day(50, 0, 24), day(50, 0, 24), day(50, 0, 24), day(50, 0, 24),
		day(50, 0, 24), day(50, 0, 24), day(50, 0, 24))
	emptyWeek = week(day(0, 12, 20), day(0, 12, 20), day(0, 12, 20), day(0, 12, 20),
		day(0, 12, 20), day(0, 12, 20), day(0, 12, 20))
	// Every day a different percentage and different hours, so misattributing any stretch to a
	// neighbouring day or to the wrong side of an hour changes the total.
	spikyWeek = week(day(11, 1, 5), day(97, 8, 17), day(83, 9, 12), day(71, 0, 24),
		day(59, 13, 14), day(43, 22, 24), day(29, 6, 7))
)

// What the shipped week adds up to: five full days and two at a fifth, of eight hours each.
const shippedWeeklyMinutes = 5*480*1.0 + 2*480*0.2 // 2592

func at(t *testing.T, s string) time.Time {
	t.Helper()
	v, err := time.ParseInLocation("2006-01-02 15:04", s, weekLocation())
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return v
}

func near(got, want, tol float64) bool { return math.Abs(got-want) <= tol }

// bruteForceWeightedMinutes is an independent oracle: minute by minute, it asks which day and hour
// that minute falls on and adds it up. Slow, and deliberately shares none of the boundary search —
// it never asks where a stretch begins or ends, which is the part that can go wrong.
func bruteForceWeightedMinutes(from, to time.Time) float64 {
	zone := weekLocation()
	total := 0.0
	for cur := from; cur.Before(to); cur = cur.Add(time.Minute) {
		step := time.Minute
		if rest := to.Sub(cur); rest < step {
			step = rest
		}
		local := cur.In(zone)
		d := workingWeek[int(local.Weekday())]
		h, m, _ := local.Clock()
		since := 60*h + m
		if d.ToMin > d.FromMin && since >= d.FromMin && since < d.ToMin {
			total += step.Minutes() * d.PercentOfDay / 100
		}
	}
	return total
}

// A week worked round the clock at one rate is exactly the calendar week.
func TestCalendarWeekEqualsCalendar(t *testing.T) {
	withZone(t, "UTC", func() {
		withWeek(t, calendarWeek, func() {
			from, to := at(t, "2026-09-07 18:00"), at(t, "2026-09-14 18:00")
			if got := weightedMinutes(from, to); !near(got, 7*24*60, 0.001) {
				t.Fatalf("round-the-clock week = %v, want %v", got, 7*24*60.0)
			}
		})
	})
}

// An exact seven-day window holds one of every weekday whatever hour it starts at, so the working
// week is the same length wherever the reset happens to fall.
func TestWorkingWeekLengthIsStartIndependent(t *testing.T) {
	withZone(t, "UTC", func() {
		withWeek(t, shippedWeek, func() {
			for _, start := range []string{
				"2026-09-07 18:00", "2026-09-09 03:12", "2026-09-12 23:59", "2026-09-13 00:00",
			} {
				from := at(t, start)
				got := weightedMinutes(from, from.Add(7*24*time.Hour))
				if !near(got, shippedWeeklyMinutes, 0.001) {
					t.Fatalf("week from %s = %v, want %v", start, got, shippedWeeklyMinutes)
				}
			}
		})
	})
}

// Hand-computed spans, so a change to the walk shows up as a wrong number rather than a plausible one.
func TestWeightedMinutesKnownSpans(t *testing.T) {
	withZone(t, "UTC", func() {
		withWeek(t, shippedWeek, func() {
			cases := []struct {
				name     string
				from, to string
				want     float64
			}{
				// Mon 18:00 to Wed 13:40: the end of one working day, a whole one, part of a third.
				{"across working days", "2026-09-07 18:00", "2026-09-09 13:40", 120 + 480 + 100},
				// Fri 18:00 to Sun 18:00: an evening at full rate and two weekend days at a fifth.
				{"across the weekend", "2026-09-11 18:00", "2026-09-13 18:00", 120 + 480*0.2 + 360*0.2},
				// Nothing is worked between eight in the evening and noon.
				{"overnight", "2026-09-07 20:00", "2026-09-08 12:00", 0},
				{"deep night", "2026-09-08 02:00", "2026-09-08 05:30", 0},
				{"part of one working day", "2026-09-08 14:00", "2026-09-08 18:00", 240},
				{"whole day from midnight", "2026-09-08 00:00", "2026-09-09 00:00", 480},
				{"zero length", "2026-09-08 14:00", "2026-09-08 14:00", 0},
				{"reversed", "2026-09-08 18:00", "2026-09-08 14:00", 0},
			}
			for _, c := range cases {
				if got := weightedMinutes(at(t, c.from), at(t, c.to)); !near(got, c.want, 0.001) {
					t.Errorf("%s: got %v, want %v", c.name, got, c.want)
				}
			}
		})
	})
}

// The gauge must stand still overnight — that is the whole point of having working hours.
func TestGaugeDoesNotMoveOvernight(t *testing.T) {
	withZone(t, "Europe/Kyiv", func() {
		withWeek(t, shippedWeek, func() {
			reset := at(t, "2026-09-14 18:00")
			_, evening := weeklyWindowGauge(reset, at(t, "2026-09-08 20:00"))
			_, midnight := weeklyWindowGauge(reset, at(t, "2026-09-09 00:00"))
			_, dawn := weeklyWindowGauge(reset, at(t, "2026-09-09 11:59"))
			if !near(evening, midnight, 0.0001) || !near(evening, dawn, 0.0001) {
				t.Fatalf("gauge moved overnight: 20:00 %v, midnight %v, 11:59 %v", evening, midnight, dawn)
			}
			_, afternoon := weeklyWindowGauge(reset, at(t, "2026-09-09 13:00"))
			if afternoon <= evening {
				t.Fatalf("gauge did not resume in the afternoon: %v then %v", evening, afternoon)
			}
		})
	})
}

// A daylight-saving day is 23 or 25 hours long, and its working hours are still the hours on the
// clock: the afternoon is eight hours whatever the night did.
func TestWorkingHoursSurviveDST(t *testing.T) {
	withZone(t, "America/New_York", func() {
		withWeek(t, shippedWeek, func() {
			// 2026-03-08 springs forward at 02:00 and 2026-11-01 falls back at 02:00, both well
			// before the working day starts.
			for _, d := range []string{"2026-03-08", "2026-11-01"} {
				got := weightedMinutes(at(t, d+" 00:00"), at(t, d+" 00:00").Add(24*time.Hour))
				if !near(got, 480*0.2, 0.001) {
					t.Errorf("%s = %v, want a weekend afternoon at a fifth (%v)", d, got, 480*0.2)
				}
			}
		})
	})
}

// A clock change inside the working hours really does lengthen or shorten that day's working time,
// because the hours are wall-clock ones and the day had more or fewer of them.
func TestClockChangeInsideWorkingHours(t *testing.T) {
	withZone(t, "America/New_York", func() {
		w := week(day(100, 0, 6), day(100, 0, 6), day(100, 0, 6), day(100, 0, 6),
			day(100, 0, 6), day(100, 0, 6), day(100, 0, 6))
		withWeek(t, w, func() {
			spring := weightedMinutes(at(t, "2026-03-08 00:00"), at(t, "2026-03-08 23:00"))
			if !near(spring, 5*60, 0.001) {
				t.Errorf("spring-forward morning = %v, want five hours", spring)
			}
			fall := weightedMinutes(at(t, "2026-11-01 00:00"), at(t, "2026-11-01 23:00"))
			if !near(fall, 7*60, 0.001) {
				t.Errorf("fall-back morning = %v, want seven hours", fall)
			}
		})
	})
}

// The implementation has to agree with the oracle in every zone, including the ones that move their
// clock at exactly midnight — where a local day has no 00:00 at all.
func TestWeightedMinutesAgreesWithBruteForce(t *testing.T) {
	zones := []string{
		"America/Santiago", "America/Havana", // move the clock at 00:00
		"Asia/Beirut", "Africa/Cairo", // midnight-adjacent
		"Australia/Lord_Howe", "Pacific/Chatham", "Asia/Kathmandu", // half-hour and 45-minute offsets
		"America/New_York", "Europe/Kyiv", "Australia/Sydney", "UTC",
	}
	starts := []string{"2026-03-05", "2026-04-02", "2026-09-03", "2026-10-29", "2026-11-01", "2026-06-10"}
	for _, w := range [][7]workingDay{shippedWeek, spikyWeek} {
		withWeek(t, w, func() {
			for _, z := range zones {
				withZone(t, z, func() {
					for _, d := range starts {
						for step := 0; step < 24*4; step++ {
							from := at(t, d+" 00:00").Add(time.Duration(step) * 15 * time.Minute)
							to := from.Add(7 * 24 * time.Hour)
							got, want := weightedMinutes(from, to), bruteForceWeightedMinutes(from, to)
							if !near(got, want, 1.0) {
								t.Fatalf("%s from %s: weighted %v, oracle %v (off by %v minutes)",
									z, from.Format("2006-01-02 15:04 -0700"), got, want, got-want)
							}
						}
					}
				})
			}
		})
	}
}

// offsetChanges finds every instant in the year at which the zone's UTC offset moves, to the minute.
// Scanned rather than declared: which zones move their clock, and when, is a property of the zone
// database this binary carries, not something a test should hard-code and then drift from.
func offsetChanges(t *testing.T, year int, loc *time.Location) []time.Time {
	t.Helper()
	var out []time.Time
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	_, prev := start.In(loc).Zone()
	for h := 1; h <= 366*24; h++ {
		cur := start.Add(time.Duration(h) * time.Hour)
		_, off := cur.In(loc).Zone()
		if off == prev {
			continue
		}
		lo := cur.Add(-time.Hour)
		for m := 1; m <= 60; m++ {
			c := lo.Add(time.Duration(m) * time.Minute)
			if _, o := c.In(loc).Zone(); o != prev {
				out = append(out, c)
				break
			}
		}
		prev = off
	}
	return out
}

// Around a clock change the total must still be the honest integral, whatever the zone does.
func TestWeightedMinutesAroundClockChanges(t *testing.T) {
	zones := []string{
		"America/Santiago", "America/Havana", "Asia/Beirut", "Africa/Cairo",
		"Australia/Lord_Howe", "Pacific/Chatham", "America/New_York", "Europe/Kyiv", "Australia/Sydney",
	}
	for _, w := range [][7]workingDay{spikyWeek, shippedWeek} {
		withWeek(t, w, func() {
			for _, z := range zones {
				withZone(t, z, func() {
					loc := weekLocation()
					changes := offsetChanges(t, 2026, loc)
					if len(changes) == 0 {
						t.Fatalf("%s: expected the zone database to hold a clock change in 2026", z)
					}
					for _, c := range changes {
						for m := -125; m <= 125; m += 7 {
							from := c.Add(time.Duration(m) * time.Minute)
							to := from.Add(30 * time.Hour)
							got, want := weightedMinutes(from, to), bruteForceWeightedMinutes(from, to)
							if !near(got, want, 1.0) {
								t.Fatalf("%s window from %s: weighted %v, oracle %v (off by %v)",
									z, from.In(loc).Format("2006-01-02 15:04 -0700"), got, want, got-want)
							}
						}
					}
				})
			}
		})
	}
}

// A working window shorter than the coarse search stride must not be stepped over.
func TestShortWorkingWindowIsNotSteppedOver(t *testing.T) {
	withZone(t, "Europe/Kyiv", func() {
		var w [7]workingDay
		for i := range w {
			w[i] = workingDay{PercentOfDay: 100, FromMin: clockMin(13, 0), ToMin: clockMin(13, 5)}
		}
		withWeek(t, w, func() {
			from := at(t, "2026-09-07 00:00")
			if got := weightedMinutes(from, from.Add(7*24*time.Hour)); !near(got, 35, 0.001) {
				t.Fatalf("five minutes a day for a week = %v, want 35", got)
			}
			if got := weightedMinutes(at(t, "2026-09-08 13:01"), at(t, "2026-09-08 13:03")); !near(got, 2, 0.001) {
				t.Fatalf("two minutes inside the window = %v, want 2", got)
			}
		})
	})
}

// The gauge itself, against the calendar one it replaces.
func TestWeeklyWindowGauge(t *testing.T) {
	withZone(t, "UTC", func() {
		withWeek(t, shippedWeek, func() {
			reset := at(t, "2026-09-14 18:00")
			now := at(t, "2026-09-09 13:40")
			_, pct := weeklyWindowGauge(reset, now)
			want := (120 + 480 + 100) / shippedWeeklyMinutes * 100
			if !near(pct, want, 0.01) {
				t.Errorf("midweek = %v, want %v", pct, want)
			}
			if _, p := weeklyWindowGauge(reset, reset.Add(-7*24*time.Hour)); !near(p, 0, 0.001) {
				t.Errorf("window start = %v, want 0", p)
			}
			if _, p := weeklyWindowGauge(reset, reset); !near(p, 100, 0.001) {
				t.Errorf("window end = %v, want 100", p)
			}
		})
	})
}

// A week worked round the clock at one rate must be indistinguishable from the calendar gauge it
// replaced — number, bar and caption alike. This is what "leave it as it was" means.
func TestCalendarShapedWeekMatchesCalendarGauge(t *testing.T) {
	withZone(t, "UTC", func() {
		for _, w := range [][7]workingDay{calendarWeek, halfCalendarWeek, emptyWeek} {
			withWeek(t, w, func() {
				reset := at(t, "2026-09-14 18:00")
				for _, s := range []string{"2026-09-08 00:00", "2026-09-09 13:40", "2026-09-12 18:00"} {
					now := at(t, s)
					bar, pct := weeklyWindowGauge(reset, now)
					calBar, cal := windowGauge(reset.Sub(now).Minutes(), weeklyWindowMin)
					if !near(pct, cal, 0.001) || bar != calBar {
						t.Errorf("%+v at %s: weighted %v/%q vs calendar %v/%q", w[0], s, pct, bar, cal, calBar)
					}
				}
				if !workingScaleIsCalendar() {
					t.Errorf("%+v should read as the calendar scale", w[0])
				}
			})
		}
	})
}

// A reset already in the past is a stale cache, and the row keeps its old behaviour: no gauge at all.
func TestWeeklyWindowGaugeSilentOnStaleOrMissingReset(t *testing.T) {
	withZone(t, "UTC", func() {
		withWeek(t, shippedWeek, func() {
			now := at(t, "2026-09-09 13:40")
			if bar, pct := weeklyWindowGauge(now.Add(-time.Minute), now); bar != "" || pct != 0 {
				t.Errorf("past reset = %q/%v, want empty", bar, pct)
			}
			if bar, pct := weeklyWindowGauge(time.Time{}, now); bar != "" || pct != 0 {
				t.Errorf("zero reset = %q/%v, want empty", bar, pct)
			}
		})
	})
}

// The gauge only ever moves forwards, and reaches exactly 100 at the reset.
func TestWeeklyWindowGaugeIsMonotonic(t *testing.T) {
	withZone(t, "Europe/Kyiv", func() {
		withWeek(t, shippedWeek, func() {
			reset := at(t, "2026-09-14 18:00")
			minutes := make([]int, 0, 600)
			for m := 0; m < 7*24*60; m += 17 {
				minutes = append(minutes, m)
			}
			minutes = append(minutes, 7*24*60)
			prev := -1.0
			for _, m := range minutes {
				now := reset.Add(-7 * 24 * time.Hour).Add(time.Duration(m) * time.Minute)
				_, pct := weeklyWindowGauge(reset, now)
				if pct < prev-0.000001 {
					t.Fatalf("minute %d: %v went below %v", m, pct, prev)
				}
				if pct < 0 || pct > 100 {
					t.Fatalf("minute %d: %v outside 0..100", m, pct)
				}
				prev = pct
			}
			if !near(prev, 100, 0.001) {
				t.Fatalf("final = %v, want 100", prev)
			}
		})
	})
}

func TestWorkingScaleIsCalendar(t *testing.T) {
	withWeek(t, shippedWeek, func() {
		if workingScaleIsCalendar() {
			t.Error("noon-to-eight with a lighter weekend is not the calendar scale")
		}
	})
	withWeek(t, week(day(100, 9, 17), day(100, 9, 17), day(100, 9, 17), day(100, 9, 17),
		day(100, 9, 17), day(100, 9, 17), day(100, 9, 17)), func() {
		if workingScaleIsCalendar() {
			t.Error("office hours every day is not the calendar scale")
		}
	})
}

// The zone decides where a day starts and which hours are worked, so the same instant is a working
// afternoon in one zone and the middle of the night in another.
func TestWorkingRateFollowsTheConfiguredZone(t *testing.T) {
	moment := time.Date(2026, 9, 11, 14, 30, 0, 0, time.UTC) // Friday afternoon UTC
	withWeek(t, shippedWeek, func() {
		withZone(t, "UTC", func() {
			if got := workingRateAt(moment); !near(got, 1, 0.001) {
				t.Errorf("UTC: rate %v, want a full working rate", got)
			}
		})
		withZone(t, "Asia/Tokyo", func() { // 23:30 there, long past the working day
			if got := workingRateAt(moment); got != 0 {
				t.Errorf("Tokyo: rate %v, want nothing at half eleven at night", got)
			}
		})
	})
}

func TestDefaultTimeZoneResolves(t *testing.T) {
	loc, err := time.LoadLocation(defaultTimeZone)
	if err != nil {
		t.Fatalf("default zone %q does not load: %v", defaultTimeZone, err)
	}
	if loc == time.UTC && defaultTimeZone != "UTC" {
		t.Fatalf("default zone %q resolved to UTC", defaultTimeZone)
	}
	prevName, prevLoc := timeZoneName, weekLoc
	timeZoneName, weekLoc = defaultTimeZone, nil
	defer func() { timeZoneName, weekLoc = prevName, prevLoc }()
	if got := weekLocation().String(); got != defaultTimeZone {
		t.Fatalf("weekLocation = %q, want %q", got, defaultTimeZone)
	}
}

func TestWeekLocationFallsBack(t *testing.T) {
	prevName, prevLoc := timeZoneName, weekLoc
	defer func() { timeZoneName, weekLoc = prevName, prevLoc }()
	for _, name := range []string{"", "Not/AZone"} {
		timeZoneName, weekLoc = name, nil
		if got := weekLocation(); got != time.Local {
			t.Errorf("zone %q resolved to %v, want the machine zone", name, got)
		}
	}
}

func TestParseClock(t *testing.T) {
	for text, want := range map[string]int{
		"00:00": 0, "12:00": 720, " 9:05 ": 545, "23:59": 1439, "24:00": 1440,
	} {
		if got, ok := parseClock(text); !ok || got != want {
			t.Errorf("parseClock(%q) = %d,%v; want %d,true", text, got, ok, want)
		}
	}
	for _, text := range []string{"", "12", "12:60", "25:00", "24:01", "-1:00", "noon", "12:00:00", "12-00"} {
		if got, ok := parseClock(text); ok {
			t.Errorf("parseClock(%q) = %d,true; want it refused", text, got)
		}
	}
}

func TestWeekdayIndex(t *testing.T) {
	for key, want := range map[string]int{
		"sun": 0, "Sunday": 0, "SUNDAY": 0, " mon ": 1, "Monday": 1,
		"tue": 2, "tuesday": 2, "wed": 3, "thu": 4, "fri": 5, "sat": 6, "Saturday": 6,
		"saturdayy": -1, "satur": -1, "tu": -1, "weekend": -1, "": -1, "monday.": -1,
	} {
		if got := weekdayIndex(key); got != want {
			t.Errorf("weekdayIndex(%q) = %d, want %d", key, got, want)
		}
	}
}

func pctOf(v float64) *float64 { return &v }
func clockOf(s string) *string { return &s }

func TestApplyWorkingWeek(t *testing.T) {
	withoutConfigProblems(t, func() {
		withWeek(t, calendarWeek, func() {
			applyWorkingWeek(map[string]workingDayConfig{
				"Monday":    {Percent: pctOf(50)},                                              // percentage only, hours kept
				"tue":       {From: clockOf("09:00"), To: clockOf("17:30")},                    // hours only
				"WED":       {Percent: pctOf(250), From: clockOf("08:00")},                     // clamped, one bound moved
				"thu":       {To: clockOf("teatime")},                                          // unreadable
				"fri":       {From: clockOf("18:00"), To: clockOf("09:00")},                    // ends before it starts
				"sat":       {Percent: pctOf(-5), From: clockOf("0:00"), To: clockOf("24:00")}, // clamped the other way
				"xyz":       {Percent: pctOf(77)},                                              // not a weekday
				"saturdayy": {Percent: pctOf(44)},                                              // a typo that must not bind
			})
			if w := workingWeek[1]; w.PercentOfDay != 50 || w.FromMin != 0 || w.ToMin != 1440 {
				t.Errorf("Monday = %+v, want half rate round the clock", w)
			}
			if w := workingWeek[2]; w.PercentOfDay != 100 || w.FromMin != 540 || w.ToMin != 1050 {
				t.Errorf("Tuesday = %+v, want 09:00 to 17:30 at full rate", w)
			}
			if w := workingWeek[3]; w.PercentOfDay != 100 || w.FromMin != 480 || w.ToMin != 1440 {
				t.Errorf("Wednesday = %+v, want a clamped rate from 08:00", w)
			}
			if w := workingWeek[4]; w.FromMin != 0 || w.ToMin != 1440 {
				t.Errorf("Thursday = %+v, want the unreadable bound ignored", w)
			}
			if w := workingWeek[5]; w.ToMin != w.FromMin {
				t.Errorf("Friday = %+v, want a backwards range to count as no work", w)
			}
			if w := workingWeek[6]; w.PercentOfDay != 0 {
				t.Errorf("Saturday = %+v, want a negative rate clamped to nothing", w)
			}
			for _, want := range []string{"xyz", "saturdayy", "teatime", "ends before it starts"} {
				found := false
				for _, p := range configProblems {
					if strings.Contains(p, want) {
						found = true
					}
				}
				if !found {
					t.Errorf("%q passed without a word: %v", want, configProblems)
				}
			}
		})
	})
}

// Two keys naming the same day is a contradiction, and resolving it by ranging a map moved the
// gauge between runs. The winner must be the same every time, and it must be said out loud.
func TestApplyWorkingWeekIsDeterministicOnDuplicateKeys(t *testing.T) {
	cfg := map[string]workingDayConfig{
		"mon": {Percent: pctOf(0)}, "Monday": {Percent: pctOf(100)},
		"sat": {Percent: pctOf(20)}, "Saturday": {Percent: pctOf(0)},
	}
	var first [7]workingDay
	var firstProblems []string
	for run := 0; run < 40; run++ {
		withoutConfigProblems(t, func() {
			withWeek(t, calendarWeek, func() {
				applyWorkingWeek(cfg)
				if run == 0 {
					first, firstProblems = workingWeek, append([]string(nil), configProblems...)
					return
				}
				if workingWeek != first {
					t.Fatalf("run %d gave %+v, run 0 gave %+v", run, workingWeek, first)
				}
				if len(configProblems) != len(firstProblems) {
					t.Fatalf("run %d reported %v, run 0 reported %v", run, configProblems, firstProblems)
				}
			})
		})
	}
	// Sorted order decides, so the capitalised spelling wins over the three-letter one.
	if first[1].PercentOfDay != 100 || first[6].PercentOfDay != 0 {
		t.Fatalf("week = %+v; want Monday 100 and Saturday 0", first)
	}
	if len(firstProblems) != 2 {
		t.Fatalf("problems = %v, want one per duplicated day", firstProblems)
	}
	for _, want := range []string{"Monday", "Saturday"} {
		found := false
		for _, p := range firstProblems {
			if strings.Contains(p, want) && strings.Contains(p, strconv.Quote(strings.ToLower(want[:3]))) {
				found = true
			}
		}
		if !found {
			t.Errorf("%s was resolved without naming both keys: %v", want, firstProblems)
		}
	}
}

// The weekly rows must not carry the same caption as the calendar gauge once the two measure
// different things — and must keep it when they do not.
func TestLimitRowsCaption(t *testing.T) {
	now := time.Now()
	reset7 := now.Add(72 * time.Hour)
	st := &state{
		s5: 20, s7: 30,
		r5Local: now.Add(2 * time.Hour), r7Local: reset7,
		minToReset5: 120, minToReset7: 72 * 60,
		scoped: []scopedBucket{{Model: "Fable", Percent: 5, ResetsAt: reset7.Format(time.RFC3339)}},
	}
	withWeek(t, shippedWeek, func() {
		rows := limitRows(st)
		if rows[0].windowCaption != "" {
			t.Errorf("session row caption = %q, want the default", rows[0].windowCaption)
		}
		for _, i := range []int{1, 2} {
			if rows[i].windowCaption != captionWorking {
				t.Errorf("row %d caption = %q, want %q", i, rows[i].windowCaption, captionWorking)
			}
			if rows[i].windowBar == "" {
				t.Errorf("row %d has no gauge", i)
			}
		}
	})
	withWeek(t, calendarWeek, func() {
		for _, r := range limitRows(st) {
			if r.windowCaption != "" {
				t.Errorf("a calendar-shaped week should leave every caption alone, got %q", r.windowCaption)
			}
		}
	})
}

// --- the config loader --------------------------------------------------------
//
// This is the only place the hook reads its settings, and it was rewritten twice; a regression here
// changes every number on screen at once, so it is exercised end to end from a real file.

// withConfigFile writes a config into a throwaway project and loads it, restoring every global the
// loader touches. The environment variable is what the binary itself resolves the project from.
func withConfigFile(t *testing.T, body string, fn func()) {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".claude", "usage-limits-config.json"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	prev := struct {
		ttl      int
		gate     bool
		ctx      float64
		zone     string
		loc      *time.Location
		week     [7]workingDay
		step     int
		problems []string
		z5, z7   map[string]float64
	}{fetchTTLSec, gateEnabled, contextWindowTokens, timeZoneName, weekLoc, workingWeek, searchStepMin,
		configProblems, copyThresholds(z5), copyThresholds(z7)}
	defer func() {
		fetchTTLSec, gateEnabled, contextWindowTokens = prev.ttl, prev.gate, prev.ctx
		timeZoneName, weekLoc = prev.zone, prev.loc
		workingWeek, searchStepMin, configProblems = prev.week, prev.step, prev.problems
		z5, z7 = prev.z5, prev.z7
	}()
	configProblems = nil
	t.Setenv("CLAUDE_PROJECT_DIR", dir)
	loadConfig()
	fn()
}

func copyThresholds(m map[string]float64) map[string]float64 {
	out := map[string]float64{}
	for k, v := range m {
		out[k] = v
	}
	return out
}

func problemsMentioning(sub string) bool {
	for _, p := range configProblems {
		if strings.Contains(p, sub) {
			return true
		}
	}
	return false
}

// Everything the config can set, set — and read back off the globals the rest of the hook uses.
func TestLoadConfigAppliesEverySetting(t *testing.T) {
	withConfigFile(t, `{
      "fetch_ttl_sec": 5,
      "gate_enabled": false,
      "context_window_tokens": 200000,
      "time_zone": "Asia/Tokyo",
      "working_week": { "mon": { "percent": 40, "from": "09:00", "to": "17:00" } },
      "zones": { "session": { "YELLOW": 11, "ORANGE": 12, "RED": 13 },
                 "weekly":  { "YELLOW": 21, "ORANGE": 22, "RED": 23 } }
    }`, func() {
		if fetchTTLSec != 5 || gateEnabled || contextWindowTokens != 200000 {
			t.Errorf("ttl %v, gate %v, context %v", fetchTTLSec, gateEnabled, contextWindowTokens)
		}
		if timeZoneName != "Asia/Tokyo" || weekLocation().String() != "Asia/Tokyo" {
			t.Errorf("zone %q resolved to %v", timeZoneName, weekLocation())
		}
		if w := workingWeek[1]; w.PercentOfDay != 40 || w.FromMin != 540 || w.ToMin != 1020 {
			t.Errorf("Monday = %+v", w)
		}
		if z5["YELLOW"] != 11 || z5["RED"] != 13 || z7["YELLOW"] != 21 || z7["RED"] != 23 {
			t.Errorf("thresholds %v %v", z5, z7)
		}
		if len(configProblems) != 0 {
			t.Errorf("a good config complained: %v", configProblems)
		}
	})
}

// A JSON null must leave the setting alone and say so, never apply the zero it decodes to. Setting
// the cache lifetime to zero this way made the hook call the API on every prompt, and setting the
// gate to false turned off the only thing that refuses a spawn in a hot zone.
func TestLoadConfigRefusesNulls(t *testing.T) {
	withConfigFile(t, `{
      "fetch_ttl_sec": null, "gate_enabled": null, "context_window_tokens": null,
      "time_zone": null, "working_week": null, "zones": null
    }`, func() {
		if fetchTTLSec != 60 || !gateEnabled || contextWindowTokens != 1_000_000.0 {
			t.Errorf("a null overwrote a default: ttl %v, gate %v, context %v",
				fetchTTLSec, gateEnabled, contextWindowTokens)
		}
		if timeZoneName != defaultTimeZone {
			t.Errorf("a null zone left %q", timeZoneName)
		}
		if workingWeek != shippedDefaultWeek() {
			t.Errorf("a null week left %+v", workingWeek)
		}
		for _, name := range []string{"fetch_ttl_sec", "gate_enabled", "time_zone", "working_week", "zones"} {
			if !problemsMentioning(name) {
				t.Errorf("%s was nulled without a word: %v", name, configProblems)
			}
		}
	})
}

// shippedDefaultWeek is the table constants.go declares, for tests that need to prove nothing moved.
func shippedDefaultWeek() [7]workingDay {
	return week(day(20, 12, 20), day(100, 12, 20), day(100, 12, 20), day(100, 12, 20),
		day(100, 12, 20), day(100, 12, 20), day(20, 12, 20))
}

// One value of the wrong type costs that value and nothing else — this is what the whole loader was
// rewritten for, and it is checked one neighbour at a time.
func TestLoadConfigIsolatesOneBadSetting(t *testing.T) {
	bad := map[string]string{
		"fetch_ttl_sec":         `"often"`,
		"gate_enabled":          `"yes"`,
		"context_window_tokens": `"lots"`,
		"time_zone":             `12`,
		"working_week":          `"weekdays"`,
		"zones":                 `"strict"`,
	}
	for field, value := range bad {
		body := `{"fetch_ttl_sec": 5, "gate_enabled": false, "context_window_tokens": 200000,
                  "time_zone": "Asia/Tokyo",
                  "working_week": {"mon": {"percent": 40}},
                  "zones": {"session": {"YELLOW": 11}}}`
		body = strings.Replace(body, `"`+field+`": `+map[string]string{
			"fetch_ttl_sec": "5", "gate_enabled": "false", "context_window_tokens": "200000",
			"time_zone": `"Asia/Tokyo"`, "working_week": `{"mon": {"percent": 40}}`,
			"zones": `{"session": {"YELLOW": 11}}`,
		}[field], `"`+field+`": `+value, 1)
		withConfigFile(t, body, func() {
			if !problemsMentioning(field) {
				t.Errorf("%s: bad value passed without a word: %v", field, configProblems)
			}
			// Every other setting still applied.
			if field != "fetch_ttl_sec" && fetchTTLSec != 5 {
				t.Errorf("%s took fetch_ttl_sec with it (%v)", field, fetchTTLSec)
			}
			if field != "gate_enabled" && gateEnabled {
				t.Errorf("%s took gate_enabled with it", field)
			}
			if field != "context_window_tokens" && contextWindowTokens != 200000 {
				t.Errorf("%s took context_window_tokens with it (%v)", field, contextWindowTokens)
			}
			if field != "time_zone" && timeZoneName != "Asia/Tokyo" {
				t.Errorf("%s took time_zone with it (%q)", field, timeZoneName)
			}
			if field != "working_week" && workingWeek[1].PercentOfDay != 40 {
				t.Errorf("%s took working_week with it (%+v)", field, workingWeek[1])
			}
			if field != "zones" && z5["YELLOW"] != 11 {
				t.Errorf("%s took zones with it (%v)", field, z5)
			}
		})
	}
}

// One bad threshold costs that threshold, not all six, and a name that is not a zone is named.
func TestLoadConfigIsolatesOneBadThreshold(t *testing.T) {
	withConfigFile(t, `{"zones": {"session": {"YELLOW": "11", "ORANGE": 12, "AMBER": 5},
                                  "weekly": {"RED": 99}}}`, func() {
		if z5["ORANGE"] != 12 || z7["RED"] != 99 {
			t.Errorf("a bad threshold took its neighbours with it: %v %v", z5, z7)
		}
		if z5["YELLOW"] != 50 {
			t.Errorf("the bad threshold was applied anyway: %v", z5)
		}
		for _, want := range []string{"YELLOW", "AMBER"} {
			if !problemsMentioning(want) {
				t.Errorf("%s passed without a word: %v", want, configProblems)
			}
		}
	})
}

// A file that really is malformed still says so, and nothing it held is applied.
func TestLoadConfigReportsMalformedFile(t *testing.T) {
	withConfigFile(t, `{"gate_enabled": false,`, func() {
		if !gateEnabled {
			t.Error("a malformed file was partly applied")
		}
		if !problemsMentioning("not valid JSON") {
			t.Errorf("a malformed file passed without a word: %v", configProblems)
		}
	})
}

// The key the working week replaced must be named rather than silently unread, or a config that
// still carries it loses its weekend weighting without a word.
func TestLoadConfigNamesTheSupersededKey(t *testing.T) {
	withConfigFile(t, `{"week_day_weights": {"sat": 20}}`, func() {
		if !problemsMentioning("week_day_weights") {
			t.Errorf("the superseded key passed without a word: %v", configProblems)
		}
	})
}

// An unloadable zone is named, and the day boundaries fall back to the machine's own zone.
func TestLoadConfigReportsUnloadableZone(t *testing.T) {
	withConfigFile(t, `{"time_zone": "Mars/Olympus"}`, func() {
		if weekLocation() != time.Local {
			t.Errorf("an unloadable zone resolved to %v", weekLocation())
		}
		if !problemsMentioning("Mars/Olympus") {
			t.Errorf("an unloadable zone passed without a word: %v", configProblems)
		}
	})
}

// A percentage outside the scale is clamped, and clamping is said out loud — a config asking for 500
// is asking for something it will not get.
func TestLoadConfigReportsClampedPercent(t *testing.T) {
	withConfigFile(t, `{"working_week": {"mon": {"percent": 500}, "tue": {"percent": -5}}}`, func() {
		if workingWeek[1].PercentOfDay != 100 || workingWeek[2].PercentOfDay != 0 {
			t.Errorf("not clamped: %+v %+v", workingWeek[1], workingWeek[2])
		}
		if !problemsMentioning("500") || !problemsMentioning("-5") {
			t.Errorf("clamping passed without a word: %v", configProblems)
		}
	})
}

// The config problems must survive into the JSON document, which is the only machine-readable way
// to see them.
func TestConfigProblemsReachTheJSONDocument(t *testing.T) {
	withConfigFile(t, `{"working_week": {"weekend": {"percent": 20}}}`, func() {
		if len(configProblems) == 0 {
			t.Fatal("expected a complaint about an unknown day")
		}
		b, err := json.Marshal(struct {
			ConfigProblems []string `json:"config_problems"`
		}{configProblems})
		if err != nil || !strings.Contains(string(b), "weekend") {
			t.Fatalf("problems did not survive encoding: %s (%v)", b, err)
		}
	})
}

// A day written with the wrong field names decodes into an empty struct without error and changes
// nothing at all — the exact silence the CONFIG line exists to end.
func TestLoadConfigNamesUnknownDayFields(t *testing.T) {
	withConfigFile(t, `{"working_week": {"mon": {"pct": 50, "start": "09:00", "to": "17:00"}}}`, func() {
		if w := workingWeek[1]; w.PercentOfDay != 100 || w.FromMin != clockMin(12, 0) {
			t.Errorf("Monday = %+v, want the defaults left alone", w)
		}
		if workingWeek[1].ToMin != clockMin(17, 0) {
			t.Errorf("the one good field was not applied: %+v", workingWeek[1])
		}
		for _, want := range []string{"pct", "start"} {
			if !problemsMentioning(want) {
				t.Errorf("%q passed without a word: %v", want, configProblems)
			}
		}
		// Exactly the two bad names, and no more — the message itself lists the valid ones, so
		// searching for them in the text proves nothing.
		if len(configProblems) != 2 {
			t.Errorf("problems = %v, want one per unknown field", configProblems)
		}
	})
}

// A threshold outside 0..100 is a typo, not a strict setting: below zero pins the zone on for good
// and denies every spawn from then on.
func TestLoadConfigRefusesThresholdsOutsideTheScale(t *testing.T) {
	withConfigFile(t, `{"zones": {"session": {"YELLOW": 5000, "ORANGE": -20, "RED": 91}}}`, func() {
		if z5["YELLOW"] != 50 || z5["ORANGE"] != 80 {
			t.Errorf("an out-of-scale threshold was applied: %v", z5)
		}
		if z5["RED"] != 91 {
			t.Errorf("a good threshold was taken down with them: %v", z5)
		}
		if !problemsMentioning("5000") || !problemsMentioning("-20") {
			t.Errorf("out-of-scale thresholds passed without a word: %v", configProblems)
		}
	})
}

// Two numeric settings whose out-of-range values used to be dropped or applied in silence. A
// negative cache lifetime makes every document look expired, so the API is called on every prompt.
func TestLoadConfigRefusesOutOfRangeNumbers(t *testing.T) {
	withConfigFile(t, `{"context_window_tokens": 0, "fetch_ttl_sec": -1}`, func() {
		if contextWindowTokens != 1_000_000.0 {
			t.Errorf("a zero context window was applied: %v", contextWindowTokens)
		}
		if fetchTTLSec != 60 {
			t.Errorf("a negative cache lifetime was applied: %v", fetchTTLSec)
		}
		if !problemsMentioning("context_window_tokens") || !problemsMentioning("fetch_ttl_sec") {
			t.Errorf("both passed without a word: %v", configProblems)
		}
	})
}

// The session-start block points at the behaviour rule, and must only do so where the project
// actually carries it — an instruction that cannot be followed is worse than none.
func TestBudgetRuleInstalled(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", dir)
	if budgetRuleInstalled() {
		t.Error("reported installed with no rules folder at all")
	}
	if err := os.MkdirAll(filepath.Join(dir, ".claude", "rules"), 0o755); err != nil {
		t.Fatal(err)
	}
	if budgetRuleInstalled() {
		t.Error("reported installed with an empty rules folder")
	}
	if err := os.WriteFile(filepath.Join(dir, ".claude", "rules", "session-budget.md"), []byte("# rule\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if !budgetRuleInstalled() {
		t.Error("did not see the rule that is there")
	}
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	if budgetRuleInstalled() {
		t.Error("claimed a rule with no project to look in")
	}
}
