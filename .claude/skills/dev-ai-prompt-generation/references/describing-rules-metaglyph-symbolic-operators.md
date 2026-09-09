## About this file

**What’s inside:** MetaGlyph-style symbolic logic, operator reliability, an example, and ASCII alternatives — flagged experimental.

**Hub:** [../SKILL.md](../SKILL.md)

---

## MetaGlyph for compact logic

> ⚠️ **Experimental / not a default.** Symbolic-glyph compression rests on a single recent preprint; reported token savings vary by model and accuracy **degrades badly on smaller models** (near-zero operator fidelity), holding up only at large scale. Do not adopt it for general prompting. The one portable lesson: a few stable logic symbols (`¬`, `→`, `∈`) read fine; `∩` is unreliable — write it in prose. Prefer plain conditional rules or [pseudocode](describing-rules-pseudocode-if-switch-function-style.md).

**When to apply:** Rarely — only very long filter/set conditions on a large model, and only after measuring that the symbols help on your own evals. Use ASCII alternatives (`&&`, `||`, `!`) when symbols are unavailable.

### Symbols

| Group       | Symbols                                                                         |
| ----------- | ------------------------------------------------------------------------------- |
| Logic       | `∧` and · `∨` or · `¬` not · `→` therefore · `⇒` if-then · `↔` equivalent       |
| Sets        | `∈` in · `∉` not in · `⊂`/`⊆` subset · `∩` intersection · `∪` union · `∅` empty |
| Comparison  | `>` `<` `≥` `≤` `≠` `=`                                                         |
| Quantifiers | `∀` for all · `∃` exists · `\|` such that                                       |
| Operations  | `◦` compose · `↦` map · `∑` sum · `±` plus-minus · `≈` approx                   |

### Operator reliability (model-dependent)

Operator fidelity is **model-dependent and low on some models** — even membership (`∈`) scores poorly on some, and `∩` is widely confused with "list". Treat glyphs as unreliable unless measured: write set/membership conditions in prose (`∈(A), ∈(B), ¬(C)`), avoid `→` as a transformation verb (use "select" / "filter"), and verify on your own model. Per-operator accuracy percentages circulating online trace to one preprint and do not reproduce as stated.

### Example

```
users → apply:
  ∈(admin) ⇒ access = full
  ∈(moderator) ⇒ access = limited
  ∈(user) ⇒ access = basic
```

## Checklist

- [ ] Symbolic glyphs avoided by default; used only after measuring on the target model.
- [ ] Intersection and transformation arrows written in prose; only stable symbols kept.
- [ ] No per-operator accuracy percentages quoted as fact.
