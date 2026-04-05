# Task Roadmap Restructure

**Date:** 2026-04-05  
**Focus:** Reassessing and restructuring the active task roadmap around the new `smaqit-adk advanced` execution scope  
**Tasks completed:** none (planning session)  
**Tasks referenced:** 006, 013, 014, 015, 018 (new), 019 (new), 020, 021

## Actions Taken

- Loaded session context: README, PLANNING.md, history 014
- Reframed Task 015 scope: original CLI session-chaining approach replaced by VS Code-native skill approach
- Deferred Task 015 (Full Compilation Chain CLI) — VS Code-native `smaqit.compile` skills supersede it on the VS Code surface; CLI chain preserved in Future for CI/CD use cases
- Closed Task 006 (smaqit.new-principle Skill) as Abandoned — absorbed into new Task 018
- Created Task 018 (Level Skills Completion): authors `smaqit.new-principle`, `smaqit.new-template`, `smaqit.new-rules`; updates L0 definition file input pattern; updates installer; absorbs Task 006
- Created Task 019 (Cross-Level Compilation): designs `smaqit.compile.principle`, `smaqit.compile.template`, `smaqit.compile.agent` skills; runs L0→L1→L2 chain via sequential subagent invocations; depends on Task 018
- Registered Tasks 020 and 021 (created in a parallel session) into PLANNING.md Active table
- Fixed PLANNING.md structural issue: duplicate `## Completed` headers collapsed into one; Task 017 was lost in the collapse and restored
- Confirmed separate skills for L1: `smaqit.new-template` (template structure) and `smaqit.new-rules` (compilation rules) — same narrow scope as L0/L2 pairs

## Decisions Made

- **Task 015 deferred, not deleted**: the CLI chain has a legitimate CI/CD use case for the future; parking rather than discarding
- **Task 006 absorbed, not merged**: 006's acceptance criteria are fully included in 018 — no standalone task needed
- **`smaqit.compile` is a distinct task (019)**: cross-level compilation is a separate concern from authoring individual Level artifacts; 019 depends on 018 being complete first
- **Separate L1 skills confirmed**: `smaqit.new-template` and `smaqit.new-rules` are separate skills — mirrors the one-skill-per-artifact-type pattern of L0 and L2; keeps scope tight
- **`smaqit.compile.agent` has no 018 dependency**: it only needs L2, which already has a definition file pattern — noted in 019 but not blocking anything

## Problems Solved

- PLANNING.md had two `## Completed` blocks (017 in its own block, then the rest); collapsed and Task 017 was missing after the fix — caught and restored
- Tasks 020 and 021 were created in a separate session and not registered in PLANNING.md — added to Active table

## Files Modified

- `.smaqit/tasks/015_full_compilation_chain_cli.md` — Status: Deferred, reason added
- `.smaqit/tasks/006_create_new_principle_skill.md` — Status: Abandoned, reason added
- `.smaqit/tasks/PLANNING.md` — Active table updated (removed 006/013/015, added 018/019/020/021); duplicate Completed header fixed; Task 017 restored

## Files Created

- `.smaqit/tasks/018_level_skills_completion.md` — Level Skills Completion task
- `.smaqit/tasks/019_cross_level_compilation.md` — Cross-Level Compilation task

## Next Steps

- **Task 018 (Level Skills Completion)** — recommended next pick; fully unblocked; foundational for 019 and 021
  - First sub-step: update `smaqit.L0` with definition file input pattern
  - Then: author `smaqit.new-principle`, `smaqit.new-template`, `smaqit.new-rules`
  - Finally: update `smaqit-adk advanced` installer to ship all five Level skills
- Task 020 (Lite-Tier Behavioral Evals) and Task 014 (CLI fix) are parallel tracks — can be picked up any time independently

## Session Metrics

- Duration: medium  
- Tasks completed: 0 (planning session)  
- Task files created: 2 (018, 019)  
- Task files modified: 3 (006, 015, PLANNING.md)  
- History file: this file
