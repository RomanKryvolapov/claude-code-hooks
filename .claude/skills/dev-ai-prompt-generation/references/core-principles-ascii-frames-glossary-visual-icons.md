## About this file

**What’s inside:** ASCII frames for immutable rules, glossary blocks, and visual markers / emoji section labels.

**Hub:** [../SKILL.md](../SKILL.md)

---

## ASCII frames

**When to apply:** Immutable rules, anti–prompt-injection rules, or any block that must not be skipped.

**Why it works (heuristic):** A visual frame can make a must-not-skip block more salient — a heuristic, not a measured effect, and in tension with the 2026 guidance to dial back heavy emphasis on frontier models. Measure on your evals; prefer putting critical rules at a prompt edge.

### Frame characters

```
Double:  ╔ ╗ ╚ ╝ ═ ║ ╠ ╣ ╦ ╩ ╬
Single:  ┌ ┐ └ ┘ ─ │ ├ ┤ ┬ ┴ ┼
Heavy:   ┏ ┓ ┗ ┛ ━ ┃ ┣ ┫ ┳ ┻ ╋
```

### Frame template

```
╔══════════════════════════════════════╗
║  [HEADING]                            ║
╠══════════════════════════════════════╣
║  • [item 1]                           ║
║  • [item 2]                           ║
╚══════════════════════════════════════╝
```

---

## Glossary

**When to apply:** Domain terms, acronyms (USP, CTA, TA, TOV), or project-specific abbreviations used later in the prompt.

**Why it works:** Underspecified terms are a major source of inconsistency — vague or undefined terms make a prompt more likely to regress across model or prompt changes. A glossary at the start removes ambiguity, reduces response variability, and saves tokens when you reuse short forms. Background: [What Prompts Don't Say: Understanding and Managing Underspecification in LLM Prompts](https://arxiv.org/abs/2505.13360).

```
## GLOSSARY
H1 = main heading
USP = unique selling proposition
CTA = call to action
TA = target audience
TOV = tone of voice
MP = marketplace platform

## Task
Write H1 + USP + 3 CTA variants for TA "young moms 25-35". Platform: MP. TOV: friendly, no slang.
```

---

## Visual markers (icons)

**When to apply:** When the answer must be split into fixed categories (SWOT, code review, risk analysis).

**Why it works:** Fixed categories and forced choice improve adherence to structure; the effect comes from categorization, not the icon. Alternatives to emoji: `### RISKS`, `[RISKS]`, `**RISKS:**`.

```
Business plan:  💡 Innovations  🚩 Risks  ⚠️ Ambiguities  ✅ Strengths  ❌ Weaknesses  🎯 Recommendations
Code review:    🐛 Bugs  ⚡ Performance  🔒 Security  📖 Readability  ♻️ Refactoring
SWOT:           💪 Strengths  😰 Weaknesses  🌟 Opportunities  ⚠️ Threats
Content:        📌 Main idea  💬 Quotes  📊 Statistics  🔗 Sources
```

## Checklist

- [ ] ASCII frame treated as a salience heuristic, not relied on; critical rules also at a prompt edge.
- [ ] Domain terms/acronyms defined in a glossary before use.
- [ ] Fixed categories chosen for structure (the icon is optional).
