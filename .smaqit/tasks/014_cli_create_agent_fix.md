# CLI create-agent / create-skill Fix

**Status:** Completed  
**Mode:** Assisted  
**Created:** 2026-03-31  
**Started:** 2026-04-09  
**Completed:** 2026-04-09

## Description

The `smaqit-adk create-agent` and `create-skill` CLI commands have known issues discovered during live testing. This task fixes all of them and validates the commands end-to-end with the correct embedded agent.

## Background

The following changes have already been made in the local workspace (not yet committed or released):
- Removed 15-minute timeout (`context.WithTimeout` → `context.WithCancel`)
- Removed progress ticker (`[working... Xs]` noise)
- Swapped system context from `smaqit.L2 + smaqit.new-agent/SKILL.md` to `smaqit.create-agent.agent.md` (self-contained lite-tier agent)
- Same swap for `create-skill` → `smaqit.create-skill.agent.md`

The original problem with L2 + new-agent: the skill's compilation step invokes L2 as a subagent, which is not supported in CLI sessions. The agent completed after one question and wrote nothing.

## Acceptance Criteria

- [x] `smaqit-adk create-agent` uses `smaqit.create-agent.agent.md` as system context
- [x] `smaqit-adk create-skill` uses `smaqit.create-skill.agent.md` as system context
- [x] No session timeout — user can take as long as needed
- [x] No progress ticker output during the session
- [x] Agent scans repo before asking questions; asks only name + description/purpose explicitly; infers remaining sections from stated purpose and repo context; presents full draft for one confirmation pass before compiling
- [x] Output file is written to `.github/agents/[name].agent.md`
- [x] `make build` passes cleanly
- [x] Changes committed and released (patch version bump)

## Notes

- `smaqit.create-agent` is self-contained: all gathering logic and base foundation directives are embedded inline. No subagent invocations, no project file reads.
- Unused imports (`sync/atomic`, `time`) must be removed after ticker removal.
- The `skills/` embed directives for `smaqit.new-agent` and `smaqit.new-skill` are no longer needed by `cmdCreate` — verify whether they are still needed by `cmdInit` before removing.
