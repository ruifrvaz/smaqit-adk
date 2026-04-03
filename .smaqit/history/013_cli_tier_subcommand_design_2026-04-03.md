# CLI Tier Subcommand Design + Release adk-v0.4.0

**Date:** 2026-04-03  
**Focus:** Retroactive task capture for routing skills work + CLI tier subcommand redesign + release adk-v0.4.0  
**Tasks referenced:** 016 (completed), 017 (created, high priority)

## Actions Taken

- Created Task 016 retroactively — captures all routing skills implementation work from the previous session (skills/smaqit.create-agent, smaqit.create-skill, installer changes, tests, docs) with 7/8 acceptance criteria pre-checked; user testing is the single remaining criterion
- Answered local install question: `cd installer && make build` produces `dist/smaqit-adk-dev`; no `make install` target exists
- Created Task 017 — initially drafted as `init --lite` / `init --advanced` flags, then immediately refined
- Refined Task 017 design: replace `smaqit-adk init` entirely with two explicit subcommands — `smaqit-adk lite` and `smaqit-adk advanced`; marked high priority
- User testing completed successfully — confirmed natural language entry point works after `smaqit-adk init`; Task 016 closed as Completed (8/8 criteria)
- Released `adk-v0.4.0` — full release workflow executed: analysis → approval → CHANGELOG → commits → tag → push

## Decisions Made

- **`init` is removed, not flagged**: `smaqit-adk lite` and `smaqit-adk advanced` are the new top-level subcommands. Installation intent is stated explicitly via the subcommand name — no flags, no default fallback.
- **`lite`** = current `init` behavior: 2 agents + 2 routing skills into `.github/`
- **`advanced`** = full ADK install into `.smaqit/`: framework, templates, L0/L1/L2, advanced skills
- Task 017 is a breaking change — requires a version bump at release
- Task 017 still depends on Task 015 (Full Compilation Chain CLI) for the advanced install content, but is high priority and can proceed on the `lite` side independently
- Release severity assessed as **MINOR** — new routing skills feature, no breaking changes

## Problems Solved

- No explicit `make install` target exists — noted for future addition (or use absolute path from `dist/`)
- Task 016 had no tracking entry despite being fully implemented — retroactively created with correct status and pre-checked criteria

## Files Modified

- `.smaqit/tasks/016_lite_tier_routing_skills.md` — created (retroactive task), then completed
- `.smaqit/tasks/017_cli_init_tier_flags.md` — created, then rewritten to reflect `lite`/`advanced` subcommand design
- `.smaqit/tasks/PLANNING.md` — added 016 (completed), added 017 (Not Started, HIGH PRIORITY)
- `CHANGELOG.md` — `[0.4.0]` entry added with Added/Changed sections

## Release: adk-v0.4.0

**Commits pushed to main:**
1. `feat: add lite-tier routing skills (smaqit.create-agent, smaqit.create-skill)` — new skill files, installer, unit tests
2. `docs: update README, install.sh, copilot-instructions for natural language UX`
3. `chore: eval runner improvements and README cleanup`
4. `chore: session history and task tracking updates` — history 011, 012, 013; tasks 016, 017
5. `Release adk-v0.4.0` — CHANGELOG only

**Tag:** `adk-v0.4.0` (annotated) — pushed to `origin`

## Next Steps

- Task 017: implement `lite`/`advanced` subcommands — can start `lite` side without Task 015; `advanced` gated on 015
- Consider adding `make install` target to Makefile for local dev convenience

## Session Metrics

- Duration: medium  
- Tasks created: 2 (016, 017); tasks completed: 1 (016)  
- Files created: 2 task files  
- Files modified: CHANGELOG.md, PLANNING.md  
- Release: adk-v0.4.0 — commit `afeb318`, tag pushed to origin
