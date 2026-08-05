# README Documents Non-Existent CLI Commands

**Status:** Not Started
**Created:** 2026-08-05

## Description

[README.md](../../README.md) documents two CLI commands that the shipped binary does not implement. The command dispatch in [installer/main.go:39-70](../../installer/main.go#L39-L70) handles only `lite`, `advanced`, `init`, `help`, `uninstall`, and `version` — there is no `create-agent` or `create-skill` case, and no corresponding `cmdCreate*` function exists in the file.

The README nonetheless presents them as working features in three places:

| Location | Claim |
|---|---|
| [README.md:16-19](../../README.md#L16-L19) | "Advanced tier — a globally installed CLI… `smaqit-adk create-agent`, `smaqit-adk create-skill`" |
| [README.md:131-140](../../README.md#L131-L140) | Commands table lists both with `--output <dir>` |
| [README.md:150-167](../../README.md#L150-L167) | A full "CLI (Advanced Tier)" section with usage examples and `COPILOT_GITHUB_TOKEN` auth setup |

A user following the README today gets a usage dump and exit code 1.

The history is consistent with a feature that was built and then withdrawn: Task 011 delivered the interactive CLI, Task 014 fixed it and shipped in adk-v0.5.0 — but Tasks 013 and 015 are both deferred in PLANNING.md with the note *"CLI work paused"* / *"VS Code-native approach taken via smaqit.compile skills"*. The commands appear to have been removed from the dispatch when that pivot happened, without the README following.

## Design Decisions

- **Establish ground truth before editing.** Confirm via `git log -S` whether the dispatch cases were deliberately removed or lost in a merge. If they were removed deliberately, the README is simply stale and should be corrected. If they were lost accidentally, this becomes a restore decision rather than a docs fix — surface that to the user rather than deciding unilaterally.
- **Do not silently delete user-facing capability claims.** If the commands are genuinely withdrawn, say what replaced them (the lite-tier routing skills and `/smaqit.create-agent` in Copilot chat) rather than removing the section and leaving a gap.

## Implementation Steps

1. Run `git log -S'case "create-agent"' -- installer/main.go` and `git log --oneline -- installer/` around the adk-v0.5.0 → v0.6.0 range to determine whether removal was deliberate.
2. Report the finding and confirm intent with the user: **correct the docs** (commands withdrawn) or **restore the commands** (accidental loss).
3. If correcting the docs: update all three README locations, ensuring the "Advanced tier" description no longer promises a standalone interactive CLI, and point users to the in-editor routing skills instead.
4. Check `installer/main.go`'s `printUsage`/`cmdHelp` output and `docs/wiki/` for the same claim and correct any other occurrence.
5. Add a CHANGELOG entry under `[Unreleased]` recording the documentation correction.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] Determined from git history whether the CLI commands were deliberately removed or accidentally lost, and the finding reported
- [ ] User confirmed the intended resolution (correct docs vs. restore commands) before any edit
- [ ] Every README claim about `smaqit-adk create-agent` / `create-skill` matches the binary's actual behavior
- [ ] `printUsage`, `cmdHelp`, and `docs/wiki/` carry no contradicting claim
- [ ] CHANGELOG `[Unreleased]` entry added
- [ ] Every command documented in README is verified to exist by running the built binary

## Findings

[Populated by smaqit.task-complete. Do not fill in manually before task is complete.]

**Implementation approach:**
- TBD

**Decisions made:**
- TBD

**Blockers encountered:**
- TBD

**Follow-up identified:**
- TBD

## Files to Create / Modify

| File | Action |
|------|--------|
| `README.md` | Modify — three locations |
| `installer/main.go` | Modify — only if `printUsage`/`cmdHelp` contradict, or if restoring commands |
| `docs/wiki/*.md` | Modify — if the claim appears there |
| `CHANGELOG.md` | Modify — `[Unreleased]` entry |

## Notes

This is a small task, but it carries a signal worth heeding before Task 023 lands: **CLI-shaped scope in this repo has been built and then abandoned once already.** Task 023 adds a substantial new subcommand surface to the same binary. Whatever caused the earlier pivot away from the interactive CLI is worth understanding before repeating the shape.
