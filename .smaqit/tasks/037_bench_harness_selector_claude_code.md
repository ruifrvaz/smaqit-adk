---
status: Not Started
created: "2026-10-11"
---

# Bench Harness Selector: First-Class Claude Code Harness Alongside Codex

## Description

Bench's engine doesn't care which harness it runs: the `process` adapter (`src/bench/adapter.go`) executes whatever `variants[].process.executable` a manifest names. Everything around it, however, assumes Codex CLI:

- the reusable process block in `.smaqit/bench/README.md` (`codex exec --sandbox danger-full-access --skip-git-repo-check --cd {workspace} {briefFile}`)
- the preflight checks and gotchas in `smaqit.bench-run` and `smaqit.bench-scaffold`
- the credential handling that task 035 found leaking Codex credentials

Running a Claude Code Case today means hand-rolling all of the following, and two of the items are real blockers:

1. **Invocation:** headless `claude -p` flags (output format, permission mode, turn limit, working directory) are undocumented in Bench and differ from Codex.
2. **Staging (blocker):** Bench stages variant treatments into the read-only `.smaqit-bench-input/treatment/` sidecar. Claude Code discovers agents, commands and skills only from a project's `.claude/` directory or from its config directory. Personal skills in the config directory override project skills of the same name. On a machine that has the released product installed globally, the installed copies therefore shadow the treatment under test. A sound run needs an isolated `CLAUDE_CONFIG_DIR` per run, with the treatment staged where Claude Code discovers it.
3. **Authentication (blocker):** an isolated config directory has no login. Credentials must reach the harness through an explicit, opt-in source (for example an inherited `ANTHROPIC_API_KEY`, or a credentials file copied in and removed afterwards). They must never be frozen into evidence. Task 035 is the precedent: a leaked workspace held a copy of real Codex CLI credentials.
4. **Evidence:** capture the session transcript (e.g. `--output-format stream-json`) into the run's trace directory, and record the harness name and version in the report.
5. **Conversation turns:** headless Claude Code has no user to answer `AskUserQuestion`. That's sufficient for single-turn Cases. A Case that tests an ask-and-resume loop also needs scripted follow-up turns, for example `--resume <session-id>` with prepared answers.

A **harness selector** makes Claude Code first-class next to Codex. It would be a declared `harness` that expands to a vetted invocation, staging layout, environment and trace capture, while raw `process` stays available as the escape hatch.

**Motivating case:** smaqit task 131 (2026-10-11). It made smaqit agents return `NEEDS-INPUT`/`REVIEW-REQUESTED` signals instead of asking. On Claude Code it preloads each agent's input skill through `skills:` frontmatter, and the main conversation resumes subagents with `SendMessage`. Its live acceptance check depends on behavior specific to Claude Code: whether `skills:` preload actually injects content, that subagents don't get `AskUserQuestion`, and that resume works. A Codex Bench run can't exercise any of that.

## Issue Triage Context

**Mode:** Auto
**Technologies:** Claude Code, Codex CLI, Go
**Platforms/Environments:** Linux, WSL2, macOS
**Features/Integrations:** Bench process adapter, headless claude -p, CLAUDE_CONFIG_DIR isolation, subagents, Agent Skills, session resume, stream-json transcripts
**Versions/Constraints:** Bench schemaVersion 2, smaqit-adk 3.0.x, credentials must never enter frozen evidence, Claude Code personal skills override project skills of the same name

## Design Decisions

TBD — to be confirmed during assessment with `task.plan`. Open questions:

- Is the selector a preset on `process` variants (`variants[].harness: codex | claude`, expanded at plan time) or a new adapter (`adapter: claude`)? Is a `schemaVersion` bump needed?
- Where are treatments staged for Claude Code (an isolated config directory, the project's `.claude/`, or both), and how are shadowing installed copies prevented or detected?
- Which credential sources are supported, and how is their cleanup guaranteed on success, failure and timeout? This relates to task 035.
- Do multi-turn scripted Cases belong here or in a follow-up task?
- How does this relate to task 031's scenario matrix, where harness A/B is already a documented Variant dimension?

## Implementation Steps

1. **Discovery:** confirm the current Claude Code headless contract:
   - flags, output formats and permission modes
   - `CLAUDE_CONFIG_DIR` isolation
   - credential sources
   - agent, skill and command discovery precedence (project vs. personal)
   - session resume and exit codes

   Record the findings in this task.
2. **Schema:** design the harness selector, validate it in `src/bench/manifest.go`, and keep raw `process` as the escape hatch.
3. **Staging:** map treatments of agent, skill and command shape into the locations Claude Code discovers, inside a per-run isolated config directory, and fail when an installed copy would shadow a treatment.
4. **Authentication:** add an explicit opt-in credential source, keep credentials out of frozen evidence, and clean up on every exit path. Fold in or depend on task 035.
5. **Evidence:** write the transcript into the trace directory, and record the harness and version in the report.
6. **Conversation turns** (optional, or a follow-up): scripted answers for resume-style Cases.
7. **Docs and skills:** add a Claude Code process block alongside the Codex one in the README, and add harness selection plus Claude Code preflight and gotchas to `smaqit.bench-run` and `smaqit.bench-scaffold`. Coordinate with task 031's matrix.
8. **Tests and live verification:** unit tests for manifest validation and staging, then a live run against authenticated Claude Code. Dogfood it with smaqit task 131's single-turn `NEEDS-INPUT` Case (thin requirements → `NEEDS-INPUT`, no specification written).

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] A manifest can select the Claude Code harness declaratively, without hand-writing its invocation, and raw `process` variants keep working unchanged
- [ ] Agent, skill and command treatments are discoverable by Claude Code in the run, and a run fails rather than silently testing an installed copy that shadows the treatment
- [ ] Credentials reach the harness only through an explicit opt-in source, never appear in frozen evidence, and are cleaned up on success, failure and timeout
- [ ] The session transcript is captured into the trace directory, and the report records the harness and its version
- [ ] `smaqit.bench-run`, `smaqit.bench-scaffold` and the README document harness selection, with Claude Code preflight and gotchas
- [ ] Unit tests cover selector validation and staging, and a live run against authenticated Claude Code passes for a single-turn Case

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
| `src/bench/manifest.go`, `src/bench/adapter.go`, `src/bench/workspace.go`, `src/bench/process.go` | Modify — selector, Claude Code staging, environment, trace |
| `src/bench/*_test.go` | Modify — selector and staging tests |
| `skills/smaqit.bench-run/SKILL.md`, `skills/smaqit.bench-scaffold/SKILL.md` | Modify — harness selection |
| `.smaqit/bench/README.md`, `docs/wiki/benchmarking.md`, `README.md` | Modify — Claude Code process block and docs |
| `CHANGELOG.md` | Modify |

## Notes

- Filed 2026-10-11 from a smaqit session. smaqit task 131's live Claude Code check is blocked on this task, or on a hand-rolled one-off run with an isolated `CLAUDE_CONFIG_DIR`.
- Related: task 031 (scenario matrix; harness A/B as a Variant dimension), task 035 (credential leak and cleanup), task 033 (`prepare` background-process teardown).
- Claude Code docs to anchor discovery: https://code.claude.com/docs/en/sub-agents, https://code.claude.com/docs/en/skills, https://code.claude.com/docs/en/headless

Child tasks inherit their active parent's branch, worktree, and workflow mode. Only a standalone or parent task owns Git lifecycle cleanup.
