# Instructions — smaqit-adk

## What this repository is

This is the **smaqit-adk** repository — an Agent Development Kit. It ships a compilation framework (principles → templates → agents) and the tools to use it. It targets Claude Code and Codex CLI as compiled, first-class platforms; GitHub Copilot participates through accepted standards this file and the shared skills path already satisfy — it is not a degraded fallback, just a different mechanism.

## ADK Artifacts (in scope)

When reasoning about or modifying the ADK, work exclusively with root-level artifacts:

| Path | Purpose |
|------|---------|
| `agents/` | Shipped agents: Level agents (L0, L1, L2) — shared body + per-platform metadata, compiled to Claude `.md` and Codex `.toml` |
| `skills/` | ADK skills: create-agent, create-skill, new-principle, bench-run, bench-scaffold — single cross-platform `SKILL.md` format |
| `framework/` | L0 principle files |
| `templates/` | L1 templates and compilation rules (agents and skills) |
| `src/` | Go source for shipped engine capabilities (e.g. `src/bench`, `src/benchcli` — the Bench evaluation/benchmark engine and CLI) |
| `installer/` | Go CLI that packages and distributes the above, plus the generator that renders per-platform agent output |
| `docs/` | ADK documentation |

## `.smaqit/` in this repo is dogfood data, not an ADK artifact

`.smaqit/bench/` holds this repo's own Bench manifests — smaqit-adk using its own shipped `bench` engine (`src/bench`/`src/benchcli`) to benchmark its own skills and agents. It is local project state, in the same sense `.smaqit/tasks/` or `.smaqit/compendium.md` are, not ADK-shipped source and not something the global installer writes into a consuming project. See `.smaqit/bench/README.md`.

# Directives

**MUST NOT:**
- Treat any project-local `.github/agents/`, `.github/skills/`, `.claude/agents/`, `.claude/skills/`, `.codex/agents/` directory as an ADK install target — the ADK installs globally only, never into a project
- Assume smaqit-adk ships or controls any project-local agent/skill directory
- Run `git commit`, `git push`, or any git write operation unless explicitly instructed by the user

## Architecture mental model

```
smaqit-adk (this repo, root)
├── agents/          ← ADK ships these (L0, L1, L2 — shared body + per-platform metadata)
├── skills/          ← ADK ships these (create-agent, create-skill, new-principle, bench-run, bench-scaffold)
├── framework/       ← ADK ships these (principle files)
├── templates/       ← ADK ships these (compilation templates)
├── src/             ← ADK ships these (bench engine + CLI Go source)
└── installer/       ← packages all of the above into a binary, including the per-platform generator

Global install targets (outside any project, one per machine):
~/.claude/agents/        ← L0, L1, L2 (Claude Code)
~/.codex/agents/         ← L0, L1, L2 (Codex CLI)
~/.agents/skills/        ← all 5 skills (Copilot + Codex)
~/.claude/skills/        ← all 5 skills (Claude Code)
~/.agents/smaqit-adk/    ← templates + framework
```

`install.sh` triggers the global install automatically after downloading the binary. No project directory receives any smaqit-adk artifact. A consuming project's own `.smaqit/definitions/` only appears lazily, created by `create-agent`/`create-skill` the first time they're used in that project.

## Routing between Level agents

The ADK has three Level agents with distinct responsibilities:

| Agent | Responsibility | Invoke when |
|-------|---------------|-------------|
| `smaqit.L0` | Maintain framework principles | Changing `framework/*.md` |
| `smaqit.L1` | Compile principles to templates | Changing `templates/` |
| `smaqit.L2` | Compile templates to agents and skills | Creating new agents or skills |

Skills that route to a Level agent do so by naming it explicitly in their compilation step, with platform-conditional invocation: a native subagent/custom-agent call on Claude Code or Codex CLI, or a direct-read-and-follow-inline instruction (reading the Claude-format agent body) on GitHub Copilot, which has no dedicated compiled file of its own.

## Agent Catalog

### Agent Model

**Level Agents** — Specialist meta-agents that maintain and compile the framework. Three exist: L0 (principle curator), L1 (template compiler), L2 (agent compiler). Level agents are not primary user-facing entry points; they are subagent targets invoked by skills or switched to deliberately by expert users.

**Product Agents** — Custom agents compiled by the ADK for use in a specific project. Produced by the compilation chain (L0 → L1 → L2). Not part of the ADK itself.

### Invocation Model

Skills are the primary user-facing entry point into ADK workflows.

- Users invoke skills via slash command (`/[skill-name]`) or semantic trigger ("create a new agent")
- The active agent loads the skill, follows its gathering and execution instructions
- When compilation is required, the skill instructs the active agent to invoke the appropriate Level agent as a subagent (Claude Code/Codex CLI native call, or Copilot's read-and-follow-inline fallback)
- Level agents may also be switched to directly by expert users for deliberate specialist work

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

Each of L0/L1/L2's per-platform metadata declares its own explicit tool mapping (see `installer/` generator) — no blanket default. Typical tool sets by role, expressed in each platform's own vocabulary:

| Role | Typical Tools |
|------|---------------|
| Read-only agents (Q&A, helpers) | Read, Grep, Glob, WebFetch |
| Authoring agents (specification) | Read, Write, Edit, Grep, Glob, WebFetch, TodoWrite |
| Compilation agents (L0, L1, L2) | Read, Write, Edit, Bash, Grep, Glob, TodoWrite |
| Execution agents (implementation) | Read, Write, Edit, Bash, Grep, Glob, TodoWrite |
| Subagent-invoking agents | above + Task |

## Skill Catalog

### Location and Shipping

Skills live in `skills/` at the ADK root, installed globally to `~/.agents/skills/` and `~/.claude/skills/` — no tiers, all 5 always installed:

| Skill | Purpose |
|-------|---------|
| `smaqit.create-agent` | Gather name+purpose, infer spec, write definition file, invoke L2 to compile agent (Claude `.md` + Codex `.toml`) |
| `smaqit.create-skill` | Gather name+purpose, infer spec, write definition file, invoke L2 to compile skill (single cross-platform `SKILL.md`) |
| `smaqit.new-principle` | Add or refine a principle in the ADK framework files |
| `smaqit.bench-run` | Preflight, structurally validate, confirm, and run a project's `.smaqit/bench/` suite; report and diagnose failures |
| `smaqit.bench-scaffold` | Author a new `.smaqit/bench/` manifest for a skill/agent lacking one; delegates any live trial to `smaqit.bench-run` |

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

Root artifacts (`agents/`, `skills/`, `framework/`, `templates/`) are the source of truth. The `installer/` directory contains build intermediates that are `.gitignore`d and regenerated on every build, including the per-platform agent output produced by the generator.

**MUST NOT** manually copy files into `installer/framework/`, `installer/skills/`, `installer/agents/`, or `installer/templates/`. These are overwritten by `make prepare`.

**To build after editing root artifacts:**
```
cd installer && make build
```

The `prepare` target (run automatically by `build`) copies all root artifacts into the installer and regenerates per-platform agent output before compilation. No manual sync step is required.
