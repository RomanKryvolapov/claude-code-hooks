## About this file

**What’s inside:** Pseudocode for conditional logic (`IF`/`ELSE`, `SWITCH`/`CASE`), defaults/fallbacks, and function-calling / docstring-style rule blocks.

**Hub:** [../SKILL.md](../SKILL.md)

---

## Pseudocode for conditional logic

**When to apply:** Conditional rules (if X then Y), multi-branch logic, or priority/format rules that depend on request type.

**Why it works:** Models read `IF/ELSE` and `SWITCH/CASE` as structure and apply branch logic more consistently than from prose. Gains are modest and task-dependent on older models — the 2023 source reports classification/generation improvements, not the round "+36% accuracy / −87% tokens" figure that circulates online. Use pseudocode for **deterministic branch rules** where it is the clearest form, **not** to force a reasoning sequence on a thinking model (prescriptive steps fight internal planning — see [reasoning-models-thinking-effort-and-cot.md](reasoning-models-thinking-effort-and-cot.md)). Source: [Prompting with Pseudo-Code Instructions](https://arxiv.org/abs/2305.11790).

### IF/ELSE

```
IF length(text) > 500 words:
    1. Extract 3-5 key theses
    2. Brief analysis per thesis
ELSE:
    1. Analyze text as a whole
    2. One paragraph of conclusions

IF language(input) != target_language:
    1. Detect source language
    2. Answer STRICTLY in target_language
```

### SWITCH/CASE and fallbacks

```
SWITCH request_type:
    CASE "question": → short answer + explanation
    CASE "task":     → step-by-step solution
    CASE "code":     → code only, no explanation
    DEFAULT:         → ask the user to clarify
```

`DEFAULT` is the SWITCH fallback branch; `FALLBACK` sets a value when data is missing (`tone: FALLBACK "neutral"`, `price: FALLBACK "on request"`).

---

## Function-calling style

Frame the task as a typed Python function with a docstring — models have seen large volumes of such code, so a signature + docstring reads as a structured contract. NOT for creative tasks; rigid structure kills variation.

```
def classify_ticket(text: str, categories: list[str] = ["Technical", "Billing", "Account"]) -> dict:
    """
    Classify a support ticket.
    Returns: {"category": str, "confidence": float 0.0-1.0, "reasoning": str}
    Constraints: confidence < 0.7 → category = "Unknown"; reasoning ≤ 50 words.
    """
```

## Checklist

- [ ] Pseudocode used for deterministic branch rules, not to script a reasoning model.
- [ ] No fixed accuracy/token numbers claimed.
- [ ] Function/docstring style used for contracts, not for creative tasks.
