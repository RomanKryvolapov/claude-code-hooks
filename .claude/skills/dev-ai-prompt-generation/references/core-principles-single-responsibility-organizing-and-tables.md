## About this file

**What’s inside:** Single responsibility for rules/skills, prompt structure (delimiters, headings, arrows), markdown formatting, and tables for structured data.

**Hub:** [../SKILL.md](../SKILL.md)

---

## Single responsibility for rules and skills

Each rule or skill owns exactly one topic. When a canonical file owns a topic (e.g. this skill's [structured-outputs-constrained-decoding.md](structured-outputs-constrained-decoding.md) owns schema enforcement), other files **reference** it, never duplicate it.

Allowed elsewhere:

- A brief "see `<canonical-skill>`" reference.
- A one-line summary naming the topic without restating specifics.

Forbidden elsewhere:

- Listing fields, thresholds, patterns, or steps the canonical rule defines.
- Rephrasing the same rules in other words (shadow copy).
- An alternative or extended version of the canonical content.

When the canonical rule lacks something another file needs, add it there and reference it. One source of truth is updated; all consumers stay in sync.

---

## Delimiters and headings

**When to apply:** Any prompt with multiple blocks (role, task, data, examples).

**Why it works:** Delimiters set clear boundaries so the model does not glue instructions to data. Prompt-format sensitivity is real and sometimes large — equivalent reformattings of one prompt can move accuracy materially, and the chosen delimiter acts as a hidden control parameter. The takeaway is not a fixed percentage: pick one delimiter convention, state it once ("Examples are separated by ==="), and stay consistent. Measure on your own evals → [evals-llm-as-judge-and-regression-gates.md](evals-llm-as-judge-and-regression-gates.md). Background: [A single character can make or break your LLM evals](https://arxiv.org/abs/2510.05152).

### Reliable delimiters

| Delimiter | Purpose                   | Example                                  |
| --------- | ------------------------- | ---------------------------------------- |
| `---`     | between logical blocks    | Role --- Task --- Rules                  |
| `===`     | between few-shot examples | === Example 1 === Example 2 ===          |
| `###`     | section with heading      | ### Processing rules                     |
| `***`     | semantic break            | End of instructions \*\*\* Start of data |
| `│`       | split for counting        | System-2 counting                        |
| `◆◆◆`     | system instructions       | prompt-injection guard                   |

### Avoid these delimiters

| Delimiter        | Problem                            |
| ---------------- | ---------------------------------- |
| `~~~~`           | confused with markdown code blocks |
| `____`           | visually merges, weak signal       |
| `....`           | reads as continuation of text      |
| `////`           | confused with code comments        |
| blank lines only | weak signal, may be ignored        |

### Explicit delimiter rule

State the delimiter's purpose ("(=== separates examples)") to reduce ambiguity; measure the effect on your own evals (no fixed percentage).

```
## Data format in this prompt
- Examples are separated by "==="
- Sections are separated by "---"
- User data is wrapped in triple quotes """

## Examples
===
Input: "Great product!"
Output: {"sentiment": "positive"}
===
Input: "Terrible quality"
Output: {"sentiment": "negative"}
===

## Task
Process the user data.
```

### Headings and hierarchy

| Level | Use                        | Per prompt |
| ----- | -------------------------- | ---------- |
| `#`   | main topic / title         | 0–1        |
| `##`  | main sections (Role, Task) | 3–7        |
| `###` | sub-sections               | as needed  |

### Arrow → for transformations

`→` shows "before → after": few-shot (`"payment not working" → Billing, high`), priority rules (`premium → always high`), transformation (`raw text → structured prompt`), pseudocode (`IF condition: → action`).

---

## Markdown formatting

**When to apply:** Any prompt; models are trained on markdown.

**Why it works:** Headings, lists, and emphasis read as structure and help the model find the right part of a prompt. Do not over-read this as "formatting beats content": on reasoning models a clear goal, hard constraints, and the right effort level dominate, and heavy formatting can hurt smaller models. Structure for clarity, not as a substitute for a well-specified task.

| Element     | Syntax       | When to use                 |
| ----------- | ------------ | --------------------------- |
| **Bold**    | `**text**`   | key terms, emphasis         |
| _Italic_    | `*text*`     | notes, names                |
| Inline code | `` `code` `` | variables, commands, values |
| List        | `- item`     | requirements                |
| Numbering   | `1. step`    | ordered steps               |
| Quote       | `> text`     | examples, excerpts          |

---

## Tables for structured data

**When to apply:** Comparing objects by parameters or filtering by several conditions. Not for ordered steps (use a list).

**Why it works:** Row = one object, column = one property, which the model localizes well — tables help factual retrieval and comparison. The gain is task-dependent: treat it as qualitative, not a fixed percentage, and confirm on your own data. Background: [Talking with Tables for Better LLM Factual Data Interactions](https://arxiv.org/abs/2412.17189).

**❌ Prose:** Netflix is $16, 4K HDR, has a family plan. Hulu is $12, 1080p, has a family plan. Disney+ is $18, 4K HDR, no family plan.

**✅ Table:**

| Service | Price | Quality | Family |
| ------- | ----- | ------- | ------ |
| Netflix | $16   | 4K HDR  | Yes    |
| Hulu    | $12   | 1080p   | Yes    |
| Disney+ | $18   | 4K HDR  | No     |

### When to use which format

| Situation                       | Format | Why                             |
| ------------------------------- | ------ | ------------------------------- |
| Comparing objects by parameters | table  | row = object, column = property |
| Filtering by several conditions | table  | exact condition matching        |
| Sequence of steps               | list   | order matters, not parameters   |
| Unstructured text               | prose  | no clear structure              |

## Checklist

- [ ] Each rule/skill owns one topic; canonical topics referenced, not duplicated.
- [ ] One delimiter convention, stated once; logical blocks separated.
- [ ] Comparison data in a table (row = object, column = property); steps in a list.
