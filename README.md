# smaQit Agent Development Kit

smaQit-adk is an **Agent Development Kit**. It ships a compilation framework (principles → templates → agents) and the tools to create custom agents and skills — installed once, globally, and available in every project you work in.

## Compatibility

| Platform | Support |
|----------|---------|
| Claude Code | Compiled `.md` subagents at `~/.claude/agents/`; skills at `~/.claude/skills/`. Native subagent invocation. |
| Codex CLI | Compiled `.toml` custom agents at `~/.codex/agents/`; skills at `~/.agents/skills/`. Native named-agent spawning. |
| GitHub Copilot | `AGENTS.md` (root, canonical instructions) and `~/.agents/skills/` — both accepted standards Copilot already reads natively. Routing skills fall back to reading the Claude-format agent body directly when invoking L0/L1/L2. |

## What is smaQit-adk?

A single global install puts:

- 3 Level agents (`smaqit.L0`, `smaqit.L1`, `smaqit.L2`) into `~/.claude/agents/` and `~/.codex/agents/`
- 5 skills (`smaqit.create-agent`, `smaqit.create-skill`, `smaqit.new-principle`, `smaqit.bench-run`, `smaqit.bench-scaffold`) into `~/.agents/skills/` and `~/.claude/skills/`
- Compilation templates and framework principle files into `~/.agents/smaqit-adk/`

No per-project install step, no tiers. Say "create a new agent" (or skill) in Claude Code or Codex, in any project, and the routing skill takes it from there.

## What can you build with smaQit-adk?

- Custom agents for any domain (Q&A bots, specification agents, implementation agents), compiled for Claude Code and Codex CLI
- Skills that package domain knowledge as reusable slash-command workflows
- Agent-based development workflows for your team

## Example: smaQit Product

**[smaQit](https://github.com/ruifrvaz/smaqit)** is a proof-of-concept built with smaQit-adk, demonstrating a five-layer specification system with compiled agents for each development phase.

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/ruifrvaz/smaqit-adk/main/install.sh | bash
```

This downloads the `smaqit-adk` binary to `~/.local/bin/` and immediately installs agents, skills, templates, and framework files to their global locations. Nothing is written into your project directory.

Or build from source:

```bash
git clone https://github.com/ruifrvaz/smaqit-adk
cd smaqit-adk/installer
make build
```

Product implementation lives under `src/`. The `installer/` directory is the Make-driven packaging and embedded-artifact staging boundary; it delegates benchmark commands to the source module.

## Quick Start

### Evaluate or compare an agent harness

Define cases, visible inputs, expected outputs, and one or more local harness variants in YAML, then validate, plan, and run:

```bash
smaqit-adk bench validate examples/bench/single-eval/bench.yaml
smaqit-adk bench plan examples/bench/single-eval/bench.yaml
smaqit-adk bench run examples/bench/single-eval/bench.yaml
```

One variant produces an evaluation; multiple variants produce a controlled comparison. `bench run` displays lifecycle progress by default; agents and applications can use `-events jsonl` for ordered machine-readable state updates. See [Benchmarking and evaluation](docs/wiki/benchmarking.md) for the manifest, process adapter, scoring, artifacts, security model, and application-facing interfaces.

## Creating Agents and Skills

### Create an agent

```
create a new agent for [purpose]
```
or explicitly:
```
/smaqit.create-agent
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

Then it compiles and writes both a Claude Code `.md` subagent and a Codex CLI `.toml` custom agent.

> **Note:** `smaqit.create-agent` compiles **base agents** — agents with foundation behaviors customized for a specific purpose. For specification or implementation agents (which require ADK extension rules), use the full ADK compilation chain (see [ADK Source](#adk-source-expert-use)).

### Create a skill

```
create a new skill for [purpose]
```
or explicitly:
```
/smaqit.create-skill
```

`smaqit.create-skill` gathers 5 specification sections:
1. Identity (name, description, version)
2. Steps with fragility levels
3. Output
4. Scope
5. Failure handling

Then it compiles and writes `SKILL.md` — a single, cross-platform format already compatible with Claude Code, Codex CLI, and GitHub Copilot, with no per-platform variants needed.

## Commands

| Command | Description |
|---------|-------------|
| `smaqit-adk bench <validate\|plan\|run\|grade\|compare\|report>` | Run config-first local evaluations and comparisons |
| `smaqit-adk help` | Show detailed command help |
| `smaqit-adk uninstall` | Remove smaqit-adk's global agents, skills, templates, and framework files |
| `smaqit-adk version` | Show ADK version |

Global installation has no user-facing subcommand — it happens automatically when `install.sh` runs.

## Agents and Skills

| Artifact | Invocation | Purpose |
|----------|------------|---------|
| `smaqit.create-agent` (skill) | "create a new agent" or `/smaqit.create-agent` | Routes to `smaqit.L2` |
| `smaqit.create-skill` (skill) | "create a new skill" or `/smaqit.create-skill` | Routes to `smaqit.L2` |
| `smaqit.new-principle` (skill) | "add a principle" or `/smaqit.new-principle` | Routes to `smaqit.L0` |
| `smaqit.L0` / `smaqit.L1` / `smaqit.L2` (agents) | Invoked as subagents by the above skills | Maintain the framework, compile templates, compile agents and skills |

On Claude Code and Codex CLI, routing is a native subagent/custom-agent call against the compiled file. On GitHub Copilot, which has no dedicated compiled agent file, the routing skill instructs it to read the Claude-format agent body directly and follow it inline for the current turn — same content, no native subagent context isolation.

## ADK Source (Expert Use)

For framework extension, specification agents, implementation agents, or direct compilation chain access, contributors work with the ADK's own source directly:

1. Write a definition file to `.smaqit/definitions/agents/[name].md` or `.smaqit/definitions/skills/[name].md` — the same shape `smaqit.create-agent`/`create-skill` would produce
2. Invoke `smaqit.L2` as a subagent to compile it

See the ADK source at `agents/`, `skills/`, `framework/`, and `templates/` for the full L0 → L1 → L2 compilation chain.

## Philosophy

- **Globally installed, project-agnostic** — one install, every project
- **Compilation-based** — Principles → Templates → Agents (the compilation chain is internalized, not distributed)
- **Multi-platform by construction** — Claude Code and Codex CLI compiled natively; GitHub Copilot compatible via accepted standards (`AGENTS.md`, `~/.agents/skills/`)
- **Subagent isolation** — Clean context via subagent invocation is a first-class design goal
- **Generic by design** — No domain-specific assumptions
- **Traceable** — Clear L0 → L1 → L2 lineage (visible in the ADK source)

## License

MIT License - see [LICENSE](LICENSE)

## Credits

Created by [ruifrvaz](https://github.com/ruifrvaz)
