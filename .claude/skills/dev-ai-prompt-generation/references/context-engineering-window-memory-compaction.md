## About this file

**What’s inside:** Context engineering — the successor framing to prompt engineering for agentic/RAG apps — and managing the window as a resource: context rot, ordering, compaction vs context editing vs memory, and injecting dynamic context safely.

**Hub:** [../SKILL.md](../SKILL.md)

---

## From prompt string to context window

Once an app has tools, multi-turn state, or retrieval, the wording of a single prompt matters far less than **what enters the context window and when**. The job becomes a pipeline: select → rank → format → order → compact. This reframing (widely adopted across the industry since mid-2025) is "context engineering". Prompt-string techniques elsewhere in this skill still apply to each piece you place; placement and budget decisions live here.

```
retrieve → rank → format → place stable content first / volatile last → compact near the limit
```

---

## Context rot: more context is not free

Model quality degrades as the window fills, well below the nominal limit, and information buried in the **middle** is used worse than information at the start or end ("lost in the middle", not recency). Implications:

- Put the question and the most load-bearing evidence at an **edge** (top or bottom), not the middle.
- **Cap** retrieved chunks; prefer precision over volume — flooding a reasoning model dilutes accuracy (see [rag-grounding-and-citations.md](rag-grounding-and-citations.md)).
- **Prune** stale tool results instead of letting them accumulate.

---

## Three distinct window-management tools

Different features — do not confuse them:

| Tool                     | What it does                                                              | When                                                      |
| ------------------------ | ------------------------------------------------------------------------- | --------------------------------------------------------- |
| **Compaction**           | Summarizes earlier turns into a compact block when nearing the limit      | Long conversations/agent loops that may exceed the window |
| **Context editing**      | Clears stale tool results / old thinking from the transcript (no summary) | Long agent loops where old tool output is irrelevant      |
| **Cross-session memory** | A persistent store the model reads/writes across sessions                 | Remembering preferences, project facts, prior corrections |

Memory hygiene: tell the model where to write, when to consult it, and use a one-lesson-per-entry format. **Never store secrets/PII in memory or prompts** — see [untrusted-content-and-injection-defenses.md](untrusted-content-and-injection-defenses.md).

---

## Ordering for the cache

Render order is normally tools → system → messages. Put **stable** content first and **volatile** content last so the cacheable prefix stays intact — also the prompt-caching invariant ([prompt-caching-and-cost-levers.md](prompt-caching-and-cost-levers.md)).

---

## Inject dynamic context late, never into the frozen prefix

Per-request facts (current date, user state, mode switches) must **not** go into the system prompt — that breaks the cache and, for per-user data, leaks across the prefix. Inject them as a **late message**. A mid-conversation `system` message (where supported) is the cache-safe, non-spoofable operator channel for mode switches and runtime state; phrase it as context, not as override commands, and fall back to a clearly-marked reminder block on models that don’t support it.

---

## Long-horizon agents

For tasks spanning multiple context windows: have the agent write structured state to a file (progress, decisions, a checklist), checkpoint with version control, and tell it that its context will be compacted so it does not stop early. Have the harness track the remaining budget, and instruct the agent to checkpoint and resume from written state before compaction.

---

## Sources

- Karpathy popularized the "context engineering" framing (2025, social posts); Anthropic, "Effective context engineering for AI agents".
- "Lost in the Middle" (Liu et al., arXiv:2307.03172); 2025 long-context degradation studies.
- Anthropic — compaction, context editing, and memory-tool docs.

## Checklist

- [ ] Load-bearing content placed at a window edge, not the middle; retrieved chunks capped.
- [ ] Compaction / context editing / memory used appropriately; secrets and PII never stored.
- [ ] Dynamic per-request facts injected as a late message, not in the frozen prefix.
