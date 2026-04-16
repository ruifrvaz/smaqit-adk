---
name: smaqit.new-principle
description: Adds or refines a principle in the ADK framework. Use when the user wants to add, update, or consolidate a principle in a framework file.
metadata:
  version: "0.1.0"
---

# New Principle

Invokes the `smaqit.L0` agent as a subagent to gather, validate, and add a principle to the appropriate `framework/*.md` file. Use when the user wants to add or refine a framework principle.

## Purpose

Provides a natural language and slash-command entry point for principle authoring. All specification gathering, form validation, conflict checking, and framework file editing logic lives inside `smaqit.L0`. Running as a subagent gives the agent a clean context — free of the current session's loaded agents, instructions, and file context.

## Steps

1. Invoke `smaqit.L0` as a subagent. Pass the instruction: "The user wants to add or refine a framework principle."

## Output

A principle added or updated in the appropriate `framework/*.md` file, produced by the `smaqit.L0` subagent.

## Scope

This skill triggers principle authoring only. It does not gather or validate specifications directly — all gathering, validation, and framework editing happen inside the subagent.

Out of scope:
- Template authoring — use the `smaqit.new-template` skill
- Compilation rules authoring — use the `smaqit.new-rules` skill
- Agent or skill creation — use `smaqit.new-agent` or `smaqit.new-skill`

## Completion

- [ ] `smaqit.L0` subagent invoked
- [ ] Subagent confirms principle added or updated in `framework/*.md`

## Failure Handling

| Situation | Action |
|-----------|--------|
| `smaqit.L0` agent not installed | Instruct the user to run `smaqit-adk advanced`, or install the agent manually into `.smaqit/agents/` |
| Subagent invocation fails | Report the failure with context; do not silently retry |
| User wants to author templates or rules instead | Stop; redirect to the appropriate skill |
| User abandons creation midway | The subagent handles this case; this skill does not intervene |
