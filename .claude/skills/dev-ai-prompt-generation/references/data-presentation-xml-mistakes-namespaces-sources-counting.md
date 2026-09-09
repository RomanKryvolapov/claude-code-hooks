## About this file

**What’s inside:** Common XML mistakes, namespaces, source marking (numbered docs), attribution/grounding, two-stage prompts, and System-2 counting with chunking.

**Hub:** [../SKILL.md](../SKILL.md)

---

## Common XML mistakes

- Meaningless tags (`<xyz>`, `<aaa>`, `<block1>`) are ignored — only semantic tags work (`<context>`, `<draft>`, `<solution>`).
- Every opening `<tag>` needs a closing `</tag>`.
- Avoid 3+ levels of nesting; prefer JSON inside one tag over 5 levels of XML.
- With 5+ tags, prefix them (`input:`, `rules:`, `output:`) to group by function.

---

## Source marking

**When to apply:** Multiple documents or RAG; answers must be grounded; attribution required.

**Why it works:** Structured context and explicit source marking improve attribution and reduce hallucinations; two-stage "first quote the relevant passage, then answer" improves document adherence. Numbered blocks tell the model each document's boundaries and the total count, so it does not mix contexts or misattribute. Source marking is also a **security boundary** — mark untrusted spans as data, not only for attribution. See [rag-grounding-and-citations.md](rag-grounding-and-citations.md) and [untrusted-content-and-injection-defenses.md](untrusted-content-and-injection-defenses.md).

### Numbered documents

```
[DOCUMENT 1 OF 3]
first document text
[END DOCUMENT 1]
[DOCUMENT 2 OF 3]
second document text
[END DOCUMENT 2]
```

`[DOCUMENT N OF M] ... [END DOCUMENT N]` gives the total count and each document's boundaries.

### Attributes and attribution

```
<doc id="smith" author="John Smith" source="Forbes 2024">
"AI market will triple by 2027."
</doc>
<doc id="jones" author="Jane Jones" source="Analytics Report">
"Do not overestimate AI growth."
</doc>
```

When citing, always use the document id `[doc_id]`; format `"quote" — [author, source]`; never attribute one document's words to another author.

### Context grounding

Constrain the model to the source:

```
<context source="Report Q3 2024">[document text]</context>

RULES:
- Use ONLY information from <context>
- Do not add external knowledge
- If not in context, write: "Data not in document"
```

Two-stage variant: ask the model to quote the relevant passage first, then answer from it — this reduces hallucinations.

---

## System-2 counting

**When to apply:** Counting words, items, or occurrences in long text. Accuracy drops to near zero when counting "in head".

**Why it works:** Counting reliability degrades as the list grows. Split the text, count each chunk separately, write the partial counts as text, then sum — the model must "see" the partial numbers in its own output to add them. Keep this qualitative; the exact accuracy / element-count figures once quoted here are not robust. In 2026 the correct fix for an exact count is a **code-execution tool**, not prompt chunking — use this only when no tool is available. Use small chunks (~5–10 elements each).

```
Text below is split by │. Count the word "type" in each part, write each partial count, then sum.

Part 1: [n]
Part 2: [n]
Total: [sum]

Text: [first 10 sentences] │ [next 10] │ [next 10]
```

## Checklist

- [ ] Only semantic tags; every tag closed; nesting <= 3 levels.
- [ ] Multiple sources numbered and attributed; grounding stated; untrusted spans marked as data.
- [ ] Exact counts done with a code-execution tool, not prompt chunking.
