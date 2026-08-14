# Bench Scenario Matrix & Generic Scaffolding

**Status:** Not Started
**Created:** 2026-08-15

## Description

Make Bench's supported evaluation shapes discoverable and actionable. Add a canonical scenario matrix to the wiki, link it from the README, and generalize the Bench scaffold skill from artifact-only baseline comparisons to scoped prompt, model, harness, version, and factorial authoring.

Bench schema v2 already executes every in-scope shape through Variants, but its comparison report intentionally ranks flattened Variant IDs rather than analyzing first-class factors. This task improves authoring and documentation without claiming unsupported axis-aware reporting, remote-model provenance locking, semantic grading, provider integration, or red-teaming.

## Issue Triage Context

**Mode:** Auto
**Technologies:** Go Bench schema v2, YAML manifests, Markdown, SKILL.md
**Platforms/Environments:** Local process harnesses; Codex only when selected by the author
**Features/Integrations:** Promptfoo research reference only; no runtime integration
**Versions/Constraints:** Bench schemaVersion 2; preserve deterministic, local-first evidence model

## Design Decisions

- **Canonical matrix location:** `docs/wiki/benchmarking.md` owns the complete scenario matrix; `README.md` links directly to its anchor and does not duplicate it.
- **Comparison representation:** A Case has one shared prompt. Candidate prompts are variant treatments; model and harness candidates are explicit process-level Variant differences; factorials are named, flattened Variant tuples in one manifest.
- **Neutral placement:** Artifact-only evaluations remain under `skills/` or `agents/`; prompt, model, harness, version, and mixed comparisons use `.smaqit/bench/scenarios/<id>/`.
- **Scope boundary:** This is a documentation and authoring-contract task. First-class axes, matrix-completeness validation, factor-effect reports, remote-model provenance, semantic grading, and red-teaming require a later engine task.
- **Promptfoo relationship:** Promptfoo informs matrix presentation only. It is neither a dependency nor a feature-parity target; retain Bench's local process-harness, immutable-plan, and frozen-evidence model.

## Implementation Steps

1. Add a `Bench scenario matrix` section before lifecycle instructions in `docs/wiki/benchmarking.md`, covering single evaluation, artifact baseline/A-B, prompt A-B, model A-B, harness/version A-B, controlled factorial, and reliability/regression shapes.
2. Define comparison controls and boundaries in that section: one manifest per claimed comparison; fixed Case prompt, fixture, preparation, shared inputs, expectations, graders, budgets, and non-candidate environment; no cross-manifest winner; no oracle or grader candidate.
3. Correct the existing guidance that limits credible comparisons to treatment changes so it instead permits explicitly declared treatment, executable/version, process-argument/model-selector, or non-secret configured-environment candidate dimensions.
4. Add a direct `#bench-scenario-matrix` link from the Bench quick start in `README.md`.
5. Generalize `smaqit.bench-scaffold` into a scenario-first workflow: select dimensions before artifact discovery, render only applicable matrix rows, create a Case-row/Variant-column comparison card, inspect every candidate and selected harness, and use the Codex process block only for Codex harnesses.
6. Update scaffold placement, drafting, validation, examples, completion criteria, and failure handling for neutral scenarios plus correct prompt, model/harness, and flattened-factorial patterns.
7. Update the dogfood Bench README's layout convention to mention the neutral `scenarios/` namespace without duplicating the general scenario matrix.
8. Preserve the Promptfoo parity assessment as task research provenance, then run build, structural, documentation-link, example, and dogfood validation gates.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] `docs/wiki/benchmarking.md` contains a canonical scenario matrix covering single evaluation, artifact, prompt, model, harness, version, factorial, and reliability/regression shapes, with representation and fixed-control guidance.
- [ ] `README.md` links directly to the matrix from its Bench quick-start section without duplicating the full table.
- [ ] `smaqit.bench-scaffold` selects candidate dimensions before artifact discovery and presents only applicable scenario shapes.
- [ ] `smaqit.bench-scaffold` supports neutral `scenarios/<id>` placement for non-artifact and mixed comparisons while retaining `skills/` and `agents/` for artifact-only evaluation.
- [ ] `smaqit.bench-scaffold` defines correct prompt A-B, model/harness A-B, and flattened-factorial patterns plus explicit authoring guardrails.
- [ ] Documentation accurately distinguishes supported flattened Variant comparisons from unimplemented axis-aware analysis, remote-model provenance locking, semantic grading, and red-teaming.
- [ ] Installer build, test suite, example validation, documentation-link checks, and dogfood suite structural validation pass.

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
| `docs/wiki/benchmarking.md` | Modify — canonical scenario matrix, controls, boundaries, and corrected comparison wording |
| `README.md` | Modify — direct matrix link beside Bench quick start |
| `skills/smaqit.bench-scaffold/SKILL.md` | Modify — scenario-first generic authoring workflow and guardrails |
| `.smaqit/bench/README.md` | Modify — neutral scenario layout convention for dogfood Bench |
| `docs/parity/promptfoo/` | Create — Promptfoo research provenance and validated parity diagrams |

## Notes

Research: `docs/parity/promptfoo/ASSESSMENT.md` records Promptfoo's prompt/provider/test matrix as the presentation reference. Bench remains a generic local process-harness engine; a future engine task, not this one, would add first-class axes, matrix completeness checks, factor-aware reports, or provider provenance.
