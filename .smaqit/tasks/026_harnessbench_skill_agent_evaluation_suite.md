# HarnessBench Skill and Agent Evaluation Suite

**Status:** Not Started
**Created:** 2026-08-10

## Description

Replace the repository's Copilot SDK behavioural-evaluation system with a HarnessBench suite run through `smaqit-adk bench`. The replacement evaluates real local harnesses, initially `codex exec` through the generic `process` adapter, and produces planned, reproducible run evidence instead of mutable Copilot-session reports.

Adopt NVIDIA's colocated `evals/` convention for skills and an equivalent namespace for flat agent files. Migrate the useful intent of the existing JSON evals into configuration-first benchmark cases, then remove the obsolete Copilot runner, its dependencies, credentials workflow, and documentation.

## Design Decisions

- **HarnessBench is the only behavioural-evaluation path:** Task 026 supersedes Tasks 020 and 021. The Copilot SDK suite is removed after equivalent replacement coverage is verified.
- **Configuration-first corpus:** Each case declares its task, visible prompt/spec/file/directory/image inputs, fixture, deterministic expectations, optional command graders, process variants, and output location in Bench YAML. Oracle values and grader assets remain outside staged inputs.
- **Colocate skills; namespace agents:** Skill suites live at `skills/<skill-id>/evals/`. Since agents are flat files, their suites live at `agents/evals/<agent-id>/`.
- **Codex is the reference live harness, not a Bench dependency:** A manifest invokes `codex exec` via `adapter: process`; the generic adapter remains usable for another local harness through configuration alone.
- **Contract evaluation before native-discovery claims:** The with-artifact variant explicitly makes the target skill or agent available in its disposable workspace and directs the harness to use it. The without-artifact variant does neither. Native discovery is a separate capability that must be demonstrated before it is reported as covered.
- **Deterministic grading only:** Rewrite legacy transcript criteria as file, JSON, text, directory, runtime, and command expectations. Do not replace the old Copilot judge with a hidden LLM judge.
- **Explicit live execution:** `make evals` runs the Bench suite deliberately; ordinary offline `test` and `test-all` do not consume model quota or require a logged-in CLI.

## Implementation Steps

1. Inventory the seven legacy JSON cases and classify each expected outcome into a Bench task, fixture/input, deterministic oracle, and command grader. Retire criteria that only asserted an implementation-specific Copilot conversation rather than an observable skill or agent outcome.
2. Define and document the benchmark layout, case naming, hidden-oracle rules, output retention, and process-harness contract. Add a repository-level suite runner that discovers the configured manifests, validates/plans/runs them, and forwards Bench lifecycle events.
3. Add reusable Codex process-variant configuration and deterministic fake-harness integration coverage. Materialize each selected skill/agent into the disposable workspace without staging oracle assets; configure the no-artifact variant as the valid baseline.
4. Create benchmark cases for every retained legacy scenario under the new skill and agent locations. Include at least one skill and one agent case with a real Codex run, using fixture-based file-output graders rather than transcript judging.
5. Replace the installer targets and test documentation: `make evals` invokes the Bench suite; ordinary test targets remain offline. Update CI to validate benchmark manifests and run live benchmarks only in an explicitly authorised environment.
6. Once replacement coverage passes, remove `tests/evals/runner/`, `tests/evals/host-setup.sh`, the legacy JSON corpus, the Copilot SDK module dependency, obsolete run-report handling, and all Copilot-token instructions. Mark Tasks 020 and 021 obsolete in their records and the planning index.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] Every retained legacy scenario has a documented HarnessBench replacement or a recorded reason it was retired as Copilot-implementation-specific
- [ ] Skill benchmarks are colocated under `skills/<skill-id>/evals/`; agent benchmarks live under `agents/evals/<agent-id>/`
- [ ] Each benchmark uses a strict Bench manifest with declared inputs, fixture, output directory, and hidden deterministic oracle/grader assets
- [ ] A with-artifact and without-artifact comparison is available for each applicable skill or agent benchmark
- [ ] The suite runner emits or preserves Bench lifecycle events and supports validate, plan, and run without custom Copilot credentials
- [ ] Deterministic fake-harness tests cover the suite runner, artifact staging boundary, grading, and comparison result
- [ ] At least one skill and one agent benchmark execute successfully against the authenticated local Codex CLI as an explicit live test
- [ ] `make evals` runs the HarnessBench suite; offline `test` and `test-all` do not run live model workloads
- [ ] The Copilot SDK runner, legacy JSON eval corpus, Copilot SDK Go dependency, Copilot-token setup, and obsolete documentation are removed after migration verification
- [ ] Tasks 020 and 021 are marked obsolete as superseded by Task 026

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
|---|---|
| `skills/*/evals/` | Create benchmark manifests, fixtures, visible inputs, hidden oracles, and graders |
| `agents/evals/` | Create equivalent benchmark suites for flat agent artifacts |
| `tests/bench/` or equivalent suite-runner location | Create manifest discovery and deterministic integration tests |
| `installer/Makefile` | Modify — replace the Copilot eval target with the explicit HarnessBench suite |
| `tests/go.mod`, `tests/go.sum` | Modify — remove Copilot SDK dependency after runner removal |
| `tests/evals/runner/` | Delete after replacement verification |
| `tests/evals/host-setup.sh` | Delete after replacement verification |
| `tests/evals/**/*.json` | Delete after each retained scenario is migrated |
| `tests/README.md`, `tests/evals/README.md`, `.github/workflows/*` | Modify — remove obsolete Copilot guidance and document Bench execution |
| `.smaqit/tasks/020_lite_tier_behavioral_evals.md`, `.smaqit/tasks/021_advanced_tier_behavioral_evals.md`, `.smaqit/tasks/PLANNING.md` | Modify — record supersession |

## Notes

Task 023 supplies the core Bench CLI and generic process adapter. This task depends on that implementation being accepted before it starts. Existing Copilot eval reports are historical evidence only and are not migrated as runtime artifacts.
