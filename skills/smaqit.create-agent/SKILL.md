---
name: smaqit.create-agent
description: Creates a new agent for this project. Use when the user asks to create, define, or build a new custom agent.
metadata:
  version: "2.0.0"
---

# Create Agent

## Purpose

Creates a new agent by inferring a complete specification from minimal input and project context, writing a definition file, and invoking `smaqit.L2` to compile the agent.

## Steps

### 1. Gather

Ask the user for two things in a single message:
- **Name** — lowercase, hyphens allowed (e.g., `my-reviewer`)
- **Purpose** — what the agent does and what it produces

### 2. Scan

Before writing anything, read:
- All existing files in `.github/agents/` — for naming conventions and tool patterns already used in this project
- Project README — for domain, stack, and conventions
- Any project manifests or config files that reveal project structure

Also extract any relevant detail the user has already provided in the conversation (tools, scope constraints, directives).

### 3. Infer and write definition file

Using name, purpose, and scanned context, infer a complete agent specification. Do not ask further questions.

Write the inferred specification to `.smaqit/definitions/agents/[name].md`. Create the directory if it does not exist.

For any field where the correct value is genuinely ambiguous, suffix the value with `[?]` and add a brief inline note explaining the uncertainty.

The definition file must cover:
- Name and description
- Tools (inferred from purpose and project patterns)
- Input sources (what the agent reads or receives)
- Output (what it produces and where)
- Directives: MUST, MUST NOT, SHOULD
- Scope boundaries (what is out of scope; redirect targets if applicable)
- Completion criteria (testable, checkbox-style)
- Failure scenarios (likely failure modes and responses)

### 4. Compile

Invoke `smaqit.L2` as a subagent with:
> "Compile the agent definition at `.smaqit/definitions/agents/[name].md`. Write the compiled agent to `.github/agents/[name].agent.md`. After compilation, list any fields annotated with `[?]` and suggest a resolution for each."

### 5. Report

After L2 completes, report to the user:
- Path of the compiled agent file
- Any `[?]`-annotated items and L2's suggested resolutions
- How to adjust: edit `.smaqit/definitions/agents/[name].md` and re-invoke `/smaqit.create-agent`, or switch to `smaqit.L2` directly

## Output

- `.smaqit/definitions/agents/[name].md` — inferred specification (scaffolding)
- `.github/agents/[name].agent.md` — compiled agent file (source of truth)

## Scope

Does not create skills, framework files, templates, or Level agents. Does not modify existing agents.

## Completion

- [ ] Name and purpose obtained from user
- [ ] Repository scanned for context
- [ ] Definition file written to `.smaqit/definitions/agents/[name].md`
- [ ] `smaqit.L2` invoked and compilation completed
- [ ] Compiled agent exists at `.github/agents/[name].agent.md`

## Failure Handling

| Situation | Action |
|-----------|--------|
| Name or purpose not provided | Request before proceeding |
| `.smaqit/templates/` not present | Inform the user that ADK templates are required — run `smaqit-adk lite` in this repository first |
| Output artifact already exists | Report the conflict; do not overwrite without user confirmation |
| L2 invocation fails | Report the error and include the path to the definition file so the user can inspect or correct it |
