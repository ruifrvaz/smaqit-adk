# Framework Behavioral Recalibration

**Date:** 2026-03-01  
**Session Focus:** Recalibrate all five framework files to describe behavioral principles of agentic constructs rather than cataloging what they are or how smaqit-adk works internally. Remove self-referencing product content throughout. Introduce `templates/skills/` structure for future skill compilation.  
**Tasks Completed:** Task 005 (Redesign Framework Files), Task 008 (retroactive — Framework Philosophy Recalibration)  
**Tasks Referenced:** Task 006 (Create smaqit.new-principle Skill — In Progress)

---

## Actions Taken

- Assessed all five framework files against the behavioral principle intent: files should describe how agents/skills/templates/artifacts must behave, not what they are or how the ADK is structured
- Identified two contamination types across files: (1) ADK self-referencing (product names, level names, file paths) and (2) catalog/meta content that belongs elsewhere
- Rewrote `SMAQIT.md` — replaced smaQit product content (5 layers, 3 phases, lifecycle states) with 5 cross-cutting system principles: Explicit Over Implicit, Single Source of Truth, Traceability, Composability, Validate Behavior Not Reproduction
- Rewrote `SKILLS.md` — kept 5 behavioral principles, removed all content below Core Principles (product catalog, format docs, loading mechanism, historical "Skills vs Input Records" rationale)
- Rewrote `TEMPLATES.md` — kept 5 structural principles, removed `## Agent Templates` section (file tree, section table, GitHub Custom Agent format note), fixed `Agent-L2` references to `the compiler`
- Voice-cleaned `AGENTS.md` — removed "in the ADK," "Level and product," "defined at compilation," "compilation chain"; changed `Compiled agents do not` to `Agents do not`
- Voice-cleaned `ARTIFACTS.md` — removed "in the ADK," changed "compilation chain" to "system"
- Moved catalog content to correct homes: Agent Catalog and Skill Catalog to `copilot-instructions.md`; Downstream Input Patterns to `docs/wiki/extending-smaqit.md`
- Created `templates/skills/compiled/skill.rules.md` — L1 vocabulary: skill format, loading stages, agent-skill interaction model, placeholder convention note
- Created `templates/skills/skill.template.md` — stub placeholder for future `smaqit.new-skill` compilation
- Updated installer Makefile to copy `templates/skills/` during `prepare`
- Updated PLANNING.md: closed Tasks 005 and 008, added Task 009 (`smaqit.new-skill`) to Future, kept Task 006 In Progress
- Updated individual task files 005 and 006 with correct statuses
- Identified and assessed gap in `smaqit.task-complete` skill: description does not trigger on "update task status" or retroactive closure scenarios

---

## Problems Solved

**Framework files were self-referencing** — files like `SMAQIT.md` and `AGENTS.md` described how smaqit-adk works internally (L0/L1/L2 levels, named agents, file paths, product names) rather than describing how the concepts they own should behave. An agent developer consuming these files was reading ADK documentation, not universal behavioral principles. Resolution: all product names, level names, and self-referential content removed from all five files.

**SMAQIT.md had no valid L0 content** — the file was entirely smaQit product domain (5-layer system, 3-phase workflow, lifecycle states). No ADK-level principle was present. Resolution: full rewrite as cross-cutting principles that apply across agents, skills, templates, and artifacts simultaneously.

**SKILLS.md body was half catalog** — the five correct principles were followed by product skill catalog, format documentation, loading mechanism details, and historical decision rationale — all of which are either catalog content (copilot-instructions), vocabulary (templates/skills/rules), product guidance (wiki), or history (session records). Resolution: body pruned to principles only; content relocated to appropriate homes.

**`templates/skills/` did not exist** — the ADK had no structural home for skill compilation vocabulary, despite skills being a first-class concept. Resolution: directory created with stub template and L1 vocabulary rules file.

**`smaqit.task-complete` skill not activated** — task status updates were applied as direct file edits rather than through the skill. Root cause: skill description says "finishing tasks" which doesn't match "update task status" or retroactive closure scenarios. Description does not follow its own Description-Driven Activation principle fully.

---

## Decisions Made

- **SMAQIT.md subject = cross-cutting principles** — not compilation philosophy, not framework integrity rules. Content that applies equally to agents, skills, templates, and artifacts simultaneously belongs here. Concept-specific content belongs in its own file.
- **Catalog content moves to copilot-instructions, not dropped** — Agent Catalog (model, invocation, naming, extensions, tooling) and Skill Catalog (location, format, loading stages) were moved to copilot-instructions rather than deleted. This content has value for AI contributors to this repo; it just doesn't belong in L0 framework files.
- **`templates/skills/` created now** — despite `smaqit.new-skill` being deferred, creating the directory now means content lands in the right architecture from the start. No future migration needed.
- **`skill.rules.md` is vocabulary only** — no MUST/MUST NOT/SHOULD directives until a skill template and `smaqit.new-skill` compilation workflow are formally built by L1. Four-type model respected.
- **"Skills vs Input Records" dropped entirely** — historical decision rationale with no persistent home. Captured in session history; not needed in any active artifact.
- **`smaqit.task-complete` description needs improvement** — the fix is in the description, not agent behavior. Proposed rewrite: "Update a task's status to completed, verify its acceptance criteria, and record it in PLANNING.md. Use when marking a task as done — whether just implemented, retroactively closing completed work, or responding to a status update request." Not applied this session; flagged for Task 006 or a separate task.

---

## Files Modified

| File | Change |
|------|--------|
| `framework/SMAQIT.md` | Full rewrite: 5 cross-cutting behavioral principles; all smaQit product content removed |
| `framework/SKILLS.md` | Full rewrite: 5 behavioral principles only; all catalog and meta content removed |
| `framework/TEMPLATES.md` | Full rewrite: 5 structural principles; `## Agent Templates` section dropped; compiler references genericized |
| `framework/AGENTS.md` | Voice-cleaned: removed ADK/compilation self-references; 4 targeted edits |
| `framework/ARTIFACTS.md` | Voice-cleaned: removed "in the ADK" and "compilation chain"; 2 targeted edits |
| `installer/framework/*.md` | Synced (all 5 files) |
| `templates/skills/compiled/skill.rules.md` | Created: L1 vocabulary for skill format, loading stages, agent-skill interaction |
| `templates/skills/skill.template.md` | Created: stub placeholder pending `smaqit.new-skill` compilation workflow |
| `installer/templates/skills/compiled/skill.rules.md` | Created (synced from root) |
| `installer/templates/skills/skill.template.md` | Created (synced from root) |
| `installer/Makefile` | Added `templates/skills/` copy steps to `prepare` target |
| `.github/copilot-instructions.md` | Added Agent Catalog section; added Skill Catalog section; updated `templates/` row |
| `docs/wiki/extending-smaqit.md` | Appended `## Downstream Input Patterns` section; updated Further Reading |
| `.smaqit/tasks/PLANNING.md` | Closed Tasks 005 and 008; added Task 009 to Future; Task 006 marked In Progress |
| `.smaqit/tasks/005_redesign_framework_files.md` | Status → Completed, completion date added |
| `.smaqit/tasks/006_create_new_principle_skill.md` | Status → In Progress |

---

## Next Steps

- **Task 006** — Create `smaqit.new-principle` skill. Prerequisites per task file: (1) revisit `agents/smaqit.L0.agent.md` description and apply Description-Driven Activation; (2) assess whether L0 needs a definition file input pattern before the skill can invoke it as a subagent.
- **`smaqit.task-complete` description** — Update to match natural trigger phrasing for status updates and retroactive closure. Consider creating a task for this or including it in Task 006 as a prerequisite.
- **Task 009** — `smaqit.new-skill` skill; deferred until Task 006 pattern is proven. `templates/skills/` structure already in place.

---

## Session Metrics

- **Tasks completed:** 2 (005, 008 retroactive)
- **Framework files rewritten:** 3 (SMAQIT.md, SKILLS.md, TEMPLATES.md)
- **Framework files voice-cleaned:** 2 (AGENTS.md, ARTIFACTS.md)
- **New files created:** 2 (skill.rules.md, skill.template.md)
- **Files modified total:** 17
- **Principles removed (product content):** 11 from SMAQIT.md alone
- **Principles added (new cross-cutting):** 5 in SMAQIT.md, 1 (Composability — new)
- **Key insight:** Framework files were doing two jobs (behavioral principles + ADK documentation); separating those jobs required routing catalog content to copilot-instructions rather than deleting it
