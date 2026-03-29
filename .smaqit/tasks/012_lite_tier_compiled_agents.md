# Task 012: Lite Tier — Compiled Standalone Create Agents

**Status:** Complete
**Created:** 2026-03-29

## Description

Compile two standalone ADK-shipped agents via L2 and repurpose `smaqit-adk init` to install only these two files. This is the **lite tier** of the smaqit-adk product: users without a global CLI installation get immediate agent and skill creation in VS Code Copilot chat by running `smaqit-adk init` — no framework files, no templates, no Level agents dropped into their project.

The two agents are **self-contained**: they carry the full gathering + compilation behavior internally, without delegating to L2 at runtime. L2 is used once at compile time (by an ADK contributor) to produce the agents; end users never interact with L2.

**The self-compilation thesis:** smaqit-adk's core deliverable is itself compiled using its own pipeline. The agents shipped in `agents/` are the output of L2 running against the ADK's own framework and templates.

**Context isolation via subagent invocation:** Users are strongly guided to invoke these agents as subagents (not by switching agent context directly). Running as a subagent gives the agent a clean context — the parent session's conversation history, loaded agents, and open file context do not bleed in. This is the same isolation pattern the ADK uses internally (L2 runs as subagent when invoked by skills). The README and agent descriptions steer users toward subagent invocation explicitly. For users who need process-level isolation independent of VS Code, the advanced tier CLI (Task 011) remains the correct path.

## Acceptance Criteria

- [x] `smaqit.create-agent.agent.md` exists in `agents/`, compiled by L2, self-contained (no runtime L2 dependency)
- [x] `smaqit.create-skill.agent.md` exists in `agents/`, compiled by L2, self-contained (no runtime L2 dependency)
- [x] `smaqit-adk init` drops only `smaqit.create-agent.agent.md` and `smaqit.create-skill.agent.md` into `.github/agents/` of the target project
- [x] No framework files, templates, Level agents, or skills are written by `init`
- [x] `smaqit-adk uninstall` removes only the two ADK-installed agent files by name; does not delete user-created agents in `.github/agents/`
- [x] README documents the lite tier model and the subagent invocation pattern as the recommended usage

## Phases

### Phase 1 — Agent Specification ✓ COMPLETE (2026-03-29)

Definition files written to:
- `.smaqit/definitions/agents/smaqit.create-agent.md`
- `.smaqit/definitions/agents/smaqit.create-skill.md`

Both definitions inline:
- The full 8-section (agent) / 6-section (skill) gathering interview
- All base foundation directives verbatim (MUST/MUST NOT/SHOULD from `base.rules.md`)
- Inline compilation logic (no runtime L2 call)
- Fragility-based step writing rules (create-skill)
- Conciseness filter rules (create-skill)
- Base failure handling rows (create-skill)
- Subagent invocation guidance in `description` field (discovery-stage signal)

**Resolution of the "how much verbatim" open question:** All base foundation directives are listed verbatim in the compilation MUST directives. The agents do not rely on any runtime framework file presence; everything needed to produce a spec-compliant output is encoded in the definition.

---

### Phase 2 — L2 Compilation (Manual Release Step)

Run an L2 compilation session in VS Code Copilot to produce both agents. This is a **manual release step** — it requires a live Copilot session and cannot be automated in `make build`.

Steps:
1. Switch to `@smaqit.L2` agent context in VS Code Copilot
2. Provide the Phase 1 specification for `smaqit.create-agent`
3. L2 reads framework + templates and compiles the agent
4. Review output: self-containment, gathering completeness, output format compliance, no L2 runtime dependency
5. Repeat for `smaqit.create-skill`
6. Commit both compiled agents to `agents/` as ADK artifacts

**Build pipeline note:** `make prepare` copies `agents/*.agent.md` into `installer/agents/`. The existing `//go:embed agents/*.md` directive in `main.go` picks them up automatically. No Makefile or embed changes required — only `cmdInit` logic must change (Phase 3).

---

### Phase 3 — Simplify `init` in Go CLI

Update `installer/main.go`:

**`cmdInit` changes:**
- Drop only `smaqit.create-agent.agent.md` and `smaqit.create-skill.agent.md` from the embedded `agents/` FS
- Remove all framework, template, and skill file copying
- Directory creation: only `.github/agents/`; remove `.smaqit/`, `.smaqit/framework/`, `.smaqit/templates/`, `.github/skills/`
- Update success message: list the two installed agents, explain how to invoke them in Copilot chat

**`cmdUninstall` changes:**
- Remove by file name (`smaqit.create-agent.agent.md`, `smaqit.create-skill.agent.md`) rather than wiping `.github/agents/` entirely
- Preserve any user-created agents in `.github/agents/`
- Remove all `.smaqit/` cleanup (no longer installed by `init`)
- Remove `.github/skills/` cleanup (no longer installed by `init`)

**`cmdHelp` and `printUsage` changes:**
- `init` described as: "Install smaqit.create-agent and smaqit.create-skill into your project"
- Remove references to per-project scaffolding, framework files, and Level agents

---

### Phase 4 — Update README and Documentation ✓ COMPLETE (2026-03-29)

- Update README: lite tier model; what `init` installs; VS Code Copilot usage; context-isolation tradeoff note; pointer to advanced tier (Task 011)
- Update `install.sh` if needed (global install story unchanged)

## Notes

**Self-contained vs. subagent-delegating:**
The current `smaqit.new-agent` skill delegates compilation to L2 as a subagent. These lite tier agents cannot do that — L2 is not in the user's environment. The compiled agents must internalize all of the compilation behavior that L2 would otherwise provide. This makes Phase 1 specification more demanding: it must tell L2 exactly what to bake in.

**Build pipeline compatibility:**
No Makefile changes needed. The two compiled agents go into `agents/` (ADK root). `make prepare` copies them to `installer/agents/`. The existing embed directive captures them. Only `cmdInit` logic changes.

**Relationship to existing skills:**
`smaqit.new-agent` and `smaqit.new-skill` remain unchanged — they are the skill invocation path (slash command). Three paths exist to the same outcome:
1. `/smaqit.new-agent` skill (requires skills in project; delegates to L2 at runtime)
2. `@smaqit.create-agent` agent invoked as subagent (lite tier; self-contained; clean context; requires only `init`)
3. `smaqit-adk create-agent` CLI command (advanced tier; process-level isolated context; Task 011)

**Subagent invocation as the recommended pattern:**
The `smaqit.create-agent` and `smaqit.create-skill` agents should be invoked as subagents, not switched to directly. Running as a subagent provides a clean LLM context — no inherited conversation history, no other loaded agents, no open file context from the parent session. This is the same mechanism L2 uses when invoked by skills. The agent descriptions and README must make this explicit. The practical instruction to users: ask Copilot to "create a new agent" and let the active agent invoke `smaqit.create-agent` as a subagent, rather than switching to it manually.

**Relationship to Task 011 (Advanced Tier):**
These tasks are independent. Task 012 can ship before Task 011. Task 011's `create-agent` and `create-skill` CLI commands replicate the same workflows in an isolated context — they are not blocked on Task 012 and do not consume its agents.
