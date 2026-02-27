# Skills

Skills are the user-facing input mechanism for agent workflows. They contain instructions for agents — what to ask, what to validate, how to compile.

## Core Principles

### Instructions, Not Data

**Skills contain instructions for agents, not accumulated requirements or user data.**

A skill tells an agent what to gather, how to validate, and how to compile. User requirements belong in conversation context during execution, or in compilation artifacts after completion. A skill that stores user data conflates the instruction format with the data format and breaks reproducibility across runs.

### Progressive Disclosure

**Agents load only the skill name and description at startup — the full body loads only on activation.**

Discovery is cheap: a few tokens per skill across all installed skills. The full `SKILL.md` body is read only when the agent determines the skill is relevant to the current task. This means the body can carry detailed instructions without startup cost, and the `description` field carries the full weight of the activation decision.

### Instruction-Only Content

**Skill bodies contain gathering steps, validation rules, and compilation guidance — nothing else.**

Skills do not accumulate user requirements over time. If a skill references supporting material, it does so through explicit file paths in `references/` or `assets/`. The body is a procedure, not a record.

### Context-Driven Input

**Requirements gathered during skill execution live in conversation context, not in the skill file.**

If persistence is needed after execution, the agent documents gathered input in a compilation log or a definition file. The skill itself is unchanged between runs. This preserves the skill as a read-only instruction artifact and prevents version-controlled skill files from diverging between projects.

### Description-Driven Activation

**Skill descriptions are the sole signal used at discovery — they must explain what the skill is and when to use it, not just name it.**

At discovery, the agent reads only `name` and `description` to decide whether a skill matches the user's request. The description is not a label or a tagline — it is an explanation. Write it to complete two sentences: *"This skill does..."* and *"Use this skill when..."*. The agent uses that explanation semantically, so a well-formed description that answers those two questions will match correctly even without keyword overlap.

Length is not the constraint. A description can be several sentences or a short paragraph. Brevity is not a virtue here; precision is. An explanation long enough to rule out adjacent skills is better than a short one that matches ambiguously.

**What makes a description precise:**

- **Explain, don't label** — "Guided specification gathering for a new agent's purpose, tools, directives, scope, and behavior, compiled into a new agent file" explains. "New agent creation" labels. The explanation gives the agent context to distinguish this skill from adjacent ones.
- **State the output** — Include what is produced. Two skills sharing the same verb (`"create"`, `"start"`, `"run"`) are distinguished by what they produce, not by the verb alone.
- **Embed keywords in explanation** — Keywords are useful but only when they appear inside explanatory sentences. A list of keywords (`"orchestrator, Q&A, utility"`) broadens the trigger surface unpredictably; the same words inside a sentence that describes the skill boundary do not.
- **Exclude mechanism** — How the skill works internally (`"via Agent-L2"`, `"interactively"`, `"using a 3-way merge"`) is body content. The description states what the user gets, not how it is produced.
- **Test for false positives** — For every distinctive noun in the description, ask: "Could a user with a different intent trigger this?" If yes, rephrase to narrow the match. A description that misses occasional valid invocations is recoverable; one that triggers falsely cannot be untriggered.

## Skills vs Input Records

**Previous model (prompts as input records):** Users wrote requirements into `.github/prompts/[domain].prompt.md`. Agents read the file as their primary input. The filled file was committed alongside specs.

**Current model (skills + context):** Skills instruct agents on how to gather input interactively. Requirements live in conversation context during execution. The agent documents what it received in a compilation log if auditability is needed. Downstream products that need persistent requirement files should define their own input pattern (e.g., a `requirements/` directory with user-managed files that their agents read directly — distinct from skills).

**Why the change:** Skills are a read-only format for skill authors. Treating them as user-editable input records conflates instructions with data and breaks the progressive disclosure model.

## Skill Structure

### Location

Skills live in `skills/` at the root of the ADK repository (shipped artifact). When installed via `smaqit-adk init`, they are copied to `.github/skills/` in the consuming project.

**ADK skills structure:**
```
skills/
└── smaqit.new-agent/
    └── SKILL.md
```

### Format

**YAML Frontmatter + Markdown Instructions**

```
---
name: skill-name
description: What this skill does and when to use it.
metadata:
  version: "1.0"
---

# Skill Title

## Steps
...
```

### Loading Stages

| Stage | What loads | Size |
|-------|-----------|------|
| Discovery | `name` + `description` only | ~100 tokens |
| Activation | Full `SKILL.md` body | < 5000 tokens recommended |
| Execution | Referenced files (references/, assets/) | On demand |

## Skill Types

### Workflow Skills

Manage session, task, and release workflows:

| Skill | Purpose |
|-------|---------|
| `smaqit.session-start` | Load project context at session start |
| `smaqit.session-finish` | Document session and create history entry |
| `smaqit.task-create` | Create a new task with auto-numbering |
| `smaqit.task-start` | Begin work on a task |
| `smaqit.task-complete` | Mark a task complete with verification |
| `smaqit.task-list` | Show active task overview |
| `smaqit.release-analysis` | Assess changes and suggest next version |
| `smaqit.release-prepare-files` | Prepare CHANGELOG and version files |
| `smaqit.release-git-local` | Execute git operations for local release |
| `smaqit.release-git-pr` | Execute git operations for PR-based release |

### Agent Creation Skill

Guides Agent-L2 through gathering specifications for new custom agents:

| Skill | Agent Type | Invokes |
|-------|------------|---------|
| `smaqit.new-agent` | All base agents | Agent-L2 |

**Agent Creation Workflow:**

1. **User invokes skill** — Via slash command (`/smaqit.new-agent`) or semantic trigger ("create a new agent", "I need an agent that...")
2. **Active agent loads skill** — Reads full `SKILL.md` body into context
3. **Interactive gathering** — Active agent follows skill's gathering steps, requesting user input for each specification (name, description, tools, directives, scope, completion criteria, failure scenarios)
4. **Write definition file** — Skill instructs active agent to write `.smaqit/definitions/agents/[name].md` containing all gathered specifications
5. **Subagent invocation** — Skill instructs active agent to invoke `smaqit.L2` as a subagent, passing the definition file path
6. **3-way merge** — L2 reads definition file + L1 base template + L1 base rules → compiles custom agent
7. **Compilation log** — L2 writes `.smaqit/logs/[name]-compilation-[YYYY-MM-DD].md` documenting the process
8. **Output agent** — `agents/[name].agent.md` created and returned

## Agent Interaction with Skills

Agents interact with skills through progressive loading:

1. **Discovery** — At startup, agent loads `name` and `description` from all available skills to know when each is relevant
2. **Activation** — When a task matches a skill's description, agent reads the full `SKILL.md` into context
3. **Execution** — Agent follows instructions, optionally loading `references/` or `assets/` files as needed

## Downstream Input Patterns

ADK skills do not prescribe how downstream products should handle persistent requirement storage. Products that need users to write requirements into files should define their own input file pattern — separate from skills. Common approaches:

- A `requirements/` directory with user-managed markdown files that product agents read directly
- A `context/` or `prompts/` directory maintained by the product (not by ADK)
- Any other user-editable file convention the product defines

The ADK's generic directive form (`MUST read from [user-defined input path]`) leaves this decision to the product.
