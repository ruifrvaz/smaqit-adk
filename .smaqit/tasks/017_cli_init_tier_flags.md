# CLI Tier Subcommands — Replace `init` with `lite` and `advanced`

**Status:** In Progress  
**Created:** 2026-04-03

## Description

Replace the `smaqit-adk init` subcommand with two explicit tier subcommands. Installation intent is expressed through the subcommand name — no flags needed.

- `smaqit-adk lite` — installs lite tier: `.github/agents/` (create-agent, create-skill) + `.github/skills/` (routing skills). Equivalent to current `init`.
- `smaqit-adk advanced` — installs the full ADK into `.smaqit/`: framework files, templates, L0/L1/L2 agents, and advanced-tier skills. Enables the full compilation chain locally.

`init` is removed. No default fallback — the user always states intent explicitly.

## Requirements

- [x] `smaqit-adk lite` installs what `init` installs today: 2 agents + 2 routing skills into `.github/`
- [x] `smaqit-adk advanced` installs framework files, templates, L0/L1/L2 agents, and advanced skills into `.smaqit/`
- [x] `smaqit-adk init` is removed (prints a clear migration message and exits non-zero)
- [x] `uninstall` handles both tiers (removes what was installed, or accepts `lite`/`advanced` argument)
- [x] `printUsage()` and `help` updated to document `lite` and `advanced` instead of `init`
- [x] Unit tests updated: renamed `init` test cases to `lite`; added `advanced` install and uninstall cases
- [x] `install.sh` and `README.md` updated to use new subcommand names

## Acceptance Criteria

- [x] `smaqit-adk lite` produces identical file output to current `smaqit-adk init`
- [x] `smaqit-adk advanced` writes L0/L1/L2 agents, framework, templates, and advanced skills to `.smaqit/`
- [x] `smaqit-adk init` is gone or prints a clear migration message
- [x] Unit tests: all lite install/uninstall/idempotent/already-exists cases pass
- [x] Unit tests: advanced install and uninstall cases pass
- [x] Build clean
- [x] `README.md` Quick Start uses `smaqit-adk lite`
- [x] All unit tests pass (`cd tests && go test ./unit/...`)
- [x] All structural tests pass (`cd tests && go test ./structural/...`)
- [ ] Changes committed

## Notes

- Depends on Task 015 (Full Compilation Chain CLI) for the advanced install content — may be implemented together
- `create-agent` and `create-skill` subcommands are unaffected
- This is a breaking change to the CLI surface — version bump required at release
