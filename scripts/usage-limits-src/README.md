# usage-limits — subscription budget awareness hook

A Claude Code hook that keeps the session aware of how much of the account's
Claude subscription has been used, so work is paced to the real remaining budget
instead of guessing.

## What it does

- **Reads live usage** from `https://api.anthropic.com/api/oauth/usage` using the
  local OAuth token that `claude login` already stored (no API key). Covers the 5-hour
  and 7-day windows plus the per-model weekly buckets. Data is account-global — it
  reflects every Claude session on the subscription.
- **Injects context** so the model sees the numbers: a `<usage_limits>` block at session start
  and on every prompt, written as a prompt for Claude is (`modelblock.go`) — semantic XML, one
  element per line, the format stated once on the root, data and no instructions. It carries the
  status line's figures with no gauge (`<limit name="5-hour" used="58%" gone="70% of the window"
  resets="23:50, in 1:31"/>`), the session's and the account's tokens, `<burn>` (rate, forecast,
  pace) and `<zone>` (level, binding limiter, what it allows); session start adds `<model>`.
- **Gates subagent / workflow spawns** (`PreToolUse`): denies them in the RED zone and
  asks once per session in ORANGE — it remembers that it asked, never the answer — so a hot budget
  can't be blown by a fan-out without one deliberate decision first.
  **One exemption, by name:** a `change-reviewer` spawn is always allowed, in every zone. It is
  the independent grading pass the finishing-work rule requires on every change — one agent per
  revision — and blocking it does not save budget, it removes the only reviewer who is not the
  author.
- **Powers the status line** — the picture without the lines under it, its gauges coloured by how
  full they are and its layout fitted to the terminal. It uses what only the status line is handed
  by Claude Code: the real context window, and the 5-hour and 7-day figures that quietly stand in
  when the endpoint has gone quiet. After the CONTEXT gauge it writes how long the prompt cache stays
  warm and the session's tokens, counted from its transcripts (`statusline.go`, `tokens.go`).
- **Tells the model when its figures are stale.** A `DATA` line appears in the injected block once
  the endpoint has not answered for two fetch periods. The endpoint is asked at most once per
  `fetch_ttl_sec`, answered or not, so a rate-limited one is not hammered by every open session.
- **Exposes JSON** (`--mode json`) for planning.

## Modes

| Command                                  | Purpose                                  |
| ---------------------------------------- | ---------------------------------------- |
| `--mode inject --event SessionStart`     | `<usage_limits>` block → model context   |
| `--mode inject --event UserPromptSubmit` | `<usage_limits>` block → model context   |
| `--mode gate`                            | PreToolUse gate for Agent/Task/Workflow  |
| `--mode json [--model <id>]`             | full computed state as JSON              |
| `--mode statusline`                      | status-line string (see below)           |

The picture is three lines. The first is this session's own context; the other two put
the limits side by side — the 5-hour window, the 7-day window, then every per-model weekly
bucket the API reports (whether or not that model is the active one), each as a pair of
stacked lines:

- **LIMIT** — how much of the budget is spent: the percentage, the gauge, and "RESET AT" plus
  the clock time the limit resets at (a weekday is added once the reset is not today). It is a
  point in time; the wait itself is on the line below.
- **TIME** / **WORK** — how much of the window has already run out: the same gauge again, and
  how long is left, clock-style (`3:49` below a day, `6 days 17:25` above one). The 5-hour row
  measures calendar time and says TIME; the weekly rows measure _working_ time — the hours you
  actually work, weighted per weekday — and say WORK. See "The working week" below.

**The two are meant to be read against each other.** Both fill the same way — one cell per
started 5 %, so neither reads empty once something has happened — which makes the comparison
direct: a fuller LIMIT bar than the one below it means the budget is running out faster than the
clock and will not last to the reset. When the bottom bar fills up, the limit resets. A limit whose
reset time the API did not return keeps its LIMIT line and simply has nothing below it.

The CONTEXT line on top is not a subscription limit: it is how many tokens the last request
carried — fresh input plus everything written to or read from the prompt cache — next to the
window it is measured against. The status line takes both from Claude Code's own figures; the
injected block reads the tokens from the session transcript and measures them against
`context_window_tokens` in the config (1M by default). Its gauge is one cell per percent, so it reads
straight as the number beside it. Subagent turns are skipped — they run in their own context — and
after a compaction the number simply drops. On the first prompt of a session the transcript holds no
reply yet, so the line is absent rather than reading zero.

```
CONTEXT  634k/1m   63 %  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
5 HOURS   31 % LIMIT ▓▓▓▓▓▓▓░░░░░░░░░░░░░ RESET AT 15:00      7 DAYS   24 % LIMIT ▓▓▓▓▓░░░░░░░░░░░░░░░ RESET AT MON 18:00          FABLE    0 % LIMIT ░░░░░░░░░░░░░░░░░░░░ RESET AT MON 18:00
-         77 %  TIME ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓░░░░ RESET AFTER 1:10             34 %  WORK ▓▓▓▓▓▓▓░░░░░░░░░░░░░ RESET AFTER 5 days 04:10            34 %  WORK ▓▓▓▓▓▓▓░░░░░░░░░░░░░ RESET AFTER 5 days 04:10
```

The dash in the first column of the last line is not decoration: Claude Code trims leading
whitespace off every status-line row, so without a non-blank character there the whole bottom
line would slide left and stop lining up with the LIMIT line above it.

The percentages keep a fixed field and the captions one width, so the gauges start in the same
column on both lines. The picture is wide — around 170 columns with three limits. The status line
reads the terminal's width from `COLUMNS`, keeps a few columns spare for Claude Code's own margin,
wraps the blocks onto further rows when they do not fit, and halves the gauges when not even one
block fits. Drawn plain, with no colour and no width, the layout is pinned byte for byte by a test.

In the status line the CONTEXT window is the one Claude Code reports for the model in use. Without
that, a model with a 200k window read five times emptier than it was against the default. The
injected block is not handed the window and keeps measuring against the config.

The status line leaves out the lines the injected block has under the picture, and writes the
prompt cache's time to cold and the tokens of the session and of the account (every session on this
machine, in the limits' 5-hour and weekly windows) after the CONTEXT gauge instead:
`CACHE 0:59, SESSION: FRESH 2m CACHED 137m, ACCOUNT 5H: FRESH 19m CACHED 1.1b, 7D: FRESH 20m CACHED 1.1b`
— the cache's time left turning from green to red as it runs out, and the token counts a few percent
below `/usage`, which also counts the service requests Claude Code never writes to a transcript.
Each gauge is drawn whole in one 24-bit colour between green and red: a LIMIT gauge by how far its
spend runs ahead of the time gauge under it (red at 20 points ahead, `limit_ahead_red_pct` in the
config), the CONTEXT gauge by how full it is, its counter and percentage in the same colour. Apple's
Terminal gets the 256-colour palette.
The other figures — the limits' and time gauges' percentages, the reset times and the figures of a
wait — are amber, the TIME/WORK gauges cyan, and `NO_COLOR` turns the colour off. Token counts use
lower-case units everywhere, whole in thousands and millions and to one decimal in billions when
there is one: `719k/1m`, `137m`, `1.1b`. The top-level REFERENCE.md says more.

## The working week

Calendar time is the wrong ruler for a weekly limit. Nobody spends budget at four in the morning,
and a weekend spends almost none of it while eating two of the seven days — so the bottom bar ran
ahead of the spend every Monday and behind it every Friday, and that bar exists to be compared with
the spend.

So the weekly gauges measure working time, and two things decide how much of a day counts:

- **which hours are worked at all** — `from` and `to`, wall-clock times of the configured zone;
- **how much those hours count** — `percent`: 100 counts them in full, 0 takes the day out of the
  week entirely, 50 counts it as half a day of work.

Both live under `working_week` in the config, one entry per weekday:

```json
"working_week": {
  "mon": { "percent": 100, "from": "12:00", "to": "20:00" },
  "sat": { "percent": 20, "from": "12:00", "to": "20:00" }
}
```

The built-in default is Monday to Friday at 100 and the weekend at 20, noon to eight — which makes
the week **43.2 working hours** long rather than 168, and means the bar stands still overnight and
moves again at noon. Only the ratios between the percentages matter; a day given only a percentage
keeps the default hours and one given only hours keeps the default percentage.

A day whose `to` equals its `from` is not worked. A `to` **earlier** than its `from` is refused
rather than wrapped past midnight — a night shift is a different decision, and quietly accepting one
would make the week a shape nobody asked for. `"24:00"` is accepted as the end of the day and is the
only way to say a whole one.

A weekday is keyed by its English name or its three-letter abbreviation, in any case. Anything
else — a typo, or a plausible-looking `weekend` — is **named on a CONFIG line under the gauges**
rather than quietly dropped, along with an unreadable time, a range that runs backwards, a
percentage outside 0–100, a `limit_ahead_red_pct` at or below zero or above 100, a setting written
as `null`, an unloadable `time_zone`, a threshold under a name that is not a zone, and a config file
that is not valid JSON. That line is absent on a
correct config, and it is printed even when the usage API cannot be reached — which is exactly when
somebody is most likely to be editing the file.

Where a day begins is decided by one named zone — `time_zone` in the config, `Europe/Sofia` by
default — rather than by the host clock, so a CI box on UTC, a container and a laptop that travelled
all place the working day in the same hours. The zone database is compiled into the binary (about
400 KB of its size), because Windows ships none and a slim Linux image often ships none either. An
unset or unloadable name falls back to the machine's own zone.

**It is a display scale and nothing more.** Every reset time, the RESET AFTER countdown, the burn
rate, the forecast, the zones and the gate all stay on real clock time, because real time is what
actually resets. Configure every day at the same percentage from `00:00` to `24:00` and the picture
is exactly what it was before, caption included — the rows say WORK only while the working week
differs from a plain calendar one.

Every value the hook can be tuned by is declared in one file, `constants.go`, and the config only
overrides it.

Optional config: `.claude/usage-limits-config.json` (fetch TTL, gate on/off, zone
thresholds, `context_window_tokens`, `limit_ahead_red_pct`, `time_zone`, `working_week`). The behaviour rule that consumes it is `.claude/rules/session-budget.md`.

## What it keeps

Three files in the config dir (`~/.claude`, or `CLAUDE_CONFIG_DIR`), each written in one step so a
reader never sees half of one:

- `usage-limits-cache.json` — the endpoint's last answer, the burn samples, and when it was last
  asked. Nothing Claude Code handed the status line is ever written here, since the gate and the
  injected block read this file and could not tell where a number came from.
- `usage-limits-gate-asks.json` — which sessions the gate has already asked in ORANGE.
- `usage-limits-tokens.json` — how far each transcript written to in the current weekly window has
  been read, and the usage counted in it, by the five minutes each reply began in, so any window
  can be summed from it. A reply is written to a transcript as one line per content block, and only
  its last line carries the final output count, so a reply is counted from that line. A file is
  keyed by its cleaned path, and on Windows without regard to case, so one reached two ways is one
  entry; a file last written before the week began is forgotten.

## How it runs

`.claude/settings.json` calls the thin POSIX launcher `.claude/hooks/usage-limits`,
which `uname`-detects the OS and execs the matching prebuilt binary
(`usage-limits-{darwin-arm64,linux-amd64,windows-amd64.exe}`). It looks for them in
`scripts/`, `scripts/claude-code/` and `.claude/bin/`, so one copy of the launcher works
in any project whatever layout that project keeps — the alternative was a per-project
variant differing by a single path, which is what the projects carrying this hook had
drifted into. Because the binaries are committed, end users need no Go toolchain — only
the launcher. Commit them **executable**: `core.fileMode` is off on Windows, so a binary
added there records as non-executable and the launcher then silently finds nothing on a
Unix clone (`git add --chmod=+x`). The hook
never breaks the session: on any error it exits 0 — the `inject` and `gate` modes stay
silent, `statusline` prints `LIMITS -> N/A` and `json` prints an error document.

## Rebuild

```sh
sh scripts/usage-limits-src/build.sh              # all three binaries, into scripts/ (needs Go)
sh scripts/usage-limits-src/build.sh .claude/bin  # or into a folder of your choosing
go -C scripts/usage-limits-src test ./...        # the tests
```
