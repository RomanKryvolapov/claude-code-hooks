## About this file

**What’s inside:** JSON as structured input, and comparison anchors / benchmarks in JSON.

**Hub:** [../SKILL.md](../SKILL.md)

---

## JSON structuring of input data

**When to apply:** Many related attributes; nested structures; lists of same-type items; data from an API/DB.

**Why it works:** JSON keeps related facts "neighbors" in context, which helps extraction and reasoning on long inputs. For complex output formats (XML, BPMN), emit a simple intermediate JSON and convert in code rather than generating the complex format directly — better still, enforce that JSON with structured outputs ([structured-outputs-constrained-decoding.md](structured-outputs-constrained-decoding.md)).

- One pair of braces `{}` per object; extra braces (`{{{{...}}}}`) add tokens and confusion.
- A JSON schema in the prompt describes structure; it is not a multi-brace template.
- Use JSON for many related attributes, nested structures, lists of same-type items, and API/DB data.

### Plain text vs JSON

**❌ Plain text:** "Name: Alex Smith, age 34, city New York, position Senior Developer at TechCorp, salary 350000, married, two kids…" — attributes scattered.

**✅ JSON:** related attributes grouped as neighbors.

```
{
  "profile": {"name": "Alex Smith", "age": 34, "city": "New York"},
  "work": {"position": "Senior Developer", "company": "TechCorp", "salary": 350000},
  "family": {"status": "married", "children": 2},
  "interests": ["skiing", "programming"],
  "purchases": [
    {"date": "2026-01-15", "item": "MacBook Pro", "price": 250000}
  ]
}
```

### Reference points as comparison anchors

Adding reference data (market averages, benchmarks, history) lets the model draw sharper comparative conclusions:

```
{
  "current": {"revenue": 1200000, "margin": 15},
  "benchmarks": {
    "industry_avg": {"revenue": 800000, "margin": 12},
    "top_10_percent": {"revenue": 2500000, "margin": 22}
  },
  "history": [{"year": 2025, "revenue": 1100000, "margin": 14}]
}
```

## Checklist

- [ ] Related attributes grouped in JSON; one {} per object; no excess braces.
- [ ] Machine-consumed JSON enforced via structured outputs, not described in prose.
- [ ] Reference anchors added when a comparison is needed.
