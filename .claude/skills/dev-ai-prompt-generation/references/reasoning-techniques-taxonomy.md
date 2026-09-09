## About this file

**What’s inside:** The reasoning-technique map — The Prompt Report’s six families, a one-line definition per named method, and a 2026 verdict (still useful, subsumed by native reasoning, or high-cost orchestration).

**Hub:** [../SKILL.md](../SKILL.md)

---

## The six families

Most named "techniques" are variants inside six families (Schulhoff et al., _The Prompt Report_, arXiv:2406.06608): **In-Context Learning**, **Zero-Shot**, **Thought Generation**, **Decomposition**, **Ensembling**, **Self-Criticism**. Pick by task shape, then check whether a reasoning model already does it internally.

> 2026 framing: on reasoning models, thought-generation and decomposition are largely **internal**. Reach for an explicit technique only when the task has structure the model cannot infer, or when you need an inspectable pipeline. Pair this file with [reasoning-models-thinking-effort-and-cot.md](reasoning-models-thinking-effort-and-cot.md).

---

## Techniques and 2026 verdicts

| Technique                                       | Family                   | What it does                                             | 2026 verdict                                                                                            |
| ----------------------------------------------- | ------------------------ | -------------------------------------------------------- | ------------------------------------------------------------------------------------------------------- |
| Few-shot / Few-shot CoT                         | In-Context Learning      | Worked input→output exemplars steer format and reasoning | Top lever for **format/tone** on non-reasoning models; can **degrade** reasoning models                 |
| Zero-shot CoT (`let’s think step by step`)      | Thought Generation       | Elicits intermediate reasoning with no exemplars         | **Subsumed** on reasoning models (redundant/harmful); useful on fast/legacy models                      |
| Self-Consistency                                | Ensembling               | Sample N reasoning paths, majority-vote the answer       | High-stakes single answers only; linear cost, diminishing returns on models that self-verify            |
| Tree-of-Thoughts (ToT)                          | Decomposition / search   | Branch, evaluate, backtrack over a thought tree          | Niche; search-structured puzzles; heavy orchestration                                                   |
| Graph-of-Thoughts (GoT)                         | Decomposition / search   | Generalize ToT to a graph (aggregate/merge thoughts)     | Research-grade; rarely worth the overhead                                                               |
| ReAct                                           | Agent / tool use         | Interleave reasoning with tool actions + observations    | **Core** agent loop, still central — see [agentic-tool-use-prompting.md](agentic-tool-use-prompting.md) |
| Reflexion                                       | Self-criticism / agent   | Reflect on failures, store critique, retry               | Useful **only** with a real success signal (tests, task outcome)                                        |
| Least-to-Most                                   | Decomposition            | Solve ordered subproblems, feed forward                  | Compositional generalization; often unneeded on reasoning models                                        |
| Step-Back / abstraction                         | Thought Generation       | Ask the governing principle first, then answer           | Helps knowledge/STEM questions                                                                          |
| Self-Ask                                        | Decomposition            | Model asks+answers sub-questions before answering        | Multi-hop QA; pairs with a search tool                                                                  |
| Plan-and-Solve                                  | Decomposition            | Devise a plan, then execute it                           | Mostly redundant on models with native planning                                                         |
| Self-Refine                                     | Self-criticism           | Draft → self-critique → revise, iterate                  | Open-ended generation; only if self-critique is genuinely useful                                        |
| Chain-of-Verification (CoVe)                    | Self-criticism           | Draft → verification questions → verified answer         | Good for hallucination-prone factual lists                                                              |
| Skeleton-of-Thought                             | Decomposition            | Outline first, expand points (in parallel)               | A **latency** trick, not an accuracy one                                                                |
| Analogical prompting                            | Thought Generation       | Model self-generates relevant exemplars                  | When you lack curated examples                                                                          |
| Program-Aided / Program-of-Thoughts             | Decomposition / tool use | Offload computation to executed code                     | Now delivered via **code-execution tools**                                                              |
| Emotion / politeness / tipping / expert-persona | Zero-shot framing        | Affective or role framing for "better" answers           | **Folklore** for accuracy — see Debunked below                                                          |

---

## Cost-vs-benefit notes

- **Decomposition + Ensembling** trade large inference/orchestration cost for accuracy; their marginal value **shrinks** on native reasoning models. Reserve for search-structured or genuinely high-stakes tasks.
- **Self-criticism** (Self-Refine, CoVe, Reflexion) only helps when the model can produce a useful critique or has an external success signal; weak self-evaluation adds cost. Self-evaluation can be unreliable — treat self-check as a cheap safety net, not a guarantee.
- **Prompt sensitivity is real and large:** equivalent reformattings of one prompt move accuracy materially (research reports ~10 points on average, larger on some model/task pairs); non-semantic features (separators, casing, option order) act as hidden control parameters. Mitigation: pick one convention and **measure** — see [evals-llm-as-judge-and-regression-gates.md](evals-llm-as-judge-and-regression-gates.md).

---

## Debunked / overrated on frontier models

- **Emotion prompting** ("this is important to my career"): small, inconsistent, model-dependent; no durable benefit. (It can _raise_ jailbreak success — a safety negative.)
- **Tipping / threats**: no statistically significant effect on output quality.
- **Politeness for accuracy**: at most a few points and inconsistent; one study found rude marginally beat very polite. Politeness is a UX/ethics choice, not a performance lever.
- **Expert-persona for accuracy** ("you are a world-class expert"): often net-negative for correctness; personas steer **tone/format**, not accuracy.

---

## Sources

- _The Prompt Report_ — Schulhoff et al., arXiv:2406.06608.
- DAIR.ai Prompt Engineering Guide — promptingguide.ai.
- Origin papers: ReAct (Yao et al.), Reflexion (Shinn et al.), Self-Refine, Chain-of-Verification (Dhuliawala et al.), Tree-of-Thoughts (Yao et al.), Step-Back (DeepMind).

## Checklist

- [ ] Technique chosen by task shape; checked whether the model already does it internally.
- [ ] Decomposition/ensembling reserved for search-structured or high-stakes work.
- [ ] Folklore framings (emotion/tipping/persona for accuracy) not used.
