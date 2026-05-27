---
name: smaqit.compile
description: Orchestrates the full or partial ADK compilation chain (L0→L1→L2) — collects target type, name, and execution mode, then delegates to smaqit.orchestrate to propagate principle, template, or agent definition changes downstream through sequential level-agent subagent invocations.
metadata:
  version: "1.0.0"
---

# Compile

## Steps

### 1. Gather

Ask the user in a single message:
- What is the **compilation target**? (principle, template/rules, or agent/skill)
- What is the **target name or file path**?
- What **execution mode** do you prefer? (`assisted` — review each phase before continuing; `autonomous` — run all phases without interruption)

If any of this information was already provided in the conversation, skip asking for it.

### 2. Invoke

Invoke `smaqit.orchestrate` as a subagent with all gathered context:

> "Compile [target type] '[target name/path]' in [mode] mode. [Any additional context the user provided.]"

### 3. Report

After the orchestration agent completes, report to the user:
- All output file paths produced by the chain
- Any phases that were skipped (e.g., L0 and L1 skipped for `agent` targets)
- How to adjust: edit the relevant definition file and re-invoke `/smaqit.compile`

## Output

Compilation outputs produced through `smaqit.orchestrate`:

- `framework/*.md` — updated principle file (principle targets only)
- `templates/**/*.template.md` or `templates/**/compiled/*.rules.md` — updated template or rules (principle and template targets)
- `agents/*.agent.md` or `skills/[name]/SKILL.md` — compiled artifact (all targets)

## Scope

Does not compile anything directly. All compilation happens inside the `smaqit.orchestrate` subagent and the level agents it invokes.

Out of scope:
- Creating new agents from scratch — use `/smaqit.create-agent`
- Creating new skills from scratch — use `/smaqit.create-skill`
- Adding or refining a single principle only — use `/smaqit.new-principle`

## Examples

**Input:** "I updated the principle in `framework/SMAQIT.md`. Compile the full chain."  
**Output:** Updated `framework/SMAQIT.md` (L0), updated `templates/agents/compiled/base.rules.md` (L1), recompiled `agents/smaqit.L2.agent.md` (L2).

**Input:** "Recompile `smaqit.create-agent` against the current rules."  
**Output:** Recompiled `.github/agents/smaqit.create-agent.agent.md` via L2 only.

## Gotchas

- `smaqit.orchestrate`, `smaqit.L0`, `smaqit.L1`, and `smaqit.L2` must all be installed. Run `smaqit-adk advanced` if any are missing.
- Compilation chains are ordered: L0 must complete before L1; L1 must complete before L2. The orchestration agent enforces this ordering.
- For `agent` targets, the chain is L2-only. Do not use this skill if only a framework principle needs updating without downstream propagation — use `/smaqit.new-principle` instead.

## Completion

- [ ] Target type, name, and mode gathered from user
- [ ] `smaqit.orchestrate` invoked with full context
- [ ] All output file paths reported to the user

## Failure Handling

| Situation | Action |
|-----------|--------|
| Target type or name not provided | Ask for the missing information before invoking the subagent |
| `smaqit.orchestrate` agent not installed | Instruct the user to run `smaqit-adk advanced` to install the full ADK |
| Subagent invocation fails | Report the failure with context; include all output paths collected before the failure |
| User wants to create a new agent, not recompile | Redirect: "Use `/smaqit.create-agent` to create a new agent from scratch" |
| User wants to add a principle only (no downstream compilation) | Redirect: "Use `/smaqit.new-principle` to add or refine a principle without triggering a compilation chain" |
