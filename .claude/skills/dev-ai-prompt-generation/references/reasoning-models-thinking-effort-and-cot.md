## About this file

**What’s inside:** Reasoning models make thinking depth an API knob — per-vendor depth controls, the rule for when NOT to write chain-of-thought, effort calibration, and tricks deprecated on frontier models.

**Hub:** [../SKILL.md](../SKILL.md)

---

## The shift

Frontier models (Claude Opus 4.x / Fable 5, OpenAI GPT-5 / o-series, Gemini 3, Grok reasoning, DeepSeek-R1) reason internally before answering. Depth is set by a **model/API parameter**, not by prompt wording. The classic "make the model think" toolkit (manual `think step by step`, prescriptive step plans, leading few-shot) was built for pre-2024 completion models and now adds friction, verbosity, or accuracy loss on reasoning models.

> Rule of thumb: on a reasoning model, **raise the effort knob, do not add reasoning scaffolding.** Give a clear goal, hard constraints, and an output contract; let the model plan.

---

## Decision rule

```
IF model is a reasoning model (thinking on / effort knob exists):
    → state outcome + constraints + output contract; do NOT prescribe steps
    → do NOT write "think step by step" / manual CoT
    → start zero-shot; add few-shot only if an eval shows a specific gap
    → if reasoning is shallow → raise the effort/thinking level, not prompt tricks
    → audit for contradictory/redundant instructions (they cause overthinking)
ELSE (fast / minimal-effort / legacy instruct model):
    → CoT, few-shot, and explicit step plans still help (see reasoning-techniques-taxonomy.md)
```

---

## Per-vendor depth knobs

Set the depth control on the API; do not emulate it in the prompt.

| Vendor    | Parameter                               | Values                                                             | Notes                                                                                                                                                                                                      |
| --------- | --------------------------------------- | ------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Anthropic | adaptive thinking + `effort`            | `low` / `medium` / `high` (default) / `xhigh` / `max`              | `thinking:{type:'adaptive'}` lets the model decide whether/how much to think; auto-enables interleaved thinking with tools. On Opus 4.8/4.7 adaptive is the **only** mode (manual `type:'enabled'` → 400). |
| OpenAI    | `reasoning_effort` (+ `text.verbosity`) | `none` / `minimal` / `low` / `medium` (default) / `high` / `xhigh` | `none`/`minimal`/`xhigh` are model-specific. `verbosity` controls **final answer length** independently of thinking depth. Pin `reasoning_effort` first when migrating.                                    |
| Google    | `thinking_level`                        | `minimal` / `low` / `medium` / `high`                              | Replaces the legacy numeric `thinking_budget`; cannot send both. `minimal` is Flash / Flash-Lite only (not Pro). Default high on Gemini 3 Pro and Flash.                                                   |
| xAI       | `reasoning_effort`                      | `none` / `low` / `medium` / `high`                                 | Model-specific support; some variants reject it (e.g. Grok 4 errors on `reasoning_effort`).                                                                                                                |
| DeepSeek  | R1 reasons by default                   | —                                                                  | Force `<think>\n` start for thorough reasoning; no effort knob. `deepseek-chat` (V3) behaves like a normal model.                                                                                          |

Higher effort is **not** automatically better: with contradictory instructions or weak stop criteria it causes overthinking and wasted tool calls. Calibrate per task — high/xhigh for hard coding, planning, or accuracy-critical work; low/none for latency-bound extraction and simple lookups.

---

## When manual CoT and few-shot still help

They are **non-reasoning-model** techniques — keep them for fast/minimal-effort configs and legacy instruct models (e.g. arithmetic/logic on a fast model). On reasoning models:

- Manual CoT duplicates the hidden chain into visible tokens, inflates cost, and lets the narration drift from the actual reasoning.
- Few-shot can **degrade** reasoning models (DeepSeek-R1 README; OpenAI advises zero-shot first): the model imitates the exemplars instead of reasoning from scratch. See [output-format-few-shot-examples-and-order.md](output-format-few-shot-examples-and-order.md).
- Word-sensitivity: with thinking off, some Claude models over-react to the literal word "think" — prefer "consider", "evaluate", "reason through".

---

## Deprecated on frontier models

| Practice                                                 | Status                                                                    | Use instead                                                                                                                                           |
| -------------------------------------------------------- | ------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| Fixed `budget_tokens` thinking budget                    | Rejected (400) on Anthropic frontier; deprecated elsewhere                | Adaptive thinking + `effort`                                                                                                                          |
| Prefilling the assistant turn to force shape/labels      | Returns 400 on Anthropic Opus 4.6/4.7/4.8, Sonnet 4.6, Fable 5 / Mythos 5 | Structured outputs / tool enums / system instruction ([structured-outputs-constrained-decoding.md](structured-outputs-constrained-decoding.md))       |
| Stacking manual CoT on top of extended/adaptive thinking | Redundant, can hurt                                                       | One depth knob; clear brief                                                                                                                           |
| `temperature=0` "for determinism"                        | Removed (400) on some Anthropic frontier models; inverted on Gemini       | Steer via effort/prompt; sampling is vendor-specific ([vendor-portability-cross-model-conventions.md](vendor-portability-cross-model-conventions.md)) |

---

## Failure mode inversion (2026)

The dominant failure on frontier models is **over-triggering / overeagerness** — the opposite of 2023's "lazy model" problem. Aggressive language (`CRITICAL`, `You MUST`, `ALWAYS use the tool`) now causes over-triggering and over-exploration. Use plain conditional phrasing and state scope explicitly — see the de-prescription rule in [core-principles-task-markdown-tokens-caps.md](core-principles-task-markdown-tokens-caps.md) and [agentic-tool-use-prompting.md](agentic-tool-use-prompting.md).

---

## Sources

- Anthropic — Prompting best practices, Adaptive thinking, Effort (platform.claude.com).
- OpenAI — Reasoning best practices; GPT-5 / GPT-5.5 prompting guides (developers.openai.com).
- Google — Gemini 3 developer guide (`thinking_level`, temperature 1.0).
- DeepSeek-R1 README (no system prompt, force `<think>`, few-shot degrades).

## Checklist

- [ ] Depth set via the vendor effort/thinking knob, not manual chain-of-thought.
- [ ] Zero-shot outcome + constraints + output contract; force-language dialed back.
- [ ] Deprecated tricks (budget_tokens, prefill, temperature=0) avoided on frontier models.
