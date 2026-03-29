# L0 Principle Invariant Layering

**Date:** 2026-02-27
**Session Focus:** Establishing and applying the principle + invariant + vocabulary layering model to framework files; TEMPLATES.md redesign
**Tasks Completed:** Task 007 (L0 Principle + Invariant + Vocabulary Layering) — retroactive
**Tasks Referenced:** Task 004 (distill AGENTS-old), Task 005 (redesign framework files)

---

## Actions Taken

- Loaded session context; reviewed open tasks (004, 005, 006) and previous session history
- Initiated assessment of Task 004 (distill AGENTS-old into AGENTS.md): flagged Role Architecture Pattern as L1 content (wrong destination), noted boundary enforcement step needed scrubbing of product-domain language, identified structural decision between new Validation section vs enriching existing foundational behaviors
- Reassessed all five framework files: identified AGENTS.md and SKILLS.md as near-clean; SMAQIT.md, ARTIFACTS.md, and TEMPLATES.md as primarily product-domain content
- Established the **principle + invariant** distinction in response to user question about framework files lacking directional content
  - Principles answer "why" (rationale prose, L0)
  - Invariants answer "what is always true" (declarative present-tense, L0)
  - This gives L1 a concrete compilation surface without importing imperative MUST/MUST NOT language into L0
- Demonstrated invariant authoring on TEMPLATES.md content (Template Hierarchy, Placeholder Convention)
- Identified a fourth type: **Vocabulary / Catalog** — lists of named things that belong at L1, not L0
- Applied this four-type model to `framework/TEMPLATES.md`:
  - Rewrote as five principle + invariant blocks (Template as Compilation Surface, Placeholder Convention, Template Hierarchy, Section Structure as Behavioral Contract, Extension Inheritance)
  - Removed specification template format (product-domain) and prompt template format (product-domain)
  - Moved all placeholder catalog tables to `compiled/*.rules.md`
  - Used flagged transition section to document removed content with rationale; removed the flag once dispositions were confirmed
- Added `## Placeholder Catalog` section to all three compiled rules files, documenting placeholder meanings co-located with L2 compilation guidance
- Confirmed TEMPLATES.md placeholder principle is adequately documented at L0; noted one gap (L1 authors define placeholders, L2 resolves them — authorship not stated)
- Decided flagged product-domain content has no ADK home — not wiki, not copilot instructions; the flag served its audit purpose and was removed
- Created Task 007 retroactively to document the content model as a reference pattern for future cleanups
- Added Framework Content Model section to `.github/copilot-instructions.md` with the four-type table and key distinctions as MUST NOT directives
- Expanded Level Boundaries section in `docs/wiki/extending-smaqit.md` with the full model table, invariant/directive distinction, vocabulary/principle distinction, and the 5-question audit test

---

## Problems Solved

**Framework files had no home for descriptive/directional content between pure philosophy and MUST directives** — The existing L0/L1 binary (principles vs directives) left a gap that caused vocabulary catalogs and invariant-style content to accumulate incorrectly at L0. Resolution: introduced the invariant and vocabulary types to complete the model.

**Placeholder catalog was misplaced at L0 in TEMPLATES.md** — Tables listing `[LAYER]`, `[PHASE]`, etc. are vocabulary definitions (which things exist), not principles (why they exist). Resolution: moved all catalogs to the three `compiled/*.rules.md` files, where L2 can find them co-located with compilation guidance.

**Product-domain content had no flagging mechanism during cleanup** — Removing content without a trace risked silent information loss. Resolution: used inline "Flagged: Out-of-Scope Content" section as a transition artifact with "Why excluded" and "Where it belongs" entries; removed the section once each item was disposed of.

---

## Decisions Made

- **Four content types, not two** — Principle, Invariant, Vocabulary/Catalog, Directive. The invariant/directive distinction gives L1 a compilation surface without importing imperative language into L0. The vocabulary/principle distinction prevents named-thing catalogs from being treated as framework philosophy.
- **Invariants are declarative, not imperative** — "An agent that receives out-of-scope work stops, identifies the correct agent, and redirects" is an invariant. "Agents MUST NOT plan when out of scope" is a directive. L1 reads the former and produces the latter.
- **Vocabulary catalogs belong at L1** — Placeholder tables, agent name mappings, layer/phase tables are consequences of template design decisions; they belong in the rules file for the template that uses them, not in the framework.
- **Product-domain content has no ADK home** — Spec template format, prompt template format, and frontmatter schema are smaQit product decisions. They are not relocated within smaqit-adk; they belong in the consuming product's documentation.
- **Flag-before-remove pattern for cleanup** — During framework file cleanup, flag misplaced content inline with rationale before deleting; remove the flag section once disposition is confirmed. Prevents silent information loss during incremental cleanup.
- **Framework Content Model documented in copilot instructions** — Copilot reads this every session and applies it when reasoning about framework files. This is the highest-leverage location for the authoring convention.

---

## Files Modified

| File | Change |
|------|--------|
| `framework/TEMPLATES.md` | Complete rewrite: five principle + invariant blocks; removed spec/prompt template content; removed placeholder catalog; cleaned to pure L0 |
| `templates/agents/compiled/base.rules.md` | Added `## Placeholder Catalog` section with 15 base template placeholder definitions |
| `templates/agents/compiled/specification.rules.md` | Added `## Placeholder Catalog` section with 16 specification-extension placeholder definitions |
| `templates/agents/compiled/implementation.rules.md` | Added `## Placeholder Catalog` section with 16 implementation-extension placeholder definitions |
| `.github/copilot-instructions.md` | Added `## Framework Content Model` section with four-type table and MUST NOT directives |
| `docs/wiki/extending-smaqit.md` | Expanded `### Level Boundaries` with the full model, key distinctions, and 5-question audit test |
| `.smaqit/tasks/007_l0_principle_invariant_vocabulary_layering.md` | Created (retroactive task documenting the content model and changes) |
| `.smaqit/tasks/PLANNING.md` | Added Task 007 to Completed table |

---

## Next Steps

- **Apply the same approach to `framework/ARTIFACTS.md`** — This is the most contaminated framework file. Audit each content block against the 5-question test. Much of it (spec IDs with BUS/FUN/STK prefixes, acceptance criteria format, spec lifecycle states, coverage translation, traceability matrices, frontmatter schema) is product-domain content with no ADK home. The ADK-level artifact principle should be thin: artifacts are compilation outputs that reference their source and carry a type. Flag before removing; remove flags once dispositions are confirmed.
- **Task 004:** Distill 9 ADK-level principles from `framework/AGENTS-old.md` into `framework/AGENTS.md` using the three-layer model (principle + invariant at L0; vocabulary and directives already in rules files). Delete AGENTS-old.md after. Assessment from this session still stands — Role Architecture Pattern goes to templates, not AGENTS.md.
- **Task 005:** Redesign SMAQIT.md (replace product overview with ADK compilation philosophy) after Task 004 is done.

---

## Session Metrics

- **Tasks created:** 1 (007, retroactive)
- **Files modified:** 8
- **Content types established:** 4 (principle, invariant, vocabulary/catalog, directive)
- **Principles written for TEMPLATES.md:** 5
- **Placeholder entries documented in rules files:** 47 (15 + 16 + 16)
