# Before starting work

**Every message you are about to act on opens the same way**: bring the branch up to date, read
what arrived, and let it change the work in hand. Then, for work that reads or changes the
repository, check that what the project stands on is still current. All of it happens _before_
planning, editing or reviewing — not after, and not "if there is time".

Skipping any of it has the same shape of cost, and it is never visible on the day. A stale branch
means work built on a base that no longer exists. Commits pulled but not read mean a bug fixed twice
or a requirement obeyed in the version it had last week. A stale dependency means work built on a
version whose vendor has stopped supporting it — which shows up as a surcharge, a security hole, or
a migration three times harder than it would have been.

---

## 1. Pull before you act on the message

**The trigger is the developer's message, not the size of the work.** Every message you are about to
act on opens by bringing the branch up to date — before planning it, before reading code for it,
before answering a question about the state of the work, and long before the first edit. Not once a
session, and not "when the task looks big enough to bother": the repository moves while you are
reading, and a message answered against yesterday's branch is answered about a repository nobody
has.

It is a standing default rather than a judgement made afresh each time, because the judgement is
what fails: a stale answer costs nothing visible on the day, so a threshold applied by feel drifts
until it is never met. **Two things bound it, and neither of them is an opinion about how big the
task looks.** A fetch already done in this turn is not done again — one pull per message, not one
per step. And a message that needs nothing from this repository at all needs no pull, which is the
whole of the exception at the foot of this file. [session-budget](session-budget.md) does not
disagree: a fetch and a look at what came in cost a handful of tokens, and the reading that follows
is scoped to the area the message touches rather than to everything that landed.

### Which branch to bring in

1. **Fetch, then look.** `git fetch` and see how far behind you are and what is coming in.
   A quiet fetch costs nothing; it also tells you whether anything relevant to your task moved.
2. **When you are on the project's main working branch** — the branch feature work is cut from,
   whatever this project calls it — fast-forward it with `git pull --ff-only` on a clean tree, and
   pull nothing else. Where the project promotes along a chain of branches, the ones downstream of
   the working branch carry release merges of their own, and pulling one of those backwards drags
   that history into the trunk. Downstream is not a parent.
3. **When you are on a branch cut from it, bring that parent in as well.** Your own branch being
   level with its own remote says nothing about the branch it was cut from — the parent is where
   everybody else's work lands, and it is the base your change still has to be correct against. So
   merge the parent's tip into your branch: the working branch above, and where your branch was cut
   from another side branch, that one first. Never a rebase of anything already pushed.

   **Into your own branch, never the other way.** This is the one merge
   [finishing-work](finishing-work.md)'s ban does not reach, and that file says so: the ban is about
   publishing your work — a commit, a push, a pull request, a merge into a branch other people
   build on — and taking the parent into your own branch publishes nothing. Merging your branch
   *into* the parent still has to be asked for, in the developer's own words.

   **Three preconditions, and each one stops the merge rather than being worked around.** A **clean
   tree**, exactly as the fast-forward above needs one: a merge into a dirty tree is a different
   operation with different failure modes, and "When the tree is not clean" below is written about
   the fast-forward only. A **parent you can name**: git records nothing about where a branch was
   cut from, so where the answer is not plain from the branch's own name, the task, or CLAUDE.md,
   **ask rather than guess** — merging the wrong branch in costs far more than an answer given a
   minute later. And a **branch to merge into**: on a detached HEAD there is none, so say so and
   stop. Where the merge conflicts, [merge-conflicts](merge-conflicts.md) governs and nothing here
   permits forcing past it.

The same pull happens again at the other end of the work — before any turn that changed files ends, and
again if a commit is asked for — so that the state reviewed is the state that lands. That half
belongs to [finishing-work](finishing-work.md).

### Read what arrived, and let it change what you are about to do

**A pull nobody read is a pull that did not happen.** The point is not a green fast-forward; it is
that the work you are about to start already knows what landed. Go through the incoming commits and
the files they touch, and answer three questions before writing anything.

**What moved under the task itself** — the files your work will touch, the contract it depends on,
the schema or migration it assumed, the neighbouring feature it calls. Someone may have already done
part of it, or moved the thing you were about to change, or changed it in a way that makes your plan
wrong rather than merely late.

**What arrived in whatever tracks the work.** This is the one that catches people out, because the
record of the work moves faster than the code does. Where the project keeps bugs and tasks — in the
repository, or in an issue tracker beside it — entries are filed continuously by whoever is working
alongside you, and several of them will be in the area you are about to touch. Read the ones that
came in. A new bug may be the same defect you are about to fix from the other end. A new task may
already own the change you were about to make. An entry whose status moved may mean the work is
finished, cancelled, or waiting on somebody — and starting it anyway is the most expensive way to
find that out. Where the project keeps nothing of the kind, this question is nothing to do.

**What arrived in the statement of what correct means.** Where the project keeps requirements, specs
or acceptance criteria under version control, read the ones your work touches rather than working
from the version you last saw — the gap between them is invisible in the code and decides whether
what you build is right. A requirement that changed is a change to what *correct* means, and it
outranks any older description of the same behaviour; a new one is a surface that did not exist when
your task was written. Where the project keeps nothing of the kind, the code and its tests are that
statement.

Then **say in one line what came in and what it changes for the work in hand** — including "nothing
relevant arrived", which is a finding rather than an omission. A silent pull tells the developer
nothing about whether anybody looked. Where the turn ends in the four-part work report
[response-style](response-style.md) sets out, that line belongs inside its first part rather than
beside it; on a turn that changes no file it stands on its own. Where the project has a tracker, that
line is also the one place an entry id belongs in a reply, because the id is what the developer would
look it up by.

### When the tree is not clean

Uncommitted work does not excuse skipping the pull — it changes how you do it.

- **Fetch first and compare** the incoming file list against your own changed files. No overlap →
  fast-forward proceeds cleanly with the changes still in the tree.
- **Overlap on generated files** (index files, a regenerated client, a lock file) → restore them to
  the committed state, pull, then regenerate. Never resolve a generated file by hand.
- **Overlap on real work** → stop and tell the developer what collides, in plain language, before
  touching anything. Do not force, do not reset, do not discard their changes to make the pull go
  through.

### When the fetch itself fails

A fetch that fails is not permission to carry on without one — say so and stop, the same as any
other blocked step.

One failure reads as something it is not: **"repository not found" on a private repository almost
always means the wrong account, not a missing repository.** The host hides the existence of a
private repository from anyone who cannot see it, so a failed authentication and a deleted
repository give the identical message. Where several accounts are signed in, only one is active at a
time and the git credential helper uses that one — an account switch elsewhere silently breaks every
fetch here. Check which account is active and whether it can see the repository before concluding
anything, and tell the developer what you found rather than switching their accounts around on a
guess.

---

## 2. Check what the project stands on

Version currency is part of starting work, not a separate project. Two duties, one mandatory and
one proportional.

### A platform you deploy on is never left behind — where there is one, this is not optional

A runtime past its vendor's support window is the one kind of staleness that costs money and
options rather than comfort: it is billed at a penalty rate where the vendor charges for it, stops
receiving security patches, and is eventually upgraded on the vendor's schedule rather than yours.
What counts as "the runtime" differs by project — a managed cluster, a language line, a base image, a
database engine, a hosting platform — and each of those has its own vendor clock.

- **Where the project deploys onto a platform with a vendor support window, check it on every piece
  of work that reads or changes the repository** — not on every message, which is section 1's trigger
  and not this one's: what is deployed, and what the newest supported version is. Compare them.
  Where the project writes down what its runtime is and what "supported" means for it, that document
  decides; where it does not, name what you take the runtime to be and check that. Where the project
  deploys nothing, this duty is empty and the section below is the whole of it.
- **If it is behind, that is a finding you act on, not one you note.** Bring the repository's
  declarations current in the same change — every place the version is written, so local and
  deployed stay on the same line.
- **Moving a running environment is announced first.** It is an operation on a live system, it goes
  one major version at a time, and the pieces around it do not follow on their own. Say what it will
  take and get the developer's go-ahead, then do it. What is never acceptable is leaving it behind
  in silence.

### Everything else: update what is quick, report what is not

Dependencies, base images, build actions, language and tool versions.

- **Quick — just do it, as part of the task.** A version bump that is a number change plus green
  gates: patch and minor releases, and a major whose only effect is a rename the tooling itself
  migrates. Do it, run the gates, mention it in one line.
- **Not quick — tell the developer and offer.** A bump needing code changes, a configuration
  rewrite, a data migration, or one whose result only a human eye can judge (anything that changes
  how the product looks). Name what is out, what it would take, and what standing still costs.
  Then let them decide.
- **A ceiling the project set on purpose is not stale.** Where a version is deliberately held back
  and the reason is written down, respect it and raise it as a question rather than lifting it.
- **Check what is actually published, not what is advertised.** A registry's "latest" and "beta"
  labels are a curated pointer, not the list of releases — a stable line can sit unlabelled behind
  a newer major. Read the published versions before concluding a version does not exist.

**Never skip silently.** Either the update is done, or the developer is told it is available. A
version left behind with nobody informed is the one outcome this section exists to prevent.

---

## When this does not apply

**Only where there is no repository in play at all** — a question about a language, a tool or an
idea, answered without opening anything here. Everything else pulls first, reading included: a
question about the state of the work or a piece of code is answered *from* the repository, and
answering it from a stale one is how a developer is told something is still open after somebody
finished it.

The version check in section 2 is the narrower half: it belongs to work that reads or changes the
repository, not to every message.

## Checklist

- [ ] Branch brought up to date **before acting on the message at all** — not once a session, and
      not only for work that looked big enough. Once per message, not once per step: a fetch already
      done in this turn is not repeated.
- [ ] On a branch cut from another, the parent merged in too, up the chain to the working branch —
      into your own branch only, on a clean tree, and never a rebase of anything already pushed.
- [ ] A parent that could not be named with confidence, or a detached HEAD, was **asked about**
      rather than guessed at; a branch downstream of the working branch was not mistaken for a parent.
- [ ] Incoming commits actually read for what they change under the task: the files it touches, and
      **anything the project records about what the work is meant to do** — where it records it.
- [ ] What came in, and what it changes for the work in hand, said in one line — "nothing relevant"
      included.
- [ ] Dirty tree handled by comparing file lists first; generated files regenerated, never hand-merged.
- [ ] A collision with the developer's own uncommitted work reported, not forced through.
- [ ] A failed fetch reported, not worked around; "repository not found" on a private repository
      treated as the wrong account until proven otherwise.
- [ ] Where the project deploys a runtime, it was checked against its vendor's supported versions;
      if behind, the repository brought current and the live move proposed — never left behind in
      silence.
- [ ] Other versions: the quick ones updated with the gates re-run, the slow ones named to the
      developer with what they would take.
- [ ] Deliberate version ceilings respected and raised as questions, not lifted.
- [ ] Version claims made from the list of published releases, not from a registry's labels.
