# Benchmarking and evaluation

`smaqit-adk bench` runs repeatable local evaluations against any harness that can be launched as a process. Configuration comes first: validate a YAML manifest, inspect a saved plan, then execute exactly that plan.

The implementation is in `src/bench` with CLI handling in `src/benchcli`. `installer/` only packages the resulting command into the existing distributable.

```text
manifest.yaml -> bench validate -> bench plan -> saved plan -> bench run -> experiment artifacts
```

One variant is an evaluation. Two or more variants add a comparison. Every case declares the prompt and visible inputs supplied to the harness plus required deterministic expectations. Optional weighted command graders measure secondary qualities but cannot make a functionally failing run eligible.

## Lifecycle

```bash
smaqit-adk bench validate bench.yaml
smaqit-adk bench plan -out experiment.plan.json bench.yaml
smaqit-adk bench run experiment.plan.json
```

`bench run bench.yaml` is shorthand that writes a plan into the configured output directory before running it. A saved plan references and hashes the manifest, every input, fixtures, oracle/grader assets, and process executables. `run` re-hashes them and exits before launching a harness if anything drifted.

Use `run -case <id>`, `-variant <id>`, or `-repetition <n>` for a diagnostic subset. Subset artifacts are explicitly marked incomplete and their comparison outcome is always inconclusive.

By default, `bench run` reports readable run and attempt state as work progresses. An agent or application can request an ordered newline-delimited JSON stream instead:

```bash
smaqit-adk bench run -events jsonl experiment.plan.json
```

The stream uses typed `run.started`, `attempt.started`, `attempt.completed`, `run.progress`, and `run.completed` events. Failures end with `run.failed`. Every event has a schema version, sequence number, and timestamp. `run.completed` identifies the outcome, experiment directory, and Markdown report; `run.failed` identifies the failure phase/message and validation diagnostics when available. Once an experiment directory exists, its event sequence is durably appended to `<experiment>/events.jsonl` before it is emitted. Use `-events quiet` when only the final result is wanted.

Applications can add `-json` to every subcommand for one final JSON document. For `bench run`, `-json` implies quiet events; use `-events jsonl` instead when streaming state is required. Exit codes are stable: `0` means success, `2` means a valid completed experiment was ineligible or inconclusive, `3` means invalid CLI/configuration, `4` means plan drift, and `5` means an infrastructure failure.

## Suites

Every command above operates on exactly one manifest. `bench suite <validate|plan|run> <directory>` discovers every `bench.yaml` found anywhere under a directory tree (sorted, so order is deterministic) and drives each one through the same validate/plan/run pipeline in turn — useful once a project has many independent benchmarks rather than one:

```bash
smaqit-adk bench suite validate ./benchmarks
smaqit-adk bench suite plan ./benchmarks
smaqit-adk bench suite run ./benchmarks
```

A manifest that fails to load, plan, or run is recorded against it individually and does not stop the rest of the suite from proceeding. `bench suite run` forwards each manifest's lifecycle events (prefixed with its path) to `-events plain|jsonl|quiet` the same way `bench run` does, and its final JSON document reports per-manifest results plus suite-level `passed`/`failed`/`errored` counts. There is no cross-manifest comparison — each manifest's own `Variants` comparison (see below) is computed independently; the suite only aggregates pass/fail across manifests, not across variants in different files. Exit codes follow the same convention as `bench run`, applied to the suite as a whole.

smaqit-adk's own `.smaqit/bench/` is an example consumer of `bench suite` — it dogfoods the engine against smaqit-adk's own skills and agents. See `.smaqit/bench/README.md` for that specific layout convention; it is this repository's own usage, not part of the `bench` command's contract.

## Manifest

The schema is strict YAML with `schemaVersion: 2`; unknown fields fail validation. Paths are resolved relative to the manifest. Version 2 uses **Case** for an evaluation scenario, **Prompt** for author-supplied `given.prompt`, and **Case brief** for the rendered prompt plus declared input locations delivered to a harness. A minimal evaluation is:

```yaml
schemaVersion: 2
name: greeting
cases:
  - id: hello
    given:
      prompt: {text: "Return hello"}
    expect:
      - {id: output, type: text, actual: stdout, operator: exact, value: hello}
variants:
  - id: local-agent
    adapter: process
    process:
      executable: my-agent
      arguments: ["--brief-file", "{briefFile}"]
      inputMode: argument
execution: {repetitions: 3, randomizeOrder: true, seed: 23, timeoutSeconds: 300}
comparison: {minimumRequiredPassRate: 1, tieThreshold: 0.01}
output: {directory: ./bench-results}
```

Each Case supports exactly one inline or file prompt and named `specs`, `files`, `directories`, and `images`. Every shared input has an `id`, `source`, optional contained `destination`, and optional `mediaType`. Bench separates three data planes:

- `fixture` copies a common source directory into a fresh writable workspace. Its optional `destination` is a safe workspace-relative path; `.` is the default.
- Case-level `prepare` commands run after fixture copy but before the baseline snapshot and sidecar staging. They are common to every variant and may use only `{workspace}` and `{caseId}`.
- Variant `treatment` assets are read-only, variant-specific files or directories staged at Bench-managed sidecar paths. They have `id`, `source`, and optional `mediaType` fields—no workspace destination. A variant with treatments must declare `intendedDifferences`; a baseline normally has an empty treatment set.

Shared `given` inputs are also staged read-only in the sidecar, but unlike treatments they are available to every variant.

The output directory must be outside every fixture, shared input directory, and treatment source directory. This prevents prior plans, oracle artifacts, and experiment results from entering a later harness workspace and keeps directory hashes stable.

Process arguments support `{brief}`, `{briefFile}`, `{inputRoot}`, `{input:<id>}`, `{treatment:<id>}`, `{workspace}`, `{caseId}`, and `{variantId}`. Placeholder IDs are validated against the Case and variant where they are used. Placeholders are expanded inside individual argument values; prompt text itself is not interpolated. `{brief}` and `inputMode: stdin` provide the rendered Case brief, while `{briefFile}` names its read-only file. The Case brief preserves the author prompt and renders separate shared-input and variant-treatment tables; an empty baseline treatment is explicit. No shell string is evaluated. Environment inheritance is opt-in by variable name; `environment.set` is intended only for non-secret values.

## Expectations and scoring

Required expectation types are:

- `text`: exact, contains, or regular-expression matching against stdout, stderr, or a submission file.
- `file`: exists, absent, exact bytes, SHA-256, size, or normalized content.
- `directory`: exists, absent, count, exact inventory/tree, or required/forbidden paths.
- `json`: strict exact or recursive subset comparison.
- `runtime`: exit code, stdout, or stderr checks.
- `image`: exact bytes or SHA-256 (perceptual grading is deliberately out of scope).
- `command`: runs a direct executable against a disposable grading copy.

Optional graders use `type: command`; each returns score `1` for exit code zero and `0` otherwise. Their positive weights must sum exactly to `1.0`. Required expectations are evaluated first and determine eligibility. For multiple variants, the engine applies minimum pass rate, optional score, tie threshold, then declared tie-breakers. Missing token/cost/tool metrics stay JSON `null` and are counted as missing.

## Evidence and re-derivation

An experiment contains its plan, sanitized resolved manifest, ordered `events.jsonl` lifecycle journal, immutable per-run requests/results, bounded stdout/stderr traces, frozen submissions, revisioned grades, aggregate statistics, comparison JSON, and Markdown report. `bench grade`, `bench compare`, and `bench report` create new derived revisions without launching the harness again. Because sidecar contents are deliberately not persisted as harness evidence, grade re-derivation verifies the saved digests of referenced prompt files, shared inputs, and treatments before reconstructing a temporary read-only grading sidecar; missing or changed references fail explicitly as reference drift.

Expected values, golden assets, grader assets, saved plans, and output directories are never staged into the harness workspace or Case brief. This protects against accidental disclosure; it is not a sandbox against a malicious same-user process. A local harness has the operating-system permissions of the invoking user. Run only trusted executables and use stronger OS/container isolation for hostile code.

For credible comparisons, keep the Case prompt, fixture, preparation, shared inputs, budgets, and environment constant; vary only declared treatment artifacts; use multiple repetitions; randomize with a recorded seed; and treat ties, missing metrics, or incomplete matrices as inconclusive rather than manufacturing certainty. The entire sidecar is excluded from snapshots, repository metrics, and frozen submissions, so evaluator-provided resources never count as harness output.

Runnable examples are under `examples/bench/`: `single-eval`, `multi-variant`, and the POSIX `process-harness` example.
