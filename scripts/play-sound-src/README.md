# play-sound — notification sound hook

A small Claude Code hook that plays a short notification sound when Claude finishes a
turn or a turn fails, so you notice when it needs you or is done.

## What it does

Plays the bundled `.claude/sounds/notify.wav` whenever the session wants the developer back —
on `Stop` and `StopFailure` (the turn finished or died) and on `Notification` (Claude is waiting:
a question, a permission prompt, an idle session). Cross-platform:

- **Windows** — the built-in .NET `SoundPlayer` (via PowerShell), no extra tools.
- **macOS** — `afplay`.
- **Linux** — the first available of `pw-play`, `paplay`, `aplay`, `play`, `ffplay`,
  `canberra-gtk-play`.

The launcher passes the wav path as the first argument. It is best-effort: a missing
player or absent audio device never breaks the session — any failure is swallowed and
the process exits 0.

## How it runs

`.claude/settings.json` calls the POSIX launcher `.claude/hooks/play-sound` on `Stop`,
`StopFailure` and `Notification`. The launcher `uname`-detects the OS and execs the matching prebuilt
binary committed under `scripts/` (`play-sound-{darwin-arm64,linux-amd64,windows-amd64.exe}`),
so end users need no Go toolchain. The sound asset stays at `.claude/sounds/notify.wav`.

## Rebuild

```sh
sh scripts/play-sound-src/build.sh   # cross-compiles all three binaries (needs Go)
```
