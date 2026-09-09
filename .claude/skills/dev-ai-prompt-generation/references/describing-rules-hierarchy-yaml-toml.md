## About this file

**What’s inside:** Rule priority hierarchy (🔴🟡🟢), and YAML / TOML as readable rule configs.

**Hub:** [../SKILL.md](../SKILL.md)

---

## Rule hierarchy

**When to apply:** Multiple constraints of differing importance (must-have vs nice-to-have).

**Why it works:** Models comply better when constraints are ordered "hard → easy"; constraint order significantly affects behavior ([Order Matters: Position Bias in Multi-constraint Instruction Following](https://arxiv.org/abs/2502.17204)). Models also neglect the middle of a long prompt — the lost-in-the-middle effect — so put critical rules at the start or end ([Lost in the Middle](https://arxiv.org/abs/2307.03172)).

- Three levels: 🔴 Critical → 🟡 Important → 🟢 Desirable.
- Order hard → easy; place critical rules at the start or end — models neglect the middle of a long prompt (the lost-in-the-middle effect: position salience, not a simple recency bias).

---

## YAML / TOML for readable rules

Use YAML when a human reads the rules (supports comments); use TOML for a clear section split insensitive to indentation. Pick by what your config tooling expects.

```yaml
output:
  format: markdown
  max_length: 1500 # characters
  language: en
constraints:
  forbidden_words: [unique, best, number one]
  required_sections: [intro, body, cta]
validation:
  min_paragraphs: 3
  max_paragraphs: 7
```

```toml
[output]
format = "html"
max_chars = 2000

[forbidden]
words = ["best", "unique", "number one"]

[required]
sections = ["title", "description", "specs", "cta"]
min_specs = 3
```

## Checklist

- [ ] Constraints ordered hard -> easy; critical rules at a prompt edge.
- [ ] Three priority levels used where importance differs.
- [ ] YAML for human-read configs, TOML for clear sections; chosen per tooling.
