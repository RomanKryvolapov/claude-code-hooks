## About this file

**What’s inside:** Minimal quick templates (task/context/rules/format), few-shot block, self-check snippet, key-value mini-config, the quick pre-ship gate, and the full verification-vs-practices checklist.

**Hub:** [../SKILL.md](../SKILL.md)

---

## Quick templates

**Minimal structure:**

```markdown
## Task

[What to do]

---

## Context / Data

[Background or input]

---

## Rules

- [rule 1]
- [rule 2]

---

## Format

[Expected output shape or example]
```

**Few-shot block:**

```markdown
## Examples (=== separates examples)

Input: "[example 1]"
→ Output: [result 1]

===

Input: "[example 2]"
→ Output: [result 2]

---

## Task

[Actual input to process]
```

**Self-check block:**

```markdown
<self_check>
Before output, check:
□ [requirement 1]?
□ [requirement 2]?
□ [requirement 3]?

If any item fails — fix BEFORE output.
</self_check>
```

**Key-value (short config):**

```
Role: [role]
Tone: friendly | formal | neutral
Length: ~500 chars
Format: JSON
```

---

## Quick gate (before shipping any prompt)

A coarse go/no-go. For the exhaustive per-practice audit, use the verification checklist below — it is the single source for the detailed items, so this gate stays a short subset.

- [ ] Sections and delimiters (`---`, `###`) between blocks
- [ ] Context/task/rules separated; critical rules at start or end
- [ ] No wall of text; key terms in **bold** or `backticks`
- [ ] Data in tags or JSON/table; sources numbered if multiple
- [ ] Rules in IF/ELSE or priority (🔴🟡🟢) when conditional
- [ ] Output format explicit (schema, example, or template)
- [ ] Self-check when quality is critical
- [ ] No duplicate or compensating instructions; one formulation per requirement
- [ ] **2026 essentials** — reasoning-model handling, structured outputs, agent/tool prompts, vendor fit, and eval-gating all considered (full items in the verification checklist below)
- [ ] **If skill (SKILL.md):** name matches folder; description = what + when + keywords; layout matches the **Anthropic Agent Skills spec** (a project skill lives at `.claude/skills/<skill-name>/SKILL.md`); `SKILL.md` links to any `references/` / `assets/` / `scripts/`

Good prompt: readable in one pass, primary vs secondary clear, solved by formulation, explicit format (schema + examples + self-check where needed).

---

## Verification checklist: practices vs prompt (for agent)

Use this **after** drafting a prompt, rule, or skill text. For each category: (1) does it apply to this prompt? (2) if yes, is the prompt aligned? (3) does anything in the prompt break it? If a practice applies and the prompt violates it — fix the prompt.

- [ ] **Core principle** — Task is minimal formulation; no pile of compensating instructions. Root cause fixed, not patched.
- [ ] **Language** — Prompt/rule text in English (except allowed examples/variables). Priority labels 🔴🟡🟢 used where priorities exist.
- [ ] **Brevity** — Short, dense; each sentence adds meaning; nothing redundant for unambiguity.
- [ ] **Reformulate vs extend** — No instruction added to patch bad formulation; no old + new conflicting wording; no duplicate formulations.
- [ ] **Structure** — Flat; one task → one block; goal/constraints/format separate; one sentence per line where it helps.
- [ ] **Wording** — Positive, direct, verifiable; outcome requirements, not vague "try to" / "preferably".
- [ ] **Examples** — Only when needed; no example that doesn’t change understanding; no corner-case overload; focus on the main scenario.
- [ ] **Parameters** — Variables/placeholders, no hardcoding; prompt stays generic when inputs change.
- [ ] **Size control** — No unnecessary growth; if long, task reformulated/simplified, not extended.
- [ ] **Delimiters** — Reliable delimiters (`---`, `===`, `###`); avoid `~~~~`, `____`, `....`, `////`; delimiter purpose stated where it matters.
- [ ] **Headings** — Clear hierarchy (# main, ## sections, ### subs); arrow `→` for transformations where useful.
- [ ] **Tables** — "Object — properties" in tables, not prose lists where comparison matters.
- [ ] **Token separation** — Enumerations separated (newline, comma, pipe) so elements don’t glue; numbers in one canonical form if specified.
- [ ] **CAPS** — At most one critical prohibition in CAPS; at start or end, not buried in the middle; exact forbidden values in quotes.
- [ ] **XML/semantic tags** — Where used: meaningful tags (`<context>`, `<task>`, `<rules>`); all closed; nesting ≤ 2–3 levels; 5+ tags consider namespace/prefix.
- [ ] **Source marking** — Multiple sources separated and attributed; context grounding stated when the model must not add external knowledge.
- [ ] **JSON** — One `{}` pair for structure; no `{{{{...}}}}`; LangChain/f-string escaping only where applicable.
- [ ] **Conditional logic** — IF/ELSE or SWITCH/CASE for deterministic branch rules; not used to force a reasoning sequence on a thinking model. (MetaGlyph is experimental — avoid by default.)
- [ ] **Rule hierarchy** — Critical → Important → Desirable; critical at start or end.
- [ ] **Output format** — Schema and/or template and/or examples; types/constraints explicit where strictness is needed; self-check when quality is critical.
- [ ] **Few-shot** — Count appropriate (0–2–3–5); examples separated (e.g. `===`); last example = main; contrasting pairs if boundaries matter.
- [ ] **Self-check** — If quality is critical: explicit checklist (□ …) before output.
- [ ] **No contradictions** — Full read-through: nothing contradicts the practices above; no forbidden patterns (meaningless tags, unclosed tags, vague wording, duplicate/compensating rules).
- [ ] **Technique fit** — For the prompt’s goal, the "When to use which technique" table was used; applied techniques match the task (delimiters for multi-block, tables for object comparison, self-check when quality is critical).
- [ ] **If skill** — Skill (`SKILL.md`): matches the **Anthropic Agent Skills spec** (name matches folder, description = what + when + keywords, optional folders); the **Agent Skills docs were consulted** (see "Agent Skills documentation (mandatory)").
- [ ] **Skill routing:** `description` is keyword-rich and task-oriented so the agent selects the skill by relevance.
- [ ] **Single responsibility** — no inline specifics about topics owned by another canonical rule/skill; only "see X" references.
- [ ] **Reasoning model** — no manual "think step by step" / prescriptive steps / leading few-shot; depth set via the effort/thinking knob; force-language dialed back. See [reasoning-models-thinking-effort-and-cot.md](reasoning-models-thinking-effort-and-cot.md).
- [ ] **Structured output** — machine-consumed output enforced with a schema, not prose JSON or prefill. See [structured-outputs-constrained-decoding.md](structured-outputs-constrained-decoding.md).
- [ ] **Agent / tool prompt** — "when to call" + side effects in tool descriptions; eagerness/persistence set deliberately; tool/retrieved/user content treated as untrusted. See [agentic-tool-use-prompting.md](agentic-tool-use-prompting.md), [untrusted-content-and-injection-defenses.md](untrusted-content-and-injection-defenses.md).
- [ ] **Vendor fit** — structuring/role/sampling conventions match the target model; not assumed to port. See [vendor-portability-cross-model-conventions.md](vendor-portability-cross-model-conventions.md).
- [ ] **Measured, not asserted** — non-trivial techniques validated on a dataset; deploys gated on no-regression; no unsourced accuracy percentages quoted as fact. See [evals-llm-as-judge-and-regression-gates.md](evals-llm-as-judge-and-regression-gates.md).

If any item is unchecked or the prompt still violates a practice — update the prompt and re-run this checklist before considering the task done.

## Checklist

- [ ] Quick gate is a coarse subset; the verification checklist is the single detailed source (no duplication).
- [ ] Example snippets use the same arrow, checkbox, and bullet glyphs as the rest of the skill.
