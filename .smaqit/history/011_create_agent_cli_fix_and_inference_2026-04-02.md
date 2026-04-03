# Create Agent CLI Fix and Inference

## Metadata

- **Date:** 2026-04-02
- **Session focus:** Fix `create-agent`/`create-skill` CLI bugs and add repo-scanning inference to agent gathering flow
- **Tasks completed:** Task 014 (CLI create-agent / create-skill Fix)
- **Tasks referenced:** Task 006, 011, 013, 015

## Actions Taken

- Recapped session state after context loss; identified all completed local changes not yet released
- Updated Task 014 acceptance criteria to reflect inference-based gathering (replacing "ask all 8 sections explicitly" with "scan repo, ask 2 things, present full draft")
- Modified `agents/smaqit.create-agent.agent.md`:
  - Added `read` and `search` to frontmatter tools
  - Updated Input section to include project files as an input source
  - Replaced sequential 8-question gathering directive with: scan repo → ask only name + description/purpose → infer remaining sections → present full draft for one confirmation pass
  - Removed MUST NOT prohibition on reading project files
- Applied identical changes to `agents/smaqit.create-skill.agent.md`
- Confirmed `qa-helper.agent.md` deletion (test artifact, not ADK catalog)
- Updated `installer/main.go` version constant to `0.3.2`
- Finalized CHANGELOG.md with `[0.3.2]` section
- Released `adk-v0.3.2` with 4 logical commits + annotated tag pushed to remote

## Problems Solved

- **Wrong agent in cmdCreate**: `smaqit.L2 + smaqit.new-agent` was the system context; the skill tries to invoke L2 as a subagent which CLI doesn't support — swapped to self-contained `smaqit.create-agent`
- **Session timeout**: 15-minute `WithTimeout` replaced with `WithCancel`; user exits via Ctrl-C only
- **Progress ticker noise**: ticker goroutine and `inputActive` atomic removed entirely
- **Too many questions**: `create-agent` was asking each of 8 specification sections one by one; new flow scans repo context first, asks 2 things, infers the rest and confirms in one pass

## Decisions Made

- **Inference approach (Option A)**: Changed the agent directive rather than creating a CLI-specific variant agent — improves both CLI and VS Code paths with no file divergence
- **Repo scan scope**: reads existing agents in `.github/agents/`, existing skills in `.github/skills/`, README, and config/manifest files — broad enough to infer domain and tooling conventions without being unbounded
- **`qa-helper.agent.md` dropped**: confirmed as a test artifact from eval work; not part of the shipped ADK agent catalog
- **4 logical commits**: grouped by purpose (agent files, main.go, build/eval infra, release prep) rather than one monolithic commit

## Files Modified

| File | Change |
|------|--------|
| `agents/smaqit.create-agent.agent.md` | Added `read`/`search` tools; updated Input section; replaced 8-question gathering with repo-scan + inference flow; removed project-file prohibition |
| `agents/smaqit.create-skill.agent.md` | Same changes as create-agent |
| `agents/qa-helper.agent.md` | Deleted (test artifact) |
| `installer/main.go` | Version `0.3.1` → `0.3.2`; swapped embed directives to `adkCreateAgentFile`/`adkCreateSkillFile`; removed `WithTimeout`; removed ticker goroutine; removed unused imports |
| `installer/Makefile` | Eval target auto-detects GitHub token via `gh auth token` |
| `CHANGELOG.md` | Added `[0.3.2]` section |
| `.smaqit/tasks/014_cli_create_agent_fix.md` | Created; acceptance criteria updated to reflect inference model |
| `.smaqit/tasks/015_full_compilation_chain_cli.md` | Created (design task, not started) |
| `.smaqit/tasks/PLANNING.md` | Added Task 014 and 015 to Active table |
| `tests/evals/runner/main.go` | Updated (eval runner improvements) |
| `tests/evals/README.md` | Updated |
| `tests/evals/host-setup.sh` | Created |

## Next Steps

- **Task 006** — Create `smaqit.new-principle` skill (unblocks Task 013)
- **Task 011** — Interactive CLI: `create-principle` and `validate` commands (deferred from prior session)
- **Task 013** — CLI `create-principle` and `validate` commands
- **Task 015** — Full compilation chain CLI (L0→L1→L2 as three sequential SDK sessions, file-chained in Go)

## Session Metrics

- Duration: multi-day context (prior session changes + this session completion)
- Tasks completed: 1 (Task 014)
- Files created: 3 (task files + host-setup.sh)
- Files modified: 8
- Files deleted: 1 (`qa-helper.agent.md`)
- Release: `adk-v0.3.2` — 4 commits, tag pushed
