## About this file

**What’s inside:** The full “universal” prompt/rules template — Glossary, Role, Context, Task, Rules with priorities, Output, Examples, self-check.

**Hub:** [../SKILL.md](../SKILL.md)

---

## Universal prompt (and rules) template

> **A menu of slots, not a mandate.** Include only the blocks the task needs — a kitchen-sink prompt contradicts the skill's minimal-formulation / size-control principle. This template is for **general / non-reasoning** prompts; for a **reasoning model**, drop prescriptive steps and worked-reasoning examples and rely on goal + constraints + an output contract + the right effort knob. For **machine-consumed** output enforce native structured outputs, not the JSON block below. See [reasoning-models-thinking-effort-and-cot.md](reasoning-models-thinking-effort-and-cot.md) and [structured-outputs-constrained-decoding.md](structured-outputs-constrained-decoding.md).

**One structuring convention:** markdown headings give the structure; reserve XML tags for **injected data** (a boundary the model — and your security layer — can see). Do not double-label a block with both a heading and a same-named tag; if you mark injected data with a tag, drop the heading for that block.

```
# PROMPT / RULES: [name] — v1.0.0

## Glossary
H1 = main heading · CTA = call to action · TA = target audience

---

## Role
You are [role with experience and specialization].

---

## Context
{{data_to_process}}

---

## Task
[Clear description of what to do]

---

## Rules

### 🔴 Critical
- NEVER [prohibition 1]
- NEVER [prohibition 2]

### 🟡 Important
- [requirement 1]
- [requirement 2]

### 🟢 Desirable
- [recommendation]

---

## Output format
Machine-consumed → enforce a schema (structured outputs). Human-facing → describe it:
{
  "field1": "type | allowed values",
  "field2": number 1-10,
  "field3": ["array of strings"] | null
}

---

## Examples
=== separates examples
===
Input: [example 1]
Output: [result 1]
===
Input: [example 2]
Output: [result 2]
===

---

## Self-check
□ All fields filled?  □ Types match schema?  □ Prohibitions followed?  □ Format correct?
If any item fails — fix BEFORE output.
```

## Checklist

- [ ] Only the slots the task needs are included (a menu, not a mandate).
- [ ] One structuring convention: no heading paired with a same-named tag.
- [ ] Machine-consumed output uses a schema; reasoning models drop prescriptive steps.
