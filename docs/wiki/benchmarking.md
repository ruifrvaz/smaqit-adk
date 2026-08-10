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

## Manifest

The schema is strict YAML with `schemaVersion: 1`; unknown fields fail validation. Paths are resolved relative to the manifest. A minimal evaluation is:

```yaml
schemaVersion: 1
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
      arguments: ["--task-file", "{taskFile}"]
      inputMode: argument
execution: {repetitions: 3, randomizeOrder: true, seed: 23, timeoutSeconds: 300}
comparison: {minimumRequiredPassRate: 1, tieThreshold: 0.01}
output: {directory: ./bench-results}
```

Each case supports exactly one inline or file prompt and named `specs`, `files`, `directories`, and `images`. Every input has an `id`, `source`, optional contained `destination`, and optional `mediaType`. A fixture directory is copied into a fresh writable workspace; visible inputs are staged separately and read-only.

The output directory must be outside every fixture and declared input directory. This prevents prior plans, oracle artifacts, and experiment results from entering a later harness workspace and keeps directory hashes stable.

Process arguments support `{task}`, `{taskFile}`, `{inputRoot}`, `{input:<id>}`, `{workspace}`, `{caseId}`, and `{variantId}`. Placeholders are expanded inside individual argument values. No shell string is evaluated. `inputMode: stdin` also writes the prompt to standard input. Environment inheritance is opt-in by variable name; `environment.set` is intended only for non-secret values.

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

An experiment contains its plan, sanitized resolved manifest, ordered `events.jsonl` lifecycle journal, immutable per-run requests/results, bounded stdout/stderr traces, frozen submissions, revisioned grades, aggregate statistics, comparison JSON, and Markdown report. `bench grade`, `bench compare`, and `bench report` create new derived revisions from existing evidence without launching the harness again.

Expected values, golden assets, grader assets, saved plans, and output directories are never staged into the harness workspace or task envelope. This protects against accidental disclosure; it is not a sandbox against a malicious same-user process. A local harness has the operating-system permissions of the invoking user. Run only trusted executables and use stronger OS/container isolation for hostile code.

For credible comparisons, keep task, fixture, inputs, budgets, and environment constant; declare intended treatment differences; use multiple repetitions; randomize with a recorded seed; and treat ties, missing metrics, or incomplete matrices as inconclusive rather than manufacturing certainty.

Runnable examples are under `examples/bench/`: `single-eval`, `multi-variant`, and the POSIX `process-harness` example.
