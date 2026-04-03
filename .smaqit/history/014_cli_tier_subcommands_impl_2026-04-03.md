# CLI Tier Subcommands Implementation

**Date:** 2026-04-03  
**Focus:** Implementation of task 017 — replace `smaqit-adk init` with `lite` and `advanced` subcommands  
**Tasks completed:** 017  
**Tasks referenced:** 015 (advanced content dependency, satisfied)

## Actions Taken

- Loaded session context: README, PLANNING.md, task 017, history 013
- Explored codebase: `installer/main.go`, `installer/Makefile`, `tests/unit/`, `tests/structural/`, `agents/`, `framework/`, `templates/`, `skills/`
- Implemented `smaqit-adk lite` — identical behavior to removed `init`; installs 2 agents + 2 routing skills into `.github/`
- Implemented `smaqit-adk advanced` — installs L0/L1/L2 agents, framework, templates, and advanced skills (smaqit.new-agent, smaqit.new-skill) into `.smaqit/`
- Changed `smaqit-adk init` to print a clear migration message and exit non-zero (breaking change)
- Extended `uninstall` to accept optional `lite` or `advanced` argument; without argument removes all installed tiers
- Added `copyEmbedDir` helper with error context wrapping for recursive embed.FS → disk copy
- Added embed directives for `framework/`, `templates/`, `skills/smaqit.new-agent/SKILL.md`, `skills/smaqit.new-skill/SKILL.md`
- Rewrote `tests/unit/init_test.go` to contain only shared helpers (`TestMain`, `runBinary`) + non-tier tests
- Created `tests/unit/lite_test.go` with renamed lite tests and new advanced/migration tests (13 tests total)
- Updated `tests/unit/embed_test.go` to use `mustLite` instead of `mustInit`
- Updated `README.md`: Quick Start now shows `smaqit-adk lite`; Commands table shows both `lite` and `advanced`
- Updated `install.sh`: `smaqit-adk init` → `smaqit-adk lite` in user-facing messages
- Updated `installer/Makefile` test-scaffold: uses `lite` and validates `.github/` structure
- Marked task 017 Completed in both task file and PLANNING.md
- Ran parallel validation (Code Review + CodeQL): 0 security alerts; addressed error wrapping feedback in `copyEmbedDir`

## Decisions Made

- **`advanced` writes exclusively to `.smaqit/`**: L0/L1/L2 agents, framework, templates, advanced skills all go to `.smaqit/`. Lite tier (.github/) is independent — a user can run both.
- **`init` exits non-zero with migration message**: Not silently aliased. Users see explicit guidance pointing to `lite` and `advanced`.
- **`uninstall` accepts optional tier arg**: `uninstall lite`, `uninstall advanced`, or bare `uninstall` (removes all detected tiers). Detects what is installed and skips cleanly if nothing found.
- **Test structure split**: `init_test.go` reduced to shared helpers + non-tier tests; `lite_test.go` owns all tier-specific tests. No duplication.
- **Error context in `copyEmbedDir`**: Each failure includes the path, making advanced install failures debuggable.

## Problems Solved

- All existing tests used `mustInit` referencing `init` command — updated to `mustLite` without disrupting structural tests
- `embed.FS` with nested directories: used `//go:embed framework` and `//go:embed templates` (directory embed), walked with `io/fs.WalkDir`
- Uninstall for advanced tier cannot remove all of `.smaqit/` (contains user content) — removed only the four ADK-specific subdirectories: `agents/`, `framework/`, `templates/`, `skills/`

## Files Modified

- `installer/main.go` — complete rewrite of switch, added `cmdLite`, `cmdAdvanced`, `copyEmbedDir`, refactored `cmdUninstall(tier string)`, updated `printUsage`, `cmdHelp`; added embed directives and `io/fs` import
- `installer/Makefile` — test-scaffold updated to use `lite` and validate `.github/` structure
- `tests/unit/init_test.go` — stripped to helpers + `TestCmdUninstall_NotInitialized` + `TestCmdVersion`
- `tests/unit/lite_test.go` — new file: all lite and advanced tests (13 test functions)
- `tests/unit/embed_test.go` — `mustInit` → `mustLite`
- `README.md` — Quick Start and Commands table updated
- `install.sh` — user-facing `init` references replaced with `lite`
- `.smaqit/tasks/017_cli_init_tier_flags.md` — all criteria checked, status Completed
- `.smaqit/tasks/PLANNING.md` — task 017 moved to Completed with date

## Next Steps

- Task 017 is a breaking change — version bump required at release (MINOR or MAJOR depending on versioning policy)
- Task 014 (CLI create-agent/create-skill fix) still pending
- Task 015 (Full Compilation Chain CLI) still pending
- Task 006, 011, 013 still active

## Session Metrics

- Duration: short  
- Tasks completed: 1 (017)  
- Files created: 1 (`tests/unit/lite_test.go`)  
- Files modified: 8  
- Tests: 13 unit tests pass; structural tests pass; build clean  
- Commit: `9c58bdc` on `copilot/vscode-mniz4m1g-j671`
