# Lite Tier — Routing Skills for Natural Language Entry Points

**Status:** Completed  
**Created:** 2026-04-03

## Description

Add thin routing skills (`smaqit.create-agent`, `smaqit.create-skill`) to the lite tier so users can trigger agent and skill creation via natural language or slash command without explicitly switching agent context. The skills are installed by `smaqit-adk init` alongside the existing agents and delegate all logic to the corresponding subagents.

## Requirements

- [x] `skills/smaqit.create-agent/SKILL.md` created — thin routing skill with semantic description, invokes `smaqit.create-agent` as subagent
- [x] `skills/smaqit.create-skill/SKILL.md` created — thin routing skill with semantic description, invokes `smaqit.create-skill` as subagent
- [x] `installer/main.go` updated — embeds both skill files; `cmdInit` writes them; `cmdUninstall` removes them
- [x] `installer/Makefile` updated — `prepare` target copies lite-tier routing skills into installer
- [x] Structural tests pass — both new skills validate against all 5 skill validators
- [x] Unit tests pass — `init_test.go` updated to assert skill files are installed and uninstalled correctly
- [x] Documentation updated — README, copilot-instructions.md, install.sh reflect natural language entry point framing
- [x] User testing — end-to-end scenario: `smaqit-adk init` on a fresh project, then trigger agent/skill creation via natural language

## Acceptance Criteria

- [x] `skills/smaqit.create-agent/SKILL.md` exists with correct frontmatter, all required sections, and working failure table
- [x] `skills/smaqit.create-skill/SKILL.md` exists with correct frontmatter, all required sections, and working failure table
- [x] `smaqit-adk init` installs `.github/skills/smaqit.create-agent/SKILL.md` and `.github/skills/smaqit.create-skill/SKILL.md`
- [x] `smaqit-adk uninstall` removes both skill files and their directories (and `.github/skills/` if empty)
- [x] Structural tests: 21/21 PASS
- [x] Unit tests: 8/8 PASS
- [x] Build clean (`make build` produces `dist/smaqit-adk-dev` without errors)
- [x] User testing: end-to-end scenario validates that after `init`, saying "create a new agent" or "create a new skill" in VS Code Copilot correctly activates the routing skill and invokes the subagent

## Notes

- Skills and agents share the same identifier (e.g. `smaqit.create-agent`) in different VS Code namespaces — no conflict
- Skills cannot self-invoke as subagents; the skill+agent pairing is the correct model
- Task 014 (CLI create-agent/create-skill Fix) is closely related and may be bundled in the same release
