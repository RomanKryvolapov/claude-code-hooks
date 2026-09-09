# Finishing work — one thorough revision, and only then a commit

**The revision runs before the end of any turn in which files changed** — whether or not the message
you are about to send mentions them. Editing a file and then closing the turn with a question, a
status note or an unrelated answer does not postpone it: the change is in the tree either way, and
the tree is what the next turn builds on. That is the trigger, and it is deliberately something
anyone can check from the outside: not a commit request, not the size of the change, not how
confident you feel, and **not "the moment you believe the work is done"**.

**There is no floor and no exception.** A docs-only change, a one-liner, a typo somebody pointed at —
all revised. A threshold you are allowed to apply yourself is a threshold you will eventually apply
to something that did not qualify, and an informal "skip it just this once" is how a rule quietly
stops existing.

Committing is a **separate** act with its own permission, covered at the end.

---

## What this is, and why it is one pass and not many

**Repetition is not verification.** Reading the same diff a fourth time finds the fourth wording
problem, not the first real defect. Defects are found by doing three specific things once, properly:

1. **reading the change in the order the code actually executes**, so you see what each step receives
   rather than what it was meant to receive;
2. **deliberately constructing what could go wrong** — inputs, states and sequences the author did
   not have in mind — instead of waiting for a problem to announce itself;
3. **running it, with real data, and watching the result.**

The third is the one that decides. Everything else is inference; only running it is evidence. **Work
is not finished because the code looks right. It is finished because you watched it do the right
thing.**

This is deliberately a **sequential** examination, not a fan-out. The code runs in an order; the
revision follows that order. Splitting it across parallel readers loses exactly the thing that makes
it work — one mind holding the whole path from input to result.

---

## Stage 0 — pull first

Bring the working branch up to date by the steps in
[before-starting-work](before-starting-work.md). A branch that moved during the session makes the
whole revision worthless: the diff would be read against a base that no longer exists. Where the
pull brings a conflict, [merge-conflicts](merge-conflicts.md) governs — and nothing in this rule
permits forcing past one.

## Stage 1 — read every change, in execution order

`git status` and `git diff` (plus `--staged`), every file and every hunk. **No change is reported
unread.** Then read them again in the order they run: entry point → what it calls → what that
returns → what is stored → what the caller does with it. Git lists files alphabetically; that
ordering hides the defects that live between two files.

- **Scope.** Every change belongs to the task. Accidental edits, formatter runs over unrelated
  files, leftover experiments, stray or generated files — reverted or split out with a stated
  reason. Anything in the diff you could not honestly describe as part of this task is a problem.
- **No secrets or junk:** no `.env`, credentials, tokens, large binaries, IDE files.

## Stage 2 — interrogate each change

For every changed piece, answer these out loud. A question you cannot answer is a finding.

- **What does this do, in one sentence?** A line you cannot explain in one sentence is one you do not
  understand well enough to ship.
- **What does it assume?** About its inputs, about what ran before it, about what exists, about
  ordering, about the caller. Then: **what happens when each assumption is false?**
- **Who else uses this?** Every changed function, class, endpoint, DTO, schema, prompt or config
  key — find the references and decide whether the change still holds for each. **The list goes into
  the report**, one line per consumer, with what was checked; a consumer you could not check is
  written down as unchecked. This is the step whose absence lets a change made for one caller quietly
  alter another.
- **Contract surfaces.** An interface change is never finished in one place. Where the project
  names its contract surfaces — which document is the source of truth, what must be regenerated,
  which mirror is hand-synced — follow that list; where it names none, work them out from the change
  itself. Source of truth first, then the code, same change.

## Stage 3 — build the failure hypotheses

Do not wait for defects to show themselves. **Write down what could go wrong, then check each one.**

- **Corner cases**, deliberately enumerated: empty, null, zero, one, very many; the boundary and one
  either side; duplicates; out of order; the same call twice; unicode, encoding, timezones; first run
  versus re-run; the largest realistic input.
- **Failure paths.** For every interaction with the outside world — database, network, file system,
  subprocess, model provider — what happens when it fails, hangs, returns garbage, returns half.
- **Silent failures specifically.** A default that hides a missing answer, a caught exception that
  logs nothing, a step that is skipped but still reports success, a fallback indistinguishable from a
  real result. These are the ones that reach users, because nothing announces them.
- **Concurrency and repetition.** Two of these at once; the same request twice; a retry landing after
  the original succeeded.
- **What the change made stale** — a comment, a rule, a table, a document that now describes
  behaviour that no longer exists.

## Stage 4 — run it

**This stage is not optional and cannot be replaced by reasoning.**

- **Execute the changed path with real inputs.** Call the function, hit the endpoint, run the script,
  drive the flow. Feed it test data you chose — including the corner cases from Stage 3 — and
  **compare what came out against what you expected before running it.**
- **Watch the effects, not just the return value.** What was written, what was logged, what was left
  behind. A function that returns the right thing and writes the wrong row has not passed.
- **Run the project's gates** — lint, build, type-check, the tests covering what was touched. Use
  whatever commands the project documents, in its CLAUDE.md, its README or its build scripts; where
  none are written down, find them before running. Scope them to what actually changed: a turn whose
  only change is a note or a log entry touches no code, so no code gate applies to it — that is
  scoping, not an exception, and the verdict says which gates were in scope.
- **Read the output, never the exit code.** A suite that skips a block whose dependency is missing
  still prints success.
- **New tests are watched failing first.** A test that has never failed proves nothing about the
  defect it claims to cover — break the fix in memory, see the test go red, restore it.
- **Non-deterministic behaviour is never called fixed on one passing run.** Anything depending on a
  model, a race, a clock or external data: state how many runs were made; where repetition is
  impossible, the status is "not verified", not "fixed".

**Where there is genuinely nothing to execute** — a change that is only documentation, a rule, a
log entry or a comment — this stage scopes down the same way the gates do: the runnable path is
empty, so there is nothing to run, and the verdict says that plainly instead of claiming a run. That
is scoping, not an exception, and it does **not** extend to a change that touches code, however
small, nor to a documented command or example, which is executed to confirm it works.

**If there is something to run and you cannot run it**, that is a result, not a formality to skip past. Say precisely what could
not be exercised and why — and where the obstacle is a missing key or environment, go through
[working-autonomously](working-autonomously.md) first: search the repository for it, and stop and
ask rather than reporting an unexercised path as working.

## Stage 5 — judge it as an architect

Correct is not sufficient. Apply [code-quality](code-quality.md) to what you just built:

- Does the fix remove the **cause**, or instruct against the symptom? **Say which, in one line.** A
  symptom fix is allowed only with a named reason why the clean one is out of reach, plus a recorded
  follow-up — the same defect patched twice is what this line exists to stop.
- Any crutch — a special case for the failing input, a retry over a race, a swallowed error, a
  loosened check, a second source of truth?
- Any over-engineering — an abstraction with one implementation, an extension point for a variation
  that does not exist, indirection nobody needs?
- Duplicated logic, dead code left behind, names that no longer say what the thing is?
- **Claims nothing backs.** A comment or description asserting a checkable fact — a count, a
  guarantee, a "verified that…" — is a finding unless a test holds it. Either a test pins the number
  or the number does not appear.
- **Promises with no consumer.** A field, flag, column or documented capability introduced here must
  have something in this change that uses it and a test that proves it. Nothing half-built ships
  described as whole.
- **Version ranges you wrote** must not admit a version the rest of the toolchain rejects.

## Stage 6 — the independent grading, once

**The pass that grades the work does not get the author's reasoning.** Run the `change-reviewer`
subagent — defined in `.claude/agents/change-reviewer.md` — with the diff and the requirement, and
nothing else. **This one subagent is standing authorisation:** it is launched here without asking,
and a session-level instruction to avoid subagents does not cover it. A spawn gate that puts the
question to the user is not such an instruction — answer it, never route around it.

**One agent, once.** Not a fan-out, not a panel: the work is sequential and so is its examination.

Do not tell it what you were trying to do, do not defend a choice, do not answer a finding with your
reasoning — check the finding against the code instead.

**Then finish, rather than looping.** Its findings are triaged and fixed; each fix is verified the
way Stage 4 demands — **by running it** — and the reviewer is **not** re-run for a whole new pass
unless the fixes changed behaviour or structure rather than wording. A second full pass over
corrected prose finds new prose to correct, indefinitely, and buys nothing. What gets re-run is the
code.

If the agent cannot be launched at all, say so in the verdict and do the pass deliberately cold:
re-read the diff without consulting your own notes, and try to make each change fail. A cold
self-pass is the fallback, never the plan.

---

## The verdict

One verdict, and it is always **scoped**: what was run, what was only read, and what class of defect
the gates that ran cannot see. "PASS" on its own is not a verdict — it is a mood. Where the automatic
gates use a stand-in for something real (a fake model, a stubbed service, an in-memory store), the
verdict says so, because that is exactly the class of defect that reaches the user instead.

- **PASS** — the change was run and did what it should; everything found is fixed and the fixes were
  themselves run.
- **FIX, then PASS** — real findings → **fix them yourself now**, re-run the affected path, and tell
  the developer plainly what was found and fixed. Problems are never parked and never merely
  reported.
- **STOP** — data loss, a security hole, a broken contract, the change contradicting its requirement,
  or a genuine fork where the choice belongs to the developer. Explain simply and concretely; for a
  fork, [architectural-forks](architectural-forks.md) says how to present it.

---

## The report

Every turn that changed files is announced with: what was reviewed, what was found and fixed in
plain language, the verdict, and which gates ran with their real results. Failures are reported as
failures — never smoothed over. The shape is in [response-style](response-style.md).

**Three lines it always carries**, because they let the developer judge the work without reading it:

- **How each claim was established** — every statement marked **ran it** / **read it** / **assumed
  it**. A report that mixes them silently reads as if everything was run.
- **Who else this touched** — the consumer list from Stage 2, unchecked ones named as unchecked.
- **What would prove it wrong** — one sentence, "this is wrong if …", the cheapest observation that
  would falsify the work.

When the revision turned problems up, they are told in the shape
[response-style](response-style.md) sets out for a problem report: the ground first, then one finding
at a time, each saying how the thing is meant to work, what goes wrong, and what the user actually
gets.

---

## Committing, which has to be asked for

**Do NOT run `git commit` / `git push` / open a PR / merge unless the user has explicitly asked for
it in their own words.** Finishing the work is not permission. A green build is not permission.

**The one merge this does not cover** is bringing a parent branch *into* your own working branch,
which [before-starting-work](before-starting-work.md) requires before the work starts and again
before a requested commit. This ban is about publishing your work — putting it somewhere other
people build on — and taking somebody else's finished work into your own branch publishes nothing
and changes nobody else's history. Merging your branch *into* the parent is publishing, and stays
on the list above.

Once a commit has been asked for:

- **Pull again, then revise again** — the tree may have aged, and the state reviewed must be the
  state that lands.
- **Only a PASS may be committed.** `--no-verify` is forbidden; when a hook fails, fix the cause.
- **No force, ever, on anything shared** — see [merge-conflicts](merge-conflicts.md) for the full
  list and why.
- **The message truthfully describes the whole diff.** Anything it cannot honestly cover is a scope
  problem, not a wording problem. Unrelated changes do not share a commit.
- **The audit trail goes in with the work, wherever the project version-controls it.** Log or audit
  output the repository tracks — `management/logs/` is where the work-audit hook writes it, and
  `.gitignore` says whether this clone keeps it — is part of the commit: never left behind, never
  edited to tidy the diff. **Stage it yourself unless you have checked that something else did:** a
  hook that stages it may not be enabled in this clone, and may deliberately not stage for a
  path-limited commit. Check `git status` and confirm. Where the project ignores that output
  instead, there is nothing to stage and this does not apply.

---

## Checklist

- [ ] The revision ran **before this turn ended, because files changed in it** — no floor, no
      exception for size.
- [ ] Full diff read hunk by hunk, then re-read **in execution order**; scope verified; no secrets or junk.
- [ ] Each change explained in one sentence; its assumptions named, and what happens when each is false.
- [ ] **Consumers enumerated for the report**, unchecked ones named; contract surfaces updated at the source of truth first.
- [ ] Failure hypotheses **written down and checked**: corner cases, failure paths, silent failures, concurrency and repetition, documentation made stale.
- [ ] **The change was run with real test data and produced the expected result** — effects watched, not just return values; gates run and their output read, not their exit code; new tests watched failing first. Where the change was documentation only, the verdict says the runnable path was empty rather than claiming a run.
- [ ] Anything that could not be exercised is named as unverified, after searching the repository for the missing access.
- [ ] Non-deterministic behaviour not called fixed on one run — the number of runs stated, or "not verified".
- [ ] Judged as an architect: cause not symptom (**said which**), no crutch, no over-engineering, no duplicated or dead code, no unbacked claim, no promise without a consumer.
- [ ] **`change-reviewer` run once**, given the diff and the requirement only; its findings fixed and the fixes **run**; no repeat pass over wording.
- [ ] Verdict pronounced and **scoped**: what ran, what was only read, which defect classes the gates cannot see.
- [ ] Report delivered with the three lines: ran / read / assumed, who else was touched, what would prove it wrong.
- [ ] **No commit, push, PR or merge without an explicit request in the user's own words**; a requested commit preceded by a fresh pull and a re-run revision; any version-controlled audit output verified staged, or confirmed to be ignored by the project.
