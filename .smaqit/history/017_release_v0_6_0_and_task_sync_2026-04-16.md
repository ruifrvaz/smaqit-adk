# Release v0.6.0 and Task Sync

## Metadata

- **Date:** 2026-04-16
- **Focus:** Test fix, release execution (adk-v0.6.0), post-release task state sync
- **Tasks referenced:** Task 018 (Level Skills Completion)

## Actions Taken

### 1. Test fix — structural skill heading

Two skill files (`smaqit.create-agent/SKILL.md`, `smaqit.create-skill/SKILL.md`) had been reverted by the user between sessions. Resuming the session revealed `TestSkillRequiredSections` failing for both: section heading was `## Completion Criteria` but the structural test requires exactly `## Completion`. Renamed in both files. Re-ran tests with `go clean -testcache` to bypass cache; all passing (unit + structural).

### 2. Release — adk-v0.6.0

Full local release workflow executed:

- **Analysis:** Last tag was `adk-v0.5.0`. Working tree had 16 changed files + 2 untracked (`docs/wiki/agent-frontmatter.md`, `skills/smaqit.new-principle/`). Key changes: new inference-first `create-*` skills, `smaqit.new-principle`, L2+templates in lite tier, deleted old agents/skills, installer rewrite, go.mod cleanup.
- **Severity:** MINOR (pre-1.0; removals increment Y). Suggested `adk-v0.6.0`. User approved.
- **CHANGELOG.md:** `[Unreleased]` → `[0.6.0] - 2026-04-16` with Added/Changed/Removed sections. Footer links updated.
- **Pre-release commit:** `feat: lite tier ships L2+templates+skills; new-principle added; create-* skills inference-first` — sha `27fea14`, 18 files, 478 insertions / 1146 deletions
- **Release commit:** `Release adk-v0.6.0` — sha `2635372`, CHANGELOG.md only
- **Tag:** `git tag -a adk-v0.6.0 -m "Release adk-v0.6.0"` (annotated)
- **Push:** `git push origin main && git push origin adk-v0.6.0` — exit 0; both pushed successfully

### 3. Task recap

User requested recap of active tasks. PLANNING.md read; 4 active tasks returned (018, 019, 020, 021) with blocker state. Noted Task 018 had a head start — `smaqit.new-principle` already shipped in v0.6.0.

### 4. Task 018 planning sync

Updated PLANNING.md and `018_level_skills_completion.md` to reflect v0.6.0 outcomes:
- Current state table: L0 `new-principle` ✅; L2 shows inference-first `create-agent`/`create-skill` (replacing deleted `new-agent`/`new-skill`)
- Deliverables: `new-principle` removed (done); `new-template` and `new-rules` remain
- Acceptance criteria: `new-principle` checked off
- Dependencies updated to reference adk-v0.6.0 instead of Task 009
- Task 006 absorption confirmed complete

## Problems Solved

- **Structural test cache masking failures:** `make test` returned `(cached)` after initial fix attempt; `go clean -testcache` was required to expose the real failure from the reverted files.
- **Skill heading regression:** reverted files had `## Completion Criteria`; test requires `## Completion`. Fixed and verified.

## Decisions Made

- `smaqit.new-principle` ships in advanced tier (not lite), installed by `smaqit-adk advanced` — confirmed in v0.6.0
- Task 018 remains "Not Started" despite `new-principle` being done — two deliverables plus L0 definition file pattern still unbuilt
- Advanced installer now targets three Level skills total: `new-principle` (done), `new-template`, `new-rules`
- Task 006 fully absorbed by v0.6.0; no separate close-out needed

## Files Modified

| File | Change |
|------|--------|
| `skills/smaqit.create-agent/SKILL.md` | `## Completion Criteria` → `## Completion` (structural test fix) |
| `skills/smaqit.create-skill/SKILL.md` | Same heading fix |
| `CHANGELOG.md` | `[0.6.0] - 2026-04-16` section added; `[Unreleased]` cleared; footer links updated |
| `docs/wiki/agent-frontmatter.md` | New file — committed in pre-release commit |
| `skills/smaqit.new-principle/SKILL.md` | New file — committed in pre-release commit |
| `agents/smaqit.L2.agent.md` | Template paths updated to `.smaqit/templates/`; attribution changed |
| `installer/main.go` | Full rewrite — SDK removed, `installLiteComponents()` helper, cmdCreate deleted |
| `installer/go.mod` | Copilot SDK dependency removed |
| `tests/unit/lite_test.go` | Updated to reflect new architecture |
| `tests/unit/embed_test.go` | `expectedFiles` updated to `smaqit.L2.agent.md` |
| `.smaqit/tasks/PLANNING.md` | Task 018 note updated post-v0.6.0 |
| `.smaqit/tasks/018_level_skills_completion.md` | Full current state refresh |

## Next Steps

- **Task 018 (Level Skills Completion):** Remaining — `smaqit.new-template`, `smaqit.new-rules`, L0 definition file pattern, installer update for two new skills
- **Task 020 (Lite-Tier Behavioral Evals):** Unblocked and independent — ready to start
- Tasks 019 and 021 blocked on Task 018 completion

## Session Metrics

- Duration: Full session
- Tasks completed: 0 active tasks formally closed (v0.6.0 delivered the work; 018 partially satisfied)
- Files modified: ~14 (including pre-release commit files)
- Key quantitative outcomes: adk-v0.6.0 released; 18 files changed, 478+/1146- in pre-release commit; all tests passing
