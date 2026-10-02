# Claude Code hooks (Go)

- The Go hooks here (`usage-limits`, `work-audit`, `play-sound`) are the same code in all of the owner's projects: a change made in one is rolled out to every project that carries them, not only committed here.
- Rebuild with `sh <hook>-src/build.sh` from the folder that holds this file (needs Go). It writes darwin-arm64, linux-amd64 and windows-amd64 binaries; `usage-limits` also takes an output directory as its first argument.
- The launcher `.claude/hooks/<hook>` looks for the binaries in `scripts/`, `scripts/claude-code/` and `.claude/bin/`, so a project keeps them in whichever of the three it already uses.
- Commit the binaries executable (`git add --chmod=+x`): `core.fileMode` is off on Windows, and a non-executable binary is skipped silently by the launcher, so the hook does nothing on a Unix clone.
- The hooks are wired in `.claude/settings.json`.
