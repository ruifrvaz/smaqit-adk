# Task Planning

## Active

| ID | Title | Status | Notes |
|----|-------|--------|-------|
| 006 | Create smaqit.new-principle Skill | Not Started | Depends on Task 009 (done) — use smaqit.new-skill; unblocks Task 013 create-principle |

| 013 | CLI create-principle and validate Commands | Not Started | Deferred from Task 011; create-principle depends on Task 006; validate needs design decision |
| 014 | CLI create-agent / create-skill Fix | Not Started | Fix wrong agent context, remove timeout + ticker; local changes already made, need commit + release |
| 015 | Full Compilation Chain CLI (L0→L1→L2) | Not Started | New `compile` command or `--full` flag; three SDK sessions file-chained in Go; ADK stays in binary |

## Completed

| ID | Title | Completed | Notes |
|----|-------|-----------|-------|
| 017 | CLI Tier Subcommands — Replace `init` with `lite` and `advanced` | 2026-04-03 | All 7/7 criteria met; breaking change; version bump required at release |

## Completed

| ID | Title | Completed | Notes |
|----|-------|-----------|-------|
| 011 | Interactive CLI Product (Advanced Tier) | 2026-04-03 | create-agent + create-skill complete; create-principle + validate deferred to Task 013 |
| 016 | Lite Tier — Routing Skills for Natural Language Entry Points | 2026-04-03 | All 8/8 criteria met; user testing passed; natural language entry point working |
| 012 | Lite Tier — Compiled Standalone Agents | 2026-03-29 | smaqit.create-agent + smaqit.create-skill compiled via L2; init repurposed to drop only these two files; no boilerplate |
| 010 | Test Framework | 2026-03-29 | Three-layer test suite complete: embed bug fix, Go unit/structural tests, Copilot SDK eval runner; 1/7 evals passing on last run |
| 009 | Create smaqit.new-skill Skill | 2026-03-29 | All criteria met; architectural correction: skill compilation moved from L1 → L2; reference chain constraint clarified |
| 008 | Framework Philosophy Recalibration | 2026-03-01 | All 5 framework files rewritten to behavioral principles only; removed self-referencing and product content; SMAQIT.md made cross-cutting; templates/skills/ created; catalog content moved to copilot-instructions and wiki |
| 005 | Redesign Framework Files | 2026-03-01 | All 5 files redesigned: SMAQIT.md (cross-cutting principles), AGENTS.md (behavioral, voice-cleaned), SKILLS.md (principles only, catalog removed), TEMPLATES.md (Agent Templates section dropped), ARTIFACTS.md (minor cleanup) |
| 004 | Distill AGENTS-old into AGENTS.md | 2026-02-28 | Added 3 invariants/behaviors (assumption-flagging, blocker-stop, skill-mediated workflows); deleted AGENTS-old.md; synced installer |
| 007 | L0 Principle + Invariant + Vocabulary Layering | 2026-02-27 | Established three-layer content model; rewrote TEMPLATES.md as clean L0; moved placeholder catalogs to compiled/*.rules.md; reference pattern for future cleanups |
| 003 | Skill-First Invocation Model | 2026-02-27 | Skills are entry points; L2 invoked as subagent by skill; no orchestrator; AGENTS.md rewritten; L0/L1/L2 framing updated |
| 002 | Migrate Prompts to Skills | 2026-02-27 | Migrated to agentskills.io format; dropped input-record philosophy; all L0/L1/L2 agents, installer, README, wiki updated |
| 001 | Clean L2 Agent Contamination | 2026-02-27 | Removed smaQit product-specific content; generalized to domain-agnostic ADK model |

## Abandoned

| ID | Title | Date | Reason |
|----|-------|------|--------|

## Future

| ID | Title | Notes |
|----|-------|-------|
