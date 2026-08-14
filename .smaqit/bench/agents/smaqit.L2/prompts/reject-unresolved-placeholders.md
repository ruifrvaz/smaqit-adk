Read and follow the L2 compiler instructions at the path listed for declared input `l2agent`. This isolated run supplies the base-agent template and base rules at the paths listed for `base-agent-template` and `base-rules`; use those declared inputs instead of any global ADK installation.

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
