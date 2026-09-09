## About this file

**What’s inside:** Core principle, prompt/rule language, brevity, reformulation, structure, wording, examples, parameters, and size control.

**Hub:** [../SKILL.md](../SKILL.md)

---

## Core principle

- A prompt (or rule block) is a minimal formulation of the task, not a pile of instructions.
- If it works wrong → rewrite; do not add conditions.
- Fix the root cause of the formulation, not the symptom.

---

## Language

Write all prompts and rule text in English — it is the industry standard, models are trained mostly on English, and it keeps the codebase consistent and predictable.

- Exceptions: example data may be in any language when it is part of the input; variable names, placeholders, and technical terms stay as-is.
- Priority labels: 🔴 Critical · 🟡 Important · 🟢 Desirable.

---

## Brevity and density

- Short, concise, readable in one pass.
- Each sentence adds new meaning.
- Remove everything the task stays unambiguous without.
- One phrase → one instruction; do not split.

---

## Reformulate instead of extending

When logic needs clarification, first reformulate the existing instruction; add a new one only if reformulating is impossible.

Forbidden:

- Adding instructions to compensate for a bad formulation.
- Keeping the old formulation and "overriding" it with a new one.
- Leaving the old formulation next to the new behavior.

---

## Structure and clarity

- Simple, flat structure; one logical task → one block.
- Goal, constraints, and format as separate blocks.
- One sentence per line; multi-line format with triple quotes.
- JSON with line breaks, not a single line.

---

## Wording

- Positive, direct, verifiable wording.
- State outcome requirements, not prohibitions.
- Avoid "try to", "preferably", "when possible", "usually".

---

## Examples and detail

- Add an example only when it changes understanding; otherwise remove it.
- Do not describe corner cases.
- Focus on the main scenario and target outcome.

---

## Parameters and data

- Use variables, parameters, and placeholders; do not hardcode values.
- The prompt/rule stays generic when inputs change.

---

## Size control

A large prompt is a design mistake. When a prompt (or rule set) grows, the task is formulated wrong — simplify and rebuild rather than extend. Priority: clarity and brevity over covering every scenario.

## Checklist

- [ ] Prompt is a minimal formulation, not a pile of compensating instructions.
- [ ] English text; one instruction per line; no hedge words ("try to", "preferably").
- [ ] If it grew large, the task was reformulated, not extended.
