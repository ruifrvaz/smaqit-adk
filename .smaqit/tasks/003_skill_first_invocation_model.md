# Skill-First Invocation Model

**Status:** Completed
**Created:** 2026-02-27
**Completed:** 2026-02-27

## Description

Refactor the ADK to reflect the correct invocation architecture, established through assessment:

- **Skills are the user-facing API.** Users invoke skills via slash command (`/smaqit.new-agent`) or semantic trigger ("I need to create a new agent"). Not Level agents.
- **Level agents are specialist subagents.** They are invoked programmatically by the active agent following skill instructions, or switched to deliberately by expert users. They are not primary entry points.
- **No orchestrator agent.** A dedicated orchestrator is unnecessary and unreachable (slash commands do not work for agents — only for prompts and skills). The default Copilot agent + skill instructions + subagent invocation is the full pattern.
- **Skill instructs, agent executes.** The skill's compilation step must explicitly name the target Level agent (`smaqit.L2`) and instruct the active agent to invoke it as a subagent. This is not left to inference.

## Confirmed Architecture

```
User
  │ /smaqit.new-agent  (slash command)
  │ "create a new agent"  (semantic trigger)
  ▼
Default Copilot agent
  ├── loads smaqit.new-agent skill
  ├── executes gathering steps (skill instructions)
  └── follows skill compilation step:
          invoke smaqit.L2 as subagent with gathered specs in context
              └── L2 compiles → outputs agent file
```

Level agents (L0, L1, L2) are also reachable by deliberate user switch for expert work (e.g., user wants to modify a framework principle → switches to `@smaqit.L2`). But this is not the primary path.

## Acceptance Criteria

### `skills/smaqit.new-agent/SKILL.md`
- [ ] "When to use" removes any reference to invoking `/smaqit.L2` first
- [ ] Compilation step explicitly instructs: invoke `smaqit.L2` as subagent, pass gathered specs as context
- [ ] Skill is self-contained: a user can trigger it without knowing L2 exists

### `agents/smaqit.L2.agent.md`
- [ ] Framing updated: L2 is invoked as subagent (receives specs in context) OR switched to directly by expert user
- [ ] Interactive gathering loop removed from "For Base Agents" section — gathering already completed by skill before L2 is invoked
- [ ] Input section: gathered specs via subagent context listed as primary input source
- [ ] "MUST activate the smaqit.new-agent skill" directive removed — L2 no longer initiates the skill
- [ ] "MUST read the skill instructions from..." replaced with "MUST read gathered specs from context"

### `agents/smaqit.L0.agent.md`
- [ ] Invocation framing: "switched to by user for expert work" or "invoked as subagent by skill" — not slash-command primary
- [ ] Boundary enforcement updated to reflect subagent model

### `agents/smaqit.L1.agent.md`
- [ ] Same invocation framing update as L0

### `framework/AGENTS.md`
- [ ] Full rewrite: remove all smaQit product contamination (business/functional/stack agents, prompt file references, BUS-LOGIN-001 examples, spec frontmatter timestamps, Task 072/073 notes)
- [ ] Documents ADK's 3 Level agents: L0 (principle curator), L1 (template compiler), L2 (agent compiler)
- [ ] Documents invocation model: skills as entry points, Level agents as subagents/expert contexts
- [ ] Documents naming convention for ADK-shipped agents
- [ ] Preserves generic concepts (naming patterns, tool lists, scope boundaries) without product specifics

### `framework/SKILLS.md`
- [ ] "Agent Creation Workflow" steps updated: skill activates on default agent → gathers → instructs subagent invocation of L2
- [ ] Removes "User invokes Agent-L2" as step 1
- [ ] Documents both invocation paths: slash command and semantic trigger

### `README.md`
- [ ] Quick Start step 2: remove `/smaqit.L2` as entry point; replace with `/smaqit.new-agent` slash or semantic
- [ ] Agents table: reflect that Level agents are specialist/subagent targets, not primary user commands
- [ ] Level Architecture section: add note on invocation model

### `docs/wiki/extending-smaqit.md`
- [ ] Workflow description updated to reflect skill-first model

## Out of Scope

- Creating a `smaqit` orchestrator agent (explicitly decided against)
- Changing compilation logic inside L2 (3-way/4-way merge stays as-is)
- Modifying `framework/SMAQIT.md`, `framework/TEMPLATES.md`, `framework/ARTIFACTS.md`
- Any `.github/agents/` or `.github/skills/` content (smaqit product, not ADK)

## Notes

The gathering steps in `smaqit.new-agent/SKILL.md` and the "For Base Agents" interactive gathering sequence in `smaqit.L2.agent.md` are currently duplicated. Once the skill owns gathering and L2 receives specs from context, the L2 duplication is removed. L2's compilation architecture (merge patterns, section-level compilation, placeholder resolution) stays intact.
