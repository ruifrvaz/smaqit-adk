# Task 013: CLI create-principle and validate Commands

**Status:** Not Started
**Created:** 2026-03-29

## Description

Follow-on to Task 011. Implements the two remaining `smaqit-adk` CLI commands that were deferred because they depend on work not yet done at Task 011 close time.

- `create-principle` — blocked on Task 006 (the `smaqit.new-principle` skill must exist before it can be embedded)
- `validate` — blocked on a design decision (what eval criteria apply to user-compiled files)

Both commands follow the same `cmdCreate` pattern established in Task 011. Once the blockers are resolved, the Go implementation is small.

## Acceptance Criteria

- [ ] `smaqit-adk create-principle` runs interactively from any directory, gathers principle spec via `smaqit.new-principle` skill, and writes a principle file into `.smaqit/framework/` of the current project (depends on Task 006)
- [ ] `smaqit-adk validate <file>` runs eval criteria against a compiled agent or skill file and reports pass/fail (design decision required first — see Notes)
- [ ] Both commands appear in `smaqit-adk help` output with accurate descriptions
- [ ] Binary builds clean after both additions

## Notes

**`create-principle` implementation** (after Task 006 completes):
- Add `//go:embed skills/smaqit.new-principle/SKILL.md` to `installer/main.go`
- Add `create-principle` switch case wired to `cmdCreate("principle", outputDir)`
- Default output dir: `./.smaqit/framework/`
- Makefile `prepare` already copies `skills/` — no Makefile changes needed

**`validate` design question** (must be answered before implementation):
- What does "valid" mean for a user-compiled agent or skill file?
- Options: (A) structural — required sections present, frontmatter well-formed; (B) behavioral — compiled output matches the definition file it was gathered from; (C) heuristic — all template placeholders resolved, no unresolved `[[...]]` markers
- Option A is cheapest and immediately implementable without the Copilot SDK
- Option B requires a definition file to exist alongside the compiled file — introduces a storage convention not yet decided
- Option C is a useful lint check and complements structural checks

**Dependency chain:**
- Task 013 (`create-principle`) → Task 006 (`smaqit.new-principle` skill)
- Task 013 (`validate`) → design decision (standalone, no external task dependency)
