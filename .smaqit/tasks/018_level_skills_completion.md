# Level Skills Completion

**Status:** Not Started  
**Created:** 2026-04-05  
**Updated:** 2026-04-16

## Description

Complete the skill layer for all Level agents. Every Level agent must have at least one scoped skill entry point that follows the gather → definition file → subagent invocation pattern. All Level skills are shipped by `smaqit-adk advanced`.

Absorbs Task 006 (smaqit.new-principle skill).

## Current State

| Level | Agent | Skill(s) | Gap |
|-------|-------|----------|-----|
| L0 | `smaqit.L0` | `smaqit.new-principle` ✅ (shipped adk-v0.6.0) | ❌ definition file input pattern not yet on L0 agent |
| L1 | `smaqit.L1` | none | ❌ missing |
| L2 | `smaqit.L2` | `smaqit.create-agent`, `smaqit.create-skill` | ✅ complete (inference-first, adk-v0.6.0) |

## Deliverables

### L0 prerequisites

- Update `smaqit.L0` agent to accept a definition file input pattern (`.smaqit/definitions/principles/[name].md`), matching the pattern already established for L2

### New skills

- `smaqit.new-template` — gathers template structure specs; writes `.smaqit/definitions/templates/[name].md`; invokes `smaqit.L1`; output: `templates/[agents|skills]/[name].template.md`
- `smaqit.new-rules` — gathers compilation rule specs; writes `.smaqit/definitions/rules/[name].md`; invokes `smaqit.L1`; output: `templates/[agents|skills]/compiled/[name].rules.md`

### Installer update

- `smaqit-adk advanced` installs all three Level skills: `smaqit.new-principle` (already done), `smaqit.new-template`, `smaqit.new-rules`

## Acceptance Criteria

- [x] `skills/smaqit.new-principle/SKILL.md` created and shipped (adk-v0.6.0)
- [ ] `smaqit.L0` updated with definition file input pattern (same pattern as L2)
- [ ] `skills/smaqit.new-template/SKILL.md` created — invokes `smaqit.L1`; output is a new `.template.md` file
- [ ] `skills/smaqit.new-rules/SKILL.md` created — invokes `smaqit.L1`; output is a new `.rules.md` file in `compiled/`
- [ ] `smaqit-adk advanced` installs new-template and new-rules into `.github/skills/`
- [ ] `make build` passes cleanly
- [ ] Structural tests pass

## Dependencies

- Task 017 (CLI tier subcommands) — completed; `smaqit-adk advanced` install target is ready
- adk-v0.6.0 — `smaqit.new-principle` shipped; inference-first pattern established via `smaqit.create-*`

## Notes

- Follow the inference-first pattern established in `smaqit.create-agent` and `smaqit.create-skill` (v2.0.0)
- Definition file namespace: `.smaqit/definitions/principles/`, `.smaqit/definitions/templates/`, `.smaqit/definitions/rules/`
- `smaqit.new-principle` scope note: targets the ADK framework only — product-domain principles belong in the product extension, not the ADK
- Task 006 absorbed: fully satisfied by adk-v0.6.0
- Unblocks Task 019 (Cross-Level Compilation)
