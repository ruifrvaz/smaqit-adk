# HarnessBench Evaluation CLI

**Date:** 2026-08-10  
**Session focus:** Deliver, validate, document, and complete the configuration-first HarnessBench CLI.  
**Tasks completed/referenced:** 023 completed; 026 created and parked; 020 and 021 superseded.

## Actions Taken

- Reviewed NVIDIA's colocated skill-evaluation convention and mapped it to HarnessBench's manifest, process-adapter, deterministic-grading, and evidence model.
- Implemented and verified durable plain and JSONL lifecycle state for `bench run`, including the persisted `events.jsonl` journal.
- Ran deterministic process-harness checks and a live authenticated Codex E2E benchmark with lifecycle evidence, drift/error coverage, and immutable artifacts.
- Fixed temporary workspace cleanup after read-only staged inputs prevented `os.RemoveAll` from deleting benchmark directories.
- Added file-level comments to every HarnessBench Go source and test file.
- Created Task 026 to migrate the obsolete Copilot behavioural-evaluation suite to colocated HarnessBench skill and agent benchmarks; marked Tasks 020 and 021 obsolete.
- Committed the implementation in focused groups, completed Task 023, merged it into `main`, and removed its worktree and branch.

## Problems Solved

- Bench workspaces leaked under `/tmp/smaqit-bench-*` because input staging removed write permission from directories that cleanup later tried to remove. Cleanup now restores permissions before deletion and tests confirm no new leak.
- The CI workflow selected Go 1.23 even though the new source module requires Go 1.24. CI now selects Go 1.24.
- The prior repository compendium entry described an unimplemented installer-owned, Copilot-SDK design. It now describes the shipped `src`-owned generic-process architecture.

## Decisions Made

- Keep HarnessBench as a general, plan-first local process evaluation engine; Codex is a configured live harness, not an engine dependency.
- Use deterministic output and command graders for Phase 1 rather than a hidden LLM judge.
- Treat skill and agent benchmark activation as contract evaluation until native discovery is independently demonstrated.
- Park Task 026 after its design record was created; do not extend the legacy Copilot SDK suite.

## Files Modified

- `src/bench/*.go`, `src/benchcli/bench.go`, `src/go.mod`, `src/go.sum` — new benchmark engine, CLI, tests, and module metadata.
- `installer/main.go`, `installer/Makefile`, `installer/go.mod`, `installer/go.sum` — package the Bench command and run its checks.
- `tests/unit/bench_cli_test.go`, `.github/workflows/test-integration.yml` — black-box coverage and Go 1.24 CI execution.
- `examples/bench/*`, `docs/wiki/benchmarking.md`, `README.md` — runnable manifests and user documentation.
- `.smaqit/user-testing/tests/023_harnessbench-phase-1-smaqit-adk-bench-subcommand.md` — end-to-end test playbook.
- `.smaqit/tasks/020_lite_tier_behavioral_evals.md`, `.smaqit/tasks/021_advanced_tier_behavioral_evals.md`, `.smaqit/tasks/023_harnessbench_phase1_bench_subcommand.md`, `.smaqit/tasks/024_repair_broken_eval_artifact_references.md`, `.smaqit/tasks/026_harnessbench_skill_agent_evaluation_suite.md`, `.smaqit/tasks/PLANNING.md` — completion, supersession, and parked follow-up state.
- `.smaqit/references/project-research.md`, `.smaqit/compendium.md`, this history file — refreshed project knowledge.

## Next Steps

- Start Task 026 when ready to replace the Copilot SDK eval suite with colocated HarnessBench skill and agent benchmarks.
- Decide whether and how to demonstrate native skill discovery for the target local harness before claiming discovery coverage.

## Session Metrics

- Tasks completed: 1
- New CLI commands: 6
- Runnable benchmark examples: 3
- HarnessBench Go files documented: 20
- Focused implementation commits merged: 5 plus the merge commit
