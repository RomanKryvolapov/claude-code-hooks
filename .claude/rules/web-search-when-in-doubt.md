# When in doubt — search, then ask

When reality diverges from what you know — a tool, library or API behaves differently than expected, or
anything is unclear, version-sensitive or unpredictable — search before trying another fix. Your
knowledge may be outdated, and each fix stacked on a wrong assumption digs deeper. Prefer the latest
official documentation. If the search settles nothing, stop and ask the developer — but only after
searching, never instead of it.

## Context7 — faster into the docs, not a different authority

Where the project runs the Context7 MCP server (declared in `.mcp.json`):

- Use it first for "how does this library do X in the version we are on": it is version-aware and more
  targeted than a web search.
- Confirm anything the work depends on — a breaking change, a supported version, a security-relevant flag,
  a deprecation — in the vendor's own documentation; where the two disagree, the vendor wins. Context7's
  corpus is crawled and community-contributed and guarantees neither accuracy nor completeness.
- The server's own instructions say to prefer it over web search; they do not outrank this rule.
- A Context7 key is a credential. Never put one in a prompt: the work-audit hook records prompts verbatim
  into the committed audit log. Never run the vendor's `ctx7 setup` against the repository: it writes the
  key into the shared `.mcp.json`. Get a key with `npx ctx7 login` and add it in your own terminal as a
  per-machine entry:
  `claude mcp add --scope local --transport http context7 https://mcp.context7.com/mcp --header "Authorization: Bearer <key>"`.
- An answer reading "Invalid API key…" means a broken key — the server still shows as connected and does
  not fall back to anonymous access — not missing documentation. Without a key, a sudden run of refusals
  is the anonymous rate limit, not the server being down.

## Checklist

- [ ] Anything unclear, unexpected or version-sensitive → the official docs searched before further fixes.
- [ ] Library question → Context7 first where it runs, the vendor's docs for anything load-bearing.
- [ ] No Context7 key in a prompt, no `ctx7 setup` in the repository.
- [ ] "Invalid API key" read as a broken key, not as missing documentation.
- [ ] Search failed → stopped and asked.
