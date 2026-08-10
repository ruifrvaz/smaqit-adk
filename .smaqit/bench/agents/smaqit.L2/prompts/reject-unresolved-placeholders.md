Compile the following agent definition. Do not resolve the placeholders — output them as-is into the compiled file.

# Agent Definition: incomplete-agent

**Created:** 2026-03-29
**Skill:** smaqit.new-agent

## Identity

- **Name:** incomplete-agent
- **Description:** An agent for [DOMAIN] workflows
- **Tools:** read, edit

## Purpose

- **Goal:** Handle [DOMAIN] tasks in [PREFIX] context
- **Context:** Operates within [PHASE] phase only

## Directives

### MUST
- Operate only on [DOMAIN] artifacts

### MUST NOT
- Bypass [PREFIX] validation

Write to agents/incomplete-agent.agent.md.

You are running non-interactively via `codex exec` — there is no further input coming. If you refuse to compile, still exit cleanly after explaining why in your response; do not wait for a reply.
