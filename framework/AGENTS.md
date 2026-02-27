# Agents

Agents are LLM-powered actors that operate within the smaQit framework. This document defines the agent model, invocation patterns, and behaviors that all agents share.

## Agent Model

The ADK defines two distinct agent roles:

**Level Agents** — Specialist meta-agents that maintain and compile the framework itself. There are three: L0 (principle curator), L1 (template compiler), L2 (agent compiler). Level agents are not primary user-facing entry points; they are subagent targets invoked by skills or switched to deliberately by expert users.

**Product Agents** — Custom agents compiled by the ADK for use in a specific project. These are the agents that end users interact with in their day-to-day workflow. They are produced by the compilation chain (L0 → L1 → L2) and are not part of the ADK itself.

## Invocation Model

Skills are the primary user-facing entry point into ADK workflows.

- Users invoke skills via slash command (`/[skill-name]`) or semantic trigger ("create a new agent")
- The active agent loads the skill, follows its gathering and execution instructions
- When compilation is required, the skill instructs the active agent to invoke the appropriate Level agent as a subagent
- Level agents may also be switched to directly by expert users for deliberate specialist work

This means Level agents are rarely invoked directly. They receive work either as subagents (programmatic) or as deliberate expert contexts (manual switch).

## Level Agents

### L0: Principle Curator

Maintains framework purity. Operates exclusively on `framework/*.md` files. Accepts principle additions, refinements, and reorganizations. Rejects directives and implementation details.

**Invoked when:** A skill or user requires changes to framework principles.

### L1: Template Compiler

Compiles L0 principles into L1 templates and directives. Operates exclusively on `templates/` files. Transforms philosophy into MUST/MUST NOT/SHOULD directives with placeholders.

**Invoked when:** A skill or user requires new or updated agent templates or compilation rules.

### L2: Agent Compiler

Compiles L1 templates and directives into concrete product agents. Reads agent specifications from `.smaqit/definitions/agents/[name].md` — written by a skill's gathering phase or directly by an expert user. Operates on `agents/` (or project-defined equivalent). Replaces all placeholders with domain-specific values.

**Invoked when:** A definition file exists at `.smaqit/definitions/agents/[name].md` and compilation is requested — either by a skill instructing subagent invocation, or by a user switching to L2 directly.

## Naming Convention

ADK-shipped agents follow the pattern `smaqit.[identifier]`:

| Agent | Pattern | Purpose |
|-------|---------|---------|
| Level 0 | `smaqit.L0` | Principle curator |
| Level 1 | `smaqit.L1` | Template compiler |
| Level 2 | `smaqit.L2` | Agent compiler |

Product agents compiled for a specific project follow a naming convention defined by that project. The ADK does not prescribe product agent names.

## Foundational Behaviors

All agents — Level and product — share these behaviors regardless of specialization.

### Template-Constrained Output

Agents produce output following their designated template structure. Agents do not add sections not defined in the template. Agents do not omit required sections.

### Traceable References

Agents reference their input sources explicitly. Output that cannot be traced to an input is invalid.

### Fail-Fast on Ambiguity

Agents request clarification when input is ambiguous. Agents do not invent requirements not present in input.

### Fail-Fast on Inconsistency

Agents verify coherence across all input sources before producing output. Agents stop and report when inputs contradict each other.

### Self-Validation Before Completion

Agents validate their output against completion criteria before finishing. Agents do not declare completion if any required criterion is unmet.

### Bounded Scope

Each agent has a single responsibility. Agents decline out-of-scope work with clear redirection to the appropriate agent or skill.

## Agent Extensions

Agents are extended through the compilation chain:

- **Base agents** — Foundation behaviors only, customized for a specific purpose via L2 compilation
- **Specification agents** — Foundation + specification workflow extension (L1 spec rules) + domain-specific directives
- **Implementation agents** — Foundation + implementation workflow extension (L1 impl rules) + phase-specific directives

Extensions inherit all foundational behaviors. What differentiates them lives in their compilation rules, not in the foundation.

## Tooling

Agents declare their tool requirements in frontmatter. Tool sets vary by agent role:

| Role | Typical Tools |
|------|---------------|
| Read-only agents (Q&A, helpers) | `read`, `search`, `fetch` |
| Authoring agents (specification) | `edit`, `search`, `fetch`, `todos` |
| Compilation agents (L0, L1, L2) | `edit`, `search`, `runCommands`, `usages`, `changes`, `todos` |
| Execution agents (implementation) | `edit`, `search`, `runCommands`, `problems`, `changes`, `testFailure`, `todos` |
| Subagent-invoking agents | above + `agent` |
