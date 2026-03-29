# GitHub Copilot Instructions — smaqit-adk

## What this repository is

This is the **smaqit-adk** repository — an Agent Development Kit. It ships a compilation framework (principles → templates → agents) and the tools to use it.

## ADK Artifacts (in scope)

When reasoning about or modifying the ADK, work exclusively with root-level artifacts:

| Path | Purpose |
|------|---------|
| `agents/` | Shipped agents: Level agents (L0, L1, L2) + compiled product agents (smaqit.create-agent, smaqit.create-skill) |
| `skills/` | Advanced-tier skills (smaqit.new-agent, smaqit.new-skill) — require full ADK at runtime, not installed by `init` |
| `framework/` | L0 principle files |
| `templates/` | L1 templates and compilation rules (agents and skills) |
| `installer/` | Go CLI that packages and distributes the above |
| `docs/` | ADK documentation |

## `.github/` contents are NOT ADK artifacts

The `.github/agents/` and `.github/skills/` directories in this repository contain **smaqit product extensions** — agents and skills built using the ADK by the smaQit product. They are a consumer of the ADK, not part of it.

# Directives

**MUST NOT:**
- Use `.github/agents/` or `.github/skills/` as examples of ADK design
- Treat smaqit extension skills (session, task, release, user-testing) as ADK-owned
- Assume the ADK ships or controls any content under `.github/`
- Run `git commit`, `git push`, or any git write operation unless explicitly instructed by the user

**The only exception:** `.github/workflows/` (CI/CD) and `.github/copilot-instructions.md` (this file) are ADK project infrastructure.

## Architecture mental model

```
smaqit-adk (this repo, root)
├── agents/          ← ADK ships these (L0, L1, L2, smaqit.create-agent, smaqit.create-skill)
├── skills/          ← ADK ships these (smaqit.new-agent, smaqit.new-skill) — expert/advanced tier
├── framework/       ← ADK ships these (principle files)
├── templates/       ← ADK ships these (compilation templates)
└── installer/       ← packages all of the above into a binary

.github/             ← NOT shipped by ADK
├── agents/          ← may contain smaqit product agents (external)
└── skills/          ← may contain smaqit product skills (external)
```

When a project runs `smaqit-adk init`, it receives only `smaqit.create-agent.agent.md` and `smaqit.create-skill.agent.md` into `.github/agents/`. No framework files, templates, or skills are written. The `.github/` content in this development repo is smaqit product work that happens to live here — it is not installed by `init`.

## Routing between Level agents

The ADK has three Level agents with distinct responsibilities:

| Agent | Responsibility | Invoke when |
|-------|---------------|-------------|
| `smaqit.L0` | Maintain framework principles | Changing `framework/*.md` |
| `smaqit.L1` | Compile principles to templates | Changing `templates/` |
| `smaqit.L2` | Compile templates to agents and skills | Creating new agents or skills |

Skills that route to a Level agent do so by naming it explicitly in their compilation step — not via frontmatter (which is an industry format not to be extended).

## Agent Catalog

### Agent Model

**Level Agents** — Specialist meta-agents that maintain and compile the framework. Three exist: L0 (principle curator), L1 (template compiler), L2 (agent compiler). Level agents are not primary user-facing entry points; they are subagent targets invoked by skills or switched to deliberately by expert users.

**Product Agents** — Custom agents compiled by the ADK for use in a specific project. Produced by the compilation chain (L0 → L1 → L2). Not part of the ADK itself.

### Invocation Model

Skills are the primary user-facing entry point into ADK workflows.

- Users invoke skills via slash command (`/[skill-name]`) or semantic trigger ("create a new agent")
- The active agent loads the skill, follows its gathering and execution instructions
- When compilation is required, the skill instructs the active agent to invoke the appropriate Level agent as a subagent
- Level agents may also be switched to directly by expert users for deliberate specialist work

Level agents are rarely invoked directly. They receive work either as subagents (programmatic) or as deliberate expert contexts (manual switch).

### Naming Convention

ADK-shipped agents follow the pattern `smaqit.[identifier]`:

| Agent | Pattern | Purpose |
|-------|---------|---------|
| Level 0 | `smaqit.L0` | Principle curator |
| Level 1 | `smaqit.L1` | Template compiler |
| Level 2 | `smaqit.L2` | Agent compiler |

Product agents compiled for a specific project follow a naming convention defined by that project. The ADK does not prescribe product agent names.

### Agent Extensions

Agents are extended through the compilation chain:

- **Base agents** — Foundation behaviors only, customized for a specific purpose via L2 compilation
- **Specification agents** — Foundation + specification workflow extension (L1 spec rules) + domain-specific directives
- **Implementation agents** — Foundation + implementation workflow extension (L1 impl rules) + phase-specific directives

Extensions inherit all foundational behaviors. What differentiates them lives in their compilation rules, not in the foundation.

### Tooling by Role

Agents declare their tool requirements in frontmatter. Tool sets vary by role:

| Role | Typical Tools |
|------|---------------|
| Read-only agents (Q&A, helpers) | `read`, `search`, `fetch` |
| Authoring agents (specification) | `edit`, `search`, `fetch`, `todos` |
| Compilation agents (L0, L1, L2) | `edit`, `search`, `runCommands`, `usages`, `changes`, `todos` |
| Execution agents (implementation) | `edit`, `search`, `runCommands`, `problems`, `changes`, `testFailure`, `todos` |
| Subagent-invoking agents | above + `agent` |

## Skill Catalog

### Location and Shipping

Skills live in `skills/` at the ADK root. They are **advanced-tier ADK infrastructure** — they require L2, framework files, and templates to be present at runtime. They are not installed by `smaqit-adk init` and are intended for ADK contributors and expert users operating the full compilation chain.

Skills in `.github/skills/` of this repository are smaqit product skills, not ADK-shipped skills.

### ADK-Shipped Skills

| Skill | Purpose |
|-------|--------|
| `smaqit.new-agent` | Gather agent specifications interactively, write a definition file, and invoke L2 as a subagent to compile the agent |
| `smaqit.new-skill` | Gather skill specifications interactively, write a definition file, and invoke L2 as a subagent to compile the skill |

### Skill Format

YAML frontmatter + markdown instructions:

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

| Stage | What loads | Constraint |
|-------|-----------|------------|
| Discovery | `name` + `description` only | ~100 tokens |
| Activation | Full `SKILL.md` body | < 5000 tokens recommended |
| Execution | Referenced external files | On demand |

## Framework Content Model

When authoring or reviewing `framework/` files and `templates/agents/compiled/*.rules.md`, apply this four-type model to every content block:

| Type | Answers | Language | Lives at |
|------|---------|----------|----------|
| **Principle** | Why does this matter? | Rationale prose | L0 `framework/` |
| **Invariant** | What is always true when this principle is applied? | Declarative present-tense | L0 `framework/` |
| **Vocabulary / Catalog** | What named things exist and what do they mean? | Definitions, tables, placeholder lists | L1 `templates/agents/compiled/*.rules.md` |
| **Directive** | What must an agent do? | MUST / MUST NOT / SHOULD | L1 `templates/agents/compiled/*.rules.md` |

**Invariant vs directive:** An invariant states what is *true* about a compliant agent (declarative). A directive instructs an agent what to *do* (imperative). L1 reads invariants and compiles them into directive form. Invariant language never uses MUST/MUST NOT/SHOULD.

**Vocabulary vs principle:** A placeholder catalog or named-things table is L1 vocabulary — it describes which things exist in a specific template, not why they exist. Principles are prior to and independent of which specific agents, layers, or placeholders exist.

**MUST NOT** place directives, placeholder catalogs, or product-domain vocabulary tables in `framework/` files.

## Build Workflow

Root artifacts (`agents/`, `skills/`, `framework/`, `templates/`) are the source of truth. The `installer/` directory contains build intermediates that are `.gitignore`d and regenerated on every build.

**MUST NOT** manually copy files into `installer/framework/`, `installer/skills/`, `installer/agents/`, or `installer/templates/`. These are overwritten by `make prepare`.

**To build after editing root artifacts:**
```
cd installer && make build
```

The `prepare` target (run automatically by `build`) copies all root artifacts into the installer before compilation. No manual sync step is required.
