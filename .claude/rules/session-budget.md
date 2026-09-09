# Session budget — spend to the task, limits are a guardrail

**Spend tokens in proportion to the task's real complexity** — the cheapest path that reliably
solves it, not the most thorough one the budget could afford.
Limits are a **guardrail**: they mark where you would hit a wall or a rate-limit so you can stop
short, not a budget to fill.
A GREEN zone does NOT mean "burn freely"; nearing a threshold means narrow scope, never spend up to
it.
Default to inline work — subagents and fan-outs cost multiples of one thread and are only worth a
clear net saving.

**One standing exception, and it is not discretionary:** the `change-reviewer` pass that
[finishing-work](finishing-work.md) requires at the grading stage. The usage gate exempts it by
name, so it is not blocked in any zone. It is a single agent, and it runs before the end of any turn
in which files changed, whether or not the message mentions them. It is never weighed against the
budget: the developer asked for it precisely because a self-graded verdict keeps passing work that
then comes back. A hot zone can no longer prevent it. A per-session spawn cap still can — and where
it does, the revision says so and falls back to a cold self-pass; it is never quietly skipped to
save tokens.

The failure mode this prevents: inflating a small task into a large one — extra passes, unrequested
tests/refactors, blind full-file scans, re-reads, and subagents that cost multiples of doing the
work inline — just because budget was available.

## Live numbers (from the hook)

Project hooks (`.claude/hooks/usage-limits`, wired in `.claude/settings.json`) inject live
subscription usage: a `[session-budget]` block at session start and a `[limits]` block on every
prompt. Both are the same picture as the status line: a `CONTEXT` line — how much of this session's
context window the last request carried, with its own gauge (one cell per percent) and a
`used/window` counter; it is not a subscription limit, it is the reason a long session gets
compacted — and under it the limits side by side (5-hour window, 7-day window, then every per-model
weekly bucket), each a pair of stacked lines: `LIMIT` (percentage of the budget spent, with its
gauge and the reset time) over the share of the window already gone (same gauge, and how long is
left). **Read one against the other:** a fuller `LIMIT` bar than the one under it means the budget
is running out faster than the clock and will not last to the reset. That bottom bar is captioned
`TIME` on the 5-hour row, which measures calendar time — and `WORK` on the weekly rows, which
measure **working** time: only the hours the config's `working_week` says are worked count, each
weekday for as much as its own percentage says. Shipped as noon to eight on weekdays and a fifth of
that at the weekend, which makes the week 43 working hours long rather than 168, so the bar stands
still overnight. It exists so the comparison means something: nobody spends budget at four in the
morning, and a weekend eats two of the seven days while spending almost none — which used to leave
the bar ahead of the spend every Monday and behind it every Friday. Where a day begins is fixed by
`time_zone` in the same config rather than by the host clock, and anything in that config the hook
cannot use is named on a `CONFIG` line rather than dropped in silence. **It is a display scale
only** — every reset time, the `RESET AFTER` countdown, the burn rate, the forecast, the zones and
the gate all stay on real clock time. Under them come a `BURN` line (rate, forecast, pace) and a
`ZONE` line (zone, binding limiter, what it allows); session start adds a `MODEL` line.

The hook is a **cross-platform Go binary** — a thin POSIX launcher (`.claude/hooks/usage-limits`)
runs the prebuilt binary for the current OS, one per OS committed in this project's scripts folder
with its source beside them, so it needs no Node or other runtime.

The data is **account-global** — it reflects ALL Claude sessions on this subscription; trust it over
any internal assumption about remaining quota.

The PreToolUse gate is the enforcement layer that backs this guidance: it refuses subagent and
workflow spawns in a hot zone. A second, harder ceiling is available but not switched on by default
— `CLAUDE_CODE_MAX_SUBAGENTS_PER_SESSION` in the `env` block of `settings.json` caps spawns per
session outright, and the count resets on `/clear`. Set it if you want a fixed ceiling as well as a
zone-based one.

For a precise guardrail check, run it from the repo root and pass the current model id so the
per-model bucket resolves:

```
.claude/hooks/usage-limits --mode json --model <current-model-id>
```

Key fields: `session_pct`, `weekly_pct`, `zone`, `limiter`, `burn_pct_per_min` (+ `burn_source`),
`min_to_exhaust`, `horizon_min`, `active_model_bucket`, `weekly_scoped`.
`{"error": ...}` → plan by time only.

## Step 1 — size the task by complexity, not by remaining budget

Classify by the _real work_ (files touched × passes needed), independent of how much window is left:

| Class   | Rough shape                       | Default depth                                                  |
| ------- | --------------------------------- | -------------------------------------------------------------- |
| trivial | one-line / typo / obvious fix     | edit directly; no thinking, no exploration                     |
| small   | 1–2 files, clear change           | targeted reads, inline, minimal thinking                       |
| medium  | a feature slice, 3–6 files        | scoped reads, one plan pass, high effort only on the hard call |
| large   | cross-cutting / migration / audit | phase it; subagents only where they clearly pay off (Step 3)   |

Complexity — not the budget — sets how much you read, whether you think, whether you delegate, how
many passes, and which model.
Never scale work _up_ just because the window has room.

## Step 2 — pick the minimum sufficient effort

- **Model:** the cheapest model that can do it by default; reserve the expensive model for genuinely
  hard reasoning / design calls. Subagents take a model override — route mechanical delegation to a
  cheap model.
- **Thinking / effort:** proportional — none for trivial/small; high effort only on the hard fork.
- **Context:** read by name (files/symbols), not blind directory scans; never re-read what you
  already have; pull only what the task needs.
- **No inflation:** do not add passes, tests, refactors, or "while I'm here" changes the user did
  not ask for.

## Step 3 — subagents & fan-out: spend only for a net saving

Subagents are **not cheaper by default.**
Each runs its own context window and reloads a bootstrap (agent system prompt + tools) on every
spawn; multi-agent fan-out runs **~4–7× the tokens** of doing the same work in one thread.

**Spawn only for a clear net win:** read-heavy exploration across **many files** (roughly >3–4 large
files/dirs) where the verbose output stays isolated in the subagent and the main context stays
clean.

**Do it inline instead** when editing 1–2 files, running shell/git, or anything the main session
finishes faster than a spawn would even start.

**When you delegate, keep it cheap:** route mechanical work (search, summarize logs, extract data,
audits) to a cheap model via the agent's model option; count fan-out of N agents as **N× the spend**
(sequential when budget-tight, parallel only when time-tight and worth it); scope narrowly (2–3
agents, tight tool lists, small outputs); never leave agents running unbounded.

### Ultracode / Workflow mode — the highest-risk spend

Ultracode and the Workflow tool fan out aggressively by design: one run can spawn dozens to hundreds
of subagents, and `CLAUDE_CODE_MAX_SUBAGENTS_PER_SESSION` is the only thing that puts a hard
per-session ceiling on it, if you have set one.
Treat it as opt-in for **genuinely large, decomposable work only** — never the default for an
ordinary task, and never by reflex because the word "ultracode" appeared.

- **Scout first, then size the fleet to the real work-list.** Discover the actual items inline, then
  spawn about **one agent per item** — no concrete work-list → no fan-out.
- **Cap the count deliberately.** Most tasks that justify delegation need **2–6 agents**; tens only
  for a proven large sweep. **Hundreds is a decomposition bug, not a plan** — stop and rethink, do
  not launch.
- **One-thread work stays in one thread**; state the count and the work-list before fanning out, and
  if you cannot justify the number, cut it.

## Step 4 — limits as a guardrail

Zones only **narrow** what you may do (worst of session / weekly / active-model bucket):

| Zone   | Session (5h) | Weekly & buckets (7d) | Behavior                                                                                                                                                                                                                          |
| ------ | ------------ | --------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| GREEN  | <50%         | <80%                  | spend to task size; subagents/fan-out only for a clear net saving                                                                                                                                                                  |
| YELLOW | 50–80%       | 80–90%                | ≤2 parallel subagents, no heavy fan-outs, avoid re-reads                                                                                                                                                                           |
| ORANGE | 80–90%       | 90–95%                | the **first** subagent of the session is asked about, and only that one — the gate asks once per session, because a gate that asks on every spawn is one people switch off. `change-reviewer` is never asked. Essential edits only |
| RED    | ≥90%         | ≥95%                  | finish + write `.claude/CHECKPOINT.md`, no new tasks until reset; `change-reviewer` still runs                                                                                                                                     |

Those are the shipped defaults; a project may set its own in `.claude/usage-limits-config.json`, and
the hook reports the zone it actually computed rather than the one this table implies.

The gate asks **once**, and it remembers only that it asked — a hook cannot learn what you answered.
So refusing that one spawn stops that spawn and nothing after it. Treat the single prompt as the
moment to decide the shape of the rest of the session, not as a per-spawn brake.

**Reviewing a change is sequential — one agent, once.** The code runs in an order and the
examination follows it; parallel readers lose the one thing that finds defects, which is a single
mind holding the whole path from input to result. Fan-out is for breadth, and a revision is depth.

`limiter` names the binding constraint: `session` (5h — switching model won't help, phase around the
reset), `weekly` (7d — schedule heavy work past the reset), `model-bucket` (only the active model is
constrained — route heavy work / subagents to a model with a freer bucket).
Nearing a threshold is a signal to narrow scope, not permission to spend up to it.

**Before an M/L/XL task** (>10 min active work): size to the task first, then use the numbers only
as a guardrail — if the estimate would hit the window before you finish (roughly ≥ 0.8 × the
forecast horizon), split at natural seams, checkpoint between phases (write `.claude/CHECKPOINT.md`:
done / in-flight / next / open questions), and schedule the rest past the reset.
The check exists to avoid hitting the wall mid-task, not to pack work up to the ceiling.

## Checklist

- [ ] Effort sized by the task's complexity, **not** by remaining budget; a small task was not inflated.
- [ ] Cheapest sufficient model / thinking level chosen; expensive model reserved for hard reasoning.
- [ ] Context read by name and scoped to the task; no blind scans, no re-reads.
- [ ] Subagent / fan-out used **only** for a clear net token saving; trivial work inline; mechanical delegation on a cheap model; fan-out counted as N×. **The `change-reviewer` pass is outside this line entirely** — it is not weighed, not economised and not skipped for a small change.
- [ ] Ultracode / Workflow fan-out reserved for genuinely large, decomposable work; fleet sized to a real scouted work-list (typically 2–6 agents; hundreds is a decomposition bug).
- [ ] **Reviewing a change was done sequentially, by one agent** — not fanned out. The code runs in an order and the examination follows it; parallel readers lose the one thing that finds defects, which is a single mind holding the whole path from input to result.
- [ ] Limits used as a guardrail (do not hit the wall / rate-limit), not as a target to fill; zone restrictions respected.
- [ ] A large task that will not fit the window was split into phases with a checkpoint.
