---
name: smaqit.new-principle
description: Adds or refines a principle in an ADK framework file — delegates gathering, validation, and authoring to smaqit.L0 to produce a compliant principle entry in the appropriate framework/*.md file.
metadata:
  version: "0.2.0"
---

# New Principle

Invokes the `smaqit.L0` agent to gather, validate, and add a principle to the appropriate `framework/*.md` file. Use when the user wants to add or refine a framework principle. Framework files are installed globally (`~/.agents/smaqit-adk/framework/`), so an edit applies across every project on the machine, not just the current one.

## Steps

1. Invoke `smaqit.L0`:
   - **On Claude Code or Codex CLI:** invoke it as a native subagent/custom-agent call.
   - **On GitHub Copilot** (no dedicated compiled `smaqit.L0` file exists): read `~/.claude/agents/smaqit-L0.md`'s body directly and follow its instructions inline for this turn, as if it had been invoked as a subagent.

   Pass the instruction: "The user wants to add or refine a framework principle."

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
| `smaqit.L0` agent not installed | Instruct the user to run the smaqit-adk installer (`install.sh`), which installs `smaqit.L0` globally to `~/.claude/agents/` and `~/.codex/agents/` |
| Subagent invocation fails | Report the failure with context; do not silently retry |
| User wants to author templates or rules instead | Stop; redirect to the appropriate skill |
| User abandons creation midway | The subagent handles this case; this skill does not intervene |
