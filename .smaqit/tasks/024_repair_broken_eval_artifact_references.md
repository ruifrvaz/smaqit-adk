# Repair Broken Eval Artifact References

**Status:** Completed
**Created:** 2026-08-05
**Started:** 2026-08-05
**Completed:** 2026-08-06
**Mode:** Assisted

## Description

Five of the seven behavioral eval files point at artifact files that **no longer exist**. The eval runner reads `artifact_file` relative to the repo root and injects its contents as the session system message; when the path is missing, `runEval` returns early with a `read artifact` error ([tests/evals/runner/main.go:190-195](../../tests/evals/runner/main.go#L190-L195)), so these evals cannot execute at all.

The cause is a rename: `smaqit.new-agent` and `smaqit.new-skill` became `smaqit.create-agent` and `smaqit.create-skill`, but the eval files were never repointed.

Verified state as of 2026-08-05:

| Eval directory | `artifact_file` | Exists? |
|---|---|---|
| `tests/evals/skills/smaqit.new-agent/` (3 evals) | `skills/smaqit.new-agent/SKILL.md` | **No** |
| `tests/evals/skills/smaqit.new-skill/` (2 evals) | `skills/smaqit.new-skill/SKILL.md` | **No** |
| `tests/evals/agents/smaqit.L2/` (2 evals) | `agents/smaqit.L2.agent.md` | Yes |

[Task 021](021_advanced_tier_behavioral_evals.md#L47) anticipates these evals being "stale" and needing review, but understates the situation — they are **broken**, not merely outdated. This task is the narrow repair. The broader review-and-extend effort for these specific artifacts belongs to **Task 020** (Lite-Tier Behavioral Evals) — `create-agent`/`create-skill` are lite-tier artifacts, not advanced-tier, so Task 021 was the wrong cross-reference; corrected during implementation.

## Design Decisions

- **Repair before rewrite.** Repointing is a small mechanical change that restores a working baseline signal. Task 021's larger review should start from evals that at least load and run.
- **Do not assume the eval content is still correct.** The rename accompanied a behavioral change: `create-agent`/`create-skill` are inference-first (v2.0.0) whereas the `new-*` skills gathered specs turn by turn. Turns and criteria that simulate a full interactive gathering flow may no longer describe the artifact's actual behavior, so each eval needs its assertions re-read against the current `SKILL.md`, not just its path swapped.

## Implementation Steps

1. Read [skills/smaqit.create-agent/SKILL.md](../../skills/smaqit.create-agent/SKILL.md) and [skills/smaqit.create-skill/SKILL.md](../../skills/smaqit.create-skill/SKILL.md) in full to establish current behavior.
2. For each of the 5 broken eval files: repoint `artifact_file`, then re-read `turns`, `expected_behavior`, and `forbidden_behavior` against the current skill and correct anything the inference-first rework invalidated.
3. Rename the eval directories to `tests/evals/skills/smaqit.create-agent/` and `tests/evals/skills/smaqit.create-skill/` so directory names match the artifacts under test.
4. Run `make -C installer evals` and confirm every eval at minimum **loads and executes** (no `read artifact` errors). Genuine behavioral failures are a legitimate outcome — record them rather than tuning criteria until they pass.
5. Record the run report path and the pass/fail tally in Findings.

## Known Issues Triage
**Triaged:** 2026-08-05
**Tools searched:** Copilot SDK (Go) — `github/copilot-sdk`, GitHub CLI — `cli/cli`
**Result:** Advisory

### Advisory Issues
- [#1958 [.NET 1.0.6][FFI] CreateSessionAsync can hang after successful startup in Linux/Kubernetes](https://github.com/github/copilot-sdk/issues/1958) — `github/copilot-sdk` — opened 2026-07-09 — bug (matches platform "Linux" only; report is scoped to the .NET FFI binding, not the Go SDK the eval runner uses — noted defensively, not expected to reproduce here)

### Historical (Closed)
- [#2249 Fine-grained PAT authenticates on Windows but not Debian Linux](https://github.com/github/copilot-sdk/issues/2249) — `github/copilot-sdk` — closed 2026-08-04 — root-caused to a stripped Docker image missing `ca-certificates`, not an SDK defect; this dev machine already has working TLS (verified via `curl` this session), so not expected to apply

### Unresolvable Tools
- (none — both tools resolved)

## Acceptance Criteria

- [x] All 7 eval files reference an `artifact_file` that exists on disk
- [x] Eval directory names match the artifacts under test
- [x] `make -C installer evals` runs all 7 without any `read artifact` load error
- [x] Each repointed eval's criteria have been re-read against the current `SKILL.md` and corrected where the inference-first rework invalidated them
- [x] Pass/fail tally and run report path recorded in Findings (behavioral failures are acceptable and reported honestly, not tuned away)

## Findings

**Implementation approach:**
- Repointed and renamed the 5 broken eval files (`smaqit.new-agent/` → `smaqit.create-agent/`, `smaqit.new-skill/` → `smaqit.create-skill/`) via `git mv`, then rewrote every turn/criterion in all 5 against the current inference-first (single-question) `create-agent`/`create-skill` `SKILL.md` text — the old files simulated a 10–15 turn interactive gathering flow that no longer exists.
- Ran `make -C installer evals` repeatedly to validate. Along the way, root-caused and fixed two real bugs in the shared eval runner (`tests/evals/runner/main.go`) that were blocking any real tally: a process leak (`copilot.Client` never called `.Stop()`, leaking one `copilot --headless` OS process per session and per graded criterion — confirmed via the vendored SDK source's own documented `defer client.Stop()` pattern, and via 23 orphaned processes observed pre-fix) and a wrong template path (`setupWorkspace()` copied ADK templates to `dir/templates` instead of `dir/.smaqit/templates`, the path every current skill/agent actually reads — confirmed by inspecting the provisioned workspace on disk).
- Obtained a final, clean, uninterrupted run after both fixes: 2/7 passed.

**Decisions made:**
- Corrected the task's own cross-reference from Task 021 (advanced-tier) to Task 020 (lite-tier), which actually owns `create-agent`/`create-skill` eval coverage; noted in both task files.
- Fixed the two runner-infrastructure bugs in-session rather than deferring them, since they directly blocked this task's own "record a real pass/fail tally" criterion. Both fixes are minimal and evidence-backed, and validated by a behavioral flip: `L2/001_compile_base_agent` — the one eval that must actually read and merge template files — timed out in every pre-fix run and passed in every post-fix run.
- Did not chase a third, distinct issue (permission-blocked file writes, below) inside this task — it's a deeper, evidently pre-existing defect unrelated to the artifact_file repair or the two fixes already made. Recorded as follow-up per this task's own "behavioral failures are acceptable and reported honestly" criterion rather than expanding scope further.

**Blockers encountered:**
- Initial auth failures (`Authorization error, you may need to run /login`) — resolved when the user added "Copilot Requests" permission to the fine-grained PAT used for `GH_TOKEN`.
- A background run was killed by an unrelated Claude Code process/harness restart mid-run (no completion record); required a clean re-run.
- All 5 `create-agent`/`create-skill` evals fail even after the templates-path fix, every one citing the same cause: blocked by filesystem permission errors, unable to write any file. Root cause not identified — `.smaqit/definitions/`, `.github/agents/`, and `.github/skills/` are all pre-created with 0755 by `setupWorkspace()`, ruling out the obvious explanation. The two `L2` evals never actually verify a successful on-disk write (they only check chat-transcript content), so they don't exercise this path and can't confirm whether the defect is pre-existing or specific to the inference-first skills.

**Follow-up identified:**
- New task needed: investigate why `create-agent`/`create-skill` sessions cannot write any file in the eval workspace. This blocks Task 020 (Lite-Tier Behavioral Evals) from ever reaching a genuine pass, since any new eval Task 020 authors will hit the same workspace/permission setup.
- Final tally — **2/7 passed**: `agents/smaqit.L2/001_compile_base_agent` PASS, `agents/smaqit.L2/002_reject_unresolved_placeholders` PASS; all 5 `create-agent`/`create-skill` evals FAIL/ERROR on the permission-write issue above. Run report: `tests/evals/runs/20260806-113717/` (`results.json`, `report.md`).
- Task 020 already carries a note (added this session) pointing at the corrected evals and flagging the auth/leak/path gotchas found here; the new permission-write follow-up should be added there, or filed as its own task, before Task 020 starts.

## Files to Create / Modify

| File | Action |
|------|--------|
| `tests/evals/skills/smaqit.new-agent/*.json` (3 files) | Modify + move to `smaqit.create-agent/` |
| `tests/evals/skills/smaqit.new-skill/*.json` (2 files) | Modify + move to `smaqit.create-skill/` |
| `.smaqit/tasks/020_lite_tier_behavioral_evals.md` | Modify — note the repair once done (corrected from Task 021, which does not cover these lite-tier artifacts) |
| `tests/evals/runner/main.go` | Modify — added `defer client.Stop()` at both `NewClient()` sites (process leak); fixed `setupWorkspace()` to copy templates/framework under `.smaqit/` instead of top-level (wrong runtime path) |

## Notes

Running the evals requires an OAuth token; `make evals` auto-detects one via `gh auth token`. Classic PATs are rejected by the Copilot agent API. An explicit token is mandatory — without one the Copilot CLI reads shared XDG config and loads smaqit-adk's own VS Code workspace context, which invalidates results ([tests/evals/README.md:31-39](../../tests/evals/README.md#L31-L39)).

The last recorded eval outcome was 1/7 passing (2026-03-29, per Task 010's completion note), consistent with 5 of 7 having been unable to load.
