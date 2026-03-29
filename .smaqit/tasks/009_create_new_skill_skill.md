# Task 009: Create smaqit.new-skill Skill

**Status:** In Progress
**Created:** 2026-03-02

## Objective

Create the `smaqit.new-skill` skill — a meta-skill that guides creation of new skills. This skill is a prerequisite for Task 006 (`smaqit.new-principle`), which should itself be created using `smaqit.new-skill`.

## Acceptance Criteria

- [ ] `templates/skills/base-skill.template.md` authored with all placeholders
- [ ] `templates/skills/compiled/skill.rules.md` contains full L1 directives (not vocabulary-only)
- [ ] `agents/smaqit.L1.agent.md` scope extended to include skill compilation
- [ ] `skills/smaqit.new-skill/SKILL.md` created following the `smaqit.new-agent` pattern
- [ ] `installer/Makefile` prepare target includes `smaqit.new-skill`
- [ ] `cd installer && make build` passes without errors
- [ ] Installer copies `skills/smaqit.new-skill/SKILL.md` into embedded directory

## Files

| File | Change |
|------|--------|
| `templates/skills/base-skill.template.md` | Renamed from `skill.template.md`; fully authored with 8 placeholders |
| `templates/skills/compiled/skill.rules.md` | Added frontmatter, Source L0 Principles, Placeholder Catalog, L1 Directive Compilation, Compilation Guidance |
| `agents/smaqit.L1.agent.md` | Extended scope, input, output, completion criteria, and failure handling for skill compilation |
| `skills/smaqit.new-skill/SKILL.md` | New — 6-section gathering flow, validation, compilation via L1 subagent |
| `installer/Makefile` | Added `smaqit.new-skill` to prepare target; updated template filename |

## Design Decisions

- **L1 is the subagent** for skill compilation — L1 already owns template/skills/ machinery; extending to compile final skill output is the natural fit
- **`base-skill.template.md` uses `[SCREAMING_CASE]` placeholders** — same convention as agent templates; L1 resolves them at compile time from the definition file
- **Skill bodies are authored prose, not substitution strings** — L1 writes content using degrees-of-freedom judgment; placeholders are structural anchors in the template, not runtime strings
- **Task 006 depends on Task 009** — `smaqit.new-principle` should be created using `smaqit.new-skill`, not before it
