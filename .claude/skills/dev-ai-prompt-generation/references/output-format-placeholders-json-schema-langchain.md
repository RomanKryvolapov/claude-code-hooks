## About this file

**What’s inside:** Placeholders `{var}` vs `[model fills]`, JSON schema in prompts, LangChain `{{` escaping, and the intermediate-JSON strategy for complex formats.

**Hub:** [../SKILL.md](../SKILL.md)

---

## Placeholders

Two kinds: the model fills them, or you supply them.

| Syntax              | Who fills     | Use                 | Example                       |
| ------------------- | ------------- | ------------------- | ----------------------------- |
| `[text]`            | model         | generation template | Hello, `[client_name]`!       |
| `{variable}`        | you           | your data slot      | Analyze: `{text}`             |
| `[a/b/c]`           | model chooses | limited choice      | `[positive/negative/neutral]` |
| `[1-5]`             | model         | numeric range       | Rating: `[1-5]`               |
| `[up to 100 words]` | model         | length limit        | Summary: `[up to 100 words]`  |

```
## Input (you fill)
Product: {product_name}
Price: {price}

## Output format (model fills)
# [Sales headline — up to 60 chars]
[Description — 2-3 sentences]
[Call to action — 1 sentence]
```

---

## JSON schema in the prompt

> For **machine-consumed** output the real contract is **native structured outputs / constrained decoding**, which guarantees schema-valid JSON — see [structured-outputs-constrained-decoding.md](structured-outputs-constrained-decoding.md). A schema described in the prompt is a **fallback** for chat/UX text or models without the feature; it steers but does not guarantee — validate the result. Assistant-prefill tricks to force JSON are discouraged where native structured outputs exist; verify prefill behavior against the current API docs.

Describe the contract with types and allowed values:

```
Return JSON:
{
  "sentiment": "positive" | "negative" | "neutral",
  "confidence": number 0.0–1.0,
  "keywords": array of strings (max 5),
  "issues": array of strings | null
}
Rules: sentiment is one of the three; issues is array OR null, never empty [].
```

Use one pair of braces `{}` per object; union types via `|`; nest without excess braces.

**LangChain / f-string escaping:** in LangChain templates `{}` are template variables — escape literal JSON braces by doubling (`{{` → `{`). In a Python f-string, double again: `{{{{"key": "v"}}}}` → the model sees `{"key": "v"}`. Source: [INVALID_PROMPT_INPUT](https://docs.langchain.com/oss/python/langchain/errors/INVALID_PROMPT_INPUT).

---

## Intermediate JSON for complex formats

LLMs are unreliable at generating complex nested formats directly (XML, HTML, BPMN). Ask for a simple intermediate JSON, then convert in code.

- **❌** "Generate a BPMN diagram in XML."
- **✅** "Return JSON `{nodes:[{id,label,type}], edges:[{from,to}]}`" → convert to XML in code.

Enforce that intermediate JSON with structured outputs / constrained decoding ([structured-outputs-constrained-decoding.md](structured-outputs-constrained-decoding.md)) rather than prose instructions.

## Checklist

- [ ] {var} (you fill) vs [model fills] distinguished.
- [ ] Machine-consumed output enforced via structured outputs; prefill not used to force shape.
- [ ] LangChain/f-string braces escaped where applicable; complex formats via simple intermediate JSON.
