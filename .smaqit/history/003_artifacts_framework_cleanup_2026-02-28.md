# ARTIFACTS Framework Cleanup

**Date:** 2026-02-28
**Session Focus:** Audit and rewrite of `framework/ARTIFACTS.md` — removal of smaQit product-domain content, rewrite as clean L0 using the four-type content model; relocation of misplaced Isolation Principle to AGENTS.md
**Tasks Completed:** Step 2 of Task 005 (ARTIFACTS.md redesign)
**Tasks Referenced:** Task 004 (distill AGENTS-old), Task 005 (redesign framework files), Task 006 (new-principle skill)

---

## Actions Taken

- Loaded session context; reviewed open tasks (004, 005, 006) and prior history
- Assessed `framework/ARTIFACTS.md` — identified ~75–80% product-domain content tied to smaQit's five-layer system; identified 6 ADK-level principle candidates
- Performed flag pass inline on ARTIFACTS.md — 8 flags placed with rationale and "Where it belongs" dispositions
- Confirmed all 8 dispositions: 7 deleted (product-domain content with no ADK home), 1 relocated (Isolation Principle → AGENTS.md)
- Rewrote ARTIFACTS.md as clean L0: 6 principle + invariant blocks, ~98 lines (down from ~300)
- Moved Isolation Principle to `framework/AGENTS.md` as a new foundational behavior: **Reference-Only Access to Sensitive Input**
- Synced both files to `installer/framework/`

---

## Problems Solved

**ARTIFACTS.md was ~75% smaQit product content** — BUS/FUN/STK/INF/COV layer prefix scheme, checkbox lifecycle, frontmatter state machine, Implements/Enables/Foundation Reference taxonomy, five-layer directory tree, Develop/Deploy/Validate phase report conventions, and `smaQit status` CLI command were all smaQit product decisions, not ADK principles. Resolution: removed all product-domain content; retained ADK-level ideas as principle + invariant blocks.

**Isolation Principle was in the wrong file** — "Agents operate on references, never values" is an agent security behavior principle, not an artifact property. It had been placed in the Implementation Artifacts section of ARTIFACTS.md. Resolution: relocated to AGENTS.md as a named foundational behavior shared by all agents.

---

## Decisions Made

- **7 of 8 flagged items deleted, not relocated** — All smaQit five-layer content (layer prefixes, checkbox syntax, state machine, cross-layer taxonomy, directory tree, phase reports) has no ADK home. The ADK does not own or ship those conventions; they belong in the smaQit product repo. Nothing from these sections was preserved in any ADK file.
- **Isolation Principle renamed on relocation** — "The Isolation Principle" is a presentation label from the original file. Renamed to "Reference-Only Access to Sensitive Input" to match the naming style of other foundational behaviors in AGENTS.md (descriptive, not themed).
- **Flag-before-remove pattern confirmed effective** — Used the same inline flag approach as TEMPLATES.md cleanup (Session 002). Flags capture why content is removed before it disappears, enabling review and disagreement before the rewrite.
- **ARTIFACTS.md now follows the TEMPLATES.md pattern exactly** — `## Core Principles` section, named subsections, bold principle statement, rationale paragraph, `**Invariants:**` bullet block. Reference pattern established in Session 002 applied consistently.

---

## Files Modified

| File | Change |
|------|--------|
| `framework/ARTIFACTS.md` | Complete rewrite: 6 principle + invariant blocks; all product-domain content removed; ~300 lines → ~98 lines |
| `framework/AGENTS.md` | Added `### Reference-Only Access to Sensitive Input` foundational behavior (relocated Isolation Principle) |
| `installer/framework/ARTIFACTS.md` | Synced from root |
| `installer/framework/AGENTS.md` | Synced from root |

---

## Next Steps

- **Task 004** — Distill 9 ADK-level principles from `framework/AGENTS-old.md` into `framework/AGENTS.md`. Isolation Principle is now already there; avoid duplication. Delete AGENTS-old.md after. Assessment from Session 002 still stands: Role Architecture Pattern goes to templates, not AGENTS.md.
- **Task 005 (remaining)** — Redesign `framework/SMAQIT.md` — replace product overview with ADK compilation philosophy. Depends on Task 004.
- **Task 006** — Create `smaqit.new-principle` skill after the framework files are clean.

---

## Session Metrics

- **Tasks completed:** Step 2 of Task 005 (ARTIFACTS.md)
- **Files modified:** 4
- **Lines removed:** ~200 (product-domain content)
- **Principle blocks written:** 6
- **Flags placed:** 8 (7 deleted, 1 relocated)
- **Principles relocated to AGENTS.md:** 1 (Isolation Principle)
