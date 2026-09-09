## About this file

**What’s inside:** Prompt caching (the prefix-match invariant and authoring rules), plus model routing and effort tiering — the levers that actually cut token cost, instead of symbolic-glyph compression.

**Hub:** [../SKILL.md](../SKILL.md)

---

## Caching beats compression

Squeezing wording (symbolic glyphs) saves a few tokens and is brittle. Caching a stable prefix saves most of the cost on repeated calls and is robust. Any workload that re-sends a large stable prefix — system prompt, tool defs, few-shot block, retrieved docs, conversation history — should cache it.

---

## The prefix-match invariant

The cache key is the **exact bytes** up to each breakpoint. Any byte change anywhere in the prefix invalidates the cache for everything after it. Reads are far cheaper than fresh input; writes cost slightly more than fresh.

> One stray byte in the prefix — a timestamp, a UUID, a per-user id, reordered tools — silently drops the cache and you pay full price with no error.

---

## Authoring rules

- **Render order:** tools → system → messages. Cache from the front.
- **Freeze the prefix.** No `now()`, per-request UUID, or per-user id in the system prompt or tool defs. Dynamic per-request facts go in a **late message**, not the prefix ([context-engineering-window-memory-compaction.md](context-engineering-window-memory-compaction.md)).
- **Stable first, volatile last.** Order content so the unchanging part is a contiguous prefix.
- **Place breakpoints** on the last stable block (e.g. `cache_control` at the end of the system/tool/document section). Explicit breakpoints are limited (Anthropic: up to 4).
- **Mind the minimum.** Prefixes below the model’s minimum cacheable length aren’t cached. Minimums vary by model — e.g. 1,024 tokens on current Sonnet/Opus, up to 4,096 on some Haiku and older Opus versions. Check the model’s page.
- **Verify.** Read cache-read token counts in the usage response; zero reads means a silent invalidator upstream.

---

## Caveat for reasoning models (Anthropic)

On Anthropic, changing the thinking mode/budget invalidates **message-level** cache (system prompt and tool definitions stay cached). Keep the thinking mode stable across a cached conversation.

---

## Complementary cost levers

- **Effort tiering.** The depth knob is the in-model cost lever: `low`/`none` for extraction and simple sub-tasks, `high`/`xhigh` only for hard reasoning/agentic work ([reasoning-models-thinking-effort-and-cot.md](reasoning-models-thinking-effort-and-cot.md)).
- **Model routing.** Route to the cheapest model that meets the bar — small models for classification/extraction, top tier for agentic/reasoning. Spawn a cheaper-model subagent for simple sub-tasks rather than switching the main loop’s model mid-session (a model switch invalidates the cache).
- **Context editing / compaction** cut tokens on long loops ([context-engineering-window-memory-compaction.md](context-engineering-window-memory-compaction.md)).

---

## Token-count gotcha

Counting tokens with another vendor’s tokenizer (e.g. an OpenAI tokenizer for Claude) is wrong by a wide margin, more so on code/non-English. Use the model’s own token-count endpoint and re-baseline per model on migration ([vendor-portability-cross-model-conventions.md](vendor-portability-cross-model-conventions.md)).

---

## Sources

- Anthropic — Prompt caching docs (prefix invariant, breakpoints, minimums, adaptive-thinking caveat).
- OpenAI / Google — automatic and explicit prompt caching docs.

## Checklist

- [ ] Prefix frozen (no timestamps/UUIDs/per-user ids); stable first, volatile last.
- [ ] Cache reads verified in the usage response.
- [ ] Effort tiering / model routing used as complementary cost levers.
