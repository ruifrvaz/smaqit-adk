# HarnessBench Phase 1 — `smaqit-adk bench` Subcommand

**Status:** Not Started
**Created:** 2026-08-05

## Description

Build **HarnessBench Phase 1**: a controlled, repeatable A/B evaluation engine shipped as a `bench` subcommand of the `smaqit-adk` binary. Its purpose is to answer the question smaqit-adk currently cannot answer with evidence — *does injecting the ADK actually make a coding agent better, and at what cost?*

Phase 1 is deliberately the **smallest system that produces real evidence**. It runs N repetitions of a fixed task across two or more variants (e.g. baseline vs. `smaqit-adk lite` installed), grades each attempt with **deterministic** graders, computes distributions rather than single-run anecdotes, and selects a winner (or declares a tie) with an explanation traceable to raw measurements.

Phase 1 drives the **Copilot SDK in-process**. External harness CLIs (codex, claude-code, opencode), the worker-process boundary, git diff metrics, and HTML reporting are explicitly **Phase 2+** and out of scope here. The adapter interface is designed so Phase 2 adds a `process` adapter without touching orchestration.

This task originates from a one-shot implementation prompt at `assets/HARNESSBENCH_ONE_SHOT_PROMPT.md`. **That prompt was written for a different repository** — a .NET/C# solution built on a "Daisy workflow orchestration engine". Its entire "Required Daisy workflow model" section (receivers, abilities, transmitters, traversal rules, loopbacks, impulses) and all .NET-specific API guidance are **inapplicable to this Go codebase and must be ignored**. What remains valuable and MUST be carried forward is: the product boundary, the scientific-validity controls, the manifest shape, the grader taxonomy, the scoring/winner-selection ordering, the artifact contract, and the security constraints.

## Design Decisions

- **Placement: subcommand of the `smaqit-adk` binary** (user decision, 2026-08-05). Chosen over a separate module/binary for discoverability and a single distributable. **Accepted cost:** `installer/` today has *zero* external dependencies and produces a small static binary that every `curl | bash` user downloads; this task adds `copilot-sdk/go` and a YAML parser to it. **Mitigation:** all bench code lives under `installer/bench/`, importing nothing from the installer's install/uninstall paths, so extraction into its own module is a mechanical move if binary size later becomes a problem.
- **Phase 1 only — prove the thesis** (user decision, 2026-08-05). In scope: variants, repetitions, deterministic graders, statistics, gating, winner selection, JSON + Markdown reports. Out of scope: external process adapters, worker-process boundary + recursion guard, descendant-process kill, generalized overlays with path-containment, git diff metrics, HTML reporting, blind LLM judge.
- **Deterministic graders are the primary signal.** An LLM judge is *not* implemented in Phase 1. The existing eval runner grades exclusively by LLM, which is precisely the weakness HarnessBench exists to correct. Note that `assets/test-harness.png` places a "Blind comparative judge" on the main path between graders and ranking — this **contradicts** the prompt text ("must not be necessary to select a winner", "can supplement but not override required executable failures"). The prompt wins: the judge is a deferred, optional extension point.
- **Harvest, don't rewrite, the isolation code.** [tests/evals/runner/main.go](../../tests/evals/runner/main.go) already solves two hard problems that MUST be preserved: (a) the workspace is created *outside* the repo tree because a workspace inside it causes git-root discovery to leak full project context to the agent ([main.go:197-201](../../tests/evals/runner/main.go#L197-L201)); (b) an explicit token is mandatory so sessions never route through VS Code's shared XDG config and inherit smaqit-adk's own workspace context ([evals/README.md:31-39](../../tests/evals/README.md#L31-L39)). Violating either silently invalidates every result.
- **Adapter interface from day one, with a `mock` adapter.** Required so the test suite is fully deterministic and needs no model API or network access.
- **Unit tests live in-package** (`installer/bench/*_test.go`), diverging from the repo's `tests/` module convention. Rationale: `tests/` is deliberately black-box (drives the built binary via `SMAQIT_ADK_BIN`); bench unit tests need access to unexported types. Black-box CLI smoke tests still belong in `tests/`.

## Implementation Steps

1. **Add dependencies to `installer/go.mod`:** `github.com/github/copilot-sdk/go v0.2.0` (same version already used by `tests/go.mod`) and `gopkg.in/yaml.v3`. Run `go mod tidy` in `installer/`.

2. **`installer/bench/manifest.go` — manifest types, parsing, validation.**
   - Support the YAML shape below (a Phase-1 subset of the prompt's schema, preserving field names so Phase 2 is additive).
   - Decode with `yaml.Decoder.KnownFields(true)` so a typo cannot silently invalidate an experiment.
   - Resolve every relative path against the **manifest's own directory**, not the process CWD.
   - Validate: at least 2 variants, unique variant IDs, weights sum to 1.0 (or normalize with a recorded warning), required graders present, repetitions ≥ 1, timeout > 0.
   - Emit diagnostics naming the exact offending field path.

   ```yaml
   schemaVersion: 1
   name: adk-vs-baseline
   fixture:
     source: ./fixtures/empty-python        # optional local dir; copied per attempt
   task:
     instructionsFile: ./tasks/hello-mario.md
   variants:
     - id: baseline
       adapter: copilot                     # copilot | mock
     - id: adk-lite
       adapter: copilot
       setup:                               # executable + argument array; never a shell string
         - executable: smaqit-adk
           arguments: ["lite"]
   execution:
     repetitions: 10
     randomizeOrder: true
     seed: null                             # generated and recorded when null
     timeoutSeconds: 300
   graders:
     - id: runtime-output
       type: commandAssertion
       required: true
       weight: 0.5
       executable: python
       arguments: ["main.py"]
       assertions: { exitCode: 0, stdoutContains: ["hello", "mario"], ignoreCase: true }
     - id: scope
       type: repository
       required: false
       weight: 0.5
       assertions: { maximumFilesCreated: 5, forbiddenContent: ["TODO", "NotImplementedError"] }
   comparison:
     minimumRequiredPassRate: 0.9
     tieThreshold: 0.01
     tieBreakers: [higherRequiredPassRate, higherMedianScore, lowerMedianDuration]
   output:
     directory: ./bench-results
   ```

3. **`installer/bench/plan.go` — run plan.** Expand `variants × repetitions` into a deterministic ordered plan. When `randomizeOrder` is true, shuffle with a seeded PRNG; **generate the seed if absent and always persist it** to `run-plan.json`. The same seed MUST reproduce the same order.

4. **`installer/bench/adapter.go` — harness adapter boundary.**
   ```go
   type Adapter interface {
       Name() string
       Execute(ctx context.Context, req Request) (Result, error)
   }
   ```
   - `copilot` adapter: creates a session against the isolated workspace, sends the task instructions, waits for completion, records duration, terminal status, and any usage the SDK exposes. Deny shell permission requests only if the task's graders do not require generated code to run; otherwise approve — record which policy applied.
   - `mock` adapter: writes a configured set of files and returns a configured status. Deterministic, no network. Used by every integration test.
   - **Missing token/cost/usage data MUST be `nil`, never `0`.** Use pointer or `sql.Null`-style optional fields so "unknown" and "zero" stay distinguishable through statistics and reporting.

5. **`installer/bench/workspace.go` — per-attempt isolation.**
   - `os.MkdirTemp` **outside the repo tree** (never inside the project — see Design Decisions).
   - Copy the fixture in without modifying the source; reject any resolved destination that escapes the workspace root.
   - Run variant `setup` commands via `exec.CommandContext` with an **argument slice** — never build a shell string, never pass through `sh -c`.
   - Record a baseline file inventory after setup and before harness execution.
   - Preserve the workspace into the run directory afterward, then remove the temp dir.

6. **`installer/bench/grader.go` — deterministic graders.** Registry keyed by `type`:
   - `command` — run an executable, capture stdout/stderr/exit code/duration, honour a timeout, pass on exit 0.
   - `commandAssertion` — as above plus assertions on exit code, `stdoutContains`, `stderrContains`, regex, with `ignoreCase`.
   - `repository` — filesystem-only assertions: `fileExists`, `fileAbsent`, `maximumFilesCreated` (vs. the recorded baseline inventory), `forbiddenContent`, `requiredContent`. No git dependency in Phase 1.
   - **Grading runs against a frozen copy** of the submission, so grader-created files and caches cannot contaminate the recorded metrics.
   - **Grader definitions and any grader assets MUST live outside the target workspace** and must never be copied into it — the agent under test must not be able to read or game them.

7. **`installer/bench/stats.go`** — count, success/failure/timeout counts, mean, median, min, max, stddev. Values that are unknown propagate as unknown and are excluded from aggregates, with the exclusion counted and reported.

8. **`installer/bench/compare.go` — gating and winner selection, in this exact order:**
   1. mark each run's required-grader status;
   2. compute required pass rate per variant;
   3. disqualify variants below `minimumRequiredPassRate`;
   4. normalize optional grader scores to `[0,1]`;
   5. compute the weighted score;
   6. compute statistics per metric;
   7. rank eligible variants;
   8. apply `tieThreshold`, then declared tie-breakers in order;
   9. if still tied or evidence is insufficient, return **tie / inconclusive** — never invent a winner.
   - **Functional correctness dominates:** a variant that fails a required grader can never outrank one that passes, regardless of speed or size.
   - Emit **comparison warnings** when variants differ in materially uncontrolled ways (differing adapters, differing setup, missing metrics on one side).

9. **`installer/bench/report.go`** — write `experiment.json`, `resolved-manifest.json`, `run-plan.json`, `comparison.json`, and `report.md`. Write **atomically**: temp file in the same directory, flush, then rename, so a crash never leaves valid-looking partial JSON. The Markdown narrative MUST be generated from structured facts — never ask a model to author the experiment's conclusion.

10. **`installer/bench/run.go`** — orchestration: validate → plan → per-attempt (workspace → setup → execute → freeze → grade → persist) → aggregate → compare → report. One attempt failing must not abort collection of the others unless fail-fast is set.

11. **`installer/cmd_bench.go` + `installer/main.go`** — add the `bench` case to the dispatch switch at [installer/main.go:39](../../installer/main.go#L39) and subcommands: `validate <manifest>`, `run <manifest> [--variant id] [--repetition n]`, `grade <run-dir>`, `compare <experiment-dir>`, `report <experiment-dir> [--format json|markdown]`. Every subcommand gets `--help`. Update `printUsage` and `cmdHelp`. Exit codes: `0` success, `1` completed but no eligible successful result, `2` invalid CLI input or manifest, `3` infrastructure/orchestration failure.

12. **Artifact layout** under `output.directory`:
    ```
    experiment-<id>/
      experiment.json  resolved-manifest.json  run-plan.json
      comparison.json  report.md
      runs/<run-id>/
        request.json  result.json
        workspace/  submission/
        traces/harness.stdout.log  traces/harness.stderr.log
        grades/results.json
    ```
    `result.json` normalizes at minimum: `schemaVersion`, `experimentId`, `runId`, `variantId`, `repetition`, `status` (enum: `completed|failed|timedOut|cancelled|invalid`), `startedAt`, `completedAt`, `durationMs`, `harness{adapter,exitCode,timedOut}`, `usage{...nullable}`, `grades{requiredPassed,score}`, `failure` (phase, error type, concise message, log path — never a serialized exception graph).

13. **Tests.** Unit (`installer/bench/*_test.go`): manifest parsing incl. unknown-field rejection, relative path resolution, plan expansion + seed determinism, required-grader gating, weight validation, statistics + tie threshold, tie-breaker ordering, unknown-metrics-stay-unknown, status transitions. Integration (mock adapter, no network): a full two-variant multi-repetition experiment; required-grader failure disqualifies a higher optional score; a known winner is selected; equal results produce a tie; parallel-safe distinct workspaces; grader assets absent from the workspace; source fixture byte-for-byte unchanged; `grade`/`compare`/`report` operate on existing artifacts without re-running a harness. Add a black-box CLI smoke test in `tests/`.

14. **Example + docs.** Ship a runnable `examples/bench/adk-vs-baseline/` (manifest, task instructions, minimal fixture, deterministic grader, and a `mock`-adapter variant usable with no model access). Write `docs/wiki/benchmarking.md` covering what HarnessBench measures and what it does not, the control/treatment model, manifest reference, scoring and winner selection, artifact layout, and — stated plainly — that Phase 1 is local process isolation, **not** a security sandbox for hostile generated code.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] `smaqit-adk bench validate|run|grade|compare|report` all exist with `--help`, and appear in `printUsage`/`cmdHelp`
- [ ] Manifest parsing rejects unknown fields and reports the offending field path
- [ ] Relative paths resolve against the manifest directory, not the process CWD
- [ ] Run plan expands variants × repetitions; seed is generated when absent, persisted, and reproduces the same order
- [ ] Each attempt runs in a fresh workspace created **outside** the repo tree
- [ ] Setup and grader commands execute via argument slices — no shell string interpolation anywhere
- [ ] `command`, `commandAssertion`, and `repository` graders are implemented and grade a frozen submission copy
- [ ] Grader definitions and assets are provably absent from the target workspace (asserted by a test)
- [ ] Required graders gate eligibility: a required-grader failure can never outrank a passing run
- [ ] Comparison produces an evidence-backed winner, tie, or inconclusive result, with tie-breakers applied in declared order
- [ ] Unknown usage metrics are reported as null, never zero (asserted by a test)
- [ ] `experiment.json`, `resolved-manifest.json`, `run-plan.json`, `comparison.json`, and `report.md` are written atomically
- [ ] Full deterministic test suite passes with no model API or network access (mock adapter)
- [ ] The `examples/bench/adk-vs-baseline/` example runs end to end using the mock adapter
- [ ] `docs/wiki/benchmarking.md` written, including the explicit security-limitations statement
- [ ] `cd installer && make build` passes; `cd tests && go test ./...` passes
- [ ] No TODOs, placeholder methods, or commented-out core behavior remain

## Findings

[Populated by smaqit.task-complete. Do not fill in manually before task is complete.]

**Implementation approach:**
- TBD

**Decisions made:**
- TBD

**Blockers encountered:**
- TBD

**Follow-up identified:**
- TBD

## Files to Create / Modify

| File | Action |
|------|--------|
| `installer/bench/manifest.go` | Create |
| `installer/bench/plan.go` | Create |
| `installer/bench/adapter.go` | Create |
| `installer/bench/workspace.go` | Create |
| `installer/bench/grader.go` | Create |
| `installer/bench/stats.go` | Create |
| `installer/bench/compare.go` | Create |
| `installer/bench/report.go` | Create |
| `installer/bench/run.go` | Create |
| `installer/bench/*_test.go` | Create |
| `installer/cmd_bench.go` | Create |
| `installer/main.go` | Modify — add `bench` dispatch case, update usage/help |
| `installer/go.mod`, `installer/go.sum` | Modify — add copilot-sdk + yaml.v3 |
| `installer/Makefile` | Modify — add a `bench-example` target |
| `examples/bench/adk-vs-baseline/` | Create — manifest, task, fixture, grader |
| `docs/wiki/benchmarking.md` | Create |
| `tests/unit/bench_cli_test.go` | Create — black-box CLI smoke test |
| `assets/HARNESSBENCH_ONE_SHOT_PROMPT.md` | Modify — commit as the source reference, annotated that the Daisy/.NET sections do not apply |

## Notes

**Scientific-validity controls that MUST survive implementation** (from the source prompt — these are the reason the tool exists, not decoration):

- identical task and starting state for paired variants
- fresh isolated workspace per attempt
- configurable repetitions; one stochastic run is never conclusive
- randomized execution order with the seed recorded
- immutable raw run artifacts; raw measurements retained even when an aggregate score is computed
- deterministic executable graders preferred over model judgment
- graders hidden from the target harness
- missing metrics are unknown, never zero
- ties and statistically weak differences reported honestly — no manufactured causal certainty
- material uncontrolled differences between variants surfaced as warnings

**Recursion guard (deferred but noted):** Phase 2 introduces a worker-process boundary, at which point `harnessbench experiment → worker → external harness` must be enforced with a depth limit so a misconfiguration cannot spawn recursively. Phase 1 has no worker process and therefore no recursion risk, but variant `setup` commands MUST NOT be permitted to invoke `smaqit-adk bench`.

**A note on the first experiment:** the honest expected result is that a trivial task favors the baseline, because kit overhead can exceed its value at low complexity. The eventual benchmark corpus needs a complexity ladder — trivial generation, multi-requirement application, existing-repo modification, ambiguity, multi-module change, mid-task requirement change, regression repair. Do not tune the first task until the ADK wins; report what is measured.

**Related:** Task 020 and Task 021 (behavioral evals) answer *"does the artifact behave as specified?"*. HarnessBench answers *"is the kit worth it?"*. They are complementary, not overlapping, and neither blocks the other. Task 024 (broken eval references) is worth resolving first if this task ends up reusing code from `tests/evals/runner/`.
