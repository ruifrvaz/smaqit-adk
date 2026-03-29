# Task Planning

## Active

| ID | Title | Status | Notes |
|----|-------|--------|-------|
| 009 | Create smaqit.new-skill Skill | Completed | All criteria met; architectural correction: skill compilation moved from L1 → L2; reference chain constraint clarified |
| 006 | Create smaqit.new-principle Skill | Not Started | Depends on Task 009 — should be created using smaqit.new-skill |

## Completed

| ID | Title | Completed | Notes |
|----|-------|-----------|-------|
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
| 012 | Lite Tier — Compiled Standalone Agents | Compile smaqit.create-agent + smaqit.create-skill via L2; repurpose init to drop only these two files; no boilerplate |
| 011 | Interactive CLI Product (Advanced Tier) | Full Go CLI developer suite: create-agent, create-skill, create-principle, validate; Copilot SDK (go); Phase 0 resolved — SDK confirmed viable; init fate resolved by Task 012 |
| 010 | Test Framework | Three-layer: embed bug fix + Go unit tests + structural validation + Copilot SDK eval runner (isolated workspace, scripted turns, LLM grading); Phases 0–2 unblocked; Phase 3 uses same Copilot SDK as Task 011 |
