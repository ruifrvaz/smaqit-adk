# Agents Old Distillation

**Date:** 2026-03-02
**Session Focus:** Task 004 — Audit `framework/AGENTS-old.md`, distill ADK-level principle content into `framework/AGENTS.md`, delete the old file
**Tasks Completed:** Task 004 (Distill AGENTS-old into AGENTS.md)
**Tasks Referenced:** Task 005 (SMAQIT.md redesign), Task 006 (new-principle skill)

---

## Actions Taken

- Loaded session context; reviewed open tasks (004, 005, 006) and prior history (Session 003)
- Read both `framework/AGENTS-old.md` and `framework/AGENTS.md` in full
- Conducted a line-by-line audit of all 9 principle candidates in AGENTS-old.md
- Identified 3 gaps between AGENTS-old.md content and current AGENTS.md
- Added 3 targeted additions to `framework/AGENTS.md`
- Confirmed AGENTS-old.md was already deleted (prior cleanup); synced installer
- Updated PLANNING.md to mark Task 004 as completed

---

## Problems Solved

**AGENTS-old.md contained ~90% smaQit product content** — five-layer agent mappings (business/functional/stack/infrastructure/coverage), phase orchestration model, `smaqit plan --phase=` CLI, YAML frontmatter lifecycle, cross-layer consolidation, pre-orchestration validation, implementation completion checklists. None of this belongs in ADK framework principles. All discarded.

**3 ADK-level behavioral gaps identified** — The current AGENTS.md was missing two invariants and one foundational behavior that had valid ADK-level expression in AGENTS-old.md.

---

## Decisions Made

- **Flag-assumptions invariant belongs in Fail-Fast on Ambiguity** — "When clarification is unavailable, agents flag assumptions explicitly rather than embedding them silently." This is a universal agent behavior, not smaQit-specific. Added as a third sentence to the existing behavior.

- **Blocker-stop invariant belongs in Self-Validation Before Completion** — "When completion criteria cannot be met, agents flag the blocker and stop rather than lowering quality standards or inventing solutions." Prevents quality drift under failure conditions. Added as a third sentence to the existing behavior.

- **Skill-Mediated Workflows is a new foundational behavior** — "Agents invoke skills for specialized, cross-cutting workflows rather than implementing those workflows inline. Skills provide the structure; agents provide the trigger." This is the behavioral complement to the Invocation Model (which describes the system-level flow). The foundational behavior specifies what individual agents must do. Added as a new `### Skill-Mediated Workflows` section.

- **All other AGENTS-old.md content discarded** — smaQit five-layer conventions, phase orchestration, CLI command coupling, and YAML frontmatter state machines have no ADK home. None were relocated.

---

## Files Modified

| File | Change |
|------|--------|
| `framework/AGENTS.md` | Added 3 items: assumption-flagging invariant, blocker-stop invariant, Skill-Mediated Workflows behavior |
| `installer/framework/AGENTS.md` | Synced from root |
| `.smaqit/tasks/PLANNING.md` | Task 004 moved to Completed |

---

## Next Steps

- **Task 005** — Rewrite `framework/SMAQIT.md`: replace smaQit product overview with ADK compilation philosophy. This is the last remaining framework file with product-domain content.
- **Task 006** — Create `smaqit.new-principle` skill after framework files are clean.

---

## Session Metrics

- **Tasks completed:** 1 (Task 004)
- **Files modified:** 3
- **Principle candidates audited:** 9
- **Items added to AGENTS.md:** 3 (2 invariants, 1 new behavior)
- **Items discarded:** ~6 product-domain sections (all smaQit-specific conventions)
