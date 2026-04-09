# Task 017 Close and Release v0.5.0

## Metadata

- **Date:** 2026-04-05
- **Session focus:** Close out Task 017 (CLI tier subcommands), plan eval tasks 020/021, release adk-v0.5.0
- **Tasks completed:** 017 (CLI Init Tier Flags)
- **Tasks created:** 020 (Lite-Tier Behavioral Evals), 021 (Advanced-Tier Behavioral Evals)
- **Tasks referenced:** 018 (Level Skills Completion), 019 (Cross-Level Compilation)

## Actions Taken

### Task 020 and 021 created
- **020** — Lite-Tier Behavioral Evals: new eval files for `smaqit.create-agent` and `smaqit.create-skill` (skill + agent), unblocked
- **021** — Advanced-Tier Behavioral Evals: review/extend existing 7 evals + add evals for new skills from Task 018; blocked on Task 018

### Release adk-v0.5.0
- Ran release workflow: analysis → approval → prepare files → git operations
- Severity assessed as MINOR: new `lite` and `advanced` CLI subcommands added; `init` deprecated (not removed)
- CHANGELOG.md updated: new `[0.5.0]` entry with Added/Changed sections; footer links corrected (were stale at `adk-v0.3.1`)
- `installer/main.go` version constant bumped from `0.3.2` → `0.5.0`
- Three pre-release commits created before release commit:
  1. `fix: replace stale init references with lite/advanced tier commands` — CI workflow, README, wiki, skill error tables
  2. `chore: update task planning — add tasks 018-021, close task 017` — task files + history
  3. `Release adk-v0.5.0` — CHANGELOG.md + main.go
- Annotated tag `adk-v0.5.0` created
- Commit and tag pushed to `origin/main`

## Problems Solved

- **CHANGELOG footer was stale**: links stopped at `adk-v0.3.1` and skipped `0.3.2` and `0.4.0` — corrected in the same prepare step
- **`installer/main.go` version hardcode out of date**: was `0.3.2` despite being at tag `0.4.0` — bumped to `0.5.0`; ldflags override at build time anyway
- **SSH passphrase blocked first push attempt**: first `git push` exited with 130; user unlocked key and retry succeeded

## Decisions Made

- **Eval tasks numbered 020/021** — user confirmed Task 018 and 019 were already taken by parallel session; eval tracks recalibrated accordingly
- **020 is unblocked; 021 is blocked on Task 018** — lite artifacts are stable; advanced artifacts are not yet complete
- **Three commits before release commit** — logical grouping: stale-ref fix, planning chore, then release; keeps history readable

## Files Modified

| File | Change |
|---|---|
| `CHANGELOG.md` | Added `[0.5.0]` entry; fixed footer links |
| `installer/main.go` | Version constant `0.3.2` → `0.5.0` |
| `.smaqit/tasks/020_lite_tier_behavioral_evals.md` | Created |
| `.smaqit/tasks/021_advanced_tier_behavioral_evals.md` | Created |
| `.github/workflows/test-integration.yml` | `init` → `lite`/`advanced` (committed this session) |
| `README.md` | `smaqit-adk init` → `smaqit-adk lite` (committed this session) |
| `docs/wiki/extending-smaqit.md` | Structure section corrected (committed this session) |
| `skills/smaqit.create-agent/SKILL.md` | Error table reference updated (committed this session) |
| `skills/smaqit.create-skill/SKILL.md` | Error table reference updated (committed this session) |

## Next Steps

- **Task 018** — Level Skills Completion: `smaqit.new-principle`, `smaqit.new-template`, `smaqit.new-rules`; update `smaqit.L0`; expand `smaqit-adk advanced` installer
- **Task 019** — Cross-Level Compilation (`smaqit.compile.*`); depends on Task 018
- **Task 020** — Lite-Tier Behavioral Evals; unblocked; ready to start
- **Task 021** — Advanced-Tier Behavioral Evals; blocked on Task 018

## Session Metrics

- Duration: ~1 session
- Tasks completed: 1 (Task 017)
- Tasks created: 2 (020, 021)
- Files committed: 9 (across 3 pre-release commits + 1 release commit)
- Release: `adk-v0.5.0` tagged and pushed
