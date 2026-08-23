# Compilation Log: smaqit.new-template

**Compilation timestamp:** 2026-08-20
**Compiled by:** Agent-L2 (smaqit.L2)
**Pattern:** Pattern 4 — Skill Compilation (3-way merge)

## L1 Sources Read

- `agents/smaqit.L2.md` — role, directives, "For Skills" compilation procedure (steps 1-8), "Section-Level Compilation" guidance
- `templates/skills/base-skill.template.md` — structural template with `[PLACEHOLDER]` markers (frontmatter + `## Steps`/`## Output`/`## Scope`/`## Examples`/`## Gotchas`/`## Completion`/`## Failure Handling`)
- `templates/skills/compiled/skill.rules.md` — placeholder catalog, Degrees of Freedom directive, Conciseness directive, Reference Structure directive, Base Failure Handling Pattern table, "Compilation Guidance for Agent-L2" (steps 1-9)

All three sources were read from the local worktree copies (`templates/...` relative to the worktree root), per the case instruction that this repository's own `templates/` tree is the ADK source of truth being shipped — not the globally-installed `~/.agents/smaqit-adk/` copies.

## Definition File

- `.smaqit/definitions/skills/smaqit.new-template.md` (tier: advanced) — Identity, Steps (with fragility annotations), Output, Scope, Completion, Failure Scenarios, Gotchas, Examples, Compatibility. All required sections present (identity, steps with fragility levels, output, scope, completion, failure handling); no gaps requiring a stop-and-request.

## Merge Process Summary

- `[SKILL_NAME]` → `smaqit.new-template` (from Identity)
- `[SKILL_DESCRIPTION]` → definition's draft description rewritten into the compiled form: "Adds or refines a Level 1 template structure file in templates/agents/ or templates/skills/ — delegates gathering, validation, and authoring to smaqit.L1, then cascades the change by recompiling every registered dependent agent or skill via smaqit.L2." Verified capability-oriented, present-tense, declarative, states what the skill produces, no conversational/user-centric gating, no first person, no can/may/might/will. Backticks dropped from the frontmatter description to match the plain-scalar YAML style already established by `skills/smaqit.new-principle/SKILL.md` and the sibling `skills/smaqit.new-rules/SKILL.md` compiled the same day.
- `[SKILL_VERSION]` → `"0.1.0"` (from Identity — used the definition's stated version directly, since the "default 1.0.0" rule in skill.rules.md's Compilation Guidance only applies when a version is unspecified)
- `[SKILL_TITLE]` → "New Template"
- `[STEPS_CONTENT]` → definition's 4 steps, degrees-of-freedom applied per fragility marking: Step 1 (low fragility) as prose under a `### 1. Invoke smaqit.L1` heading, with the exact instruction text passed to the subagent rendered as a blockquote; Step 2 (high fragility) rendered with the literal `smaqit-adk registry get --scope base|extension --depends-on <path>` commands preserved verbatim in fenced code blocks, no paraphrasing of the base/extension scope distinction or the "missing registry file = zero dependents" rule; Step 3 (medium fragility) kept as directive prose describing the pattern (single-hop `Task` call, mirrors `smaqit.development`'s orchestration, never nested inside `smaqit.L1`) rather than literal pseudocode, since the invocation shape varies by dependent type; Step 4 (low fragility) as a single prose instruction. Parenthetical fragility labels themselves were stripped (compile-time signal for Agent-L2, not runtime instruction for the executing agent).
- `[OUTPUT_CONTENT]` → definition Output section, carried through near-verbatim as three bullets (template file changed, dependents recompiled, cascade summary)
- `[SCOPE_CONTENT]` → definition Scope section, condensed to one paragraph naming the two redirect skills (`smaqit.new-rules`, `smaqit.new-principle`) and the registry-format boundary
- `[EXAMPLES_CONTENT]` → definition's single example, **reworded** (see Validation Fix below)
- `[GOTCHAS_CONTENT]` → definition's 2 gotchas, verbatim (already concise, no trimming needed)
- `[COMPLETION_CONTENT]` → definition Completion checklist, verbatim
- `[FAILURE_HANDLING_CONTENT]` → see Failure Handling Merge below
- `compatibility` frontmatter → included (definition specifies: "Requires the `smaqit-adk` binary on `PATH` for the registry subcommands used in Step 2."), backticks dropped to match plain-scalar YAML style, flat top-level field placed after `metadata`
- `allowed-tools` frontmatter → omitted; definition does not specify a value

## Failure Handling Merge

Base Failure Handling Pattern (4 rows, from skill.rules.md) merged against the definition's 6 Failure Scenarios rows:

- Base "Subagent invocation fails" is textually identical to the definition's own "Subagent invocation fails" row (both: "Report the failure with context; do not silently retry") — deduped to a single row.
- Base "Required input not provided" and "Gathered input is ambiguous" have no direct definition counterpart but remain applicable at the orchestration level (this skill must have a clear instruction to hand `smaqit.L1`, and a registry response could be malformed/ambiguous) — kept as foundation rows per skill.rules.md's "Insert the following as the foundation of `[FAILURE_HANDLING_CONTENT]`" instruction.
- Base "Output artifact already exists" has no exact definition counterpart but is genuinely applicable: the Step 3 cascade recompiles pre-existing dependent agent/skill files via `smaqit.L2` — kept as a foundation row rather than dropped as inapplicable.
- Remaining 5 definition rows (`smaqit.L1` not installed, registry file missing/unreadable, a dependent's recompile fails, user wants rules/principles instead, user abandons midway) had no base overlap — carried through unchanged.

Result: 9-row single coherent table (4 base − 1 dedupe + 6 definition = 9), header `Situation | Action` matching the base table's header (definition used `Situation | Response`, normalized to `Action`).

## Validation Fix

Ran the repo's skill validator (`~/.agents/skills/smaqit.create-skill/scripts/validate-skill.go`, installed globally, invoked with `go run`) against the first draft as an extra check beyond this procedure's own step 6. It reported one violation: `unresolved placeholder "[TYPE_SPECIFIC_RULES]" on body line 57`. The definition's own Example text ('User says "add a `[TYPE_SPECIFIC_RULES]` placeholder to the base agent template."') was carried through verbatim in the first draft; the validator's bracket-token heuristic cannot distinguish a genuine unresolved compile-time placeholder from a bracketed string quoted as illustrative user input. Fixed by rewording the Example's Input line to describe the same scenario without bracket syntax ("add a new placeholder section to the base agent template for domain-specific completion criteria"), preserving the Output line unchanged. Re-ran the validator: passed, along with all 6 other already-compiled skills in `skills/` (confirmed as a sanity check that the validator itself was not newly broken).

## Conciseness Filter

Removed the definition's inline parenthetical fragility labels from the compiled Steps. No other trims were needed; the definition's prose was already concise and each remaining sentence (exact registry commands, the `smaqit.development` orchestration analogy, the no-nested-Task rule) carries information the executing agent could not safely infer.

## Progressive Disclosure

Not needed — compiled body is 90 lines, well under the 400-line threshold. No `references/` subdirectory created.

## Validation Checklist

- [x] No unresolved `[PLACEHOLDER]`-style uppercase-bracket tokens (`grep -n` for all skill.rules.md placeholder names → no matches)
- [x] Frontmatter `name` field exactly `smaqit.new-template`
- [x] Description is capability-oriented, declarative, present-tense; no "Use when...", no first person, no can/may/might/will
- [x] All required sections present, in template order: `## Steps`, `## Output`, `## Scope`, `## Examples`, `## Gotchas`, `## Completion`, `## Failure Handling`
- [x] `compatibility` frontmatter field present (sourced from definition); `allowed-tools` omitted (not specified)
- [x] No nested reference chains — skill body references `~/.claude/agents/smaqit-L1.md` (an external agent file, read inline for one turn on Copilot only) but does not reference any file bundled within its own skill directory, so no `SKILL.md` → file → file chain exists
- [x] Body under 400 lines (90 lines)
- [x] Third-party validator (`validate-skill.go`) passes with zero violations after the Example rewording

## Decisions Made

1. Used `### N. Title` subheadings for the four Steps (matching the style already established in `skills/smaqit.bench-run/SKILL.md`), rather than the flat numbered-list style used by `skills/smaqit.new-principle/SKILL.md` and the sibling `skills/smaqit.new-rules/SKILL.md` compiled the same day. Both are valid renderings of `[STEPS_CONTENT]` per skill.rules.md, which does not mandate a specific list form — flagging the stylistic divergence for awareness, not resolving it here.
2. Reworded the Example's Input line to remove bracket syntax that the repo's validator flags as an unresolved placeholder, while preserving the scenario's meaning (adding a placeholder-bearing section to a template) from the definition file.
3. Placed `compatibility` as a flat frontmatter field (matching `skills/smaqit.new-rules/SKILL.md`'s same-day compilation and skill.rules.md's literal placeholder-catalog description) rather than as body prose under the title (the style used by the earlier `skills/smaqit.bench-run/SKILL.md` and `skills/smaqit.bench-scaffold/SKILL.md`) — the frontmatter form is the literal instruction in skill.rules.md's Placeholder Catalog and Compilation Guidance.
4. Performed the full base-pattern merge (4 rows, with 1 dedupe) into Failure Handling per skill.rules.md's literal instruction, consistent with the `smaqit.new-rules` sibling compiled the same day.
