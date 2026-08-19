# Bench `Case.prepare` Has No Lifecycle Hook to Clean Up a Backgrounded Process

**Status:** Not Started
**Created:** 2026-08-19

## Description

Found while designing a Bench manifest in a downstream consuming project (smaqit's task 110 — Vault Loader slug derivation and non-interactive secret-write safety) that needed a live, ephemeral `vault server -dev` running for the duration of a Case (through the variant's process execution and into `expect` checks), not just during `prepare` itself.

Read `src/bench/run.go` and `src/bench/process.go`/`process_unix.go` directly rather than guessing, since no existing manifest anywhere (this repo's own dogfood suite included) launches a background/daemon process from `Case.prepare`. Confirmed:

- `runAttempt` (`run.go`) runs every `caseConfig.Prepare[i]` command sequentially via `executeCommand`, waiting for each to exit before proceeding — this part is fine and matches the README's documented contract.
- `processAdapter.Execute` (`process.go:20-81`) sets `Setpgid: true` on the launched command (`configureProcessGroup`) and only calls `terminateProcessTree` (which `SIGKILL`s the whole process group via `syscall.Kill(-pid, SIGKILL)`) in the `ctx.Done()` branch — i.e. **only on timeout or cancellation**. On normal successful completion (`case waitErr := <-done`), no process-group cleanup happens at all.
- A prepare command that backgrounds a child (e.g. `sh -c "vault server -dev ... & "`) therefore returns exit 0 immediately, the prepare step is marked done, and the backgrounded process keeps running — sharing the same process group, but with no `Wait()` blocking on it and no cleanup path once the run finishes successfully.
- Workspace cleanup (`defer removeWorkspace(...)` in `runAttempt`) is filesystem-only (`os.RemoveAll`-shaped), not process cleanup — it does not touch the process group.

**Net effect:** a Case author who backgrounds a long-lived process in `prepare` (a dev server, a mock service, anything a Case might legitimately need alive across its own variant execution and `expect` checks) gets an orphaned process on the host after every successful run, with no engine-provided way to guarantee it is torn down — not in `prepare` itself (nothing runs *after* success there), not in `expect` (meant to be re-gradable without rerunning the harness, per `bench grade`, so teardown there would be semantically wrong and skipped on regrade anyway), and not via any documented Case-level `teardown`/`cleanup` field (none exists in the schema).

This is a real gap, not a documentation gap: `.smaqit/bench/README.md`'s "Case data planes" section documents `fixture`/`prepare`/`given`/`treatment` but has no teardown story at all.

## Issue Triage Context

**Mode:** Skip
**Technologies:** None
**Platforms/Environments:** None
**Features/Integrations:** None
**Versions/Constraints:** None

## Design Decisions

TBD — open questions for whoever picks this up:

- Add an explicit `Case.teardown` (or `Case.cleanup`) command list, run unconditionally after the run's outcome is determined (success, failure, or timeout alike) — mirroring `prepare`'s shape but guaranteed to execute regardless of variant outcome, and itself bounded by a timeout so a hanging teardown command can't block the run indefinitely.
- Alternatively (cheaper, narrower): have the engine track PIDs/process-groups spawned during `prepare` and forcibly terminate them (SIGTERM then SIGKILL after a grace period) once the run's own workspace cleanup fires, regardless of exit status — this would require no new manifest schema field, just extending `runAttempt`'s existing cleanup path to also reap prepare-spawned process groups.
- Either way: document the guarantee (or lack of one) explicitly in `.smaqit/bench/README.md`'s "Case data planes" section once resolved, so a future Case author doesn't have to re-derive this from source the way this task did.

## Implementation Steps

TBD — sketch, not committed:

1. Decide between an explicit `teardown` schema field vs. engine-tracked automatic process-group reaping (see Design Decisions).
2. Implement the chosen mechanism in `src/bench/run.go`/`process.go`, ensuring it fires on every exit path (success, failure, timeout, cancellation) — not just the `ctx.Done()` branch that already exists.
3. Add a regression test proving a backgrounded process spawned in `prepare` is confirmed gone after the run completes (e.g. assert the PID no longer exists, or the bound port is free).
4. Document the guarantee in `.smaqit/bench/README.md`.
5. Consider adding a minimal example manifest under `examples/bench/` demonstrating a backgrounded-service Case, once the mechanism exists — useful as a template for the next Case author who needs this (e.g. smaqit's task 110, once this ships).

## Known Issues Triage

**Triaged:** 2026-08-19
**Tools searched:** none
**Result:** Clear — internal engine gap in `src/bench`, not a third-party dependency issue.

## Acceptance Criteria

- [ ] A backgrounded process launched in a Case's `prepare` step is guaranteed torn down after the run completes, on every exit path (success, failure, timeout, cancellation) — not just on timeout/cancellation as today
- [ ] The mechanism requires no manual `pkill`/cleanup step from the Case author or a human running the suite
- [ ] `.smaqit/bench/README.md`'s "Case data planes" section documents the guarantee
- [ ] A regression test proves the backgrounded process is actually gone post-run, not just that the run reported success

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
| `src/bench/run.go` | Modify — cleanup path for prepare-spawned processes on every exit outcome |
| `src/bench/process.go` / `process_unix.go` / `process_windows.go` | Modify — teardown/reaping mechanism |
| `src/bench/manifest.go` | Modify, if a new `teardown` schema field is chosen |
| `.smaqit/bench/README.md` | Modify — document the guarantee |

## Notes

Discovered live while planning (not yet implementing — blocked on this gap) a Bench case for smaqit's task 110 that needed a real ephemeral `vault server -dev` alive across a Case's full lifecycle to prove a non-interactive secret-write bug fix end-to-end. Worked around downstream by scoping that Bench case to a Vault-free scenario instead (slug derivation only) and relying on a static-analysis check plus manual live verification for the Vault-dependent half — not blocked on this task, but would directly benefit from it. Filed from the smaqit repo's own session (not raised by direct smaqit-adk maintenance work), so cross-check `git log`/`PLANNING.md` here before treating this as independently prioritized against 031/032.
