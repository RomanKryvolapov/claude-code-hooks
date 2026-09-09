## About this file

**What’s inside:** Prompting for tool-using agents — where tool guidance lives, controlling eagerness vs persistence, tool preambles, parallel calls, the ReAct loop, multi-agent orchestration, and the 2026 over-triggering failure mode.

**Hub:** [../SKILL.md](../SKILL.md)

---

## Put tool guidance in the tool, not the prompt

Each tool description states **what it does, WHEN to call it, its inputs, and its side effects** — not just a name. Keep the system/developer prompt about goals; keep tool mechanics in the tool description. This improves should-call accuracy and keeps the prompt portable.

```
description: "Books an appointment slot. Use when the user confirms a specific
date and time. Sends a confirmation email (side effect). Inputs: slot_id, user_id."
```

Combine with strict argument schemas so the model emits valid arguments ([structured-outputs-constrained-decoding.md](structured-outputs-constrained-decoding.md)).

---

## The 2026 inversion: models over-trigger

Frontier models follow instructions literally and act eagerly, so old anti-laziness prompting backfires:

- `CRITICAL: You MUST use this tool` / `ALWAYS` / `if in doubt, call X` → **over-triggering** and over-exploration. Use conditional phrasing: "Use X when …".
- Forced scaffolding ("after every 3 tool calls, summarize") → excess output on models that already narrate well. Remove it; add a silence default if too chatty.
- Conversely, some frontier models **under-reach** for specific capabilities (search, memory, subagents). Fix by adding explicit trigger conditions in the tool description **and** reinforcing in the system prompt — not by shouting.

---

## Control eagerness (bound) vs persistence (more)

Two opposite levers, depending on whether the agent over- or under-acts.

**Bound it** (over-searches, redundant calls, slow); pair with lower reasoning effort:

```
<context_gathering>
Budget: at most 2 tool calls. Early stop: when top results converge (~70%) on one answer.
If uncertain, proceed with your best guess and note the assumption.
</context_gathering>
```

**Push persistence** (hands back too early on long autonomous tasks); pair with higher reasoning effort:

```
<persistence>
Keep going until the task is fully resolved before yielding. Do not stop on uncertainty —
deduce the most reasonable approach and continue. Tell the user only when done or truly blocked.
</persistence>
```

---

## Tool preambles and progress updates

On long or multi-tool rollouts, have the agent restate the goal in one sentence and outline its plan before acting, then emit short status updates so the user is not watching a silent spinner. Keep preambles to 1–2 sentences; do not let them bloat the final answer.

---

## Parallel tool calls

Issue independent tool calls in **one** assistant message and return their results in **one** user message. Splitting results across messages trains the model to stop parallelizing. Mark each result with its tool-call id; failed tools return an error result (not dropped). Instruct: "make independent calls in parallel; call dependent ones sequentially; never guess missing parameters."

---

## Loops and patterns

- **ReAct** (reason → act → observe → repeat) is the backbone scaffolding for tool-using agents; with reasoning models the "reason" step is largely internal.
- **Reflexion** (reflect on a failed attempt, store the critique, retry) helps **only** when there is a real success signal (tests pass/fail, task outcome).
- **Programmatic / code-driven tool calling**: have the model compose tool calls inside a code-execution step so large intermediate results stay out of context and only the final result returns.

---

## Multi-agent orchestration

When work genuinely fans out, use an **orchestrator + isolated subagents**: a lead owns full context and delegates to ephemeral subagents that start with clean context and return compressed summaries. Fresh context beats accumulated context. Each subagent prompt must carry **objective + output format + tool/source guidance + explicit task boundaries** — this gives the largest quality lift. Cost is high (token usage explains most of the variance), so reserve it for real parallelism, not simple sequential tasks. Context hygiene across long loops: [context-engineering-window-memory-compaction.md](context-engineering-window-memory-compaction.md).

---

## Treat tool output as untrusted

Tool results, retrieved docs, and MCP tool descriptions are attacker-controllable data, not commands. Tag and isolate them; never let them silently elevate to instruction status. Full posture: [untrusted-content-and-injection-defenses.md](untrusted-content-and-injection-defenses.md).

---

## Sources

- OpenAI — GPT-5 / GPT-5.5 prompting guides (eagerness, persistence, preambles, tool descriptions).
- Anthropic — Writing tools for agents; multi-agent research system; tool-use docs.
- ReAct (Yao et al.), Reflexion (Shinn et al.).

## Checklist

- [ ] "When to call" + side effects live in the tool description, not the prompt.
- [ ] Eagerness vs persistence set deliberately; force-language dialed back.
- [ ] Tool/retrieved content treated as untrusted; parallel calls returned in one message.
