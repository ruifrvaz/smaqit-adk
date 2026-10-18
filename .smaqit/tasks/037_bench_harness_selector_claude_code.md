---
status: Completed
created: "2026-10-11"
mode: Assisted
started: "2026-10-18"
completed: "2026-10-18"
---

# Bench Harness Selector: First-Class Claude Code Harness Alongside Codex

## Description

Bench's engine doesn't care which harness it runs: the `process` adapter (`src/bench/adapter.go`) executes whatever `variants[].process.executable` a manifest names. Everything around it, however, assumes Codex CLI:

- the reusable process block in `.smaqit/bench/README.md` (`codex exec --sandbox danger-full-access --skip-git-repo-check --cd {workspace} {briefFile}`)
- the preflight checks and gotchas in `smaqit.bench-run` and `smaqit.bench-scaffold`
- the credential handling that task 035 found leaking Codex credentials

Running a Claude Code Case today means hand-rolling all of the following, and two of the items are real blockers:

1. **Invocation:** headless `claude -p` flags (output format, permission mode, turn limit, working directory) are undocumented in Bench and differ from Codex.
2. **Staging (blocker):** Bench stages variant treatments into the read-only `.smaqit-bench-input/treatment/` sidecar. Claude Code discovers agents, commands and skills only from a project's `.claude/` directory or from its config directory. Personal skills in the config directory override project skills of the same name. On a machine that has the released product installed globally, the installed copies therefore shadow the treatment under test. A sound run needs an isolated `CLAUDE_CONFIG_DIR` per run, with the treatment staged where Claude Code discovers it.
3. **Authentication (blocker):** an isolated config directory has no login. Credentials must reach the harness through an explicit, opt-in source (for example an inherited `ANTHROPIC_API_KEY`, or a credentials file copied in and removed afterwards). They must never be frozen into evidence. Task 035 is the precedent: a leaked workspace held a copy of real Codex CLI credentials.
4. **Evidence:** capture the session transcript (e.g. `--output-format stream-json`) into the run's trace directory, and record the harness name and version in the report.
5. **Conversation turns:** headless Claude Code has no user to answer `AskUserQuestion`. That's sufficient for single-turn Cases. A Case that tests an ask-and-resume loop also needs scripted follow-up turns, for example `--resume <session-id>` with prepared answers.

A **harness selector** makes Claude Code first-class next to Codex. It would be a declared `harness` that expands to a vetted invocation, staging layout, environment and trace capture, while raw `process` stays available as the escape hatch.

**Motivating case:** smaqit task 131 (2026-10-11). It made smaqit agents return `NEEDS-INPUT`/`REVIEW-REQUESTED` signals instead of asking. On Claude Code it preloads each agent's input skill through `skills:` frontmatter, and the main conversation resumes subagents with `SendMessage`. Its live acceptance check depends on behavior specific to Claude Code: whether `skills:` preload actually injects content, that subagents don't get `AskUserQuestion`, and that resume works. A Codex Bench run can't exercise any of that.

## Issue Triage Context

**Mode:** Auto
**Technologies:** Claude Code, Codex CLI, Go
**Platforms/Environments:** Linux, WSL2, macOS
**Features/Integrations:** Bench process adapter, new claude-code adapter, headless claude -p, CLAUDE_CONFIG_DIR isolation, CLAUDE_CODE_OAUTH_TOKEN, subagents, Agent Skills, stream-json transcripts, Bench freeze and workspace cleanup
**Versions/Constraints:** Bench schemaVersion 2, smaqit-adk 3.0.x, Claude Code 2.1.296, credentials must never enter frozen evidence, Claude Code personal skills override project skills of the same name, `--bare` disables skill/agent discovery

## Design Decisions

Resolved in `task.plan` (2026-10-18).

- **Selector: a new `adapter: claude-code`** with a typed `claudeCode:` block, not a plan-time preset on `process`. Creating a per-run config directory and staging treatments into it are runtime acts, so a preset alone can't do them. Raw `process` stays unchanged as the escape hatch. Manifest and plan stay at `schemaVersion: 2`, since the change is additive.
- **Staging: an isolated per-run config directory.** Bench creates a "harness home" outside the workspace (never snapshotted or frozen) holding an isolated `HOME` and `CLAUDE_CONFIG_DIR`. Treatments listed in `install` are copied into its `skills/`, `agents/` and `commands/`. Bench never stages into the workspace's `.claude/`. A same-name definition in the fixture's own `.claude/` or in a managed/enterprise location fails the run.
- **Credentials: exactly one opt-in environment variable**, `ANTHROPIC_API_KEY` or `CLAUDE_CODE_OAUTH_TOKEN`. No credential file is copied, and every other `ANTHROPIC_*`, `CLAUDE_*` and `ANTHROPIC_PROFILE` variable is absent from the harness environment. After freeze, a canary scan of the run's evidence redacts and fails on any literal occurrence of the credential value.
- **Credential source for live runs: a direct loader, not Vault.** The same shell command that runs `smaqit-adk bench run` reads `.claudeAiOauth.accessToken` from the invoking user's `~/.claude/.credentials.json` into `CLAUDE_CODE_OAUTH_TOKEN`. The value is never printed or written anywhere else. It's documented in `smaqit.bench-run`'s preflight and the README, not built into the engine. The loader extracts exactly one field, because the same file also holds the user's MCP OAuth tokens. It's Linux/WSL only, since macOS keeps the token in the Keychain.
- **Task 035 is folded in.** Its freeze crash, its swallowed cleanup errors, and a newly found leak are fixed here. The new leak: registry-cascade copies `~/.codex` into `{workspace}`, which freeze would capture into `submission/` once the socket crash is fixed. 035's open question is resolved as **skip with a `special` tree marker**, not an error. Task 035 is marked superseded by 037.
- **Multi-turn scripted Cases (`--resume` with prepared answers) are a follow-up task.** Single-turn covers smaqit task 131's `NEEDS-INPUT` Case.
- **Task 031 boundary:** 037 owns only the Claude Code adapter docs, the README process block, and the harness sections of the bench skills. 031 keeps the generic scenario matrix.
- **Release:** a new adapter is a MINOR bump.

### Discovery findings (Claude Code 2.1.296, 2026-10-18)

- `-p --output-format stream-json` requires `--verbose`. There is no `--cwd` flag; the working directory is the process cwd. `--session-id <uuid>` preassigns the session. `--no-session-persistence` exists.
- `--bare` skips auto-discovery of skills, agents and commands, and accepts only API-key auth, so it can't be used to stage treatments. `--safe-mode` disables customizations entirely. Neither replaces `CLAUDE_CONFIG_DIR` isolation.
- `CLAUDE_CONFIG_DIR` relocates settings, agents, skills, commands, plugins, transcripts (`projects/<encoded-cwd>/<session>.jsonl`) and `.credentials.json`. The docs don't say it relocates `~/.claude.json`, so an isolated `HOME` is also needed.
- Discovery precedence: skills are enterprise > personal > project. Agents are managed > `--agents` > project > user. Project discovery walks from cwd up to the repo root.
- Auth precedence in `-p`: cloud-provider variables > `ANTHROPIC_AUTH_TOKEN` > `ANTHROPIC_API_KEY` (always used when present) > apiKeyHelper > `CLAUDE_CODE_OAUTH_TOKEN` > Anthropic profile > `/login`.
- Exit code: an unauthenticated run exits 1, yet stdout's JSON `result` shows `subtype: "success"` with `is_error: true`. Success has to be judged from `is_error`, not the exit code or subtype alone.
- Subagents never get `AskUserQuestion`. `skills:` frontmatter on a subagent injects the full skill content at startup.
- The workspace trust dialog is skipped in `-p`, but a project's `.claude/settings.json` hooks and `.mcp.json` servers still run.
- **Step 1 probe, 2026-10-18.** Setup: Claude Code 2.1.296, `env -i` with only `PATH`, `HOME`, `CLAUDE_CONFIG_DIR` and `DISABLE_AUTOUPDATER`, an empty temp config dir, one Haiku turn costing $0.0014. Results:
  - A `/login` access token passed as `CLAUDE_CODE_OAUTH_TOKEN` **authenticates**: exit 0, `is_error: false`, `apiKeySource: "none"`, and no onboarding gate.
  - `system/init` **lists `skills`, `agents` and `slash_commands`**. A staged `probe-skill` and `probe-agent` both appeared. Only skills bundled with the CLI loaded alongside them, with no personal `smaqit.*` skills, which confirms the isolation. The discovery assertion is viable.
  - With `CLAUDE_CONFIG_DIR` set, `.claude.json` was created inside the config dir. `HOME` is still isolated for issues #97680 and #80580.
  - The transcript was written to `<config>/projects/<cwd with non-alphanumerics as ->/<session-id>.jsonl`, and `--session-id` was honored.
  - The credential string appeared in no stdout/stderr output and no harness-home file.

## Implementation Steps

**Phase A: Engine safety (folds in Task 035)**

1. **Probe** (done before any code; no API calls except one minimal turn). Confirm that a `/login` access token passed as `CLAUDE_CODE_OAUTH_TOKEN` authenticates in an empty `CLAUDE_CONFIG_DIR`. Confirm whether `system/init` lists the loaded skills and agents. Record both results under Design Decisions. If the token isn't accepted, fall back to a one-time `claude setup-token`, with the token passed directly and still not stored in Vault.
2. **Freeze handles non-regular files.** `snapshotTree` (`src/bench/inventory.go`) records sockets, FIFOs and devices as `Kind: "special"` with no digest, mirroring `hash.go`'s `IsRegular()` guard. `compareTrees` and the submission copy in `runAttempt` skip special entries, and `RepositoryMetrics` counts them.
3. **Cleanup failures are always reported.** Remove the `&& result.Failure == nil` guard in `runAttempt`'s deferred cleanup (`src/bench/run.go`). Add a separate `RunResult.Cleanup` failure. Surface it in `report.md` and in the plain and JSONL `attempt.completed` events, and mark the run failed even when an earlier failure exists.
4. **Per-run harness home.** Add `Workspace.HarnessHome` (`src/bench/workspace.go`): a separate `os.MkdirTemp("", "smaqit-bench-home-")` directory, mode 0700, never inside `Root`. The same deferred cleanup removes it, and its failures go through step 3. Expose a `{harnessHome}` placeholder to `process` arguments (`allowedPlaceholders` in `manifest.go`, `renderArguments` in `adapter.go`).
5. **Repair registry-cascade.** `.smaqit/bench/scenarios/registry-cascade/bench.yaml`'s `sh -c` wrapper copies `~/.codex` into `{harnessHome}` and exports `HOME`/`CODEX_HOME` there instead of `{workspace}`.

**Phase B: Claude Code adapter** (depends on Phase A)

6. **Schema.** Add `Variant.ClaudeCode *ClaudeCodeConfig` with these fields:
   - `executable` (default `claude`)
   - `model`
   - `permissionMode` (required)
   - `maxTurns`
   - `allowedTools` / `disallowedTools`
   - `arguments` (extra, placeholder-validated)
   - `credential` (enum `ANTHROPIC_API_KEY` | `CLAUDE_CODE_OAUTH_TOKEN`)
   - `install[]` (`{treatment, as: skill|agent|command}`)
   - `environment` (non-secret only; rejects `HOME`, `CLAUDE_CONFIG_DIR` and credential-shaped names)

   `Validate` adds a `"claude-code"` adapter case, mutually exclusive with `process` and `mock`. It checks that each install treatment ID exists on the variant, and checks install shape: a skill is a directory with `SKILL.md`; an agent or command is a `.md` file.
7. **Plan.** `BuildPlan` (`src/bench/plan.go`) resolves and hashes the `claude` executable through `resolveExecutable`, runs `<exe> --version` once with an empty environment, and records `HarnessVersion` per variant. It warns when the credential variable differs between variants.
8. **Runtime adapter.** Add a new `claudeCodeAdapter` in `src/bench/claudecode.go`, registered in `adapterFor`:
   1. Install treatments into `HarnessHome/home/.claude/{skills/<name>/, agents/<name>.md, commands/<name>.md}`, with names taken from frontmatter.
   2. Fail on shadowing: a same-name definition in the fixture's `.claude/` or in a managed location.
   3. Build the environment from scratch: inherited `PATH`, `HOME`/`CLAUDE_CONFIG_DIR` pointed into the harness home, `DISABLE_AUTOUPDATER=1`, `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1`, and the declared credential copied from the parent environment. The parent's own `CLAUDE_CODE_*` session variables never pass through.
   4. Build the argument list: `-p --output-format stream-json --verbose --permission-mode … [--model] [--max-turns] [--allowedTools …] --session-id <uuid>`, with the Case brief on stdin and the workspace root as cwd.
   5. Extract and reuse the process-group spawn, timeout and trace helper from `process.go`.
9. **Evidence.**
   - stream-json stdout goes to `traces/harness.stdout.log`, and the session file `<config>/projects/*/<session-id>.jsonl` is copied to `traces/transcript.jsonl`.
   - Parse the final `result` event: `is_error: true` means the harness failed. Fill `Usage` with tokens, `total_cost_usd` as estimated cost, and the tool-use count.
   - If `system/init` lists them, assert that every installed treatment was discovered.
   - `HarnessResult` gains `Name`, `Version` and `SessionID`, and `report.md` gains a per-variant harness/version line.
10. **Credential canary.** After freeze, scan `submission/`, `traces/` and the run's JSON evidence for the literal credential value. On a match, redact it and fail the run with phase `evidence`.

**Phase C: Skills, docs, verification** (depends on Phase B)

11. **Skills.** Update `.smaqit/definitions/skills/smaqit.bench-run.md` and `smaqit.bench-scaffold.md` with:
    - harness selection
    - Claude Code preflight: `command -v claude`, `claude --version`, the direct credential loader, and a `expiresAt` check against the run budget
    - gotchas: shadowing, `is_error`, why `--bare` can't be used
    - harness-neutral wording in place of Codex-only wording

    Recompile both via `smaqit.L2`.
12. **Docs.** Add a Claude Code variant block next to the Codex one in `.smaqit/bench/README.md`, and move the misplaced Codex YAML under its own heading. Document the adapter, harness home, credential model and loader, special-file handling and the cleanup guarantee in `docs/wiki/benchmarking.md`. Add a harness line to `README.md` and an entry to `CHANGELOG.md`.
13. **Tests.** Unit tests cover:
    - a socket or FIFO in the workspace
    - a forced cleanup failure alongside an earlier failure
    - `{harnessHome}` sitting outside `Root` and being removed
    - `claude-code` validation
    - install staging and shadowing detection
    - stream-json and `is_error` parsing
    - the credential canary
    - the parent's `CLAUDE_CODE_*` variables not leaking into the harness environment

    These run against a fake `claude` script that emits canned stream-json, with no network.
14. **Live verification.** Run a new `.smaqit/bench/scenarios/claude-code-isolation/bench.yaml` through the direct loader. Its skill treatment is deliberately named after a skill already installed in `~/.claude/skills`, carrying a canary behavior, with `with-treatment` and baseline variants. Afterwards, confirm that searching the experiment directory for the credential value finds nothing, and that `/tmp/smaqit-bench-*` is empty. smaqit task 131's `NEEDS-INPUT` Case runs downstream once this is released.

## Known Issues Triage
**Triaged:** 2026-10-18
**Tools searched:** Claude Code (`anthropics/claude-code`), Codex CLI (`openai/codex`), Go (`golang/go`)
**Result:** Advisory

Queries: claude-code `Linux` + `CLAUDE_CONFIG_DIR`, and `headless` + `CLAUDE_CODE_OAUTH_TOKEN`; codex `Linux` + `ipc socket`; go `Linux` + `WalkDir socket`. Each was run open and closed. Opened and closed dates weren't captured in the compact projection.

### Blocking Issues
- None confirmed. #97680 below meets the literal label criteria: open, `bug`, `platform:linux`, and it names `CLAUDE_CONFIG_DIR`. It was downgraded to Advisory because this task's design isolates `HOME` as well as `CLAUDE_CONFIG_DIR`, so paths that ignore the config dir land in the per-run harness home, not the real `~/.claude`.

### Advisory Issues
- [#97680 Daemon paths and session metadata ignore `CLAUDE_CONFIG_DIR`, writing to `~/.claude` (Linux)](https://github.com/anthropics/claude-code/issues/97680) — `anthropics/claude-code` — bug, has repro, platform:linux, area:core. Mitigated by the isolated `HOME`. Verify that no writes reach the real `~/.claude` during the live run.
- [#80580 CLAUDE_CONFIG_DIR does not fully isolate user-level CLAUDE.md discovery from `$HOME/.claude`](https://github.com/anthropics/claude-code/issues/80580) — `anthropics/claude-code` — reproduced. Confirms the isolated `HOME` is required, not optional.
- [#80791 A `.claude` config root at or above cwd is treated as project scope and overrides the CLAUDE_CONFIG_DIR profile's settings](https://github.com/anthropics/claude-code/issues/80791) — `anthropics/claude-code` — stale, reproduced. The workspace lives under the system temp dir; the shadowing check should also reject a `.claude` in workspace ancestors.
- [#97264 With CLAUDE_CONFIG_DIR set, starting in `$HOME` records a project-scope install of every user plugin](https://github.com/anthropics/claude-code/issues/97264) — `anthropics/claude-code` — bug, has repro, platform:linux, area:plugins. The cwd is the workspace, never `HOME`.
- [#78988 `CLAUDE_CONFIG_DIR` containing `~` is treated as a cwd-relative path](https://github.com/anthropics/claude-code/issues/78988) — `anthropics/claude-code` — bug, platform:macos, area:core. Always pass an absolute path.
- [#102985 Headless Linux server loses its OAuth login when many Claude Code processes share `~/.claude/.credentials.json`](https://github.com/anthropics/claude-code/issues/102985) — `anthropics/claude-code` — bug, has repro, platform:linux, area:auth. The loader only reads the access token, and the harness never sees or writes the user's credentials file.
- [#98693 OAuth session invalidated by concurrent headless `claude --print` calls](https://github.com/anthropics/claude-code/issues/98693) — `anthropics/claude-code` — bug, platform:macos, area:auth. Keep live runs sequential (no parallel repetitions) on the user's login token.
- [#84903 `CLAUDE_CODE_OAUTH_TOKEN` from `claude setup-token` returns 401 Invalid bearer token](https://github.com/anthropics/claude-code/issues/84903) — `anthropics/claude-code` — stale. Relevant to the step 1 probe and to the `setup-token` fallback.

### Historical (Closed)
- [#73403 `CLAUDE_CODE_OAUTH_TOKEN` ignored on first run due to onboarding gate](https://github.com/anthropics/claude-code/issues/73403) — `anthropics/claude-code` — platform:windows. An empty config dir is exactly a first run; the step 1 probe must confirm this doesn't happen on Linux.
- [#68241 `CLAUDE_CODE_OAUTH_TOKEN` shadowed by a stale `~/.claude/.credentials.json`](https://github.com/anthropics/claude-code/issues/68241) — `anthropics/claude-code` — regression, platform:linux. The isolated config dir holds no credentials file, so this can't recur.
- [#33313 YOLO mode still prompts for local commands due to temporary IPC sockets](https://github.com/openai/codex/issues/33313) — `openai/codex` — sandbox, CLI. Context for Codex creating sockets under its home (Task 035's crash).

### Unresolvable Tools
- None

### Omitted Tools
- None

### Search Warnings
- None. The research map's task block supplied the GitHub repositories for Claude Code and Codex CLI; Go was resolved through the helper.

## Acceptance Criteria

- [x] A manifest can select the Claude Code harness declaratively, without hand-writing its invocation, and raw `process` variants keep working unchanged
- [x] Agent, skill and command treatments are discoverable by Claude Code in the run, and a run fails rather than silently testing an installed copy that shadows the treatment
- [x] Credentials reach the harness only through an explicit opt-in source, never appear in frozen evidence, and are cleaned up on success, failure and timeout
- [x] The session transcript is captured into the trace directory, and the report records the harness and its version
- [x] `smaqit.bench-run`, `smaqit.bench-scaffold` and the README document harness selection, with Claude Code preflight and gotchas
- [x] Unit tests cover selector validation and staging, and a live run against authenticated Claude Code passes for a single-turn Case
- [x] A harness-created non-regular file (socket, FIFO, device) in the workspace no longer crashes the freeze step; it is recorded as a `special` entry and excluded from the submission (from Task 035)
- [x] A workspace or harness-home cleanup failure is always surfaced in `result.json`, `report.md` and CLI output, never dropped because an earlier failure was already recorded (from Task 035)
- [x] No harness configuration or credential copy is placed inside the frozen workspace: the per-run harness home lives outside it, and registry-cascade no longer copies `~/.codex` into `{workspace}`
- [x] `.smaqit/bench/README.md` documents the special-file and cleanup guarantees (from Task 035)

## Findings

**Implementation approach:**
- Added a first-class `adapter: claude-code` (`src/bench/claudecode.go`) with a typed `claudeCode` block, validated in `manifest.go`, resolved, hashed and version-probed at plan time in `plan.go`, and dispatched from `adapterFor`. Raw `process` is unchanged.
- Extracted the process-group spawn, timeout and trace plumbing from `process.go` into a shared `runProcess`, used by both adapters.
- Added a per-attempt harness home (`Workspace.HarnessHome`, `{harnessHome}` for variant process arguments) outside the workspace and removed with it, holding the isolated `HOME`/`CLAUDE_CONFIG_DIR` and installed treatments.
- Folded in Task 035: `snapshotTree` records sockets, FIFOs and devices as `special` entries excluded from submissions; cleanup failures go to a dedicated `RunResult.Cleanup`, the `attempt.completed` event, `report.md` and a CLI warning; a post-attempt credential canary scan redacts and fails on leaks.
- Moved the registry-cascade manifest's Codex credential copy from `{workspace}` into `{harnessHome}` via `CODEX_HOME`.
- Brought the stale `bench-run`/`bench-scaffold` definitions (unchanged since Task 028, while the compiled skills had been hand-edited by Tasks 029 and 030) back in line with the compiled skills, added harness selection, and recompiled both via `smaqit.L2` to v0.3.0.
- Repo identity moved back to the public `github.com/ruifrvaz/smaqit-adk`, at the user's request during completion, after `daisy-tribe/smaqit-adk` became unreachable: Go module paths, imports, CLI help, `install.sh` (gh-only flow kept) and README/wiki install docs. Local `main` (all Task 032/034/036 history) was merged with the user's README web edit and pushed once to `ruifrvaz` main. No tags were pushed and no release was triggered. v2.1.0 and v3.0.0 are deliberately not backfilled.
- Testing: 26 new unit tests against a fake `claude` script plus real sockets and FIFOs, and a live scenario (`.smaqit/bench/scenarios/claude-code-isolation`) whose treatment shares its name with an installed personal skill.

**Decisions made:**
- A new adapter, not a plan-time preset, because config-dir creation and treatment staging are runtime acts. Both schema versions stay at 2, since the change is additive.
- Exactly one credential variable (`CLAUDE_CODE_OAUTH_TOKEN` or `ANTHROPIC_API_KEY`), never a credentials file. For live runs it is loaded directly from the invoking user's login inside the run command, not through Vault (user decision).
- Any same-named skill, command or agent in the workspace's or an ancestor's `.claude/` or a managed location fails the attempt. Discovery is asserted against the `system/init` skill and agent lists.
- `DISABLE_AUTOUPDATER=1` is the only default. `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC` and `DISABLE_TELEMETRY` are deliberately not set, because both also disable feature-flag fetching. This departs from the original plan.
- Symlinks keep their existing handling. Only sockets, FIFOs and devices became `special`.
- A cleanup failure marks the attempt not passing, and the first recorded failure is preserved.

**Blockers encountered:**
- `daisy-tribe/smaqit-adk` became unreachable (HTTP 404, and the org no longer appears for the user). This was resolved by repointing `origin` to `ruifrvaz/smaqit-adk`; the old URL is kept as the `daisy-tribe` remote.
- The installed `smaqit-adk` is v2.0.1, which predates the registry, so the L2 recompiles were not registered.
- Discovered during implementation: opening a FIFO in the old freeze path would have hung forever rather than erroring.

**Follow-up identified:**
- Multi-turn scripted Claude Code Cases (`--resume` with prepared answers) need a follow-up task.
- smaqit task 131's single-turn `NEEDS-INPUT` Case should run downstream against this adapter once it is released.
- Task 031 (scenario matrix) should build on the harness selector. A Codex counterpart of the per-attempt isolation (`CODEX_HOME` in `{harnessHome}`) could become a first-class Codex adapter.
- Task 036's PR #3 lived on the unreachable `daisy-tribe` repo. Its branch needs re-pushing to `ruifrvaz` and a new PR, or its one-line fix can ride along with a later release.
- The real install should be upgraded from v2.0.1 so registry-driven propagation tracks the recompiled bench skills.

## Files to Create / Modify

| File | Action |
|------|--------|
| `src/bench/manifest.go`, `src/bench/adapter.go`, `src/bench/workspace.go`, `src/bench/process.go` | Modify — selector, Claude Code staging, environment, trace |
| `src/bench/*_test.go` | Modify — selector and staging tests |
| `skills/smaqit.bench-run/SKILL.md`, `skills/smaqit.bench-scaffold/SKILL.md` | Modify — harness selection |
| `.smaqit/bench/README.md`, `docs/wiki/benchmarking.md`, `README.md` | Modify — Claude Code process block and docs |
| `CHANGELOG.md` | Modify |

## Notes

- Filed 2026-10-11 from a smaqit session. smaqit task 131's live Claude Code check is blocked on this task, or on a hand-rolled one-off run with an isolated `CLAUDE_CONFIG_DIR`.
- Related: task 031 (scenario matrix; harness A/B as a Variant dimension), task 035 (credential leak and cleanup), task 033 (`prepare` background-process teardown).
- Claude Code docs to anchor discovery: https://code.claude.com/docs/en/sub-agents, https://code.claude.com/docs/en/skills, https://code.claude.com/docs/en/headless

Child tasks inherit their active parent's branch, worktree, and workflow mode. Only a standalone or parent task owns Git lifecycle cleanup.
