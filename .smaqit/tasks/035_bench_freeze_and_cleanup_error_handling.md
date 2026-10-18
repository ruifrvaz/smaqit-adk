---
status: Abandoned
created: "2026-08-23"
superseded_by: "037"
---

# Bench Freeze Step Crashes on Non-Regular Files, and Cleanup Failures Are Silently Swallowed

## Description

Found live while running Task 032's registry-cascade Bench scenario (`.smaqit/bench/scenarios/registry-cascade/`), which deliberately runs `codex exec` with a workspace-local `HOME` to sandbox the registry-driven cascade mechanism under test. Two real `src/bench` engine gaps surfaced together in the same live run:

1. **The post-harness "freeze" step crashes on non-regular files instead of handling them.** Running `codex exec` with `HOME={workspace}` causes Codex CLI to create its own runtime IPC socket at `$HOME/.codex/ipc/ipc.sock` as a normal side effect of its own execution — nothing to do with anything the Case manifest staged. When `runAttempt` (`src/bench/run.go`) calls `snapshotTree(workspace.Root)` afterward to freeze the final tree state, the walk tries a raw `open()` on this socket and fails with `open .../.codex/ipc/ipc.sock: no such device or address` (ENXIO), aborting the whole attempt with `phase: "freeze"` before any `expect` check ever runs. This happened even though the harness itself completed successfully (`exitCode: 0`) and — confirmed by manually reading the trace logs — had already done everything the Case was testing for correctly (added a principle, updated compiled rules, recompiled a registered dependent, updated the registry). `src/bench/hash.go`'s `digestPathExcluding` (used for plan-time asset hashing) already has the right pattern to mirror: it explicitly checks `!info.Mode().IsRegular()` and returns a clear error rather than blindly calling `open()`. `snapshotTree`/`compareTrees` (`src/bench/inventory.go`) has no equivalent guard.

2. **Worse: when the freeze failure above occurs, the deferred workspace cleanup's own failure is silently swallowed — with zero visible warning that a workspace leaked.** `runAttempt`'s deferred cleanup (`src/bench/run.go`) reads: `if cleanupErr := removeWorkspace(workspace.Root); cleanupErr != nil && result.Failure == nil { result = finishFailure(result, "cleanup", cleanupErr) }`. The `&& result.Failure == nil` guard means a cleanup failure is only ever reported when no other failure was already recorded. Since the freeze failure above had already set `result.Failure`, `removeWorkspace`'s own failure (confirmed live: it also failed, on the same socket file and/or the write-protected `.smaqit-bench-input/` treatment sidecar) vanished with no trace in `result.json`, `report.md`, or any CLI output — the run just reported the freeze failure and said nothing about cleanup at all. The temp workspace was left sitting on disk under `/tmp/smaqit-bench-*`. In this specific manifest's case that workspace held a locally-copied set of real Codex CLI credentials (staged deliberately, with the user's explicit sign-off, to authenticate the sandboxed `codex exec` call) — found only by manually `ls`-ing `/tmp` after the run, and cleaned up by hand.

Both are genuine `src/bench` correctness gaps, not documentation gaps or manifest-authoring mistakes — confirmed by reading the actual engine source rather than guessing.

## Issue Triage Context

**Mode:** Skip
**Technologies:** None
**Platforms/Environments:** None
**Features/Integrations:** None
**Versions/Constraints:** None

## Design Decisions

TBD — open questions for whoever picks this up:

- For the freeze crash: either skip non-regular files during `snapshotTree` (recording them in the tree listing with a `"type": "special"` marker and no content hash) or fail with a clear, actionable message naming the offending path — either is better than the current raw OS `errno` leaking through as an opaque `phase: "freeze"` failure. Mirror `hash.go`'s existing `IsRegular()` check rather than inventing a new pattern.
- For the swallowed cleanup error: a cleanup failure is an independent fact from whatever failure preceded it — conflating or dropping one because the other already exists is the actual bug. Remove the `&& result.Failure == nil` guard and always surface a cleanup failure distinctly (e.g. a separate `result.CleanupFailure` field, or an explicit warning line in `report.md`/CLI output) regardless of whether an earlier failure was already recorded. This is arguably the more severe of the two gaps: it defeats the "ephemeral, always-cleaned-up workspace" safety premise the whole engine design relies on, silently.

## Implementation Steps

TBD — sketch, not committed:

1. Add an `IsRegular()` guard to `snapshotTree`'s per-file handling in `src/bench/inventory.go`, mirroring `hash.go`'s `digestPathExcluding` pattern. Decide skip-with-marker vs. error-with-clear-message (see Design Decisions).
2. Fix the cleanup-error-swallowing bug in `runAttempt`'s deferred cleanup (`src/bench/run.go`) — always report a cleanup failure distinctly, never conditionally drop it because an earlier failure already exists.
3. Add a regression test: run a harness that deliberately leaves a non-regular file (e.g. a Unix socket created via a small test helper, or `mkfifo`) in the workspace; confirm the run reports a clear diagnostic instead of a raw `ENXIO`, and confirm a cleanup failure (if forced) is always visibly reported rather than silently dropped when another failure already exists.
4. Document both behaviors in `.smaqit/bench/README.md`'s "Case data planes" section — a harness may leave non-regular files behind, and a cleanup failure is always surfaced, never silent.

## Known Issues Triage

**Triaged:** 2026-08-23
**Tools searched:** none
**Result:** Clear — internal engine gap in `src/bench`, not a third-party dependency issue.

## Acceptance Criteria

- [ ] A harness-created non-regular file (socket, FIFO, device) in the workspace no longer crashes the freeze step with a raw OS error; it is handled with a clear, actionable message or skipped gracefully
- [ ] A workspace-cleanup failure is always surfaced (in `result.json`/`report.md`/CLI output), never silently dropped because an earlier failure was already recorded
- [ ] A regression test proves both behaviors
- [ ] `.smaqit/bench/README.md` documents both guarantees

## Findings

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
| `src/bench/inventory.go` | Modify — non-regular-file handling in `snapshotTree`/`compareTrees` |
| `src/bench/run.go` | Modify — fix the cleanup-error-swallowing guard in `runAttempt`'s deferred cleanup |
| `src/bench/hash.go` | Reference only — existing `IsRegular()` pattern to mirror, no change expected |
| `.smaqit/bench/README.md` | Modify — document both guarantees |

## Notes

Discovered live while running Task 032's `.smaqit/bench/scenarios/registry-cascade/bench.yaml` — a scenario proving the registry-driven principle-propagation cascade (`smaqit.new-principle` → `smaqit.L0` → `smaqit.L1` → registry lookup → `smaqit.L2` recompile + re-register) end-to-end via a real `codex exec` run. The cascade mechanism itself was confirmed fully correct by direct trace-log inspection (principle landed, rules updated, dependent recompiled, registry updated, all with the exact expected content) despite Bench reporting the run as failed — the failure was entirely these two infrastructure gaps, not the mechanism under test. Task 032 treats this live run as sufficient proof and does not block on this task.

The leaked-credential-copy workspace from the live run was found and removed by hand (verified via `ls`, confirmed the path matched Bench's own `smaqit-bench-*` temp-dir naming and was not `$HOME`, then removed — plus a `/tmp` sweep confirming no other stray Bench workspaces existed). No credential material was left on disk after manual cleanup.
