# work-audit — automatic work audit log

A Claude Code hook that mechanically records the sequence of work into the project's
audit log, so anyone can see what happened in a session without relying on the model
to write a summary.

## What it writes

One append-only file per session at
`management/logs/<author-slug>/<date>/<session8>.log` — a folder per developer, a folder
per day inside it, one file per session (author from `git user.name`; `session8` = first
8 chars of the session id — so parallel sessions never interleave).
Entries are blank-line separated, no timestamps, no token counts — each event writes one
entry in a single append, so entries never interleave and the file reads in the order the
events reached the hook. Each turn records:

- **`prompt:`** — the user's request (blank lines stripped), preceded by a separator.
  A message typed while the turn was still running is queued instead of submitted, so it
  never reaches the hook as an event; it is recovered from the transcript at the end of the
  turn and written as its own entry, **`prompt (typed while the turn was running):`**, right
  before the turn it interrupted. A finished background task arrives through the same event
  as a prompt but nobody typed it, so it is recorded as `background task finished` instead.
- **`question:` / `user chose:`** — for every `AskUserQuestion`, the question and the option
  the user picked. It reads `user chose:` and not `answer:` so it can never be mistaken for
  Claude's own reply below. The pick is read from the structured tool result, with the older
  text form as a fallback; if neither can be parsed the line says `(not captured)` rather than
  inventing an answer.
- **`turn:`** — model, effort, duration, tool count with a breakdown, and subagent count.
  Subagents are counted from the `SubagentStart` events, not from the tool calls, because a
  Workflow fan-out is one `Workflow` tool call that spawns dozens of agents. Tool counts come
  from the session transcript; subagent sidechains are excluded from them.
- **`files:`** — the files Claude edited that turn via its file tools (changes made
  through shell commands or by subagents are not itemized): `+` created, `~` modified,
  `-` deleted, with `+added/-removed` line counts and, for modified files, the exact
  line ranges `+[new lines] -[old lines]` — computed via `git diff -U0 HEAD` (vs the
  last commit).
- **`answer:`** — Claude Code's final "what was done" reply (blank lines stripped; very
  long texts truncated).

It also logs `session start` / `session end` (with total duration), `subagent start` /
`subagent done`, API `error`, and `compaction`. Agents from a Workflow fan-out get no
start/done line — dozens of them would bury the turn they belong to — only the count.

Two counters live outside the log, in a per-session file under the OS temp folder: how many
subagents started since the last turn, and how many queued messages have already been written
(a queued message stays in the transcript for good, with nothing marking it as consumed). They
are throwaway state — a temp sweep mid-session costs a turn its subagent count and can repeat a
queued message once. Nothing that belongs in the repository is kept there.

## Wiring

`.claude/settings.json` registers the launcher `.claude/hooks/work-audit` (async) on:
`SessionStart`, `UserPromptSubmit`, `Stop`, `StopFailure`, `SessionEnd`,
`SubagentStart`, `SubagentStop`, `PreCompact`, and `PostToolUse` (matcher
`AskUserQuestion` only — never per-tool noise).

## Safe by design

Must never take the session down: every error is swallowed and the process exits 0.
**git is optional** — if git is missing or the folder isn't a repo, the file-change
statistics degrade to a bare list of paths while everything else (prompt, turn stats,
answer, session events) is still written.

## How it runs

The POSIX launcher `.claude/hooks/work-audit` `uname`-detects the OS and execs the
matching prebuilt binary committed under `scripts/`
(`work-audit-{darwin-arm64,linux-amd64,windows-amd64.exe}`), so end users need no Go
toolchain. Full workflow and format: `.claude/rules/project-tracking.md`.

## Rebuild

```sh
sh scripts/work-audit-src/build.sh   # cross-compiles all three binaries (needs Go)
```
