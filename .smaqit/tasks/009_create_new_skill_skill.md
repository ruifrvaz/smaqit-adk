# Task 009: Create smaqit.new-skill Skill

**Status:** Completed
**Created:** 2026-03-02
**Completed:** 2026-03-29

## Objective

Create the `smaqit.new-skill` skill — a meta-skill that guides creation of new skills. This skill is a prerequisite for Task 006 (`smaqit.new-principle`), which should itself be created using `smaqit.new-skill`.

## Acceptance Criteria

- [x] `templates/skills/base-skill.template.md` authored with all placeholders
- [x] `templates/skills/compiled/skill.rules.md` contains full L2 directives (vocabulary + compilation guidance)
- [x] `agents/smaqit.L2.agent.md` scope extended to include skill compilation (see Design Decisions)
- [x] `skills/smaqit.new-skill/SKILL.md` created following the `smaqit.new-agent` pattern
- [x] `installer/Makefile` prepare target includes `smaqit.new-skill`
- [x] `cd installer && make build` passes without errors
- [x] Installer copies `skills/smaqit.new-skill/SKILL.md` into embedded directory

## Files

| File | Change |
|------|--------|
| `templates/skills/base-skill.template.md` | Renamed from `skill.template.md`; fully authored with 8 placeholders |
| `templates/skills/compiled/skill.rules.md` | Added frontmatter, Source L0 Principles, Placeholder Catalog, L2 Directive Compilation, Compilation Guidance; corrected Agent-L1 → Agent-L2 references |
| `agents/smaqit.L2.agent.md` | Extended scope, input, output, compilation patterns (Pattern 4), and procedure for skill compilation; L1 cleaned of skill output responsibility |
| `agents/smaqit.L1.agent.md` | Corrected: removed skill output compilation scope; boundary enforcement redirects skill compilation requests to L2 |
| `skills/smaqit.new-skill/SKILL.md` | New — 6-section gathering flow, validation, compilation via L2 subagent |
| `installer/Makefile` | Added `smaqit.new-skill` to prepare target; updated template filename |

## Design Decisions

- **L2 is the subagent** for skill compilation — L1 owns templates and rules only; L2 compiles all concrete outputs (agents and skills). Initial design assigned this to L1, which was corrected during implementation as an architectural category error.
- **`base-skill.template.md` uses `[SCREAMING_CASE]` placeholders** — same convention as agent templates; L2 resolves them at compile time from the definition file
- **Skill bodies are authored prose, not substitution strings** — L2 writes content using degrees-of-freedom judgment; placeholders are structural anchors in the template, not runtime strings
- **Skill reference files may live in subdirectories** — the constraint is chain depth (no nested references), not directory depth; skill folders may contain `scripts/`, `references/`, `assets/` etc.
- **Task 006 depends on Task 009** — `smaqit.new-principle` should be created using `smaqit.new-skill`, not before it
