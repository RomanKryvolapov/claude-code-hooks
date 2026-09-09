# Stop at architectural forks

Some decisions are an edit. Some are a rewrite. This rule is about the second kind.

**When you notice that a piece of work can reasonably be built in more than one way, and the
choice is expensive to undo — stop before writing the code, study the whole picture, and put the
fork in front of the developer with the background and the options laid out.** Choosing silently
and moving on is the failure this prevents: the cost does not show up in the diff, it shows up
months later as work that has to be thrown away.

The opposite failure is just as real: a developer asked to arbitrate every small choice stops
reading. See "Do not ask about everything" at the end — the filter is the important half of this
rule.

## What counts as a fork

### The four signs — answered from evidence, and any one of them is enough

**Check these before writing code, every time.** Three of them are answered by looking. The first is
answered from a list you must actually produce — which is what keeps it from sliding back into a
feeling, and why it is worded as it is.
If any one is true, the decision **is** a fork and goes to the developer — however small the diff
looks, however obvious the answer feels, however much it is "just how it has to work".

1. **Something with more than one caller changes behaviour for a caller other than the one you are
   serving.** A shared function, a shared client, a base class, a middleware, a config key, a schema
   several features read. Merely editing a shared file is not the sign — the question is whether
   anybody else's behaviour moves. **You are allowed to answer that only from a list of the callers,
   made now — before the code, not at review time.** [finishing-work](finishing-work.md) demands the
   same list again later, but a list produced after the branch is built arrives too late to stop it.
   Answering from a sense of the risk, with no list in front of you, is not answering it: the list is
   the mechanical part, and it is what keeps this sign from becoming a judgement call again.
2. **The shape of something that outlives the request changes.** Anything stored, persisted,
   cached, queued, or sent across a boundary: a column, a stored document, a message payload, an
   API field, a file format.
3. **The behaviour of something you were not asked to touch changes.** A neighbouring feature, a
   different endpoint, another team's surface. Reaching a goal by altering a shared path is the
   common case here, and it is a fork even when the alteration looks like an improvement.
4. **New state is introduced that outlives the request, or that more than one place writes.** A
   column, a stored document, a cache entry, a persisted flag, a counter, a queued record — where it
   lives, who owns it and what happens when it is missing are all decisions, and all of them are
   expensive to move later. A field on an object inside one flow, or a variable that dies with the
   request, is not this.

**Why mechanical.** The weighing version of this test fails in one predictable direction: raising a
fork stops the work, so a mind optimising for finishing finds reasons the sign does not apply. A
worked example: the way a model is asked for a structured answer gets switched in the one shared
place every such request passes through. Sign 1 is plainly true — several unrelated features use
that place and every one of them starts being asked differently — and yet the change goes in as an
implementation detail, with the other users of it never examined. It could break them outright, and
nobody would know until a user did.

**Why signs 1 and 4 are worded narrowly.** Taken at their widest — "touches a shared file", "adds a
field" — they fire on most ordinary work, and a stop that fires on everything is one nobody reads.
The narrow wording keeps the stops rare enough to carry information. Nothing is lost by narrowing:
what no longer stops still appears in the consumer list every report carries, which costs the
developer no attention at all. A stop is expensive and must be worth it; a disclosure is free and
should be exhaustive.

### If none of the four is true, then weigh it

The four signs are the floor, not the ceiling. Absent all four, a decision is still a fork when
**two or more designs are genuinely defensible** and it is either **hard to reverse** — undoing it
means rewriting work built on it, migrating data, or changing a contract outside this repository —
or it **shapes money, security or privacy**: what is charged, what is trusted, what is stored about
a person.

**Not a fork — decide it yourself and say so in one line:** names, file layout inside a module, two
equivalent library calls, test structure, formatting, anything the codebase already settled by
precedent, anything a later change undoes in an afternoon.

> Rule of thumb: **any of the four signs → stop. Otherwise, if the wrong choice costs a rewrite,
> stop; if it costs an edit, decide.**

## What to do when you spot one

**Before writing any code on either branch:**

1. **Stop.** Do not start on one option "just to see" — half-built code is an argument for keeping it.
2. **Raise the context above the current task.** Read the requirements that touch this decision, not
   only the task you were handed. Look at how the neighbouring parts are actually built today. Find
   every place the decision would reach, and name those places. A fork judged from inside one file
   is judged wrong.
3. **Find all the options, not the first two.** Include the ones the system already implies: if
   there are several existing mechanisms of the same kind, each is a candidate. Include the hybrid,
   and include "postpone / do the minimum that keeps both open" — often the best answer.
4. **Cost each option honestly**: what it takes to build, what it takes to live with, what it rules
   out, and what breaks if the decision is reversed a year later.

## How to present it

Write for a competent developer who is **not** in this project. They must be able to answer without
reading the code. Plain language, no file paths, no identifiers, no snippets.

- **Say what it is, first line.** "This is an architectural decision and it needs your call."
- **Background.** What the system does here today, and what changed to make the question arise.
  A few sentences — enough that the options make sense.
- **The question.** One sentence, decision-shaped: "we must decide how X is done".
- **What depends on the answer.** What gets built differently, and what becomes hard later.
- **The options — usually two to four.** Each: how it works in a couple of sentences, what it costs
  to build, what it costs to keep, what it forecloses. Advantages and disadvantages both, for every
  option, including the one you prefer.
- **Your recommendation.** Pick one, say why, and say what would change your mind. A fork presented
  without a recommendation pushes the whole analysis back onto the developer.
- **What you need from them.** The smallest question that unblocks the work.

## While the answer is missing

- **Keep working on everything that does not depend on it**, and say which part that is.
- **Never pick a branch quietly to keep moving.** If some work genuinely cannot wait, choose the
  option that is cheapest to reverse, state the assumption in the open, and keep it isolated.
- **Do not re-ask.** One clear presentation; then wait, and work elsewhere.

## Record the decision

A fork settled in a chat and never written down gets re-litigated by the next person — or by you,
next month. Write it **where this project keeps decisions** — an ADR directory, a decisions section
in `CLAUDE.md`, a note beside the work it came out of, whatever the project already uses. If the
project has not settled on a place, ask where it should go rather than inventing one. Record the
options that were rejected and _why_, and who decided. The rejected branch is the valuable part: it
is what stops the same discussion from restarting.

## Do not ask about everything

Raising a fork costs the developer real attention, and
[working-autonomously](working-autonomously.md) is explicit that a task is yours from beginning to
end: you stop only for a decision that genuinely is not yours, and only after trying to settle it
yourself. **This filter applies only to decisions that trip none of the four signs** — a sign is a
stop, and no question below reopens it.

- Is it reversible within a day? → decide yourself.
- Has the project already made an equivalent decision? → follow it; a precedent beats a preference.
- Would you be content with any of the options? → then it is not a fork, it is a preference. Choose.
- Is it only unclear because you have not looked? → look first. Most "forks" dissolve on reading.

Expect a handful of real forks in a project, not one per task. If you are raising them weekly and
none of them trips a sign, the filter is wrong.

## Checklist

- [ ] **The four signs were checked by looking, before any code:** callers **listed** and their
      behaviour judged from that list, outliving shapes named, untouched-feature behaviour
      considered, state that outlives the request noticed. Any one true → stopped.
- [ ] Recognised the fork **before** writing code on either branch; nothing was half-built to "see".
- [ ] Context raised above the current task: requirements read, neighbouring implementation looked
      at, every place the decision reaches named.
- [ ] All real options listed — including the hybrid and the postpone-it option — not just the first two.
- [ ] Each option costed: build, live-with, foreclosed, cost of reversing.
- [ ] Presented in plain language an outsider could act on: background, the question, what depends
      on it, options with pros **and** cons, a recommendation, and the one question to answer.
- [ ] Work continued on everything the answer does not block; no branch chosen silently.
- [ ] Decision and rejected options recorded where the project keeps decisions.
- [ ] Filter applied: reversible, precedented or acceptable-either-way choices were made, not asked.
