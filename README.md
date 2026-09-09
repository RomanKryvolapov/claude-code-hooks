# Claude Code hooks: subscription limits, a work log, and a sound when it stops

Three hooks for [Claude Code](https://claude.com/claude-code), ready to drop into any project.

- **`usage-limits`** — puts your live Anthropic subscription usage in front of the model and in your
  status line, and refuses subagent spawns once you are close to a limit.
- **`work-audit`** — records what actually happened in each session into an append-only log, so you
  can see the work without relying on the model to summarise itself.
- **`play-sound`** — plays a short sound when Claude Code stops or wants your attention, so you can
  look away while it works.

All three are single static binaries built from the Go sources in this repository, committed for
macOS on Apple Silicon, Linux on x86-64 and Windows on x86-64. **There is no runtime to install** —
no Node, no Python, no Go. Copy two folders and it works. On any other platform — an Intel Mac,
arm64 Linux — run `sh scripts/<name>-src/build.sh` once; you need Go only for that.

A set of behaviour rules ships alongside them in `.claude/rules/`, together with one subagent and
one skill. The limits hook supplies numbers; the `session-budget` rule is what makes the model act
on them, and copied without it the numbers are decoration. See
[What else is in here](#what-else-is-in-here) for what each rule, the subagent and the skill are
for — they are useful on their own and can be taken separately from the hooks.

---

## Why the limits hook exists

A Claude subscription has three kinds of limit running at once: a **5-hour session window**, a
**7-day window**, and a **per-model weekly bucket** for some models. They are account-global — every
Claude session you have open shares them — and nothing inside a coding session tells you where you
stand. So you either stop early to be safe, or you run into a wall in the middle of something.

This hook reads the real numbers from Anthropic's own usage endpoint and shows them in one picture,
in three places: injected into the model's context at session start and on every prompt (so the
model can size its own work), rendered in the status line, and available as JSON for scripting.

```
CONTEXT  634k/1M   63 %  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
5 HOURS   31 % LIMIT ███████░░░░░░░░░░░░░ RESET AT 15:00      7 DAYS   24 % LIMIT █████░░░░░░░░░░░░░░░ RESET AT MON 18:00          FABLE    0 % LIMIT ░░░░░░░░░░░░░░░░░░░░ RESET AT MON 18:00
-         77 %  TIME ████████████████░░░░ RESET AFTER 1:10             31 %  WORK ███████░░░░░░░░░░░░░ RESET AFTER 5 days 03:46            31 %  WORK ███████░░░░░░░░░░░░░ RESET AFTER 5 days 03:46
BURN  0.16 %/min (sampled) - safe to reset 15:00, pace x0.6
ZONE  GREEN (limiter: weekly) -> spend to task size; subagents only when they clearly save tokens.
```

---

## Reading the picture

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

---

## The working week — why the weekly bar is not calendar time

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

### Where a day begins

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

---

## The gate

Beyond showing numbers, the hook is wired to `PreToolUse` for `Agent`, `Task` and `Workflow`, and
can refuse a spawn:

| Zone   | Session (5h) | Weekly & buckets (7d) | What the gate does                                  |
| ------ | ------------ | --------------------- | --------------------------------------------------- |
| GREEN  | <50%         | <80%                  | nothing                                             |
| YELLOW | 50–80%       | 80–90%                | nothing                                             |
| ORANGE | 80–90%       | 90–95%                | asks about the **first** spawn of the session, once |
| RED    | ≥90%         | ≥95%                  | denies the spawn                                    |

It asks **once per session**, not per spawn: a gate that interrupts every time is a gate people
switch off. The thresholds are configurable; set `gate_enabled` to `false` to turn the gate off and
keep the display.

---

## Configuration

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

---

## The rule that makes the numbers matter

A hook can put numbers in front of the model. It cannot make the model do anything about them —
Claude Code has no built-in notion of a budget, and nothing in the injected block says whether 34%
on a Wednesday is fine or alarming.

`.claude/rules/session-budget.md` is that missing half. It tells the model to size effort by the
task's own complexity rather than by what is left, that subagents cost several times what inline
work costs and are worth spawning only for a clear net saving, that reviewing a change is one agent
working sequentially rather than a fan-out, what each zone allows, what the binding limiter means for
whether switching model would help at all, and when to split a long task and write a checkpoint
instead of running into the wall.

**Copying the file is the whole installation.** Claude Code discovers every `.md` under
`.claude/rules/` at launch, recursively, and loads it with the same priority as `.claude/CLAUDE.md`.
No import, no line in `CLAUDE.md`, nothing to wire.

---

## What else is in here

The rules, the subagent and the skill are independent of the hooks — take the ones you want. Only
`session-budget` has a tie to them, and it is the one described above.

### The rules

`.claude/rules/`, nine files, about 1,500 lines. Because rules load unconditionally, **all of them
sit in the context window of every session** — that is what they cost. A rule that should only apply
to part of a codebase can carry a `paths:` frontmatter key with glob patterns, and then loads only
when Claude touches a matching file; none of these nine do, because none of them are about a
particular kind of file.

| Rule                      | What it is for                                                                                                                                                                                            |
| ------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `session-budget`          | The other half of the limits hook: size effort by the task's own complexity rather than by what is left, when a subagent is worth spawning at all, what each zone allows, when to split a task and checkpoint |
| `before-starting-work`    | Pull before acting on the message, read what actually arrived, and check what the project stands on — all of it before planning, editing or reviewing                                                       |
| `working-autonomously`    | A task handed over is yours from beginning to end: decide what it contains, carry it to a verified finish, report once. Names the only two reasons to stop early                                            |
| `architectural-forks`     | The opposite brake: when a choice is expensive to undo, stop and put the fork to the developer with the options laid out — and, just as firmly, what not to ask about                                       |
| `code-quality`            | Find the mechanism that produces a defect instead of patching the symptom; clean without over-designing; SOLID as judgement. Its middle section is about prompts — when a model misbehaves, look at what reached it and in what order |
| `finishing-work`          | One thorough revision before any turn that changed files ends: read in execution order, interrogate, build failure hypotheses, run it, judge it as an architect, then grade it independently                |
| `merge-conflicts`         | A conflict is two people's intentions meeting, not two blocks of text disagreeing. Resolve by hand, never by force, and prove afterwards that both intentions survived                                      |
| `response-style`          | Answer in the language of the message, and the shape of the two reports — the work report at the end of a turn that changed files, the problem report when a review turns something up                      |
| `web-search-when-in-doubt`| When reality diverges from what you know, search before stacking another fix on an unverified assumption; the vendor's own documentation for the version in use is the authority                            |

### The subagent

`.claude/agents/change-reviewer.md` is the independent grading pass `finishing-work` requires. It
receives the diff and the requirement and nothing else — no access to the author's reasoning, which
is the point: an author who knows the intent reads the code as the intent. It is told to refute
rather than confirm, it may not edit anything, and it works sequentially rather than broadly.

**The limits hook knows about it.** The gate exempts `change-reviewer` by name, so the one pass a
revision depends on is never the spawn that gets refused in a hot zone.

### The skill

`.claude/skills/dev-ai-prompt-generation/` — a hub `SKILL.md` plus 30 reference files on writing and
reviewing prompts, Agent Skills, rules and `CLAUDE.md`: structure and delimiters, output contracts
and structured outputs, reasoning-model effort knobs, few-shot, context engineering, prompt caching,
evals and LLM-as-judge, injection defences, RAG grounding, and cross-vendor conventions.

Unlike rules, **a skill costs nothing until it is used.** Only its name and description sit in
context; the body loads when the model judges it relevant or you invoke it, and a reference file
loads only when something opens it.

---

## The work log

`work-audit` writes one append-only file per session under `management/logs/`, a folder per
developer (from `git user.name`) and a folder per day inside it. Entries are separated by blank
lines and each event is written in a single append, so parallel sessions never interleave and the
file reads in the order things happened.

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

The log is append-only and one file per session, so `management/logs/** merge=union` in
`.gitattributes` keeps both sides of a merge instead of conflicting. It is in this repository's
`.gitattributes` and should be copied along with it.

---

## The sound hook

`play-sound` plays `.claude/sounds/notify.wav` when Claude Code stops, fails, or raises a
notification — a permission prompt, an idle prompt, a question. It is the same shape as the limits
hook: a POSIX launcher picking a committed binary per platform, no runtime.

Replace `.claude/sounds/notify.wav` with any WAV file to change the sound. On macOS it plays through
`afplay`, on Linux through whichever of `paplay`, `aplay` or `ffplay` exists, on Windows through the
system player. If none is available it exits quietly — a notification hook must never interrupt a
session.

---

## What none of them ever do

- **It never breaks a session.** Every failure path exits 0: the injection and gate modes stay
  silent, the status line prints `LIMITS -> N/A`, JSON prints an error document. A missing binary, a
  broken config, an expired token, no network — all end the same way.
- **It sends nothing anywhere.** The only outbound request is to `api.anthropic.com`, authenticated
  with the OAuth token `claude login` already stored in `~/.claude/.credentials.json`. No API key is
  needed and the token is never printed.
- **The limits hook writes only two files**, both under `~/.claude`: a usage cache and a note of
  whether the gate has asked this session. The audit hook writes only its own log under
  `management/logs/`, plus two throwaway counters in the OS temp folder.

---

## Modes

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

## Installing it in a project

Copy the hooks and their binaries into the project root:

```sh
cp -r .claude/hooks .claude/sounds .claude/usage-limits-config.json <project>/.claude/
cp -r scripts/usage-limits-* scripts/work-audit-* scripts/play-sound-* <project>/scripts/
```

All three hooks must be copied together with all three sets of binaries. A launcher whose binary is
missing exits 0 in silence — that is right for a hook and means the omission never announces itself.

The rules, the subagent and the skill are separate and optional:

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

## Rebuilding

Only needed if you change the Go sources. The committed binaries are what make a copy work without a
toolchain.

```sh
sh scripts/usage-limits-src/build.sh              # all three, into scripts/
sh scripts/usage-limits-src/build.sh .claude/bin  # or into a folder of your choosing
go -C scripts/usage-limits-src test ./...          # the working-week arithmetic
```

Builds are reproducible — `-trimpath -buildvcs=false` — so the same source gives the same bytes in
any checkout, and a committed binary can be verified against its source.

The test suite is worth knowing about: the working-week arithmetic is checked against an independent
minute-by-minute oracle across every awkward time zone, at quarter-hour offsets, around every real
clock change of the year, for monotonicity, and for agreement with plain calendar time whenever the
configured week is a calendar-shaped one.

---

## Requirements

Claude Code, signed in (`claude login`). Nothing else.

Binaries are committed for **darwin/arm64**, **linux/amd64** and **windows/amd64**. The launchers
pick one by `uname` and do not check the architecture, so on a platform outside that list — an Intel
Mac, an arm64 Linux box or container, FreeBSD — the exec fails rather than falling back. Build for it
once with `sh scripts/<name>-src/build.sh`, which needs Go 1.25+. That is also the only thing Go is
ever needed for.
