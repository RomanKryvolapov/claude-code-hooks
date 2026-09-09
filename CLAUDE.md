# claude-code-hooks

Three ready-to-install Claude Code hooks, shipped as committed static binaries (macOS, Linux,
Windows) with the Go sources beside them, plus the behaviour rule that makes the first of them
useful. No runtime to install — no Node, no Python, no Go.

- **`usage-limits`** — live Anthropic subscription usage (5-hour window, 7-day window, per-model
  weekly buckets) injected into the model's context and drawn in the status line, plus a gate that
  refuses subagent spawns near a limit. Its weekly gauge measures **working time** — configurable
  hours and a weight per weekday — not calendar time.
- **`work-audit`** — an append-only log of what actually happened in each session: the prompts, the
  choices made in interactive questions, per-turn model and tool and subagent counts, the files
  changed with line ranges, and the final answer. Written mechanically, so it does not depend on the
  model summarising itself honestly.
- **`play-sound`** — a sound when Claude Code stops or needs attention.

**`.claude/rules/session-budget.md` is the other half of the limits hook.** The hook supplies the
numbers; the rule is what tells the model to size work to the task, when a subagent is worth
spawning, and what each zone allows. Copied without it, the numbers are decoration.

**To install all of this into another project, follow [AGENTS.md](AGENTS.md).** It is written for
you: the files to copy, the `settings.json` to merge, how to make the rule actually load, the two
invisible steps that break the hooks on somebody else's machine (line endings and the executable
bit), and how to check the install took.

**To understand or change what they do, read [README.md](README.md).** It covers every line of the
display, how the working week is computed, the gate's thresholds, what the audit log records, and
every configuration key.

The Go sources are `scripts/usage-limits-src/`, `scripts/work-audit-src/` and
`scripts/play-sound-src/`. Every tunable default of the limits hook is declared in
`scripts/usage-limits-src/constants.go`; the config file only overrides it. Rebuild with
`sh scripts/<name>-src/build.sh`, test with `go -C scripts/usage-limits-src test ./...`.
