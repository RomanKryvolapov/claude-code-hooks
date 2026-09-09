# For an agent reading this repository

**What this is:** three ready-to-install [Claude Code](https://claude.com/claude-code) hooks, shipped
as committed static binaries for macOS, Linux and Windows plus the Go sources they were built from,
and one behaviour rule.

- `usage-limits` — shows live Anthropic subscription usage (5-hour window, 7-day window, per-model
  weekly buckets) in the model's context and the status line, and gates subagent spawns near a limit.
  Its weekly gauge measures **working time**, not calendar time.
- `work-audit` — writes an append-only record of each session (prompts, choices, per-turn stats,
  files changed, final answer) under `management/logs/`.
- `play-sound` — plays a sound when Claude Code stops or needs attention.
- `.claude/rules/` — nine behaviour rules, of which `session-budget.md` is the one that makes the
  limits hook worth having. The hook supplies numbers; the rule is what tells the model to act on
  them. **Install it too, or the numbers are decoration.** The other eight are independent of the
  hooks; `README.md` says what each is for.
- `.claude/agents/change-reviewer.md` — the independent grading pass the `finishing-work` rule
  requires. Install it whenever you install that rule; the limits gate exempts it by name.
- `.claude/skills/dev-ai-prompt-generation/` — a skill on writing prompts, skills and rules.
  Independent of everything else here.

**There is no runtime to install.** No Node, no Python, no Go. The binaries are committed on purpose.

`README.md` is the full description — what every line of the display means, how the working week is
computed, what the gate does, every config key. Read it if you need to explain the hooks or change
their behaviour. **This file is only about installing them into another project.**

---

## Installing into an arbitrary project

`<target>` is the project root — the folder holding its `.claude` directory. Nothing below needs
network access or a toolchain.

### 1. Copy the files

```sh
mkdir -p <target>/.claude/hooks <target>/.claude/sounds <target>/scripts
mkdir -p <target>/.claude/rules <target>/.claude/agents <target>/.claude/skills
cp .claude/hooks/usage-limits .claude/hooks/work-audit .claude/hooks/play-sound <target>/.claude/hooks/
cp .claude/sounds/notify.wav                            <target>/.claude/sounds/
cp .claude/usage-limits-config.json                     <target>/.claude/
cp .claude/rules/session-budget.md                      <target>/.claude/rules/
cp -r scripts/usage-limits-src scripts/work-audit-src scripts/play-sound-src <target>/scripts/
cp scripts/usage-limits-darwin-arm64 scripts/usage-limits-linux-amd64 scripts/usage-limits-windows-amd64.exe <target>/scripts/
cp scripts/work-audit-darwin-arm64 scripts/work-audit-linux-amd64 scripts/work-audit-windows-amd64.exe       <target>/scripts/
cp scripts/play-sound-darwin-arm64 scripts/play-sound-linux-amd64 scripts/play-sound-windows-amd64.exe       <target>/scripts/
```

The other eight rules, the subagent and the skill are optional and independent of the hooks. Ask
whether they are wanted before copying them — they change how the model behaves, and eight rules is
about 1,400 lines in every session's context:

```sh
cp .claude/rules/*.md          <target>/.claude/rules/
cp .claude/agents/*.md         <target>/.claude/agents/
cp -r .claude/skills/dev-ai-prompt-generation <target>/.claude/skills/
```

`change-reviewer.md` is not optional if `finishing-work.md` is installed — that rule launches it by
name at its grading stage, and without the file the pass silently cannot run.

**If the target already has these hooks**, overwrite the launchers, the binaries and the sources, but
**do not overwrite `.claude/usage-limits-config.json`** — it holds that project's own thresholds.
Merge instead: add any key it lacks, change no key it already has. If it still carries the old
`week_day_weights` key, replace it with `working_week`, carrying each day's number into that day's
`percent` and taking the default hours (`"from": "12:00"`, `"to": "20:00"`).

**If the target keeps its binaries somewhere else** — `scripts/claude-code/` or `.claude/bin/` — put
them there instead and leave its layout alone. The launcher searches all three.

### 2. Wire `.claude/settings.json`

If the target has no `settings.json`, copy this repository's whole. Otherwise merge these into the
existing document, leaving every other setting untouched. Add nothing twice: if an entry already
runs `usage-limits` or `play-sound` for the same event, repoint that entry's `command` rather than
appending a second one.

```json
{
  "statusLine": {
    "type": "command",
    "command": "\"${CLAUDE_PROJECT_DIR}/.claude/hooks/usage-limits\" --mode statusline"
  },
  "hooks": {
    "SessionStart": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "\"${CLAUDE_PROJECT_DIR}/.claude/hooks/usage-limits\" --mode inject --event SessionStart",
            "timeout": 20
          },
          {
            "type": "command",
            "command": "\"${CLAUDE_PROJECT_DIR}/.claude/hooks/work-audit\"",
            "async": true,
            "timeout": 10
          }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "\"${CLAUDE_PROJECT_DIR}/.claude/hooks/usage-limits\" --mode inject --event UserPromptSubmit",
            "timeout": 20
          },
          {
            "type": "command",
            "command": "\"${CLAUDE_PROJECT_DIR}/.claude/hooks/work-audit\"",
            "async": true,
            "timeout": 10
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "Agent|Task|Workflow",
        "hooks": [
          {
            "type": "command",
            "command": "\"${CLAUDE_PROJECT_DIR}/.claude/hooks/usage-limits\" --mode gate",
            "timeout": 20
          }
        ]
      }
    ],
    "Notification": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "\"${CLAUDE_PROJECT_DIR}/.claude/hooks/play-sound\"",
            "timeout": 15
          }
        ]
      }
    ],
    "Stop": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "\"${CLAUDE_PROJECT_DIR}/.claude/hooks/play-sound\"",
            "timeout": 15
          },
          {
            "type": "command",
            "command": "\"${CLAUDE_PROJECT_DIR}/.claude/hooks/work-audit\"",
            "async": true,
            "timeout": 10
          }
        ]
      }
    ],
    "StopFailure": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "\"${CLAUDE_PROJECT_DIR}/.claude/hooks/play-sound\"",
            "timeout": 15
          },
          {
            "type": "command",
            "command": "\"${CLAUDE_PROJECT_DIR}/.claude/hooks/work-audit\"",
            "async": true,
            "timeout": 10
          }
        ]
      }
    ],
    "SessionEnd": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "\"${CLAUDE_PROJECT_DIR}/.claude/hooks/work-audit\"",
            "async": true,
            "timeout": 10
          }
        ]
      }
    ],
    "SubagentStart": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "\"${CLAUDE_PROJECT_DIR}/.claude/hooks/work-audit\"",
            "async": true,
            "timeout": 10
          }
        ]
      }
    ],
    "SubagentStop": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "\"${CLAUDE_PROJECT_DIR}/.claude/hooks/work-audit\"",
            "async": true,
            "timeout": 10
          }
        ]
      }
    ],
    "PreCompact": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "\"${CLAUDE_PROJECT_DIR}/.claude/hooks/work-audit\"",
            "async": true,
            "timeout": 10
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "AskUserQuestion",
        "hooks": [
          {
            "type": "command",
            "command": "\"${CLAUDE_PROJECT_DIR}/.claude/hooks/work-audit\"",
            "async": true,
            "timeout": 10
          }
        ]
      }
    ]
  }
}
```

### 3. Say what you installed — the rules need no wiring

**Copying the files is the whole installation.** Claude Code discovers every `.md` under
`.claude/rules/` at launch, recursively, and loads it with the same priority as `.claude/CLAUDE.md`.
There is no import to add and no line to put in `CLAUDE.md`. The same is true of
`.claude/agents/` and `.claude/skills/`: a subagent is available by its filename, and a skill is
offered by its `description`.

What that does require is **telling the person plainly what you added**, in the report, by name.
Rules are the one part of this install that changes how the model behaves rather than what it can
see, they take effect in the next session without anyone opting in, and every one of them lands in
the context window of every session from then on. Somebody who asked for a usage gauge should not
discover a rule about merge conflicts by having it obeyed.

If a rule should apply only to part of a codebase, give it a `paths:` frontmatter key and it loads
only when Claude touches a matching file:

```markdown
---
paths:
  - "src/api/**/*.ts"
---
```

None of the nine shipped rules carry one, because none of them are about a particular kind of file.

### 4. The two steps that are easy to miss

Both are invisible on the machine doing the install and break the hooks on somebody else's.

**Line endings.** The launchers are POSIX scripts, and a CRLF shebang breaks them on Linux and
macOS. Git converts line endings on Windows checkouts by default, so the target needs a
`.gitattributes` rule. Add these lines if they are not already there:

```
.claude/hooks/* text eol=lf
*.sh text eol=lf

# The audit log is append-only, one file per session: keep both sides of a merge.
management/logs/** merge=union
```

**The executable bit.** On Windows `core.fileMode` is off, so a newly added file records as
non-executable, and the launcher's `[ -x ... ]` test then fails on a Unix clone — the hook silently
does nothing, with no error anywhere. Record it explicitly:

```sh
git -C <target> add --chmod=+x .claude/hooks/usage-limits .claude/hooks/work-audit .claude/hooks/play-sound \
  scripts/usage-limits-darwin-arm64 scripts/usage-limits-linux-amd64 scripts/usage-limits-windows-amd64.exe \
  scripts/work-audit-darwin-arm64 scripts/work-audit-linux-amd64 scripts/work-audit-windows-amd64.exe \
  scripts/play-sound-darwin-arm64 scripts/play-sound-linux-amd64 scripts/play-sound-windows-amd64.exe
```

If git refuses the repository with _"detected dubious ownership"_, run the command with
`-c safe.directory=<target>` rather than changing the user's global git config.

### 5. Check it actually works

Do not skip this: every failure path in these hooks exits 0 and prints nothing, which is right for a
hook and means a broken install looks exactly like a working one.

```sh
CLAUDE_PROJECT_DIR=<target> <target>/.claude/hooks/usage-limits --mode statusline
```

Two stacked lines per limit is a working install. `LIMITS -> N/A` means it ran but has no usage data
— usually not signed in. **Empty output means the launcher found no binary**: check that the
binaries landed in one of `scripts/`, `scripts/claude-code/`, `.claude/bin/`, and that they are
executable.

A `CONFIG` line under the gauges names anything in the config the hook could not use. It should not
be there after a clean install.

Then the sound. It prints nothing either way, so time it — a run that found its binary blocks
while the wav plays, one that found nothing returns at once:

```sh
time CLAUDE_PROJECT_DIR=<target> <target>/.claude/hooks/play-sound
```

A second or two means it worked. A few milliseconds means the launcher found no binary, and the only
symptom you will ever get is silence.

The audit hook reads its event from stdin, so hand it one:

```sh
echo '{"hook_event_name":"SessionStart","session_id":"check0000","source":"startup"}'   | CLAUDE_PROJECT_DIR=<target> <target>/.claude/hooks/work-audit
find <target>/management/logs -type f
```

A file named after the session id should appear, holding one `session start:` line. The folders are
created by the hook; the one named after a person comes from `git user.name`. Delete the check file
afterwards — it is not part of a real session.

---

## Adjusting the working week

The one setting most installs want to change. In `<target>/.claude/usage-limits-config.json`:

```json
"time_zone": "Europe/Sofia",
"working_week": {
  "mon": { "percent": 100, "from": "12:00", "to": "20:00" },
  "tue": { "percent": 100, "from": "12:00", "to": "20:00" },
  "wed": { "percent": 100, "from": "12:00", "to": "20:00" },
  "thu": { "percent": 100, "from": "12:00", "to": "20:00" },
  "fri": { "percent": 100, "from": "12:00", "to": "20:00" },
  "sat": { "percent": 20, "from": "12:00", "to": "20:00" },
  "sun": { "percent": 20, "from": "12:00", "to": "20:00" }
}
```

`percent` is how much those hours count — 100 in full, 0 not at all, 50 as half a day. `from` and
`to` are wall-clock times in `time_zone`. Hours must not run past midnight; `"24:00"` says the end of
the day. Set every day to the same percent from `00:00` to `24:00` to get plain calendar time back.

Ask the person whose machine this is what hours they actually work before changing it — the default
is somebody else's answer.

---

## Do not

- **Do not commit anything** unless the person asked. Copying and wiring is the job; the commit is
  theirs.
- **Do not overwrite a config that already exists.** Its thresholds are a decision somebody made.
- **Do not add a second hook entry** for an event that already has one — both would run.
- **Do not report success without running step 5.** A silent hook is indistinguishable from a
  working one until somebody needs it.
