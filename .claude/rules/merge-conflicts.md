# Merge conflicts — resolved by hand, never by force

A merge conflict is **two people's intentions meeting**, not two blocks of text disagreeing. Every
resolution decides whose work survives, and getting it wrong reinstates a solution somebody already
replaced. That is how a fixed bug comes back, long after anyone is still watching for it.

The damage is invisible afterwards. A bad resolution leaves a clean-looking merge: no failing build,
no obvious diff, nothing for a reviewer to catch. The defect only reappears in the product weeks
later, and nobody connects it to the merge. **This is why the bar here is higher than anywhere else
in these rules.**

---

## Never force anything

**These commands are forbidden unless the developer asks for that exact command, in their own words,
in this conversation.** Not "resolve the conflict however", not "just get it merged" — the command
itself.

- `git push --force` and `--force-with-lease` — both. The lease makes it safer for you, not for the
  person whose commits vanish.
- `git reset --hard` on anything you did not create in this session.
- `git checkout --ours` / `--theirs`, and `git merge -X ours` / `-X theirs` — these discard one side
  wholesale, unread. They are the single most common cause of the returning-bug failure.
- `git rebase --skip`, `git rerere` applied blindly, `git commit --no-verify`.
- Deleting or recreating a branch to escape a conflict.
- Any `-f` / `--force` flag on a shared branch.

**If you believe one of these is genuinely the right answer, that is a fork** — stop and put it to
the developer with what would be lost, per [architectural-forks](architectural-forks.md). It is
never a step you take to keep moving.

---

## Before resolving a single hunk, understand both sides

A conflict cannot be resolved from the conflict markers alone. They show two texts; you need two
**intentions**.

For every conflicting hunk, establish and be able to say in one sentence each:

1. **What was each side trying to achieve?** Find the commit behind each side — `git log --merge`,
   `git log -L <range>:<file>`, `git blame` on each parent — and read its message and its diff. A
   commit that says "fix the duplicate question" tells you something the code alone does not.
2. **Which one is newer, and does newer mean better here?** Usually the later change knows about the
   earlier one and supersedes it. Sometimes it does not — a long-lived branch can carry an older fix
   the other side never had. Establish it; do not assume from the date.
3. **Are they actually in conflict?** Very often both intentions can survive: two independent
   additions to one function, a rename on one side and a behaviour change on the other. The default
   answer to a conflict is **both**, not one of them.

**A hunk you cannot explain both sides of is a hunk you may not resolve.** Keep looking, or stop and
ask.

---

## Resolve hunk by hunk, by hand

- **One hunk at a time**, deciding each on its own evidence. Taking a whole file from one side is
  allowed only when you have read both versions of **every** hunk in it and can say why each one
  goes that way.
- **Never resolve inside someone else's uncommitted work.** If the conflict lands in a tree holding
  changes you did not make, stop and say what collides — do not decide for them.
- **Generated files are the exception, and it runs the other way:** a lock file, a generated client,
  an index, a migration snapshot is never merged by hand. Restore it to one side's committed state,
  finish the merge, then regenerate it and check the result. Hand-editing generated output produces
  a file that matches neither source.
- **Reformatting during a conflict is forbidden.** A formatter run inside a merge hides the real
  resolution in noise nobody can review.

---

## After resolving, prove both intentions survived

Compiling is not evidence. A resolution can compile perfectly and still have deleted a fix.

- **For each side, name the behaviour it was adding or fixing, and check that behaviour still
  works** — its tests, and the code path itself where a test does not exist. Both sides, not the one
  you were working on.
- **Read the merge result as a whole**, not as a diff against your side. The question is whether the
  file now makes sense, not whether it matches what you had.
- **Where a hunk resolved in favour of one side, say so in the merge commit message** when the
  reason is not obvious from the code. The next person to hit this conflict reads that message.

---

## When to stop and ask

Stop and put it to the developer — with both versions explained in plain language, what each was
for, and what would be lost either way — when:

- you cannot establish what one side was trying to do;
- both intentions are real and genuinely incompatible;
- the resolution would drop somebody's committed work;
- the conflict is in a file whose ownership you are unsure of;
- or you find yourself reaching for any command in the forbidden list.

Asking costs a message. A wrong resolution costs a returning bug that nobody traces back to here.

---

## Checklist

- [ ] No force command used — no force push, no hard reset of others' work, no `--ours`/`--theirs`,
      no strategy option that discards a side, no branch deleted to escape the conflict.
- [ ] For every conflicting hunk, both sides' **intentions** were established from their commits and
      can each be stated in one sentence.
- [ ] Which side is newer was established from history, not assumed from the file.
- [ ] "Both" was considered first; one side was dropped only where they genuinely cannot coexist.
- [ ] Resolved hunk by hunk; a whole file taken from one side only after reading every hunk of both.
- [ ] No conflict resolved inside someone else's uncommitted work.
- [ ] Generated files regenerated, never hand-merged; no formatter run inside the merge.
- [ ] After resolving, **each side's behaviour was verified to still work** — not just that it builds.
- [ ] Non-obvious resolutions explained in the merge commit message.
- [ ] Anything unclear or genuinely incompatible was put to the developer instead of decided.
