## About this file

**What’s inside:** XML tag semantics (role/context/task/rules), nesting, attributes, namespaces, and the universal XML prompt template.

**Hub:** [../SKILL.md](../SKILL.md)

---

## XML tags and semantics

**When to apply:** Prompts with three or more logical blocks (context, task, rules, output, examples), **when targeting Claude** — XML tags are Claude's primitive. If the target model is not Claude (OpenAI, Gemini), use Markdown sections instead (see vendor portability). Use semantic tags in English — fewer tokens, models expect them.

**Why it works:** Tagged structure improves parsing of instructions and reduces errors; combining role + task + output + examples improves consistency on structured tasks (measure on your evals — no fixed percentage). Tags mark clear boundaries: where context, where task, where rules. Anthropic recommends XML tags as Claude’s preferred structuring primitive — [Use XML tags to structure your prompts](https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/use-xml-tags).

> **Vendor portability:** XML tags are Claude’s preferred primitive. OpenAI and Gemini work equally well with **Markdown sections** and put app rules in the **developer** message; Mistral/Grok accept either. Use one convention per prompt. Durable product rules go at system/developer level, never the user turn; DeepSeek-R1 is the exception (no system prompt, all in the user turn). See [vendor-portability-cross-model-conventions.md](vendor-portability-cross-model-conventions.md).

### Basic tags

| Component | Main tag     | Alternatives            | When to use                      |
| --------- | ------------ | ----------------------- | -------------------------------- |
| Role      | `<role>`     | `<persona>`             | at prompt start                  |
| Context   | `<context>`  | `<background>`          | when background is needed        |
| Task      | `<task>`     | `<objective>`, `<goal>` | required — core of prompt        |
| Rules     | `<rules>`    | `<constraints>`         | constraints and quality criteria |
| Output    | `<output>`   | `<format>`              | when output format matters       |
| Example   | `<example>`  | `<sample>`              | complex formats, few-shot        |
| Input     | `<input>`    | `<data>`                | input data to process            |
| Document  | `<document>` | `<doc>`, `<source>`     | cited sources                    |

Minimum prompt = 3 tags: `<context>` + `<task>` + `<rules>`. English tags cost ~1 token vs 2–3 for other languages and are the standard.

### Nested tags and attributes

Nesting creates hierarchy (2–3 levels is enough). Attributes carry metadata: `source="..."` (origin), `id="..."` (attribution), `lang="..."` (language), `do_not_copy_language="true"` (avoid style drift).

```
<document>
    <metadata>
        <title>Report Q3 2024</title>
        <author>Analytics team</author>
    </metadata>
    <content>Revenue grew 15%...</content>
</document>
```

### Namespaces for grouping

With 5+ tags, prefix them to group by function so the model sees the hierarchy:

```
<input:article>Article text...</input:article>
<input:comments>User comments...</input:comments>
<rules:content>Use ONLY facts from input:article</rules:content>
<output:format>JSON: {"summary": "...", "facts": [...]}</output:format>
```

### Universal XML prompt template

```
<role>You are [specialization]. Your task is [main function].</role>
<context>[Background, situation, constraints]</context>
<task>[Concrete task: what exactly to do]</task>
<rules>
- [Constraint 1]
- [Output format]
</rules>
<output>[Expected output structure]</output>
<example>
Input: [example input]
Output: [example correct answer]
</example>
```

## Checklist

- [ ] One convention per prompt (XML tags OR Markdown sections), never a heading paired with a same-named tag.
- [ ] Semantic tags only; all closed; nesting <= 3 levels; 5+ tags grouped by prefix.
- [ ] Tags in English; vendor fit checked (XML for Claude; Markdown + developer message for OpenAI/Gemini).
