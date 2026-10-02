# Merge conflicts — resolved by hand, never by force

A conflict is two people's intentions meeting, and every resolution decides whose work survives. A bad
resolution leaves a clean-looking merge and brings back a bug somebody already fixed, weeks later, with
nothing tying it to the merge — so the bar here is higher than anywhere else.

## Never force anything

These are forbidden unless the developer asks for that exact command, in their own words, in this
conversation — "just get it merged" is not that:

- `git push --force` and `--force-with-lease`;
- `git reset --hard` on anything you did not create in this session;
- `git checkout --ours` / `--theirs`, and `git merge -X ours` / `-X theirs` — they discard a side unread;
- `git rebase --skip`, `git rerere` applied blindly, `git commit --no-verify`;
- deleting or recreating a branch to escape a conflict;
- any `-f` / `--force` flag on a shared branch.

If one of these seems genuinely right, that is a fork: put it to the developer with what would be lost
(rule `architectural-forks`).

## Understand both sides before resolving a hunk

For every conflicting hunk, be able to say in one sentence each:

1. What each side was trying to achieve — from the commit behind it (`git log --merge`,
   `git log -L <range>:<file>`, `git blame` on each parent), its message and its diff.
2. Which side is newer, and whether newer is better here — established from history, not assumed from a
   date; a long-lived branch can carry a fix the other side never had.
3. Whether they really conflict — the default answer is both: two additions to one function, a rename on
   one side and a behaviour change on the other.

A hunk whose two sides you cannot explain is a hunk you may not resolve.

## Resolve hunk by hunk

- Decide each hunk on its own evidence. Take a whole file from one side only after reading both versions
  of every hunk in it.
- Never resolve inside someone else's uncommitted work — stop and say what collides.
- Generated files (lock files, generated clients, indexes, migration snapshots): restore one side's
  committed state, finish the merge, then regenerate and check. Never merge them by hand.
- No formatter runs during a merge.

## Prove both intentions survived

- For each side, name the behaviour it added or fixed and check that it still works — its tests, and the
  code path itself where there is no test. Compiling is not evidence.
- Read the merged file as a whole, not as a diff against your side.
- Where a hunk went one way for a reason the code does not show, say so in the merge commit message.

## When to stop and ask

When you cannot tell what one side was trying to do; when both intentions are real and incompatible;
when the resolution would drop somebody's committed work; when you are unsure who owns the file; or when
you are reaching for a forbidden command. Explain both versions in plain language and what each choice
would lose.

## Checklist

- [ ] No force command, no side discarded wholesale, no branch deleted to escape the conflict.
- [ ] Both sides' intentions established from their commits for every hunk; which is newer taken from history.
- [ ] "Both" considered first; one side dropped only where they cannot coexist.
- [ ] Resolved hunk by hunk; never inside someone else's uncommitted work.
- [ ] Generated files regenerated, never hand-merged; no formatter run.
- [ ] Each side's behaviour verified after the merge; non-obvious choices explained in the commit message.
- [ ] Anything unclear or incompatible put to the developer.
