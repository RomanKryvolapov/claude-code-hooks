## About this file

**What’s inside:** Prompt decorators (`+++Reasoning`, `+++Tone`, stacking, role debates) — compact control tokens vs long prose. Flagged experimental.

**Hub:** [../SKILL.md](../SKILL.md)

---

## Prompt decorators — compact control tokens

> ⚠️ **Community DSL, not a vendor feature.** Decorators (`+++Reasoning`, `+++OutputFormat`) are a community convention, not something the APIs natively parse. Where the API exposes a real knob, **use the knob**: depth via `reasoning_effort` / `thinking_level` / `effort`, answer length via verbosity, output shape via structured outputs. On a reasoning model `+++Reasoning` is redundant with the effort knob. See [reasoning-models-thinking-effort-and-cot.md](reasoning-models-thinking-effort-and-cot.md) and [structured-outputs-constrained-decoding.md](structured-outputs-constrained-decoding.md).

**When to apply:** Only as a lightweight in-prompt convention for behaviors with **no** API control, on models that follow the pattern. Do not rely on them cross-vendor; adherence is best on strong instruction-following models, loose on smaller/older ones.

**Why it works (when it does):** Compact tokens act as meta-instructions and read as structural anchors; stacking gives a predictable how-to-think → how-to-express → how-to-format order. Community spec: [Prompt Decorators Specification](https://synaptiai.github.io/prompt-decorators/prompt-decorators-specification-v1.0/).

### Families

- **Cognitive / generative (how to think):** `+++Reasoning`, `+++Refine(iterations=N)`, `+++Debate(roles=[…])`, `+++Verify`, `+++Hypothesize`, `+++Synthesize`.
- **Expressive / systemic (how to output):** `+++Tone(style=…)`, `+++OutputFormat(type=json|markdown|table)`, `+++Length(…)`, `+++Audience(…)`, `+++Language(…)`.

### Common decorators

| Decorator                    | Replaces                      |
| ---------------------------- | ----------------------------- |
| `+++Reasoning`               | "show reasoning step by step" |
| `+++Tone(style=formal)`      | "use a formal tone"           |
| `+++OutputFormat(type=json)` | "output result as JSON"       |
| `+++Debate`                  | "consider multiple angles"    |
| `+++Refine(iterations=3)`    | "improve iteratively"         |

### Stacking

Decorators apply top to bottom in the order how-to-think → how-to-express → how-to-format:

```
+++Debate
+++Reasoning
+++OutputFormat(type=markdown)

Evaluate a startup's market-entry strategy.
```

Parameterized roles are possible (`+++Debate(roles=[…], rounds=3)`), but on 2026 models prefer the native effort / verbosity / structured-output knobs over decorator syntax.

## Checklist

- [ ] Native API knobs (effort/verbosity/structured outputs) preferred over decorators.
- [ ] Decorators used only where no API control exists; not relied on cross-vendor.
