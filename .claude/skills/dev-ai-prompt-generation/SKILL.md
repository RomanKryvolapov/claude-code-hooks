---
name: dev-ai-prompt-generation
description: Techniques for writing and reviewing effective prompts, Agent Skills (SKILL.md), rules, and CLAUDE.md. Use for any prompting or LLM-instruction task — drafting or reviewing a prompt (system, user, tool, RAG), creating or editing a skill, a rule, or CLAUDE.md, or choosing structure, delimiters, output format, few-shot, reasoning-model handling, structured outputs, tool and agent prompting, context engineering, prompt caching, evals, injection defenses, or RAG grounding.
---

# Prompt and Agent Skill generation (spec)

Techniques for generating effective prompts and for writing Agent Skills and project rules (`CLAUDE.md`), based on research and practice.

When creating prompts or rules:

- Choose techniques that fit the task
- Use structure (delimiters, tags, headings)
- Control output format (schema, examples, self-check)
- Follow brevity and reformulation principles

Rules of thumb:

- Order: structure → data → rules → format
- Combine techniques for complex tasks; start with basics, add advanced as needed
- Prefer the **Modern model behavior** index for anything that ships

---

## Read this first (2026 model behavior)

**Pick the target by what you’re writing — the artifact decides the conventions, not a blanket default.**

- **A Claude Code artifact** — an Agent Skill (`SKILL.md`), a rule, `CLAUDE.md`, or any prompt that Claude itself will run — is written for **Claude**: use Anthropic’s conventions (XML tags where they help, a system prompt, the structure Anthropic recommends). These ship on Claude, so don’t write them in OpenAI style.
- **An application prompt for the API** follows the conventions of the vendor it will actually ship on. If no vendor is named, assume the **OpenAI API** — Markdown sections (not XML tags), application rules in the **developer** message, reasoning depth via `reasoning_effort` (+ `text.verbosity`), and machine-consumed output via strict `json_schema` (`text.format`).

For any other vendor, switch conventions per [references/vendor-portability-cross-model-conventions.md](references/vendor-portability-cross-model-conventions.md).

Three notes below override older folklore in this skill. They apply to every prompt, rule, and skill you write for a frontier model.

> **1. Reasoning models change the rules.** On frontier reasoning models (Claude Opus 4.x / Fable 5, OpenAI GPT-5 / o-series, Gemini 3, Grok reasoning, DeepSeek-R1): do **not** hand-write chain-of-thought; control depth with the vendor’s **effort / thinking knob**; prefer **zero-shot** goal + constraints + output-contract over few-shot and prescriptive steps; and **dial back** `CRITICAL` / `You MUST` force-language, which now over-triggers. → [references/reasoning-models-thinking-effort-and-cot.md](references/reasoning-models-thinking-effort-and-cot.md)

> **2. Vendor portability is not assumed.** Structuring, role placement, and sampling differ by vendor: XML tags are Claude’s primitive; OpenAI/Gemini use Markdown sections and the **developer** message; sampling advice is partly inverted (Gemini keep temp 1.0; DeepSeek-R1 0.6; Claude frontier removes temperature); DeepSeek-R1 wants no system prompt. Pick one convention per prompt. → [references/vendor-portability-cross-model-conventions.md](references/vendor-portability-cross-model-conventions.md)

> **3. Quantitative claims are not load-bearing.** Accuracy percentages from older prompt folklore (delimiters +24%, pseudocode +36% / −87%, MetaGlyph 62–81%, INoT +7.95%, self-check 60→97%, counting 5→95%) are unsubstantiated or context-dependent and have been removed or softened across this skill. **Measure correctness on your own evals.** → [references/evals-llm-as-judge-and-regression-gates.md](references/evals-llm-as-judge-and-regression-gates.md)

---

## Progressive disclosure (how skills load)

Skills load progressively — metadata, then body, then references on demand. Treat the token budgets below as rough design guidance (from Anthropic’s Agent Skills design), not hard limits; confirm specifics against Anthropic’s Agent Skills docs (code.claude.com/docs/skills).

| Level            | When it loads      | Rough budget                      | Contents                             |
| ---------------- | ------------------ | --------------------------------- | ------------------------------------ |
| 1 — Metadata     | always (discovery) | ~100 (hard cap 1,536 chars/skill) | YAML `description` (+ `when_to_use`) |
| 2 — Instructions | skill activated    | ~5,000                            | body of `SKILL.md`                   |
| 3 — Resources    | only when needed   | opened file(s)                    | `references/` linked from `SKILL.md` |

- Keep `SKILL.md` a hub (≤500 lines); put depth in `references/`.
- One hop only: `SKILL.md` → `references/foo.md`; no reference-to-reference primary chains.
- Index every reference with a link and a **Contents** cell specific enough to choose one file without opening it.
- **Agent rule:** open only the reference row that matches the task, not the whole set.

---

## How to split a large skill into `references/`

Use when `SKILL.md` would exceed ~500 lines, or a topic is large enough to load on its own.

### Size targets

| Layer                  | Target                                                                                                         |
| ---------------------- | -------------------------------------------------------------------------------------------------------------- |
| `SKILL.md`             | ≤ ~500 lines — hub only: purpose, progressive disclosure, index tables, short mandatory notes, checklist.      |
| Each `references/*.md` | ~150 lines (±). If a topic is still too long, split into two semantically-named files, not one oversized file. |

### Filenames (semantic)

Pattern `short-topic-kebab-slug.md`, English, kebab-case, stable. Good: `structured-outputs-constrained-decoding.md`. Bad: `part-01.md` — the agent cannot pick it without opening it. Order by topic, not by a `part-01`/`part-02` number.

### Safe cut boundaries

Split only where Markdown stays valid: after a complete `##` / `###` section or a closed fenced block. Never split inside a fence, a table row, or a normative rule. If one `##` exceeds ~150 lines, subdivide it by `###` into separate files.

### Reference header

Start every `references/*.md` with:

```markdown
## About this file

**What’s inside:** One or two sentences — topics/techniques covered.

**Hub:** [../SKILL.md](../SKILL.md)
```

### Hub index (required)

A table with a **Contents** cell (specific enough to choose one file) and a **File** link. Every reference appears exactly once — no orphans.

### After splitting

Keep a single source of truth (no duplicate monolith). Sync `CLAUDE.md` (and any other repo index you keep) when layout or discoverability changes.

---

## Agent Skills: official docs (mandatory) — summary

**Agent Skills** (`SKILL.md`) are Anthropic’s open standard. The contract is **Anthropic’s Agent Skills docs** (code.claude.com/docs/skills, platform.claude.com). Claude Code skills live under `.claude/skills/` in a project, or `~/.claude/skills/` for personal ones, and may also ship inside a plugin.

When creating, editing, or reviewing a skill:

1. Load the current official docs (code.claude.com/docs/skills) with web access, so behavior matches the shipping product.
2. Treat those docs as the contract for the path, frontmatter fields (all optional — `description` is the routing signal), how a skill is selected (by `description` relevance), and optional folders such as `references/`.
3. Resolve conflicts in favor of the official docs; update this skill if it drifts.
4. If the project has other skills that no longer match the documented structure, migrate them in the same change set.
5. After substantive skill work, sync `CLAUDE.md` (and any other index your repo requires).

Full wording, the skill-docs sections, and the **when to use which technique** table → [references/skills-docs-and-technique-table.md](references/skills-docs-and-technique-table.md).
Frontmatter fields & invocation control → [references/skill-frontmatter-and-invocation.md](references/skill-frontmatter-and-invocation.md). Dynamic context, arguments, `context: fork`, lifecycle, skill evals → [references/skill-dynamic-context-arguments-lifecycle.md](references/skill-dynamic-context-arguments-lifecycle.md).

---

## Reference index (read on demand)

Open only the row that matches the task.

### Skill docs & technique table

| Contents                                                                                                                                                                              | File                                                                                                               |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| Agent Skills docs workflow, prompts in code, how skills & project rules apply, `SKILL.md` layout, when to use which technique                                                         | [references/skills-docs-and-technique-table.md](references/skills-docs-and-technique-table.md)                     |
| **Skill frontmatter** — every field, invocation control (`disable-model-invocation`, `user-invocable`), tool pre-approval, `paths`, command naming, listing limits, where skills live | [references/skill-frontmatter-and-invocation.md](references/skill-frontmatter-and-invocation.md)                   |
| **Skill mechanics** — dynamic context `` !`cmd` ``, `$ARGUMENTS` / `${CLAUDE_*}` substitution, `context: fork` subagents, content lifecycle & compaction, skill evals                 | [references/skill-dynamic-context-arguments-lifecycle.md](references/skill-dynamic-context-arguments-lifecycle.md) |

### Core principles

| Contents                                                                                                                                             | File                                                                                                                                                     |
| ---------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Core principle, language, brevity, reformulation, structure, wording, examples, parameters, **size control**                                         | [references/core-principles-basics-language-brevity-through-size-control.md](references/core-principles-basics-language-brevity-through-size-control.md) |
| Single responsibility, organizing structure, delimiters/headings, markdown, **tables**                                                               | [references/core-principles-single-responsibility-organizing-and-tables.md](references/core-principles-single-responsibility-organizing-and-tables.md)   |
| Task block patterns, table-filter example, **token separation**, number format, CAPS/quotes, **force-language & positive phrasing (2026 dial-back)** | [references/core-principles-task-markdown-tokens-caps.md](references/core-principles-task-markdown-tokens-caps.md)                                       |
| **ASCII frames**, glossary, visual markers / icons                                                                                                   | [references/core-principles-ascii-frames-glossary-visual-icons.md](references/core-principles-ascii-frames-glossary-visual-icons.md)                     |

### Modern model behavior (read first for any production prompt)

The 2025–2026 essentials. For a prompt that ships, these usually matter more than the formatting tricks below.

| Contents                                                                                           | File                                                                                                                     |
| -------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| **Reasoning models** — thinking/effort knobs per vendor, when NOT to write CoT, deprecations       | [references/reasoning-models-thinking-effort-and-cot.md](references/reasoning-models-thinking-effort-and-cot.md)         |
| **Reasoning-technique taxonomy** — CoT, self-consistency, ToT, ReAct, Reflexion, step-back, CoVe…  | [references/reasoning-techniques-taxonomy.md](references/reasoning-techniques-taxonomy.md)                               |
| **Structured outputs** — strict JSON-schema / constrained decoding per vendor; prefill deprecated  | [references/structured-outputs-constrained-decoding.md](references/structured-outputs-constrained-decoding.md)           |
| **Agentic & tool-use** — tool descriptions, eagerness vs persistence, parallel calls, multi-agent  | [references/agentic-tool-use-prompting.md](references/agentic-tool-use-prompting.md)                                     |
| **Context engineering** — window as a resource, context rot, compaction vs editing vs memory       | [references/context-engineering-window-memory-compaction.md](references/context-engineering-window-memory-compaction.md) |
| **Prompt caching & cost** — prefix-match invariant, model routing, effort tiering                  | [references/prompt-caching-and-cost-levers.md](references/prompt-caching-and-cost-levers.md)                             |
| **Evals & LLM-as-judge** — regression gates, analytic rubrics, bias controls, auto-optimization    | [references/evals-llm-as-judge-and-regression-gates.md](references/evals-llm-as-judge-and-regression-gates.md)           |
| **Injection defenses** — instruction hierarchy, spotlighting, isolation, egress, untrusted content | [references/untrusted-content-and-injection-defenses.md](references/untrusted-content-and-injection-defenses.md)         |
| **RAG grounding** — grounding contract, citations, retrieval pipeline (hybrid + rerank)            | [references/rag-grounding-and-citations.md](references/rag-grounding-and-citations.md)                                   |
| **Vendor portability** — XML vs Markdown, system vs developer vs no-system-prompt, sampling, Llama | [references/vendor-portability-cross-model-conventions.md](references/vendor-portability-cross-model-conventions.md)     |

### Data presentation (XML, JSON, sources)

| Contents                                                                       | File                                                                                                                                                 |
| ------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| XML tags, templates, universal XML prompt                                      | [references/data-presentation-xml-tags-and-universal-template.md](references/data-presentation-xml-tags-and-universal-template.md)                   |
| XML mistakes, namespaces, **source marking**, attribution, counting / chunking | [references/data-presentation-xml-mistakes-namespaces-sources-counting.md](references/data-presentation-xml-mistakes-namespaces-sources-counting.md) |
| **JSON** input structure, benchmarks / anchors                                 | [references/data-presentation-json-input-and-anchors.md](references/data-presentation-json-input-and-anchors.md)                                     |

### Describing rules (pseudocode, MetaGlyph, YAML, decorators)

| Contents                                                                                         | File                                                                                                                                     |
| ------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------- |
| **Pseudocode** `IF`/`ELSE`, `SWITCH`, function-style rules                                       | [references/describing-rules-pseudocode-if-switch-function-style.md](references/describing-rules-pseudocode-if-switch-function-style.md) |
| **MetaGlyph** symbolic operators, example, ASCII alternatives **(experimental — not a default)** | [references/describing-rules-metaglyph-symbolic-operators.md](references/describing-rules-metaglyph-symbolic-operators.md)               |
| Rule **priority hierarchy**, **YAML** & **TOML** configs                                         | [references/describing-rules-hierarchy-yaml-toml.md](references/describing-rules-hierarchy-yaml-toml.md)                                 |
| **Prompt decorators** (`+++…`), stacking **(experimental — prefer native API knobs)**            | [references/describing-rules-prompt-decorators.md](references/describing-rules-prompt-decorators.md)                                     |

### Output format (schema, few-shot, self-check)

| Contents                                                                           | File                                                                                                                                     |
| ---------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| Placeholders, JSON schema, LangChain escaping, intermediate JSON                   | [references/output-format-placeholders-json-schema-langchain.md](references/output-format-placeholders-json-schema-langchain.md)         |
| TypeScript-style contract, comparison tables, ladder, **Schema / Examples / Task** | [references/output-format-typescript-schema-examples-task-ladder.md](references/output-format-typescript-schema-examples-task-ladder.md) |
| Code fences, **self-check**, verification-first, **INoT** _(INoT experimental)_    | [references/output-format-self-check-verification-inot.md](references/output-format-self-check-verification-inot.md)                     |
| **Few-shot** patterns, delimiters, ordering, contrasting pairs                     | [references/output-format-few-shot-examples-and-order.md](references/output-format-few-shot-examples-and-order.md)                       |

### Universal template & checklists

| Contents                                                                     | File                                                                                                                           |
| ---------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| Full universal template (role/context/task/rules/output/examples/self-check) | [references/universal-prompt-template-full-example.md](references/universal-prompt-template-full-example.md)                   |
| Quick templates + **quick gate** + **verification** checklist                | [references/universal-prompt-quick-templates-and-checklists.md](references/universal-prompt-quick-templates-and-checklists.md) |

---

## When to use this skill

- Drafting or reviewing **LLM prompts** (system, user, tool, RAG).
- Creating or editing **Agent Skills** (`SKILL.md`, Claude Code skills) or project rules (`CLAUDE.md`).
- Choosing **structure, delimiters, data format, few-shot, and self-check** patterns for a task.
- Working with **reasoning models, structured outputs, tool/agent prompts, context engineering, prompt caching, evals, injection defenses, or RAG** — see the **Modern model behavior** index.

---

## Checklist (this skill hub)

- [ ] **Progressive disclosure:** `SKILL.md` stays lean; depth lives under `references/` with explicit links from this file.
- [ ] **Splitting:** semantic filenames, ~150 lines per ref, safe cut boundaries, `## About this file`, hub table with a **Contents** column.
- [ ] **One hop:** primary navigation is `SKILL.md` → `references/*.md`; no deep reference-only chains.
- [ ] **Reference size & naming:** each `references/*.md` stays ~150 lines (±); filename + `## About this file` explain the contents without reading the whole file.
- [ ] **Official docs:** skill (`SKILL.md`) conventions confirmed against Anthropic’s Agent Skills docs (code.claude.com/docs/skills).
- [ ] **Writing a Claude Code skill?** invocation mode set deliberately (`disable-model-invocation` for side-effect workflows, `user-invocable: false` for background knowledge); data injected via dynamic context (`` !`cmd` ``) instead of asking the model to fetch it; body written as standing instructions (content persists all session; first ~5k tokens survive compaction).
- [ ] **Reasoning model?** removed manual chain-of-thought; set the vendor effort/thinking knob; preferred zero-shot; dialed back force-language.
- [ ] **Machine-consumed output?** used structured outputs / strict schema (constrained decoding), not prose JSON or assistant prefill.
- [ ] **Agent / tool prompt?** trigger conditions + side effects in tool descriptions; tool/retrieved/user content treated as untrusted data.
- [ ] **Changed a prompt or bumped a model?** re-ran evals / re-baselined tokens before shipping; **no unsourced accuracy percentages stated as fact**.
- [ ] **Repo sync:** after substantive edits, `CLAUDE.md` (and any other repo index) updated if your process requires it.
- [ ] **Full checklists** for prompts and rules: [references/universal-prompt-quick-templates-and-checklists.md](references/universal-prompt-quick-templates-and-checklists.md); full template body: [references/universal-prompt-template-full-example.md](references/universal-prompt-template-full-example.md).
