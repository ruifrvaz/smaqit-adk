# Bench Release Matrix

**Date:** 2026-08-15
**Session focus:** Complete Bench v2/release recovery, establish evaluation terminology, and plan scenario-matrix authoring.
**Tasks completed/referenced:** 029 completed; 030 completed earlier in the parent flow; 031 created; extensions Task 028 planned.

## Actions Taken

- Reviewed and completed the Task 029 / Task 030 Bench v2 work: Case-brief terminology, declarative fixtures/shared inputs/variant treatments, L2 source and output contracts, and four repaired dogfood manifests.
- Diagnosed the failed v2.0.0 post-merge release: cross-compilation variables caused the Go asset generator to be built for a target platform and then executed on the Linux runner.
- Released v2.0.1 after changing preparation to run host-native, moving tag creation behind the successful build matrix, and updating the release workflow's safety and action pinning.
- Confirmed the globally installed `smaqit-adk v2.0.1` and validated all four dogfood Bench manifests through the installed binary.
- Clarified Bench vocabulary: Case is an evaluation scenario, Variant is a named candidate configuration, Fixture is shared writable starting material, and a Case brief is the harness-delivered prompt plus staged-resource tables.
- Researched Promptfoo's prompt/provider/test matrix and recorded a parity assessment. Decided to use its matrix presentation as a design reference, not as a runtime dependency or broad feature-parity target.
- Created Task 031 to introduce a canonical Bench scenario matrix, generic scenario-first scaffolding, and neutral layout for prompt/model/harness comparisons.

## Problems Solved

- **Release cross-build failure:** `make prepare` inherited target `GOOS`/`GOARCH`; `generate-agents.go` consequently failed with an executable-format error. Host-native generation followed by target-only final builds fixes every release target.
- **Incomplete release safety:** the old workflow tagged before artifact builds. The repaired dependency graph creates a tag only after all binaries succeed, preventing another tagged-but-unpublished release.
- **Bench discoverability gap:** existing documentation explained schema mechanics but not which evaluation shapes Bench can express. The planned scenario matrix separates supported flattened Variant comparisons from future axis-aware analysis.

## Decisions Made

- Preserve `adk-v2.0.0` as an audit record and ship the pipeline repair as `adk-v2.0.1`.
- Keep Bench local-process and deterministic-evidence first. Provider adapters, model-graded scoring, interactive UI, and red-teaming are separate concerns.
- Treat prompt A/B as two variant treatments under one stable Case prompt; a suite cannot establish a winner across separate manifests.
- Use `scenarios/<id>` for generic and mixed comparisons while retaining `skills/<id>` and `agents/<id>` for artifact-only evaluation.

## Files Modified

- `.github/workflows/post-merge-release.yml`, `installer/Makefile`, `installer/main.go`, `CHANGELOG.md` — released v2.0.1 cross-build and release-order repair.
- `src/bench/`, `examples/bench/`, `docs/wiki/benchmarking.md`, root `README.md`, and root Bench skills — Task 029/030 Bench v2 delivery/treatment work and its verification coverage.
- `.smaqit/bench/` and `.smaqit/compendium.md` — repaired dogfood Bench suite and recorded the current evaluation contract.
- `agents/smaqit.L2.md`, `templates/skills/compiled/skill.rules.md`, `agents/smaqit.L1.md`, `skills/smaqit.create-skill/SKILL.md`, `AGENTS.md` — caller-provided source precedence and output-profile alignment.
- `docs/parity/promptfoo/` — created Promptfoo architecture, flow, feature-gap diagrams, and parity assessment; currently uncommitted task research provenance.
- `.smaqit/tasks/029_repair_dogfood_bench_manifests.md`, `.smaqit/tasks/PLANNING.md` — completed Task 029.
- `.smaqit/tasks/031_bench_scenario_matrix_generic_scaffolding.md`, `.smaqit/tasks/PLANNING.md` — created Task 031; currently uncommitted.
- `/home/ruifrvaz/projects/smaqit-extensions/.smaqit/tasks/028_benchmark_glossary_skill_invocation.md` — planned the canonical glossary-skill benchmark follow-up.

## Next Steps

1. Start Task 031 to implement the scenario matrix, README navigation, generic scaffold flow, and neutral scenario layout.
2. Start extensions Task 028 to benchmark and repair automatic glossary updates after clear definition questions.
3. Keep the Promptfoo assessment as reference material; assess any future engine-factor or model-provenance feature separately.

## Session Metrics

- **Tasks completed:** 1 owner task (029); Task 030 was completed in the parent flow.
- **Release outcome:** v2.0.1 published with five platform binaries.
- **Installed-Bench check:** 4 dogfood manifests structurally valid through the globally installed v2.0.1 binary.
- **Research outputs:** 1 Promptfoo parity assessment and 3 validated Mermaid diagrams.
- **New planned work:** Task 031 in smaqit-adk and Task 028 in smaqit-extensions.
