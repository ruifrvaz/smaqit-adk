# smaQit Agent Development Kit

smaQit-adk is an **Agent Development Kit** for GitHub Copilot. It gives you two compiled agents — `@smaqit.create-agent` and `@smaqit.create-skill` — that let you build new Copilot agents and skills interactively, without requiring any framework files or external compilation steps in your project.

## What is smaQit-adk?

smaQit-adk ships two self-contained agents that know how to gather agent and skill specifications from you and compile them directly into your project. Install once, create agents anytime.

- **`@smaqit.create-agent`** — Interactively gathers specs and writes a `.agent.md` into `.github/agents/`
- **`@smaqit.create-skill`** — Interactively gathers specs and writes a `SKILL.md` into `.github/skills/`

Both agents carry all ADK compilation knowledge inline — no framework files, no templates, no Level agents required in your project.

## What can you build with smaQit-adk?

- Custom Copilot agents for any domain (Q&A bots, specification agents, implementation agents)
- Skills that package domain knowledge as reusable slash-command workflows
- Agent-based development workflows for your team

## Example: smaQit Product

**[smaQit](https://github.com/ruifrvaz/smaqit)** is a proof-of-concept built with smaQit-adk, demonstrating a five-layer specification system with compiled agents for each development phase.

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/ruifrvaz/smaqit-adk/main/install.sh | bash
```

Or build from source:

```bash
git clone https://github.com/ruifrvaz/smaqit-adk
cd smaqit-adk/installer
make build
./dist/smaqit-adk-dev init
```

## Quick Start

1. **Initialize ADK in your project:**

```bash
smaqit-adk init
```

This installs two agents into `.github/agents/`:
```
.github/
└── agents/
    ├── smaqit.create-agent.agent.md
    └── smaqit.create-skill.agent.md
```

That's it — no framework files, no templates, no skills directory.

2. **Create a new agent:**

Open GitHub Copilot chat and ask:
```
Create a new agent for [your purpose]
```

Copilot will invoke `@smaqit.create-agent` as a subagent to gather your specs and compile the agent file.

> **Tip — Clean context via subagent:** For the best results, let Copilot invoke these agents as subagents rather than switching to them directly. Running as a subagent provides a clean LLM context free of the current session's loaded agents, conversation history, and open file context. This is how the ADK is designed to be used.

## Creating Agents and Skills

### Create an agent

```
create a new agent for [purpose]
```
or explicitly:
```
@smaqit.create-agent
```

`smaqit.create-agent` gathers 8 specification sections interactively:
1. Identity (name, description, tools)
2. Purpose
3. Input sources
4. Output format
5. Directives (MUST / MUST NOT / SHOULD)
6. Scope boundaries
7. Completion criteria
8. Failure scenarios

Then it compiles and writes `.github/agents/[name].agent.md`.

> **Note:** `smaqit.create-agent` compiles **base agents** — agents with foundation behaviors customized for a specific purpose. For specification or implementation agents (which require ADK extension rules), use the full ADK compilation chain (see [Advanced Use](#advanced-use)).

### Create a skill

```
create a new skill for [purpose]
```
or explicitly:
```
@smaqit.create-skill
```

`smaqit.create-skill` gathers 6 specification sections:
1. Identity (name, description, version)
2. Purpose
3. Steps with fragility levels
4. Output
5. Scope
6. Failure handling

Then it compiles and writes `.github/skills/[name]/SKILL.md`.

## Commands

| Command | Description |
|---------|-------------|
| `smaqit-adk init [dir]` | Install create-agent and create-skill into `.github/agents/` |
| `smaqit-adk help` | Show detailed command help |
| `smaqit-adk uninstall` | Remove smaqit-adk agents from project |
| `smaqit-adk version` | Show ADK version |

## Agents

| Agent | Invocation | Purpose |
|-------|------------|---------| 
| `smaqit.create-agent` | `@smaqit.create-agent` or "create a new agent" | Gather specs and compile a new `.agent.md` |
| `smaqit.create-skill` | `@smaqit.create-skill` or "create a new skill" | Gather specs and compile a new `SKILL.md` |

## Compatibility

| Platform | Status |
|----------|--------|
| GitHub Copilot (VS Code) | ✅ Supported |
| Other AI assistants | Planned |

## Advanced Use

For advanced use cases — specification agents, implementation agents, framework extension, or CLI-driven compilation — the ADK ships a full compilation chain (L0 → L1 → L2), a framework of principles and templates, and two advanced-tier creation skills:

- **`smaqit.new-agent`** — Gather agent specs interactively, write a definition file to `.smaqit/definitions/`, and invoke L2 to compile the agent. Produces a full audit trail (definition file + compilation log).
- **`smaqit.new-skill`** — Same workflow for skills.

These skills require the full ADK stack at runtime (L2, framework files, templates). They are not installed by `smaqit-adk init` and are intended for ADK contributors and expert users operating the full compilation chain.

See the ADK source at `agents/`, `skills/`, `framework/`, and `templates/` for the full compilation chain.

## Philosophy

- **Self-contained agents** — No framework files needed in the consuming project
- **Compilation-based** — Principles → Templates → Agents (the compilation chain is internalized, not distributed)
- **Subagent isolation** — Clean context via subagent invocation is a first-class design goal
- **Generic by design** — No domain-specific assumptions
- **Traceable** — Clear L0 → L1 → L2 lineage (visible in the ADK source)

## License

MIT License - see [LICENSE](LICENSE)

## Credits

Created by [ruifrvaz](https://github.com/ruifrvaz)
