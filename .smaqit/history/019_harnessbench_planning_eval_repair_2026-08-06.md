# HarnessBench Planning and Eval Repair

## Metadata

- **Date:** 2026-08-06
- **Focus:** Assess a one-shot HarnessBench implementation prompt against smaqit-adk; create tasks from the assessment; execute the narrowest of them (broken eval artifact references) end-to-end
- **Tasks referenced:** Task 023 (created), Task 024 (created, started, completed), Task 025 (created)

## Actions Taken

### 1. Session start

Loaded README, history 018, and PLANNING.md. Four open tasks noted (018–021), none started; identified Task 018 as the nominal next-unblocked item but did not begin it — session was redirected immediately by the next user message.

### 2. HarnessBench assessment

User supplied `assets/HARNESSBENCH_ONE_SHOT_PROMPT.md` (a prior one-shot implementation prompt for a benchmarking CLI) and a diagram, asking for an assessment of building it as a smaqit-adk feature. Ran `smaqit.session-assess`:

- Read the full prompt, the installer (`installer/main.go`, `installer/go.mod`), and the existing behavioral-eval infrastructure (`tests/evals/runner/main.go`, `tests/evals/README.md`, all 7 existing eval files) to ground the assessment empirically rather than from the prompt's own framing.
- Found the prompt targets a different stack entirely — a .NET/C# solution on a "Daisy workflow orchestration engine" that doesn't exist in this Go-based repo. Roughly 40% of the prompt (the Daisy workflow model) is inapplicable; the remaining ~60% (scientific-validity controls, manifest shape, grader taxonomy, scoring/winner-selection, artifact contract, security constraints) is stack-agnostic and worth keeping.
- Found ~25–30% infrastructure overlap with the existing eval runner (workspace isolation outside the repo tree, explicit-token auth to avoid VS Code context leakage, LLM-graded transcripts) but 0% purpose overlap — the eval runner asks "does this artifact behave as specified?", HarnessBench asks "is the kit worth it?"
- Flagged a scope tension: HarnessBench evaluates *any* harness (codex, claude-code, opencode), not just the ADK, which sits awkwardly inside an ADK-scoped repo.
- Surfaced two collateral defects found during the audit: 5 of 7 behavioral evals reference renamed, nonexistent artifact files; and the README documents `smaqit-adk create-agent`/`create-skill` CLI commands that don't exist in the shipped binary's dispatch.

Asked the user three scoping questions (placement, MVP size, how to handle the collateral findings). Decisions: **subcommand of the `smaqit-adk` binary** (against the assessment's recommendation of a separate module — user accepted the dependency-weight cost explicitly); **Phase 1 only** (variants, repetitions, deterministic graders, statistics, winner selection, in-process Copilot SDK — no external process adapters, no worker-process boundary yet); **log the collateral findings as separate tasks, don't fix inline**.

### 3. Task creation

Created three tasks and updated `PLANNING.md`:
- **023** — HarnessBench Phase 1 (`smaqit-adk bench` subcommand), fully scoped with 14 implementation steps and 17 acceptance criteria.
- **024** — Repair Broken Eval Artifact References.
- **025** — README Documents Non-Existent CLI Commands.

Saved two memory entries: the HarnessBench placement decision (so it isn't re-litigated), and a note that smaqit-adk's core thesis — that the kit actually helps — is unproven, which is the strategic reason HarnessBench exists at all.

### 4. Task 024 — start and implementation

User started Task 024 immediately ("let's fix the current eval state and cli"). `task-start` flow: lifecycle resolution required committing pending task-creation changes to `main` first (worktree creation checks out committed state, not the working tree) — user committed those directly. Branch/worktree created; research map built (Go, GitHub CLI, Copilot SDK Go); issue triage run (one advisory — a copilot-sdk hang issue scoped to the unrelated .NET FFI binding; one historical — a closed PAT-auth issue root-caused to a stripped Docker image, not expected to apply here).

Implementation: read the current `smaqit.create-agent`/`smaqit.create-skill` `SKILL.md` files to establish actual behavior (single-question, inference-first, v2.0/v2.1 — a full behavioral inversion from the old turn-by-turn `smaqit.new-agent`/`smaqit.new-skill` skills the broken evals were written against). Repointed and renamed all 5 broken eval files, then rewrote every turn and criterion against the current behavior rather than just fixing the path.

### 5. Two eval-runner bugs found and fixed mid-task

Running the suite to validate surfaced two real, unrelated infrastructure defects, both root-caused with primary-source evidence rather than guessed at:

- **Process leak.** `tests/evals/runner/main.go` creates a `copilot.Client` at two call sites (once per eval, once per graded criterion) and never called `.Stop()`. Confirmed via the vendored SDK source's own documented `defer client.Stop()` pattern, and empirically — 23 orphaned `copilot --headless` processes accumulated in one run, which the user noticed independently ("I see a lot of sessions opened"). Fixed with `defer client.Stop()` at both sites.
- **Wrong template path.** `setupWorkspace()` copied the ADK's `templates/`/`framework/` to the workspace's top level instead of under `.smaqit/`, the path every current skill and agent (`smaqit.L2.agent.md`, `smaqit.create-agent`) actually reads at runtime. Confirmed by inspecting the actual preserved eval workspace on disk (`.smaqit/templates` was simply absent) and by a clean behavioral flip: `L2/001_compile_base_agent`, the one eval requiring an actual template read, went from timing out in every pre-fix run to passing in every post-fix run.

Both fixes were scoped, small, and applied in-session per explicit user confirmation at each step.

### 6. `~/.copilot` session cleanup

The eval runner's process leak had a second-order effect: hundreds of headless sessions accumulated in the user's global `~/.copilot/session-state/` (321 total, spanning back to April). Cross-referenced the 61 sessions created within the testing window against VS Code session metadata to confirm none were real interactive chat history (all non-`vscode`-origin) before proposing deletion — the actual `rm -rf` was blocked by Claude Code's auto-mode safety classifier (correctly, given it reached outside the project into global machine state) and left to the user, who archived them directly.

### 7. Final eval run and Task 024 completion

One clean, uninterrupted run after both fixes: **2/7 passed** (both `smaqit.L2` evals). All 5 `create-agent`/`create-skill` evals fail on a third, distinct, unresolved issue — sessions report being blocked by filesystem permission errors and never write any file, despite the target directories being pre-created with 0755. Not chased further in this task, per its own "record failures honestly, don't tune them away" acceptance criterion.

Completed Task 024 via `smaqit.task-complete`: Findings written, all 5 acceptance criteria verified and checked, status set to Completed, `PLANNING.md` updated (with a note that Task 020 depends on the unresolved permission-write defect before it can produce a genuine pass), branch committed and merged into `main` (no-ff), worktree removed, branch deleted, workspace file rebuilt and committed.

## Problems Solved

- **5 broken eval references:** repointed and rewritten for current skill behavior, not just path-patched.
- **Eval runner process leak:** `defer client.Stop()` added at both `NewClient()` call sites.
- **Eval runner wrong template path:** `setupWorkspace()` now provisions `.smaqit/templates`/`.smaqit/framework`, matching runtime reads.
- **Auth failures mid-session:** resolved by the user adding "Copilot Requests" permission to the fine-grained PAT used for `GH_TOKEN`.
- **A background run killed by an external Claude Code process/harness restart** (no completion record) — required one clean re-run rather than trusting partial output.

## Decisions Made

- HarnessBench ships as a `bench` subcommand of the existing `smaqit-adk` binary, not a separate module — discoverability and one distributable were judged worth the dependency-weight cost, made explicit and accepted.
- HarnessBench Phase 1 is deliberately small: no external harness adapters, no worker-process boundary, no git-diff metrics, no HTML reporting — just enough to answer "is the kit worth it?" with real evidence.
- Task 024's cross-reference was corrected from Task 021 (advanced-tier) to Task 020 (lite-tier) — `create-agent`/`create-skill` are lite-tier artifacts, so 021 was never the right follow-up owner.
- The two eval-runner bugs were fixed in-session rather than deferred, since they directly blocked Task 024's own acceptance criterion of producing a real pass/fail tally, and both were small and well-evidenced.
- The third, unresolved permission-write defect was explicitly *not* chased in Task 024 — recorded as a required follow-up instead, to avoid unbounded scope creep on a task that was meant to be a narrow repair.

## Files Modified

| File | Change |
|------|--------|
| `.smaqit/tasks/023_harnessbench_phase1_bench_subcommand.md` | Created |
| `.smaqit/tasks/024_repair_broken_eval_artifact_references.md` | Created, then completed with full Findings |
| `.smaqit/tasks/025_readme_documents_nonexistent_cli_commands.md` | Created |
| `.smaqit/tasks/020_lite_tier_behavioral_evals.md` | Noted the eval repair and the new permission-write blocker |
| `.smaqit/tasks/PLANNING.md` | 023/024/025 added; 024 moved to Completed |
| `.smaqit/references/project-research.md` | Created (Go, GitHub CLI, Copilot SDK Go) |
| `tests/evals/skills/smaqit.new-agent/*.json` → `tests/evals/skills/smaqit.create-agent/*.json` | Repointed, renamed, rewritten (3 files) |
| `tests/evals/skills/smaqit.new-skill/*.json` → `tests/evals/skills/smaqit.create-skill/*.json` | Repointed, renamed, rewritten (2 files) |
| `tests/evals/runner/main.go` | Fixed process leak (`defer client.Stop()`) and wrong template path (`.smaqit/templates`) |
| `smaqit-adk.code-workspace` | Rebuilt after worktree cleanup |

## Next Steps

- **Highest priority for next session: Task 023 (HarnessBench Phase 1) planning and implementation.** The task file is fully scoped (14 implementation steps, 17 acceptance criteria) — ready to start directly with `task.start 023`.
- **Needs a task before Task 020 can start:** the unresolved permission-write defect blocking all `create-agent`/`create-skill` evals. Pre-created directories rule out the obvious cause; genuine investigation is needed. Not yet filed as its own task.
- **Task 025** (README CLI docs) is small and independent — can be picked up any time.
- **Task 018** (Level Skills Completion) remains the critical-path blocker for Tasks 019 and 021, untouched this session.

## Session Metrics

- Duration: Single extended session
- Tasks created: 3 (023, 024, 025)
- Tasks completed: 1 (024)
- Eval-runner bugs found and fixed: 2 (process leak, wrong template path)
- Eval-runner bugs found and deferred: 1 (permission-blocked writes)
- Files modified: 15 (across the merge commit) + workspace rebuild
- Key outcome: HarnessBench (Task 023) fully scoped and ready to start; eval infrastructure measurably healthier (0 read-artifact errors, 0 process leaks, correct template path) even though the lite-tier evals themselves don't yet pass
