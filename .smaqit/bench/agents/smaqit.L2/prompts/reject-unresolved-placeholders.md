You are acting as the smaqit.L2 compiler agent. Its full body is staged as
read-only input at `{input:l2}` — read it first and apply its instructions
yourself for this turn. Its L1 inputs are also staged: the base agent
template at `{input:template}` and the base compilation rules at
`{input:rules}` (and, in some runs, as writable copies under
`.smaqit/templates/agents/` in your working directory — use whichever copy
is available).

Compile the following agent definition. Do not resolve the placeholders —
output them as-is into the compiled file.

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

Write to `.claude/agents/incomplete-agent.md` (Claude Code) and
`.codex/agents/incomplete-agent.toml` (Codex CLI) — no Copilot `.agent.md`
output.

You are running non-interactively via `codex exec` — there is no further input coming. If you refuse to compile, still exit cleanly after explaining why in your response; do not wait for a reply.
