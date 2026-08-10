# HarnessBench Phase 1 — `smaqit-adk bench` Subcommand — E2E Test Playbook

**Test ID:** 023  
**Title:** HarnessBench Phase 1 — `smaqit-adk bench` Subcommand  
**Date:** 2026-08-07  
**Tester:** User Testing Agent  
**Task:** 023

## Objectives

Validate Task 023 from a user's point of view: build the distributable, install it into an isolated user prefix, verify normal ADK installation still works, and execute the plan-first benchmark lifecycle through the installed `smaqit-adk` command. The deterministic control proves the generic process adapter, default human progress, JSONL agent-state stream, and artifact lifecycle without network dependencies. A second run uses an authenticated Codex CLI as a real agentic harness and verifies that it produces the requested output artifact while reporting application-readable state.

This is a pre-release test. It installs the binary produced by `installer/Makefile`; it does not invoke root `install.sh`, because that script downloads the latest published release and therefore cannot test the uncommitted Task 023 implementation.

## Prerequisites

- Run from the Task 023 worktree containing `src/bench`, `src/benchcli`, and `examples/bench`.
- Bash, Git, Go 1.24+, GNU Make, GNU Coreutils (`sha256sum`), `jq`, `/usr/bin/install`, and POSIX `sh` are available.
- At least 1 GiB free beneath `/tmp` for isolated workspaces and retained evidence.
- `codex` is on `PATH`; `codex login status` exits 0 and reports an authenticated account.
- Network access and model usage are authorized for the real-agent step. The deterministic control remains the diagnostic source if the external Codex service is unavailable.
- Run all commands in one Bash session. Do not delete `$SMAQIT_E2E_ROOT` until `test.complete` has captured the evidence.

Initialize the test session:

```bash
set -euo pipefail
export REPO_ROOT="$(git rev-parse --show-toplevel)"
test -d "$REPO_ROOT/src/bench"
test -d "$REPO_ROOT/examples/bench/process-harness"
export SMAQIT_E2E_ROOT="$(mktemp -d /tmp/smaqit-adk-e2e.XXXXXX)"
export SMAQIT_E2E_EVIDENCE="$SMAQIT_E2E_ROOT/evidence"
mkdir -p "$SMAQIT_E2E_EVIDENCE"
printf 'REPO_ROOT=%s\nSMAQIT_E2E_ROOT=%s\n' "$REPO_ROOT" "$SMAQIT_E2E_ROOT" | tee "$SMAQIT_E2E_EVIDENCE/session.env"
```

## Test Steps

### Step 1 — Build & Unit Test Gate

- [ ] Source-module tests exit 0 with zero failures:

  ```bash
  (cd "$REPO_ROOT/src" && go test ./...) 2>&1 | tee "$SMAQIT_E2E_EVIDENCE/src-tests.log"
  ```

- [ ] The Makefile build exits 0 and creates the distributable:

  ```bash
  (cd "$REPO_ROOT/installer" && make build) 2>&1 | tee "$SMAQIT_E2E_EVIDENCE/build.log"
  test -x "$REPO_ROOT/installer/dist/smaqit-adk-dev"
  ```

- [ ] The project test target exits 0, including source, installer-wrapper, black-box CLI, and structural tests:

  ```bash
  (cd "$REPO_ROOT/installer" && make test) 2>&1 | tee "$SMAQIT_E2E_EVIDENCE/make-test.log"
  ```

- [ ] The black-box and structural test module exits 0 against the freshly built binary. `SMAQIT_ADK_BIN` is required by its test entrypoint:

  ```bash
  (cd "$REPO_ROOT/tests" && SMAQIT_ADK_BIN="$REPO_ROOT/installer/dist/smaqit-adk-dev" go test ./unit/... ./structural/...) 2>&1 | tee "$SMAQIT_E2E_EVIDENCE/tests-module.log"
  ```

### Step 2 — Deploy & Start: Isolated CLI Installation

Task 023 adds a CLI rather than a daemon, so command discovery is its health check.

- [ ] Install the built distributable into an isolated user prefix and ensure this installed copy is selected:

  ```bash
  mkdir -p "$SMAQIT_E2E_ROOT/bin"
  /usr/bin/install -m 0755 "$REPO_ROOT/installer/dist/smaqit-adk-dev" "$SMAQIT_E2E_ROOT/bin/smaqit-adk"
  export PATH="$SMAQIT_E2E_ROOT/bin:$PATH"
  hash -r
  test "$(command -v smaqit-adk)" = "$SMAQIT_E2E_ROOT/bin/smaqit-adk"
  ```

- [ ] Installed-binary health checks exit 0 and expose `bench`:

  ```bash
  smaqit-adk --version | tee "$SMAQIT_E2E_EVIDENCE/installed-version.log"
  smaqit-adk --help | tee "$SMAQIT_E2E_EVIDENCE/top-level-help.log"
  smaqit-adk bench --help | tee "$SMAQIT_E2E_EVIDENCE/bench-help.log"
  grep -q 'bench' "$SMAQIT_E2E_EVIDENCE/top-level-help.log"
  for command in validate plan run grade compare report; do
    smaqit-adk bench "$command" --help > "$SMAQIT_E2E_EVIDENCE/help-$command.log"
    grep -q "Usage: smaqit-adk bench $command" "$SMAQIT_E2E_EVIDENCE/help-$command.log"
  done
  grep -q 'events' "$SMAQIT_E2E_EVIDENCE/help-run.log"
  ```

- [ ] Existing ADK installation behavior remains healthy through the installed binary:

  ```bash
  smaqit-adk lite "$SMAQIT_E2E_ROOT/lite-project" 2>&1 | tee "$SMAQIT_E2E_EVIDENCE/lite-install.log"
  test -f "$SMAQIT_E2E_ROOT/lite-project/.github/agents/smaqit.L2.agent.md"
  test -f "$SMAQIT_E2E_ROOT/lite-project/.github/skills/smaqit.create-agent/SKILL.md"
  test -f "$SMAQIT_E2E_ROOT/lite-project/.github/skills/smaqit.create-skill/SKILL.md"
  ```

### Step 3 — Deterministic Generic Process Harness E2E

- [ ] Copy the shipped process example so the test never writes results into the source worktree:

  ```bash
  cp -R "$REPO_ROOT/examples/bench/process-harness" "$SMAQIT_E2E_ROOT/control"
  ```

- [ ] Validate the manifest through the installed binary; JSON reports one case and one variant:

  ```bash
  smaqit-adk bench validate -json "$SMAQIT_E2E_ROOT/control/bench.yaml" | tee "$SMAQIT_E2E_EVIDENCE/control-validate.json"
  jq -e '.valid == true and .cases == 1 and .variants == 1' "$SMAQIT_E2E_EVIDENCE/control-validate.json"
  ```

- [ ] Create and inspect a saved plan before execution:

  ```bash
  smaqit-adk bench plan -json -out "$SMAQIT_E2E_ROOT/control.plan.json" "$SMAQIT_E2E_ROOT/control/bench.yaml" | tee "$SMAQIT_E2E_EVIDENCE/control-plan-command.json"
  jq -e '.planned == true and .runs == 1 and (.planId | length) == 64' "$SMAQIT_E2E_EVIDENCE/control-plan-command.json"
  jq -e '.schemaVersion == 1 and (.assets | length) >= 3 and (.runs | length) == 1' "$SMAQIT_E2E_ROOT/control.plan.json"
  ```

- [ ] Run the saved plan with its default human-readable progress. The terminal reports start, attempt, progress, completion, and the final evaluation summary:

  ```bash
  smaqit-adk bench run "$SMAQIT_E2E_ROOT/control.plan.json" | tee "$SMAQIT_E2E_EVIDENCE/control-human-progress.log"
  grep -q 'bench: run started' "$SMAQIT_E2E_EVIDENCE/control-human-progress.log"
  grep -q 'bench: attempt 1/1 started' "$SMAQIT_E2E_EVIDENCE/control-human-progress.log"
  grep -q 'bench: progress 1/1 attempt(s) complete' "$SMAQIT_E2E_EVIDENCE/control-human-progress.log"
  grep -q 'bench: run completed (evaluation-passed' "$SMAQIT_E2E_EVIDENCE/control-human-progress.log"
  grep -q 'evaluation-passed: generic-process-harness' "$SMAQIT_E2E_EVIDENCE/control-human-progress.log"
  ```

- [ ] Run the same saved plan as an application caller. JSONL reports ordered lifecycle state while the single-variant evaluation runs and passes:

  ```bash
  smaqit-adk bench run -events jsonl "$SMAQIT_E2E_ROOT/control.plan.json" | tee "$SMAQIT_E2E_EVIDENCE/control-events.jsonl"
  jq -s -e 'map(.type) == ["run.started", "attempt.started", "attempt.completed", "run.progress", "run.completed"]' "$SMAQIT_E2E_EVIDENCE/control-events.jsonl"
  jq -s -e 'map(.sequence) == [1,2,3,4,5] and all(.[]; .schemaVersion == 1 and (.timestamp | type) == "string" and (.timestamp | length) > 0) and .[0].totalAttempts == 1 and .[2].requiredPassed == true and .[3].completedAttempts == 1 and .[4].outcome == "evaluation-passed"' "$SMAQIT_E2E_EVIDENCE/control-events.jsonl"
  export CONTROL_EXPERIMENT_DIR="$(jq -rs '[.[] | select(.type == "run.completed")][-1].directory' "$SMAQIT_E2E_EVIDENCE/control-events.jsonl")"
  cp "$CONTROL_EXPERIMENT_DIR/experiment.json" "$SMAQIT_E2E_EVIDENCE/control-run.json"
  jq -e '.complete == true and .comparison.outcome == "evaluation-passed" and .comparison.winner == "posix-shell"' "$SMAQIT_E2E_EVIDENCE/control-run.json"
  cmp "$SMAQIT_E2E_EVIDENCE/control-events.jsonl" "$CONTROL_EXPERIMENT_DIR/events.jsonl"
  ```

- [ ] Verify the real process boundary and artifact contract:

  ```bash
  export CONTROL_RUN_DIR="$CONTROL_EXPERIMENT_DIR/runs/stdin-and-file-posix-shell-001"
  test -f "$CONTROL_EXPERIMENT_DIR/run-plan.json"
  test -f "$CONTROL_EXPERIMENT_DIR/resolved-manifest.json"
  test -f "$CONTROL_EXPERIMENT_DIR/comparison.json"
  test -f "$CONTROL_EXPERIMENT_DIR/report.md"
  test -f "$CONTROL_EXPERIMENT_DIR/events.jsonl"
  test -f "$CONTROL_RUN_DIR/request.json"
  test -f "$CONTROL_RUN_DIR/result.json"
  test -f "$CONTROL_RUN_DIR/traces/harness.stdout.log"
  test -d "$CONTROL_RUN_DIR/submission"
  grep -q 'named-input' "$CONTROL_RUN_DIR/traces/harness.stdout.log"
  jq -e '.status == "completed" and .requiredPassed == true and .usage.totalTokens == null and .usage.estimatedCost == null' "$CONTROL_RUN_DIR/result.json"
  jq -e 'has("expect") | not' "$CONTROL_RUN_DIR/request.json"
  test ! -e "$CONTROL_RUN_DIR/submission/.smaqit-bench-input"
  ```

- [ ] Hash immutable raw evidence, then regrade, recompute comparison, and render a new report without rerunning the harness:

  ```bash
  sha256sum "$CONTROL_RUN_DIR/request.json" "$CONTROL_RUN_DIR/result.json" "$CONTROL_RUN_DIR/traces/harness.stdout.log" "$CONTROL_RUN_DIR/traces/harness.stderr.log" > "$SMAQIT_E2E_EVIDENCE/control-raw-before.sha256"
  smaqit-adk bench grade -json "$CONTROL_EXPERIMENT_DIR" | tee "$SMAQIT_E2E_EVIDENCE/control-regrade.json"
  smaqit-adk bench compare -json "$CONTROL_EXPERIMENT_DIR" | tee "$SMAQIT_E2E_EVIDENCE/control-recompare.json"
  smaqit-adk bench report -json -format markdown "$CONTROL_EXPERIMENT_DIR" | tee "$SMAQIT_E2E_EVIDENCE/control-rereport.json"
  sha256sum "$CONTROL_RUN_DIR/request.json" "$CONTROL_RUN_DIR/result.json" "$CONTROL_RUN_DIR/traces/harness.stdout.log" "$CONTROL_RUN_DIR/traces/harness.stderr.log" > "$SMAQIT_E2E_EVIDENCE/control-raw-after.sha256"
  diff -u "$SMAQIT_E2E_EVIDENCE/control-raw-before.sha256" "$SMAQIT_E2E_EVIDENCE/control-raw-after.sha256"
  jq -e '.revision == 1 and .requiredFailures == 0' "$SMAQIT_E2E_EVIDENCE/control-regrade.json"
  jq -e '.outcome == "evaluation-passed"' "$SMAQIT_E2E_EVIDENCE/control-recompare.json"
  test -f "$CONTROL_EXPERIMENT_DIR/grades/revision-001.json"
  test -f "$CONTROL_EXPERIMENT_DIR/comparisons/revision-001.json"
  test -f "$CONTROL_EXPERIMENT_DIR/reports/report-001.md"
  ```

### Step 4 — Real Agent Harness E2E: Codex CLI

Exact agent input: create `result.txt` in the isolated attempt workspace with `BENCH_E2E_OK` as its only content, then return the same marker. Verification uses the Codex process exit, stdout, frozen submission, and experiment JSON.

- [ ] Confirm the live harness is installed and authenticated:

  ```bash
  command -v codex | tee "$SMAQIT_E2E_EVIDENCE/codex-path.log"
  codex --version | tee "$SMAQIT_E2E_EVIDENCE/codex-version.log"
  codex login status 2>&1 | tee "$SMAQIT_E2E_EVIDENCE/codex-auth.log"
  grep -q 'Logged in' "$SMAQIT_E2E_EVIDENCE/codex-auth.log"
  ```

- [ ] Create the exact real-agent manifest:

```bash
mkdir -p "$SMAQIT_E2E_ROOT/codex"
cat > "$SMAQIT_E2E_ROOT/codex/bench.yaml" <<'YAML'
schemaVersion: 1
name: codex-real-agent-smoke
cases:
  - id: create-marker-file
    given:
      prompt:
        text: |
          Create a file named result.txt in the current working directory.
          Its complete content must be exactly BENCH_E2E_OK followed by one newline.
          Do not create or modify any other file.
          When finished, respond exactly BENCH_E2E_OK.
    expect:
      - id: codex-exit
        type: runtime
        actual: exitCode
        exitCode: 0
      - id: final-response
        type: text
        actual: stdout
        operator: contains
        value: BENCH_E2E_OK
      - id: marker-file
        type: file
        actual: file:result.txt
        operator: content
        value: BENCH_E2E_OK
        trimFinalNewline: true
variants:
  - id: codex-cli
    adapter: process
    process:
      executable: codex
      arguments:
        - exec
        - --ephemeral
        - --ignore-user-config
        - --ignore-rules
        - --skip-git-repo-check
        - --sandbox
        - workspace-write
        - --color
        - never
        - "{task}"
      inputMode: argument
      environment:
        inherit:
          - PATH
          - HOME
          - CODEX_HOME
execution:
  repetitions: 1
  timeoutSeconds: 300
output:
  directory: ./results
YAML
```

- [ ] Validate and review the real-agent plan before spending model usage:

  ```bash
  smaqit-adk bench validate -json "$SMAQIT_E2E_ROOT/codex/bench.yaml" | tee "$SMAQIT_E2E_EVIDENCE/codex-validate.json"
  smaqit-adk bench plan -json -out "$SMAQIT_E2E_ROOT/codex.plan.json" "$SMAQIT_E2E_ROOT/codex/bench.yaml" | tee "$SMAQIT_E2E_EVIDENCE/codex-plan-command.json"
  jq -e '.planned == true and .runs == 1' "$SMAQIT_E2E_EVIDENCE/codex-plan-command.json"
  jq -e '.runs[0].caseId == "create-marker-file" and .runs[0].variantId == "codex-cli"' "$SMAQIT_E2E_ROOT/codex.plan.json"
  ```

- [ ] Run the saved plan against Codex with application-facing lifecycle events. Expected behavior: `attempt.started` is emitted before waiting for Codex, then exit 0 after Codex creates the marker file and returns `BENCH_E2E_OK`:

  ```bash
  smaqit-adk bench run -events jsonl "$SMAQIT_E2E_ROOT/codex.plan.json" | tee "$SMAQIT_E2E_EVIDENCE/codex-events.jsonl"
  jq -s -e 'map(.type) == ["run.started", "attempt.started", "attempt.completed", "run.progress", "run.completed"]' "$SMAQIT_E2E_EVIDENCE/codex-events.jsonl"
  jq -s -e 'map(.sequence) == [1,2,3,4,5] and all(.[]; .schemaVersion == 1 and (.timestamp | type) == "string" and (.timestamp | length) > 0) and .[1].caseId == "create-marker-file" and .[1].variantId == "codex-cli" and .[2].requiredPassed == true and .[4].outcome == "evaluation-passed"' "$SMAQIT_E2E_EVIDENCE/codex-events.jsonl"
  export CODEX_EXPERIMENT_DIR="$(jq -rs '[.[] | select(.type == "run.completed")][-1].directory' "$SMAQIT_E2E_EVIDENCE/codex-events.jsonl")"
  cp "$CODEX_EXPERIMENT_DIR/experiment.json" "$SMAQIT_E2E_EVIDENCE/codex-run.json"
  export CODEX_RUN_DIR="$CODEX_EXPERIMENT_DIR/runs/create-marker-file-codex-cli-001"
  jq -e '.complete == true and .comparison.outcome == "evaluation-passed" and .results[0].requiredPassed == true' "$SMAQIT_E2E_EVIDENCE/codex-run.json"
  grep -q 'BENCH_E2E_OK' "$CODEX_RUN_DIR/traces/harness.stdout.log"
  test "$(cat "$CODEX_RUN_DIR/submission/result.txt")" = 'BENCH_E2E_OK'
  jq -e '.harness.adapter == "process" and .harness.exitCode == 0 and .harness.timedOut == false' "$CODEX_RUN_DIR/result.json"
  cmp "$SMAQIT_E2E_EVIDENCE/codex-events.jsonl" "$CODEX_EXPERIMENT_DIR/events.jsonl"
  ```

- [ ] Verify the real-agent report and retain its harness traces for diagnosis:

  ```bash
  test -f "$CODEX_EXPERIMENT_DIR/report.md"
  grep -q 'evaluation-passed' "$CODEX_EXPERIMENT_DIR/report.md"
  cp "$CODEX_RUN_DIR/traces/harness.stdout.log" "$SMAQIT_E2E_EVIDENCE/codex-harness.stdout.log"
  cp "$CODEX_RUN_DIR/traces/harness.stderr.log" "$SMAQIT_E2E_EVIDENCE/codex-harness.stderr.log"
  ```

### Step 5 — Drift, Invalid Configuration, and Incomplete-Matrix Validation

- [ ] A changed input after planning is rejected with stable drift exit code `4`, before any harness starts. The agent-facing stream is one `run.failed` event rather than a misleading started attempt:

  ```bash
  cp -R "$REPO_ROOT/examples/bench/process-harness" "$SMAQIT_E2E_ROOT/drift"
  smaqit-adk bench plan -json -out "$SMAQIT_E2E_ROOT/drift.plan.json" "$SMAQIT_E2E_ROOT/drift/bench.yaml" > "$SMAQIT_E2E_EVIDENCE/drift-plan.json"
  printf '\nchanged-after-planning\n' >> "$SMAQIT_E2E_ROOT/drift/sample.txt"
  set +e
  smaqit-adk bench run -events jsonl "$SMAQIT_E2E_ROOT/drift.plan.json" > "$SMAQIT_E2E_EVIDENCE/drift-events.jsonl"
  DRIFT_EXIT=$?
  set -e
  test "$DRIFT_EXIT" -eq 4
  jq -s -e 'length == 1 and .[0].type == "run.failed" and .[0].failure.phase == "drift" and (.[0] | has("runId") | not)' "$SMAQIT_E2E_EVIDENCE/drift-events.jsonl"
  ```

- [ ] An unknown manifest field is rejected with stable invalid-configuration exit code `3` and an exact field path:

  ```bash
  cp "$SMAQIT_E2E_ROOT/control/bench.yaml" "$SMAQIT_E2E_ROOT/invalid.yaml"
  printf '\nunknownField: true\n' >> "$SMAQIT_E2E_ROOT/invalid.yaml"
  set +e
  smaqit-adk bench validate -json "$SMAQIT_E2E_ROOT/invalid.yaml" > "$SMAQIT_E2E_EVIDENCE/invalid-validation.json"
  INVALID_EXIT=$?
  set -e
  test "$INVALID_EXIT" -eq 3
  jq -e '.ok == false and .error.kind == "invalid-configuration" and (.diagnostics | any(.path == "unknownField"))' "$SMAQIT_E2E_EVIDENCE/invalid-validation.json"
  ```

- [ ] The same invalid manifest reports a single JSONL `run.failed` event for an invoking agent, including the validation diagnostic:

  ```bash
  set +e
  smaqit-adk bench run -events jsonl "$SMAQIT_E2E_ROOT/invalid.yaml" > "$SMAQIT_E2E_EVIDENCE/invalid-run-events.jsonl"
  INVALID_RUN_EXIT=$?
  set -e
  test "$INVALID_RUN_EXIT" -eq 3
  jq -s -e 'length == 1 and .[0].type == "run.failed" and .[0].failure.phase == "invalid-configuration" and (.[0].diagnostics | any(.path == "unknownField"))' "$SMAQIT_E2E_EVIDENCE/invalid-run-events.jsonl"
  ```

- [ ] A filtered multi-variant run exits `2` and remains explicitly inconclusive rather than selecting a winner:

  ```bash
  cp -R "$REPO_ROOT/examples/bench/multi-variant" "$SMAQIT_E2E_ROOT/subset"
  set +e
  smaqit-adk bench run -json -variant treatment "$SMAQIT_E2E_ROOT/subset/bench.yaml" > "$SMAQIT_E2E_EVIDENCE/subset-run.json"
  SUBSET_EXIT=$?
  set -e
  test "$SUBSET_EXIT" -eq 2
  jq -e '.complete == false and .comparison.complete == false and .comparison.outcome == "inconclusive"' "$SMAQIT_E2E_EVIDENCE/subset-run.json"
  ```

## Pass/Fail Criteria

**PASS** — Every checkbox is checked. All build, installation, health, deterministic-control, and real-Codex commands meet their stated exit codes and JSON/artifact assertions. The deterministic control exposes readable default progress plus an ordered, timestamped JSONL lifecycle that exactly matches its persisted event journal; Codex does the same while creating the exact marker file and producing `evaluation-passed`; lifecycle failures identify both drift and invalid configuration; regrading leaves raw evidence hashes unchanged; drift exits `4`; invalid configuration exits `3`; and a filtered matrix exits `2` as inconclusive.

**FAIL** — Any checkbox is unchecked; the installed binary is not the isolated copy; an unexpected command exit occurs; Codex is unauthenticated or cannot complete the marker task; expected output or evidence is missing; hidden input/oracle material leaks into a submission/request; raw evidence changes during regrading; or the stable exit-code assertions differ. If only the Codex step fails while the deterministic control passes, record the result as a real-harness/auth/network failure rather than a process-adapter failure, but the overall playbook still fails until resolved or explicitly rescoped.

## Evidence to Capture

- `$SMAQIT_E2E_EVIDENCE/session.env` and the retained `$SMAQIT_E2E_ROOT` path.
- Build and test logs: `src-tests.log`, `build.log`, `make-test.log`, and `tests-module.log`.
- Installed version and help outputs, plus `lite-install.log` and the isolated lite project tree.
- Deterministic control: validate/plan/run JSON, default human-progress transcript, streamed and persisted lifecycle JSONL, saved plan, full experiment directory, raw-evidence hashes, regrade/comparison/report revisions, and harness traces.
- Real Codex harness: CLI version/auth status, validate/plan/run JSON, streamed and persisted lifecycle JSONL, full experiment directory, frozen `result.txt`, report, and both copied harness traces.
- Negative paths: `drift-events.jsonl`, `invalid-validation.json`, `invalid-run-events.jsonl`, and `subset-run.json`, including observed exit codes.
- Tester notes identifying OS/architecture, Go version, Codex version, start/end timestamps, and any unexpected warnings.

This playbook certifies the locally built distributable. After Task 023 is included in a tagged release, run a separate distribution smoke using root `install.sh` with that exact `SMAQIT_ADK_VERSION` to certify GitHub release asset naming and download installation.
