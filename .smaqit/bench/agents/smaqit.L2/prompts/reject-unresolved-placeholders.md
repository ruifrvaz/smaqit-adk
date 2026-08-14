If the Case brief lists a variant treatment artifact named `l2agent`, read and follow its compiler instructions. In that same treatment table, use `base-agent-template` and `base-rules` as the authoritative compilation sources. If those treatments are absent, proceed without ADK guidance and do not search project-global or user-global ADK locations.

Compile the following agent definition. Do not resolve the placeholders — output them as-is into either compiled file.

# Agent Definition: incomplete-agent

**Created:** 2026-03-29
**Skill:** smaqit.create-agent

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

Write to `.claude/agents/incomplete-agent.md` and `.codex/agents/incomplete-agent.toml`.

You are running non-interactively via `codex exec` — there is no further input coming. If you refuse to compile, still exit cleanly after explaining why in your response; do not wait for a reply.
