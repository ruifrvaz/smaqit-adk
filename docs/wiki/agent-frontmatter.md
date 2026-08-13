# Agent Metadata Reference

smaqit-adk compiles each agent from one shared, platform-neutral body (Role/Input/Output/Directives prose) plus per-platform metadata that controls the compiled file's wrapper — frontmatter shape, tool declarations, and discovery behavior differ by platform even though the body content is identical.

## Claude Code — YAML frontmatter (`.md`)

Claude Code discovers agents at `.claude/agents/` (project scope) or `~/.claude/agents/` (user scope, respecting `CLAUDE_CONFIG_DIR` if set).

| Field | Type | Purpose |
|---|---|---|
| `name` | string | Agent name used for subagent invocation (Task tool) |
| `description` | string | Shown when the agent is selected or listed |
| `tools` | comma-separated list | Explicit tool allowlist. **Omitted entirely → the subagent inherits all tools** — this is not documented as a default in Claude's own frontmatter, but confirmed behavior: absence of the field means unrestricted, not empty |

```yaml
---
name: smaqit-L2
description: Level 2 Compiler - Compiles Level 1 template directives and definition files into concrete agent and skill implementations
tools: Bash, Read, Write, Edit, Grep, Glob, Task, TodoWrite
---
```

## Codex CLI — TOML custom agents (`.toml`)

Codex CLI discovers agents at `.codex/agents/` (project scope) or `~/.codex/agents/` (user scope, respecting `CODEX_HOME` if set). The `name` field is the identity Codex matches on when spawning — filename and `name` should agree by convention.

| Field | Type | Purpose |
|---|---|---|
| `name` | string | Matched when Codex spawns the agent |
| `description` | string | Shown when the agent is selected or listed |
| `developer_instructions` | TOML literal multi-line string | The agent body — identical content to the Claude render |
| `[tools]` | table | Optional. Grants *additional* capabilities beyond Codex's default core tools (e.g. `view_image = true`) — not a restrict-to-list mechanism like Claude's `tools:`. Omit unless the agent genuinely needs something extra |
| `[mcp_servers.*]` | table | Optional MCP server bindings |

```toml
name = "smaqit.L2"
description = "Level 2 Compiler - Compiles Level 1 template directives and definition files into concrete agent and skill implementations"
developer_instructions = '''
# Level 2: Agent and Skill Compiler
...
'''
```

## GitHub Copilot — no dedicated compiled file

Copilot is not an authored target: `smaqit.L0`/`L1`/`L2` have no `.agent.md` render, and `smaqit.create-agent` does not produce one for user-created agents either. Copilot compatibility comes from two accepted standards it already reads natively:

- **`AGENTS.md`** (root, canonical instructions) — read via `chat.useAgentsMdFile`
- **`~/.agents/skills/`** — Copilot auto-discovers personal skills here, the same shared path Codex CLI uses

When a skill needs to invoke a Level agent on Copilot, it reads the Claude-format body directly (`~/.claude/agents/smaqit-L0.md` etc.) and follows it inline for the current turn — same content, no native subagent context isolation.

## ADK Level Agents

L0, L1, and L2 are compilation specialists, not primary user-facing entry points. Use the corresponding skill instead of invoking them directly:

| Want to... | Use this skill |
|---|---|
| Add or refine a framework principle | `/smaqit.new-principle` |
| Create a new agent | `/smaqit.create-agent` |
| Create a new skill | `/smaqit.create-skill` |
| Update a template or compilation rules | Switch to `smaqit.L1` directly — no routing skill exists for this yet |

Expert users who need to switch directly to a Level agent can do so via their platform's own subagent-switching mechanism (e.g. `@smaqit-L2` in Claude Code, spawning `smaqit.L2` by name in Codex CLI).

## Tool Name Mapping (Copilot-era → Claude)

The three Level agents originally declared Copilot tool IDs before this multi-platform migration; the explicit mapping used to produce each agent's Claude `tools:` list:

| Copilot tool ID | Claude tool name |
|---|---|
| `execute/runInTerminal`, `execute/getTerminalOutput`, `execute/awaitTerminal` | `Bash` |
| `read/readFile` | `Read` |
| `edit/createFile`, `edit/editFiles` | `Write`, `Edit` |
| `edit/createDirectory` | (implicit in `Write`/`Bash`) |
| `search` | `Grep`, `Glob` |
| `agent` | `Task` |
| `todo` | `TodoWrite` |
| `web/fetch` | `WebFetch` |

## References

- [Claude Code: Create custom subagents](https://code.claude.com/docs/en/sub-agents)
- [Codex CLI: Custom agent definitions (TOML)](https://learn.chatgpt.com/docs/codex/cli)
- [VS Code Copilot: `AGENTS.md` support](https://code.visualstudio.com/docs/agent-customization/custom-instructions)
