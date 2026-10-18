---
name: smaqit.bench-run
tier: advanced
---

# Skill Definition: smaqit.bench-run

## Identity

- **Name:** `smaqit.bench-run`
- **Description (draft, L2 may refine):** Automates running a project's `.smaqit/bench/` Bench suite end-to-end — per-harness preflight (Codex CLI or Claude Code), structural validation, a plain-chat confirmation gate, live execution, pass/fail/inconclusive reporting, and failure diagnosis against known gotchas before treating anything as a novel bug.
- **Version:** 0.3.0

## Steps (with fragility)

1. **Preflight** (high fragility — exact commands). First determine which harnesses the selected manifests use: a variant with `adapter: claude-code` uses Claude Code; a `process` variant whose executable is `codex` (directly or inside an `sh -c` wrapper) uses Codex CLI. Run only the checks for harnesses actually in scope:
   - **Codex CLI:** `command -v codex`. If missing, stop with install/auth instructions — do not attempt to install it.
   - **Claude Code:** `command -v claude` and `claude --version`. Read each claude-code variant's `claudeCode.credential` (`CLAUDE_CODE_OAUTH_TOKEN` or `ANTHROPIC_API_KEY`). For `CLAUDE_CODE_OAUTH_TOKEN` on Linux/WSL, the credential is loaded at execution time directly from the invoking user's own login — no vault, no prompt to paste it: confirm `~/.claude/.credentials.json` exists and that `.claudeAiOauth.expiresAt` (epoch milliseconds) is later than now plus the run's timeout budget, checking only that field and never printing the token (for example `jq -e --argjson budget_ms <run-timeout-budget-ms> '.claudeAiOauth.expiresAt > (now * 1000 + $budget_ms)' ~/.claude/.credentials.json`). For `ANTHROPIC_API_KEY`, confirm the variable is set by name only (`[ -n "$ANTHROPIC_API_KEY" ]`), never echoing it. If the login has expired or the credential is unavailable, stop and tell the user to refresh it (for example by using Claude Code interactively once, or `claude setup-token` on macOS where the login lives in the Keychain).
   - If the project ships its own `installer/` (this repo's own dev convention), run `cd installer && make build` so `dist/smaqit-adk-dev` and any pre-compiled graders are current; otherwise assume the already-installed `smaqit-adk` binary on `PATH` is current.
2. **Structural validation** (high fragility). Run `smaqit-adk bench suite validate .smaqit/bench` (or `smaqit-adk bench validate <manifest.yaml>` when scoped to one manifest). Stop and report diagnostics verbatim on failure — no live spend past this point.
3. **Scope selection** (medium fragility). If the caller specified a manifest/case/variant, use it. Otherwise ask: "Run the whole `.smaqit/bench` suite, or a single manifest?" Preflight (Step 1) applies to the harnesses of the chosen scope: if the chosen scope uses a harness Step 1 did not check, run that check before the confirmation gate.
4. **Confirmation gate** (high fragility — this is a hard invariant, not a suggestion). Always ask before live execution, stating manifest count, the harnesses involved, and approximate cost/time: "This will run N manifest(s) live against <the authenticated Codex CLI / Claude Code (your own login) / both>. Proceed?" No auto-confirm path exists. Do not proceed without an explicit yes, regardless of caller framing.
5. **Execute** (high fragility). Run `smaqit-adk bench suite run .smaqit/bench` (or the narrowed `bench run` invocation). When any claude-code variant in scope uses `CLAUDE_CODE_OAUTH_TOKEN`, load it in the same command, never in a separate step and never printed: `CLAUDE_CODE_OAUTH_TOKEN="$(jq -r .claudeAiOauth.accessToken ~/.claude/.credentials.json)" smaqit-adk bench suite run .smaqit/bench`. Extract only that one field (the same file also holds MCP OAuth tokens). Never copy the credentials file anywhere. Stream lifecycle events as they arrive, and surface any `WARNING attempt cleanup failed` line prominently.
6. **Report** (medium fragility). For each manifest: pass/fail/inconclusive outcome, comparison winner if multi-variant, and the `report.md` path (`<output.directory>/experiment-<id>/report.md`). For claude-code variants, include the harness version from the report's Harnesses table and the token/cost usage from `result.json` when present. Report any attempt with a `cleanup` failure, naming the directory to inspect.
7. **Diagnose on failure** (low-medium fragility — judgment call, but check known gotchas first as a hard rule). For each failing manifest, read grade messages, trace logs (`traces/harness.std{out,err}.log`, plus `traces/transcript.jsonl` for Claude Code), `result.json`'s `failure`/`cleanup`, and the submission tree before concluding anything. Rule out known gotchas first:
   - Snap-packaged Go toolchain vs. `go run` in a command grader (see `.smaqit/bench/README.md`)
   - Codex: missing `--sandbox`/`--skip-git-repo-check` flags; Codex's own file-discovery missing dotfiles, or its built-in skill-authoring feature colliding with a generic prompt
   - Claude Code: a `treatment would be shadowed` failure (a same-named skill/agent/command in the fixture's or an ancestor's `.claude/`, or a managed location); `installed treatment not discovered` (wrong install `as`, a skill directory without `SKILL.md`, or an agent without a frontmatter `name`); `reportedError` (the final `is_error: true` — often an expired login or a permission denial under `dontAsk`; check the stdout log's final `result` event); a missing-credential failure; an `evidence`-phase failure (the harness echoed its credential and Bench redacted it — a genuine finding about the Case, not an engine bug)
   - Offer `smaqit-adk bench grade <experiment-dir>` to re-grade without spending more quota when the failure looks grading-only.

## Output

- Plain-chat pass/fail/inconclusive summary per manifest, with `report.md` paths, harness versions for claude-code variants, and any cleanup failures.
- For failures: a diagnosis distinguishing infra/known-gotcha flakes from genuine content failures, citing actual grade messages and trace logs as evidence.
- No subagent invoked — this skill executes directly.

## Scope

Runs and reports on an existing `.smaqit/bench/` suite. Does not author new manifests (that's `smaqit.bench-scaffold`). Does not patch `src/bench` engine gaps inline — a genuine engine limitation is reported precisely with a recommendation to file a follow-up task, never silently worked around in this skill. Never stores, prints, or copies a harness credential; the Claude Code login token is only ever loaded inline into the executing command.

## Completion

- [ ] Harnesses in scope identified; preflight passed for each (`codex` found, and/or `claude` found with a valid unexpired credential checked by name/expiry only), or preflight stopped cleanly with instructions
- [ ] Structural validation passed before any live execution
- [ ] Scope confirmed (whole suite or single manifest)
- [ ] User explicitly confirmed live execution before Step 5 ran
- [ ] Credential (when needed) loaded inline in the executing command, never printed or copied
- [ ] Per-manifest outcome, `report.md` path, and any cleanup failures reported
- [ ] Any failures diagnosed against known gotchas before being called novel

## Failure Scenarios

| Situation | Response |
|-----------|----------|
| `codex` not on `PATH` or unauthenticated | Stop with install/auth instructions; no live execution attempted |
| `claude` not on `PATH` | Stop with install instructions; no live execution attempted |
| Claude Code login missing or `expiresAt` too soon for the run budget | Stop and ask the user to refresh the login; never ask them to paste a token |
| Structural validation fails | Report diagnostics verbatim; stop before the confirmation gate |
| User declines the confirmation gate | Stop cleanly; do not execute |
| A manifest run times out | Report `timedOut` in the summary; check the harness's known flag/permission gotchas before escalating |
| A command grader fails specifically on `go run` | Check the Snap/Go toolchain conflict documented in `.smaqit/bench/README.md` before treating it as a content failure |
| An attempt reports a `cleanup` failure | Report the leaked directory path from the message and tell the user to inspect and delete it; never treat it as a pass |
| An attempt fails in phase `evidence` | Report that the harness echoed its credential into output and Bench redacted it; treat it as a Case finding |
| `src/bench` has a genuine limitation, not a known gotcha | Report it precisely with the reproducing manifest/case; recommend a follow-up task rather than patching the engine inline |
| `.smaqit/bench/` does not exist in this project | Report there is no suite to run; suggest `smaqit.bench-scaffold` |

## Gotchas

- Shared `given` inputs and variant `treatment` artifacts land in separate read-only Case-brief tables within a Bench-managed, submission-excluded subtree under the workspace root. Diagnose staged-content failures against the relevant table and ID; `{input:<id>}` and `{treatment:<id>}` expand only in permitted process arguments, not in raw prompt text.
- `codex exec` requires `--sandbox danger-full-access` and `--skip-git-repo-check` in every process-variant block used by this repo's manifests; a failure that looks like a hang is often a missing one of these flags on a *new* manifest, not a `bench-run` bug.
- A Claude Code run succeeds or fails on its final `is_error`, not its exit code alone: an unauthenticated run can exit 1 with `subtype: "success"` and `is_error: true`. Bench records this as `reportedError`.
- Claude Code live runs on one login token must stay sequential; concurrent headless calls on one token have been reported to invalidate the session (`anthropics/claude-code#98693`).

## Examples

**Input:** User says "run the bench suite" (or invokes `/smaqit.bench-run`) in a project with `.smaqit/bench/skills/my-skill/bench.yaml` (a Codex `process` comparison) and no other manifests.
**Output:** Preflight passes (`codex` found). `bench suite validate .smaqit/bench` passes. Skill asks "Run the whole suite, or a single manifest?" — user says "whole suite". Skill asks for confirmation stating "1 manifest, live against Codex CLI — proceed?" — user confirms. Skill runs `bench suite run .smaqit/bench`, streams progress, and reports: "my-skill: winner (with-artifact passed, without-artifact correctly failed as baseline). Report: .smaqit/bench/runs/my-skill/experiment-.../report.md".

**Input:** User asks to run `.smaqit/bench/scenarios/claude-code-isolation/bench.yaml`, whose two variants use `adapter: claude-code` with `credential: CLAUDE_CODE_OAUTH_TOKEN`.
**Output:** Preflight finds `claude` (2.1.296), confirms `~/.claude/.credentials.json` exists and its `expiresAt` covers the 300 s budget without printing the token, validation passes, and the skill asks "1 manifest, 2 attempts live against Claude Code using your own login — proceed?". On yes it runs `CLAUDE_CODE_OAUTH_TOKEN="$(jq -r .claudeAiOauth.accessToken ~/.claude/.credentials.json)" smaqit-adk bench run .smaqit/bench/scenarios/claude-code-isolation/bench.yaml` and reports "winner: with-treatment; harness claude-code 2.1.296; $0.004 total; no cleanup failures".

## Compatibility

Live execution requires the harness CLIs the selected manifests use: an authenticated `codex` on `PATH` for Codex `process` variants, and `claude` on `PATH` plus either the invoking user's Claude Code login (Linux/WSL, loaded inline as `CLAUDE_CODE_OAUTH_TOKEN`) or an `ANTHROPIC_API_KEY` for claude-code variants; `jq` is used to load the login token. Structural validation requires only the `smaqit-adk` binary.
