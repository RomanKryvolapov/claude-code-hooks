# Architectural forks — stop and raise them, never choose quietly

When a piece of work can reasonably be built more than one way and the choice is expensive to undo,
stop before writing code on either branch and put the fork to the developer with the background and the
options. A fork chosen silently costs nothing in the diff and a rewrite months later; a developer asked
to arbitrate every small choice stops reading, so the filter at the end matters as much as the stop.

## The four signs — any one makes it a fork

Check them before writing code, every time, from evidence:

1. **Something with more than one caller changes behaviour for a caller other than yours** — a shared
   function, client, base class, middleware, config key, a schema several features read. Answer only from
   a list of the callers made now, before the code; a feeling about the risk is not an answer. Editing a
   shared file is not the sign — another caller's behaviour moving is.
2. **The shape of something that outlives the request changes** — a column, a stored document, a message
   payload, an API field, a file format.
3. **The behaviour of something you were not asked to touch changes** — a neighbouring feature, another
   endpoint, another team's surface. Reaching your goal by altering a shared path counts, even when it
   looks like an improvement.
4. **New state appears that outlives the request, or that more than one place writes** — a column, a
   stored document, a cache entry, a persisted flag, a counter, a queued record. A field on an object
   inside one flow, or a variable that dies with the request, is not this.

The signs are worded narrowly on purpose: a stop that fires on everything is ignored. What does not stop
still goes into the consumer list every report carries.

## If no sign is true, weigh it

It is still a fork when two or more designs are genuinely defensible and the choice is hard to reverse
(it means rewriting work built on it, migrating data, or changing a contract outside this repository) or
shapes money, security or privacy.

Not a fork — decide it and say so in one line: names, file layout inside a module, two equivalent library
calls, test structure, formatting, anything the codebase already settled, anything a later change undoes
in an afternoon.

## When you spot one

Before writing code on either branch:

1. **Stop.** Half-built code becomes an argument for keeping it.
2. **Raise the context above the task** — read the requirements that touch the decision, look at how the
   neighbouring parts are built, and name every place the decision reaches.
3. **Find all the options** — including those the system already implies, the hybrid, and "postpone, or
   do the minimum that keeps both open".
4. **Cost each one** — what it takes to build and to live with, what it forecloses, what reversing it a
   year later breaks.

## How to present it

Write for a competent developer who is not in this project: plain language, no file paths, identifiers or
snippets.

- First line: "This is an architectural decision and it needs your call."
- Background: what the system does here today and what changed to raise the question.
- The question, in one decision-shaped sentence, and what depends on the answer.
- Two to four options, each with how it works, its cost to build and to keep, and what it forecloses —
  advantages and disadvantages for every option, yours included.
- Your recommendation, why, and what would change your mind.
- The smallest question that unblocks the work.

## While the answer is missing

- Keep working on everything that does not depend on it, and say which part that is.
- Never pick a branch quietly. If something genuinely cannot wait, take the option cheapest to reverse,
  state the assumption openly and keep it isolated.
- Ask once, then wait and work elsewhere.

## Record the decision

Write it where CLAUDE.md says decisions are recorded ("Where decisions are recorded"): the options
rejected and why, and who decided. The rejected branch is what stops the discussion from restarting.

## Do not ask about everything

This filter applies only to decisions that trip none of the four signs — a sign is a stop, and nothing
below reopens it.

- Reversible within a day → decide.
- The project already made an equivalent decision → follow it; a precedent beats a preference.
- Content with any of the options → it is a preference, not a fork. Choose.
- Unclear only because you have not looked → look first.

Expect a handful of real forks in a project, not one per task.

## Checklist

- [ ] The four signs checked before any code, the callers listed rather than estimated.
- [ ] A fork recognised before code on either branch; nothing half-built.
- [ ] Context raised above the task; every place the decision reaches named.
- [ ] All real options found, the hybrid and the postpone option included, each costed.
- [ ] Presented in plain language: background, question, options with pros and cons, a recommendation, one question.
- [ ] Work continued on what the answer does not block; no branch chosen silently.
- [ ] Decision and rejected options recorded where the project keeps decisions.
- [ ] Reversible, precedented or either-way choices made, not asked.
