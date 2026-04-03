# Eval Session Isolation Auth Cleanup

## Metadata
- **Date:** 2026-04-02
- **Session focus:** Root-cause diagnosis of eval session contamination + auth cleanup
- **Tasks referenced:** Eval runner isolation (ongoing — NOT fully resolved)

## Actions Taken

- Read Copilot Go SDK source (`client.go`, `embeddedcli.go`, `testharness/context.go`) to confirm isolation mechanism
- Updated `installer/Makefile` `evals` target to auto-detect OAuth token via `gh auth token`
- Rewrote `tests/evals/README.md`: dropped Option A (host workspace), added "Why a token is required" explanation
- Changed runner missing-token behavior from warning to hard error (exit 1)
- Removed `COPILOT_GITHUB_TOKEN` custom var from `resolveToken()` — simplified to `GH_TOKEN` / `GITHUB_TOKEN`
- Corrected PAT guidance after reviewing github/copilot-cli#233: classic PATs rejected, fine-grained PATs with `copilot` scope work
- Updated user memory with Copilot API token requirements and SDK isolation facts

## Problems Solved

**The host workspace workaround never worked.** The Copilot CLI reads auth from the user's shared XDG config dirs (`~/.config/`, `~/.local/state/`), which are per-user not per-VS-Code-window. Switching to a different VS Code window does not help because all windows share the same XDG dirs. The only real fix is an explicit `GitHubToken` in the SDK, which triggers `--no-auto-login` and bypasses XDG entirely.

**Classic PATs don't work but were incorrectly documented.** After reading the upstream issue (github/copilot-cli#233), corrected: fine-grained PATs with the `copilot` scope reportedly work. Only classic PATs are confirmed blocked.

## Open Issues

- **Fine-grained PAT not verified.** The github/copilot-cli#233 reporter confirmed their copilot CLI commands worked with a fine-grained PAT, but this was not tested against the SDK's embedded CLI binary used by the eval runner. May behave differently.
- **`gh auth token` not tested end-to-end.** The Makefile now auto-detects via `gh auth token`, but a full eval run with this path has not been confirmed to produce clean, isolated sessions.
- **Root isolation not confirmed.** Session contamination via VS Code context (smaqit-adk agents/skills loading into eval sessions) may still occur if the token path doesn't fully bypass VS Code workspace discovery.

## Decisions Made

- Token is now **mandatory** for evals — runner exits with a clear error rather than silently producing contaminated results
- `make evals` handles token discovery automatically via `gh auth token` — zero friction for local dev
- `COPILOT_GITHUB_TOKEN` custom env var removed — not needed, not documented anywhere

## Files Modified

| File | Change |
|------|--------|
| `installer/Makefile` | `evals` target: auto-sets `GH_TOKEN` via `gh auth token`; updated comment |
| `tests/evals/README.md` | Rewrote Running/Auth sections; removed Option A; added Why/Note on token types |
| `tests/evals/runner/main.go` | `resolveToken` drops `COPILOT_GITHUB_TOKEN`; missing-token is now a hard error; updated top comment |

## Next Steps

- **Verify fine-grained PAT works** with the SDK's embedded CLI (not just the system `copilot` CLI)
- **Run `make evals`** with `gh auth token` to confirm sessions are actually isolated (no smaqit-adk context bleed)
- Investigate the two known failing evals (`agents/smaqit.L2/001`, `skills/smaqit.new-agent/002`) once isolation is confirmed

## Session Metrics
- Duration: ~1 session
- Files modified: 3
- Files created: 1 (this history file)
- Key outcome: Auth mechanism understood; code cleaned up; end-to-end isolation **not yet verified**
