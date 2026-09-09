## About this file

**What’s inside:** Native schema enforcement / constrained decoding as the correct way to get machine-consumable output, why it replaces "reply only in JSON" prose and assistant-prefill tricks, plus per-vendor syntax, requirements, and caveats.

**Hub:** [../SKILL.md](../SKILL.md)

---

## What constrained decoding is

With strict structured outputs the API compiles your JSON Schema into a grammar and **masks tokens at each step** so the model cannot emit output that violates the schema. You get guaranteed-valid JSON **and** schema adherence — no parsing fallbacks, no retry-on-malformed loops.

> Default for any **machine-consumed** output (extraction, classification, agent state, tool arguments): enforce a schema on the API. Reserve "reply only in JSON" prose for chat/UX text a human reads, or models without the feature. Describing a JSON shape in the prompt (see [output-format-placeholders-json-schema-langchain.md](output-format-placeholders-json-schema-langchain.md)) is a **fallback**, not the contract.

---

## Per-vendor

| Vendor        | Field                                                                            | Key requirements                                                                                                                                                                                     |
| ------------- | -------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| OpenAI        | `text.format` = `{type:"json_schema", strict:true, schema:{…}}`                  | **Every** property in `required`; `additionalProperties:false`. Check the `refusal` field. `json_object` mode = valid JSON only (no schema); superseded by strict `json_schema` but still supported. |
| Anthropic     | `output_config.format` json_schema; or `messages.parse()`                        | Strict tools: `strict:true` on the tool def + `additionalProperties:false` + `required`. **Incompatible with Citations.**                                                                            |
| Google Gemini | `response_mime_type:"application/json"` + `response_schema`                      | Convey intent via the schema’s `description` fields. Gemini 3 can combine schema with built-in tools.                                                                                                |
| Self-hosted   | grammar-constrained runtimes (Outlines, Guidance, XGrammar, llguidance via vLLM) | Force tokens to a JSON Schema / regex / BNF when no native feature exists.                                                                                                                           |

Prefer schema enforcement over prose JSON on **every** vendor for machine-consumed output — a 2026 convergence point ([vendor-portability-cross-model-conventions.md](vendor-portability-cross-model-conventions.md)).

---

## Authoring rules

- **List every property in `required`** and set `additionalProperties:false` — strict modes demand it (OpenAI) and it stops invented fields everywhere.
- **Null over guess:** make optional/absent values explicit (`"effective_date": string | null`) and instruct "if a field is absent in the source, output null — never infer". Pairs with strict schema to kill hallucinated fields.
- **Describe fields in the schema**, not only in the prompt — schema `description`/`title` are read by the model and survive prompt edits.
- **Tool/function arguments** get the same strict treatment so the model emits valid arguments; put usage guidance in the tool description, not the prompt ([agentic-tool-use-prompting.md](agentic-tool-use-prompting.md)).

---

## Caveats

- **Validity ≠ correctness.** The schema guarantees shape, not that values are right. Validate semantics in code (enums against real data, ranges, referential integrity).
- **Not all constraints are enforced.** OpenAI silently does **not** enforce string `minLength`/`maxLength`, numeric `minimum`/`maximum`, or format hints — re-check in code. Anthropic **rejects** these (and `$ref` / recursive schemas) with a 400 as unsupported. Either way, do not rely on the schema for value-range validation.
- **Feature conflicts.** Strict/structured output can be incompatible with other features (e.g. Anthropic Citations; some programmatic/forced tool-calling modes). Read the vendor matrix before combining.
- **It is a jailbreak surface.** Constrained decoding is a control plane an attacker can exploit to coerce outputs; "it returned valid JSON" is not "it’s safe". See [untrusted-content-and-injection-defenses.md](untrusted-content-and-injection-defenses.md).

---

## Deprecated alternatives

| Old trick                                             | Why it’s gone                                                 | Use instead                                                        |
| ----------------------------------------------------- | ------------------------------------------------------------- | ------------------------------------------------------------------ |
| Prefill the assistant turn with `{` to force JSON     | Rejected by current Claude models; no schema guarantee anyway | Schema enforcement                                                 |
| `response_format: json_object` (valid-JSON-only)      | No schema guarantee                                           | Strict `json_schema`                                               |
| "Return ONLY JSON, no prose" prose with regex cleanup | Brittle; breaks on edge cases                                 | Schema enforcement (keep prose JSON only as a no-feature fallback) |

> **Not deprecated** — intermediate JSON for complex formats: for XML/BPMN/HTML, emit a simple intermediate JSON (ideally schema-enforced) and convert it in code rather than generating the complex format directly.

---

## Sources

- OpenAI — Structured Outputs guide + "Introducing Structured Outputs" (constrained decoding via context-free grammar).
- Anthropic — Structured outputs / strict tools docs.
- Google — Gemini structured output / response schema docs.
- Grammar-constrained decoding: Outlines, Guidance, XGrammar, llguidance.

## Checklist

- [ ] Machine-consumed output enforced with a strict schema (required keys, additionalProperties:false).
- [ ] Value ranges/lengths validated in code, not assumed from the grammar.
- [ ] Semantics validated even when the schema is satisfied.
