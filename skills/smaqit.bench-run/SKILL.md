---
name: smaqit.bench-run
description: Automates running a project's .smaqit/bench/ Bench suite end-to-end — preflight checks, structural validation, a plain-chat confirmation gate, live execution, pass/fail/inconclusive reporting, and failure diagnosis against known gotchas before treating anything as a novel bug.
metadata:
  version: "0.2.0"
---

# Bench Run

Structural validation requires only the smaqit-adk binary. Live execution requires an authenticated `codex` executable on `PATH`.

## Steps

### 1. Preflight

Confirm `codex` is on `PATH`:

```bash
command -v codex
```

If missing, stop and report install/auth instructions — do not attempt to install it.

If this project ships its own `installer/` (this repo's own dev convention), build the current dev binary and any pre-compiled graders first:

```bash
cd installer && make build
```

Otherwise assume the already-installed `smaqit-adk` binary on `PATH` is current.

### 2. Structural validation

Run before any live spend:

```bash
smaqit-adk bench suite validate .smaqit/bench
```

(or `smaqit-adk bench validate <manifest.yaml>` when scoped to one manifest — see Step 3.)

On failure, stop and report the diagnostics verbatim. Do not proceed to Step 4.

### 3. Scope selection

If the caller specified a manifest, case, or variant, use it. Otherwise ask:

> Run the whole `.smaqit/bench` suite, or a single manifest?

### 4. Confirmation gate

This is a hard invariant, not a suggestion. Always ask before live execution, stating manifest count and approximate cost/time:

> This will run N manifest(s) live against the authenticated Codex CLI. Proceed?

No auto-confirm path exists. Do not proceed without an explicit yes, regardless of caller framing.

### 5. Execute

```bash
smaqit-adk bench suite run .smaqit/bench
```

(or the narrowed `bench run` invocation for a single-manifest scope). Stream lifecycle events as they arrive.

### 6. Report

For each manifest: pass/fail/inconclusive outcome, comparison winner (if multi-variant), and the `report.md` path (`<output.directory>/experiment-<id>/report.md`).

### 7. Diagnose on failure

For each failing manifest, read its grade messages, trace logs (`traces/harness.std{out,err}.log`), and submission tree before concluding anything.

Rule out known gotchas first, as a hard rule, before calling anything a novel bug:
- Snap-packaged Go toolchain vs. `go run` in a command grader (see `.smaqit/bench/README.md`)
- Missing `--sandbox`/`--skip-git-repo-check` flags
- Codex's own file-discovery missing dotfiles, or its built-in skill-authoring feature colliding with a generic prompt

Offer to re-grade without spending more quota when the failure looks grading-only:

```bash
smaqit-adk bench grade <experiment-dir>
```

## Output

- Plain-chat pass/fail/inconclusive summary per manifest, with `report.md` paths.
- For failures: a diagnosis distinguishing infra/known-gotcha flakes from genuine content failures, citing actual grade messages and trace logs as evidence.
- No subagent invoked — this skill executes directly.

## Scope

Runs and reports on an existing `.smaqit/bench/` suite. Does not author new manifests — that's `smaqit.bench-scaffold`. Does not patch `src/bench` engine gaps inline: a genuine engine limitation is reported precisely with a recommendation to file a follow-up task, never silently worked around in this skill.

## Examples

**Input:** User says "run the bench suite" (or invokes `/smaqit.bench-run`) in a project with `.smaqit/bench/skills/my-skill/bench.yaml` and no other manifests.

**Output:** Preflight passes (`codex` found). `bench suite validate .smaqit/bench` passes. Skill asks "Run the whole suite, or a single manifest?" — user says "whole suite". Skill asks for confirmation stating "1 manifest, live against Codex CLI — proceed?" — user confirms. Skill runs `bench suite run .smaqit/bench`, streams progress, and reports: "my-skill: winner (with-artifact passed, without-artifact correctly failed as baseline). Report: `.smaqit/bench/runs/my-skill/experiment-.../report.md`".

## Gotchas

- Shared `given` inputs and variant `treatment` artifacts land in separate read-only Case-brief tables within a Bench-managed, submission-excluded subtree under the workspace root. Diagnose staged-content failures against the relevant table and ID; `{input:<id>}` and `{treatment:<id>}` expand only in permitted process arguments, not in raw prompt text.
- `codex exec` requires `--sandbox danger-full-access` and `--skip-git-repo-check` in every process-variant block used by this repo's manifests; a failure that looks like a hang is often a missing one of these flags on a *new* manifest, not a `bench-run` bug.

## Completion

- [ ] `codex` on `PATH` confirmed, or preflight stopped cleanly with instructions
- [ ] Structural validation passed before any live execution
- [ ] Scope confirmed (whole suite or single manifest)
- [ ] User explicitly confirmed live execution before Step 5 ran
- [ ] Per-manifest outcome and `report.md` path reported
- [ ] Any failures diagnosed against known gotchas before being called novel

## Failure Handling

| Situation | Action |
|-----------|--------|
| Required input not provided | Request the missing information before proceeding |
| Gathered input is ambiguous | Flag the ambiguity and ask for clarification |
| Subagent invocation fails | Report the failure with context; do not silently retry |
| Output artifact already exists | Confirm with user before overwriting |
| `codex` not on `PATH` or unauthenticated | Stop with install/auth instructions; no live execution attempted |
| Structural validation fails | Report diagnostics verbatim; stop before the confirmation gate |
| User declines the confirmation gate | Stop cleanly; do not execute |
| A manifest run times out | Report `timedOut` in the summary; check `--skip-git-repo-check`/`--sandbox` gotchas before escalating |
| A command grader fails specifically on `go run` | Check the Snap/Go toolchain conflict documented in `.smaqit/bench/README.md` before treating it as a content failure |
| `src/bench` has a genuine limitation, not a known gotcha | Report it precisely with the reproducing manifest/case; recommend a follow-up task rather than patching the engine inline |
| `.smaqit/bench/` does not exist in this project | Report there is no suite to run; suggest `smaqit.bench-scaffold` |
