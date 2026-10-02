# Before starting work — pull, read what arrived, check versions

Every message you act on opens the same way: bring the branch up to date, read what arrived and let it
change the work in hand; for work that reads or changes the repository, also check that what the
project stands on is still current. All of it before planning, editing or reviewing — a stale branch,
an unread commit or an unsupported runtime costs nothing visible on the day and a great deal later.

## 1. Pull before acting on the message

- The trigger is the message, not the size of the work: before planning it, before reading code for it,
  before answering a question about a tracker entry. Once per message, not once per step.
- Skip it only when no repository is in play — a question answered without opening anything here.

### Which branch to bring in

1. `git fetch`, and look at how far behind you are and what is coming in.
2. On the working branch CLAUDE.md names under "Branch and runtime": `git pull --ff-only` on a clean
   tree, and pull nothing else. Branches downstream of it carry release merges — downstream is not a
   parent.
3. On a branch cut from it: also merge the parent's tip into your branch — the working branch, and first
   any side branch yours was cut from. Never rebase anything already pushed. This merge goes into your
   own branch only; merging your branch into the parent publishes your work and needs the developer's
   words (rule `finishing-work`). Three preconditions, each a stop rather than something to work around:
   - a clean tree;
   - a parent you can name — git records nothing about where a branch was cut from, so where the branch
     name, the task or CLAUDE.md does not make it plain, ask;
   - a branch to merge into — on a detached HEAD, say so and stop.

   A conflict is handled by rule `merge-conflicts`.

### Read what arrived

A pull nobody read did not happen. Before writing anything, go through the incoming commits and answer:

- **What moved under the task** — the files it touches, the contract, schema or migration it assumes, the
  neighbouring feature it calls. Part of it may already be done, moved, or made wrong.
- **What arrived in the trackers** — new bugs and tasks in the area, entries whose status moved. A new bug
  may be the defect you are about to fix; a new task may already own your change.
- **What arrived in the requirements** — wherever the project keeps them: its `req-*` skills, a requirements
  document, or the skill that defines what a correct run is. A changed requirement changes what correct
  means and outranks any older tracker entry about the same behaviour.

Then say in one line what came in and what it changes for the work — "nothing relevant" included. In a
work report that line belongs in its first part, and it is where a tracker id belongs in a reply.

### When the tree is not clean

Fetch first and compare the incoming files with your changed files:

- No overlap → fast-forward with the changes in place.
- Overlap on generated files (indexes, a generated client, a lock file) → restore them to the committed
  state, pull, regenerate. Never resolve a generated file by hand.
- Overlap on real work → stop and tell the developer what collides before touching anything. No force, no
  reset, nothing of theirs discarded.

### When the fetch fails

A failed fetch is a stop, not permission to continue. "Repository not found" on a private repository
almost always means the wrong account is active, not a missing repository: check which account the
credential helper uses and whether it can see the repository, and report what you found rather than
switching accounts on a guess.

## 2. Check what the project stands on

On every piece of work that reads or changes the repository — not on every message.

### The deployed runtime is never left behind

- Compare what is deployed with the newest version its vendor supports; CLAUDE.md names the runtime and
  its vendor clocks under "Branch and runtime". A runtime past its support window costs money, security
  patches and options.
- Behind → bring every place the version is declared current in the same change, so local and deployed
  stay on one line.
- Moving a live environment is announced first: say what it takes, get the go-ahead, move one major
  version at a time. Leaving it behind in silence is never acceptable.

### Everything else: update what is quick, report what is not

- Quick — a patch or minor bump, or a major whose only effect is a rename the tooling migrates: do it as
  part of the task, run the gates, mention it in one line.
- Not quick — code changes, a configuration rewrite, a data migration, or a result only a human eye can
  judge: name what is out, what it would take and what standing still costs, and let the developer decide.
- A version held back on purpose, with the reason written down, is not stale: raise it as a question.
- Read the list of published versions, not a registry's "latest" label.
- Never skip silently: either the update is done or the developer is told it is available.

## Checklist

- [ ] Branch brought up to date before acting on the message; once per message.
- [ ] On a cut branch, the parent merged in — into your own branch, on a clean tree, never a rebase of pushed work.
- [ ] A parent you could not name, or a detached HEAD, asked about rather than guessed.
- [ ] Incoming commits read for the task's files, the tracker entries in its area and the requirements that moved.
- [ ] What came in, and what it changes, said in one line.
- [ ] A dirty tree handled by comparing file lists; generated files regenerated; collisions with the developer's work reported.
- [ ] A failed fetch reported; "repository not found" treated as the wrong account first.
- [ ] The deployed runtime checked against its vendor's support; if behind, declarations updated and the live move proposed.
- [ ] Quick bumps done with the gates re-run, slow ones named, deliberate ceilings respected, published versions checked.
