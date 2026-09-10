# A whole `.claude/` for a project

Everything [Claude Code](https://claude.com/claude-code) reads out of a repository, ready to copy into
one: **three hooks** shipped as committed static binaries, **nine behaviour rules**, **one reviewing
subagent**, **one skill about writing prompts**, and the `settings.json` that wires it all together.

**There is no runtime to install** — no Node, no Python, no Go. The hooks are single static binaries,
committed for macOS on Apple Silicon, Linux on x86-64 and Windows on x86-64, with the Go sources they
were built from beside them. Everything else is Markdown and JSON.

The three groups — the hooks, the rules with their subagent, and the skill — do not need each other.
Take any one of them and leave the rest. **Inside** the rules, though, the files reference one
another by name, so they are installed as a set rather than picked one at a time;
[How the parts connect](#how-the-parts-connect) says exactly what depends on what.

---

## What is in here

| Component                                                                       | Where it lives                             | What it is                                                                                                                         |
| ------------------------------------------------------------------------------- | ------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------- |
| [`usage-limits`](#usage-limits--your-subscription-budget-in-front-of-the-model)  | `.claude/hooks/`, `scripts/usage-limits-*` | Live Anthropic subscription usage in the model's context and the status line, plus a gate that refuses subagent spawns near a limit |
| [`work-audit`](#work-audit--an-append-only-record-of-the-session)                | `.claude/hooks/`, `scripts/work-audit-*`   | An append-only log of what actually happened in each session, written mechanically rather than summarised by the model              |
| [`play-sound`](#play-sound--a-sound-when-it-wants-you-back)                      | `.claude/hooks/`, `scripts/play-sound-*`   | A sound when Claude Code stops, fails, or is waiting on you                                                                        |
| [The nine rules](#the-rules)                                                     | `.claude/rules/`                           | How the model works: budget, starting, finishing, autonomy, architecture, code quality, merges, answers, searching                  |
| [`change-reviewer`](#the-change-reviewer-subagent)                               | `.claude/agents/`                          | A grading pass over a change that is deliberately not told what its author intended                                                |
| [`dev-ai-prompt-generation`](#the-dev-ai-prompt-generation-skill)                | `.claude/skills/`                          | A hub plus 30 reference files on writing and reviewing prompts, skills, rules and `CLAUDE.md`                                      |
| [`settings.json`](#settingsjson--what-is-wired-to-what)                          | `.claude/`                                 | The wiring: which hook runs on which Claude Code event, and the status line                                                        |
| [`usage-limits-config.json`](#configuration)                                     | `.claude/`                                 | The one file you are meant to edit: working week, time zone, zone thresholds                                                       |
| [`.gitattributes` and `.gitignore`](#gitattributes-and-gitignore)                | repository root                            | LF-only launchers, binaries marked binary, a union merge for the audit log, and what this repository deliberately does not track   |
| [The Go sources](#the-go-sources-and-rebuilding)                                 | `scripts/*-src/`                           | The source of all three hooks, a build script each, and the test suite behind the limits hook's working-week arithmetic            |
| [`AGENTS.md`](AGENTS.md)                                                         | repository root                            | The install procedure, written for Claude Code to follow rather than for a person to read                                          |
| [`CLAUDE.md`](CLAUDE.md)                                                         | repository root                            | The short orientation Claude Code loads when working _on this repository_                                                          |

```
.claude/
  hooks/                            three POSIX launchers, one per hook
  rules/                            nine .md files — loaded into every session automatically
  agents/change-reviewer.md         the reviewing subagent
  skills/dev-ai-prompt-generation/  SKILL.md + references/ (30 files)
  sounds/notify.wav                 what play-sound plays
  settings.json                     events → hooks, and the status line
  usage-limits-config.json          working week, time zone, zone thresholds
scripts/
  usage-limits-*  work-audit-*  play-sound-*   committed binaries, three platforms each
  *-src/                                       Go sources and a build script each; tests for the
                                               limits hook only
management/logs/                    where work-audit writes (this repository ignores it)
```

### How the parts connect

**The rules are a set, not a menu.** Eight of the nine link to each other by name — the revision
rule sends you to the pull rule for its first stage, to the quality rule for its architectural
judgement and to the answer rule for the shape of its report; the autonomy rule and the fork rule are
written as two halves of one idea. Install one alone and the model is handed instructions pointing at
files that are not there, which nothing announces. Only `web-search-when-in-doubt` refers to nothing
else. The subagent belongs to that set too: it applies the quality rule by name.

Across the groups there are four ties, and they are the only ones:

1. **`session-budget` is the other half of `usage-limits`.** The hook supplies numbers; the rule is
   what makes the model act on them. Copied without it, the numbers are decoration.
2. **`finishing-work` launches `change-reviewer` by name.** Install that rule without the subagent
   file and its grading stage silently cannot run.
3. **The gate exempts `change-reviewer` by name**, so the one pass a revision depends on is never the
   spawn refused in a hot zone.
4. **`finishing-work` expects the audit log to be committed with the work** where the project tracks
   it — that log is what `work-audit` writes, and the rule says to check it was staged.

Beyond those, the hooks do not need the rules, the rules do not need the hooks, and the skill needs
nothing at all.

---

# The hooks

All three are the same shape: a small POSIX launcher in `.claude/hooks/`, and a static binary per
platform under `scripts/`. The launcher picks the binary by `uname` and looks in `scripts/`,
`scripts/claude-code/` and `.claude/bin/`, so an existing project's layout is respected.

**A launcher that finds no binary exits 0 in silence.** That is right for a hook — a notification must
never interrupt a session — and it also means a broken install looks exactly like a working one, which
is why [AGENTS.md](AGENTS.md) carries a step that checks the install took.

---

## `usage-limits` — your subscription budget in front of the model

A Claude subscription has three kinds of limit running at once: a **5-hour session window**, a
**7-day window**, and a **per-model weekly bucket** for some models. They are account-global — every
Claude session you have open shares them — and nothing inside a coding session tells you where you
stand. So you either stop early to be safe, or you run into a wall in the middle of something.

This hook reads the real numbers from Anthropic's own usage endpoint and shows them in one picture,
in three places: injected into the model's context at session start and on every prompt (so the model
can size its own work), rendered in the status line, and available as JSON for scripting.

```
CONTEXT  634k/1M   63 %  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
5 HOURS   31 % LIMIT ███████░░░░░░░░░░░░░ RESET AT 15:00      7 DAYS   24 % LIMIT █████░░░░░░░░░░░░░░░ RESET AT MON 18:00          FABLE    0 % LIMIT ░░░░░░░░░░░░░░░░░░░░ RESET AT MON 18:00
-         77 %  TIME ████████████████░░░░ RESET AFTER 1:10             31 %  WORK ███████░░░░░░░░░░░░░ RESET AFTER 5 days 03:46            31 %  WORK ███████░░░░░░░░░░░░░ RESET AFTER 5 days 03:46
BURN  0.16 %/min (sampled) - safe to reset 15:00, pace x0.6
ZONE  GREEN (limiter: weekly) -> spend to task size; subagents only when they clearly save tokens.
```

### Reading the picture

Each limit is a **pair of stacked lines**, and the whole point is to read one against the other.

**Top line — `LIMIT`:** how much of that budget is spent, as a percentage, a 20-cell gauge, and
`RESET AT` plus the clock time the limit resets (a weekday appears once the reset is not today).

**Bottom line — `TIME` or `WORK`:** how much of the window has already gone, drawn with the same
gauge, plus `RESET AFTER` — the wait, clock-style (`3:49` under a day, `6 days 17:25` over one).

**A fuller top bar than bottom bar means the budget is running out faster than the clock**, and will
not last to the reset. When the bottom bar fills, the limit resets.

**`CONTEXT`** on top is not a subscription limit. It is how many tokens the last request carried —
fresh input plus everything written to or read from the prompt cache — against the context window.
It is why a long session eventually gets compacted. Its gauge is one cell per percent and uses a
lighter block so it is never confused with the limit gauges. Subagent turns are skipped; they run in
their own context.

**`BURN`** is how fast the session window is filling: `%/min`, measured from samples when there are
enough of them and estimated from the elapsed pace otherwise, then a forecast — either `safe to
reset <time>` or `~N minutes to cap` — and `pace xN`, the spend divided by the share of the window
gone. Above 1.0 you are ahead of the clock.

**`ZONE`** is the part that changes behaviour: GREEN, YELLOW, ORANGE or RED, whichever is worst
across the session window, the weekly window and the active model's own bucket, plus which of the
three is binding. Knowing the binding limiter is what tells you whether switching model helps.

**`MODEL`** (session start only) names the model in use and whether it has a weekly bucket of its
own, with that bucket's burn and forecast.

### The working week — why the weekly bar is not calendar time

Calendar time is the wrong ruler for a weekly limit. Nobody spends budget at four in the morning,
and a weekend eats two of the seven days while spending almost none of it. On the calendar scale the
bottom bar therefore ran ahead of the spend every Monday and behind it every Friday — and that bar
exists precisely to be compared with the spend.

So the **weekly** gauges measure working time, and two things decide how much of a day counts:

- **which hours are worked at all** — `from` and `to`, wall-clock times;
- **how much those hours count** — `percent`: 100 counts them in full, 0 takes the day out of the
  week entirely, 50 counts it as half a day of work.

```json
"working_week": {
  "mon": { "percent": 100, "from": "12:00", "to": "20:00" },
  "sat": { "percent": 20,  "from": "12:00", "to": "20:00" }
}
```

Shipped as Monday to Friday at 100 and the weekend at 20, noon to eight — a week of **43.2 working
hours** rather than 168. The bar stands still overnight and moves again at noon.

Details worth knowing:

- Only the ratios between the percentages matter; the week is their sum.
- A day given only a `percent` keeps the default hours, and one given only hours keeps the default
  percent.
- `to` equal to `from` means the day is not worked. `to` **earlier** than `from` is refused rather
  than wrapped past midnight — a night shift is a different decision, and quietly accepting one
  would make the week a shape nobody asked for.
- `"24:00"` is accepted as the end of the day, and is the only way to say a whole one.
- Set every day to the same percent from `00:00` to `24:00` and the bar is calendar time again,
  caption included. The rows say `WORK` only while the working week differs from a plain calendar
  one.

The **5-hour row is never weighted** — it is five hours of one day, and weighting it would say
nothing.

**This is a display scale and nothing more.** Every reset time, the `RESET AFTER` countdown, the
burn rate, the forecast, the zones and the gate all stay on real clock time, because real time is
what actually resets.

#### Where a day begins

In one named time zone, written down in the config (`Europe/Sofia` by default) rather than taken from
the host clock — so a CI box on UTC, a container and a laptop that travelled all place the working
day in the same hours. Any IANA name works; an empty value means "use this machine's zone".

The zone database is compiled into the binaries (about 400 KB of their size), because Windows ships
none and a slim Linux image often ships none either. A name this build cannot load falls back to the
machine's own zone and says so.

Clock changes are handled by asking the calendar rather than assuming: a daylight-saving day of 23
or 25 hours counts as the day it really was, an hour repeated when a clock goes back is counted under
the day it repeats on, and zones that move their clock at exactly midnight (Chile, Cuba — where a
local day has no `00:00` at all) are correct too.

### The gate

Beyond showing numbers, the hook is wired to `PreToolUse` for `Agent`, `Task` and `Workflow`, and
can refuse a spawn:

| Zone   | Session (5h) | Weekly & buckets (7d) | What the gate does                                  |
| ------ | ------------ | --------------------- | --------------------------------------------------- |
| GREEN  | <50%         | <80%                  | nothing                                             |
| YELLOW | 50–80%       | 80–90%                | nothing                                             |
| ORANGE | 80–90%       | 90–95%                | asks about the **first** spawn of the session, once |
| RED    | ≥90%         | ≥95%                  | denies the spawn                                    |

It asks **once per session**, not per spawn: a gate that interrupts every time is a gate people
switch off. `change-reviewer` is exempt by name in every zone. The thresholds are configurable; set
`gate_enabled` to `false` to turn the gate off and keep the display.

### Configuration

`.claude/usage-limits-config.json`, every key optional:

| Key                     | Default         | What it sets                                                 |
| ----------------------- | --------------- | ------------------------------------------------------------ |
| `working_week`          | noon–8, wknd 20 | Per weekday: `percent`, `from`, `to` — see above             |
| `time_zone`             | `Europe/Sofia`  | The zone the working week is measured in                     |
| `zones`                 | see the table   | `session` and `weekly` thresholds for YELLOW / ORANGE / RED  |
| `gate_enabled`          | `true`          | Whether the gate refuses spawns at all                       |
| `fetch_ttl_sec`         | `60`            | How long a fetched usage document is reused before re-asking |
| `context_window_tokens` | `1000000`       | What the CONTEXT gauge measures against                      |

Days are keyed by their English name or three-letter abbreviation, in any case.

**Nothing in this config is dropped in silence.** Anything the hook cannot use is named on a
`CONFIG` line under the gauges: an unknown weekday, one day named twice, a setting inside a day that
is not `percent`, `from` or `to`, an unreadable time, a range that runs backwards, a percentage
outside 0–100, a threshold outside 0–100 or under a name that is not a zone, a cache lifetime below
zero, a context window at or below zero, a setting written as `null`, an unloadable time zone, the
superseded `week_day_weights` key, and a file that is not valid JSON. One bad value costs that value
alone — never the rest of the file. The line is absent on a correct config, and it is printed even
when the usage API cannot be reached, which is exactly when somebody is most likely to be editing the
file.

An unknown key at the top level is the one thing passed over without comment: a config may carry
settings for something else.

Every default lives in one place, `scripts/usage-limits-src/constants.go`; the config only overrides
it.

### Modes

| Command                                  | Purpose                                     |
| ---------------------------------------- | ------------------------------------------- |
| `--mode inject --event SessionStart`     | the block injected at session start         |
| `--mode inject --event UserPromptSubmit` | the block injected on every prompt          |
| `--mode gate`                            | the PreToolUse gate for Agent/Task/Workflow |
| `--mode json [--model <id>]`             | the full computed state, as JSON            |
| `--mode statusline`                      | the table alone, for the status line        |

`--mode json` is the one to script against. It carries `session_pct`, `weekly_pct`, `zone`,
`limiter`, `burn_pct_per_min`, `min_to_exhaust`, `horizon_min`, the active model's bucket, every
weekly-scoped bucket, and `config_problems`.

---

## `work-audit` — an append-only record of the session

A model summarising its own session is the least reliable witness available: what it reports is
another generated text, produced by the same process that did the work. This hook writes the record
mechanically instead, from the events Claude Code emits, so the log is evidence rather than a
retelling.

It writes one append-only file per session under `management/logs/`, a folder per developer (from
`git user.name`) and a folder per day inside it. Entries are separated by blank lines and each event
is written in a single append, so parallel sessions never interleave and the file reads in the order
things happened.

Each turn records:

- **the prompt** you sent. A message typed while the turn was still running never reaches a hook as
  an event, so it is recovered from the transcript afterwards and written in its own entry, marked as
  such, immediately before the turn it interrupted.
- **every interactive question and the option you picked**, read from the structured tool result. If
  it cannot be parsed the line says so rather than inventing an answer.
- **the turn itself** — model, effort, duration, tool calls with a breakdown, subagents. Subagents
  are counted from their own start events rather than from tool calls, because one workflow call can
  spawn dozens.
- **the files changed** through Claude's file tools: created, modified or deleted, with added and
  removed line counts and the exact line ranges, computed against the last commit. Changes made
  through shell commands or by subagents are not itemised.
- **the final answer**.

It also logs session start and end with the total duration, subagent start and finish, API errors
and compaction.

**git is optional.** Without it the file statistics degrade to a bare list of paths and everything
else is still written. Two throwaway counters live in the OS temp folder rather than in the
repository; losing them costs a turn its subagent count at worst.

**Whether the log is committed is the target project's decision.** In _this_ repository
`management/` is ignored, because the log belongs to whatever project this gets copied into rather
than to the template. Where a project does track it, the union-merge line in `.gitattributes` keeps
both sides of a merge instead of conflicting — the file is append-only and one per session, so a
union merge is always right.

---

## `play-sound` — a sound when it wants you back

`play-sound` plays `.claude/sounds/notify.wav` when Claude Code stops, fails, or raises a
notification — a permission prompt, an idle prompt, a question. It exists so you can look away while
it works.

Replace `.claude/sounds/notify.wav` with any WAV file to change the sound. On macOS it plays through
`afplay`; on Linux through the first of `pw-play`, `paplay`, `aplay`, `play`, `ffplay` and
`canberra-gtk-play` that exists, which covers PipeWire, PulseAudio, ALSA, SoX, FFmpeg and libcanberra;
on Windows through the built-in .NET sound player. If none is available it exits quietly — a
notification hook must never interrupt a session.

---

## What none of the hooks ever do

- **They never break a session.** Every failure path exits 0: the injection and gate modes stay
  silent, the status line prints `LIMITS -> N/A`, JSON prints an error document. A missing binary, a
  broken config, an expired token, no network — all end the same way.
- **They send nothing anywhere.** The only outbound request is `usage-limits` asking
  `api.anthropic.com`, authenticated with the OAuth token `claude login` already stored in
  `~/.claude/.credentials.json`. No API key is needed and the token is never printed.
- **They write almost nothing.** `usage-limits` writes two files under `~/.claude`: a usage cache and
  a note of whether the gate has asked this session. `work-audit` writes its own log under
  `management/logs/`, plus two throwaway counters in the OS temp folder. `play-sound` writes nothing.

---

# The rules

`.claude/rules/`, nine files, about 1,500 lines. They are the half of this repository that changes
how the model _works_ rather than what it can see: when to stop, when not to, what "finished" means,
what a report has to contain, and what may never be resolved by force.

**Copying a file is the whole installation.** Claude Code discovers every `.md` under
`.claude/rules/` at launch, recursively, and loads it with the same priority as `.claude/CLAUDE.md`.
No import, no line in `CLAUDE.md`, nothing to wire.

**Which also means all of them sit in the context window of every session** — that is what they cost.
A rule that should apply only to part of a codebase can carry a `paths:` frontmatter key with glob
patterns, and then loads only when Claude touches a matching file; none of these nine do, because
none of them are about a particular kind of file.

They are worth reading before installing. A rule takes effect in the next session without anybody
opting in, and somebody who wanted a usage gauge should not discover a rule about merge conflicts by
having it obeyed.

**Take them as a set.** Eight of the nine point at each other by name, so a rule installed on its own
sends the model to files that are not there. Dropping one of them is a deliberate decision, not a
free saving — and `.claude/agents/change-reviewer.md` goes with them, since two of the rules lead
to it.

## `session-budget`

The other half of the limits hook, and the only rule tied to anything else here.

A hook can put numbers in front of the model. It cannot make the model do anything about them —
Claude Code has no built-in notion of a budget, and nothing in the injected block says whether 34% on
a Wednesday is fine or alarming. This rule is that missing half: size effort by the task's own
complexity rather than by what is left, pick the cheapest model and the least thinking that will do,
read context by name instead of scanning, and treat a limit as a wall to stop short of rather than a
budget to fill.

Its sharpest part is about subagents. Each runs its own context window and reloads a bootstrap on
every spawn, so a fan-out costs roughly four to seven times what the same work costs inline — worth
it only for read-heavy exploration across many files, never for editing one or two. It also says
plainly that **reviewing a change is one agent working sequentially, not a fan-out**: the code runs
in an order, and parallel readers lose the single mind holding the whole path from input to result.
The `change-reviewer` pass is the one thing outside the calculation entirely — never weighed, never
economised, never skipped because the change looked small. The single thing that can still stop it is
a hard per-session cap on spawns, and then the revision has to say so and fall back to a cold
self-pass rather than quietly skipping the grading.

Then: what each zone allows, what the binding limiter means for whether switching model would help at
all, why Ultracode and the Workflow tool are the highest-risk spend available — one run can fan out
into dozens of subagents — and when to split a long task and write a checkpoint instead of running
into the wall.

## `before-starting-work`

Every message opens the same way: bring the branch up to date, read what actually arrived, and let it
change the work in hand — before planning, before editing, before answering a question about the
state of the work.

The trigger is deliberately the message rather than the size of the task, because a threshold applied
by feel drifts until it is never met. On a branch cut from another, the parent is merged in too, up
the chain to the working branch — into your own branch only, on a clean tree, never a rebase of
anything already pushed, and never a guess about which branch the parent is.

**A pull nobody read is a pull that did not happen.** The rule spells out what to read for: what
moved under the task itself, what arrived in whatever tracks the work, and what changed in the
statement of what _correct_ means. Then one line saying what came in and what it changes — including
"nothing relevant arrived", which is a finding rather than an omission.

It also covers the awkward cases: a dirty tree handled by comparing file lists first, generated files
regenerated rather than hand-merged, a collision with uncommitted work reported instead of forced
through, and the failure that reads as something it is not — _"repository not found"_ on a private
repository almost always meaning the wrong account rather than a missing repository.

The second half is version currency: a deployed runtime past its vendor's support window is a finding
you act on rather than note, quick dependency bumps are done as part of the task, slow ones are named
with what they would take, and a ceiling somebody set on purpose is respected and raised as a
question.

## `working-autonomously`

A task handed over is yours from beginning to end. Not a plan handed back, not a first half with
questions attached, not a draft awaiting approval.

Two failures, both expensive: stopping to ask what you could have determined turns a delegated task
into a conversation, and stopping halfway with something plausible hands back work that now has to be
inspected — which is what delegating was meant to avoid. So names, structure, patterns, test level,
error shapes and the order of steps are all decided, stated in one line each, and the work carries
on. "Shall I go on?" is not deference; the answer is always yes, because that is what the request
was.

**Only two things stop a turn.** An architectural fork, which belongs to the developer. And something
genuinely needed and unobtainable — and only after looking for it in the project's environment files,
its configuration, its documentation, the live environment and its own sanctioned substitutes, with
the report saying where you looked. What never happens instead is proceeding as if you had it: no
empty test, no "the code looks correct, so it works", no requirement marked done on a path nobody
exercised.

## `architectural-forks`

The opposite brake, and the reason the rule above does not turn into recklessness: some decisions are
an edit, some are a rewrite, and the second kind goes to the developer _before_ any code is written.

The test is deliberately mechanical, because the weighing version fails in one predictable direction —
raising a fork stops the work, so a mind optimising for finishing finds reasons the sign does not
apply. **Four signs, any one of them enough:** something with more than one caller changes behaviour
for a caller other than the one you are serving (answered from a list of callers made _now_, not from
a sense of the risk); the shape of something that outlives the request changes; the behaviour of
something you were not asked to touch changes; or new state appears that outlives the request or that
more than one place writes. Absent all four, it is still a fork when two designs are genuinely
defensible and the choice is hard to reverse or shapes money, security or privacy.

Then how to present it: in plain language a competent developer outside the project can act on —
background, the question in one sentence, what depends on the answer, every real option including the
hybrid and the do-nothing, each costed to build, to live with and to reverse, **and a
recommendation**, because a fork presented without one pushes the whole analysis back onto the
developer.

And the half that keeps the rule usable: what _not_ to ask about. Reversible within a day, already
settled by precedent, or acceptable either way — decide it yourself and say so in one line. Expect a
handful of real forks in a project, not one per task.

## `code-quality`

Every change is judged the way an architect would judge it: not "does the symptom go away" but "is
this what this code should look like now".

The first half is about causes. When something misbehaves, find the mechanism that produces it before
designing a fix, and never ship the fast answer instead: a special case for the input from the bug
report, a retry or a delay that makes a race look solved, a swallowed error, a widened type, a second
source of truth kept in sync by hand. Where the clean fix genuinely is out of reach, say so where it
sits and in the report — a temporary measure nobody knows about is a permanent one.

Its middle section is about prompts, and is the part most worth reading even if you take nothing
else. **When a model gets it wrong, the model is not the cause** — it was handed the wrong data, or
the right data in the wrong form: a bad order, a bad name, the wrong block, a missing fact it was left
to infer, two instructions that contradict each other. The symptom fix is another sentence in the
prompt, which is cheap, looks like work, and often moves the number on the next run — and that is
exactly what makes it dangerous. Prompts do not fail loudly when they get too heavy; they get
unstable, the same instruction wins on Tuesday and loses on Thursday, and nobody can say which of the
forty lines is doing the damage. Diagnose from what actually reached the model, in send order, and
fix it there. A worked example runs through a scripted interview that kept losing its last question,
where the cause was the running order rather than anything the instructions said.

Then the two opposite failures — under-designed (special cases, duplicated logic, names that lie,
state nobody owns) and over-designed (an abstraction with one implementation, an extension point for
a variation nobody has, indirection you must trace through four files) — SOLID applied as judgement
rather than ritual, when to stop patching and rewrite, and how to settle a genuine tie: from the
user's side, on transparency, simplicity and predictability, never on a guess about what somebody
would prefer.

## `finishing-work`

One thorough revision before the end of **any turn in which files changed** — whether or not the
message mentions them, whatever the size of the change, docs and one-liners included. The trigger is
something anyone can check from outside, which is the point: a threshold you are allowed to apply
yourself is one you will eventually apply to something that did not qualify.

It is one pass, not many, because repetition is not verification. Reading the same diff a fourth time
finds the fourth wording problem, not the first real defect. What finds defects is doing three things
once, properly: reading the change in the order the code actually executes, deliberately constructing
what could go wrong instead of waiting for it to announce itself, and **running it with real data**.

Seven stages: pull first; read every hunk and then re-read in execution order; interrogate each change
for what it assumes and who else uses it; write the failure hypotheses down and check them; run it and
watch the effects rather than the return value, reading the output rather than the exit code; judge it
as an architect against `code-quality`; and grade it once with `change-reviewer`, which is given the
diff and the requirement and nothing else.

It ends in a **scoped verdict** — what was run, what was only read, and which classes of defect the
gates that ran cannot see. "PASS" on its own is a mood, not a verdict. Findings are fixed rather than
parked; the fixes are run.

And the line that matters most in daily use: **no commit, push, pull request or merge without an
explicit request in the user's own words.** Finishing the work is not permission, and a green build is
not permission.

## `merge-conflicts`

A conflict is two people's intentions meeting, not two blocks of text disagreeing. Every resolution
decides whose work survives, and a wrong one reinstates a solution somebody already replaced — which
is how a fixed bug comes back long after anyone is watching for it. The damage is invisible: a bad
resolution leaves a clean-looking merge with nothing for a reviewer to catch.

So the bar is higher here than anywhere else. **Nothing is forced** — no force push (with or without
a lease), no hard reset of work you did not create, no taking one side wholesale unread, no strategy
option that discards a side, no branch deleted to escape a conflict. Believing one of those is the
right answer is itself a fork.

Before resolving a hunk you must be able to say, in one sentence each, what both sides were trying to
achieve — found from their commits, not from the markers — and which is newer, established from
history rather than assumed. **The default answer to a conflict is "both".** Generated files run the
other way: restored to one side, then regenerated, never hand-merged. Afterwards, each side's
behaviour is checked to still work, because compiling is not evidence.

## `response-style`

The shape of everything the developer actually reads.

Answer in the language of the message — all of it, thinking included. Write for a competent developer
who is _not_ deep in this project, does not have the source open, and will not decode identifiers,
paths or snippets to follow you. That means plain prose by default and no code identifiers, no
`file:line`, no snippets, no multi-row tables unless they were asked for.

Two hard-learned rules about ending a turn: **look for it before asking for it** — every profile,
account, env file, config and runbook, not the first two, and where something genuinely is missing,
say where you looked — and **finish the task rather than handing it back at the halfway point**.

Then the three report shapes. The **work report** at the end of any turn that changed files: what was
done, how solid it is (which claims were run, which read, which assumed; who else the change touched;
one sentence of "this is wrong if…"), what is left with each item on its own line carrying its own
context, and what is needed from the developer. The **problem report** when a review turns something
up: the ground first, then one finding at a time, each saying how the thing is meant to work, what
goes wrong, and what the user actually gets. And the **prompt-composition answer**, the one shape that
overrides the ban on identifiers and tables: every message that goes on the wire, in send order, one
table per message, nothing merged, omitted or tidied.

## `web-search-when-in-doubt`

The shortest rule here, and the one that prevents the most wasted work.

When reality diverges from what you know — a tool behaving unexpectedly, a version-sensitive API, a
flag that does not exist — search before applying another fix. The failure pattern it exists against
is the familiar one: something does not work, a fix based on recollection is applied, it still does
not work, and each new fix digs deeper into a wrong assumption. The vendor's own documentation for
the version actually in use is the authority; forum posts and recollection are only ways of finding
it faster. If the search fails too, stop and ask rather than guessing down a false path.

---

# The `change-reviewer` subagent

`.claude/agents/change-reviewer.md` — the independent grading pass `finishing-work` requires at its
last stage.

**It receives the diff and the requirement and nothing else.** No access to the author's reasoning,
which is the whole design: an author who knows the intent reads the code as the intent and cannot see
the gap between the two. A reviewer with only the code can.

It is told to **refute rather than confirm** — a review ending "looks good" has usually failed to
look — and, if it genuinely cannot break the change, to say so and list what it tried, which reads
very differently from an unexamined approval. It reads in execution order, judges against the
requirement rather than against taste, hunts down every other consumer of what the diff touched (the
highest-yield check there is, and the defect class that ships most often), builds failure hypotheses
instead of waiting for them, and runs the code. It may not edit anything — no fixes, no staging, no
formatting runs. It returns findings worst-first, plus what it could not check.

**It works sequentially, and one pass, not a panel.** Depth is the whole value it adds; four skims in
parallel would lose it.

**The limits hook knows about it by name.** The gate exempts `change-reviewer` in every zone, so the
one pass a revision depends on is never the spawn refused when the budget is tight.

---

# The `dev-ai-prompt-generation` skill

`.claude/skills/dev-ai-prompt-generation/` — a hub `SKILL.md` plus 30 reference files, about 2,600
lines in total, on writing and reviewing prompts, Agent Skills, rules and `CLAUDE.md`.

The hub carries what applies to everything you write — pick conventions by the artifact rather than by
habit (a Claude Code skill or rule is written for Claude; an application prompt follows the vendor it
will actually ship on), how progressive disclosure loads a skill, and three notes that override older
prompting folklore: frontier reasoning models want an effort knob rather than hand-written
chain-of-thought and over-trigger on force language of the "CRITICAL / you MUST" kind; vendor
portability is not assumed; and the accuracy percentages that float around prompt advice are not
load-bearing — measure on your own evals.

The references go deep, one topic per file: structure and delimiters, XML and JSON data presentation,
rule-description styles from pseudocode to decorators, output contracts and structured outputs,
few-shot and example ordering, reasoning-model effort, context engineering and compaction, prompt
caching and cost levers, evals and LLM-as-judge gates, injection defences, RAG grounding and
citations, agentic tool-use prompting, skill frontmatter and lifecycle, and cross-vendor conventions.

**Unlike a rule, a skill costs nothing until it is used.** Only its name and description sit in
context; the body loads when the model judges it relevant or you invoke it, and a reference file loads
only when something opens it. That is why a skill can afford to be many times the size of a rule.

---

# `settings.json` — what is wired to what

`.claude/settings.json` is the only wiring in the repository. It sets the status line and maps Claude
Code's events onto the three hooks:

| Event                              | What runs                                  |
| ---------------------------------- | ------------------------------------------ |
| status line                        | `usage-limits --mode statusline`           |
| `SessionStart`                     | `usage-limits --mode inject`, `work-audit` |
| `UserPromptSubmit`                 | `usage-limits --mode inject`, `work-audit` |
| `PreToolUse` (Agent/Task/Workflow) | `usage-limits --mode gate`                 |
| `Notification`                     | `play-sound`                               |
| `Stop`, `StopFailure`              | `play-sound`, `work-audit`                 |
| `SessionEnd`, `PreCompact`         | `work-audit`                               |
| `SubagentStart`, `SubagentStop`    | `work-audit`                               |
| `PostToolUse` (AskUserQuestion)    | `work-audit`                               |

Every command is written against `${CLAUDE_PROJECT_DIR}`, so it works whatever directory the session
was started from. The audit entries run asynchronously — the log must never make you wait — and the
limits entries do not, because an injection that arrives after the prompt is useless.

Nothing else in the file needs touching. If a project already has a `settings.json`, merge these
entries into it rather than replacing it, and repoint an existing entry rather than adding a second
one for the same event — both would run.

---

# `.gitattributes` and `.gitignore`

`.gitattributes` is small, and it is the source of the two failures that only ever appear on somebody
else's machine:

- **LF-only launchers**, which takes two lines — one for `.claude/hooks/*` and one for `*.sh`. They
  are POSIX scripts, and a Windows checkout that converts line endings turns the shebang into CRLF,
  which breaks them on Linux and macOS.
- **A union merge for the audit log**, one line: `management/logs/** merge=union`. The log is
  append-only and one file per session, so a merge should keep both sides instead of conflicting.

The binaries and the WAV are marked as binary in the same file, so no platform's heuristics get a
vote.

`.gitignore` is the shorter half, and says what this repository deliberately does not keep: editor
state, and `management/` — the audit hook's own output. The log belongs to whatever project this is
copied into rather than to the template, so it is ignored here and the target project decides for
itself. Nothing under `scripts/` is ignored: the binaries are committed on purpose, and that is the
whole reason a copy works without a Go toolchain.

---

# The Go sources, and rebuilding

`scripts/usage-limits-src/`, `scripts/work-audit-src/` and `scripts/play-sound-src/` — the full source
of all three hooks, each with its own `README.md` covering that hook in implementation terms, and its
own `build.sh` that cross-compiles that hook for all three platforms. Every tunable default of the
limits hook is declared in `scripts/usage-limits-src/constants.go`; the config file only overrides it.

Rebuilding is only needed if you change the sources. The committed binaries are what make a copy work
without a toolchain.

```sh
sh scripts/usage-limits-src/build.sh              # one hook, all three platforms, into scripts/
sh scripts/usage-limits-src/build.sh .claude/bin  # or into a folder of your choosing
go -C scripts/usage-limits-src test ./...          # the working-week arithmetic
```

Each hook has its own script, so rebuilding all of them is the first line three times, with
`work-audit-src` and `play-sound-src` in place of `usage-limits-src`. Only the limits script takes an
output directory; the other two always write next to the sources, into `scripts/`.

All three build with `-trimpath`, so no local path is baked into a binary. The limits hook adds
`-buildvcs=false` on top, which makes it reproducible byte for byte: the same source gives the same
bytes in any checkout, so a committed binary can be verified against its source. The other two leave
Go's default version stamping on, so their bytes can differ between checkouts even when the source
does not.

The test suite is worth knowing about: the working-week arithmetic is checked against an independent
minute-by-minute oracle across every awkward time zone, at quarter-hour offsets, around every real
clock change of the year, for monotonicity, and for agreement with plain calendar time whenever the
configured week is a calendar-shaped one.

---

# Installing this in another project

Copy the hooks and their binaries into the project root — the two target folders have to exist
first, so create them in the same breath:

```sh
mkdir -p <project>/.claude <project>/scripts
cp -r .claude/hooks .claude/sounds .claude/usage-limits-config.json <project>/.claude/
cp -r scripts/usage-limits-* scripts/work-audit-* scripts/play-sound-* <project>/scripts/
```

All three hooks must be copied together with all three sets of binaries. A launcher whose binary is
missing exits 0 in silence — that is right for a hook and means the omission never announces itself.

The rules, the subagent and the skill are separate and optional — and worth choosing deliberately,
since the rules change how the model behaves from the next session onwards. Copy the rules together
with the subagent, for the reason given above: they refer to each other and to it by name.

```sh
cp -r .claude/rules .claude/agents .claude/skills <project>/.claude/
```

Then merge `.claude/settings.json` from this repository into the project's own — or copy it whole if
the project has none. If you are asking Claude Code to do this, point it at
[`AGENTS.md`](AGENTS.md), which spells out every step including the two that are easy to miss.

The launcher finds its binary in `scripts/`, `scripts/claude-code/` or `.claude/bin/`, so an existing
project's layout is respected.

**On Windows, commit the launchers and binaries executable.** Git ignores filesystem modes there
(`core.fileMode` is off), so a file added on Windows records as non-executable and the launcher then
silently finds nothing on a Unix clone:

```sh
git add --chmod=+x .claude/hooks/usage-limits .claude/hooks/work-audit .claude/hooks/play-sound \
  scripts/usage-limits-* scripts/work-audit-* scripts/play-sound-*
```

The `.gitattributes` in this repository keeps the launchers LF-only; copy it too, or add the same
two lines, or a Windows checkout will turn the shebang into CRLF and break them on Linux and macOS.

---

# Requirements

Claude Code, signed in (`claude login`). Nothing else — the rules, the subagent and the skill need
nothing at all, and the hooks need no runtime.

Binaries are committed for **darwin/arm64**, **linux/amd64** and **windows/amd64**. The launchers
pick one by `uname` and do not check the architecture, so on a platform outside that list — an Intel
Mac, an arm64 Linux box or container, FreeBSD — the exec fails rather than falling back, silently, as
a hook must.

The build scripts do not solve that on their own: each cross-compiles those same three targets and
takes no platform argument. Build your own instead, and **give it the name the launcher already looks
for on that operating system** — the name is a lookup key, not a description of the architecture:

```sh
cd scripts/usage-limits-src && GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 \
  go build -trimpath -buildvcs=false -ldflags "-s -w" -o ../usage-limits-darwin-arm64 .
```

That is an Intel Mac; for arm64 Linux the target is `GOOS=linux GOARCH=arm64` and the name to write
is `usage-limits-linux-amd64`. Repeat for `work-audit-src` and `play-sound-src`, whose launchers work
the same way. Go 1.25+ is needed for this, and it is the only thing Go is ever needed for.
