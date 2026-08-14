# Repair Dogfood Bench Manifests Broken by Task 027

**Status:** PR Open
**PR:** #23
**Created:** 2026-08-13
**Started:** 2026-08-14
**Mode:** Assisted

## Description

Task 027's removal of `lite` and the Copilot `.agent.md` format broke 4 dogfood Bench manifests under `.smaqit/bench/` — one fails structural validation outright, three would fail live execution. Migrate all 4 onto the generic `given.files`/`given.directories` + `{input:<id>}` staging pattern the README already documents as canonical (rather than the special-cased `lite`-install trick three of them used), update expectations for dual Claude/Codex output, and re-verify structurally and live.

The 3 manifests using the `smaqit-adk-dev lite {workspace}` trick (`agents/smaqit.L2`, `skills/smaqit.create-agent`, `skills/smaqit.create-skill`) were always a documented special case — `.smaqit/bench/README.md`'s own canonical with/without-artifact example already shows the generic `given.directories` + `{input:<id>}` pattern instead. `smaqit.new-principle`'s manifest already follows that canonical pattern and needs only a one-line path fix. Since `lite` (a project-scoped install) has no replacement — global install only ever targets `$HOME` — the right fix isn't swapping the command, it's bringing all 3 manifests onto the pattern the other one and the README already establish.

A second, real design consideration: `create-agent`/`create-skill`'s skills now say "invoke `smaqit.L2` as a native subagent." A Bench manifest that did a real global install to make `smaqit.L2` spawnable would mutate the real machine's `~/.codex/agents/`/`~/.claude/agents/` on every run, cutting against Bench's disposable-workspace philosophy (and needs real credentials to authenticate, confirmed the hard way during Task 027). The resolution: reuse the read-and-follow-inline fallback Task 027 already designed for Copilot — stage L2's body alongside the routing skill, and tell Codex explicitly that L2 isn't registered as a spawnable agent in this sandbox, so it should read `{input:l2}` directly and apply it. `new-principle`'s manifest already does exactly this for L0, just with a stale file path — proof the pattern works, not something new to invent. Task 027's own live `codex exec` test already verified the *real* global-registration-and-spawn mechanism separately; this task doesn't need to re-prove that.

## Design Decisions

- **Bench vocabulary is authoritative.** A smaqit *Task* remains a tracked work item; Bench uses *Case* for an evaluation scenario, *Prompt* for `given.prompt`, *Case brief* for the rendered prompt plus declared inputs delivered to a harness, and *Harness workspace* for the writable execution directory.
- **Make a clean v2 breaking terminology migration.** Replace `{task}`/`{taskFile}` with `{brief}`/`{briefFile}`, `RunRequest.Task` with `RunRequest.CaseBrief`, and `Workspace.TaskFile` with `Workspace.BriefFile`. Bump manifest and plan schemas from 1 to 2 and the independent `request.json` schema to 2; reject legacy placeholders and require old manifests/plans to be migrated and regenerated. Historical smaqit task records and the platform's real `Task` subagent tool are excluded.
- **L2 conforms to the Bench case-brief contract.** When a case brief declares L2, a template, or rules as inputs, L2 reads those paths rather than consulting the developer's real global ADK installation. Global `~/.agents/smaqit-adk/` paths remain the normal fallback outside declared-input execution.
- **Compiled skill output is context-sensitive.** Consumer-project compilation writes both `.agents/skills/[name]/SKILL.md` and `.claude/skills/[name]/SKILL.md`; compiling an ADK-shipped skill in this source repository writes `skills/[name]/SKILL.md`.
- **Unify all 4 manifests onto one staging philosophy.** Every target and supporting artifact is a case-level `given.files`/`given.directories` input; each without-artifact variant removes all of those inputs during setup. No manifest performs a real global install or touches the developer machine's actual `~/.claude` or `~/.codex`.
- **L2 is never spawned as a real registered custom agent inside Bench tests.** Its body is staged and read inline. The real spawn mechanism was verified separately in Task 027.
- **Both Claude `.md` and Codex `.toml` agent outputs, and both project-local skill copies, are checked.** The create-skill validator runs against the Claude copy and the two compiled skill copies must be identical.
- **In-scope prompt corrections:** prompts name the relevant declared-input IDs in the case brief, rather than relying on prompt-body placeholder interpolation; `new-principle` states that L0 is intentionally unregistered in the isolated workspace, rather than claiming `codex exec` lacks subagent invocation.

## Implementation Steps

1. Migrate Bench's rendered harness input to the Case brief contract: rename the exported Go API, rendered filename and heading, process placeholders, stdin delivery, and request-evidence field from Task terminology to CaseBrief/Brief terminology.
2. Make the public contract a deliberate v2 break: update manifest/plan/request schemas, strict placeholder validation, migration diagnostics, and unit coverage for new placeholders, legacy-placeholder rejection, brief-file creation, stdin delivery, and request evidence.
3. Update all shipped examples, executable test fixtures, documentation, and changelog entries to the v2 Case/Prompt/Case-brief vocabulary; retain real smaqit task workflow and platform subagent-tool references.
4. Update `agents/smaqit.L2.md` so caller-provided outputs take precedence; consumer-project skill compilation creates the two project-local skill copies; ADK-source compilation retains its root `skills/` output; declared Case-brief inputs override global template/rule paths.
5. Fix `smaqit.new-principle`'s stale L0 source path; stage the skill, L0 body, and framework source as declared inputs; remove all three for the baseline; correct the isolated-workspace wording in its prompt.
6. Migrate `agents/smaqit.L2/bench.yaml` off the `lite`-install trick: stage L2, the base-agent template, and base rules per case; remove them for the baseline; update both cases' expectations and prompts for Claude/Codex dual output and declared-input use.
7. Migrate `skills/smaqit.create-agent/bench.yaml`: stage its routing skill, L2, the base-agent template, and base rules per case; remove all four for the baseline; use declared-input IDs in prompts and assert both compiled agent renders.
8. Migrate `skills/smaqit.create-skill/bench.yaml`: stage its routing skill, L2, the base-skill template, and skill rules per case; remove all four for the baseline; assert and compare both compiled skill copies and validate the Claude copy.
9. Align `.smaqit/bench/README.md` and `skills/smaqit.bench-scaffold/SKILL.md` with the Case-brief and declared-input contract, including the fact that prompt bodies name declared input IDs rather than receiving interpolation.
10. Run the focused Bench unit tests, example validation, and `smaqit-adk bench suite validate .smaqit/bench`; fix remaining structural failures until all four v2 manifests validate.
11. Run `smaqit-adk bench suite run .smaqit/bench` (live, authenticated `codex exec`, real cost); confirm no harness-level errors and diagnose pass/fail/inconclusive outcomes using the documented Bench conventions.
12. Update `.smaqit/compendium.md`'s Task-027-stale note to record the repaired suite and Case-brief terminology.

## Known Issues Triage

**Triaged:** 2026-08-14
**Tools searched:** Codex CLI
**Result:** Advisory

### Blocking Issues
- None.

### Advisory Issues
- None confirmed.

### Historical (Closed)
- None.

### Unresolvable Tools
- Codex CLI — the resolver returned unrelated repository `router-for-me/CLIProxyAPI`; no matching GitHub repository is available in the verified research map.

### Omitted Tools
- None.

### Search Warnings
- Codex CLI repository resolution — no GitHub issue search was run because the deterministic resolver returned an unrelated repository.

## Acceptance Criteria

- [x] Bench v2 uses Case/Prompt/Case-brief terminology throughout its public Go API, rendered harness input, placeholders, request evidence, examples, documentation, and executable fixtures; project-management tasks and platform `Task` tools remain unchanged
- [x] `{brief}` and `{briefFile}` work; legacy `{task}` and `{taskFile}` are rejected with a clear migration error; manifest and plan schemas are v2 and request evidence uses its v2 `caseBrief` field
- [x] L2 honors Case-brief declared inputs over global fallback paths and produces both project-local skill copies when caller-directed to compile for a consumer project
- [x] `smaqit-adk bench suite validate .smaqit/bench` reports `valid=true` for all four migrated v2 manifests
- [x] No migrated manifest references `.github/`, the `lite` command, `.agent.md` output, or legacy Task placeholders
- [x] Live `smaqit-adk bench suite run .smaqit/bench` executes without harness-level errors (pass/fail/inconclusive results are acceptable outcomes; crashes or timeouts are not)
- [x] `.smaqit/compendium.md`'s Task-027-stale note about these manifests is updated to reflect the repair

## Findings

[Populated by smaqit.task-complete. Do not fill in manually before task is complete.]

**Implementation approach:**
- Migrated Bench’s public delivery vocabulary to Case briefs and schema v2, then introduced declarative fixtures, shared inputs, and variant treatments for the dogfood suite.
- Updated L2’s caller-supplied source/output contract and migrated all four dogfood manifests, prompts, documentation, examples, and compendium guidance.

**Decisions made:**
- Replaced imperative setup-based artifact copying and baseline removal with variant-only treatment staging; raw prompts use Case-brief table IDs rather than interpolation.
- Kept historical migration references intact while enforcing the current contract in active manifests, examples, public APIs, and validations.

**Blockers encountered:**
- None; the authenticated live suite completed all four manifests with zero harness errors and zero timeouts.

**Follow-up identified:**
- The new-principle comparison tied because both variants met its current expectation; refine that scenario only if stronger treatment discrimination is needed.

## Files to Create / Modify

| File | Action |
|------|--------|
| `.smaqit/bench/agents/smaqit.L2/bench.yaml` | Modify — migrate off `lite` to `given.files` staging, dual-output expectations |
| `.smaqit/bench/agents/smaqit.L2/prompts/compile-base-agent.md` | Modify — `{input:<id>}` references, dual output paths |
| `.smaqit/bench/agents/smaqit.L2/prompts/reject-unresolved-placeholders.md` | Modify — dual output paths |
| `.smaqit/bench/skills/smaqit.create-agent/bench.yaml` | Modify — migrate off `lite` to `given.directories` staging + L2 inline-fallback, dual-output expectations |
| `.smaqit/bench/skills/smaqit.create-skill/bench.yaml` | Modify — same treatment, plus validator command path |
| `.smaqit/bench/skills/smaqit.new-principle/bench.yaml` | Modify — fix stale `l0agent` source path |
| `.smaqit/compendium.md` | Modify — update the Task-027-stale note once fixed |

## Notes

Source: follow-up item #2 from Task 027's own Findings (`.smaqit/tasks/027_migrate_to_global_user_level_installation.md`), planned via `smaqit.task-plan` in the same session Task 027 completed. Discovery for this plan was done by direct file reads (all 4 manifests, their prompts, and `.smaqit/bench/README.md`) rather than Explore subagents — the area is small and context was already fresh from implementing Task 027 itself.
