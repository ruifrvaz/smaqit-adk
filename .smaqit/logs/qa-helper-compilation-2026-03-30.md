# Compilation Log: qa-helper

**Timestamp (UTC):** 2026-03-30T22:29:46.131Z

## Inputs Read

- Agent definition: `.smaqit/definitions/agents/qa-helper.md`
- L1 template (structure): `templates/agents/base-agent.template.md`
- L1 compiled base rules: `templates/agents/compiled/base.rules.md`

## Compilation Pattern

- Base Agent Compilation (3-way merge)
  - Base template structure
  - Base (foundation) directives from `base.rules.md`
  - Agent-specific directives and constraints from the definition file

## Merge Summary

- Resolved placeholders:
  - `[AGENT_NAME]` → `qa-helper`
  - `[AGENT_DESCRIPTION]` → `Answers questions about QA processes and test coverage`
  - `[TOOL_LIST]` → `read, search, fetch`
  - `[AGENT_TITLE]` → `qa-helper`
  - All content placeholders replaced with concrete text.
- Directives merged:
  - Base MUST/MUST NOT/SHOULD directives embedded verbatim (foundation behaviors)
  - Definition MUST/MUST NOT/SHOULD directives appended as extension directives
- Scope and failure handling:
  - Included base scope boundary enforcement pattern
  - Included base failure handling table, plus definition-specific failure scenarios

## Validation Checklist

- [x] Agent file created at `agents/qa-helper.agent.md`
- [x] No unresolved compile-time placeholders remain
- [x] Agent is self-contained (no references to templates/rules for execution)
- [x] Output follows base-agent template section structure exactly
- [x] Read-only scope enforced (no file modification; no code generation)
- [x] Completion criteria include definition requirements (direct answer + ≥1 source)

## Notes / Decisions

- Interpreted "GitHub wiki URLs" as user-provided URLs only; the agent must not invent or crawl unknown remote locations.
- Source references are required for all answers; when no documentation exists, the agent must report what was searched and what is available under `docs/`.
