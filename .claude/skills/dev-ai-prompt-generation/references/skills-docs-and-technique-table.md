## About this file

**What’s inside:** Mandatory Agent Skills docs workflow, prompts-in-code notes, how Claude project rules and skills apply, `SKILL.md` layout, and the "When to use which technique" agent guide.

**Hub:** [../SKILL.md](../SKILL.md)

---

## Agent Skills documentation (mandatory)

**Agent Skills** (`SKILL.md`) are Anthropic’s **open standard**. The authoritative contract is **Anthropic’s Agent Skills docs** (code.claude.com/docs/skills, platform.claude.com agent-skills, agentskills.io). In Claude Code, skills live under `.claude/skills/` (project) or `~/.claude/skills/` (personal).

When **creating, editing, or reviewing** a skill:

1. **Load the current official docs** before relying on this skill or existing repo habits: the latest **Agent Skills** pages at **code.claude.com/docs/skills** (and platform.claude.com). Use **web access** (`WebFetch`, `WebSearch`, or browser MCP) so guidance matches the **shipping product**, not stale training data. If a page is empty or script-heavy, follow the docs index or search the site for the topic.

2. **Treat the official docs as the contract** for: the discovery path (`.claude/skills`, user-level `~/.claude/skills`), frontmatter fields, how a skill is selected (by `description` relevance), folder layout, and optional folders (`references/`, `assets/`, `scripts/`) **only as those docs currently define them**.

3. **Resolve conflicts** in favor of the official Agent Skills docs. If this skill contradicts them after verification, **update this file** in the same pass.

4. **Migrate outdated skills**: if repository files **no longer match** the documented structure, **rewrite, rename, or relocate** them in the same change set. Do not leave known drift “for later.”

5. **Index sync (if your project keeps one)**: if the project’s **`CLAUDE.md`** or another index lists skills or rules, update it after substantive skill work.

---

## Prompts in code

- Keep template and variables clear; escape `{` as `{{` in f-strings/LangChain.
- Use one canonical number/date format and state it in the prompt.

---

## How project rules and skills apply

Claude Code has two layers:

- **Always-on rules** live in `CLAUDE.md` (project root and nested dirs) and load into every session. Keep it to a few base rules — repo structure, house style — and move longer rules into separate files (a common convention is `.claude/rules/*.md`) that `CLAUDE.md` pulls in.
- **Skills** (`SKILL.md`) load on demand: the agent sees each skill's `name` + `description` first and pulls the full body only when the task matches. The `description` is the routing signal — make it **keyword-rich and task-oriented** (e.g. "Testing: unit, integration, e2e; pytest, fixtures, mocks" beats "Testing").

---

## Skills (SKILL.md)

Agent Skills are Anthropic’s **open standard** (used by Claude Code, Claude apps, and the API).

- **First** follow **Agent Skills documentation (mandatory)** above; the layout and optional folders are defined by the **Anthropic Agent Skills spec** (code.claude.com/docs/skills), not older bullets in this file.
- Base path: `.claude/skills/<skill-name>/` (user-level `~/.claude/skills/`) with **`SKILL.md`** as the entrypoint. Optional sibling folders (`references/`, `assets/`, `scripts/`) per the Agent Skills spec; link to them from `SKILL.md`. Discovery roots and frontmatter fields change over time — confirm against code.claude.com/docs/skills.
- **All frontmatter fields are optional** in Claude Code; `description` is the recommended routing signal — describe **what** + **when** + **keyword-rich** trigger phrases. The command name comes from the **directory name**; frontmatter `name` is a display label. For cross-product portability keep `name` = folder name. Full field table, invocation control (`disable-model-invocation`, `user-invocable`), tool pre-approval, listing limits → [skill-frontmatter-and-invocation.md](skill-frontmatter-and-invocation.md).
- Body: **When to use**, instructions, examples, edge cases. Keep the **primary workflow** in `SKILL.md`; move long appendices, data files, or static assets into **optional folders**, with clear relative links. Once invoked, the body **stays in context for the whole session** — write standing instructions, not one-shot steps. Dynamic context (`` !`cmd` ``), `$ARGUMENTS`, `context: fork`, lifecycle & compaction, skill evals → [skill-dynamic-context-arguments-lifecycle.md](skill-dynamic-context-arguments-lifecycle.md).
- **Custom commands are merged into skills**: `.claude/commands/*.md` is the same mechanism with fewer features; on a name clash the skill wins.

---

## When to use which technique (agent guide)

**Technique names match section headings** in the reference files — search for the exact heading.

> **No load-bearing percentages.** Older folklore attached specific accuracy numbers to these techniques (delimiters "+24%", pseudocode "+36% / −87% tokens", tables "+40%", self-check "60→97%", counting "5→95%"). They are unsubstantiated or context-dependent and have been removed. Effects are model- and task-dependent — **measure on your own evals** ([evals-llm-as-judge-and-regression-gates.md](evals-llm-as-judge-and-regression-gates.md)).

### Structure, data, rules, output

| Goal                                         | Technique (section heading)                                  | When to apply                                                                                                                                                                         |
| -------------------------------------------- | ------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Separate role, task, data, examples          | **Delimiters and headings**                                  | Every prompt with ≥2 logical blocks. Pick one delimiter convention, state it once, stay consistent — format is a hidden control parameter.                                            |
| Clear boundaries for context/task/rules      | **XML tags and semantics**                                   | 3+ logical parts. Anthropic recommends XML (not trains); Markdown sections are the cross-vendor equivalent (see vendor portability). Use semantic tags; meaningless tags are ignored. |
| Compare/filter objects by parameters         | **Tables for structured data**                               | Object–property data; comparison or multi-condition filtering. Row = object, column = property. Use lists for ordered steps.                                                          |
| Control output shape / enforce correctness   | **Placeholders**, **JSON schema with types**, **Self-check** | Strict-format or high-stakes output. For machine-consumed output prefer **structured outputs** (native schema enforcement), not prose JSON.                                           |
| Show output format and reasoning style       | **Few-shot examples**                                        | Non-obvious format; classification. **Non-reasoning models** — few-shot can degrade reasoning models (start zero-shot). Put the hardest example last.                                 |
| Conditional / multi-branch logic             | **Pseudocode for conditional logic**                         | Deterministic branch rules (IF/ELSE, SWITCH/CASE). Not to force a reasoning sequence on a thinking model.                                                                             |
| Order constraints by importance              | **Rule hierarchy**                                           | Several rules of differing importance. Critical (🔴) at start or end, not the middle (lost-in-the-middle, not recency).                                                               |
| Lists, records, counting, arithmetic         | **Token separation**                                         | Enumerations and records. Separate elements so they don’t merge into one token. For exact counts use a code-execution tool.                                                           |
| One non-negotiable prohibition               | **CAPS and emphasis**                                        | A single critical "never do X", in quotes, at start/end. Dial back `CRITICAL`/`MUST` force-language — it over-triggers frontier models.                                               |
| Multi-document Q&A; cite; reduce fabrication | **Source marking** + **RAG grounding**                       | RAG / multiple documents. Numbered blocks + "use ONLY the context; say so if absent; cite every claim". Also a security boundary.                                                     |
| Define terms / acronyms once                 | **Glossary**                                                 | Domain terms/acronyms used later. Underspecified terms cause regressions across model/prompt changes.                                                                                 |
| Force answer into fixed categories           | **Visual markers (icons)**                                   | Fixed-section analysis (SWOT, code review). Effect comes from categorization, not the icon.                                                                                           |
| Many related / nested attributes             | **JSON structuring of input data**                           | API/DB-like or nested input. Keeps related facts adjacent in context.                                                                                                                 |
| Maximum predictability                       | **Schema–Examples–Task**                                     | Strict classification/extraction on instruct/fast models. Drop worked examples on reasoning models.                                                                                   |
| Request code without preamble                | **Backticks for code**                                       | Code-only answers. (Assistant-prefill to force shape is discouraged where native structured outputs exist; verify prefill behavior against the current API docs.)                     |

### Modern model behavior (2026) — read these first for any production prompt

| Goal                                                            | Reference                                                                                          |
| --------------------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| Reasoning models, depth knob, when NOT to write CoT             | [reasoning-models-thinking-effort-and-cot.md](reasoning-models-thinking-effort-and-cot.md)         |
| Choose among CoT / ToT / ReAct / Reflexion / step-back / CoVe … | [reasoning-techniques-taxonomy.md](reasoning-techniques-taxonomy.md)                               |
| Guaranteed-valid machine output                                 | [structured-outputs-constrained-decoding.md](structured-outputs-constrained-decoding.md)           |
| Tool/agent prompts, eagerness vs persistence, multi-agent       | [agentic-tool-use-prompting.md](agentic-tool-use-prompting.md)                                     |
| Managing the context window, memory, compaction                 | [context-engineering-window-memory-compaction.md](context-engineering-window-memory-compaction.md) |
| Real token-cost lever (caching), model routing                  | [prompt-caching-and-cost-levers.md](prompt-caching-and-cost-levers.md)                             |
| Measuring prompts, LLM-as-judge, auto-optimization              | [evals-llm-as-judge-and-regression-gates.md](evals-llm-as-judge-and-regression-gates.md)           |
| Prompt-injection defenses, untrusted content                    | [untrusted-content-and-injection-defenses.md](untrusted-content-and-injection-defenses.md)         |
| RAG grounding contract + retrieval quality                      | [rag-grounding-and-citations.md](rag-grounding-and-citations.md)                                   |
| What ports across vendors and what doesn’t                      | [vendor-portability-cross-model-conventions.md](vendor-portability-cross-model-conventions.md)     |

### Experimental / not defaults

**MetaGlyph** (symbolic compression), **Prompt decorators** (`+++…`), **INoT**, **CNL-P**, **Cognitive BASIC** — niche, model-specific, or based on single preprints. Don’t reach for them by default; prefer native API knobs and plain conditional rules.

**Decision shortcut:**
Reasoning model → set the effort/thinking knob, state goal + constraints, no manual CoT. Machine-consumed output → structured outputs (schema), not prose JSON. Structured data → table or JSON. Conditional rules → pseudocode. Few-shot only on non-reasoning models. Agent/tool prompt → "when to call" in the tool, untrusted content as data. Production prompt → measure on evals; cache the stable prefix.

## Checklist

- [ ] Anthropic Agent Skills docs (code.claude.com/docs/skills) consulted before changing skill conventions.
- [ ] `description` keyword-rich for routing (key use case first — listing truncates); invocation mode chosen deliberately ([skill-frontmatter-and-invocation.md](skill-frontmatter-and-invocation.md)).
- [ ] Skill body written as standing instructions; data injected via dynamic context where a preprocessor can fetch it ([skill-dynamic-context-arguments-lifecycle.md](skill-dynamic-context-arguments-lifecycle.md)).
- [ ] Technique chosen via the guide; modern-model references checked for production prompts.
