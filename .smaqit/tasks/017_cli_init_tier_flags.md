# CLI Tier Subcommands — Replace `init` with `lite` and `advanced`

**Status:** Not Started  
**Created:** 2026-04-03

## Description

Replace the `smaqit-adk init` subcommand with two explicit tier subcommands. Installation intent is expressed through the subcommand name — no flags needed.

- `smaqit-adk lite` — installs lite tier: `.github/agents/` (create-agent, create-skill) + `.github/skills/` (routing skills). Equivalent to current `init`.
- `smaqit-adk advanced` — installs the full ADK into `.smaqit/`: framework files, templates, L0/L1/L2 agents, and advanced-tier skills. Enables the full compilation chain locally.

`init` is removed. No default fallback — the user always states intent explicitly.

## Requirements

- [ ] `smaqit-adk lite` installs what `init` installs today: 2 agents + 2 routing skills into `.github/`
- [ ] `smaqit-adk advanced` installs framework files, templates, L0/L1/L2 agents, and advanced skills into `.smaqit/`
- [ ] `smaqit-adk init` is removed (or aliased with a deprecation notice — TBD)
- [ ] `uninstall` handles both tiers (removes what was installed, or accepts `lite`/`advanced` argument)
- [ ] `printUsage()` and `help` updated to document `lite` and `advanced` instead of `init`
- [ ] Unit tests updated: rename `init` test cases to `lite`; add `advanced` install and uninstall cases
- [ ] `install.sh` and `README.md` updated to use new subcommand names

## Acceptance Criteria

- [ ] `smaqit-adk lite` produces identical file output to current `smaqit-adk init`
- [ ] `smaqit-adk advanced` writes L0/L1/L2 agents, framework, templates, and advanced skills to `.smaqit/`
- [ ] `smaqit-adk init` is gone or prints a clear migration message
- [ ] Unit tests: all lite install/uninstall/idempotent/already-exists cases pass
- [ ] Unit tests: advanced install and uninstall cases pass
- [ ] Build clean
- [ ] `README.md` Quick Start uses `smaqit-adk lite`

## Notes

- Depends on Task 015 (Full Compilation Chain CLI) for the advanced install content — may be implemented together
- `create-agent` and `create-skill` subcommands are unaffected
- This is a breaking change to the CLI surface — version bump required at release
