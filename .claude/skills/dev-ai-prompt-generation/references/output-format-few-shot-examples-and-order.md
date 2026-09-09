## About this file

**What’s inside:** Few-shot sizing, `===` delimiters, contrasting pairs, CoT inside examples, and example ordering.

**Hub:** [../SKILL.md](../SKILL.md)

---

## Few-shot examples

**When to apply:** New or non-obvious output format; classification with clear boundaries; when the model should copy both format and reasoning style.

**Why it works:** Examples steer format, tone, and edge-case handling without fine-tuning. Too many can hurt — the optimal count depends on model and task. The last example is remembered best, so make it the main or hardest. Contrasting pairs (good vs bad) clarify quality boundaries. Study: [Language Models are Few-Shot Learners (GPT-3)](https://arxiv.org/abs/2005.14165).

> ⚠️ **Reasoning-model caveat.** Few-shot is a **non-reasoning-model** technique. On frontier reasoning models (o-series, GPT-5 reasoning, Gemini 3 high-thinking, DeepSeek-R1), leading with examples can **degrade** accuracy — the model imitates the exemplars instead of reasoning from scratch. Start **zero-shot**; add 1–2 tightly-aligned examples only when an eval shows a specific format/edge-case gap. "Show reasoning in examples" applies only when thinking is off. Prefer positive style examples over negative "don't do X" ones. See [reasoning-models-thinking-effort-and-cot.md](reasoning-models-thinking-effort-and-cot.md). The guidance below is for instruct / fast models.

### How many

| Count         | When                                |
| ------------- | ----------------------------------- |
| 0 (zero-shot) | simple task, model knows the format |
| 1–2           | show a specific output format       |
| 3–5           | complex classification, edge cases  |
| 5+            | rarely needed                       |

### Format

Separate examples with `===`, set off the task with its own heading (or a `---`), and state the markup. The arrow `→` shows input → output:

```
## Examples (=== separates examples)
Input: "Card payment not working" → Billing, high (payment mentioned)
===
Input: "App crashes on launch" → Technical, medium (bug)
===
Input: "Change email in profile" → Account, low (settings)

## Now process:
"Double charge for subscription"
```

### CoT in examples (thinking off)

Show the reasoning, not just the result, so the model applies the same logic — only when extended/adaptive thinking is off:

```
Input: "Can't pay by card, shows error"
Reasoning: payment + error → payment context → Billing; blocks purchase → high.
→ Billing, high
```

### Contrasting pair (good vs bad)

One input, two outcomes with a short reason — clarifies quality boundaries better than positive-only examples:

- **❌ Bad:** "Great headphones, recommend buying." (too short, no specs)
- **✅ Good:** "Sony WH-1000XM5 with active noise cancellation. Up to 30 h battery, quick charge (3 min = 3 h). LDAC for Hi-Res." (concrete, objective)

### Order

- Last example = main (the model remembers the end best).
- Contrasting pairs improve boundary understanding; varied formats help the model adapt.

## Checklist

- [ ] Zero-shot first on reasoning models; few-shot only when an eval shows a gap.
- [ ] Examples separated with `===`; main/hardest example last; positive over negative.
- [ ] CoT-in-examples only when thinking is off.
