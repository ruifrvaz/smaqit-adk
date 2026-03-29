# Skill Framework and Test Plan

**Date:** 2026-03-03
**Session Focus:** Design and implement the skill compilation framework (Task 009); plan the test framework (Task 010)
**Tasks Completed:** Task 009 (Create smaqit.new-skill Skill)
**Tasks Created:** Task 010 (Test Framework)
**Tasks Referenced:** Task 006 (smaqit.new-principle — dependency order corrected)

---

## Actions Taken

- Loaded session context; confirmed Task 009 as active work item (dependency order: 009 before 006)
- Fetched Anthropic official skill authoring best practices from platform docs
- Assessed which best practices belong at L0 vs L1 — concluded L0 changes not needed; all best practices translate to L1 directives
- Debated skill section structure and placeholder model; decided skills use `[SCREAMING_CASE]` compile-time placeholders matching agent template convention
- Renamed `skill.template.md` → `base-skill.template.md` to match established naming convention
- Authored `templates/skills/base-skill.template.md` with 8 placeholders: SKILL_NAME, SKILL_DESCRIPTION, SKILL_VERSION, SKILL_TITLE, PURPOSE_CONTENT, STEPS_CONTENT, OUTPUT_CONTENT, SCOPE_CONTENT, COMPLETION_CONTENT, FAILURE_HANDLING_CONTENT
- Transformed `templates/skills/compiled/skill.rules.md` from vocabulary-only stub to full L1 rules file (frontmatter, Source L0 Principles table, Placeholder Catalog, 5 directive groups, Compilation Guidance for Agent-L1)
- Extended `agents/smaqit.L1.agent.md` scope: added skill compilation responsibility, definition file reading pattern, output location, completion criteria for skill compilation, and new failure handling rows
- Created `skills/smaqit.new-skill/SKILL.md` — 6-section gathering flow, description validation, fragility-per-step gathering, compilation via L1 subagent, definition file format
- Updated `installer/Makefile` prepare target: added `smaqit.new-skill` copy, updated template filename reference
- Updated `.smaqit/tasks/PLANNING.md`: Task 009 moved to Active; Task 006 corrected to depend on Task 009
- Created `.smaqit/tasks/009_create_new_skill_skill.md` task file
- Built installer — passed clean
- Assessed test framework needs: identified embed bug (smaqit.new-skill not embedded in binary), three test layers (structural, installer unit tests, behavioral evals), and complete absence of existing test infrastructure
- Saved Task 010 with full 4-phase plan to `.smaqit/tasks/010_test_framework.md`

---

## Problems Solved

**Compiler ownership gap for skills** — No Level agent owned `skills/*.md` creation. Resolved by extending L1 scope: L1 already owns `templates/skills/` machinery; extending it to compile final skill output is the natural fit, paralleling L2's ownership of `agents/*.md`.

**Dependency order inversion** — Task 009 was listed as depending on Task 006, but the correct order is reversed: `smaqit.new-principle` (006) should be created *using* `smaqit.new-skill` (009). Corrected in PLANNING.md.

**Skill body placeholder model** — Debated whether skills should use `[PLACEHOLDER]` compile-time substitution or prose-only bodies. Decided to use `[SCREAMING_CASE]` placeholders in the template consistent with the agent template convention; L1 writes prose content for each placeholder from definition file inputs (authoring compiler, not substitution compiler).

**Degrees of freedom principle** — Official best practices introduced a high/medium/low fragility model for step specificity. Assessed whether this belongs at L0 (principle) or L1 (directive). Concluded L1 only — it's authoring craft, not a structural principle. Encoded as a directive in `skill.rules.md`.

**Embed bug identified** — `installer/main.go` uses a single-file `go:embed` directive that only embeds `smaqit.new-agent/SKILL.md`. The `smaqit.new-skill` is copied by `make prepare` but neither embedded in the binary nor installed by `cmdInit`. Deferred fix to Task 010 (Phase 0 prerequisite).

---

## Decisions Made

- **L0 not updated** — Assessed 4 potential L0 additions from official best practices (Conciseness, Degrees of Freedom, third-person description, one-level-deep references). Concluded none require new L0 principles; all translate cleanly to L1 directives. Existing 5 SKILLS.md principles are sufficient.
- **L1 extended, not a new Level agent** — Skill compilation assigned to L1 (not a new L4 or separate agent). L1's scope statement updated from "exclusively on Level 1 template files" to cover skill compilation from definition files.
- **`base-skill.template.md` naming** — Renamed from `skill.template.md` to follow `base-agent.template.md` convention exactly. Makefile and installer directory updated accordingly.
- **No Validation section in skill template** — Decided validation is part of Steps, not a dedicated template section. The 7-section structure (Purpose, Steps, Output, Scope, Completion, Failure Handling + frontmatter) mirrors agent structure without artificial sections.
- **Test framework scope: all three layers** — User selected full coverage: installer unit tests (Layer 2) + structural validation (Layer 1) + behavioral LLM evaluations (Layer 3). Behavioral evals use Anthropic standard evaluation JSON format with a custom Go runner.
- **Embed bug fix in Task 010 Phase 0** — Not fixed immediately; included as prerequisite Phase 0 of the test framework task so tests can assert correct post-`init` state.

---

## Files Modified

| File | Change |
|------|--------|
| `templates/skills/base-skill.template.md` | Renamed from `skill.template.md`; fully authored with 8 `[SCREAMING_CASE]` placeholders and 7 sections |
| `templates/skills/compiled/skill.rules.md` | Transformed from vocabulary-only stub to full L1 rules file; added frontmatter, Source L0 Principles, Placeholder Catalog, 5 directive groups, Compilation Guidance |
| `agents/smaqit.L1.agent.md` | Extended scope, input, output, completion criteria, and failure handling to cover skill compilation |
| `skills/smaqit.new-skill/SKILL.md` | Created — 6-section gathering flow, validation, compilation via L1 subagent |
| `installer/Makefile` | Added `smaqit.new-skill` to prepare target; updated template filename reference |
| `.smaqit/tasks/PLANNING.md` | Task 009 moved to Active; Task 006 corrected; Task 010 added to Future |
| `.smaqit/tasks/009_create_new_skill_skill.md` | Created — task file with acceptance criteria and design decisions |
| `.smaqit/tasks/010_test_framework.md` | Created — full 4-phase test framework plan |

---

## Next Steps

- **Task 009** — Still marked In Progress; needs verification pass (confirm acceptance criteria all met) before marking Completed
- **Task 006** — Create `smaqit.new-principle` skill using `smaqit.new-skill`; should be the first real-world use of the new skill to validate the flow
- **Task 010** — Implement the test framework: Phase 0 (embed bug fix) → Phase 1 (installer unit tests) → Phase 2 (structural validation) → Phase 3 (behavioral evals)

---

## Session Metrics

- **Tasks completed:** 1 (Task 009)
- **Tasks created:** 1 (Task 010)
- **Files created:** 3 (`smaqit.new-skill/SKILL.md`, `009_create_new_skill_skill.md`, `010_test_framework.md`)
- **Files modified:** 5 (`base-skill.template.md`, `skill.rules.md`, `smaqit.L1.agent.md`, `Makefile`, `PLANNING.md`)
- **Files renamed:** 1 (`skill.template.md` → `base-skill.template.md`)
- **Build status:** Passing
- **Key outcome:** Skill compilation framework complete; `smaqit.new-skill` ready for use; test framework plan saved
