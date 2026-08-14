# Promptfoo Parity Assessment

Source: [promptfoo/promptfoo](https://github.com/promptfoo/promptfoo) and [Promptfoo evaluation configuration](https://www.promptfoo.dev/docs/configuration/guide/)  
Studied: 2026-08-15 · package version 0.122.0

## What Promptfoo is

Promptfoo is a TypeScript/Node.js CLI and library for evaluating LLM prompts, models, agents, and RAG applications. Its YAML configuration composes prompts, providers, and test inputs into an evaluation matrix, then applies optional assertions and metrics. It also provides provider integrations, interactive/CI-oriented reports, and separate red-team workflows for adversarial security testing. [Its CLI registers evaluation, red-team, generation, and reporting commands.](https://raw.githubusercontent.com/promptfoo/promptfoo/main/src/main.ts)

Key properties:

- **Configuration model** — prompts, providers, and tests combine into a visible matrix; array-valued variables expand input combinations.
- **Evaluation surface** — provider adapters and application targets return outputs that deterministic, semantic, or model-graded assertions evaluate.
- **Reporting/security** — supports interactive viewing and export formats, plus a separate red-team workflow.

---

## Structural Mapping

| Concept | Promptfoo implementation | smaqit-adk equivalent | Mapping quality |
|---------|--------------------------|-----------------------|-----------------|
| Candidate/input matrix | `prompts` x `providers` x `tests`, with variable expansion | Cases x Variants x repetitions | partial |
| Candidate execution | Provider adapters and application targets | `process` adapter launching any local harness | partial |
| Test contract | Test variables and assertions | Case prompt, fixture, shared given inputs, deterministic expectations | partial |
| Quality evaluation | Deterministic, embedding, custom-code, and model-graded assertions | Deterministic expectations and optional command graders | partial |
| Reproducibility | YAML configuration and result exports | Hashed plans, immutable run evidence, frozen submissions, re-grade/reference-drift checks | a-only |
| Result presentation | Interactive viewer, HTML, JSON, CSV, JUnit | Markdown report, JSON evidence, event stream, traces | partial |
| Security testing | Red-team probe generation, scanning, risk reporting | None | absent |

Promptfoo's `scenarios` feature groups variable sets with reusable tests, while Bench's **Case** remains one complete evaluation scenario. They are related but not interchangeable: Promptfoo expands a declarative input matrix; Bench executes an explicitly declared run matrix and preserves the run evidence.

---

## Relationship Options

### Option A — Direct integration

Run Promptfoo as a Bench process harness or command grader for provider/API-oriented evaluations.

**Pros:**

- Reuses Promptfoo's provider catalog, model-graded assertions, and HTML reports.
- Allows a Bench experiment to treat a Promptfoo configuration as one candidate execution environment.

**Cons:**

- Duplicates both systems' planning, matrix expansion, scoring, and evidence concepts.
- Adds a Node.js/runtime dependency and makes a local process Bench depend on API credentials and Promptfoo configuration semantics.

**Verdict:** Viable for a consuming project with an existing Promptfoo estate, but not as a foundational dependency of smaqit-adk Bench.

### Option B — Native feature parity

Add Promptfoo-equivalent provider adapters, model-graded metrics, interactive reports, and red-teaming to Bench.

**Pros:**

- One evaluation tool and one evidence model.

**Cons:**

- Turns a process-harness evaluation engine into a broad LLM platform, duplicating a mature dedicated project.
- Weakens Bench's local, explicit, deterministic default by making provider, credential, and judge-model policy part of its core.

**Verdict:** Reject as a general goal. Only improve capabilities that strengthen Bench's distinct process-harness contract.

### Option C — Use as a design reference

Adopt Promptfoo's user-facing matrix vocabulary and scenario selection pattern while retaining Bench's execution/evidence design.

**Verdict:** Recommended. It directly addresses the current discoverability gap without importing unrelated provider, UI, or security scope.

---

## Recommendation

**Option C — use Promptfoo as a design reference for a scoped Bench scenario matrix.**

Promptfoo makes the candidate dimensions and test-input dimensions visible before execution. Bench should copy that presentation principle, not its product surface:

1. Document a canonical scenario matrix in `docs/wiki/benchmarking.md`: single evaluation; artifact A/B; prompt A/B; model A/B; harness A/B; version comparison; and controlled factorial combinations.
2. Add a compact README link to that matrix next to the Bench quick start.
3. Make `smaqit.bench-scaffold` select the dimensions that vary before drafting a manifest, then show only the matching matrix rows and invariants.
4. Keep a single manifest for every claimed comparison; a Bench suite aggregates independent manifests and does not compare across them.
5. Preserve fixed Cases, fixtures, shared inputs, expectations, budgets, and graders across candidate variants. Only declared candidate dimensions may vary.

This keeps Bench focused on reproducible local harness experiments, while making model/prompt/artifact/harness evaluation shapes obvious to an author.

---

## Parity Roadmap

| Priority | Feature | Promptfoo reference | smaqit-adk task / component | Status | Benefit |
|----------|---------|---------------------|-----------------------------|--------|---------|
| 1 | Scoped scenario matrix | [Configuration guide](https://www.promptfoo.dev/docs/configuration/guide/) | `docs/wiki/benchmarking.md`, `README.md` | planned | Makes the legal comparison axes and controls discoverable. |
| 2 | Scenario-first scaffolding | [Scenario configuration](https://www.promptfoo.dev/docs/configuration/scenarios/) | `skills/smaqit.bench-scaffold/SKILL.md` | planned | Lets authors choose model, prompt, artifact, harness, or factorial evaluation before YAML is drafted. |
| 3 | Neutral scenario layout | Promptfoo external scenario files | Bench project layout convention | planned | Supports non-artifact benchmarks without mislabeling them as skills or agents. |
| 4 | Semantic/model-graded metrics | [Assertions and metrics](https://www.promptfoo.dev/docs/configuration/expected-outputs/) | unplanned | Valuable for subjective quality, but conflicts with Bench's deterministic default and needs an explicit cost/reproducibility policy. |
| 5 | Red-team workflows | [Red-team overview](https://www.promptfoo.dev/docs/red-team/) | unplanned | Separate security-testing product area; do not absorb into this documentation task. |

---

## Project A Advantages

| smaqit-adk Bench capability | Promptfoo equivalent |
|-----------------------------|--------------------|
| Any trusted local process can be the harness, including CLI agents and repository workflows. | Provider adapters and application targets; not the same general local-harness contract. |
| Plans hash the manifest, executable, fixture, inputs, treatments, and oracle references before execution. | No directly equivalent plan/reference-drift contract established by this assessment. |
| Frozen submissions, revisioned grades, and re-derivation preserve an audit trail without re-running candidates. | Result export exists, but not the same evidence/re-grade model. |
| Read-only shared inputs and variant treatments are explicitly excluded from submissions and repository metrics. | No directly equivalent staged sidecar plane established by this assessment. |

## Sources consulted

- [Promptfoo repository README](https://github.com/promptfoo/promptfoo)
- [Configuration guide](https://www.promptfoo.dev/docs/configuration/guide/)
- [Scenario configuration](https://www.promptfoo.dev/docs/configuration/scenarios/)
- [Assertions and metrics](https://www.promptfoo.dev/docs/configuration/expected-outputs/)
- [Output formats](https://www.promptfoo.dev/docs/configuration/outputs/)
- [Red-team overview](https://www.promptfoo.dev/docs/red-team/)
