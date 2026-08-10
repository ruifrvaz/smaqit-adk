# HarnessBench Skill and Agent Evaluation Suite

**Status:** In Progress
**Created:** 2026-08-10
**Started:** 2026-08-10
**Mode:** Assisted

## Description

Replace the repository's Copilot SDK behavioural-evaluation system with a HarnessBench suite run through `smaqit-adk bench`. The replacement evaluates real local harnesses, initially `codex exec` through the generic `process` adapter, and produces planned, reproducible run evidence instead of mutable Copilot-session reports.

Adopt NVIDIA's colocated `evals/` convention for skills and an equivalent namespace for flat agent files. Migrate the useful intent of the existing JSON evals into configuration-first benchmark cases, then remove the obsolete Copilot runner, its dependencies, credentials workflow, and documentation.

## Design Decisions

- **HarnessBench is the only behavioural-evaluation path:** Task 026 supersedes Tasks 020 and 021. The Copilot SDK suite is removed after equivalent replacement coverage is verified.
- **Configuration-first corpus:** Each case declares its task, visible prompt/spec/file/directory/image inputs, fixture, deterministic expectations, optional command graders, process variants, and output location in Bench YAML. Oracle values and grader assets remain outside staged inputs.
- **Dogfood data lives in `.smaqit/bench/`, not colocated in ADK-shipped source:** Skill suites live at `.smaqit/bench/skills/<skill-id>/`; agent suites live at `.smaqit/bench/agents/<agent-id>/`. This separates the ADK's shipped product source (`skills/`, `agents/`) from this repo's own self-benchmarking data, mirroring how a consumer project's `.smaqit/` holds its local ADK state rather than its ADK-shipped artifacts. (Supersedes the originally recorded "colocate under `skills/<skill-id>/evals/`" decision.)
- **Suite discovery/orchestration is a shipped engine capability, not repo-specific test infra:** The multi-manifest `suite` subcommand lives in `src/bench`/`src/benchcli` alongside the rest of the Bench CLI, reusable by any consumer project with multiple manifests — not a one-off tool under `tests/` or `.smaqit/`.
- **With/without-artifact comparison uses Variants, not Cases:** One Case, two Variant entries, differentiated via `Variant.Setup` commands — the without-artifact variant's setup removes/never-materializes the staged artifact. This reuses Bench's native `compare.go` grouping by `VariantID` with no new comparison code.
- **Codex is the reference live harness, not a Bench dependency:** A manifest invokes `codex exec` via `adapter: process`; the generic adapter remains usable for another local harness through configuration alone. No include/anchor mechanism exists across manifest files, so the reusable Codex process-variant block is a documented copy-paste YAML convention, not templated code.
- **Contract evaluation before native-discovery claims:** The with-artifact variant explicitly makes the target skill or agent available in its disposable workspace and directs the harness to use it. The without-artifact variant does neither. Native discovery is a separate capability that must be demonstrated before it is reported as covered.
- **Deterministic grading only:** Rewrite legacy transcript criteria as file, JSON, text, directory, runtime, and command expectations. Do not replace the old Copilot judge with a hidden LLM judge.
- **Explicit live execution; CI wiring deferred:** `make evals` runs the Bench suite deliberately; ordinary offline `test` and `test-all` do not consume model quota or require a logged-in CLI. CI gains only an offline manifest-validation step this task — a gated live-Codex CI job is explicitly deferred to a follow-up task, not built here.

## Implementation Steps

**Phase A — Inventory & layout doc**
1. Consolidate the legacy-eval classification: of 40 total criteria across 7 JSON cases, 16 map to deterministic graders (file/text/directory/command), 24 are conversational-only and are retired. Fold `smaqit.create-skill/002_reject_first_person_description.json`'s actual content (validator-gating) into the scenario already covered by `001_full_gathering_flow.json`; record that the "reject first-person description" scenario was never actually implemented. Document the mapping (e.g. `.smaqit/bench/MIGRATION.md`).
2. Write the benchmark-layout doc (extend `docs/wiki/benchmarking.md` or add `.smaqit/bench/README.md`): the `.smaqit/bench/skills/<id>/` / `.smaqit/bench/agents/<id>/` convention, the with/without-artifact one-Case-two-Variant pattern via `Variant.Setup`, the reusable `codex exec` process-variant YAML snippet, hidden-oracle placement, and `.smaqit/bench/runs/` output retention.

**Phase B — Suite capability (shipped engine feature, depends on Phase A doc)**
3. Add a `suite` subcommand to `src/benchcli/bench.go` + supporting code in `src/bench/` (e.g. `suite.go`): walks a root directory for `bench.yaml` files, runs the existing `LoadManifest → BuildPlan → RunPlanWithOptions` pipeline per manifest (reusing `benchRun`'s pattern), forwards per-manifest lifecycle events, and produces a new aggregate suite-result type. Support `validate`/`plan`/`run` at the suite level.
4. Add deterministic fake-harness tests for the suite capability using the existing `mock` adapter (`src/bench/suite_test.go`): discovery, the with/without-artifact `Setup`-command boundary, grading, and the aggregate comparison result.

**Phase C — Dogfood manifests (depends on Phase B)**
5. Create `.smaqit/bench/skills/smaqit.create-agent/bench.yaml`, `.smaqit/bench/skills/smaqit.create-skill/bench.yaml`, `.smaqit/bench/agents/smaqit.L2/bench.yaml` (+ fixtures/oracle assets), migrating retained scenarios from step 1 with deterministic `Expect` blocks and with/without-artifact Variant pairs. At least one skill and one agent manifest gets a real `codex exec` process variant.
6. Run the new suite locally against the authenticated Codex CLI as the explicit live test (not wired into CI — see Design Decisions).

**Phase D — Wiring and cleanup (depends on Phase C passing)**
7. `installer/Makefile`: point `evals` at `smaqit-adk-dev bench suite .smaqit/bench/`; drop `evals` from `test-all`'s dependency (`test-all: test`) so offline `test`/`test-all` never run it. Add an offline validation step for `.smaqit/bench/**/bench.yaml` to `.github/workflows/test-integration.yml` (extends the existing `make test-bench-examples` pattern) — no live execution, no new secrets.
8. Update `tests/README.md` and `.github/copilot-instructions.md`'s architecture table to note `.smaqit/bench/` as this repo's dogfood-suite location, distinct from ADK-shipped source.
9. Remove `tests/evals/runner/`, `tests/evals/host-setup.sh`, the legacy JSON corpus, `tests/evals/runs/`, `tests/evals/README.md`, and the Copilot SDK Go dependency (`tests/go.mod`/`go.sum`, `go mod tidy`).
10. Verify/apply the obsolete marker on Tasks 020 and 021's own task files (PLANNING.md already lists them as Abandoned; confirm the individual files match).

## Known Issues Triage
**Triaged:** 2026-08-10
**Tools searched:** Codex CLI (openai/codex)
**Result:** Advisory

### Advisory Issues
- [#36570 exec: approvals_reviewer = "auto_review" silently defeats an explicit --sandbox level](https://github.com/openai/codex/issues/36570) — `openai/codex` — opened 2026-08-02 — bug, sandbox, exec, CLI, config. No platform keyword is stated in this task, so this stays Advisory rather than Blocking per triage rules — but it's directly relevant: Bench's `process` adapter has no PTY/approval-relay, so if `codex exec`'s approval/sandbox mode isn't pinned via CLI flags, an unresolved approval prompt would hang until Bench's own run timeout kills it, turning Phase C's live benchmark into a timeout rather than a clean pass/fail. Resolve by explicitly setting a non-interactive sandbox/approval flag (e.g. full-auto / danger-full-access equivalent) in the reusable Codex process-variant block from Implementation Step 2.
- [#28476 Codex exec stalled after turn.started](https://github.com/openai/codex/issues/28476) — `openai/codex` — opened 2026-06-16 — bug, exec, CLI. A known stall/hang report for `codex exec`. Mitigated by Bench's manifest-level `Execution.TimeoutSeconds`, but worth setting deliberately (not relying on the 300s default) for the live-Codex manifests in Phase C so a stall reads as a bounded timeout, not an indefinite hang.
- [#36562 codex exec --json drops typed codexErrorInfo from terminal errors](https://github.com/openai/codex/issues/36562) — `openai/codex` — opened 2026-08-02 — bug, exec, CLI. Only relevant if a manifest parses `codex exec --json` output for grading; if Phase C uses plain-text/file-based deterministic graders instead (as the task's design decisions specify), this doesn't affect the plan.

### Unresolvable Tools
- (none — Codex CLI resolved directly via `openai/codex`; GitHub Copilot SDK was not searched since Task 026 removes it rather than depending on new behavior from it)

## Acceptance Criteria

- [ ] Every retained legacy scenario has a documented HarnessBench replacement or a recorded reason it was retired as Copilot-implementation-specific
- [ ] Skill benchmarks live under `.smaqit/bench/skills/<skill-id>/`; agent benchmarks live under `.smaqit/bench/agents/<agent-id>/` (this repo's own dogfood-suite data, not the ADK-shipped `skills/`/`agents/` trees)
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
