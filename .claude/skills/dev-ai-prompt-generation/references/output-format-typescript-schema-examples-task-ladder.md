## About this file

**What’s inside:** TypeScript-style output contracts, markdown comparison tables, the control ladder (request → schema → rules), and the Schema / Examples / Task template.

**Hub:** [../SKILL.md](../SKILL.md)

---

## TypeScript typing

Union types, optional fields, and generics give maximum strictness — the model reads TypeScript types.

> For **machine-consumed** output, enforce the shape with native structured outputs rather than the type description alone ([structured-outputs-constrained-decoding.md](structured-outputs-constrained-decoding.md)). A described type steers; constrained decoding guarantees.

```
type ReviewAnalysis = {
  sentiment: "positive" | "negative" | "neutral";
  score: 1 | 2 | 3 | 4 | 5;
  confidence: number;        // 0.0-1.0
  keywords: string[];        // max 5
  issues?: string[];         // only if score < 3
  recommendation: { action: "buy" | "wait" | "skip"; reason: string };
}
// Return an object of type ReviewAnalysis
```

---

## Markdown tables for comparison

A table forces concreteness: every cell needs a value, so the model cannot pad with fluff.

```
Compare three options. Answer ONLY as a table:

| Criterion | Option A | Option B | Option C |
| --- | --- | --- | --- |
| Price | [price] | [price] | [price] |
| Quality | [1-5] | [1-5] | [1-5] |
| Verdict | [recommendation] | ... | ... |

Then one paragraph (≤ 50 words).
```

---

## Control ladder

Six levels from chaos to guarantee; each step adds predictability.

1. **Request** — "Analyze the review" → chaos
2. **+ Example** — "Here is a good analysis…" → hint
3. **+ Template** — "Fill: Sentiment `[...]`, Score `[...]`" → structure
4. **+ Schema** — `{"sentiment": "...", "score": ...}` → fields
5. **+ Types** — `"positive" | "negative"` → validation
6. **+ Rules** — `IF score < 3 THEN issues required` → guarantee

---

## Schema–Examples–Task

**When to apply:** Output must be highly predictable and the format is strict (classification, extraction, reports).

**Why it works:** Three separated blocks — Schema (WHAT), Examples (HOW), Task (ON WHAT) — keep the contract, demonstrations, and input apart; strong for strict-format extraction on instruct / fast models. On reasoning models, drop the worked-reasoning examples and lean on the schema plus a clear task — few-shot can degrade them ([output-format-few-shot-examples-and-order.md](output-format-few-shot-examples-and-order.md)).

```
## SCHEMA
{ "category": "string", "sentiment": "positive"|"negative"|"mixed", "score": 1-5, "issues": ["string"]|null }

## EXAMPLES   (=== separates examples)
===
Review: "Great vacuum!" → {"category":"vacuum","sentiment":"positive","score":5,"issues":null}
===
Review: "Camera good but battery weak." → {"category":"camera","sentiment":"mixed","score":3,"issues":["battery"]}
===

## TASK
Review: "Phone is ok but heats up when gaming" →
```

The biggest gain comes from helping the model start the _current_ task, not from showing other tasks. On fast/non-reasoning models, CoT + a few examples is strong and this template packages it; on reasoning models, set the effort knob and state the task ([reasoning-models-thinking-effort-and-cot.md](reasoning-models-thinking-effort-and-cot.md)).

## Checklist

- [ ] Output contract uses types/schema; machine-consumed shape enforced via structured outputs.
- [ ] On reasoning models, worked-reasoning examples dropped; schema + clear task kept.
- [ ] Comparison answers forced into a table where every cell needs a value.
