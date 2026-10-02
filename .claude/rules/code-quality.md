# Code quality — clean solutions, causes not symptoms

Judge every change as an architect would: not "does the symptom go away" but "is this what this code
should look like now". A change that works and leaves the codebase worse has failed; half an hour on a
clean solution beats ten minutes on one that costs an hour of fixes afterwards.

## Find the cause; never patch the symptom

Find the mechanism that produces the misbehaviour — the exact place the wrong value, order or state
comes from — before designing a fix. Never fix with:

- a special case for the input that failed, or a guard where the crash happened instead of where the bad
  state was created;
- a retry, delay, re-fetch or "refresh twice" that makes a race look solved;
- a swallowed error — a catch that logs nothing, a default hiding a missing answer, a fallback
  indistinguishable from success;
- a widened type or a loosened check;
- a second source of truth kept in sync by hand.

When the clean fix is genuinely out of reach (a migration, someone else's decision, work far outside the
task), write down what the cause is, what the temporary measure does and what removing it takes — at the
place it sits and in the report. A temporary measure nobody knows about is a permanent one.

## When a model answers wrong, the model is not the cause

The model is far more capable than anything the product asks of it, so a wrong answer means it was handed
the wrong data, or the right data in the wrong shape. Another sentence in the prompt is the symptom fix:
it often moves the next run's number and makes the prompt less stable, and a prompt that needs a new rule
to hold the last one up is pointing at a cause elsewhere.

Read what actually reached the model — every block and message, in send order, with the real values — and
ask, in this order:

- **What was it asked to do, and in what order?** An instruction that fights how the conversation or task
  naturally goes is lost some of the time, whatever it says. Reorder it.
- **What is each value called, and which block is it in?** A value in a block carrying the product's own
  authority reads as an instruction; beside the user's words it reads as a fact.
- **What was it not given?** Anything it must infer, it will sometimes infer differently; hand over what
  the system already knows.
- **What contradicts what?** Two instructions that cannot both hold are settled differently from run to run.

The same applies to a screen people keep misusing: what were they given, in what order, and what were
they left to work out — not another warning, confirmation or tooltip on top.

When the fix is a change to prompt text — a block moved, a value named, two instructions reconciled —
write it with the `dev-ai-prompt-generation` skill. Adding a sentence is not one of its techniques.

## Clean, and no more than that

- **Not under-designed:** no layers of special cases, duplicated logic, names that lie, state nobody owns,
  or a function that grew an unrelated responsibility.
- **Not over-designed:** no abstraction with one implementation, interface "for later", unrequested
  configuration switch, factory producing one thing, or indirection a reader must trace through four
  files. Generality nobody asked for is harder to remove than a missing abstraction is to add.
- The test for both: a competent developer new to this code can say what the change does and why it is
  where it is, without you explaining.

## SOLID, as judgement

- One reason to change per class, module or function.
- Extend rather than edit where behaviour actually varies — never for a variation nobody has.
- A subtype is usable wherever its base is.
- Narrow interfaces: no consumer depends on methods it never calls.
- Depend on abstractions across the boundaries the project already draws (a port and its adapter, a
  domain layer and its infrastructure); do not invent new ones for one change.
- No duplicated logic, no dead code, names that say what the thing is, failures that are loud.

## When it is wrong, rewrite it

If the structure cannot be made clean by editing — or this would be the fourth patch on the same spot —
rewrite it, from scratch if needed. The limits: keep the behaviour others depend on (find the consumers
first — rule `architectural-forks`, sign 1), verify it as thoroughly as new code, and stay inside the
task.

## When the design is unclear, decide from the user's side

Between defensible implementations, pick what the person using it will experience as **transparent**
(they can tell what happened and why), **simple** (fewest moving parts, no mode they did not ask for) and
**predictable** (the same action does the same thing every time). Never decide by guessing what somebody
would prefer.

## What the build enforces

Where the project gates its code — linters, suppression baselines, ratchets — CLAUDE.md lists the gates
under "Gates" and the project's code-quality skill, where it has one, says what each checks. Two things
bind wherever they exist:

- A file you touch is yours, its suppressed lint debt included: clear its entries, never add to them.
- A baseline, threshold or count is never raised to make a change pass.

## Checklist

- [ ] The mechanism behind the defect found and named before the fix.
- [ ] The fix removes the cause: no special case, retry over a race, swallowed error, loosened check or second source of truth.
- [ ] An unavoidable temporary measure written down where it sits and in the report.
- [ ] A wrong model answer diagnosed from what reached it in send order; no sentence added to prop up another.
- [ ] Neither under- nor over-designed; readable without the author.
- [ ] Boundaries, responsibilities and dependencies follow the project's structure.
- [ ] Code that could not be made clean by editing was rewritten, its consumers found first.
- [ ] Judgement calls settled on transparency, simplicity and predictability.
- [ ] Touched files left with no more lint debt; no baseline or threshold raised.
