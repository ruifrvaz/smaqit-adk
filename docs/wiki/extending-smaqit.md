# Extending smaQit via smaqit-adk

This guide covers how to extend the smaQit framework by creating new agents, modifying principles, and contributing framework improvements using **smaqit-adk**.

## Overview

**smaqit-adk** (Agent Development Kit) is the framework development toolkit that contains:

- **Level agents** (L0, L1, L2) for principle curation, template compilation, and agent compilation
- **Framework files** (5 files in `framework/`) defining ADK architecture
- **Templates** (3 agent templates in `templates/agents/`) with compilation rules
- **Skills** (`skills/`) for user-facing entry points into ADK workflows

## When to Use smaqit-adk

Use **smaqit-adk** when:

- Creating custom agents for specialized domains
- Modifying framework principles or compilation rules
- Building organization-specific agent extensions
- Debugging or improving agent templates
- Scaffolding a new project with smaQit agent infrastructure

Use **smaQit** (product) when building applications with existing pre-compiled agents.

## Installation

```bash
# Install ADK
curl -fsSL https://raw.githubusercontent.com/ruifrvaz/smaqit-adk/main/install.sh | bash

# Initialize ADK project
smaqit-adk init
```

## ADK Project Structure

After `smaqit-adk init`, you'll have:

```
.smaqit/
├── framework/                    # 5 framework principle files
│   ├── SMAQIT.md                # Core principles index
│   ├── AGENTS.md                # Agent behaviors
│   ├── ARTIFACTS.md             # Artifact rules
│   ├── TEMPLATES.md             # Template structure rules
│   └── SKILLS.md                # Skill architecture
└── templates/
    └── agents/                   # 3 generic agent templates
        ├── base-agent.template.md
        ├── specification-agent.template.md
        ├── implementation-agent.template.md
        └── compiled/             # L0→L1 compilation rules
            ├── base.rules.md
            ├── specification.rules.md
            └── implementation.rules.md
.github/
├── agents/                       # 3 Level meta-agents
│   ├── smaqit.L0.agent.md       # Principle documentation
│   ├── smaqit.L1.agent.md       # Template compilation
│   └── smaqit.L2.agent.md       # Agent compilation
└── skills/
    └── smaqit.new-agent/         # Agent creation skill
        └── SKILL.md
```

## Level Agent Architecture

smaqit-adk uses a **three-level compilation chain**:

### Level 0 (L0): Principles

**Purpose:** Define framework philosophy and concepts without implementation details.

**Agent:** `/smaqit.L0`

**Content types:**
- WHY: Philosophical foundations
- WHAT: Definitions and categorizations
- HOW arranged: Structural organization

**Example:** `framework/SMAQIT.md` defines "Agents validate their own output" as a concept.

**When to use Agent-L0:**
- Documenting new framework principles
- Clarifying architectural rationale
- Defining new concepts or mappings
- Updating framework files (`framework/*.md`)
- Invoked as subagent by skills that require principle changes, or switched to directly

### Level 1 (L1): Templates

**Purpose:** Compile L0 principles into directive-based templates with structure.

**Agent:** `/smaqit.L1`

**Compilation mechanism:**
- Generic templates (`templates/agents/*.template.md`) with placeholders
- Compilation rules (`templates/agents/compiled/*.rules.md`) with L0→L1 transformation directives
- Output: Templates with directives (MUST/MUST NOT/SHOULD)

**Example:** Compilation rule transforms L0 "Agents validate output" into L1 "Agents MUST validate output before declaring completion".

**When to use Agent-L1:**
- Creating or updating agent templates (`templates/agents/*.template.md`)
- Creating or updating compilation rules (`templates/agents/compiled/*.rules.md`)
- Compiling L0 principles into structured directives
- Invoked as subagent by skills that require template changes, or switched to directly

### Level 2 (L2): Compiled Agents

**Purpose:** Compile L1 templates into concrete project agents.

**Agent:** `/smaqit.L2`

**Compilation mechanism (3-way merge):**
1. Generic agent template (`base-agent.template.md`, `specification-agent.template.md`, or `implementation-agent.template.md`)
2. Corresponding compilation rules (`base.rules.md`, `specification.rules.md`, or `implementation.rules.md`)
3. Gathered specifications — provided via context when invoked as subagent, or gathered interactively when invoked directly

**When to use Agent-L2:**
- Invoked as subagent by `smaqit.new-agent` skill after specification gathering completes
- Switched to directly by expert users for compilation work

## Creating a New Agent

### Using the smaqit.new-agent skill

1. **Invoke the skill:**

In GitHub Copilot chat, type `/smaqit.new-agent` or say *"I need to create a new agent"*.

The skill guides you interactively through specification gathering (name, description, tools, directives, scope, completion criteria, failure scenarios), then instructs the active agent to invoke Agent-L2 as a subagent to perform the 3-way compilation.

2. **Compile L1 template (if needed):**

```
/smaqit.L1
```

Agent-L1 will:
- Create or update agent template in `templates/agents/`
- Create or update compilation rules in `templates/agents/compiled/`
- Document L0 principles that informed the template

3. **Compile L2 agent:**

```
/smaqit.L2
```

Agent-L2 will:
- Read template and compilation rules
- Merge with agent creation prompt into concrete agent in `agents/` or `.github/agents/`
- Validate output structure

## Modifying Existing Agents

### Updating principles (L0)

To change framework philosophy or concepts:

```
/smaqit.L0
```

Provide context about which principle to modify and why. Agent-L0 will:
- Update relevant framework file (`framework/*.md`)
- Preserve principle purity (no directives or implementation details)
- Document rationale

### Updating templates (L1)

To change agent structure or directives:

```
/smaqit.L1
```

Provide context about which template or compilation rules to modify. Agent-L1 will:
- Update agent template or compilation rules
- Compile new directives from L0 principles
- Maintain compilation chain integrity

### Recompiling agents (L2)

After L0 or L1 changes, recompile agents:

```
/smaqit.L2
```

Agent-L2 will:
- Read updated templates and compilation rules
- Merge into concrete agent(s)
- Validate output

## Release Choreography

### ADK Release (adk-vX.Y.Z)

1. **Make framework changes:**
   - Update principles (L0), templates (L1), or compilation rules (L1)
   - Use Level agents to maintain compilation chain integrity

2. **Verify build:**
   ```bash
   cd installer && make clean build test
   ```

3. **Tag ADK release:**
   ```bash
   git tag adk-v0.2.0
   git push origin adk-v0.2.0
   ```

4. **GitHub Actions:**
   - Builds `smaqit-adk` binaries for all platforms
   - Extracts ADK section from CHANGELOG
   - Creates GitHub release with binaries

## Best Practices

### Level Boundaries

Each level has a precise content type. Applying the wrong type to a level is contamination.

| Type | Answers | Language | Lives at |
|------|---------|----------|----------|
| **Principle** | Why does this matter? | Rationale prose | L0 `framework/` |
| **Invariant** | What is always true when this principle is applied? | Declarative present-tense | L0 `framework/` |
| **Vocabulary / Catalog** | What named things exist and what do they mean? | Definitions, tables, placeholder lists | L1 `templates/agents/compiled/*.rules.md` |
| **Directive** | What must an agent do? | MUST / MUST NOT / SHOULD | L1 `templates/agents/compiled/*.rules.md` |
| **Compiled output** | Concrete executable agent | Filled template, no placeholders | L2 `agents/` or project `.github/agents/` |

**Invariant vs directive:** An invariant states what is *true* about a compliant agent in declarative present-tense. A directive instructs what an agent must *do* using MUST/MUST NOT/SHOULD. L1 reads invariants and compiles them into directives — invariant language never appears in L1 output.

**Vocabulary vs principle:** A placeholder catalog or named-things table is L1 vocabulary. It requires knowing which specific agents, layers, or placeholders exist. A principle is prior to and independent of those specifics.

#### Audit Test — 5 questions

For any content block in a framework or rules file, ask in order:

1. Does it say *why* something matters? → **Principle** (L0)
2. Does it describe what is *always true* about a compliant system? → **Invariant** (L0), stated declaratively
3. Does it list *which things exist* by name? → **Vocabulary/Catalog** (L1, in the rules file for the relevant template)
4. Does it instruct an agent what to *do*? → **Directive** (L1, MUST/MUST NOT/SHOULD)
5. Does it assume a specific product domain (layer names, phase names, file paths, CLI commands)? → **Product-domain content** — does not belong in smaqit-adk at any level

If content falls into question 5, flag it inline before removing: state "Why excluded" and "Where it belongs." Remove the flag once the content is disposed of.

### Opportunistic Cleanup

When working in contaminated areas (mixed levels), actively extract and relocate content:
1. **Don't introduce new contamination** — Respect level boundaries for new content
2. **Clean contamination within session scope** — If you spot it, fix it
3. **Document cleanup** — Note what was cleaned and where it moved
4. **Prioritize session goals** — Don't derail work, but seize opportunities

### Documentation

Document decisions in:
- **Wiki** (`docs/wiki/`) — Human-readable context and rationale
- **CHANGELOG** — User-facing changes

## Troubleshooting

### Level contamination detected

**Symptom:** L0 files contain directives or L1 files contain implementation specifics.

**Solution:**
1. Invoke appropriate Level agent (L0 or L1)
2. Extract contaminated content
3. Relocate to proper level
4. Update compilation chain if needed

### Agent compilation fails

**Symptom:** Agent-L2 can't merge template and compilation rules.

**Solution:**
1. Check template structure matches compilation rules expectations
2. Verify compilation rules have proper sections (Source Principles, Compilation Directives, Compilation Guidance)
3. Use `/smaqit.L1` to fix template or compilation rules

### Framework changes don't propagate

**Symptom:** L0 principle changes don't appear in compiled agents.

**Solution:**
1. Compile L1 templates first: `/smaqit.L1`
2. Update compilation rules if needed
3. Recompile L2 agents: `/smaqit.L2`

## Downstream Input Patterns

ADK skills do not prescribe how downstream products should handle persistent requirement storage. A product that needs users to write requirements into files should define its own input file pattern — separate from skills. Common approaches:

- A `requirements/` directory with user-managed markdown files that product agents read directly
- A `context/` or `prompts/` directory maintained by the product
- Any other user-editable file convention the product defines

The ADK's generic directive form (`MUST read from [user-defined input path]`) leaves this decision entirely to the product. Skills gather input interactively and deposit it into a designated artifact when persistence is needed — they do not double as input file locations.

## Further Reading

- [README](../../README.md) — ADK overview and quickstart
- [Framework Files](../../framework/) — Core principle definitions
- [Templates](../../templates/agents/) — Generic agent templates
- [Skill Rules](../../templates/skills/compiled/skill.rules.md) — L1 vocabulary for skill format and loading
