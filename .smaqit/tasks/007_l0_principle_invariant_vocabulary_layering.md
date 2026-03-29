# L0 Principle + Invariant + Vocabulary Layering

**Status:** Completed
**Created:** 2026-02-27
**Completed:** 2026-02-27

## Description

Established a precise three-layer content model for ADK framework files and applied it to `framework/TEMPLATES.md`, correcting a category error where vocabulary catalogs and product-domain content were mixed into L0 principle files.

### The Content Model

Prior to this task, the distinction between L0 and L1 was described as "principles vs directives." This left a gap: descriptive information (catalogs, tables of named things, reference lists) had no defined home, causing it to accumulate in framework files by default.

The refined model introduces three distinct content types:

| Type | Answers | Language | Lives at |
|------|---------|----------|----------|
| **Principle** | Why does this matter? | Rationale prose | L0 framework |
| **Invariant** | What is always true when this principle is applied? | Declarative present-tense describing compliant behavior | L0 framework |
| **Vocabulary / Catalog** | What named things exist and what do they mean? | Definitions, tables of terms, placeholder catalogs | L1 templates / rules |
| **Directive** | What must an agent do? | MUST / MUST NOT / SHOULD | L1 templates / rules |

**Key distinction between invariant and directive:**
- An invariant states what is *true* about a compliant agent: "An agent that receives out-of-scope work stops, identifies the correct agent, and redirects."
- A directive instructs an agent what to *do*: "Agents MUST NOT plan, create todos, or execute when a request falls outside their designated scope."

L1 reads invariants and compiles them into directive form. This gives L1 a concrete compilation surface without importing imperative language into L0.

**Key distinction between vocabulary and principle:**
- A vocabulary catalog (e.g., a placeholder table) describes which named things exist in a specific template. It requires knowing *what* agents, layers, phases, or placeholders exist — i.e., it is specific to an implementation.
- A principle describes *why* those things should exist and what rules govern them.

Vocabulary catalogs belong at L1 because they are a consequence of template design decisions, not prior architectural principles.

## What Was Done

### `framework/TEMPLATES.md` — rewrote as clean L0

Replaced the old mixed content with five principle + invariant blocks:

| Principle | Summary |
|-----------|---------|
| Template as Compilation Surface | Templates define invariant structure; placeholders mark variation; compilation fills without altering |
| Placeholder Convention | Consistent `[SCREAMING_CASE]` format; meaning documented in rules files, not templates |
| Template Hierarchy | Foundation + extensions strictly additive; extensions never duplicate or override foundation |
| Section Structure as Behavioral Contract | Sections define behavioral map; L2 does not add or omit sections |
| Extension Inheritance | All foundation content is included before extension content; extensions cannot weaken foundation |

Removed content (product-domain, confirmed no ADK home):
- Specification template format (smaQit five-layer spec system, frontmatter schema, state lifecycle)
- Prompt template format (prompt-as-input-record model, smaQit `[layer]` naming convention)

Relocated content (L1 vocabulary):
- All placeholder catalog tables (`[LAYER]`, `[PHASE]`, `[PHASE_NAME]`, etc.) moved to the three rules files

The `## Agent Templates` section at the end remains as a thin ADK-level reference to the actual template files and their section structure.

### `templates/agents/compiled/*.rules.md` — added placeholder catalogs

Each rules file received a `## Placeholder Catalog` section positioned between `## Source L0 Principles` and `## L1 Directive Compilation`:

| File | Catalog covers |
|------|---------------|
| `base.rules.md` | 15 base template placeholders (`[AGENT_NAME]`, `[ROLE_CONTENT]`, `[BASE_MUST_DIRECTIVES]`, etc.) |
| `specification.rules.md` | 16 specification-extension placeholders (`[LAYER]`, `[LAYER_NAME]`, `[SPECIFICATION_MUST_DIRECTIVES]`, etc.) |
| `implementation.rules.md` | 16 implementation-extension placeholders (`[PHASE]`, `[PHASE_NAME]`, `[STATE_TRACKING_CONTENT]`, etc.) |

Notable documentation decisions:
- `[STATE_TRACKING_CONTENT]` in `implementation.rules.md` notes its secondary placeholders `[STATUS_VALUE]` and `[TIMESTAMP_FIELD]` as product-defined values resolved at L2
- `[AGENT_NAME]` in `implementation.rules.md` is flagged as overriding the base `[AGENT_TITLE]` — a naming inconsistency with a paper trail

## Reference Pattern for Future Cleanups

When auditing any framework file or L1 rules file, apply this test to each content block:

1. **Does it say *why* something matters?** → Principle (L0)
2. **Does it describe what is *always true* about a compliant system?** → Invariant (L0), stated declaratively
3. **Does it list *which things exist* by name?** → Vocabulary/Catalog (L1, in the rules file for the relevant template)
4. **Does it instruct an agent what to *do*?** → Directive (L1, MUST/MUST NOT/SHOULD)
5. **Does it assume a specific product domain (layer names, phase names, file paths, CLI commands)?** → Product-domain content; does not belong in smaqit-adk at any level

For any content found at the wrong level:
- Extract it, move it, and leave no orphan
- If the content has no valid ADK home, flag it inline with "Why excluded" and "Where it belongs" before removing — the flag serves as an audit trail during transition
- Once the flag is acted on (content moved or confirmed deleted), remove the flag section itself

## Acceptance Criteria

- [x] Three-layer content model (principle / invariant / vocabulary+directive) documented in this task
- [x] `framework/TEMPLATES.md` contains only L0 principles and invariants
- [x] All placeholder catalogs moved to the three `compiled/*.rules.md` files
- [x] Product-domain content removed from `framework/TEMPLATES.md` (no replacement needed in this repo)
- [x] Flagged transition section removed from TEMPLATES.md once disposition was decided

## Notes

- Task 004 (distill AGENTS-old) and Task 005 (redesign remaining framework files) should apply this same layering model. AGENTS.md is already close; SMAQIT.md and ARTIFACTS.md need the most work.
- The "Flagged: Out-of-Scope Content" pattern used transitionally in TEMPLATES.md is a valid technique for incremental cleanup: flag before removing, then remove the flag once disposition is confirmed. Do not leave flags permanently.
