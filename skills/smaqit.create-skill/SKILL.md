---
name: smaqit.create-skill
description: Creates a new skill for this project. Use when the user asks to create, define, or build a new skill.
metadata:
  version: "1.0"
---

# Create Skill

Invokes the `smaqit.create-skill` agent as a subagent to gather specifications and compile a new `SKILL.md` file. Use when the user wants to create a new Copilot skill for their project.

## Purpose

Provides a natural language and slash-command entry point for skill creation. All specification gathering and compilation logic lives inside the `smaqit.create-skill` agent. Running as a subagent gives the agent a clean context — free of the current session's loaded agents, instructions, and file context.

## Steps

1. Invoke `smaqit.create-skill` as a subagent. Pass the instruction: "The user wants to create a new skill."

## Output

A compiled `SKILL.md` file written to `.github/skills/[name]/SKILL.md`, produced by the `smaqit.create-skill` subagent.

## Scope

This skill triggers skill creation only. It does not gather specifications directly — all gathering and compilation happen inside the subagent.

Out of scope:
- Agent creation — use the `smaqit.create-agent` skill

## Completion

- [ ] `smaqit.create-skill` subagent invoked
- [ ] Subagent confirms output file written

## Failure Handling

| Situation | Action |
|-----------|--------|
| `smaqit.create-skill` agent not installed | Instruct the user to run `smaqit-adk lite`, or install the agent manually into `.github/agents/` |
| Subagent invocation fails | Report the failure with context; do not silently retry |
| User wants to create an agent instead | Stop; redirect to the `smaqit.create-agent` skill |
| User abandons creation midway | The subagent handles this case; this skill does not intervene |
