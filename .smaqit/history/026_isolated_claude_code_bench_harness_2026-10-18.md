# Isolated Claude Code Bench Harness

**Date:** 2026-10-18
**Session focus:** Plan, implement, live-verify and release a first-class Claude Code harness for Bench. Along the way, move the repository back from the unreachable `daisy-tribe` org to the public `ruifrvaz/smaqit-adk`, and upgrade the machine's global install.
**Tasks completed/referenced:**
- 037 completed (PR #25 on ruifrvaz, released as v3.1.0)
- 035 abandoned, folded into 037
- 036 referenced: its PR #3 is stranded on daisy-tribe
- 031 and 033 referenced as related only

## Actions Taken

- Started the session and reviewed open tasks (037, 036, 035, 033, 031, 025). Found that `gh` and `git fetch` both failed against `daisy-tribe/smaqit-adk`.
- Planned Task 037 via `smaqit.task-plan`, using three parallel discovery agents:
  - Bench engine internals
  - Claude Code's headless contract (flags, `CLAUDE_CONFIG_DIR`, auth precedence, discovery precedence)
  - Existing Codex conventions and the overlap with Tasks 031, 033 and 035
- Resolved four design forks with the user: a new adapter, an isolated config dir, a single credential env var, and folding Task 035 in. Multi-turn scripted Cases became a follow-up.
- The user rejected storing the live-run credential in Vault in favour of a direct inline loader. Confirmed that `~/.claude/.credentials.json` holds the login access token (plus MCP OAuth tokens, so the loader extracts one field only).
- Updated the task file (decisions, discovery findings, 14 implementation steps, 4 criteria folded in from Task 035) and started the task. Steps covered: worktree, research-map refresh, and issue triage (Advisory: several `CLAUDE_CONFIG_DIR` isolation bugs upstream, mitigated by also isolating `HOME`).
- Ran a live probe (one Haiku turn, $0.0014). It confirmed that a login access token works as `CLAUDE_CODE_OAUTH_TOKEN` in an empty config dir, and that `system/init` lists the loaded skills and agents.
- Implemented Phase A (engine safety), Phase B (the `claude-code` adapter) and Phase C (docs, skills, live scenario). Added 26 new unit tests against a fake `claude` script and real sockets and FIFOs.
- Live-verified `.smaqit/bench/scenarios/claude-code-isolation`:
  - Its treatment skill is deliberately named after an installed personal skill.
  - with-treatment won; the baseline saw no `smaqit.*` skills.
  - No token appeared in the evidence and no `/tmp` leftovers remained.
- Found that both bench skill definitions had been stale since Task 028 (the compiled skills were hand-edited in Tasks 029 and 030). Refreshed them, then recompiled both via two parallel `smaqit.L2` subagents to v0.3.0.
- During `task.complete`, the user redirected the repo back to `ruifrvaz/smaqit-adk`:
  - repointed `origin`, keeping a `daisy-tribe` remote
  - merged the user's Aug 30 README web edit and pushed local `main` once, with no tags
  - reverted the code identity (Go module paths, imports, CLI help, `install.sh`, docs) inside 037's branch
- Opened PR #25 (`Prepare release v3.1.0`) on ruifrvaz. The user merged it; the release workflow published `adk-v3.1.0` with all binaries and `install.sh`. Ran Phase 2 cleanup.
- Upgraded the global install from v2.0.1 to v3.1.0 via the install one-liner and validated it:
  - skills, agents, framework and templates match the repo byte-for-byte
  - the change was additive only
  - Bench validate/plan works, and a second live isolation run with the installed binary passed

## Problems Solved

- **Personal skills shadow treatments.** Claude Code gives personal `~/.claude/skills` precedence over project skills, and about 60 `smaqit.*` skills are installed globally. The adapter isolates both `HOME` and `CLAUDE_CONFIG_DIR` per attempt in a harness home outside the workspace. It fails on same-named definitions in the fixture's or an ancestor's `.claude/` or managed locations, and asserts discovery through `system/init`.
- **The invoking session's own environment would leak into the harness.** This session's `CLAUDE_CODE_SESSION_ID` and messaging socket/token were visible in Bash. The adapter builds its environment from nothing, and a test proves those variables don't pass through.
- **A hidden credential leak.** The registry-cascade manifest copied `~/.codex` into `{workspace}`. Freeze only excluded the sidecar, so fixing Task 035's socket crash alone would have frozen Codex credentials into evidence. The credentials now live in `{harnessHome}` via `CODEX_HOME`.
- **FIFOs would hang the old freeze step.** Opening a FIFO blocked forever, rather than crashing like a socket did. Special files are now recorded as `special` and excluded from submissions.
- **Silently swallowed cleanup failures.** These are now always reported in `result.json`, `report.md`, the events and the CLI.
- **Stale skill definitions.** Recompiling from the stale definitions would have reverted Tasks 029 and 030. The definitions were reconciled with the compiled content first.
- **Unreachable GitHub repo.** `daisy-tribe/smaqit-adk` returned 404 and the org was no longer visible. This was resolved by moving back to the user's public repo.
- **`--bare` can't provide isolation.** It disables skill/agent/command discovery, so isolation relies on `CLAUDE_CONFIG_DIR` plus `HOME` instead.

## Decisions Made

- `adapter: claude-code` is a new runtime adapter, not a plan-time preset on `process`. Schema versions stay at 2, since the change is additive.
- Treatments are installed into an isolated per-attempt config dir, never into the workspace's `.claude/`.
- Exactly one credential env var (`CLAUDE_CODE_OAUTH_TOKEN` or `ANTHROPIC_API_KEY`) passes through; no credentials file is ever copied. A post-attempt canary scan redacts any leaked value and fails the attempt.
- The live-run credential comes from a direct inline `jq` loader of the user's own login, never from Vault (user decision).
- Only `DISABLE_AUTOUPDATER=1` is set by default. `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC` and `DISABLE_TELEMETRY` are deliberately not set, because both also disable feature-flag fetching (a deviation from the plan).
- Task 035 was folded into 037; multi-turn resume Cases are a follow-up.
- Repository decisions (user):
  - push local `main` directly to the public ruifrvaz repo
  - fold the identity revert into 037
  - keep `install.sh` gh-only
  - don't backfill the v2.1.0 and v3.0.0 releases on ruifrvaz
- Task 036's changelog bullet stays under `[Unreleased]`, because its fix is not in v3.1.0.

## Files Modified

- `src/bench/claudecode.go` (new): the claude-code adapter. Covers validation, install, the shadowing check, the isolated environment, argument building, stream-json parsing and the session UUID.
- `src/bench/{adapter,process,workspace,inventory,run,plan,manifest,report,events}.go`:
  - adapter dispatch and request/result fields
  - the shared `runProcess` helper
  - `HarnessHome` and `{harnessHome}`
  - `special` tree entries
  - the `Cleanup` failure and credential canary
  - plan-time harness version
  - the `ClaudeCodeConfig` schema
  - the report's Harnesses and Cleanup sections
- `src/bench/claudecode_test.go`, `src/bench/freeze_test.go` (new); `run_test.go` and `expect_test.go` (cleanup helper).
- `src/benchcli/bench.go`: cleanup warnings, and the module path.
- `src/go.mod`, `tests/go.mod`, `installer/go.mod`, `installer/main.go`, `install.sh`: identity moved to `ruifrvaz/smaqit-adk`.
- `.smaqit/bench/scenarios/claude-code-isolation/` (new live scenario plus canary treatment); `.smaqit/bench/scenarios/registry-cascade/bench.yaml` (credentials moved to `{harnessHome}`).
- `.smaqit/bench/README.md`, `docs/wiki/benchmarking.md`, `docs/wiki/extending-smaqit.md`, `README.md`, `CHANGELOG.md` (the `## [3.1.0]` section).
- `.smaqit/definitions/skills/smaqit.bench-{run,scaffold}.md` and `skills/smaqit.bench-{run,scaffold}/SKILL.md` (v0.3.0 via L2).
- `.smaqit/compendium.md` (freeze/cleanup entry updated, Claude Code harness entry added); `.smaqit/references/project-research.md` (Task 037 block, Claude Code row).
- `.smaqit/tasks/037_*.md` (decisions, findings, completion), `035_*.md` (abandoned), `PLANNING.md`; `smaqit-adk.code-workspace`.

## Next Steps

- File the follow-ups surfaced by the v3.1.0 install validation:
  - `--install-global` never seeds the base registry (`~/.agents/smaqit-adk/registry.json`), so cascades find no dependents on a release-installed machine.
  - `smaqit-adk sync` prints nothing and exits 0 when no registry exists.
  - The installer summary says "5 ADK skills" (there are 8).
  - There is no `smaqit-adk update` command; upgrading means re-running the install one-liner.
- Re-open Task 036 on ruifrvaz: push its branch, open a new PR (likely v3.1.1), and fold in the "5 ADK skills" message fix.
- Create the follow-up task for multi-turn scripted Claude Code Cases (`--resume` with prepared answers).
- Run smaqit task 131's single-turn `NEEDS-INPUT` Case against the new adapter.
- Task 025 (likely already resolved) and Task 031 (scenario matrix, now able to build on the harness selector) remain open.

## Session Metrics

- **Tasks completed:** 1 (037, released as v3.1.0); 1 folded and abandoned (035)
- **Code:** 27 files changed in the main implementation commit (about 2,000 lines added); 26 new unit tests
- **Subagents:**
  - 3 discovery agents
  - 2 parallel `smaqit.L2` compiles
- **Live Claude Code runs:**
  - one probe
  - two isolation-scenario runs (dev binary, then the installed v3.1.0 binary)
  - about $0.01 total
  - zero credential leaks, zero temp leftovers
- **Repository:** origin repointed to `ruifrvaz/smaqit-adk`; local `main` (21 commits) pushed; PR #25 merged; `adk-v3.1.0` released with 5 platform binaries plus `install.sh`
- **Global install:** upgraded from v2.0.1 to v3.1.0 and validated
