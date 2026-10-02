# Finishing work — one thorough revision, a commit only on request

Every turn in which files changed ends with a revision of what changed — whatever the message says and
however small the change, a docs-only edit or a one-liner included. The tree is what the next turn
builds on, and a size threshold you may apply yourself is one you will eventually apply to something that
did not qualify. Committing is a separate act with its own permission (last section).

Files means every file git does not ignore — what `git status` lists, untracked ones included. Not a
change to revise: what a run writes where git ignores it (run folders, `.env`, Claude's memory), and the
results a project's own product flow writes into folders its CLAUDE.md names as product output (a lead
tracker, dossiers, letters). That is the product running; its own checks apply.

Defects are found by three things done once, properly: reading the change in the order the code
executes, deliberately constructing what could go wrong, and running it with real data. Only running it
is evidence — work is finished when you watched it do the right thing, not when the code looks right.
The revision follows one path with one mind, so it is sequential, never split across parallel readers.

## Stage 0 — pull

Bring the branch up to date as rule `before-starting-work` sets out: a revision read against a base that
moved is worthless. A conflict is handled by rule `merge-conflicts`; nothing here permits forcing past one.

## Stage 1 — read every change in execution order

- `git status` and `git diff` (and `--staged`): every file, every hunk; nothing is reported unread.
- Then read it in the order it runs: entry point → what it calls → what that returns → what is stored →
  what the caller does with it. The defects that live between two files hide in alphabetical order.
- Scope: every change belongs to the task. Accidental edits, formatter runs over unrelated files,
  leftover experiments, stray or generated files are reverted or split out with a stated reason.
- No secrets or junk: no `.env`, credentials, tokens, large binaries, IDE files.

## Stage 2 — interrogate each change

- What it does, in one sentence — a line you cannot explain is one you do not understand well enough to
  ship.
- What it assumes — about its inputs, what ran before it, ordering, what exists, the caller — and what
  happens when each assumption is false.
- Who else uses it: for every changed function, class, endpoint, DTO, schema, prompt or config key, find
  the references and decide whether the change still holds for each. The list goes into the report, one
  line per consumer with what was checked; one you could not check is named as unchecked.
- Contract surfaces: where CLAUDE.md names them (which document is the source of truth, what must be
  regenerated, which mirror is hand-synced), follow that list; where it names none, work them out from the
  change itself. The source of truth first, then the code, in the same change.

## Stage 3 — write down what could go wrong, then check each

- Corner cases: empty, null, zero, one, very many; the boundary and one either side; duplicates; out of
  order; the same call twice; unicode, encoding, timezones; first run versus re-run; the largest
  realistic input.
- Failure paths: for every database, network, file-system, subprocess or model-provider call — it fails,
  hangs, returns garbage, returns half.
- Silent failures above all: a default hiding a missing answer, a catch that logs nothing, a skipped step
  still reporting success, a fallback indistinguishable from a real result. They reach users because
  nothing announces them.
- Concurrency and repetition: two at once, the same request twice, a retry landing after the original
  succeeded.
- What the change made stale: a comment, rule, table or document describing behaviour that is gone.

## Stage 4 — run it

- Execute the changed path with inputs you chose, the corner cases included; predict the result first,
  then compare. Watch the effects — what was written, logged, left behind — not only the return value.
- Run the gates CLAUDE.md lists under "Gates" for what changed. A turn that changed only notes or tracker
  entries has no code gate in scope — that is scoping, and the verdict says which gates were in scope.
- Read the output, never the exit code: a suite that skips a block whose dependency is missing still
  prints success.
- A new test is watched failing first: break the fix, see it go red, restore.
- Non-deterministic behaviour (a model, a race, a clock, external data) is never called fixed on one
  run: state the number of runs, or "not verified".
- A change that is only documentation, a rule, a tracker entry or a comment has nothing to run, and the
  verdict says so instead of claiming a run; a documented command or example is still executed.
- Something to run that you cannot run is a result: say precisely what was not exercised and why — after
  searching for the missing access as rule `working-autonomously` requires.

## Stage 5 — judge it as an architect

Apply rule `code-quality` to what you built:

- Cause or symptom — say which in one line. A symptom fix needs a named reason the clean one is out of
  reach and a recorded follow-up.
- No crutch (a special case for the failing input, a retry over a race, a swallowed error, a loosened
  check, a second source of truth) and no over-engineering (an abstraction with one implementation, an
  extension point for a variation that does not exist, needless indirection); no duplicated logic, dead
  code, or names that stopped saying what the thing is.
- No claim nothing backs: a comment or description asserting a checkable fact — a count, a guarantee,
  "verified that…" — is a finding unless a test holds it.
- No promise without a consumer: a new field, flag, column or documented capability is used and tested in
  the same change.
- A version range you wrote admits no version the rest of the toolchain rejects.

## Stage 6 — independent grading, once per task

When you judge the whole task finished — not after each turn — launch the `change-reviewer` subagent
(`.claude/agents/change-reviewer.md`) once, with the task's full diff and the requirement and nothing
else: none of your reasoning, no defence of a choice. This launch is standing authorisation; a
session-level instruction to avoid subagents does not cover it.

- On the turns before that, the report says the independent pass is still to come.
- Check each finding against the code, never against your own reasoning; fix the real ones and run each
  fix as Stage 4 demands.
- Do not re-run the reviewer for a new pass unless the fixes changed behaviour or structure rather than
  wording; re-run the code instead.
- If the agent cannot be launched, say so and do the pass cold: re-read the diff without your notes and
  try to make each change fail.

## The verdict

Always scoped: what was run, what was only read, and what class of defect the gates that ran cannot see —
where a fake model, a stubbed service or an in-memory store stands in for the real thing, the verdict
says so.

- **PASS** — run, and it did what it should; everything found is fixed and the fixes were run.
- **FIX, then PASS** — real findings: fix them now, re-run the affected path, and say plainly what was found
  and fixed. Problems are never parked or merely reported.
- **STOP** — data loss, a security hole, a broken contract, the change contradicting its requirement, or a
  fork whose choice belongs to the developer (rule `architectural-forks`).

## The report

Its shape is in rule `response-style`. It always carries three lines:

- how each claim was established — **ran it**, **read it** or **assumed it**;
- who else this touched — the consumer list from Stage 2, unchecked ones named;
- what would prove it wrong — one sentence, "this is wrong if …".

## Committing — only when asked

Never `git commit`, `git push`, open a pull request or merge unless the developer asked for it in their
own words; finishing the work and a green build are not permission. Merging the parent into your own
branch (rule `before-starting-work`) is outside this ban — it publishes nothing.

Once a commit is asked for:

- Pull again and revise again: the state reviewed must be the state that lands.
- Only a PASS is committed. Never `--no-verify`; when a hook fails, fix the cause.
- Never force anything shared (rule `merge-conflicts`).
- The message describes the whole diff truthfully; unrelated changes do not share a commit.
- Everything under `management/logs/` goes in with the work, unedited. Stage it yourself unless you
  checked that a hook did — the hook that stages it runs only where it is enabled and skips path-limited
  commits — and confirm with `git status`.

## Checklist

- [ ] Revision ran before the turn ended because files changed — no floor, no size exception.
- [ ] Full diff read hunk by hunk, then in execution order; scope checked; no secrets or junk.
- [ ] Each change explained in a sentence; its assumptions and their failure named; consumers listed for the report.
- [ ] Failure hypotheses written down and checked.
- [ ] The change run on real data, its effects watched, the gates' output read, new tests seen failing first.
- [ ] What could not be exercised named as unverified, after searching for the access.
- [ ] Non-deterministic behaviour not called fixed on one run.
- [ ] Judged as an architect; cause or symptom said.
- [ ] At the end of the task, `change-reviewer` run once on the full diff and the requirement; its real findings fixed and the fixes run.
- [ ] Verdict scoped; the report carries ran/read/assumed, who else was touched, what would prove it wrong.
- [ ] No commit, push, pull request or merge without the developer's words; a requested commit preceded by a fresh pull and revision, audit logs staged.
