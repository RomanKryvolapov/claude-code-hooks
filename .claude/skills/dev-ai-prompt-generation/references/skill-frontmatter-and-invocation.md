## About this file

**What’s inside:** Claude Code skill frontmatter reference — every field, invocation control (`disable-model-invocation` / `user-invocable`), tool pre-approval, model/effort overrides, `paths` scoping, command naming, description listing limits, and where skills live.

**Hub:** [../SKILL.md](../SKILL.md)

Verified against code.claude.com/docs/skills, July 2026. Fields change over time — re-verify per the mandatory docs workflow in [skills-docs-and-technique-table.md](skills-docs-and-technique-table.md).

---

## All frontmatter fields are optional

Claude Code’s contract: **every frontmatter field is optional**; only `description` is _recommended_ (without it, discovery falls back to the first paragraph of the body). For cross-product portability under the Agent Skills open standard (agentskills.io), the safe minimum is still `name` matching the folder + a filled `description`.

| Field                      | What it does                                                                                                                                                                                                                       |
| -------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `name`                     | Display label in listings only. The **command name comes from the directory name** (`.claude/skills/deploy/` → `/deploy`); the one exception is a plugin-root `SKILL.md`, where `name` does set the command.                       |
| `description`              | What the skill does and when to use it — the routing signal for automatic invocation. Put the key use case in the first sentence.                                                                                                  |
| `when_to_use`              | Extra trigger phrases / example requests, appended to `description` in the listing (shares its character cap).                                                                                                                     |
| `argument-hint`            | Autocomplete hint for expected arguments, e.g. `[issue-number]`.                                                                                                                                                                   |
| `arguments`                | Named positional arguments for `$name` substitution in the body (space-separated string or YAML list; names map to positions in order).                                                                                            |
| `disable-model-invocation` | `true` → only the user can invoke (`/name`); the description leaves Claude’s context entirely. Use for side-effect workflows (commit, deploy, send-message) where the user controls timing.                                        |
| `user-invocable`           | `false` → hidden from the `/` menu; only Claude invokes. Use for background knowledge that is not a meaningful user action.                                                                                                        |
| `allowed-tools`            | Tools pre-approved (no permission prompt) while the skill is active. Grants, does **not** restrict — all other tools stay available under normal permissions. Project-level skills need the workspace trust dialog accepted first. |
| `disallowed-tools`         | Tools removed from the pool while the skill is active (e.g. `AskUserQuestion` in an autonomous loop). Clears on the next user message.                                                                                             |
| `model`, `effort`          | Model / reasoning-effort override while the skill is active (rest of the current turn; not saved to settings).                                                                                                                     |
| `context`                  | `fork` → run the skill in an isolated subagent; the body becomes the subagent’s prompt → [skill-dynamic-context-arguments-lifecycle.md](skill-dynamic-context-arguments-lifecycle.md).                                             |
| `agent`                    | Agent type for `context: fork` (`Explore`, `Plan`, `general-purpose`, or a custom `.claude/agents/` type). Default: `general-purpose`.                                                                                             |
| `hooks`                    | Hooks scoped to this skill’s lifecycle.                                                                                                                                                                                            |
| `paths`                    | Glob patterns; the skill auto-loads only when Claude works with matching files (monorepo / area-specific guidance).                                                                                                                |
| `shell`                    | Shell for dynamic-context commands: `bash` (default) or `powershell`.                                                                                                                                                              |

Example — a user-triggered task skill with pre-approved tools:

```yaml
---
name: deploy
description: Deploy the application to production
disable-model-invocation: true
allowed-tools: Bash(npm run build) Bash(git push *)
---
```

## Invocation control (who triggers the skill)

| Frontmatter                      | User invokes | Claude invokes | Description in context |
| -------------------------------- | ------------ | -------------- | ---------------------- |
| (default)                        | Yes          | Yes            | Always                 |
| `disable-model-invocation: true` | Yes          | No             | No                     |
| `user-invocable: false`          | No           | Yes            | Always                 |

Pick by content type: **reference content** (conventions, domain knowledge Claude should apply when relevant) → default or `user-invocable: false`. **Task content** (step-by-step actions with side effects) → `disable-model-invocation: true`.

## Description listing limits

- Per skill, `description` + `when_to_use` are truncated at **1,536 characters** in the skill listing — front-load the key use case.
- The whole listing shares a budget (~1% of the context window); on overflow the least-used skills lose their descriptions first. `/doctor` shows which are shortened or dropped.

## Where skills live

| Level      | Path                               | Notes                                                    |
| ---------- | ---------------------------------- | -------------------------------------------------------- |
| Project    | `.claude/skills/<name>/SKILL.md`   | Committed with the repo.                                 |
| Personal   | `~/.claude/skills/<name>/SKILL.md` | All your projects.                                       |
| Enterprise | managed settings                   | Org-wide. Name clashes: enterprise > personal > project. |
| Plugin     | `<plugin>/skills/<name>/`          | Namespaced `plugin:skill`, cannot clash.                 |

Nested `.claude/skills/` below the working directory load on demand (monorepo packages) and get directory-qualified names (`apps/web:deploy`) on clashes. Skill directories are watched — edits to `SKILL.md` apply live within the session.

**Custom commands are merged into skills**: `.claude/commands/deploy.md` and `.claude/skills/deploy/SKILL.md` both create `/deploy`; the skill wins a name clash and adds supporting files, frontmatter, and automatic invocation.
