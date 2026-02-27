---
name: smaqit.new-agent
description: Gather agent specifications interactively to compile a new base agent via Agent-L2. Use when creating a new custom agent (Q&A, helper, orchestrator, utility).
metadata:
  version: "0.2.0"
---

# New Agent Creation

Instruct Agent-L2 to compile a new base agent by gathering specifications interactively from the user.

## When to use this skill

Activate when the user wants to create a new custom agent. Triggered by:
- Slash command: `/smaqit.new-agent`
- Semantic: "create a new agent", "build me an agent for...", "I need an agent that..."

This skill handles the full workflow: interactive specification gathering followed by compilation via Agent-L2.

## Steps

Gather the following specifications from the user in order. Request each section systematically — do not infer values without asking.

### 1. Agent Identity

**Name**
- Ask: "What is the agent name? (lowercase, single word or hyphenated — e.g., `qa`, `release-helper`)"
- Used in: filename (`agents/[name].agent.md`), frontmatter `name` field

**Description**
- Ask: "What is the agent description? (clear, concise, under 80 characters)"
- Used in: frontmatter `description` field

**Tools**
- Ask: "What tools does this agent need? (comma-separated from: fetch, search, read, edit, runCommands, usages, todos, problems, changes, testFailure)"
- Used in: frontmatter `tools` list

### 2. Agent Purpose

**Goal**
- Ask: "What is this agent's goal? (what it produces or accomplishes)"
- Example: "Fetch and answer questions about project documentation"

**Context**
- Ask: "What constraints or scope define this agent?"
- Example: "Read-only operations, documentation focus, no code generation"

### 3. Input Sources

- Ask: "What input sources does this agent read or consume?"
- Examples: "User questions", "Files from `docs/` directory", "GitHub wiki URLs"

### 4. Output Format

- Ask: "What does this agent produce?"
- Examples: "Direct answers with source references", "Generated configuration files in JSON"

### 5. Specialized Directives

**MUST**
- Ask: "What are the agent-specific MUST directives? (imperative statements, one per line)"
- Examples: "Fetch documentation from GitHub when local files unavailable", "Provide source references for all answers"

**MUST NOT**
- Ask: "What are the agent-specific MUST NOT directives? (prohibitions, one per line)"
- Examples: "Generate code or implementation", "Modify any files"

**SHOULD**
- Ask: "What are the agent-specific SHOULD directives? (recommendations, one per line)"
- Examples: "Prefer local files over remote fetch when available"

### 6. Scope Boundaries

- Ask: "What is explicitly out of scope? Include redirections to other agents where applicable."
- Examples: "Code generation → redirect to Development agent", "File modifications → read-only operations only"

### 7. Completion Criteria

- Ask: "What are the agent-specific completion criteria? (testable checks beyond foundation criteria)"
- Examples: "Answer addresses user's question directly", "At least one source reference provided"

### 8. Failure Scenarios

- Ask: "What are the agent-specific failure scenarios? Provide as situation / action pairs."
- Example pairs:
  - Documentation not found → "Respond: 'Documentation not found for [topic]. Available sections: [list]'"
  - Ambiguous question → "Request clarification: 'Did you mean [A] or [B]?'"

## Validation

After gathering all specifications:

1. **Validate directives** — Check user-provided MUST/MUST NOT/SHOULD statements are compatible with foundation directives in `templates/agents/compiled/base.rules.md`. Flag conflicts before compiling.
2. **Validate completeness** — All 8 sections must have content. Request missing information before proceeding.
3. **Validate directive form** — Directives must be imperative statements starting with verbs. Reject narrative or philosophical statements.
4. **Confirm** — Present a summary of gathered specifications and ask user to confirm before compiling.

## Compilation

Once specifications are confirmed:

1. Write a definition file to `.smaqit/definitions/agents/[name].md` containing all gathered specifications in the format below
2. Use the `agent` tool to invoke `smaqit.L2` as a subagent, passing the definition file path as context
3. Agent-L2 will read the definition file and perform the 3-way merge:
   - `templates/agents/base-agent.template.md` — structure
   - `templates/agents/compiled/base.rules.md` — foundation directives
   - `.smaqit/definitions/agents/[name].md` — gathered specifications
4. Output: `agents/[name].agent.md` created and validated by L2
5. Compilation log written to `.smaqit/logs/[name]-compilation-[YYYY-MM-DD].md` by L2

### Definition File Format

```markdown
# Agent Definition: [name]

**Created:** YYYY-MM-DD
**Skill:** smaqit.new-agent

## Identity

- **Name:** [name]
- **Description:** [description]
- **Tools:** [tool1, tool2, ...]

## Purpose

- **Goal:** [goal]
- **Context:** [context/constraints]

## Input Sources

[list of input sources]

## Output Format

[output description]

## Directives

### MUST
- [directive 1]
- [directive 2]

### MUST NOT
- [directive 1]
- [directive 2]

### SHOULD
- [directive 1]

## Scope Boundaries

[what is out of scope, redirections]

## Completion Criteria

- [ ] [criterion 1]
- [ ] [criterion 2]

## Failure Scenarios

| Situation | Action |
|-----------|--------|
| [situation] | [action] |
```

## Notes

- The definition file at `.smaqit/definitions/agents/[name].md` is the auditable record of what was requested. The compilation log documents what L2 produced from it.
- This skill covers base agents only. Specification and implementation agents require domain/phase rules (via Agent-L1) before Agent-L2 can compile them
- The `[EXTENSION_MUST_DIRECTIVES]` placeholder in the base template is filled by user-provided MUST directives — these are agent-specific behaviors, not workflow extensions
- Expert users can write a definition file directly and switch to `@smaqit.L2` without using this skill
