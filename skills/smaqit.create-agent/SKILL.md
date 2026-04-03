---
name: smaqit.create-agent
description: Creates a new custom agent for this project. Use when the user asks to create, define, or build a new agent.
metadata:
  version: "1.0"
---

# Create Agent

Invokes the `smaqit.create-agent` agent as a subagent to gather specifications and compile a new `.agent.md` file. Use when the user wants to create a new custom Copilot agent for their project.

## Purpose

Provides a natural language and slash-command entry point for agent creation. All specification gathering and compilation logic lives inside the `smaqit.create-agent` agent. Running as a subagent gives the agent a clean context — free of the current session's loaded agents, instructions, and file context.

## Steps

1. Invoke `smaqit.create-agent` as a subagent. Pass the instruction: "The user wants to create a new agent."

## Output

A compiled `.agent.md` file written to `.github/agents/[name].agent.md`, produced by the `smaqit.create-agent` subagent.

## Scope

This skill triggers agent creation only. It does not gather specifications directly — all gathering and compilation happen inside the subagent.

Out of scope:
- Skill creation — use the `smaqit.create-skill` skill

## Completion

- [ ] `smaqit.create-agent` subagent invoked
- [ ] Subagent confirms output file written

## Failure Handling

| Situation | Action |
|-----------|--------|
| `smaqit.create-agent` agent not installed | Instruct the user to run `smaqit-adk init`, or install the agent manually into `.github/agents/` |
| Subagent invocation fails | Report the failure with context; do not silently retry |
| User wants to create a skill instead | Stop; redirect to the `smaqit.create-skill` skill |
| User abandons creation midway | The subagent handles this case; this skill does not intervene |
