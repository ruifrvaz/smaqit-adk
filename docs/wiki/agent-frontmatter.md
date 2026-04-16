# Agent Frontmatter Reference

VS Code discovers `.agent.md` files in `.github/agents/` (workspace scope) or `~/.copilot/agents/` (user scope). The YAML frontmatter block controls how VS Code surfaces and restricts each agent.

## Key Fields

| Field | Type | Default | Purpose |
|---|---|---|---|
| `name` | string | filename | Agent name shown in the picker and used for subagent invocation |
| `description` | string | — | Shown as placeholder text in the chat input when the agent is active |
| `tools` | list | — | Tools available to this agent. Unlisted tools are unavailable |
| `user-invocable` | boolean | `true` | Whether the agent appears in the agents dropdown in chat |
| `disable-model-invocation` | boolean | `false` | Whether other agents can invoke this agent as a subagent |
| `model` | string or list | current model | Model to use. Array = priority fallback list |
| `agents` | list | — | Which agents this agent can invoke as subagents. Use `*` for all, `[]` for none |
| `handoffs` | list | — | Suggested next-step buttons shown after a response completes |

## Visibility vs. Invocability

These two dimensions are independent:

| Goal | Setting |
|---|---|
| Hide from picker, still usable as subagent | `user-invocable: false` |
| Show in picker, block subagent invocation | `disable-model-invocation: true` |
| Hide from picker AND block subagent invocation | both |

## ADK Level Agents

L0, L1, and L2 are set to `user-invocable: false`. They do not appear in the chat agents dropdown. They are only reachable as subagents when invoked by a skill that routes to them.

This is intentional: level agents are compilation specialists. Direct invocation from the picker bypasses the skill routing layer and skips context setup. Use the corresponding skill instead:

| Want to... | Use this skill |
|---|---|
| Add or refine a framework principle | `/smaqit.new-principle` |
| Create a new agent | `/smaqit.new-agent` |
| Create a new skill | `/smaqit.new-skill` |

Expert users who need to switch directly to a level agent can do so by typing the agent name in chat (`@smaqit.L2`) — subagent invocation is not blocked.

## Example: Subagent-Only Agent

```yaml
---
name: my-compiler
description: Compiles definition files into agents
tools: [edit, search, read]
user-invocable: false
---
```

## Example: Picker Agent with Restricted Subagent Use

```yaml
---
name: my-reviewer
description: Reviews code for security issues
tools: [read, search]
disable-model-invocation: true
---
```

## Tool Scoping

Tools not listed in `tools` are unavailable to the agent. Common tool identifiers:

| Tool | Description |
|---|---|
| `read` | File reading |
| `edit` | File editing |
| `search` | Workspace search |
| `web/fetch` | Fetch URLs |
| `agent` | Subagent invocation (required if using `agents`) |
| `todo` | Todo list management |
| `execute/runInTerminal` | Run terminal commands |

To include all tools from an MCP server: `myserver/*`

## References

- [VS Code Custom Agents documentation](https://code.visualstudio.com/docs/copilot/customization/custom-agents)
- [VS Code Agent Skills documentation](https://code.visualstudio.com/docs/copilot/customization/agent-skills)
