## About this file

**What’s inside:** Treating prompts like code — measure changes on a dataset, gate on no-regression, grade open-ended output with analytic rubrics plus LLM-as-judge (with bias controls), and let optimizers tune prompts. The home of "measure on your own evals".

**Hub:** [../SKILL.md](../SKILL.md)

---

## Why this file exists

There is no universal "+X% accuracy" for a prompt technique — effects are model- and task-dependent, and prompt-format sensitivity is large. The only trustworthy number is the one you measure on **your** task. Where older guidance quotes a fixed uplift, run it on your eval set and keep what wins.

---

## EvalOps loop

```
curate a representative dataset
  → baseline the current prompt/model
  → change ONE thing (wording, technique, schema, effort, model)
  → rerun the dataset
  → compare per criterion
  → ship only on lift / no-regression
```

Before a **model migration**, re-baseline both behavior and token counts — a new model version is a new tuning target, not a drop-in ([vendor-portability-cross-model-conventions.md](vendor-portability-cross-model-conventions.md)).

---

## Grade with analytic rubrics, not holistic scores

Score each criterion separately (correctness, completeness, format adherence, grounding, tone) instead of one opaque 1–10, so you can root-cause a regression to a specific dimension. Combine:

- **Code/deterministic checks** for cheap, exact things: schema valid, contains a citation, within length, no forbidden token.
- **LLM-as-judge** for open-ended quality that exact-match can’t capture.

---

## LLM-as-judge bias controls

A judge model has systematic biases; counteract them or the scores lie:

- **Position/order bias:** in pairwise comparison, swap the order and average — otherwise the first (or last) candidate wins on position alone.
- **Length bias:** instruct the judge to score on correctness/completeness only and not to favor longer answers.
- **Self-preference:** judges over-rate outputs from their own family; use a different model as judge where possible.
- **Validate the judge** against a sample of human labels before trusting it at scale; ask for a short per-criterion justification, not just a number.

---

## Self-check is a practice, not a guaranteed multiplier

Asking the model to verify its answer against criteria before finishing catches errors for coding/math and is vendor-endorsed. But LLMs are **unreliable at self-verifying their own reasoning**, and on reasoning models self-checking is largely internal — treat an explicit self-check as a cheap safety net and **measure** its effect rather than assuming a fixed jump. The old "60% → 97%" ladders have no traceable source; do not quote them.

---

## Automatic prompt optimization

With an eval metric, let an optimizer discover instructions/examples you wouldn’t hand-write — it beats hand-tuning. Pragmatic order:

1. **Auto few-shot** (select demos from labeled data).
2. **DSPy + MIPROv2** (Bayesian, data- and demo-aware) when formats drift across a multi-step pipeline.
3. **GEPA** (reflective/evolutionary, Pareto, sample-efficient) — often the strongest, pairs with an LLM-judge metric.
4. **Fine-tune** only if drift/traffic genuinely demands it.

Point at the ecosystem (DSPy, GEPA); don’t reproduce their APIs — wire the optimizer to the same metric your eval set already scores, rather than standing up a second harness.

---

## Sources

- Industry "EvalOps" consensus (2025); analytic-rubric and LLM-as-judge bias literature.
- Self-verification limits — research on LLMs verifying their own reasoning/planning.
- DSPy (dspy.ai), MIPROv2; GEPA (Agrawal et al., 2025).

## Checklist

- [ ] Changes measured on a curated dataset; deploys gated on no-regression; model bumps re-baselined.
- [ ] Open-ended output graded with analytic rubrics + LLM-judge bias controls.
- [ ] No unsourced accuracy percentages stated as fact.
