# Registry-Driven Cascade Live Verification

**Date:** 2026-08-24
**Session focus:** Plan, implement, live-verify, and ship Task 032 (Registry-Driven Principle Propagation); diagnose and resolve a release-pipeline blocker that surfaced after a mid-session repo migration.
**Tasks completed/referenced:** 032 completed (PR #1, released v2.1.0); 035 created (Bench engine gaps found live); 034 referenced only (concurrent repo migration to `daisy-tribe/smaqit-adk`, not worked on this session).

## Actions Taken

- Started the session, reviewed active tasks (032, 031, 025).
- Planned Task 032 via `smaqit.task-plan`, resolving two design forks through direct source discovery rather than the task file's original assumptions.
- Started Task 032 (worktree, branch, Assisted mode); converted its task file to YAML frontmatter (a pre-existing repo inconsistency also found on other older task files); refreshed the project research map.
- Implemented the registry engine (`installer/registry.go` + `registry_cli.go`: hidden `registry get/put` and public `sync` CLI subcommands), updated `agents/smaqit.L0/L1/L2.md`, revised `skills/smaqit.new-principle`, and authored `smaqit.new-template`/`smaqit.new-rules` — all compiled via a real `smaqit.L2` subagent from definition files, per this repo's dogfooding convention.
- Added `smaqit.sync` as a new interactive drift-resolution skill after the user identified that the CLI `sync` alone could only report drift, never actually resolve it.
- Authored a PlantUML sequence diagram (`docs/designs/new-principle-cascade.puml`) documenting the cascade mechanism, on request.
- Built a new Bench scenario (`.smaqit/bench/scenarios/registry-cascade/`) to live-verify the cascade end-to-end via `smaqit.bench-scaffold`, then ran it live against the authenticated Codex CLI.
- Diagnosed the live run's reported failure as two genuine `src/bench` engine bugs, not a mechanism problem — confirmed via trace-log inspection that the cascade itself worked correctly. Manually verified and cleaned up a leaked workspace, then filed Task 035.
- Completed Task 032 via `smaqit.task-complete`: computed v2.1.0, opened PR #1, pushed and promoted the CHANGELOG entry.
- Discovered mid-task that the repo had been transferred to `daisy-tribe/smaqit-adk` (private) by a concurrently running Task 034 session; the PR merged but its release workflow never fired because Actions were disabled at merge time. Fixed by pushing the `adk-v2.1.0` tag directly.
- Diagnosed a second release blocker: the org's Actions/Packages storage quota (500 MB) was exhausted — not by `smaqit-adk` itself, but by an unrelated repo in the same org, `harness-deploy`, holding three already-expired ~1.6 GB Docker-image artifacts. Deleted them, unblocking the release.
- Verified the `adk-v2.1.0` GitHub Release published with all 5 platform binaries; ran Phase 2 of task completion (merge verification, worktree/branch cleanup).

## Problems Solved

- **Cascade orchestration design gap:** the task's original plan assumed L0/L1 could chain directly into L2 via a `Task` tool. Direct source inspection confirmed no 2-hop subagent chain exists anywhere in this repo or the sibling `smaqit` product — resolved by having the skill layer (`smaqit.new-principle`/`new-template`/`new-rules`) own the orchestration itself.
- **Missing interactive reconciliation path:** the CLI `sync` can only report drift (a plain binary has no path to invoke an LLM for the semantic recompile). Resolved by adding `smaqit.sync`, which shells out to the same registry CLI but uses its own `Task` access to actually resolve conflicts.
- **Two real `src/bench` engine bugs**, found live: the post-harness "freeze" step crashes on a non-regular file (Codex CLI's own IPC socket, created as a side effect of sandboxing `HOME`) instead of handling it gracefully; and a resulting cleanup failure is silently swallowed when an earlier failure already exists, which had leaked a workspace containing copied Codex credentials to disk with zero visible warning. Manually verified cleanup; filed as Task 035 rather than fixed inline.
- **Release pipeline broke twice, for reasons outside Task 032's own scope:** GitHub Actions were disabled on the freshly transferred private repo at the exact moment PR #1 merged, silently dropping the release trigger (fixed via a direct tag push, the workflow's own documented fallback); then the tag-push run failed on the org's Actions storage quota, traced to an unrelated repo's leftover expired artifacts (fixed by deleting them).

## Decisions Made

- Cascade orchestration lives in the skill layer, never via agent-to-agent `Task` chaining — matches `smaqit.development`'s proven single-hop pattern.
- Registry I/O centralized in a shared Go package (`installer/registry.go`), not per-agent shell/JSON manipulation, to avoid hash divergence.
- `smaqit-adk sync` (CLI) stays report-only by design; `smaqit.sync` (skill) is the real interactive resolution path.
- `smaqit.new-principle`'s cascade sequencing is L0 → L1 judgment → registry lookup → L2, not L0 → registry lookup directly — the registry only has L1-template→L2-output edges, never an L0-principle→L2 edge.
- The new Bench scenario is placed under `.smaqit/bench/scenarios/` (Task 031's anticipated convention) rather than `skills/`/`agents/`, since it tests a cross-cutting mechanism, not one artifact in isolation.
- Two acceptance criteria (extension-scope duplicate rejection, cross-project cascade isolation) were checked off on implementation correctness and code-level reasoning rather than an independent live trial — documented explicitly in Findings rather than silently overclaimed.
- Release-blocker fixes favored the simplest correct mechanism already built into the system: a direct tag push over fabricating a no-op PR; deleting only the other repo's already-expired artifacts, leaving its small non-expired ones untouched.

## Blockers Encountered

- A manual `--install-global` verification pass, run without sandboxing `HOME`, briefly overwrote the real machine-wide install (`~/.claude`, `~/.codex`, `~/.agents`). Caught immediately, restored from the production binary, orphaned WIP-only skill directories removed by hand; a feedback memory was saved to prevent recurrence.
- Repeated transient `git push` 403s, later understood to be side effects of the org-transfer's token/permission churn in progress on a different session.

## Files Modified

- `installer/registry.go`, `installer/registry_cli.go` — new registry engine and CLI subcommands
- `installer/main.go`, `installer/Makefile` — subcommand dispatch, skill packaging
- `agents/smaqit.L0.md`, `smaqit.L1.md`, `smaqit.L2.md` — scope detection, definition-file pattern, registration
- `skills/smaqit.new-principle/SKILL.md` (revised), `smaqit.new-template/`, `smaqit.new-rules/`, `smaqit.sync/` (new) — all compiled via `smaqit.L2`
- `.smaqit/definitions/skills/*.md` — 4 new/revised skill definitions
- `.smaqit/bench/scenarios/registry-cascade/` — new live-verification Bench scenario and fixture
- `docs/designs/new-principle-cascade.puml` — cascade sequence diagram
- `docs/wiki/extending-smaqit.md`, `AGENTS.md`, `README.md` — propagation model and skill catalog documentation
- `tests/unit/global_install_test.go` — updated for the 8-skill install set
- `CHANGELOG.md` — v2.1.0 entry
- `.smaqit/tasks/032_registry_driven_principle_propagation.md`, `035_bench_freeze_and_cleanup_error_handling.md` (new), `PLANNING.md`

## Next Steps

- Pick up Task 035 (Bench freeze-step and cleanup-error-handling fixes) when ready.
- Consider folding or superseding the pre-existing dogfood Bench case at `.smaqit/bench/skills/smaqit.new-principle/` — it only exercises Step 1 (L0) in isolation and doesn't cover the cascade this task added — with the new `scenarios/registry-cascade/` manifest.
- A dedicated live trial for extension-scope duplicate-principle rejection and cross-project cascade isolation, if stronger proof than code-level reasoning is wanted later.
- Task 034 (repo migration to `daisy-tribe/smaqit-adk`) remains in progress in its own worktree/session — not started by this one; `install.sh`'s redesign half of that task was not observed as done.
- Tasks 031 and 025 remain Not Started, untouched this session.

## Session Metrics

- **Tasks completed:** 1 (032, released as v2.1.0); 1 filed (035)
- **Skills compiled via real `smaqit.L2` subagent invocations:** 4 (`new-template`, `new-rules`, `sync`, revised `new-principle`)
- **New Go package + CLI surface:** `installer/registry.go` + hidden `registry get/put` and public `sync` subcommands
- **Live verification:** 1 new Bench scenario authored and run against the authenticated Codex CLI; cascade confirmed correct via direct trace-log inspection, independent of Bench's own pass/fail reporting
- **Engine bugs found:** 2, in `src/bench` (freeze-step crash on non-regular files; silently swallowed cleanup failure)
- **Safety incidents:** 1 (accidental real-install overwrite, caught and restored same session)
- **Release blockers diagnosed and resolved outside the task's own scope:** 2 (Actions disabled post-migration; org storage quota exhausted by an unrelated repo)
