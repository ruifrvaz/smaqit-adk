# smaqit-adk

**Generic Agent Development Kit**

smaqit-adk provides the foundational framework for building AI agent orchestration systems. It's a toolkit of principles, templates, and compilation agents that enable developers to create custom specification-driven development workflows.

## What is smaqit-adk?

smaqit-adk is **not** an application - it's a development kit for building agent-based systems. It provides:

- **Framework principles** - Generic concepts for agent orchestration
- **Agent templates** - Reusable patterns for specification and implementation agents
- **Level agents** - Meta-agents that compile principles into executable agents (L0, L1, L2)
- **Compilation system** - Transform abstract principles into concrete agent implementations

## What can you build with smaqit-adk?

- Custom specification agents for any domain (security, compliance, performance, etc.)
- Implementation agents tailored to your stack and workflow
- Domain-specific agent orchestration frameworks
- Organization-specific development workflows

## Example: smaqit Product

**[smaqit](https://github.com/ruifrvaz/smaqit)** is a proof-of-concept built with smaqit-adk, demonstrating a five-layer specification system (business, functional, stack, infrastructure, coverage) with development/deployment/validation phases. It shows one way to use the ADK, but you can create entirely different architectures.

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

This creates:
```
.smaqit/
├── framework/          # 5 principle files
└── templates/
    └── agents/         # 3 generic templates + 3 compilation rules
.github/
├── agents/             # 3 Level agents (L0, L1, L2)
└── prompts/            # new-agent prompt template
```

2. **Define your custom agent requirements:**

Edit `.github/prompts/smaqit.new-agent.prompt.md` with:
- Agent purpose and scope
- Input requirements
- Output artifacts
- Validation criteria

3. **Compile your agent:**

Open GitHub Copilot chat:
```
/smaqit.L2
```

Agent-L2 will compile your custom agent from generic templates.

## Level Architecture

smaqit-adk uses a **three-level compilation chain**:

### L0: Principles (Framework Philosophy)

**Files:** `framework/*.md`

Define WHY and WHAT without implementation details:
- Core principles and concepts
- Structural mappings
- No directives, no file paths, no specifics

**Agent:** `/smaqit.L0` - Maintains principle purity

### L1: Templates (Compilation Rules)

**Files:** `templates/agents/*.template.md`, `templates/agents/compiled/*.rules.md`

Transform principles into structured directives:
- Generic agent templates with placeholders
- Compilation rules (L0 → L1 transformation)
- Directives (MUST/MUST NOT/SHOULD)

**Agent:** `/smaqit.L1` - Compiles templates from principles

### L2: Product Agents (Executable Implementations)

**Files:** Your custom agents

Compile templates into concrete agents:
- Merge template + compilation rules
- Produce executable agent definitions
- Ready for GitHub Copilot integration

**Agent:** `/smaqit.L2` - Compiles product agents from templates

## Commands

| Command | Description |
|---------|-------------|
| `smaqit-adk init [dir]` | Scaffold ADK structure |
| `smaqit-adk help` | Show detailed command help |
| `smaqit-adk uninstall` | Remove ADK from project |
| `smaqit-adk version` | Show ADK version |

## Agents

| Agent | Purpose |
|-------|---------|
| `/smaqit.L0` | Document framework principles |
| `/smaqit.L1` | Compile templates from principles |
| `/smaqit.L2` | Compile product agents from templates |

## ADK Contents

### Framework Files (5)

- `SMAQIT.md` - Core principles
- `AGENTS.md` - Agent concepts
- `TEMPLATES.md` - Template structure
- `ARTIFACTS.md` - Artifact patterns
- `PROMPTS.md` - Prompt architecture

### Agent Templates (3)

- `base-agent.template.md` - Common agent structure
- `specification-agent.template.md` - Pattern for spec-generating agents
- `implementation-agent.template.md` - Pattern for code-generating agents

### Compilation Rules (3)

- `base.rules.md` - Base agent compilation
- `specification.rules.md` - Specification agent compilation
- `implementation.rules.md` - Implementation agent compilation

### Level Agents (3)

- `smaqit.L0.agent.md` - Principle curator
- `smaqit.L1.agent.md` - Template compiler
- `smaqit.L2.agent.md` - Agent compiler

## Compatibility

| Platform | Status |
|----------|--------|
| GitHub Copilot (VS Code) | ✅ Supported |
| Other AI assistants | Planned |

## Philosophy

smaqit-adk embodies several key principles:

- **Generic by design** - No domain-specific assumptions
- **Compilation-based** - Principles → Templates → Agents
- **Self-describing** - Framework documents itself
- **Extensible** - Build your own architectures
- **Traceable** - Clear L0 → L1 → L2 lineage

## Documentation

- [Extending Agents](#) - Guide to creating custom agents
- [Framework Principles](framework/SMAQIT.md) - Core concepts
- [Compilation Chain](#) - Understanding L0/L1/L2

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

## License

MIT License - see [LICENSE](LICENSE)

## Credits

Created by [ruifrvaz](https://github.com/ruifrvaz)
