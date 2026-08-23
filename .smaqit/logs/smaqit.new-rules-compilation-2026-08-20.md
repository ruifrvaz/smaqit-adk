# Compilation Log: smaqit.new-rules

**Compilation timestamp:** 2026-08-20
**Compiled by:** Agent-L2 (smaqit.L2)
**Pattern:** Pattern 4 — Skill Compilation (3-way merge)

## L1 Sources Read

- `agents/smaqit.L2.md` — role, directives, "For Skills" compilation procedure (steps 1-8), "Section-Level Compilation" guidance
- `templates/skills/base-skill.template.md` — structural template with `[PLACEHOLDER]` markers (frontmatter + `## Steps`/`## Output`/`## Scope`/`## Examples`/`## Gotchas`/`## Completion`/`## Failure Handling`)
- `templates/skills/compiled/skill.rules.md` — placeholder catalog, Degrees of Freedom directive, Conciseness directive, Reference Structure directive, Base Failure Handling Pattern table, "Compilation Guidance for Agent-L2" (steps 1-9)

All three sources were read from the local worktree copies (`templates/...` relative to the worktree root), per the case instruction that this repository's own `templates/` tree is the ADK source of truth being shipped — not the globally-installed `~/.agents/smaqit-adk/` copies.

## Definition File

- `.smaqit/definitions/skills/smaqit.new-rules.md` (tier: advanced) — Identity, Steps (with fragility annotations), Output, Scope, Completion, Failure Scenarios, Gotchas, Examples, Compatibility. All required sections present (identity, steps with fragility levels, output, scope, completion, failure handling); no gaps requiring a stop-and-request.

## Merge Process Summary

- `[SKILL_NAME]` → `smaqit.new-rules` (from Identity)
- `[SKILL_DESCRIPTION]` → definition's draft description rewritten into the compiled form: "Adds or refines a Level 1 compiled rules file in templates/agents/compiled/ or templates/skills/compiled/ — delegates gathering, validation, and authoring to smaqit.L1, then cascades the change by recompiling every registered dependent agent or skill via smaqit.L2." Verified capability-oriented, present-tense, declarative, states what the skill produces (a changed rules file + recompiled dependents), no conversational/user-centric gating, no first person, no can/may/might/will. Backticks dropped from the frontmatter description to match the plain-scalar style already established by `skills/smaqit.new-principle/SKILL.md`'s compiled description.
- `[SKILL_VERSION]` → `"0.1.0"` (from Identity)
- `[SKILL_TITLE]` → "New Rules"
- `[STEPS_CONTENT]` → definition's 4 steps, degrees-of-freedom applied per fragility marking: Step 1 (low) as prose with platform sub-bullets, trimmed of the definition's parenthetical fragility label itself (the label is compile-time guidance for Agent-L2, not runtime instruction for the executing agent); Step 2 (high fragility) rendered with the literal `smaqit-adk registry get` commands preserved verbatim, no paraphrasing of the `--scope base`/`--scope extension` distinction or the "missing registry file = zero dependents" rule; Step 3 (medium) kept as directive prose describing the pattern (single-hop `Task` call, mirrors `smaqit.development`'s orchestration, never nested inside `smaqit.L1`) rather than literal pseudocode, since the invocation shape varies by dependent type; Step 4 (low) as a single prose instruction
- `[OUTPUT_CONTENT]` → definition Output section, lightly reworded to state the subagent responsible for each artifact (`smaqit.L1` for the rules file, `smaqit.L2` for dependent recompiles) per skill.rules.md's `[OUTPUT_CONTENT]` definition ("artifact produced, its path, and the subagent invoked, if any")
- `[SCOPE_CONTENT]` → definition Scope section, restructured into a lead sentence + "Out of scope" bullets pointing to `smaqit.new-template` and `smaqit.new-principle` (the only two redirects the definition itself names — did not add `smaqit.new-agent`/`smaqit.new-skill` redirects the way `smaqit.new-principle`'s existing compiled Scope does, since this skill's definition does not mention agent/skill creation as an adjacent concern)
- `[EXAMPLES_CONTENT]` → definition's single example, verbatim
- `[GOTCHAS_CONTENT]` → definition's 2 gotchas, verbatim (already concise, no trimming needed)
- `[COMPLETION_CONTENT]` → definition Completion checklist, verbatim
- `[FAILURE_HANDLING_CONTENT]` → see Failure Handling Merge below
- `compatibility` frontmatter → included (definition specifies a Compatibility note: "Requires the `smaqit-adk` binary on `PATH` for the registry subcommands used in Step 2."), paraphrased into one sentence with backticks dropped to match plain-scalar YAML style, flat top-level field placed after `metadata`
- `allowed-tools` frontmatter → omitted; definition does not specify a value

## Failure Handling Merge

Base Failure Handling Pattern (4 rows, from skill.rules.md) merged against the definition's 6 Failure Scenarios rows:

- Base "Subagent invocation fails" is textually identical to the definition's own "Subagent invocation fails" row (both: "Report the failure with context; do not silently retry") — deduped to a single row.
- Base "Required input not provided" and "Gathered input is ambiguous" have no direct definition counterpart but remain applicable at the orchestration level (this skill must have a clear instruction to hand `smaqit.L1`, and a registry response could be malformed/ambiguous) — kept as foundation rows per skill.rules.md's "Insert the following as the foundation of `[FAILURE_HANDLING_CONTENT]`" instruction.
- Base "Output artifact already exists" has no exact definition counterpart but is genuinely applicable: the Step 3 cascade recompiles dependent agent/skill files via `smaqit.L2`, which are pre-existing files being overwritten — kept as a foundation row rather than dropped as inapplicable.
- Remaining 5 definition rows (`smaqit.L1` not installed, registry file missing/unreadable, a dependent's recompile fails, user wants templates/principles instead, user abandons midway) had no base overlap — carried through unchanged.

Result: 9-row single coherent table (4 base − 1 dedupe + 6 definition = 9), header `Situation | Action` matching the base table's header.

**Note on precedent:** `skills/smaqit.new-principle/SKILL.md` (an existing compiled sibling) omits all 4 base rows entirely rather than merging them. `skills/smaqit.bench-scaffold/SKILL.md`'s compilation log instead performs the full merge-with-dedupe documented above. This compilation followed skill.rules.md's literal instruction (merge, don't omit) and the bench-scaffold precedent rather than the new-principle precedent; flagging the divergence for awareness, not resolving it here since `smaqit.new-principle` is out of this task's scope.

## Conciseness Filter

Removed the definition's inline parenthetical fragility labels ("(low fragility)", "(high fragility — ...)", "(medium fragility)") from the compiled Steps — these are compile-time signals to Agent-L2 for choosing instruction form, not information the executing agent needs at runtime. No other trims were needed; the definition's prose was already concise and each remaining sentence (exact registry commands, the `smaqit.development` orchestration analogy, the no-nested-Task rule) carries information the executing agent could not safely infer.

## Progressive Disclosure

Not needed — compiled body is 80 lines, well under the 400-line threshold. No `references/` subdirectory created.

## Validation Checklist

- [x] No unresolved `[PLACEHOLDER]`-style uppercase-bracket tokens (`grep -nE '\[[A-Z_]+\]' skills/smaqit.new-rules/SKILL.md` → no matches)
- [x] Frontmatter `name` field exactly `smaqit.new-rules`
- [x] Description is capability-oriented, declarative, present-tense; no "Use when...", no first person, no can/may/might/will
- [x] All required sections present, in template order: `## Steps`, `## Output`, `## Scope`, `## Examples`, `## Gotchas`, `## Completion`, `## Failure Handling`
- [x] `compatibility` frontmatter field present (sourced from definition); `allowed-tools` omitted (not specified)
- [x] No nested reference chains — skill body references `~/.claude/agents/smaqit-L1.md` (an external agent file, read inline for one turn on Copilot only) but does not reference any file bundled within its own skill directory, so no `SKILL.md` → file → file chain exists
- [x] Body under 400 lines (80 lines)

## Decisions Made

1. Stripped the definition's parenthetical fragility annotations from the compiled Steps text (compile-time metadata, not runtime instruction) while still honoring the fragility level each annotation indicated in how each step was written.
2. Performed the full base-pattern merge (4 rows) into Failure Handling per skill.rules.md's literal instruction, rather than mirroring `smaqit.new-principle`'s existing precedent of omitting the base rows — documented the divergence above rather than silently picking one without a trace.
3. Limited the Scope section's "Out of scope" redirects to the two skills the definition itself names (`smaqit.new-template`, `smaqit.new-principle`), not extending to `smaqit.new-agent`/`smaqit.new-skill` as `smaqit.new-principle`'s compiled Scope does, since those skill names are not part of `smaqit.new-rules`'s own definition and are not verified to exist under those names in this repository (`smaqit.create-agent`/`smaqit.create-skill` are the actual installed names).
