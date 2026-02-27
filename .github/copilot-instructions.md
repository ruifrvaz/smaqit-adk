# GitHub Copilot Instructions — smaqit-adk

## What this repository is

This is the **smaqit-adk** repository — an Agent Development Kit. It ships a compilation framework (principles → templates → agents) and the tools to use it.

## ADK Artifacts (in scope)

When reasoning about or modifying the ADK, work exclusively with root-level artifacts:

| Path | Purpose |
|------|---------|
| `agents/` | Shipped Level agents (L0, L1, L2) |
| `skills/` | Shipped skills (e.g., `smaqit.new-agent`) |
| `framework/` | L0 principle files |
| `templates/` | L1 agent templates and compilation rules |
| `installer/` | Go CLI that packages and distributes the above |
| `docs/` | ADK documentation |

## `.github/` contents are NOT ADK artifacts

The `.github/agents/` and `.github/skills/` directories in this repository contain **smaqit product extensions** — agents and skills built using the ADK by the smaQit product. They are a consumer of the ADK, not part of it.

# Directives

**MUST NOT:**
- Use `.github/agents/` or `.github/skills/` as examples of ADK design
- Treat smaqit extension skills (session, task, release, user-testing) as ADK-owned
- Assume the ADK ships or controls any content under `.github/`

**The only exception:** `.github/workflows/` (CI/CD) and `.github/copilot-instructions.md` (this file) are ADK project infrastructure.

## Architecture mental model

```
smaqit-adk (this repo, root)
├── agents/          ← ADK ships these (L0, L1, L2)
├── skills/          ← ADK ships these (smaqit.new-agent, ...)
├── framework/       ← ADK ships these (principle files)
├── templates/       ← ADK ships these (compilation templates)
└── installer/       ← packages all of the above into a binary

.github/             ← NOT shipped by ADK
├── agents/          ← may contain smaqit product agents (external)
└── skills/          ← may contain smaqit product skills (external)
```

When a project runs `smaqit-adk init`, it receives copies of `agents/`, `skills/`, `framework/`, and `templates/` into its own `.github/` and `.smaqit/`. The `.github/` content in this development repo is not that — it is smaqit product work that happens to live here.

## Routing between Level agents

The ADK has three Level agents with distinct responsibilities:

| Agent | Responsibility | Invoke when |
|-------|---------------|-------------|
| `smaqit.L0` | Maintain framework principles | Changing `framework/*.md` |
| `smaqit.L1` | Compile principles to templates | Changing `templates/` |
| `smaqit.L2` | Compile templates to product agents | Creating new agents |

Skills that route to a Level agent do so by naming it explicitly in their compilation step — not via frontmatter (which is an industry format not to be extended).
