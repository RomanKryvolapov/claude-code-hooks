## About this file

**What’s inside:** What ports across vendors and what does not — structuring (XML vs Markdown), role conventions (system vs developer vs no-system-prompt), vendor-specific sampling, Llama special tokens, and migration hygiene.

**Hub:** [../SKILL.md](../SKILL.md)

---

## Pick one convention per prompt; do not assume portability

**State one default target vendor per project** and apply its conventions consistently; if none is stated, assume the vendor whose API the surrounding code already calls. Use this file to see what changes when you target a different one.

A prompt tuned for one model family is **not** a drop-in for another. Three things diverge by vendor and even by model type: how you structure the prompt, where instructions live, and how you sample. Choose one convention per prompt, apply it consistently, and re-baseline on every model change.

---

## Structuring: XML vs Markdown (convergence point)

| Vendor            | Preferred structuring                                      | Notes                                                                                  |
| ----------------- | ---------------------------------------------------------- | -------------------------------------------------------------------------------------- |
| Anthropic         | **XML tags** (`<instructions>`, `<context>`, `<document>`) | Anthropic recommends XML tags as Claude’s preferred (canonical) structuring primitive. |
| OpenAI            | Markdown sections / headings                               | App rules go in the **developer** message.                                             |
| Google Gemini     | Markdown or XML labeled sections                           | Favors directness; terse by default.                                                   |
| xAI Grok, Mistral | Markdown or XML                                            | Both endorse labeled markup for task/constraints/context.                              |

**Convergence:** Google, xAI, and Mistral now all recommend labeled markup (XML tags or Markdown headings) for separating task / constraints / context — matching Anthropic’s long-standing advice. This is the single most portable structuring technique. Detail and the XML tag set: [data-presentation-xml-tags-and-universal-template.md](data-presentation-xml-tags-and-universal-template.md).

---

## Role conventions: where instructions live

| Vendor / model                      | Where app rules go                                              | Notes                                                                                                                                                  |
| ----------------------------------- | --------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| OpenAI reasoning models (o1+/GPT-5) | **`developer`** message (Responses API `instructions`)          | `developer` replaced `system`. Chain of command (Model Spec): platform/system > developer > user > assistant/tool (assistant/tool carry no authority). |
| Anthropic                           | `system` prompt; mid-conversation `system` messages on Opus 4.8 | Durable rules at system level; never the user turn.                                                                                                    |
| Gemini, Grok, Mistral, Llama        | `system` instruction                                            | A well-formed system prompt improves refusal calibration / reduces false refusals on Grok and Llama.                                                   |
| **DeepSeek-R1**                     | **No system prompt** — all instructions in the user turn        | R1-specific; `deepseek-chat` (V3) reverts to normal system-prompt behavior.                                                                            |

Durable product/policy rules belong at system/developer level for instruction-following **and** injection resistance — see [untrusted-content-and-injection-defenses.md](untrusted-content-and-injection-defenses.md).

---

## Sampling is vendor-specific (and partly inverted)

The old "low temperature = deterministic/reliable" heuristic does **not** transfer to reasoning models.

| Vendor / model                             | Temperature                                     | Other                                                                      |
| ------------------------------------------ | ----------------------------------------------- | -------------------------------------------------------------------------- |
| Gemini 3                                   | Keep **default 1.0**                            | Lowering can cause looping / degraded reasoning.                           |
| DeepSeek-R1                                | **0.5–0.7 (0.6)**                               | Too high fractures reasoning; too low → repetition.                        |
| Anthropic frontier (Opus 4.7/4.8 and up)   | `temperature`/`top_p`/`top_k` **removed (400)** | Steer via effort + prompt.                                                 |
| xAI Grok reasoning                         | —                                               | `presence/frequency penalty` and `stop` **error**; use `reasoning_effort`. |

Strip inherited sampling settings when migrating; do not copy a temperature from one vendor to another.

---

## Llama special tokens (self-hosted)

Llama uses explicit role/turn tokens; getting them exact matters when calling weights directly (vLLM/TGI/Bedrock raw). **Use the tokenizer’s chat template — do not hand-build.** Llama 4 renamed Llama 3’s tokens: `<|start_header_id|>`/`<|end_header_id|>`/`<|eot_id|>` became `<|header_start|>`/`<|header_end|>`/`<|eot|>`. Hand-built Llama 3 strings silently break on Llama 4.

---

## Structured output and depth knobs port as _concepts_, not syntax

- Every major vendor offers schema-enforced output, but the field differs (OpenAI `text.format` json_schema, Anthropic `output_config.format`, Gemini `response_schema`) — see [structured-outputs-constrained-decoding.md](structured-outputs-constrained-decoding.md).
- Every reasoning vendor offers a depth knob, but with different names/values — see the table in [reasoning-models-thinking-effort-and-cot.md](reasoning-models-thinking-effort-and-cot.md).

---

## Migration hygiene (every model bump)

Treat each new model version as a **new tuning target**, not a drop-in:

1. Swap the model.
2. Pin the depth knob (`reasoning_effort` / `thinking_level` / `effort`) to match prior latency.
3. Baseline on your eval set (behavior **and** token counts).
4. Change one thing at a time; re-measure.
5. Start from the smallest prompt that preserves the product contract — strip carried-over scaffolding and force-language.

Gates and measurement: [evals-llm-as-judge-and-regression-gates.md](evals-llm-as-judge-and-regression-gates.md).

---

## Sources

- Anthropic, OpenAI (Reasoning best practices, Model Spec), Google Gemini 3 guide, xAI reasoning docs, Mistral best practices, DeepSeek-R1 README, Meta Llama 4 prompt format.

## Checklist

- [ ] One convention per prompt; not assumed to port across vendors.
- [ ] Role placement and sampling match the target model.
- [ ] Re-baselined (behavior + tokens) on every model bump.
