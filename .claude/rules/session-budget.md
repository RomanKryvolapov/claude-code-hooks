# Session budget — spend to the task; limits are a guardrail

Spend tokens in proportion to the task's real complexity — the cheapest path that reliably solves it,
not the most thorough one the budget could afford. Limits mark where you would hit a wall, not a budget
to fill: a green zone is not permission to burn, and nearing a threshold means narrowing scope. Work
inline by default; subagents and fan-outs cost multiples of one thread.

**The one exception** is the `change-reviewer` pass that rule `finishing-work` runs once per task, when
the task is done. It is never weighed against the budget and the usage gate exempts it by name, so no zone
blocks it; if the agent cannot be launched at all, the revision says so and falls back to a cold
self-pass.

## Reading the live numbers

The `usage-limits` hook (wired in `.claude/settings.json`) adds a `<usage_limits>` block at session start
and on every prompt. It is data, account-wide across every session on this subscription, and it outranks
any assumption about remaining quota.

- `<context>` — how much of this session's context window the last request carried; why long sessions
  get compacted.
- `<limit … used gone resets>` — one per limit: the 5-hour window, the 7-day window, each per-model weekly
  bucket. `used` ahead of `gone` means the budget runs out before the reset. Weekly limits are scaled to
  the working week the config defines — a display scale only; resets, burn, zones and the gate run on
  clock time.
- `<tokens>` — fresh (new input, output, cache writes) and cached (read back from the prompt cache at a
  tenth of the price), for this session and for the account.
- `<burn>`, `<zone>` (level, the binding limiter, what it allows) and, at session start, `<model>`.
  `<config_problem>` names config the hook could not use; `<stale>` means the figures are that old —
  read them as a floor.

For a precise check: `.claude/hooks/usage-limits --mode json --model <current-model-id>` (fields include
`session_pct`, `weekly_pct`, `zone`, `limiter`, `burn_pct_per_min`, `min_to_exhaust`, `horizon_min`,
`active_model_bucket`); `{"error": …}` means plan by time only.

The PreToolUse gate refuses subagent and workflow spawns in a hot zone. In ORANGE it asks about the first
spawn of the session only and remembers only that it asked, not the answer — refusing that one spawn sets
the shape of the rest of the session. Claude Code itself has no per-session spawn cap any more
(`CLAUDE_CODE_MAX_SUBAGENTS_PER_SESSION` is a no-op since v2.1.224); only its limits on concurrent
subagents and on nesting depth apply.

## Size the task by its work, not by the remaining budget

| Class   | Shape                                | Depth                                                     |
| ------- | ------------------------------------ | --------------------------------------------------------- |
| trivial | one line, a typo, an obvious fix     | edit directly; no exploration, no thinking                |
| small   | 1–2 files, a clear change            | targeted reads, inline, minimal thinking                  |
| medium  | a feature slice, 3–6 files           | scoped reads, one plan, high effort only on the hard call |
| large   | cross-cutting, a migration, an audit | phases; subagents only where they clearly pay off         |

- The cheapest model that can do it; the expensive one for genuinely hard reasoning and design. Route
  mechanical delegation to a cheap model through the agent's model option.
- Thinking in proportion to the task; read by file and symbol name, never by blind scans; never re-read
  what you already have.
- No inflation: no passes, tests, refactors or "while I'm here" changes nobody asked for.

## Subagents and fan-out

- Spawn only for a clear net saving: read-heavy exploration across many files (roughly more than three or
  four large ones) whose verbose output stays in the subagent. Editing one or two files, shell and git,
  anything the main session finishes faster than a spawn starts — inline.
- A fan-out of N agents costs about N times one: keep it narrow (2–3 agents, tight tool lists, small
  outputs), sequential when the budget is tight, never unbounded.
- Ultracode and the Workflow tool are for genuinely large, decomposable work: scout the real work-list
  first, then about one agent per item — usually 2–6, tens only for a proven large sweep; hundreds is a
  decomposition bug. State the count and the work-list before fanning out.
- Reviewing a change is depth, not breadth: one agent, sequentially.

## Zones narrow what you may do

| Zone   | 5-hour | Weekly and buckets | Behaviour                                                              |
| ------ | ------ | ------------------ | ---------------------------------------------------------------------- |
| GREEN  | < 50%  | < 80%              | spend to task size; subagents only for a clear saving                  |
| YELLOW | 50–80% | 80–90%             | at most 2 parallel subagents, no heavy fan-out, no re-reads            |
| ORANGE | 80–90% | 90–95%             | the first subagent of the session is asked about; essential edits only |
| RED    | ≥ 90%  | ≥ 95%              | finish, write `.claude/CHECKPOINT.md`, no new tasks until the reset    |

These are the shipped defaults; `.claude/usage-limits-config.json` may change them, and the hook reports
the zone it actually computed. `limiter` names what binds: `session` (a model switch will not help —
phase around the reset), `weekly` (move heavy work past the reset), `model-bucket` (only the active model
is constrained — route heavy work to another).

Before a task of more than about ten minutes, size it first; if the estimate reaches about 0.8 of the
forecast horizon, split it at natural seams, write `.claude/CHECKPOINT.md` (done, in flight, next, open
questions) between phases, and schedule the rest past the reset.

## Checklist

- [ ] Effort sized by the task, not the remaining budget; a small task not inflated.
- [ ] Cheapest sufficient model and thinking; context read by name, no re-reads.
- [ ] Subagents only for a clear saving, mechanical work on a cheap model, a fan-out counted as N times.
- [ ] Workflow fan-out sized to a scouted work-list.
- [ ] Zone restrictions respected; the limits used as a guardrail, not a target.
- [ ] A task that will not fit the window split into phases with a checkpoint.
