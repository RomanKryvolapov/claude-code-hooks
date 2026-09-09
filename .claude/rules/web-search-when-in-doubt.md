# When in doubt — search the web, then ask (mandatory)

**When reality diverges from your knowledge — stop guessing, search the web.**
Your built-in knowledge is sometimes outdated or wrong, especially about the latest versions of tools, libraries, and APIs.

The known failure pattern to avoid: something doesn't work → you apply a fix based on what you "know" → it still fails → each new fix digs deeper into a solution built on a wrong assumption.

Therefore, in ALL doubtful cases — whenever anything is unclear, non-obvious, unpredictable, differs from your knowledge, or behaves differently than you expected — ALWAYS search the internet before attempting further fixes, paying special attention to the most recent version of the official documentation.

The official documentation for the exact version in use is the authority, and second-hand answers — a forum post, a blog, an index of somebody else's crawl, your own recollection — are only a way of finding it faster. Anything the work depends on — a breaking change, a supported version, a security-relevant flag, a deprecation — is confirmed against the vendor's own documentation before it is acted on.

If even the search gives no solution — STOP working and ask the user what to do.
Asking is always better than going down a false path — but only stop to ask after the search has failed, not instead of searching.

## Checklist

- [ ] Anything unclear / unexpected / version-sensitive → searched the web (latest official docs first) before further fixes.
- [ ] Did not stack fixes on an unverified assumption.
- [ ] Search failed → stopped and asked the user, instead of guessing down a false path.
