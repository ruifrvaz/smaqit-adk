# HarnessBench Phase 1 — `smaqit-adk bench` Subcommand

**Status:** In Progress
**Mode:** Assisted
**Created:** 2026-08-05
**Started:** 2026-08-06

## Description

Build **HarnessBench Phase 1**: a configurable local evaluation and benchmarking engine shipped as the `bench` subcommand of the `smaqit-adk` binary. Users and applications declare one or more cases, one or more harness variants, deterministic expected outputs, execution controls, and reporting policy in a versioned manifest.

Each case can provide a prompt plus named specs, files, directories, and images. Phase 1 executes any locally installed agentic harness through a generic process adapter, with task delivery through stdin or safe argument placeholders. One variant produces a standalone evaluation; two or more variants additionally produce a controlled comparison. Repetitions, deterministic grading, statistics, eligibility gates, and winner/tie/inconclusive outcomes remain traceable to immutable raw measurements.

The workflow is plan-first: `bench validate` checks configuration, `bench plan` resolves and hashes all referenced inputs and expands the exact run matrix, and `bench run` executes the saved plan only if drift checks pass. `bench run <manifest>` remains an auto-plan convenience. Native SDK/API adapters, remote harnesses, multi-turn protocols, semantic/LLM judging, perceptual image grading, self-contained plan bundles, worker-process crash isolation, hostile-code sandboxing, git-diff metrics, and HTML reporting are explicitly deferred.

This task originates from a one-shot implementation prompt at `assets/HARNESSBENCH_ONE_SHOT_PROMPT.md`. **That prompt was written for a different repository** — a .NET/C# solution built on a "Daisy workflow orchestration engine". Its entire "Required Daisy workflow model" section (receivers, abilities, transmitters, traversal rules, loopbacks, impulses) and all .NET-specific API guidance are **inapplicable to this Go codebase and must be ignored**. What remains valuable and MUST be carried forward is: the product boundary, the scientific-validity controls, the manifest shape, the grader taxonomy, the scoring/winner-selection ordering, the artifact contract, and the security constraints.

## Design Decisions

- **General engine, ADK benchmark as one use case** (user decision, 2026-08-06). Task 023 delivers a configurable local evaluation and benchmarking engine. Measuring ADK versus baseline remains an important example, not the engine's product boundary.
- **Placement: subcommand of the `smaqit-adk` binary** (user decision, 2026-08-05). Chosen for discoverability and one distributable.
- **Source boundary: product code under `src/`** (user correction, 2026-08-06). The engine and command implementation live in the independent `src` Go module. `installer/` remains the Make-driven packaging/staging boundary and imports only the source CLI entrypoint from its existing binary wrapper.
- **Generic local `process` adapter in Phase 1** (user decision, 2026-08-06). Any locally installed harness that can accept task input through stdin or arguments is configurable. A deterministic `mock` adapter supports tests and examples. The in-process Copilot SDK adapter is deferred, so `installer/` adds a YAML parser but not the Copilot SDK.
- **Plan-first workflow** (user decision, 2026-08-06). `bench validate` checks configuration, `bench plan` creates a reviewable immutable plan, and `bench run <plan>` executes that exact plan. `bench run <manifest>` remains an auto-plan convenience.
- **Reference-and-hash plans** (user decision, 2026-08-06). Plans resolve and hash configuration, visible inputs, hidden oracle assets, graders, fixtures, executable identity, seed, and run order without embedding source assets. Execution rejects missing or changed inputs before launching a harness.
- **Case-oriented multimodal inputs** (user decision, 2026-08-06). A manifest contains one or more cases. Each case may provide a prompt plus named specs, files, directories, and images. Generic process delivery exposes staged filesystem paths and a rendered task envelope; provider-native multimodal upload is deferred.
- **One or more variants** (user decision, 2026-08-06). One variant produces a standalone evaluation. Two or more variants additionally produce a benchmark comparison.
- **Required expectations gate eligibility** (user decision, 2026-08-06). Declared expected outputs are deterministic pass/fail gates. Optional weighted graders measure secondary qualities only after eligibility; their weights MUST sum exactly to `1.0` when present and are never silently normalized.
- **Generic output boundary.** Phase 1 evaluates process stdout/stderr and artifacts written to the workspace. Native final-chat-response extraction, multi-turn protocols, semantic/LLM judging, and perceptual image grading are deferred.
- **Hidden-oracle boundary.** Expected values, golden assets, grader definitions/scripts, saved plans, and experiment outputs never enter the harness workspace, request, arguments, environment, or logs. This is non-disclosure by staging discipline, not a security boundary against a hostile same-user process.
- **Fresh external workspaces and robust process lifecycle.** Every attempt runs outside the source repository. Setup/harness/grader commands use argument arrays, never shell strings. Timeout and cancellation terminate descendant processes and close owned resources.
- **Immutable evidence, revisioned derivations.** Raw requests, traces, submissions, and measurements are immutable. Regrading creates revisioned derived artifacts instead of overwriting prior evidence. Missing usage/cost values remain null, never zero.
- **Observable lifecycle for agents and users** (user decision, 2026-08-07). `bench run` emits readable state transitions by default and ordered JSONL with `-events jsonl` for agent/application callers. Every emitted run/attempt/progress/completion transition is first durably appended to the experiment's `events.jsonl`; `-events quiet` suppresses progress, and `-json` remains a single final result document.
- **Deterministic tests at both module boundaries.** Unit and integration tests live in `src/bench/`; black-box CLI tests live in `tests/`. Makefile and CI wiring MUST execute both suites with no model API or network requirement.

## Implementation Steps

1. **Create the versioned manifest contract.** Add strict YAML types and field-addressed diagnostics for one or more cases, one or more variants, execution controls, required expectations, optional graders, comparison policy, and output location. Resolve relative paths against the manifest directory and reject unknown fields, duplicate IDs, escaping destinations, invalid placeholders, invalid timeouts/repetitions, and optional weights that do not sum to `1.0`.

2. **Model case inputs.** Each case supports a prompt from exactly one inline value or file plus named specs, files, directories, and images. Inputs declare stable IDs, sources, optional media types, and containment-safe destinations. Keep the writable fixture distinct from staged read-only task inputs.

3. **Model process variants.** Configure executable, argument array, input mode (`stdin` or argument), working directory, explicitly inherited environment-variable names, non-secret environment overrides, setup commands, and intended treatment differences. Support argument-level placeholders `{task}`, `{taskFile}`, `{inputRoot}`, `{input:<id>}`, `{workspace}`, `{caseId}`, and `{variantId}`; never invoke a shell.

4. **Model expected outputs and optional graders.** Required expectations locate actual output in harness stdout/stderr, a submission file/directory, semantic JSON, or a command result. Optional graders remain a separate weighted mechanism for secondary metrics and never override failed expectations.

5. **Implement `bench validate`.** Parse and fully validate the manifest without running a harness or mutating a workspace. Provide human-readable field paths and stable JSON diagnostics for application callers.

6. **Implement `bench plan`.** Resolve configuration, sources, placeholders, defaults, executable identity, environment names, and intended differences; generate a seed when absent; expand `cases × variants × repetitions`; deterministically randomize order; hash all visible inputs, hidden oracle assets, fixtures, setup inputs, grader assets, and executable identity; emit comparability warnings; and atomically write a stable plan artifact.

7. **Make saved plans canonical.** `bench run <plan>` re-hashes references and rejects drift before launching anything. `bench run <manifest>` creates, persists, and executes a plan as a convenience. Case/variant/repetition filters may run subsets, but incomplete evidence is explicitly marked and cannot yield a conclusive comparison.

8. **Implement the adapter boundary.** Add `process` and `mock` adapters. Pass only a resolved run-scoped request containing run identity, workspace, rendered task input, executable/arguments/environment, and trace paths. Treat a started process's non-zero exit as a recorded result; reserve adapter errors for infrastructure failures. Keep unknown usage/cost metrics nullable.

9. **Implement process lifecycle management.** Execute in the attempt workspace, stream bounded stdout/stderr to trace files, enforce timeouts and cancellation, terminate descendant processes with platform-specific handling, and close all owned resources. Never persist inherited secret values.

10. **Implement isolated workspaces.** Create every attempt with `os.MkdirTemp` outside the repository, copy the fixture without changing its source, stage only declared agent-visible inputs under a reserved input root, render the task envelope, record the baseline inventory, freeze the completed submission, preserve evidence, and remove the temporary workspace.

11. **Implement deterministic evaluators.** Support text exact/contains/regex checks with explicit normalization; file exists/absent/byte/hash/size/content checks; stable directory inventory/tree/count checks; strict semantic JSON exact/subset checks; runtime exit/stdout/stderr checks; and exact-byte or SHA-256 image-output checks. Run commands only against disposable grading copies of frozen submissions.

12. **Enforce the hidden-oracle boundary.** Keep expected values, golden assets, grader definitions/scripts, plans, and experiment output paths out of harness workspaces, task envelopes, requests, arguments, environments, and logs. Store grader traces separately and assert the boundary in tests. Document that same-user process execution is not a hostile-code sandbox.

13. **Implement statistics and comparison.** Summarize repetitions per case and variant; emit evaluation outcomes for a single variant; for multiple variants, gate eligibility on required expectations, compute optional weighted scores and nullable statistics, apply minimum pass rate, tie threshold, and declared tie-breakers in a fixed order, surface uncontrolled differences, and return winner, tie, or inconclusive.

14. **Persist immutable evidence and revisioned derivations.** Atomically write saved plan, sanitized resolved configuration, experiment metadata, run matrix, per-attempt request/result, hashes, frozen submission, traces, grades, statistics, comparison, and Markdown report. Keep raw evidence immutable and write regrading results as revisioned derived artifacts.

15. **Wire the CLI and observable run state.** Implement `bench validate|plan|run|grade|compare|report` in `src/benchcli/` and dispatch to it from the existing `installer/main.go` packaging wrapper. Provide nested `--help`, single-result JSON modes, readable default `bench run` progress, ordered JSONL lifecycle events for agent/application callers, a durable per-experiment event journal, and stable exit codes for success, completed-but-ineligible, invalid input/configuration, drift, and infrastructure failure.

16. **Add deterministic verification.** Cover strict parsing, containment, input hashing, drift, seeded plans, placeholder expansion, process I/O and cleanup, all expectation types, required gating, optional weights, single-variant evals, multi-variant winner/tie/inconclusive outcomes, incomplete matrices, hidden-oracle exclusion, immutable fixtures, revisioned regrading, and full network-free mock experiments. Add built-binary black-box tests and wire installer-package tests into Makefile and CI.

17. **Ship examples and documentation.** Add a one-variant evaluation using prompt/spec/file/image inputs, a two-variant mock benchmark, and a generic process-harness example. Document the manifest and plan schemas, validate/plan/run lifecycle, input delivery, expectations, scoring, artifacts, application integration, secret handling, scientific-validity controls, and local-process security limits.

## Known Issues Triage

**Triaged:** 2026-08-06
**Tools searched:** go.yaml.in/yaml/v3
**Result:** Advisory

### Advisory Issues
- [#321 KnownFields(true) is not propagated when UnmarshalYAML calls Node.Decode](https://github.com/yaml/go-yaml/issues/321) — `yaml/go-yaml` — opened 2026-03-20 — no labels
- [#332 Fix: option KnownFields not being respected inside custom UnmarshalYAML()](https://github.com/yaml/go-yaml/pull/332) — `yaml/go-yaml` — opened 2026-04-14 — no labels

Implementation avoids the affected custom `UnmarshalYAML` → `Node.Decode` path and uses the maintained stable `go.yaml.in/yaml/v3` module with direct typed decoding.

## Acceptance Criteria

- [ ] `smaqit-adk bench validate|plan|run|grade|compare|report` all exist with nested `--help`, appear in top-level usage/help, and expose documented stable exit codes
- [ ] Machine-consumed commands provide stable JSON diagnostics/results without requiring applications to parse terminal prose
- [ ] `bench run` reports readable lifecycle state by default; `-events jsonl` emits ordered, timestamped, typed run/attempt/progress/completion or failure events; `-events quiet` suppresses progress
- [ ] Lifecycle events are durably recorded in each created experiment's immutable `events.jsonl` before being delivered to a caller; `run.completed` identifies the outcome, experiment directory, and report, while `run.failed` identifies the failure phase/message and diagnostics when available
- [ ] Manifest parsing rejects unknown fields and reports the exact offending field path
- [ ] Relative paths resolve against the manifest directory; source and destination containment checks reject escaping paths and unsafe symlinks
- [ ] A manifest accepts one or more cases containing prompt, spec, file, directory, and image inputs with unique stable IDs
- [ ] A manifest accepts one or more variants using the generic local `process` adapter; the deterministic `mock` adapter remains available for tests/examples
- [ ] Process harnesses accept task input through configured stdin or safe argument placeholders, including named input paths, without shell interpolation
- [ ] `bench plan` expands `cases × variants × repetitions`, generates and persists a seed when absent, and reproduces the same order from the same resolved inputs and seed
- [ ] Saved plans record hashes for configuration, visible inputs, hidden oracle assets, fixtures, setup/grader assets, and executable identity where available
- [ ] `bench run <plan>` detects missing or changed references and exits before launching any harness; `bench run <manifest>` persists the auto-generated plan
- [ ] Each attempt runs in a fresh workspace created outside the source repository, and source fixtures remain byte-for-byte unchanged
- [ ] Setup, harness, and grader commands execute with argument arrays only; timeout/cancellation terminates descendant processes and leaves no owned process running
- [ ] Text, file, directory, strict semantic JSON, runtime, and exact-byte/SHA-256 image-output expectations grade a frozen submission copy deterministically
- [ ] Expected values, golden assets, grader definitions/scripts, saved plans, and experiment outputs are absent from the target workspace and harness request surface, as asserted by tests
- [ ] Required expectations gate eligibility; a failing run cannot outrank an eligible run regardless of optional metrics
- [ ] Optional grader weights are rejected unless they sum exactly to `1.0`; no silent weight normalization occurs
- [ ] One variant produces a standalone evaluation outcome and statistics without requiring comparison
- [ ] Two or more variants produce an evidence-backed winner, tie, or inconclusive comparison with minimum pass rate, tie threshold, and tie-breakers applied in documented order
- [ ] Filtered or otherwise incomplete run matrices are reported as incomplete and cannot yield a conclusive comparison
- [ ] Unknown usage/cost metrics remain null, are excluded from aggregates, and have their exclusions counted
- [ ] Raw requests, traces, submissions, and measurements are immutable; regrading writes revisioned derived artifacts without rerunning the harness
- [ ] Plan, experiment, run, grade, comparison, and report artifacts are written atomically and remain traceable to their hashed inputs
- [ ] Deterministic unit/integration tests pass without a model API or network, including full single- and multi-variant mock experiments
- [ ] Runnable examples cover single-variant multimodal evaluation, two-variant comparison, and generic process-harness input delivery
- [ ] Benchmarking documentation covers manifest/plan formats, lifecycle, scoring, application integration, secrets, scientific-validity controls, and the explicit local-process security limitation
- [ ] `cd src && go test ./...`, `cd installer && go test ./...`, `cd installer && make build`, `cd installer && make test`, and `cd tests && go test ./...` pass; CI executes source-module, installer-wrapper, and black-box tests
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
| `src/bench/manifest.go`, `inputs.go` | Create — versioned schema, strict validation, case inputs, diagnostics, and manifest-relative resolution |
| `src/bench/plan.go`, `hash.go` | Create — immutable saved plan, hashing, drift detection, seed, and case × variant × repetition expansion |
| `src/bench/adapter.go`, `process.go`, `mock.go` | Create — adapter contract plus generic process and deterministic mock implementations |
| `src/bench/process_unix.go`, `process_windows.go` | Create — platform-specific descendant-process cleanup |
| `src/bench/workspace.go` | Create — external workspaces, safe staging, baseline inventory, submission freeze, and preservation |
| `src/bench/expect.go`, `grader.go` | Create — deterministic required expectations, optional graders, and hidden-oracle enforcement |
| `src/bench/stats.go`, `compare.go` | Create — nullable statistics, eligibility, ranking, ties, and inconclusive outcomes |
| `src/bench/artifacts.go`, `events.go`, `report.go` | Create — atomic immutable evidence, ordered lifecycle journal, revisioned derivations, JSON results, and Markdown reports |
| `src/bench/run.go` | Create — validate/plan/run/grade/compare/report orchestration |
| `src/bench/*_test.go` | Create — deterministic unit and integration coverage |
| `src/benchcli/bench.go` | Create — CLI parsing, human/JSON/JSONL output, run-state rendering, and stable exit semantics |
| `src/go.mod`, `src/go.sum` | Create — independent source module and YAML parsing dependency |
| `installer/main.go` | Modify — delegate the `bench` dispatch case to `src/benchcli` and update usage/help |
| `installer/go.mod`, `installer/go.sum` | Modify — reference the local source module for packaging |
| `installer/Makefile` | Modify — test source and installer modules and add example verification |
| `.github/workflows/test-integration.yml` | Modify — execute source-module, installer-wrapper, and black-box CLI tests |
| `examples/bench/single-eval/` | Create — prompt/spec/file/image evaluation example |
| `examples/bench/multi-variant/` | Create — deterministic two-variant mock benchmark |
| `examples/bench/process-harness/` | Create — generic stdin and named-input placeholder example |
| `docs/wiki/benchmarking.md` | Create |
| `README.md` | Modify — add benchmark/evaluation quick start |
| `tests/unit/bench_cli_test.go` | Create — black-box CLI smoke test |

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

**Process boundary:** Phase 1 launches each harness as a direct child process and MUST terminate its descendant process group on timeout or cancellation. It does not provide a separate crash-isolating worker or hostile-code sandbox. Direct recursive `smaqit-adk bench` configurations MUST fail validation when detectable; indirect recursion through arbitrary wrappers cannot be guaranteed and is documented as a limitation.

**A note on the first experiment:** the honest expected result is that a trivial task favors the baseline, because kit overhead can exceed its value at low complexity. The eventual benchmark corpus needs a complexity ladder — trivial generation, multi-requirement application, existing-repo modification, ambiguity, multi-module change, mid-task requirement change, regression repair. Do not tune the first task until the ADK wins; report what is measured.

**Related:** Task 026 will adopt HarnessBench as the repository's skill and agent evaluation suite, answering *"does the artifact behave as specified?"* alongside comparative questions such as *"which harness or treatment performs better?"*. Tasks 020 and 021 are superseded and their Copilot-specific session and LLM-grading code is not reused. Task 024 is complete; its external-workspace and process-lifecycle findings are inputs to this implementation.
