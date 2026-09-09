## About this file

**What’s inside:** Dynamic context injection (`` !`command` `` preprocessing), argument substitution (`$ARGUMENTS`, `$N`, `${CLAUDE_SKILL_DIR}`…), running a skill in a forked subagent (`context: fork`), skill content lifecycle under compaction, and evaluating skills against a no-skill baseline.

**Hub:** [../SKILL.md](../SKILL.md)

Verified against code.claude.com/docs/skills, July 2026.

---

## Dynamic context injection — `` !`command` ``

`` !`command` `` runs a shell command **before** the skill content is sent to the model; the output replaces the placeholder. The model receives actual data, never the command. This is preprocessing, not a tool call — no permission prompt, no round-trip.

```yaml
---
name: pr-summary
description: Summarize changes in a pull request
context: fork
agent: Explore
allowed-tools: Bash(gh *)
---

- PR diff: !`gh pr diff`
- Changed files: !`gh pr diff --name-only`

Summarize this pull request…
```

Rules:

- The inline form is recognized only when `!` starts a line or follows whitespace (`` KEY=!`cmd` `` stays literal).
- Multi-line commands: a fenced block opened with ` ```! `.
- One pass over the original file; command output is **not** re-scanned for further placeholders.
- `shell: powershell` in frontmatter switches the shell on Windows; `disableSkillShellExecution: true` (settings, for managed policies) replaces each command with a policy notice instead of running it.
- Prefer injection over instructions like “first run git diff”: the model gets the data pre-inlined, with zero chance of skipping the step.

## Argument substitution

| Variable                | Expands to                                                                                                      |
| ----------------------- | --------------------------------------------------------------------------------------------------------------- |
| `$ARGUMENTS`            | Everything typed after the skill name. If absent from the body, arguments are appended as `ARGUMENTS: <value>`. |
| `$ARGUMENTS[N]`, `$N`   | 0-based positional argument (shell-style quoting: `"hello world"` is one argument).                             |
| `$name`                 | Named argument declared in the `arguments` frontmatter list (names map to positions in order).                  |
| `${CLAUDE_SESSION_ID}`  | Current session id — logging, session-scoped files.                                                             |
| `${CLAUDE_EFFORT}`      | Active effort level (`low`…`max`) — adapt skill instructions to it.                                             |
| `${CLAUDE_SKILL_DIR}`   | Directory of this `SKILL.md` — reference bundled scripts portably.                                              |
| `${CLAUDE_PROJECT_DIR}` | Project root — reference project-local scripts regardless of where the skill is installed.                      |

Escape a literal `$` before a digit, `ARGUMENTS`, or a declared name with a backslash: `\$1.00`.

## Run in a subagent — `context: fork`

`context: fork` runs the skill in an isolated context: the **skill body becomes the subagent’s prompt**, with no access to the conversation history. `agent:` picks the execution environment (`Explore`, `Plan`, `general-purpose`, or a custom type). Explore/Plan skip `CLAUDE.md` and git status — a forked skill on `agent: Explore` sees only its own content plus the agent’s system prompt.

Only fork skills that contain an **explicit task**. Guideline-only content (“use these API conventions”) forked to a subagent returns nothing useful — the subagent gets guidelines but no actionable prompt.

Two directions of skill ↔ subagent composition:

| Approach                       | System prompt           | Task                                                        |
| ------------------------------ | ----------------------- | ----------------------------------------------------------- |
| Skill with `context: fork`     | From the agent type     | The `SKILL.md` content                                      |
| Subagent with a `skills` field | The subagent’s own body | Claude’s delegation message (skills preloaded as reference) |

## Content lifecycle (why bodies must be standing instructions)

- On invocation the **rendered** body enters the conversation as one message and **stays for the rest of the session**; the file is not re-read on later turns.
- Write **standing instructions** (“always X when Y”), not one-shot steps (“now do X”) — the content keeps applying long after the invoking turn.
- Every line is a recurring token cost for the whole session — the conciseness bar is higher than for a one-off prompt.
- **Compaction:** each invoked skill is re-attached after summarization keeping its first **5,000 tokens**; all re-attached skills share a **25,000-token** budget, most recently invoked first — older skills can drop out entirely. Front-load the rules that must survive; re-invoke a critical skill after compaction.
- “The skill stopped influencing behavior” usually means the content is still in context but losing to other signals — strengthen the `description` and instructions, or enforce deterministically with hooks.

## Evaluating a skill

Triggering ≠ working. Measure separately: (1) does it invoke on the prompts it should, (2) is the output right when it does.

- **Baseline comparison:** run a few realistic prompts in **fresh sessions** with the skill available vs disabled (`skillOverrides` in settings), and compare. Fresh sessions matter — authoring-session context masks gaps in the written instructions.
- The **skill-creator plugin** (official marketplace) automates the loop: test cases in `evals/evals.json` inside the skill directory, isolated per-case subagent runs, assertion grading, a with/without benchmark (pass-rate gain vs token/time overhead), blind A/B between two skill versions, and description tuning against should-trigger / should-not-trigger prompts.

Cross-reference: generic prompt-eval methodology (judges, rubrics, regression gates) → [evals-llm-as-judge-and-regression-gates.md](evals-llm-as-judge-and-regression-gates.md).
