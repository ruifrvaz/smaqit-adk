# Skills

Skills are the user-facing input mechanism for agent workflows. They contain instructions for agents — what to ask, what to validate, how to compile.

**Key Principles:**

- **Skills are instructions, not data stores** — A skill tells an agent how to gather and process input. User requirements are held in conversation context.
- **Progressive disclosure** — Agents load only the skill name and description at startup. The full `SKILL.md` is loaded only when the skill is activated for a task.
- **Instruction-only content** — Skills contain step-by-step gathering instructions, validation rules, and compilation guidance. They do not accumulate user requirements over time.
- **Context-driven input** — Requirements gathered during a skill execution live in the conversation context. If persistence is needed, the agent documents them in a compilation log or a user-managed file.

**Structure:**

- YAML frontmatter: `name`, `description`, `metadata`
- Markdown body: gathering steps, validation rules, compilation instructions
- Optional `references/` subdirectory for detailed reference material
- Optional `assets/` subdirectory for templates and static resources

**Skill types:**

- **Workflow skills** — Session, task, and release management (session-start, task-create, release-analysis, etc.)
- **Agent creation skills** — Gathering specifications to compile new agents (smaqit.new-agent)

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

### Progressive Disclosure

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
