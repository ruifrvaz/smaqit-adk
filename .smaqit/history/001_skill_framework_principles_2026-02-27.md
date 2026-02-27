# Skill Framework Principles

**Date:** 2026-02-27
**Session Focus:** Skill description quality, framework principle formatting, task planning for framework redesign
**Tasks Completed:** Skill content refinement (continuation of Task 003)
**Tasks Created:** Task 004 (Distill AGENTS-old), Task 005 (Redesign Framework Files)

---

## Actions Taken

- Refined `skills/smaqit.new-agent/SKILL.md` per progressive disclosure principles: removed "When to use this skill" section and opening meta-description line; trimmed Notes to execution-relevant content only
- Improved skill `description` field across three iterations, converging on full explanatory form
- Added `## Skill Description Quality` principle to `framework/SKILLS.md`, then restructured into `### Description-Driven Activation` following SMAQIT.md principle pattern
- Converted all of `framework/SKILLS.md` Key Principles bullets to `### Name` / bold one-liner / explanation paragraph pattern
- Updated `.github/copilot-instructions.md` twice: added MUST NOT for git write operations; added Build Workflow section documenting `make build` as the correct sync path
- Assessed `framework/AGENTS-old.md` for distillation candidates; identified 9 ADK-level principles missing or weakly stated in current AGENTS.md
- Created Task 004 and Task 005; renumbered previously created task from 004 to 005

---

## Problems Solved

**Skill description was too short and contained false-positive trigger words** — "orchestrator", "Q&A helper", "utility" in the description caused incorrect skill activation on unrelated queries. Resolution: three-iteration refinement converging on a multi-sentence explanatory description that answers "this skill does..." and "use this skill when...".

**Framework principles used bullet lists instead of SMAQIT.md pattern** — SKILLS.md used `**Key Principles:** - bullet` format. Resolution: converted to `### Principle Name` / bold one-liner / explanation paragraphs throughout.

**Copilot was manually copying files to `installer/` subdirectories** — This conflated build intermediates with source artifacts. Resolution: documented in copilot-instructions.md that `make build` handles all sync via `prepare` target; MUST NOT manually copy to installer subdirectories.

**No process for agent skill invocation on ambiguity** — AGENTS-old.md contained a principle that agents should invoke skills for ambiguity/complexity rather than implementing workflows inline. Noted as distillation candidate for Task 004.

---

## Decisions Made

- **Skill descriptions must be explanatory, not labels or taglines** — Length is not the constraint; precision is. A description can be several sentences. The two questions it must answer: "This skill does..." and "Use this skill when...". Keywords embedded in explanation are useful; bare keyword lists are not.
- **Description-Driven Activation is a Core Principle in SKILLS.md** — Not a separate section; integrated as the fifth core principle following the same pattern as the other four.
- **AGENTS-old.md distillation is a prerequisite for framework redesign** — Task 004 must complete before Task 005 so that AGENTS.md is the reference model for other framework files.
- **No git commits in this session** — User established that git commits are not part of the Copilot workflow unless explicitly requested.
- **Installer sync via `make build` only** — `installer/framework/`, `installer/skills/`, `installer/agents/`, `installer/templates/` are `.gitignore`d build intermediates; `make prepare` (run by `make build`) regenerates them.

---

## Files Modified

| File | Change |
|------|--------|
| `skills/smaqit.new-agent/SKILL.md` | Three description iterations; removed redundant sections; final description is multi-sentence explanation |
| `framework/SKILLS.md` | Converted Key Principles to Core Principles with SMAQIT.md pattern; added Description-Driven Activation principle; renamed Progressive Disclosure subsection to Loading Stages |
| `.github/copilot-instructions.md` | Added MUST NOT for git write operations; added Build Workflow section |
| `.smaqit/tasks/004_distill_agents_old_into_agents.md` | Created (new) |
| `.smaqit/tasks/005_redesign_framework_files.md` | Created (renamed from 004) |
| `.smaqit/tasks/PLANNING.md` | Updated with tasks 004 and 005 |

---

## Next Steps

- **Task 004:** Distill 9 ADK-level principles from `framework/AGENTS-old.md` into `framework/AGENTS.md`, then delete the old file
- **Task 005:** Replace `SMAQIT.md`, `ARTIFACTS.md`, and the product-domain sections of `TEMPLATES.md` with ADK-scoped content
- All changes since last commit are unstaged — commit when ready

---

## Session Metrics

- **Tasks created:** 2 (004, 005)
- **Files modified:** 6
- **Skill description iterations:** 3
- **Principle reformats:** 5 (4 existing converted + 1 new)
- **Copilot instruction additions:** 2 directives
