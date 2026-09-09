## About this file

**What’s inside:** Table-filter task example, token separation, number canonicalization, CAPS/emphasis and quotes for exact values, and 2026 force-language / positive-phrasing guidance.

**Hub:** [../SKILL.md](../SKILL.md)

---

## Example: task with a table filter

```
| Service | Price ($) | Quality | Family |
|---------|-----------|---------|--------|
| Netflix | 16        | 4K HDR  | Yes    |
| Hulu    | 12        | 1080p   | Yes    |
| Disney+ | 18        | 4K HDR  | No     |
| HBO Max | 15        | 4K HDR  | Yes    |

Find services where price ≤ $15, family = Yes, quality = 4K.
```

---

## Token separation

**When to apply:** Lists or enumerations; tabular records (name, age, city); counting or arithmetic over many elements.

**Why it works:** Tokenization affects arithmetic and symbolic reasoning: "glued" elements can merge into one token and get lost, so separating them (commas between numbers, one item per line) helps on text-only paths. Keep the gain qualitative — it is task-dependent. In 2026 the correct fix for exact counting or arithmetic is a **code-execution tool**, not prompt formatting. Background: [Tokenization counts: impact on arithmetic in frontier LLMs](https://arxiv.org/abs/2402.14903).

### Separation symbols

| Symbol       | Example                      | When to use                 |
| ------------ | ---------------------------- | --------------------------- |
| `,` comma    | `a, b, c, d`                 | element lists, enumerations |
| `\|` pipe    | `name \| age \| city`        | tabular records             |
| `\n` newline | each element on its own line | long lists                  |
| `---` line   | Block 1 --- Block 2          | semantic boundaries         |
| space        | `s t r a w b e r r y`        | character-level analysis    |

### Glued vs separated

**❌ Glued:** `Analyze: apple,pear,banana,orange` — the tokenizer may merge words and lose elements.

**✅ Separated:**

```
Analyze:
- apple
- pear
- banana
```

Each element is its own token → accurate count. Records read cleanest one per line with pipes: `John | 25 | Moscow`.

### Number canonicalization

Use one number format and state it ("All numeric values are in format X"):

| Task         | Format                          | Example                                           |
| ------------ | ------------------------------- | ------------------------------------------------- |
| Simple tasks | no separators                   | `1250` instead of `1,250.00`                      |
| Arithmetic   | digit separation or a code tool | `1 234 567`; for exact math prefer code execution |

---

## CAPS and emphasis

**When to apply:** A single non-negotiable prohibition (e.g. `NEVER use the word "unique"`). Place at the start or end of the prompt, not the middle.

**Why it works:** Over-using CAPS dilutes the signal — if everything is emphasized, nothing stands out, so reserve it for one critical prohibition. Treat this as a formatting heuristic and measure on your own evals; on frontier models, dialing back force-language matters more than emphasis placement (see _Force-language and positive phrasing_ below).

**❌ Everything shouted:**

```
NEVER use the WORD "unique".
ALWAYS write IN ENGLISH.
MUST add CTA.
```

**✅ One prohibition stands out:**

```
- Write in English
- Add a call to action
- Length: up to 500 characters
- NEVER use the word "unique"
```

### Placement

- Place at the start or end of the prompt, not the middle.
- The middle of long context is used worst — the lost-in-the-middle (not recency) effect; models attend best to the edges. Study: [Lost in the Middle](https://arxiv.org/abs/2307.03172).

### Quotes for exact values

Put a forbidden word in quotes so the model treats it literally:

| Format                         | Result                            |
| ------------------------------ | --------------------------------- |
| `Do not use the word unique`   | model may interpret loosely       |
| `Do not use the word "unique"` | model reads it as the exact value |

---

## Force-language and positive phrasing (2026)

**When to apply:** Any prompt for a frontier model (Claude Opus/Sonnet 4.x and later, GPT-5/o-series, Gemini 3, Grok, R1).

**Why it works:** Frontier models follow instructions literally and act eagerly. Aggressive emphasis carried over from older models now **over-triggers**: `CRITICAL`, `You MUST`, `ALWAYS`, `if in doubt, use the tool` cause over-eager tool calls and over-exploration. The lever is no longer "how loud" but "how clear".

- Dial back force-language to plain conditional phrasing: "Use X when …", not "CRITICAL: You MUST use X".
- Prefer positive instructions over prohibitions — a "do not" list is weaker than one positive style example.
- Explain the WHY behind a constraint so the model generalizes it to unforeseen cases.
- State scope explicitly — literal-following models do not infer "apply this to every item".
- Remove inherited anti-laziness / "be thorough" scaffolding; it makes proactive models over-act.

See the failure-mode inversion in [reasoning-models-thinking-effort-and-cot.md](reasoning-models-thinking-effort-and-cot.md) and tool over-triggering in [agentic-tool-use-prompting.md](agentic-tool-use-prompting.md).

## Checklist

- [ ] Elements separated so they do not merge into one token; exact counts use a code tool.
- [ ] At most one CAPS prohibition, at a prompt edge, exact value in quotes.
- [ ] Force-language dialed back; positive phrasing; scope stated explicitly.
