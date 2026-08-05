# Repair Broken Eval Artifact References

**Status:** Not Started
**Created:** 2026-08-05

## Description

Five of the seven behavioral eval files point at artifact files that **no longer exist**. The eval runner reads `artifact_file` relative to the repo root and injects its contents as the session system message; when the path is missing, `runEval` returns early with a `read artifact` error ([tests/evals/runner/main.go:190-195](../../tests/evals/runner/main.go#L190-L195)), so these evals cannot execute at all.

The cause is a rename: `smaqit.new-agent` and `smaqit.new-skill` became `smaqit.create-agent` and `smaqit.create-skill`, but the eval files were never repointed.

Verified state as of 2026-08-05:

| Eval directory | `artifact_file` | Exists? |
|---|---|---|
| `tests/evals/skills/smaqit.new-agent/` (3 evals) | `skills/smaqit.new-agent/SKILL.md` | **No** |
| `tests/evals/skills/smaqit.new-skill/` (2 evals) | `skills/smaqit.new-skill/SKILL.md` | **No** |
| `tests/evals/agents/smaqit.L2/` (2 evals) | `agents/smaqit.L2.agent.md` | Yes |

[Task 021](021_advanced_tier_behavioral_evals.md#L47) anticipates these evals being "stale" and needing review, but understates the situation — they are **broken**, not merely outdated. This task is the narrow repair; Task 021 remains the broader review-and-extend effort.

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

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] All 7 eval files reference an `artifact_file` that exists on disk
- [ ] Eval directory names match the artifacts under test
- [ ] `make -C installer evals` runs all 7 without any `read artifact` load error
- [ ] Each repointed eval's criteria have been re-read against the current `SKILL.md` and corrected where the inference-first rework invalidated them
- [ ] Pass/fail tally and run report path recorded in Findings (behavioral failures are acceptable and reported honestly, not tuned away)

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
| `tests/evals/skills/smaqit.new-agent/*.json` (3 files) | Modify + move to `smaqit.create-agent/` |
| `tests/evals/skills/smaqit.new-skill/*.json` (2 files) | Modify + move to `smaqit.create-skill/` |
| `.smaqit/tasks/021_advanced_tier_behavioral_evals.md` | Modify — note the repair once done |

## Notes

Running the evals requires an OAuth token; `make evals` auto-detects one via `gh auth token`. Classic PATs are rejected by the Copilot agent API. An explicit token is mandatory — without one the Copilot CLI reads shared XDG config and loads smaqit-adk's own VS Code workspace context, which invalidates results ([tests/evals/README.md:31-39](../../tests/evals/README.md#L31-L39)).

The last recorded eval outcome was 1/7 passing (2026-03-29, per Task 010's completion note), consistent with 5 of 7 having been unable to load.
