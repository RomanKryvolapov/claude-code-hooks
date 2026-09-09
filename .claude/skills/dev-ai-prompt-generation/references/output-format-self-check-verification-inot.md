## About this file

**What’s inside:** Code fences for code-only answers, self-check blocks (□ lists), verification-first and quit instructions, and the INoT solver/critic pattern.

**Hub:** [../SKILL.md](../SKILL.md)

---

## Backticks for code

To get code with no preamble, end the prompt with an opening code fence and let the model complete it — it fills the block with code, not chatter. On 2026 models a plain "return only code, no prose" instruction (or structured outputs) is more reliable.

- **❌ Chatty:** "Sure! I'll help you write a sort function. Here is a Python example…"
- **✅** End the prompt with `Write a sort function in Python` and an opening Python code fence (three backticks + `python`); the model continues inside it.

---

## Self-check

**When to apply:** Output correctness or format is critical (JSON, structured reports, factual answers).

**Why it works:** Asking the model to verify its answer against criteria before finishing catches errors for coding and math, and is vendor-endorsed. Caveats for 2026: LLMs are **unreliable** at self-verifying their own reasoning, and on reasoning models the verification is largely internal — treat an explicit self-check as a cheap safety net (a real lever on fast / non-reasoning models), not a guaranteed multiplier. Source: [SelfCheck](https://arxiv.org/abs/2308.00436).

The model does not self-check by default; ask and it will. More structure buys more reliability in roughly this order — but the accuracy percentages once attached to these levels have **no traceable source**; measure on your own evals → [evals-llm-as-judge-and-regression-gates.md](evals-llm-as-judge-and-regression-gates.md):

| Level                | Mechanism                             |
| -------------------- | ------------------------------------- |
| No control           | just answer                           |
| Ask to check         | "verify your answer before finishing" |
| Checklist            | explicit "□ …" criteria               |
| Validation           | enforce a schema (structured outputs) |
| Iterative correction | draft → critique → revise             |

A checklist block to embed in a prompt:

```
<self_check>
Before output, check:
□ All required fields filled?
□ Types match the schema?
□ No forbidden words?
□ Length within limit?
If any item fails — fix BEFORE output.
</self_check>
```

---

## Verification-first and quit instructions

- **Verification-first:** "here is an answer [any], verify it, then give the correct one." Can improve accuracy on some tasks — measure; do not assume a guaranteed lift.
- **Quit under uncertainty:** "If unsure, write 'Need clarification: …' instead of guessing."

---

## INoT (experimental — name and figures unverified)

> Experimental. INoT (a solver + critic that "debate" in one pass, then revise) comes from a single 2025 preprint; its quoted accuracy/token figures are not established. Mention it as one pattern, not a default — and on reasoning models the internal chain already does much of this.

A solver proposes, a critic finds weaknesses, the solver revises; repeat 2–3 rounds, then output the final solution. Use rarely — a high-stakes single answer where you want an explicit critique pass and have no eval-driven alternative.

## Checklist

- [ ] Self-check used as a cheap safety net, not a guaranteed multiplier; effect measured on evals.
- [ ] No fixed accuracy ladder quoted.
- [ ] On reasoning models, internal verification relied on; explicit self-check is a fallback.
