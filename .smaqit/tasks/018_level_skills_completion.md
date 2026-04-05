# Level Skills Completion

**Status:** Not Started  
**Created:** 2026-04-05

## Description

Complete the skill layer for all Level agents. Every Level agent must have at least one scoped skill entry point that follows the gather → definition file → subagent invocation pattern. All Level skills are shipped by `smaqit-adk advanced`.

Absorbs Task 006 (smaqit.new-principle skill).

## Current State

| Level | Agent | Skill(s) | Gap |
|-------|-------|----------|-----|
| L0 | `smaqit.L0` | none | ❌ missing |
| L1 | `smaqit.L1` | none | ❌ missing |
| L2 | `smaqit.L2` | `smaqit.new-agent`, `smaqit.new-skill` | ✅ complete |

## Deliverables

### L0 prerequisites

- Update `smaqit.L0` agent to accept a definition file input pattern (`.smaqit/definitions/principles/[name].md`), matching the pattern already established for L2

### New skills

- `smaqit.new-principle` — gathers principle name, one-line bold statement, explanation paragraph(s), and target framework file; writes `.smaqit/definitions/principles/[name].md`; invokes `smaqit.L0`; output: target `framework/*.md` file updated
- `smaqit.new-template` — gathers template structure specs; writes `.smaqit/definitions/templates/[name].md`; invokes `smaqit.L1`; output: `templates/[agents|skills]/[name].template.md`
- `smaqit.new-rules` — gathers compilation rule specs; writes `.smaqit/definitions/rules/[name].md`; invokes `smaqit.L1`; output: `templates/[agents|skills]/compiled/[name].rules.md`

### Installer update

- `smaqit-adk advanced` installs all five Level skills: `smaqit.new-principle`, `smaqit.new-template`, `smaqit.new-rules`, `smaqit.new-agent`, `smaqit.new-skill`

## Acceptance Criteria

- [ ] `smaqit.L0` updated with definition file input pattern (same pattern as L2)
- [ ] `skills/smaqit.new-principle/SKILL.md` created — description follows Description-Driven Activation principle
- [ ] `smaqit.new-principle` gathering: principle name, one-line bold statement, explanation paragraph(s), target framework file, rationale for framework-level classification
- [ ] `smaqit.new-principle` validation: name not already present in target file, statement is a single bold imperative sentence, explanation doesn't contradict existing principles
- [ ] `smaqit.new-principle` writes definition to `.smaqit/definitions/principles/[name].md`, then invokes `smaqit.L0` as subagent
- [ ] `skills/smaqit.new-template/SKILL.md` created — invokes `smaqit.L1`; output is a new `.template.md` file
- [ ] `skills/smaqit.new-rules/SKILL.md` created — invokes `smaqit.L1`; output is a new `.rules.md` file in `compiled/`
- [ ] `smaqit-adk advanced` installs all five Level skills into `.smaqit/skills/`
- [ ] `make build` passes cleanly
- [ ] Structural tests pass

## Dependencies

- Task 009 (smaqit.new-skill) — completed; provides the pattern reference for all new skills
- Task 017 (CLI tier subcommands) — completed; `smaqit-adk advanced` install target is ready

## Notes

- Follow the gather → definition file → subagent pattern established in `smaqit.new-agent` and `smaqit.new-skill`
- Definition file namespace: `.smaqit/definitions/principles/`, `.smaqit/definitions/templates/`, `.smaqit/definitions/rules/`
- `smaqit.new-principle` scope note: targets the ADK framework only — product-domain principles belong in the product extension, not the ADK
- Task 006 absorbed: all acceptance criteria from 006 are included above
- Unblocks Task 019 (Cross-Level Compilation)
