## About this file

**What’s inside:** Grounding as a contract (use only the context, say when unknown, cite every claim), native citations / quotes-first extraction, and the retrieval-quality essentials behind a RAG prompt. Relevant to knowledge-base chat and any answer that must be sourced.

**Hub:** [../SKILL.md](../SKILL.md)

---

## Grounding is a contract, not a hope

State it explicitly and enforce all three clauses:

```
- Answer ONLY from the provided context.
- If the context does not contain the answer, say so — do not use outside knowledge.
- Cite the source for every claim.
```

Reliability comes from making the contract auditable (citations) and from feeding good context (retrieval quality), below.

---

## Quotes-first and native citations

- **Quotes-first extraction:** for long/multi-document inputs, ask the model to pull relevant passages into `<quotes>` tags **before** answering, then answer from those quotes. Cuts noise and makes grounding visible.
- **Native citations:** vendors expose source-attribution features (e.g. Anthropic Citations) that return per-claim cited spans — prefer these where provenance matters (regulated/traceable domains). Note: Anthropic Citations is incompatible with structured-output schema enforcement, so choose per use case ([structured-outputs-constrained-decoding.md](structured-outputs-constrained-decoding.md)).

---

## Precision over volume

Flooding a reasoning model with retrieved text **dilutes** accuracy (context rot — [context-engineering-window-memory-compaction.md](context-engineering-window-memory-compaction.md)). Feed only the most relevant chunks; for long context, ask the model to outline key sections and anchor each claim to a section.

---

## Retrieval pipeline essentials

A grounding prompt is only as good as what’s retrieved. Naive single-vector top-k is insufficient in production:

```
query → hybrid retrieve (dense embeddings + sparse/BM25)
      → rerank candidates (cross-encoder) for precision
      → (structure-aware chunking upstream)
      → grounding / consistency check on the draft
      → low confidence ⇒ reformulate query / multi-hop / admit uncertainty
```

- **Hybrid search** catches both semantic and exact-keyword matches.
- **Reranking** lifts precision of the top chunks the model actually sees.
- **Grounding check:** verify generated claims are supported by retrieved text; refuse or re-retrieve when support is weak ("no evidence, no answer").

---

## Layout

Place documents near the **top**, the question/instructions at the **bottom**, bridge with a transition ("Based on the information above, …"). Keep behavioral constraints/persona in the system instruction. Long-context placement detail: [data-presentation-xml-mistakes-namespaces-sources-counting.md](data-presentation-xml-mistakes-namespaces-sources-counting.md).

---

## Untrusted by default

Retrieved chunks and user-supplied documents are attacker-controllable. Mark and isolate them; grounding is also a **security boundary**, not only attribution — see [untrusted-content-and-injection-defenses.md](untrusted-content-and-injection-defenses.md).

---

## Sources

- Anthropic — long-context grounding (quotes-first) and Citations docs.
- OpenAI / Google — limit RAG context to most-relevant chunks.
- 2025 RAG surveys; reranking and "no-evidence-no-answer" / grounding-check literature.

## Checklist

- [ ] Grounding contract enforced (only-context, say-when-unknown, cite-every-claim).
- [ ] Retrieval is hybrid + reranked; low confidence re-retrieves or admits uncertainty.
- [ ] Precision over volume; retrieved content treated as untrusted.
