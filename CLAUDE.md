# claude-code-hooks

A whole `.claude/` for a project, ready to copy into one: three hooks shipped as committed static
binaries (macOS, Linux, Windows) with the Go sources beside them, nine behaviour rules, a reviewing
subagent, a skill about writing prompts, and the `settings.json` that wires it together. No runtime
to install — no Node, no Python, no Go.

**What is in the repository, all of it equally part of the point:**

- **`usage-limits`** — live Anthropic subscription usage (5-hour window, 7-day window, per-model
  weekly buckets) injected into the model's context and drawn in the status line, plus a gate that
  refuses subagent spawns near a limit. Its weekly gauge measures **working time** — configurable
  hours and a weight per weekday — not calendar time.
- **`work-audit`** — an append-only log of what actually happened in each session: the prompts, the
  choices made in interactive questions, per-turn model and tool and subagent counts, the files
  changed with line ranges, and the final answer. Written mechanically, so it does not depend on the
  model summarising itself honestly.
- **`play-sound`** — a sound when Claude Code stops or needs attention.
- **`.claude/rules/`** — nine rules that decide how the model works: `session-budget`,
  `before-starting-work`, `working-autonomously`, `architectural-forks`, `code-quality`,
  `finishing-work`, `merge-conflicts`, `response-style` and `web-search-when-in-doubt`. README.md
  describes each one in full.
- **`.claude/agents/change-reviewer.md`** — the independent grading pass `finishing-work` launches:
  it gets the diff and the requirement and nothing else, and is told to refute rather than confirm.
- **`.claude/skills/dev-ai-prompt-generation/`** — a hub plus 30 reference files on writing prompts,
  skills, rules and `CLAUDE.md`; unlike a rule, it costs nothing until it is used.
- **`.claude/settings.json`**, **`.claude/usage-limits-config.json`** and **`.gitattributes`** — the
  wiring, the one file meant to be edited, and the two attribute lines (LF launchers, union merge for
  the log) whose absence only ever breaks somebody else's machine.

**The rules are a set, not a menu.** Eight of the nine link to each other by name, and the subagent
applies the quality rule by name, so a rule installed alone points the model at files that are not
there. Only `web-search-when-in-doubt` refers to nothing else.

**Across the groups there are four ties, and they are the only ones.** `session-budget` is the other
half of the limits hook — the hook supplies numbers, the rule is what makes the model act on them,
and copied without it the numbers are decoration. `finishing-work` launches `change-reviewer` by
name, so that rule without that file has a grading stage that silently cannot run. The limits gate
exempts `change-reviewer` by name, so the pass a revision depends on is never the spawn refused in a
hot zone. And `finishing-work` expects the audit log — what `work-audit` writes — to be staged with
the work wherever a project tracks it. Otherwise the hooks, the rules and the skill are independent
of each other.

**Rules load by themselves.** Claude Code reads every `.md` under `.claude/rules/` at launch, with
the same priority as `.claude/CLAUDE.md` — no import and no line here is needed. That also means all
nine are in the context window of every session; a rule scoped to part of a codebase would carry a
`paths:` frontmatter key instead, and none of these do.

**To install all of this into another project, follow [AGENTS.md](AGENTS.md).** It is written for
you: the files to copy, the `settings.json` to merge, how to make the rules actually load, the two
invisible steps that break the hooks on somebody else's machine (line endings and the executable
bit), and how to check the install took.

**To understand or change what any of it does, read [README.md](README.md).** It describes every
component in full — every line of the display, how the working week is computed, the gate's
thresholds, what the audit log records, every configuration key, and what each rule, the subagent and
the skill are for.

The Go sources are `scripts/usage-limits-src/`, `scripts/work-audit-src/` and
`scripts/play-sound-src/`. Every tunable default of the limits hook is declared in
`scripts/usage-limits-src/constants.go`; the config file only overrides it. Rebuild with
`sh scripts/<name>-src/build.sh`, test with `go -C scripts/usage-limits-src test ./...`.
