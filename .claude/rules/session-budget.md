# Session budget — spend to the task, limits are a guardrail

**Spend tokens in proportion to what the task actually needs** — the cheapest path that reliably
solves it, not the most thorough one the budget could afford.

The numbers the `usage-limits` hook injects are a **guardrail**: they mark where you would hit a
wall so you can stop short of it, not a budget to fill. A GREEN zone does not mean "burn freely",
and nearing a threshold means narrowing scope, never spending up to it.

The failure this prevents: inflating a small task into a large one — extra passes, unrequested
tests and refactors, blind full-file scans, re-reads, and subagents that cost several times what
doing the work inline would — merely because budget was available.

## The numbers you are given

At session start and on every prompt, a block is injected with the account's live usage. It is
**account-global**: it covers every Claude session open on this subscription, so trust it over any
internal assumption about what is left.

- **`CONTEXT`** — how full this session's own context window is. Not a subscription limit; it is the
  reason a long session eventually gets compacted.
- **`LIMIT`** over **`TIME`** / **`WORK`** — for each limit, the budget spent above the share of the
  window gone. **Read one against the other:** a fuller top bar than bottom bar means the spend is
  outrunning the clock and will not last to the reset. The 5-hour row's bottom gauge is calendar
  time (`TIME`); the weekly rows measure working time (`WORK`) — configured hours, weighted per
  weekday — so that the comparison means something across a weekend or a night.
- **`BURN`** — the current rate, a forecast, and `pace xN`: spend divided by the share of the window
  gone. Above 1.0 is ahead of the clock.
- **`ZONE`** — the worst of the session window, the weekly window and the active model's own weekly
  bucket, plus which of the three is binding.

For a precise check when planning, run the hook yourself and read the JSON:

```
.claude/hooks/usage-limits --mode json --model <current-model-id>
```

Key fields: `session_pct`, `weekly_pct`, `zone`, `limiter`, `burn_pct_per_min`, `min_to_exhaust`,
`horizon_min`, `active_model_bucket`. An `{"error": ...}` document means plan by time alone.

## Size the task first, by its own complexity

Classify by the real work — files touched times passes needed — and **not** by how much budget is
left:

| Class   | Rough shape                          | Default depth                                                       |
| ------- | ------------------------------------ | ------------------------------------------------------------------- |
| trivial | one line, a typo, an obvious fix     | edit directly; no exploration                                       |
| small   | one or two files, a clear change     | targeted reads, inline, minimal thinking                            |
| medium  | a feature slice, three to six files  | scoped reads, one planning pass, deep thought only on the hard call |
| large   | cross-cutting, a migration, an audit | phase it; delegate only where it clearly pays off                   |

Then pick the minimum sufficient effort: the cheapest model that can do it, thinking proportional to
the difficulty, context read by name rather than by scanning, and no passes, tests or refactors
nobody asked for.

## Subagents are not cheaper by default

Each subagent runs its own context window and reloads its own bootstrap on every spawn, so a
multi-agent fan-out costs several times what the same work costs in one thread.

**Spawn only for a clear net win** — read-heavy exploration across many files, where the verbose
output stays inside the subagent and the main context stays clean. Do it inline instead when editing
one or two files, running shell or git, or anything that finishes faster than a spawn would start.
When you do delegate, route mechanical work to a cheap model, scope it narrowly, and count a fan-out
of N agents as N times the spend.

**Reviewing a change is depth, not breadth.** Code runs in an order and the examination follows it;
parallel readers lose the one thing that finds defects, which is a single mind holding the whole path
from input to result. One agent, once.

## What each zone allows

| Zone   | Session (5h) | Weekly & buckets (7d) | Behaviour                                                                      |
| ------ | ------------ | --------------------- | ------------------------------------------------------------------------------ |
| GREEN  | <50%         | <80%                  | spend to task size; subagents only for a clear net saving                      |
| YELLOW | 50–80%       | 80–90%                | be lean: at most two parallel subagents, no heavy fan-outs, avoid re-reads     |
| ORANGE | 80–90%       | 90–95%                | the session's first spawn is asked about; essential edits only; offer to defer |
| RED    | ≥90%         | ≥95%                  | finish what is open, write a checkpoint, start nothing new until the reset     |

The thresholds are the defaults; a project may set its own in `.claude/usage-limits-config.json`,
and the hook reports the zone it actually computed.

`limiter` names the binding constraint and is what tells you whether anything can be done about it:
`session` — switching model will not help, phase around the reset; `weekly` — schedule heavy work
past the reset; `model-bucket` — only the active model is constrained, so route heavy work to a model
with a freer bucket.

**The gate asks once per session, not per spawn.** A gate that interrupts every time is one people
switch off, so treat that single prompt as the moment to decide the shape of the rest of the session
rather than as a per-spawn brake.

## Before a task that will run long

Size it to the task first, then use the numbers only as a guardrail: if the estimate would hit the
window before the work finishes — roughly at or beyond 80% of the forecast horizon — split it at
natural seams, write a checkpoint between phases (what is done, what is in flight, what is next, what
is still open), and schedule the rest past the reset. The check exists to avoid hitting the wall
mid-task, not to pack work up to the ceiling.

## Checklist

- [ ] Effort sized by the task's complexity, **not** by remaining budget; a small task was not inflated.
- [ ] Cheapest sufficient model and thinking level; deep reasoning reserved for the hard call.
- [ ] Context read by name and scoped to the task; no blind scans, no re-reads.
- [ ] Subagents used only for a clear net saving; trivial work inline; a fan-out counted as N times the spend.
- [ ] A change was reviewed sequentially by one agent, not fanned out.
- [ ] Limits treated as a wall to stop short of, not a target to fill; zone restrictions respected.
- [ ] A long task that will not fit the window was split into phases with a checkpoint.
