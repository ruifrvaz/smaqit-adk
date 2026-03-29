---
name: smaqit.new-agent
description: Guides the creation of a new agent for this project. Use this skill when the user wants to define and compile a new agent — the skill gathers the agent's purpose, tools, directives, scope, completion criteria, and failure scenarios through an interactive interview, then writes a definition file and invokes Agent-L2 to compile the agent file.
metadata:
  version: "0.2.0"
---

# New Agent Creation

## Purpose

Guides an interactive specification interview to define a new base agent, writes a definition file, and invokes smaqit.L2 to compile the agent. Use when the user wants to create a new custom agent for their project.

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

## Output

- **Definition file:** `.smaqit/definitions/agents/[name].md` — auditable specification record
- **Compiled agent:** `agents/[name].agent.md` — produced and validated by smaqit.L2
- **Compilation log:** `.smaqit/logs/[name]-compilation-[YYYY-MM-DD].md` — L2 audit trail

## Scope

Base agents only. Specification agents and implementation agents require domain and phase rules (via smaqit.L1) before smaqit.L2 can compile them. Redirect requests for those agent types to smaqit.L1 first.

## Completion

- [ ] All 8 specification sections gathered and confirmed by user
- [ ] Definition file written to `.smaqit/definitions/agents/[name].md`
- [ ] smaqit.L2 invoked and has produced `agents/[name].agent.md` without errors

## Failure Handling

| Situation | Action |
|-----------|--------|
| User-provided directive conflicts with foundation directives | Flag the conflict in Validation; request clarification before proceeding |
| Any specification section is incomplete | Request the missing information before moving to Validation |
| smaqit.L2 reports unresolved placeholders | Return to definition file, resolve gaps, retry compilation |
| User requests a specification or implementation agent | Explain base-only scope; redirect to smaqit.L1 for domain/phase rules first |

## Notes

- This skill covers base agents only. Specification and implementation agents require domain/phase rules (via Agent-L1) before Agent-L2 can compile them
- The definition file at `.smaqit/definitions/agents/[name].md` is the auditable record of what was requested; the compilation log documents what L2 produced from it
