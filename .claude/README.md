# Claude Code — commands and settings reference

Two large parts:

1. **Slash commands** — everything you type with a `/` at the start of a line (for example `/resume`).
2. **`settings.json`** — all configuration keys, with types, JSON examples and explanations.

Current for **Claude Code 2.1.287** (October 2026). Checked against the official documentation
at code.claude.com (Commands, All settings, Settings, Environment variables, Hooks, Memory, Mods) and against `claude --help`.
Some commands are not always visible — it depends on platform, plan and environment (for example `/desktop` is macOS and Windows x64 only,
with a subscription; `/setup-bedrock` is hidden from the menu until `CLAUDE_CODE_USE_BEDROCK=1` is set, but it still works when typed in full).
Installed version: `claude --version`.

---

# Part 1. Slash commands

Notation: **[Skill]** — a built-in skill (essentially a prompt script), **[Workflow]** — a built-in
multi-agent workflow. Arguments after the name are marked with brackets: `<required>`, `[optional]`, `a|b` — a choice.
Aliases (in the "Aliases" column) call the same command. A command is recognized only at the start of a message; the text after it is
its arguments. The exception is skills: `/skill-a /skill-b text` loads all the skills named in a row (up to six) and passes the text to each.
A command sent while Claude is replying waits until the turn ends; some run immediately, for example `/status`, `/tasks`, `/usage`, and in
fullscreen so do dialogs such as `/theme` and `/help`.

## Sessions and conversation

| Command       | Arguments      | Aliases         | What it does                                                                                                                                                                                                                  |
| ------------- | -------------- | --------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `/clear`      | `[name]`       | `/reset` `/new` | A new conversation with a clean context; the name labels the old one in the `/resume` list. The old one stays on disk. To free up context without changing the conversation, use `/compact`.                                  |
| `/resume`     | `[ID or name]` | `/continue`     | Go back to an earlier conversation by ID or name, or pick one from a list. Background sessions are marked `bg`; you can open a running one right from here — the current conversation moves to the background.                |
| `/branch`     | `[name]`       | —               | Branch the conversation from the current point; the original is kept (to go back, use `/resume`). Handy for trying a different path without losing your work so far.                                                          |
| `/fork`       | `[prompt]`     | —               | Copy the conversation into a new **background session** and keep working here. With a prompt, the copy starts on it right away; before editing code it is told to create its own worktree (except for edits made "in place"). |
| `/subtask`    | `<task>`       | —               | A background subagent that inherits the whole conversation: it works on the task while you carry on; the result comes back into this conversation.                                                                            |
| `/rename`     | `[name]`       | —               | Rename the session (the name shows in the prompt bar). With no argument, a name is generated from the topic of the conversation.                                                                                              |
| `/recap`      | —              | —               | A one-line summary of the session on request (the automatic recap after you have been away is controlled by the `awaySummaryEnabled` key).                                                                                    |
| `/btw`        | `[question]`   | —               | A short side question that does not go into the conversation. With no question, it shows your earlier side questions.                                                                                                         |
| `/export`     | `[file]`       | —               | Export the conversation as plain text: with a name, straight to a file; without one, a "clipboard or file" dialog.                                                                                                            |
| `/copy`       | `[N]`          | —               | Copy Claude's last reply (`N` — the Nth from the end). If there are code blocks, you pick a block; `w` in the picker writes to a file (handy over SSH).                                                                       |
| `/background` | `[prompt]`     | `/bg`           | Send the current session to the background (background agent) and free up the terminal; you can add one more instruction. To follow it, use `claude agents`.                                                                  |
| `/stop`       | —              | —               | Stop the current background session (only from inside it); the transcript and worktree remain. To detach without stopping it, use `/exit`.                                                                                    |
| `/exit`       | —              | `/quit`         | Exit the CLI; in an attached background session, detach (the session itself keeps running).                                                                                                                                   |

## Context and memory

| Command        | Arguments          | Aliases               | What it does                                                                                                                                                               |
| -------------- | ------------------ | --------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `/context`     | `[all]`            | —                     | Context usage as a colored grid, plus tips on what to unload (heavy tools, bloated memory). In fullscreen the breakdown is collapsed — `all` expands it.                   |
| `/compact`     | `[instructions]`   | —                     | Compress the conversation history into a short summary, freeing up context. You can say what to focus on.                                                                  |
| `/autocompact` | `[auto\|<tokens>]` | —                     | The auto-compact window: how full the context must be before it starts compacting on its own (`500k` or `auto`). Saved to user settings; with no argument, a dialog opens. |
| `/memory`      | —                  | —                     | Edit `CLAUDE.md` files, turn auto memory on and off, and view its entries.                                                                                                 |
| `/init`        | —                  | —                     | Create the project's `CLAUDE.md`. With `CLAUDE_CODE_NEW_INIT=1`, an interactive flow (it asks about skills, hooks, personal notes).                                        |
| `/add-dir`     | `<path>`           | —                     | Add a working directory so its files are accessible in this session. Its `.claude/` configuration is mostly not picked up.                                                 |
| `/cd`          | `<path>`           | —                     | Move the session to another working directory, keeping the conversation; project settings are then read from the new directory.                                            |
| `/rewind`      | —                  | `/checkpoint` `/undo` | Roll **code and/or conversation** back to a checkpoint, or compress the conversation starting from a chosen message.                                                       |

## Model and performance

| Command         | Arguments                                    | Aliases | What it does                                                                                                                                                                       |
| --------------- | -------------------------------------------- | ------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `/model`        | `[model]`                                    | —       | Switch the model and save it as the default for new sessions. With no argument, a list (`s` on a row — this session only; left/right arrows — effort level).                       |
| `/effort`       | `[level\|auto\|status\|ultracode [on\|off]]` | —       | Effort level: from `low` to `xhigh`, `max` (this session only) or `auto`; `status` prints the current one. `ultracode on` / `off` — turn ultracode on / off for the session.       |
| `/fast`         | `[on\|off]`                                  | —       | Fast mode: faster output at a higher token price; only on Opus 5.5, Opus 5 and Opus 4.8.                                                                                           |
| `/advisor`      | `[model\|off]`                               | —       | Attach an advisor — a second model that gives advice at key moments of a task: `fable`, `opus`, `sonnet` or a full ID; `off` turns it off. With no argument, you pick from a list. |
| `/plan`         | `[description]`                              | —       | Enter plan mode (Claude reads and plans; edits are blocked until the plan is approved); with a description, start on that task right away.                                         |
| `/goal`         | `[condition\|clear]`                         | —       | Set a goal: Claude keeps working turn after turn until the condition is met. With no argument, show the current one; `clear`, `stop`, `off`, `reset`, `none`, `cancel` remove it.  |
| `/output-style` | `[style]`                                    | —       | Show the output styles or switch to one (`/output-style concise`).                                                                                                                 |

## Code review and quality

| Command            | Arguments                                                                       | Aliases   | What it does                                                                                                                                                                                                                                                                            |
| ------------------ | ------------------------------------------------------------------------------- | --------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `/code-review`     | `[low\|medium\|high\|xhigh\|max\|ultra] [--fix] [--comment] [PR\|branch\|path]` | `/review` | **[Skill]** Review the current diff (or a PR, branch or path) for bugs; depending on the model and level, also for simplifications. `--fix` — apply the findings, `--comment` — post them on the PR/MR, `ultra` — a deep cloud review. With no level, the last one you entered is used. |
| `/simplify`        | `[target]`                                                                      | —         | **[Skill]** Find and apply simplifications: four reviewers in parallel (reuse, simplification, efficiency, level of abstraction). It does not look for bugs — `/code-review` does that.                                                                                                 |
| `/security-review` | —                                                                               | —         | Check the branch's changes against the default branch of `origin` for vulnerabilities: injection, authorization, data leaks. Requires an `origin` remote.                                                                                                                               |
| `/diff`            | —                                                                               | —         | View the changes in the working tree, including Claude's edits.                                                                                                                                                                                                                         |
| `/ultrareview`     | `[PR or branch]`                                                                | —         | A deep multi-agent review in a cloud sandbox; the same as `/code-review ultra`. On Pro and Max there are 3 free runs, after that you need usage credits.                                                                                                                                |

## Development and workflows

| Command                     | Arguments                                                                                                                      | Aliases | What it does                                                                                                                                                                                                                                                                      |
| --------------------------- | ------------------------------------------------------------------------------------------------------------------------------ | ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `/run`                      | —                                                                                                                              | —       | **[Skill]** Launch the project's app and check a change "live", not only with tests.                                                                                                                                                                                              |
| `/verify`                   | —                                                                                                                              | —       | **[Skill]** Confirm that a change works: build it, run it, watch the result. Runs only when invoked manually.                                                                                                                                                                     |
| `/run-skill-generator`      | —                                                                                                                              | —       | **[Skill]** Teach `/run` and `/verify` to build, launch and "drive" the app from a clean environment: it records the recipe as a project skill.                                                                                                                                   |
| `/batch`                    | `<instruction>`                                                                                                                | —       | **[Skill]** A large change across the whole codebase: a plan of 5–30 independent pieces and, once approved, one background subagent per piece in its own worktree; each implements its piece, runs the tests and publishes its change. Requires git (or a `WorktreeCreate` hook). |
| `/debug`                    | `[description]`                                                                                                                | —       | **[Skill]** Turn on debug logging for the session and work out a problem from the log.                                                                                                                                                                                            |
| `/fewer-permission-prompts` | —                                                                                                                              | —       | **[Skill]** Scan transcripts, find frequent safe calls (read-only Bash, MCP) and build an allowlist in `.claude/settings.json` so permission prompts appear less often.                                                                                                           |
| `/update-config`            | `[request]`                                                                                                                    | —       | **[Skill]** Describe a change (allow a command, set a variable, add a hook) and Claude edits the right `settings.json`. For theme and model, use `/config`.                                                                                                                       |
| `/claude-api`               | `[migrate\|upgrade\|managed-agents-onboard\|prompt-audit\|cost-optimize\|build-eval\|hillclimb\|preserved-thinking-migration]` | —       | **[Skill]** Reference material for the Claude API and Managed Agents in the project's language; it loads on its own when `anthropic` / `@anthropic-ai/sdk` is imported. The subcommands are described in the skills documentation.                                                |
| `/deep-research`            | `<question>`                                                                                                                   | —       | **[Workflow]** A fan-out of web searches, source checking and a report with links. Runs only when invoked manually.                                                                                                                                                               |
| `/workflow-authoring`       | —                                                                                                                              | —       | **[Skill]** Reference for writing dynamic workflow scripts: API, resuming, patterns, examples. Requires workflows to be enabled.                                                                                                                                                  |
| `/dataviz`                  | `[request]`                                                                                                                    | —       | **[Skill]** Guidance for charts and dashboards: choosing a form to fit the data, color by role, checking the palette for contrast and color blindness.                                                                                                                            |

## Artifacts and design

Artifacts are available on the Anthropic API; on Amazon Bedrock, Google Cloud Agent Platform, Microsoft Foundry and Claude Platform on AWS they are not, and the commands in this section are unavailable there.

| Command                  | Arguments | Aliases | What it does                                                                                                                                                                  |
| ------------------------ | --------- | ------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `/artifacts`             | —         | —       | A list of your artifacts and those shared with you: attach one to the session, open it in the browser or copy its link.                                                       |
| `/artifact-capabilities` | —         | —       | **[Skill]** Reference for the runtime capabilities of a published artifact (calling connectors, downloading a file). Claude usually loads it on its own.                      |
| `/artifact-diagramming`  | —         | —       | **[Skill]** Tips for diagrams in artifacts: when one is needed, what to draw, how to write inline SVG that reads well in both light and dark themes.                          |
| `/design`                | `[brief]` | —       | **[Skill]** Screen mockups, landing pages and posters as artboards on one canvas (a Claude Design artifact); edits in the browser save automatically, with export to PNG/PDF. |
| `/design-login`          | —         | —       | Authorize access to the design system for `/design-sync` through your claude.ai account.                                                                                      |
| `/design-sync`           | `[hint]`  | —       | **[Skill]** Convert the repository's React design system and upload it to Claude Design. The first sync of a large repository can take hours.                                 |
| `/slides`                | `[brief]` | —       | **[Skill]** A presentation as a Claude Slides artifact, built from your brief.                                                                                                |

## Subagents, background tasks, automation

| Command        | Arguments             | Aliases      | What it does                                                                                                                                                                                     |
| -------------- | --------------------- | ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `/agents`      | —                     | —            | Since v2.1.198 it only reminds you: Claude creates and edits subagents when you ask, or you edit `.claude/agents/` or `~/.claude/agents/` yourself.                                              |
| `/tasks`       | —                     | `/bashes`    | The background work of the current session, including subagents that have already finished.                                                                                                      |
| `/list-agents` | —                     | `/peers`     | Subagents, teammates and other Claude Code sessions that Claude can send messages to, with the names used to address them. Only where cross-session messaging is enabled.                        |
| `/workflows`   | —                     | —            | Workflow progress: view, pause, resume, and save running and finished ones.                                                                                                                      |
| `/loop`        | `[interval] [prompt]` | `/proactive` | **[Skill]** Repeat a prompt while the session is open (`/loop 5m …`). With no interval, Claude picks the pace itself; with no prompt, the built-in maintenance prompt or your `loop.md` is used. |
| `/schedule`    | `[description]`       | `/routines`  | Create, update, list or run routines — scheduled agents in the cloud; Claude walks you through the setup in a dialog.                                                                            |
| `/autofix-pr`  | `[prompt]`            | —            | A cloud session watches the PR of the current branch and pushes fixes when CI fails or reviewers comment. Requires `gh` and access to cloud sessions.                                            |

## Configuration and interface

| Command             | Arguments                                     | Aliases          | What it does                                                                                                                                                                                                   |
| ------------------- | --------------------------------------------- | ---------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `/config`           | `[key=value ...]`                             | `/settings`      | The settings panel: theme, model, output style and so on. `key=value` sets an option without the panel (`/config theme=dark`); `/config --help` lists the accepted keys.                                       |
| `/permissions`      | —                                             | `/allowed-tools` | Allow / ask / deny rules, working directories, recent auto mode denials; the **Auto mode** tab holds the classifier rules.                                                                                     |
| `/auto-mode-setup`  | —                                             | —                | Draft `autoMode.environment` entries from the project and recent sessions and save them to user settings. Requires Pro, Max or Team.                                                                           |
| `/sandbox`          | —                                             | —                | Toggle sandbox mode (supported platforms only).                                                                                                                                                                |
| `/theme`            | —                                             | —                | Color theme: `auto`, light/dark, color-blind-friendly, ANSI, or your own from `~/.claude/themes/` or plugins.                                                                                                  |
| `/keybindings`      | —                                             | —                | Open the keybindings file.                                                                                                                                                                                     |
| `/terminal-setup`   | —                                             | —                | Shift+Enter for a newline in VS Code, Cursor, Devin Desktop, Alacritty, Zed; Option+Enter in Apple Terminal; clipboard access for `/copy` in iTerm2.                                                           |
| `/statusline`       | —                                             | —                | Set up the status line: describe what you want, or with no arguments it is built from your shell prompt.                                                                                                       |
| `/voice`            | `[hold\|tap\|off]`                            | —                | Voice input: `hold` — talk while holding the key, `tap` — press to start, press again to stop, `off` — turn it off. Requires a Claude.ai account.                                                              |
| `/tui`              | `[default\|fullscreen]`                       | —                | The interface renderer, with a restart (the conversation is kept); `fullscreen` — flicker-free. With no argument, shows the active one.                                                                        |
| `/color`            | `[color\|default]`                            | —                | The prompt bar color for the session: `red`, `blue`, `green`, `yellow`, `purple`, `orange`, `pink`, `cyan`; `default` resets it, with no argument a random one is picked.                                      |
| `/scroll-speed`     | —                                             | —                | Mouse wheel scroll speed: a "ruler" for previewing right in the dialog (fullscreen only, not in the JetBrains terminal).                                                                                       |
| `/focus`            | —                                             | —                | Focus view: your last prompt, a one-line summary of tool calls and the final reply (fullscreen only).                                                                                                          |
| `/hooks`            | —                                             | —                | Browse the configured hooks by event.                                                                                                                                                                          |
| `/skills`           | —                                             | —                | A list of skills. `t` — sort by tokens, `Space` / `Enter` — change a skill's visibility, `Esc` — save and close.                                                                                               |
| `/skill-doctor`     | —                                             | —                | How much context each skill costs and how often it is used, so you can find the unneeded ones. Requires v2.1.252+ and feature-flag fetching.                                                                   |
| `/plugin`           | `[subcommand]`                                | —                | Plugins and marketplaces: with no argument, a menu; or `list`, `install`, `enable`, `disable`.                                                                                                                 |
| `/plugin-authoring` | —                                             | —                | **[Skill]** The reference Claude follows when it writes a mod (see the notes); Claude loads it itself when you ask for a mod. A skill of a built-in plugin, it can be turned off in `/plugin`; since v2.1.287. |
| `/reload-skills`    | —                                             | —                | Rescan the skill and command directories without restarting the session.                                                                                                                                       |
| `/reload-plugins`   | `[--force]`                                   | —                | Apply plugin changes without a restart. If reloading would reset the prompt cache, the command warns and skips it — `--force` goes ahead.                                                                      |
| `/import`           | `[codex\|gemini\|cursor] [--dry-run] [--yes]` | —                | Import configuration from OpenAI Codex, Gemini CLI or Cursor: instructions, MCP servers, commands, subagents, skills. `--dry-run` — preview only.                                                              |

## MCP and integrations

| Command             | Arguments                                                      | Aliases | What it does                                                                                                                                                                                                             |
| ------------------- | -------------------------------------------------------------- | ------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `/mcp`              | `[reconnect (<server>\|all)\|enable\|disable [<server>\|all]]` | —       | Manage MCP servers and their OAuth authorization. With no argument, an interactive list; `reconnect` — reconnect a server that has dropped, `reconnect all` — every server that failed or is waiting for authentication. |
| `/ide`              | —                                                              | —       | IDE integrations and their status.                                                                                                                                                                                       |
| `/chrome`           | —                                                              | —       | Claude in Chrome settings.                                                                                                                                                                                               |
| `/claude-in-chrome` | `[task]`                                                       | —       | **[Skill]** Do a task in the browser through Claude in Chrome: check a page, fill in a form, read the console. Requires the integration to be enabled (`claude --chrome`).                                               |

## Account and authentication

| Command               | Arguments | Aliases | What it does                                                                                                                                                             |
| --------------------- | --------- | ------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `/login`              | —         | —       | Sign in to your Anthropic account (or switch accounts).                                                                                                                  |
| `/logout`             | —         | —       | Sign out of your account.                                                                                                                                                |
| `/setup-bedrock`      | —         | —       | Amazon Bedrock wizard: authentication, region, model pins. Hidden from the menu until `CLAUDE_CODE_USE_BEDROCK=1` is set; type it in full.                               |
| `/setup-vertex`       | —         | —       | Google Cloud Agent Platform (Vertex) wizard: authentication, project, region, model pins. Hidden from the menu until `CLAUDE_CODE_USE_VERTEX=1` is set; type it in full. |
| `/install-github-app` | —         | —       | Install the Claude GitHub App for a repository (optionally with a GitHub Actions workflow and secrets). github.com only.                                                 |
| `/install-slack-app`  | —         | —       | Install the Claude Slack app (OAuth in the browser).                                                                                                                     |
| `/privacy-settings`   | —         | —       | View and change privacy settings (Pro and Max only).                                                                                                                     |

## Usage and cost

| Command               | Arguments | Aliases          | What it does                                                                                                                               |
| --------------------- | --------- | ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| `/usage`              | —         | `/cost` `/stats` | Session cost, plan limits and activity statistics; on paid plans, a breakdown of what is using up the limit. `/stats` opens the Stats tab. |
| `/usage-credits`      | —         | —                | Set up usage credits (formerly `/extra-usage`) or request them from an admin when you hit the limit.                                       |
| `/rate-limit-options` | —         | —                | What to do when you hit the claude.ai limit: wait and continue automatically, add credits or change plan. Requires a subscription.         |

## Remote work and environments

| Command           | Arguments | Aliases | What it does                                                                                                                                    |
| ----------------- | --------- | ------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| `/desktop`        | —         | `/app`  | Continue the current session in the Claude Code Desktop app. macOS and Windows x64 only; requires a Claude subscription.                        |
| `/teleport`       | —         | `/tp`   | Pull a cloud session into this terminal: pick one from a list, then its branch and conversation are fetched. Requires a claude.ai subscription. |
| `/remote-control` | —         | `/rc`   | Open this session for control from claude.ai or the Claude app (from another device). Requires a claude.ai subscription.                        |
| `/remote-env`     | —         | —       | The default cloud environment for cloud sessions started from the CLI.                                                                          |
| `/web-setup`      | —         | —       | Connect GitHub for cloud sessions using the credentials of the local `gh`.                                                                      |

## Diagnostics and help

| Command     | Arguments               | Aliases    | What it does                                                                                                                                                                                                                                                                                                                                |
| ----------- | ----------------------- | ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `/help`     | —                       | —          | Help and a list of the available commands.                                                                                                                                                                                                                                                                                                  |
| `/status`   | —                       | —          | The Status tab in settings: version, model, account, API connectivity and the settings sources that were loaded (the `Setting sources` line). Works even while Claude is replying.                                                                                                                                                          |
| `/doctor`   | `[prompt-audit [path]]` | `/checkup` | **[Skill]** Checks the installation and settings, with the option to fix problems: duplicate installs, `PATH`, broken settings files, unused skills/MCP/plugins, slow hooks. `prompt-audit` — an audit of `CLAUDE.md`, skills and other instructions for outdated content and contradictions. In the terminal, `claude doctor` (read-only). |
| `/bug`      | `[report]`              | `/share`   | A bug report or "share the conversation": you choose how much history to attach and confirm before it is sent.                                                                                                                                                                                                                              |
| `/feedback` | `[report]`              | —          | Feedback on Claude Code — the same dialog and the same consent as `/bug`. If Claude has prepared draft feedback, with no argument it opens their queue.                                                                                                                                                                                     |
| `/heapdump` | —                       | —          | A JS heap dump written to `~/Desktop` (or the home directory) for diagnosing memory. The `.heapsnapshot` contains the whole conversation and credentials — do not send it, attach only the `-diagnostics.json`. Hidden from the menu — type it in full.                                                                                     |
| `/insights` | —                       | —          | An HTML report on your recent sessions on this machine: projects, how you work, where you run into difficulty, what is worth trying.                                                                                                                                                                                                        |

## Information and documentation

| Command            | Arguments | Aliases           | What it does                                                                                                                                       |
| ------------------ | --------- | ----------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| `/release-notes`   | —         | —                 | The changelog by version: pick a version interactively or show them all.                                                                           |
| `/powerup`         | —         | —                 | Short interactive lessons on Claude Code features, with animated demos.                                                                            |
| `/mobile`          | —         | `/ios` `/android` | A QR code for installing the Claude mobile app.                                                                                                    |
| `/radio`           | —         | —                 | Claude FM lo-fi radio in the browser (without a browser, it prints the stream URL).                                                                |
| `/passes`          | —         | —                 | Share a free week of Claude Code with friends; shown only if your account is eligible.                                                             |
| `/upgrade`         | —         | —                 | Open the page for upgrading to a higher plan in the browser (not shown on Enterprise).                                                             |
| `/stickers`        | —         | —                 | Order Claude Code stickers.                                                                                                                        |
| `/team-onboarding` | —         | —                 | A team onboarding guide based on your last 30 days of usage: markdown that a colleague pastes as their first message (on paid plans, also a link). |

## Notes on commands

- MCP servers can expose their prompts as commands: in the menu they appear as `/server:prompt (MCP)`, and you can also type `/mcp__<server>__<prompt>`.
- Custom commands are skills: `.claude/skills/<name>/SKILL.md`. The old format `.claude/commands/<name>.md` still works and gives the same `/name`.
- Mods (since v2.1.287, on by default) are plugins whose code runs inside Claude Code: their own panes and a band above the prompt,
  `/commands` that run immediately, without a Claude turn, and hooks into tool calls. Which ones are loaded shows in `/plugin`
  (the `N mod active` line); built-in ones are on the Installed → Built-in tab. Among the built-in ones, `cc-plugin-you-should-know`
  is off by default — a background agent that points out what you and Claude may have missed; turn it on with
  `/plugin enable cc-plugin-you-should-know@builtin` (it works only on the Anthropic API with telemetry on, and where your
  organization has it available). Your own mods are turned off by `claude --safe-mode` for one session (along with all
  customizations) or for good by the `disableAllHooks` key; built-in mods are not turned off by either. To write your own, ask Claude
  (the `/plugin-authoring` skill).
- Removed: `/ultraplan` (use `/plan`), `/pr-comments` (since v2.1.91 — ask Claude to show the PR comments), `/vim` (since v2.1.92 — `/config` → Editor mode).
- Renamed: `/extra-usage` → `/usage-credits`. `/fork` now opens the copy as a background session, and the old "forked subagent" behavior is now `/subtask`. `/review` became an alias of `/code-review`.
- Whether a given command is available depends on the platform, plan and environment; the full current list is always in `/help` (type `/` to see the menu).

> **About `.claudeignore`:** there is no such file in the Claude Code documentation. You can keep files away from Claude with
> `permissions.deny` rules (for example `Read(./.env)`). Content search (Grep) skips files listed in `.gitignore`, Glob does not
> (change that with the variable `CLAUDE_CODE_GLOB_NO_IGNORE=false`), and the `@` file picker is controlled by the `respectGitignore` key.

---

# Part 2. `settings.json` keys

`settings.json` keys with types and examples, from the All settings reference (Claude Code 2.1.287), including enterprise and
deprecated ones.

Notation: **[managed]** — takes effect only from the managed (policy) source set by an admin;
**[user]** — read only from personal sources (the user file, `--settings`, managed) and ignored in the shared project file
(for `skipDangerousModePermissionPrompt`, `useAutoModeDuringPlan`, `syncClaudeAiSkills` and `syncClaudeAiPlugins`, local is also allowed);
**[exp]** — experimental; **[depr]** — deprecated. No marker — an ordinary key that can go in any file.

## Where the file lives and what overrides what

The same key can be set at several levels; the highest level where it is set wins. Priority (each one overrides the previous):
**user → project → local → CLI flag → managed**. Arrays (for example `permissions.allow`) are not overridden but merged across files
(`fallbackModel`, `modelPicker`, `availableModels` and `modelSettings` have their own rules).

| Level            | File                                                                                                                                                                                                                                                                                                        | Purpose                                                                                                                                      |
| ---------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| User             | `~/.claude/settings.json`                                                                                                                                                                                                                                                                                   | Personal settings for all projects                                                                                                           |
| Project          | `<project>/.claude/settings.json`                                                                                                                                                                                                                                                                           | Team settings, committed to the repository                                                                                                   |
| Local            | `<project>/.claude/settings.local.json`                                                                                                                                                                                                                                                                     | Personal settings for one specific project. Claude Code excludes the file from git itself; if you created it by hand, add it to `.gitignore` |
| CLI              | `--settings <file or JSON>`                                                                                                                                                                                                                                                                                 | For one run only                                                                                                                             |
| Managed (policy) | `managed-settings.json`: macOS `/Library/Application Support/ClaudeCode/`, Linux and WSL `/etc/claude-code/`, Windows `C:\Program Files\ClaudeCode\`; or a macOS MDM profile, the Windows registry `HKLM\SOFTWARE\Policies\ClaudeCode` (the `Settings` value), or server-managed from the claude.ai console | Enterprise policy; overrides everything (except a few restrictive keys where the stricter value wins)                                        |

Settings files are strict JSON: `//` comments and trailing commas count as a syntax error (Claude Code shows a Settings Error at startup).
To see which files loaded, use `/status` (the `Setting sources` line); to see what Claude Code rejected, use `claude doctor`. Claude Code itself also writes `~/.claude.json`
(login, MCP servers, project state and a few keys — see the section "`~/.claude.json` keys").

## How to read the examples: what goes where

Scalars (`model`, `theme`, numbers, flags) go at the top level of the object. Grouped settings
(`permissions`, `env`, `hooks`, `statusLine`, `worktree`, `voice`, `sandbox`, `sshConfigs`) are **nested objects**.
Here is a "skeleton" of the file showing what goes where:

```json
{
  "$schema": "https://json.schemastore.org/claude-code-settings.json",
  "model": "opus",
  "outputStyle": "Explanatory",
  "theme": "dark",
  "language": "russian",
  "autoCompactEnabled": true,
  "cleanupPeriodDays": 30,
  "env": {
    "CLAUDE_CODE_USE_POWERSHELL_TOOL": "1",
    "DISABLE_TELEMETRY": "1"
  },
  "permissions": {
    "allow": ["Bash(npm run build)", "Edit(src/**)"],
    "ask": ["Bash(git push *)"],
    "deny": ["Read(./.env)"],
    "defaultMode": "acceptEdits",
    "additionalDirectories": ["../shared-lib"]
  },
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [{ "type": "command", "command": "npm run lint" }]
      }
    ]
  },
  "statusLine": {
    "type": "command",
    "command": "~/.claude/statusline.sh",
    "padding": 1
  },
  "worktree": {
    "symlinkDirectories": ["node_modules"],
    "baseRef": "fresh"
  },
  "enabledPlugins": {
    "formatter@anthropic-tools": true
  }
}
```

Next come all the keys, grouped by topic. Each table has the **name**, the **type/allowed values** and a **description**.
For object settings, a JSON example follows the table where the shape is not obvious.

## Authentication and providers

| Key                          | Type / values                                     | What it does                                                                                                                                                                  |
| ---------------------------- | ------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `apiKeyHelper`               | string (command)                                  | A command that prints the credential for API requests (sent in both `X-Api-Key` and `Authorization: Bearer`). For rotating keys, for example from a secrets store.            |
| `awsCredentialExport`        | string (command)                                  | A command that prints AWS credentials as JSON (for Bedrock); only the Bedrock client sees these credentials, while Claude's shell commands get the regular ones.              |
| `awsAuthRefresh`             | string (command)                                  | A command (for example `aws sso login`) that refreshes the credentials in `.aws` when the current ones for Bedrock stop working.                                              |
| `gcpAuthRefresh`             | string (command)                                  | A command that refreshes Google Cloud Application Default Credentials when they have expired.                                                                                 |
| `otelHeadersHelper`          | string (path or command)                          | Prints JSON with HTTP headers for OpenTelemetry export — for tokens that rotate.                                                                                              |
| `forceLoginMethod`           | `"claudeai"` \| `"console"` \| `"gateway"`        | Restrict the login method: a claude.ai account, Claude Console or a cloud gateway (`gateway` is honored only from a managed source on the machine).                           |
| `forceLoginOrgUUID`          | string \| array of strings                        | Login is allowed only to the specified organization (one, or any from the list). Only a managed source enforces this; from other files it merely preselects the organization. |
| `forceLoginGatewayUrl`       | string (URL)                                      | **[managed]** The cloud gateway URL that the `/login` screen connects to.                                                                                                     |
| `gatewayInternalNetworks`    | array of strings (up to 4 IPv4 CIDRs, `/8`–`/32`) | **[managed]** Public IPv4 blocks that the organization uses to number its internal network — so that `/login` accepts a gateway within them.                                  |
| `allowedProviders`           | array of strings                                  | **[managed]** Which providers can be used: `anthropic`, `bedrock`, `vertex`, `foundry`, `anthropicAws`, `mantle`, `customEndpoint`, `gateway`. All others are rejected.       |
| `policyHelper`               | object `{path, timeoutMs?, refreshIntervalMs?}`   | **[managed]** An executable that computes managed settings at startup (`timeoutMs` default 10000; `refreshIntervalMs`: `0` or at least 60000).                                |
| `forceRemoteSettingsRefresh` | boolean                                           | **[managed]** Block startup until fresh server-managed settings are fetched; if the fetch fails, exit.                                                                        |
| `managedSourcesBehavior`     | `"first-wins"` \| `"merge"`                       | **[managed]** How to combine several managed sources: use only the highest-priority one (default) or merge them all.                                                          |
| `parentSettingsBehavior`     | `"first-wins"` \| `"merge"`                       | **[managed]** How to combine the settings of the host container (SDK, IDE) with the admin's: `first-wins` (default) — discard them; `merge` — apply them beneath the admin's. |

## Model, thinking, effort

| Key                       | Type / values                                      | What it does                                                                                                                                                                                               |
| ------------------------- | -------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `model`                   | string                                             | The model for new sessions: an alias (`opus`, `sonnet`…) or a full ID (`claude-opus-5-5`). Overridden by `ANTHROPIC_MODEL` and `--model`.                                                                  |
| `availableModels`         | array of strings                                   | Which models can be selected (alias, ID). An ID such as `claude-opus-5` also allows later versions (`claude-opus-5-5`). In managed settings it applies to the whole organization.                          |
| `availableModelsMatch`    | `"prefix"` \| `"exact"`                            | **[managed]** How to read `availableModels` entries: `prefix` (default) — the version and the IDs that extend it; `exact` — only the exact version named.                                                  |
| `deniedModels`            | array of strings                                   | **[managed]** A model denylist: overrides `availableModels`; a family alias (`opus`) blocks the whole family.                                                                                              |
| `enforceAvailableModels`  | boolean                                            | If "Default" falls outside `availableModels`, use the first available model from the list (default `false`).                                                                                               |
| `fallbackModel`           | array of strings                                   | Fallback models, in order, for when the primary model is overloaded or unavailable (the first three are used; `"default"` expands to the primary model). Not merged across files.                          |
| `modelOverrides`          | object (string → string)                           | A mapping from Anthropic IDs to provider IDs (for example a Bedrock inference profile ARN).                                                                                                                |
| `modelPicker`             | object `{options, replaceBuiltInOptions?}`         | **[user]** Your own rows in the `/model` list, with your own labels — after the built-in ones or (with `replaceBuiltInOptions: true`) instead of them. A row: `{model, label?, description?, behavesAs?}`. |
| `modelPricing`            | object `{multiplier?, overrides?}`                 | **[managed]** Show costs at the organization's contract rates rather than list prices.                                                                                                                     |
| `advisorModel`            | string                                             | The model for the server-side advisor tool (`fable`, `opus`, `sonnet` or a full ID); if unset, the advisor is off. Usually written by the `/advisor` command.                                              |
| `agent`                   | string                                             | The name of an agent (built-in or your own) for the main thread: applies its system prompt, tool restrictions and model.                                                                                   |
| `alwaysThinkingEnabled`   | boolean                                            | `false` turns off extended thinking. It is on by default; on models that always think (Opus 5.5, Sonnet 5.5, Fable), `false` has no effect.                                                                |
| `showThinkingSummaries`   | boolean                                            | Show thinking summaries when expanded (`Ctrl+O`); otherwise a collapsed placeholder.                                                                                                                       |
| `effortLevel`             | `"low"`\|`"medium"`\|`"high"`\|`"xhigh"`           | The default effort level for models without a saved level. `/effort` now writes the level to `modelSettings`, not here.                                                                                    |
| `maxEffortLevel`          | `"low"`\|`"medium"`\|`"high"`\|`"xhigh"`\|`"max"`  | A ceiling on the effort level: anything higher runs at the ceiling. Across several files the lowest one applies; `"max"` means no ceiling.                                                                 |
| `modelSettings`           | object (model → `{effortLevel?, maxEffortLevel?}`) | Effort level per model; `/effort` and the slider in `/model` write here. Takes precedence over the top-level `effortLevel`.                                                                                |
| `ultracode`               | boolean                                            | Start sessions with ultracode: Claude plans a workflow for every substantial task on its own (requires Workflows to be enabled and a model with `xhigh`). Claude Code only reads this key.                 |
| `fastMode`                | boolean                                            | `true` turns fast mode on (only Opus 5.5, Opus 5, Opus 4.8). Usually written by `/fast`.                                                                                                                   |
| `fastModePerSessionOptIn` | boolean                                            | `true` — the saved `fastMode` does not turn on at startup; `/fast` is needed in each session.                                                                                                              |
| `autoCompactEnabled`      | boolean                                            | Automatically compact the conversation when the context fills up (default `true`).                                                                                                                         |
| `autoCompactWindow`       | number (100000–1000000)                            | How full the context (in tokens) must be before auto-compact starts. Overridden by `--autocompact` and `CLAUDE_CODE_AUTO_COMPACT_WINDOW`.                                                                  |
| `switchModelsOnFlag`      | boolean                                            | What to do when the safety classifier flags a request: `true` (default) — switch to the fallback model and continue, `false` — pause and offer a choice.                                                   |
| `promptCacheTtl`          | `"5m"` \| `"1h"`                                   | The lifetime of the prompt cache for the main conversation; one hour keeps the cache longer, but writing to it costs more.                                                                                 |
| `subagentPromptCacheTtl`  | `"5m"` \| `"1h"`                                   | The same for subagents, workflows and Claude Code's background requests (compaction, session titles).                                                                                                      |

```json
{
  "fallbackModel": ["claude-sonnet-5", "claude-haiku-4-5"],
  "modelSettings": { "claude-opus-5-5": { "effortLevel": "high" } }
}
```

## Tool permissions (`permissions`)

The nested `permissions` object:

| Field                                 | Type / values                                                                                                   | What it does                                                                                                                                                                                                                                                                                                                                                 |
| ------------------------------------- | --------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `allow`                               | array of rule strings                                                                                           | What is allowed without asking. Format `Tool(specifier)`: `Bash(npm run build)`, `Edit(docs/**)`, `Read(~/.zshrc)`. `*` is a wildcard (`Bash(git *)`; the `:*` suffix is equivalent to ` *`). In an MCP rule, `*` works only in the tool name: `mcp__github__get_*`.                                                                                         |
| `ask`                                 | array of rule strings                                                                                           | What always requires confirmation — even in `acceptEdits` and `bypassPermissions` (in `dontAsk`, such calls are denied).                                                                                                                                                                                                                                     |
| `deny`                                | array of rule strings                                                                                           | What is always forbidden. For secrets: matching files drop out of search, and reading and editing are blocked.                                                                                                                                                                                                                                               |
| `defaultMode`                         | `"default"`\|`"acceptEdits"`\|`"plan"`\|`"auto"`\|`"dontAsk"`\|`"bypassPermissions"` (`"manual"` = `"default"`) | The mode for new sessions: read-only without asking / plus file edits and simple filesystem commands / analysis only / auto (a background classifier checks actions) / automatic denial instead of any prompt / everything without asking. `auto` and `bypassPermissions` do not take effect from project and local — set them in `~/.claude/settings.json`. |
| `additionalDirectories`               | array of paths                                                                                                  | Extra directories within the access scope (like `--add-dir`); their `.claude/` configuration is mostly not picked up.                                                                                                                                                                                                                                        |
| `blockReadsOutsideWorkingDirectories` | boolean                                                                                                         | File tools do not read outside the working directories in any mode, including bypass. `true` takes effect from any file.                                                                                                                                                                                                                                     |
| `disableBypassPermissionsMode`        | `"disable"`                                                                                                     | Forbid the mode that bypasses confirmations: the `--dangerously-skip-permissions` flag is rejected.                                                                                                                                                                                                                                                          |
| `disableAutoMode`                     | `"disable"`                                                                                                     | Remove auto mode from the `Shift+Tab` cycle (the same can be set with the top-level `disableAutoMode` key).                                                                                                                                                                                                                                                  |

```json
{
  "permissions": {
    "allow": ["Bash(npm run *)", "Edit(src/**)", "Read(~/.config/**)"],
    "ask": ["Bash(git push *)"],
    "deny": ["Read(./secrets/**)", "Read(./.env)"],
    "defaultMode": "acceptEdits",
    "additionalDirectories": ["../shared", "/tmp/workspace"]
  }
}
```

Related top-level keys:

| Key                                 | Type / values                                                         | What it does                                                                                                                                                                                                     |
| ----------------------------------- | --------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `skipDangerousModePermissionPrompt` | boolean                                                               | **[user]** Skip the dialog before bypass mode; Claude Code writes `true` itself once you accept the dialog.                                                                                                      |
| `skipAutoPermissionPrompt`          | boolean                                                               | **[user]** Skip the one-time explanation about auto mode the first time you enter it yourself.                                                                                                                   |
| `autoMode`                          | object `{environment, allow, soft_deny, hard_deny, classifyAllShell}` | **[user]** Your own rules for the auto mode classifier (arrays of text rules; `"$defaults"` inside an array keeps the built-in ones). `classifyAllShell: true` — run every shell command through the classifier. |
| `useAutoModeDuringPlan`             | boolean                                                               | **[user]** In plan mode, the auto mode classifier checks shell commands itself instead of prompting (default `true`); `false` — prompt.                                                                          |
| `allowManagedPermissionRulesOnly`   | boolean                                                               | **[managed]** Honor allow / ask / deny rules only from managed settings.                                                                                                                                         |

## MCP servers

| Key                                 | Type / values                                                 | What it does                                                                                                                                                                       |
| ----------------------------------- | ------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `enableAllProjectMcpServers`        | boolean                                                       | Auto-approve all MCP servers from the project's `.mcp.json` (ignored from the shared project file in a folder whose trust has not been accepted).                                  |
| `enabledMcpjsonServers`             | array of strings                                              | Servers from `.mcp.json` that are explicitly approved (by name). The approval dialog writes here (to `settings.local.json`).                                                       |
| `disabledMcpjsonServers`            | array of strings                                              | Servers from `.mcp.json` that are explicitly rejected: Claude Code will not connect them and will not ask.                                                                         |
| `allowedMcpServers`                 | array of objects `{serverName \| serverCommand \| serverUrl}` | An allowlist of MCP servers that may be added (entries from all files are merged). An empty array = no server is allowed. If set in managed settings, it is enforced for everyone. |
| `deniedMcpServers`                  | array of objects `{serverName \| serverCommand \| serverUrl}` | A denylist of MCP servers (including those from plugins and claude.ai connectors); entries from all files are merged.                                                              |
| `allowManagedMcpServersOnly`        | boolean                                                       | **[managed]** The MCP allowlist is read only from managed settings.                                                                                                                |
| `managedMcpServers`                 | object (name → `http`/`sse` server config)                    | **[managed]** Remote MCP servers that the admin provides to all users; they cannot be changed or removed.                                                                          |
| `allowAllClaudeAiMcps`              | boolean                                                       | **[managed]** Load claude.ai connectors alongside `managed-mcp.json` (otherwise it takes over control of MCP).                                                                     |
| `allowClaudeInChromeWithManagedMcp` | boolean                                                       | **[managed]** Allow the built-in Claude in Chrome server alongside `managed-mcp.json`.                                                                                             |
| `disableClaudeAiConnectors`         | boolean                                                       | Do not load claude.ai connectors; `true` takes effect from any file (the shared project file can turn them off for the repository).                                                |

> The MCP servers themselves (their commands/URLs) are defined in `.mcp.json` or via `claude mcp add`, not in `settings.json`.
> Only the access rules for them live here.

## Hooks (`hooks`)

`hooks` is an object "event → array of `{matcher, hooks}` groups". A group says which tools (or other values) the `matcher` reacts to
and which handlers to run (`hooks`). All matching handlers run in parallel. Handler type (`type`): `command` (a shell command;
the event JSON arrives on stdin), `http` (POST to a URL), `mcp_tool` (a call to an MCP server tool), `prompt` (a one-off evaluation by a model),
`agent` (a subagent checks a condition; experimental). A command can be in shell form (`command` is a string for the shell)
or exec form (`command` is an executable, `args` is an array of arguments, no shell). Common fields: `type`, `if` (a single permission rule,
for example `"Bash(git *)"` — tool events only), `timeout` (seconds), `statusMessage`; commands also have `args`, `async`, `asyncRewake`, `shell`.

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [{ "type": "command", "command": "./scripts/guard.sh" }]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          { "type": "command", "command": "npm run lint" },
          { "type": "command", "command": "node", "args": ["./scripts/notify.js"] }
        ]
      }
    ]
  }
}
```

`matcher`: `"*"`, `""` or omitted — all; only letters, digits, `_`, `-`, spaces, `,` and `|` — an exact name or a list (`Bash`, `Edit|Write`);
any other characters — a JavaScript regular expression, not anchored at the ends (`^Notebook`). For `FileChanged`, `matcher` holds the file names to watch.

Events: **session** — `SessionStart`, `Setup`, `SessionEnd`; **turn** — `UserPromptSubmit`, `UserPromptExpansion`, `Stop`, `StopFailure`, `MessageDisplay`;
**tools** — `PreToolUse`, `PermissionRequest`, `PermissionDenied`, `PostToolUse`, `PostToolUseFailure`, `PostToolBatch`;
**agents and tasks** — `SubagentStart`, `SubagentStop`, `TaskCreated`, `TaskCompleted`, `TeammateIdle`;
**environment and context** — `InstructionsLoaded`, `ConfigChange`, `CwdChanged`, `DirectoryAdded`, `FileChanged`, `PreCompact`, `PostCompact`, `WorktreeCreate`, `WorktreeRemove`;
**model and MCP** — `PreModelSwitch`, `PostModelSwitch`, `Elicitation`, `ElicitationResult`; **other** — `Notification`.
Hooks can also be set in the frontmatter of skills and agents and in plugins (`hooks/hooks.json`); to view them, use `/hooks`.

| Key                      | Type / values           | What it does                                                                                                                                                                                                                                                                                                |
| ------------------------ | ----------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `hooks`                  | object (event → groups) | Commands that run before/after tools and on session events. Hooks from all files are merged; other files cannot remove managed hooks.                                                                                                                                                                       |
| `disableAllHooks`        | boolean                 | Turn off hooks, the custom status line, the custom file-suggestion command and mods. From ordinary files, only your own are turned off (managed hooks and the organization's mods keep running); from managed settings, all of them; built-in mods always run. While hooks are off, `/goal` does not start. |
| `allowManagedHooksOnly`  | boolean                 | **[managed]** Run only hooks from managed settings (plus hooks from plugins force-enabled there).                                                                                                                                                                                                           |
| `allowedHttpHookUrls`    | array of strings        | An allowlist of URLs for HTTP hooks (`*` is a wildcard); an empty array blocks all HTTP hooks. Arrays from files are merged.                                                                                                                                                                                |
| `httpHookAllowedEnvVars` | array of strings        | An outer limit: which environment variables an HTTP hook may interpolate into headers (permission in the hook itself is also needed).                                                                                                                                                                       |

## Git: commits, PRs, attribution

| Key                      | Type / values                                   | What it does                                                                                                                                                                                                                       |
| ------------------------ | ----------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `attribution`            | object `{commit?, pr?, sessionUrl?}` \| `false` | Attribution in commits (`commit` — the `Co-Authored-By` git trailer), in PR descriptions (`pr`) and a link to the claude.ai session (`sessionUrl`, boolean). An empty string hides that part, `false` (v2.1.281+) hides all of it. |
| `includeCoAuthoredBy`    | boolean                                         | **[depr]** Replaced by `attribution` since v2.0.62. `false` hides attribution until `attribution.commit` / `attribution.pr` are set.                                                                                               |
| `includeGitInstructions` | boolean                                         | Include the built-in commit/PR instructions and a `git status` snapshot in the context (default `true`); `CLAUDE_CODE_DISABLE_GIT_INSTRUCTIONS` takes precedence.                                                                  |
| `prUrlTemplate`          | string                                          | A template for the PR link in the footer and summaries. Placeholders: `{host}` `{owner}` `{repo}` `{number}` `{url}`.                                                                                                              |

```json
{
  "attribution": { "commit": "", "pr": "" },
  "includeGitInstructions": true
}
```

## Interface and TUI

| Key                              | Type / values                                                                                                              | What it does                                                                                                                                                                                    |
| -------------------------------- | -------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `theme`                          | `"auto"`\|`"dark"`\|`"light"`\|`"dark-daltonized"`\|`"light-daltonized"`\|`"dark-ansi"`\|`"light-ansi"`\|`"custom:<name>"` | The interface color theme (default `"dark"`; `auto` adapts to the terminal background).                                                                                                         |
| `editorMode`                     | `"normal"` \| `"vim"`                                                                                                      | The key mode in the input box.                                                                                                                                                                  |
| `tui`                            | `"default"` \| `"fullscreen"`                                                                                              | The interface renderer; `fullscreen` — flicker-free (= `CLAUDE_CODE_NO_FLICKER=1`). Written by `/tui`.                                                                                          |
| `viewMode`                       | `"default"`\|`"verbose"`\|`"focus"`                                                                                        | The transcript view at startup; overrides `verbose` and the `/focus` choice. `focus` requires fullscreen.                                                                                       |
| `verbose`                        | boolean                                                                                                                    | Show full tool output instead of shortened summaries.                                                                                                                                           |
| `autoScrollEnabled`              | boolean                                                                                                                    | Follow new output downward (fullscreen; default `true`).                                                                                                                                        |
| `syntaxHighlightingDisabled`     | boolean                                                                                                                    | Turn off syntax highlighting in diffs, code blocks and file previews.                                                                                                                           |
| `prefersReducedMotion`           | boolean                                                                                                                    | Reduce or remove animations (accessibility).                                                                                                                                                    |
| `axScreenReader`                 | boolean                                                                                                                    | Output for screen readers: flat text with no borders or animations (the classic renderer; `tui` has no effect then).                                                                            |
| `showTurnDuration`               | boolean                                                                                                                    | Show "Cooked for …" after every turn (default `true`).                                                                                                                                          |
| `timeFormat`                     | `"auto"`\|`"12-hour"`\|`"24-hour"`\|`"24-hour-utc"`\| strftime template                                                    | The time format in the interface (any value with `%` is a template, for example `"%H:%M"`).                                                                                                     |
| `timeZone`                       | string (IANA, for example `"Europe/Sofia"`)                                                                                | The time zone for displayed times; ignored for `24-hour-utc`. There is no row for it in `/config` — set it in the file.                                                                         |
| `terminalProgressBarEnabled`     | boolean                                                                                                                    | Send the progress of long operations to the terminal (an indicator on the tab or in the taskbar; default `true`).                                                                               |
| `terminalTitleFromRename`        | boolean                                                                                                                    | `/rename` and `--name` change the terminal tab title (default `true`).                                                                                                                          |
| `spinnerTipsEnabled`             | boolean                                                                                                                    | Show tips in the waiting spinner (default `true`).                                                                                                                                              |
| `spinnerVerbs`                   | object `{mode: "append"\|"replace", verbs: string[]}`                                                                      | Your own spinner verbs (added to the standard ones or replacing them).                                                                                                                          |
| `spinnerTipsOverride`            | object `{tips, tipsFile?, label?, excludeDefault?}`                                                                        | Your own spinner tips: a string or `{id, text, cooldownSessions?, priority?}`; `excludeDefault: true` — show only these. Objects, `tipsFile` and `label` — only from user/managed/`--settings`. |
| `companyAnnouncements`           | array of strings                                                                                                           | Announcements shown at startup (one is picked at random if there are several).                                                                                                                  |
| `emojiCompletionEnabled`         | boolean                                                                                                                    | Emoji suggestions as you type a `:code:`, and replacement of a finished code with the emoji (default `true`).                                                                                   |
| `maxProseWidth`                  | number (columns, minimum 40)                                                                                               | Limit the width of prose in Claude's replies in a wide terminal (tables and code stay full width).                                                                                              |
| `respondToBashCommands`          | boolean                                                                                                                    | `false` — after a command run with `!`, the output goes into the context but Claude does not reply right away (default `true`).                                                                 |
| `bashEditDiffEnabled`            | boolean                                                                                                                    | **[user]** Record which files a Bash command changed: show a diff in the terminal and pass the list to PostToolUse hooks.                                                                       |
| `footerLinksRegexes`             | array of objects `{type: "regex", pattern, url, label?}`                                                                   | **[user]** Turn task and review IDs in the output (matched by regex) into clickable badges under the input box; `{name}` in `url` is taken from named groups.                                   |
| `vimInsertModeRemaps`            | object (two-character sequence → `"<Esc>"`)                                                                                | **[user]** Vim mode: a two-key sequence in INSERT (for example `"jj"`) works as Escape.                                                                                                         |
| `spellcheck`                     | object `{enabled, checker?, language?, color?}`                                                                            | **[user]** Underline typos in the input box using an installed spell checker (`aspell`, `hunspell`, `ispell` or `auto`).                                                                        |
| `wheelScrollAccelerationEnabled` | boolean                                                                                                                    | Accelerate mouse wheel scrolling when you scroll fast in fullscreen (default `true`).                                                                                                           |
| `askUserQuestionTimeout`         | `"60s"`\|`"5m"`\|`"10m"`\|`"never"`                                                                                        | **[user]** How long Claude waits on an unanswered question before continuing by itself with the option already selected (default `"never"`).                                                    |
| `autoContinueAtUsageLimit`       | boolean                                                                                                                    | **[user]** After hitting the claude.ai limit, wait in the open session and continue the task after the reset (default `true`).                                                                  |
| `dialogExpiry`                   | `"60s"`\|`"5m"`\|`"10m"`\|`"never"`                                                                                        | **[user]** How long to wait for an answer to dialogs forwarded to a remote client (Remote Control, SDK) and for consent to a cross-session message (default `"5m"`).                            |
| `promptSuggestionEnabled`        | boolean                                                                                                                    | Show gray suggestions for the next prompt (default `true`).                                                                                                                                     |
| `keybindingFlavor`               | `"classic"` \| `"readline"`                                                                                                | **[depr]** Has no effect since v2.1.261: words in the prompt are always edited with readline rules.                                                                                             |

## Status line

| Key                  | Type / values                               | What it does                                                                                                 |
| -------------------- | ------------------------------------------- | ------------------------------------------------------------------------------------------------------------ |
| `statusLine`         | object (see below)                          | A custom bottom status line, rendered by an external script.                                                 |
| `subagentStatusLine` | object `{type: "command", command: string}` | A status line for each subagent in the agent panel (the line's context reaches the script as JSON on stdin). |

`statusLine` fields: `type` (`"command"`), `command` (a string — the script), `padding` (a number), `refreshInterval` (a number ≥1 — recompute every N seconds), `hideVimModeIndicator` (boolean — hide the built-in `-- INSERT --` if the script draws the vim mode itself).

```json
{
  "statusLine": {
    "type": "command",
    "command": "~/.claude/statusline.sh",
    "padding": 1,
    "refreshInterval": 5
  }
}
```

## Context, memory, sessions

| Key                            | Type / values                               | What it does                                                                                                                                          |
| ------------------------------ | ------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| `cleanupPeriodDays`            | number ≥1                                   | How many days to keep transcripts before auto-deletion (default 30); a session older than that drops out of the `/resume` list.                       |
| `autoMemoryEnabled`            | boolean                                     | Auto memory (reading and writing personal notes; default `true`). `/memory` toggles it too.                                                           |
| `autoMemoryDirectory`          | string (path)                               | The auto memory folder (an absolute path or `~/…`). From project/local files it is honored only once the folder is trusted.                           |
| `fileCheckpointingEnabled`     | boolean                                     | Snapshot files before edits so that `/rewind` can restore them (default `true`).                                                                      |
| `outputStyle`                  | string                                      | The style of Claude's replies: `default`, `Proactive`, `Concise`, `Explanatory`, `Learning` or your own. Case matters — otherwise Default is applied. |
| `language`                     | string                                      | The language for replies and voice dictation (for example `"russian"`, `"japanese"`); the value is passed to Claude as is, without validation.        |
| `awaySummaryEnabled`           | boolean                                     | A one-line session recap when you come back after a few minutes away (on by default).                                                                 |
| `showClearContextOnPlanAccept` | boolean                                     | In the plan-approval dialog, add "clear context and execute the plan" as the first option (default `false`).                                          |
| `plansDirectory`               | string (path)                               | The folder for plan files, relative to the project root (otherwise `~/.claude/plans`).                                                                |
| `claudeMd`                     | string                                      | **[managed]** `CLAUDE.md`-style instructions, mixed in as organizational memory ahead of user and project.                                            |
| `claudeMdExcludes`             | array of strings (glob/absolute paths)      | Which `CLAUDE.md` files not to load (handy in a monorepo, to skip other teams' files).                                                                |
| `respectGitignore`             | boolean                                     | The `@` file picker does not show files listed in `.gitignore` (default `true`).                                                                      |
| `fileSuggestion`               | object `{type: "command", command: string}` | A custom source of file suggestions for `@` mentions (for example from a prebuilt index).                                                             |
| `bashOutputMaxChars`           | number (4000–128000)                        | How many characters of Bash/PowerShell output Claude gets at once (default 30000); the rest is saved to a file with a preview.                        |

Which instruction files are read is chosen by the **Project instructions** row in `/config` (since v2.1.277); in the settings file it is
`pluginConfigs["agents-md@builtin"].options.instructionFiles` (user, `--settings` or managed only). Values:
`claude-md-or-agents-md` (default: `CLAUDE.md`, and `AGENTS.md` only if neither the working folder nor any folder above it has a
`CLAUDE.md` or `CLAUDE.local.md`), `claude-md-and-agents-md` (both: `CLAUDE.md` first, then `AGENTS.md`), `claude-md` (`CLAUDE.md` only),
`managed-only` (only the managed `CLAUDE.md` and auto memory).

## Skills

| Key                          | Type / values                                                           | What it does                                                                                                                                                                            |
| ---------------------------- | ----------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `skillOverrides`             | object (name → `"on"`\|`"name-only"`\|`"user-invocable-only"`\|`"off"`) | Per-skill visibility: `name-only` — the model sees the name without the description; `user-invocable-only` — hidden from the model, but `/name` still works; `off` — hidden completely. |
| `skillListingMaxDescChars`   | number                                                                  | The character limit per skill description in the listing shown to Claude (default 1536).                                                                                                |
| `skillListingBudgetFraction` | number (0–1)                                                            | The share of context for the whole skill listing (default 0.01 = 1%).                                                                                                                   |
| `disableSkillShellExecution` | boolean                                                                 | Forbid inline shell in skills and custom commands (the command is replaced with a placeholder). `true` in managed settings cannot be overridden.                                        |
| `disableBundledSkills`       | boolean                                                                 | Remove the bundled skills and workflows; built-in commands such as `/init` stay available but are hidden from the model.                                                                |
| `syncClaudeAiSkills`         | boolean                                                                 | **[user]** `false` — do not download skills from your claude.ai account and do not load the ones already downloaded (`true` changes nothing).                                           |

## Plugins and marketplaces

| Key                             | Type / values                                                 | What it does                                                                                                                                                                                                                                                                                                                                                                             |
| ------------------------------- | ------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `enabledPlugins`                | object (`"plugin@marketplace"` → boolean)                     | Which plugins are enabled. Example: `{ "formatter@anthropic-tools": true }`. Written by `/plugin` and `claude plugin enable`.                                                                                                                                                                                                                                                            |
| `pluginConfigs`                 | object (by plugin ID)                                         | **[user]** Non-secret answers from plugin configuration dialogs (`options`, `mcpServers`); secret ones go to the system credential store. Built-in plugins carry the `@builtin` suffix: `agents-md@builtin` (see "Context, memory, sessions"), and in managed settings `cc-plugin-sec-default@builtin` with the option `allowManagedModsOnly: true` (load only the organization's mods). |
| `prependPlugins`                | array of plugin IDs                                           | **[user]** Which organization plugins' mods to run before all of the user's mods, in order; in managed settings, also list `sec-default@builtin` to keep the built-in guard. Read from the user file only when there are no managed settings and not on Team/Enterprise; never read from `--settings`.                                                                                   |
| `appendPlugins`                 | array of plugin IDs                                           | **[user]** The same, but these mods run after all of the user's mods; an ID that appears in both lists counts as prepend.                                                                                                                                                                                                                                                                |
| `extraKnownMarketplaces`        | object (name → `{source, autoUpdate?}`)                       | Extra marketplaces for a repository or organization; from the project file it is honored after the folder is trusted.                                                                                                                                                                                                                                                                    |
| `strictKnownMarketplaces`       | array of sources                                              | **[managed]** Plugins can be added and installed only from these marketplaces. An empty array blocks all.                                                                                                                                                                                                                                                                                |
| `blockedMarketplaces`           | array of sources                                              | **[managed]** Blocked marketplace sources (checked before download).                                                                                                                                                                                                                                                                                                                     |
| `pluginSuggestionMarketplaces`  | array of strings                                              | **[managed]** Whose plugins may appear in contextual install suggestions.                                                                                                                                                                                                                                                                                                                |
| `strictPluginOnlyCustomization` | boolean \| array (`"skills"`\|`"agents"`\|`"hooks"`\|`"mcp"`) | **[managed]** Skills, agents, hooks and MCP servers only from plugins and managed settings (`true` — all four).                                                                                                                                                                                                                                                                          |
| `pluginTrustMessage`            | string                                                        | **[managed]** Extra text for the trust warning shown before a plugin is installed.                                                                                                                                                                                                                                                                                                       |
| `disableCommandPluginSources`   | boolean                                                       | **[managed]** Forbid plugins with a `command` source (installed by running a command from the marketplace).                                                                                                                                                                                                                                                                              |
| `syncClaudeAiPlugins`           | boolean                                                       | **[user]** `false` — do not download plugins from your claude.ai account and do not load the ones already downloaded.                                                                                                                                                                                                                                                                    |

## Workflows, agents, auto mode

| Key                             | Type / values                                      | What it does                                                                                                                                                                                                      |
| ------------------------------- | -------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `enableWorkflows`               | boolean                                            | Turn Workflows on/off for yourself (on by default, except on the Pro plan). The **Dynamic workflows** toggle in `/config`.                                                                                        |
| `disableWorkflows`              | boolean                                            | Turn off Workflows and the built-in workflow commands for everyone the settings reach (= `CLAUDE_CODE_DISABLE_WORKFLOWS`).                                                                                        |
| `workflowKeywordTriggerEnabled` | boolean                                            | The word "ultracode" in a prompt launches a Workflow for that turn (default `true`).                                                                                                                              |
| `workflowSizeGuideline`         | `"unrestricted"`\|`"small"`\|`"medium"`\|`"large"` | A guideline for the number of agents in a workflow: `small` — fewer than 5, `medium` — fewer than 10, `large` — fewer than 50 (advice to the model, not a hard limit; default `medium`, `small` on the Pro plan). |
| `disableAgentView`              | boolean                                            | Turn off agent view and background agents (`claude agents`, `--bg`, `/background`).                                                                                                                               |
| `teammateMode`                  | `"in-process"`\|`"auto"`\|`"tmux"`\|`"iterm2"`     | Where to show the teammates of agent teams: in the main pane (default), auto (split panes in tmux/iTerm2), tmux, or native iTerm2 panes.                                                                          |
| `crossSessionInbound`           | `"accept"`\|`"hold"`\|`"refuse"`                   | What to do with messages from your other sessions: deliver them to Claude, show a notification without delivering, or drop them. From project/local, only the stricter value is honored.                          |
| `isolatePeerMachines`           | boolean                                            | Require your confirmation before `SendMessage` reaches a session on another machine. `true` takes effect from any file.                                                                                           |
| `processWrapper`                | string                                             | **[user]** macOS and Linux: an enterprise launcher placed in front of Claude Code's background processes (it must exec into Claude Code).                                                                         |

## Worktree (for the `--worktree` flag)

The nested `worktree` object:

| Field                | Type / values            | What it does                                                                                                                                                   |
| -------------------- | ------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `symlinkDirectories` | array of strings         | What to symlink from the main repository into the worktree so as not to bloat the disk (e.g. `node_modules`, `.cache`). By default nothing is symlinked.       |
| `sparsePaths`        | array of strings         | Which paths to include in the worktree via `git sparse-checkout`. Speeds things up a lot in large monorepos.                                                   |
| `baseRef`            | `"fresh"` \| `"head"`    | What new worktrees branch from: `fresh` (default) — from `origin/<default-branch>`; `head` — from the current local HEAD (with unpushed commits).              |
| `bgIsolation`        | `"worktree"` \| `"none"` | Isolation of background sessions: `worktree` (default) blocks edits in the main tree until `EnterWorktree`; `none` — allows editing the working copy directly. |

```json
{
  "worktree": {
    "symlinkDirectories": ["node_modules", ".cache"],
    "sparsePaths": ["packages/app", "packages/shared"],
    "baseRef": "fresh",
    "bgIsolation": "worktree"
  }
}
```

## Remote control, SSH, environments

| Key                                | Type / values                                                      | What it does                                                                                                                                                                                                                                   |
| ---------------------------------- | ------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `disableRemoteControl`             | boolean                                                            | Turn off Remote Control (`claude remote-control`, `--remote-control`, auto-start and the in-session toggle).                                                                                                                                   |
| `remoteControlAtStartup`           | boolean                                                            | Connect Remote Control automatically at the start of every interactive session (a toggle in `/config`).                                                                                                                                        |
| `remote.defaultEnvironmentId`      | string (`env_…` or `ccpool_…`)                                     | The default cloud environment for `claude --cloud`; written by `/remote-env`. Self-hosted `ccpool_…` — only from user/managed.                                                                                                                 |
| `sshConfigs`                       | array of objects (see below)                                       | **[user]** Preconfigured SSH connections in the Desktop environment dropdown (read by the Desktop app).                                                                                                                                        |
| `sshHostAllowlist`                 | array of strings (host patterns)                                   | **[managed]** Which hosts Desktop SSH sessions are allowed to connect to; an empty array turns SSH sessions off.                                                                                                                               |
| `disableDesktopLocalSessions`      | boolean                                                            | **[managed]** Turn off on-device Code sessions in Desktop (leave only remote SSH and cloud ones).                                                                                                                                              |
| `browserExternalPageTools`         | `"disabled"`                                                       | **[managed]** Claude's tools do not work with external pages in the Browser pane of the Desktop app.                                                                                                                                           |
| `disableBrowserExternalNavigation` | boolean                                                            | **[managed]** The Browser pane in Desktop opens only localhost — for both people and Claude.                                                                                                                                                   |
| `disableMobileSimulatorTools`      | boolean                                                            | **[managed]** Block Claude's access to the iOS Simulator pane in Desktop.                                                                                                                                                                      |
| `desktopSessionCleanupPeriodDays`  | number ≥0                                                          | **[user]** The maximum age of Desktop/Cowork session transcripts in days (`0` — no limit); they are deleted once they are older than both this value and `cleanupPeriodDays`.                                                                  |
| `disableDeepLinkRegistration`      | `"disable"`                                                        | Do not register the `claude-cli://` link handler with the OS (such links open a session with a ready-made prompt).                                                                                                                             |
| `enableArtifact`                   | boolean                                                            | `false` turns off the Artifact tool (it publishes session output as a private web page on claude.ai) for every session that reads the file; no file can turn it back on, `true` changes nothing. Written by the Artifacts toggle in `/config`. |
| `disableArtifact`                  | boolean                                                            | **[depr]** Replaced by `enableArtifact`: `true` is equivalent to `enableArtifact: false`, `false` is ignored.                                                                                                                                  |
| `channelsEnabled`                  | boolean                                                            | **[managed]** Allow channel notifications (MCP servers with the `claude/channel` capability); on Team and Enterprise they are blocked without this.                                                                                            |
| `allowedChannelPlugins`            | array of `{marketplace, plugin}` or strings `"plugin@marketplace"` | **[managed]** An allowlist of channel plugins (replaces Anthropic's default list).                                                                                                                                                             |

Fields of a `sshConfigs` item: `id`, `name`, `sshHost` (`"user@host"` or an alias from `~/.ssh/config`) are required; `sshPort`, `sshIdentityFile` are optional.

```json
{
  "remote": { "defaultEnvironmentId": "env_123" },
  "sshConfigs": [
    {
      "id": "dev-vm",
      "name": "Dev VM",
      "sshHost": "deploy@dev.example.com",
      "sshPort": 22
    }
  ]
}
```

## Notifications and voice

| Key                       | Type / values                                                                                                     | What it does                                                                                                                                                   |
| ------------------------- | ----------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `preferredNotifChannel`   | `"auto"`\|`"terminal_bell"`\|`"iterm2"`\|`"iterm2_with_bell"`\|`"kitty"`\|`"ghostty"`\|`"notifications_disabled"` | How to notify about a finished task and a pending permission: `auto` — a desktop notification in iTerm2, Ghostty, Kitty, otherwise silence (default).          |
| `inputNeededNotifEnabled` | boolean                                                                                                           | A push notification to your phone when a permission confirmation or an answer to a question is waiting (only with Remote Control connected; default `false`).  |
| `agentPushNotifEnabled`   | boolean                                                                                                           | Let Claude send proactive push notifications to your phone when it sees fit (default `false`).                                                                 |
| `voice`                   | object `{enabled?, mode?: "hold"\|"tap", autoSubmit?}`                                                            | Voice input: `hold` — hold to talk, `tap` — press to start, press again to stop; `autoSubmit` — submit the prompt on release (hold only). Written by `/voice`. |
| `voiceEnabled`            | boolean                                                                                                           | **[depr]** The old one-line way to turn dictation on (since v2.1.92 it is replaced by `voice.enabled`, which takes precedence).                                |

```json
{
  "voice": { "enabled": true, "mode": "hold", "autoSubmit": true }
}
```

## Sandbox (`sandbox`) and network

The nested `sandbox` object. It isolates Claude's Bash commands from the file system and the network (macOS, Linux, WSL2).

| Field                                                     | Type / values                                    | What it does                                                                                                                                                    |
| --------------------------------------------------------- | ------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `enabled`                                                 | boolean                                          | Turn on the sandbox for Bash (default `false`). Choosing a mode in `/sandbox` writes the key to `settings.local.json`.                                          |
| `failIfUnavailable`                                       | boolean                                          | If the sandbox is enabled but fails to start, exit Claude Code with an error instead of running commands without it (default `false`).                          |
| `autoAllowBashIfSandboxed`                                | boolean                                          | Run Bash in the sandbox without prompting (default `true`); `deny` rules and specific `ask` rules (`Bash(git push *)`) still apply.                             |
| `excludedCommands`                                        | array of strings                                 | Commands that run outside the sandbox (`Bash(...)` rule syntax: an exact command, a prefix such as `docker *`, a wildcard).                                     |
| `allowUnsandboxedCommands`                                | boolean                                          | Whether a blocked command may be retried outside the sandbox (the `dangerouslyDisableSandbox` parameter); `false` — strict mode (default `true`).               |
| `ignoreViolations`                                        | object (command substring → array of substrings) | Do not show violation reports for paths that the command is known to try and be denied (the denial itself stays).                                               |
| `enableWeakerNestedSandbox`                               | boolean                                          | Linux: a sandbox inside unprivileged Docker (it mounts the existing `/proc`) — weakens protection.                                                              |
| `enableWeakerNetworkIsolation`                            | boolean                                          | macOS: access to the TLS trust service (`com.apple.trustd.agent`) — needed by `gh`, `gcloud`, `terraform` behind a MITM proxy; weakens protection.              |
| `allowAppleEvents`                                        | boolean                                          | **[user]** macOS: let the sandbox send Apple Events (`open`, `osascript`) — removes code-execution isolation.                                                   |
| `ripgrep`                                                 | object `{command, args?}`                        | **[user]** Your own ripgrep binary for the sandbox.                                                                                                             |
| `bwrapPath` / `socatPath`                                 | string (absolute path)                           | **[managed]** Paths to bubblewrap and socat outside `PATH`.                                                                                                     |
| `filesystem.allowWrite` / `denyWrite`                     | array of paths                                   | Extra paths where writing is allowed / not allowed (by default you can write to the working folder, the temp folder and added directories).                     |
| `filesystem.denyRead` / `allowRead`                       | array of paths                                   | Paths closed to reading / reopened inside a `denyRead` area. By default almost everything can be read, including files with credentials.                        |
| `filesystem.allowManagedReadPathsOnly`                    | boolean                                          | **[managed]** Honor `allowRead` only from managed settings.                                                                                                     |
| `filesystem.disabled`                                     | boolean                                          | **[user]** Turn off file-system isolation, keeping network isolation.                                                                                           |
| `network.allowedDomains` / `deniedDomains`                | array of strings                                 | Domains that are pre-approved (no prompt) / always blocked; `*.example.com` and `:port` are allowed. `deniedDomains` takes precedence.                          |
| `network.strictAllowlist`                                 | boolean                                          | **[user]** Deny hosts outside the list instead of prompting.                                                                                                    |
| `network.allowManagedDomainsOnly`                         | boolean                                          | **[managed]** Honor only domains from managed settings; all others are blocked without asking.                                                                  |
| `network.allowUnixSockets` / `allowAllUnixSockets`        | array of strings / boolean                       | Allowed Unix sockets: paths — macOS only; on Linux and WSL2 only `allowAllUnixSockets` is available.                                                            |
| `network.allowLocalBinding`                               | boolean                                          | macOS: allow binding to localhost ports (for example a dev server).                                                                                             |
| `network.allowMachLookup`                                 | array of strings                                 | macOS: extra XPC/Mach services (iOS Simulator, Playwright); a trailing `*` means a prefix.                                                                      |
| `network.httpProxyPort` / `socksProxyPort`                | number (port)                                    | Your own HTTP / SOCKS5 proxy instead of the built-in one.                                                                                                       |
| `network.tlsTerminate`                                    | object `{caCertPath?, caKeyPath?}`               | **[exp]** **[user]** The sandbox proxy terminates TLS (needed for `mask` credential substitution); `{}` — an ephemeral CA per session.                          |
| `credentials.files` / `credentials.envVars`               | arrays of objects `{path \| name, mode}`         | Credential protection: `"deny"` — hide, `"mask"` — show a placeholder while the proxy substitutes the real value. `mask` — only from user/managed/`--settings`. |
| `credentials.allowPlaintextInject` / `awsPairs` / `sigv4` | boolean / array / object                         | **[user]** Fine-tuning `mask` for AWS: substitution over plain HTTP too, grouping of non-standard AWS variables, SigV4 forms (`deny` or `passthrough`).         |

Related top-level key: `skipWebFetchPreflight` (boolean) — skip the WebFetch domain check via `api.anthropic.com` in locked-down enterprise environments.

```json
{
  "sandbox": {
    "enabled": true,
    "autoAllowBashIfSandboxed": true,
    "network": {
      "allowedDomains": ["api.example.com", "*.githubusercontent.com"],
      "deniedDomains": ["telemetry.example.com"]
    },
    "filesystem": {
      "denyRead": ["~/.ssh", "~/.aws"]
    }
  }
}
```

## Updates and miscellaneous

| Key                          | Type / values                  | What it does                                                                                                                                                                          |
| ---------------------------- | ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `autoUpdatesChannel`         | `"latest"` \| `"stable"`       | The auto-update channel: `latest` (default) — the newest release; `stable` — a version about a week old that skips releases with serious regressions.                                 |
| `minimumVersion`             | string                         | The minimum version — prevents rolling back below it when switching to `stable`. In managed settings, a lower bound for the organization.                                             |
| `requiredMinimumVersion`     | string                         | **[managed]** The oldest version that is allowed to run; on an older one Claude Code exits at startup.                                                                                |
| `requiredMaximumVersion`     | string                         | **[managed]** The newest version that is allowed to run.                                                                                                                              |
| `defaultShell`               | `"bash"` \| `"powershell"`     | The shell for `!` commands in the input box (default `bash`; on Windows without Bash — `powershell`).                                                                                 |
| `feedbackSurveyRate`         | number (0–1)                   | The probability of showing the session quality survey; `0` — never show it.                                                                                                           |
| `feedbackDrafts`             | `"notify"`\|`"quiet"`\|`"off"` | **[user]** Draft feedback that Claude prepares on its own: show a card / quietly accumulate / turn the tool off.                                                                      |
| `disableSideloadFlags`       | boolean                        | **[managed]** Reject the flags `--plugin-dir`, `--plugin-url`, `--agents`, `--mcp-config`, which are used to get around `strictKnownMarketplaces`.                                    |
| `wslInheritsWindowsSettings` | boolean                        | **[managed, Windows]** WSL reads managed settings from the Windows policy chain.                                                                                                      |
| `$schema`                    | string (URL)                   | A link to the settings JSON schema — autocomplete and validation in the editor. The schema can be older than the CLI: a warning about a recent key does not mean the config is wrong. |

Removed, and they do nothing (you may meet them in an old config; deprecated but still readable keys are marked **[depr]** in their own sections): `taskOutputMaxChars` (v2.1.277), `permissionExplainerEnabled` (v2.1.257), `teammateDefaultModel` (v2.1.234).

## `~/.claude.json` keys (not `settings.json`)

Claude Code stores these keys in `~/.claude.json` (`/config` writes them); they do not go into the settings files.

| Key                            | Type / values            | What it does                                                                                                                              |
| ------------------------------ | ------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------- |
| `autoConnectIde`               | boolean                  | Auto-connect to a running IDE when starting from an external terminal (default `false`); `CLAUDE_CODE_AUTO_CONNECT_IDE` takes precedence. |
| `autoInstallIdeExtension`      | boolean                  | Auto-install the IDE extension when launched from the VS Code terminal (default `true`).                                                  |
| `claudeInChromeDefaultEnabled` | boolean                  | Start every interactive session with the Chrome integration, without `--chrome`.                                                          |
| `copyFullResponse`             | boolean                  | `/copy` always copies the whole reply, without the code block picker.                                                                     |
| `copyOnSelect`                 | boolean                  | Copy mouse selections to the clipboard automatically (fullscreen and agent view; default `true`).                                         |
| `defaultToAgentsView`          | boolean                  | `claude` with no arguments opens agent view instead of a new conversation.                                                                |
| `diffTool`                     | `"auto"` \| `"terminal"` | Where to show an edit's diff when an IDE is connected: in the IDE (`auto`, default) or in the terminal.                                   |
| `externalEditorContext`        | boolean                  | On `Ctrl+G`, show Claude's previous reply as `#` comments in the external editor's buffer.                                                |
| `leftArrowOpensAgents`         | boolean                  | `←` on an empty prompt sends the session to the background and opens agent view (default `true`).                                         |
| `prStatusFooterEnabled`        | boolean                  | A PR/MR status badge for the current branch in the prompt footer (default `true`).                                                        |

## Environment variables (`env`)

`env` is a nested object "name → value". **All values are strings** (put numbers and booleans in quotes:
`"PORT": "3000"`, a flag — `"1"`). These variables are set in the environment of Claude Code sessions and usually override variables of the same name from the shell.
Almost any variable is accepted; the most useful ones are listed below. Project and local settings **cannot** set some variables
(`CLAUDE_CONFIG_DIR`, `CLAUDE_CODE_TMPDIR`, the OpenTelemetry variables that enable export and choose its address) — set them in the shell,
in user settings or in managed settings.

```json
{
  "env": {
    "CLAUDE_CODE_USE_POWERSHELL_TOOL": "1",
    "ANTHROPIC_MODEL": "claude-opus-5-5",
    "BASH_DEFAULT_TIMEOUT_MS": "120000",
    "DISABLE_TELEMETRY": "1"
  }
}
```

**Providers and authentication**

| Variable                                                                         | Value type  | What it does                                                                                                                                         |
| -------------------------------------------------------------------------------- | ----------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| `ANTHROPIC_API_KEY` / `ANTHROPIC_AUTH_TOKEN`                                     | string      | A key (the `X-Api-Key` header) / a bearer token (`Authorization`). A key overrides the subscription; to return to it, run `unset ANTHROPIC_API_KEY`. |
| `CLAUDE_CODE_OAUTH_TOKEN`                                                        | string      | A claude.ai OAuth token (login without `/login`, for the SDK and automation); get one with `claude setup-token`.                                     |
| `ANTHROPIC_BASE_URL`                                                             | URL         | A custom base URL for the API (proxy/gateway). With a non-Anthropic host, Remote Control is turned off, and MCP tool search is too, by default.      |
| `ANTHROPIC_CUSTOM_HEADERS`                                                       | string      | Extra HTTP headers for the API (`Name: Value`, several separated by newlines).                                                                       |
| `CLAUDE_CODE_USE_BEDROCK` / `CLAUDE_CODE_USE_VERTEX` / `CLAUDE_CODE_USE_FOUNDRY` | `"1"`/`"0"` | Use Amazon Bedrock / Google Cloud Agent Platform (Vertex) / Microsoft Foundry as the provider.                                                       |
| `AWS_BEARER_TOKEN_BEDROCK` / `ANTHROPIC_VERTEX_PROJECT_ID`                       | string      | A Bedrock API key / the GCP project for Vertex.                                                                                                      |
| `CLAUDE_CODE_SKIP_BEDROCK_AUTH` / `CLAUDE_CODE_SKIP_VERTEX_AUTH`                 | `"1"`/`"0"` | Skip provider authentication (when a gateway does it).                                                                                               |
| `HTTPS_PROXY` / `HTTP_PROXY` / `NO_PROXY`                                        | URL / list  | A proxy for outgoing traffic / domains and IPs that bypass the proxy.                                                                                |

**Models and limits**

| Variable                                                                                            | Value type           | What it does                                                                                                                                                                                                                                                            |
| --------------------------------------------------------------------------------------------------- | -------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `ANTHROPIC_MODEL`                                                                                   | string               | The primary model; takes precedence over the `model` key, but `--model` and `/model` take precedence over it.                                                                                                                                                           |
| `ANTHROPIC_DEFAULT_OPUS_MODEL` / `ANTHROPIC_DEFAULT_SONNET_MODEL` / `ANTHROPIC_DEFAULT_HAIKU_MODEL` | string               | Which IDs stand behind the `opus` / `sonnet` / `haiku` aliases (haiku is also used for background tasks).                                                                                                                                                               |
| `ANTHROPIC_SMALL_FAST_MODEL`                                                                        | string               | **[depr]** The small fast model for background tasks; use `ANTHROPIC_DEFAULT_HAIKU_MODEL` instead.                                                                                                                                                                      |
| `CLAUDE_CODE_SUBAGENT_MODEL`                                                                        | string               | The default model for subagents, teammates and workflow agents (the `model` in an agent definition or at call time takes precedence over it).                                                                                                                           |
| `CLAUDE_CODE_MAX_OUTPUT_TOKENS` / `MAX_MCP_OUTPUT_TOKENS`                                           | number (as a string) | The output token limit (the ceiling depends on the model) / the limit on an MCP tool's response (default 25000).                                                                                                                                                        |
| `MAX_THINKING_TOKENS`                                                                               | number (as a string) | A fixed "thinking" budget; `0` — turn it off on the Anthropic API (except for models that always think). On adaptive models a non-zero value is ignored.                                                                                                                |
| `CLAUDE_CODE_EFFORT_LEVEL`                                                                          | string               | Effort level: `low`, `medium`, `high`, `xhigh`, `max` or `auto`; takes precedence over `--effort`, `/effort` and the `effortLevel` / `modelSettings` keys.                                                                                                              |
| `CLAUDE_CODE_AUTO_COMPACT_WINDOW`                                                                   | number               | The auto-compact window in tokens, 100000–1000000, as a whole number only (it does not understand `500k`); takes precedence over `autoCompactWindow` and `--autocompact`.                                                                                               |
| `CLAUDE_CODE_DISABLE_1M_CONTEXT`                                                                    | `"1"`                | No 1M window: the 1M variants disappear from `/model`, and models with native 1M (Sonnet 5.5, Fable) stay at 200K. Since v2.1.287, Opus 4.7+ and Fable on Bedrock, Vertex, Foundry and the Claude apps gateway run with 1M by default — this variable brings back 200K. |
| `DISABLE_AUTO_COMPACT`                                                                              | `"1"`                | Turn off auto-compact (manual `/compact` remains); takes precedence over `autoCompactEnabled`.                                                                                                                                                                          |
| `DISABLE_PROMPT_CACHING`                                                                            | `"1"`                | Turn off prompt caching for all models.                                                                                                                                                                                                                                 |
| `API_TIMEOUT_MS`                                                                                    | number (ms)          | The timeout for API requests (default 600000; maximum 2147483647).                                                                                                                                                                                                      |

**Behavior and environment**

| Variable                                          | Value type  | What it does                                                                                                                                                                                                             |
| ------------------------------------------------- | ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `CLAUDE_CODE_USE_POWERSHELL_TOOL`                 | `"1"`/`"0"` | The PowerShell tool. Windows: on by itself (on Bedrock, Vertex and Foundry, `1` is needed if Git Bash is present), `0` turns it off. Linux, macOS and WSL: `1` turns it on, `pwsh` is required.                          |
| `CLAUDE_CODE_DISABLE_ADAPTIVE_THINKING`           | `"1"`       | Turn off adaptive thinking and go back to the fixed budget from `MAX_THINKING_TOKENS`. Works only on Opus 4.6 and Sonnet 4.6.                                                                                            |
| `CLAUDE_CONFIG_DIR`                               | path        | The configuration directory instead of `~/.claude` (settings, history, plugins) — handy for several accounts. Ignored when set in project/local settings.                                                                |
| `CLAUDE_CODE_GIT_BASH_PATH`                       | path        | Windows: the path to Git Bash's `bash.exe` if it is not on `PATH`.                                                                                                                                                       |
| `CLAUDE_CODE_SHELL` / `CLAUDE_CODE_SHELL_PREFIX`  | string      | The shell for the Bash tool (a path to `bash` or `zsh`) / a wrapper prefix for Claude Code's shell commands (Bash, hooks, status line, stdio MCP).                                                                       |
| `CLAUDE_CODE_TMPDIR`                              | path        | The directory for temporary files. Ignored when set in project/local settings.                                                                                                                                           |
| `USE_BUILTIN_RIPGREP`                             | `"0"`       | `0` — use the system `rg` instead of the built-in one.                                                                                                                                                                   |
| `CLAUDE_CODE_NO_FLICKER`                          | `"1"`/`"0"` | The flicker-free fullscreen renderer (= `tui: "fullscreen"`; overrides `tui`).                                                                                                                                           |
| `CLAUDE_CODE_DISABLE_TERMINAL_TITLE`              | `"1"`       | Do not change the terminal title (and do not make the background request for a session title).                                                                                                                           |
| `CLAUDE_CODE_DISABLE_BACKGROUND_TASKS`            | `"1"`       | Turn off all background work: `run_in_background`, auto-backgrounding, `Ctrl+B`.                                                                                                                                         |
| `CLAUDE_CODE_DISABLE_WORKFLOWS`                   | `"1"`       | Turn off Workflows (= `disableWorkflows`).                                                                                                                                                                               |
| `CLAUDE_CODE_DISABLE_AGENT_VIEW`                  | `"1"`       | Turn off agent view and background agents (= `disableAgentView`).                                                                                                                                                        |
| `CLAUDE_CODE_DISABLE_CLAUDE_MDS`                  | `"1"`       | Do not load any `CLAUDE.md` memory files (user, project, auto memory).                                                                                                                                                   |
| `CLAUDE_CODE_DISABLE_AUTO_MEMORY`                 | `"1"`/`"0"` | `1` — turn off auto memory; `0` — force it on (even with `--bare` and `autoMemoryEnabled: false`).                                                                                                                       |
| `CLAUDE_CODE_NEW_INIT`                            | `"1"`       | `/init` runs an interactive flow (it asks which files to create).                                                                                                                                                        |
| `BASH_DEFAULT_TIMEOUT_MS` / `BASH_MAX_TIMEOUT_MS` | number (ms) | The default timeout for a foreground Bash/PowerShell command (120000) / the maximum the model can set (600000).                                                                                                          |
| `BASH_MAX_OUTPUT_LENGTH`                          | number      | How many characters of Bash output to read back (default 30000, maximum 150000); the `bashOutputMaxChars` key takes precedence.                                                                                          |
| `MCP_TIMEOUT` / `MCP_TOOL_TIMEOUT`                | number (ms) | The MCP server startup timeout (default 30000) / the MCP tool execution timeout.                                                                                                                                         |
| `CLAUDE_CODE_MAX_WEB_SEARCHES_PER_SESSION`        | number      | The limit on WebSearch calls per session (default 200); after that Claude continues with what it has already found. You can raise it as high as you like, but the limit cannot be turned off _(in this project: `500`)_. |
| `CLAUDE_CODE_MAX_SUBAGENTS_PER_SESSION`           | number      | **[depr]** Removed in v2.1.224 and does nothing (it used to limit the number of subagents per session).                                                                                                                  |

**Privacy, telemetry, updates**

| Variable                                                             | Value type          | What it does                                                                                                                                                                                                                                                     |
| -------------------------------------------------------------------- | ------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC`                           | any non-empty value | A single switch for all non-essential outgoing traffic (auto-updates, telemetry, error reporting, `/feedback`, release notes…); it also turns off feature-flag fetching. **`0` and `false` also turn it on** — to allow traffic, the variable has to be removed. |
| `DISABLE_TELEMETRY` / `DISABLE_ERROR_REPORTING`                      | any non-empty value | Opt out of telemetry / of error reporting (`DISABLE_TELEMETRY` also turns off feature-flag fetching). `0` and `false` also turn the opt-out on — you can only reverse it by removing the variable.                                                               |
| `CLAUDE_CODE_ENABLE_TELEMETRY`                                       | `"1"`               | Turn on OpenTelemetry collection (metrics and logs); without it, OTel exporters are not configured. Ignored when set in project/local settings.                                                                                                                  |
| `DISABLE_AUTOUPDATER` / `DISABLE_UPDATES`                            | `"1"`               | Turn off background auto-updates (a manual `claude update` still works) / block all updates, including `claude update` and `claude install`.                                                                                                                     |
| `DISABLE_COST_WARNINGS`                                              | `"1"`               | Hide cost warnings.                                                                                                                                                                                                                                              |
| `DISABLE_FEEDBACK_COMMAND`                                           | `"1"`               | Hide `/feedback` and Claude's draft feedback, as well as `/bug` and `/share` (the old name `DISABLE_BUG_COMMAND` is also accepted).                                                                                                                              |
| `OTEL_*` (`OTEL_METRICS_EXPORTER`, `OTEL_EXPORTER_OTLP_ENDPOINT`, …) | string              | Export of metrics, logs and traces through OpenTelemetry; some of the variables (enabling, address, content) are ignored when set in project/local settings.                                                                                                     |

> The full list of env variables is wider (Bedrock/Vertex/Foundry regions, OTEL details, internal flags) —
> only the practically useful ones are listed here. Any variable set in `env` will end up in the session environment.
