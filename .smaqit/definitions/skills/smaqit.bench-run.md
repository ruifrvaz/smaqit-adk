---
name: smaqit.bench-run
tier: advanced
---

# Skill Definition: smaqit.bench-run

## Identity

- **Name:** `smaqit.bench-run`
- **Description (draft, L2 may refine):** Automates running a project's `.smaqit/bench/` Bench suite end-to-end — preflight, structural validation, a plain-chat confirmation gate, live execution, pass/fail/inconclusive reporting, and failure diagnosis against known gotchas before treating anything as a novel bug.
- **Version:** 0.1.0

## Steps (with fragility)

1. **Preflight** (high fragility — exact commands). Confirm `codex` is on `PATH` via `command -v codex`; stop with install/auth instructions if missing (do not attempt to install it). If the project ships its own `installer/` (this repo's own dev convention), run `cd installer && make build` so `dist/smaqit-adk-dev` and any pre-compiled graders are current; otherwise assume the already-installed `smaqit-adk` binary on `PATH` is current.
2. **Structural validation** (high fragility). Run `smaqit-adk bench suite validate .smaqit/bench` (or `smaqit-adk bench validate <manifest.yaml>` when scoped to one manifest). Stop and report diagnostics verbatim on failure — no live spend past this point.
3. **Scope selection** (medium fragility). If the caller specified a manifest/case/variant, use it. Otherwise ask: "Run the whole `.smaqit/bench` suite, or a single manifest?"
4. **Confirmation gate** (high fragility — this is a hard invariant, not a suggestion). Always ask before live execution, stating manifest count and approximate cost/time: "This will run N manifest(s) live against the authenticated Codex CLI. Proceed?" No auto-confirm path exists. Do not proceed without an explicit yes, regardless of caller framing.
5. **Execute** (high fragility). Run `smaqit-adk bench suite run .smaqit/bench` (or the narrowed `bench run` invocation). Stream lifecycle events as they arrive.
6. **Report** (medium fragility). For each manifest: pass/fail/inconclusive outcome, comparison winner if multi-variant, and the `report.md` path (`<output.directory>/experiment-<id>/report.md`).
7. **Diagnose on failure** (low-medium fragility — judgment call, but check known gotchas first as a hard rule). For each failing manifest, read grade messages, trace logs (`traces/harness.std{out,err}.log`), and submission tree before concluding anything. Rule out known gotchas first: Snap-packaged Go toolchain vs. `go run` in a command grader (see `.smaqit/bench/README.md`), missing `--sandbox`/`--skip-git-repo-check` flags, Codex's own file-discovery missing dotfiles or its built-in skill-authoring feature colliding with a generic prompt. Offer `smaqit-adk bench grade <experiment-dir>` to re-grade without spending more quota when the failure looks grading-only.

## Output

- Plain-chat pass/fail/inconclusive summary per manifest, with `report.md` paths.
- For failures: a diagnosis distinguishing infra/known-gotcha flakes from genuine content failures, citing actual grade messages and trace logs as evidence.
- No subagent invoked — this skill executes directly.

## Scope

Runs and reports on an existing `.smaqit/bench/` suite. Does not author new manifests (that's `smaqit.bench-scaffold`). Does not patch `src/bench` engine gaps inline — a genuine engine limitation is reported precisely with a recommendation to file a follow-up task, never silently worked around in this skill.

## Completion

- [ ] `codex` on `PATH` confirmed, or preflight stopped cleanly with instructions
- [ ] Structural validation passed before any live execution
- [ ] Scope confirmed (whole suite or single manifest)
- [ ] User explicitly confirmed live execution before Step 5 ran
- [ ] Per-manifest outcome and `report.md` path reported
- [ ] Any failures diagnosed against known gotchas before being called novel

## Failure Scenarios

| Situation | Response |
|-----------|----------|
| `codex` not on `PATH` or unauthenticated | Stop with install/auth instructions; no live execution attempted |
| Structural validation fails | Report diagnostics verbatim; stop before the confirmation gate |
| User declines the confirmation gate | Stop cleanly; do not execute |
| A manifest run times out | Report `timedOut` in the summary; check `--skip-git-repo-check`/`--sandbox` gotchas before escalating |
| A command grader fails specifically on `go run` | Check the Snap/Go toolchain conflict documented in `.smaqit/bench/README.md` before treating it as a content failure |
| `src/bench` has a genuine limitation, not a known gotcha | Report it precisely with the reproducing manifest/case; recommend a follow-up task rather than patching the engine inline |
| `.smaqit/bench/` does not exist in this project | Report there is no suite to run; suggest `smaqit.bench-scaffold` |

## Gotchas

- Staged inputs in Bench manifests (`given.files`/`given.directories`) land in a read-only area outside the harness's actual working directory, reachable only via the `{input:<id>}` placeholder — this matters when diagnosing a failure that references staged content.
- `codex exec` requires `--sandbox danger-full-access` and `--skip-git-repo-check` in every process-variant block used by this repo's manifests; a failure that looks like a hang is often a missing one of these flags on a *new* manifest, not a `bench-run` bug.

## Examples

**Input:** User says "run the bench suite" (or invokes `/smaqit.bench-run`) in a project with `.smaqit/bench/skills/my-skill/bench.yaml` and no other manifests.
**Output:** Preflight passes (`codex` found). `bench suite validate .smaqit/bench` passes. Skill asks "Run the whole suite, or a single manifest?" — user says "whole suite". Skill asks for confirmation stating "1 manifest, live against Codex CLI — proceed?" — user confirms. Skill runs `bench suite run .smaqit/bench`, streams progress, and reports: "my-skill: winner (with-artifact passed, without-artifact correctly failed as baseline). Report: .smaqit/bench/runs/my-skill/experiment-.../report.md".

## Compatibility

Requires `codex` CLI on `PATH`, authenticated, for live execution. Structural validation (Steps 1–2) requires only the `smaqit-adk` binary.
