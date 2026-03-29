# Task 011: Interactive CLI Product

**Status:** Not Started
**Created:** 2026-03-29

## Description

**Advanced Tier** — the full smaqit-adk developer suite as a globally installed, standalone interactive CLI. This task covers the Go CLI (Copilot SDK) implementation of the complete `create-*` command surface, Level agents accessible as CLI modes, framework files shipped with the binary, and the eval kit for validating compiled outputs. The lite tier (`init` + two compiled standalone agents) is handled in Task 012.

Users install smaqit-adk once. The CLI drives each interactive workflow in a fresh, isolated LLM context — no VS Code, no Copilot extension required, no project context contamination.

**Problem with current model:**
- `smaqit-adk init` requires per-project installation of framework files, templates, and Level agents
- Creation workflows live inside Copilot chat — VS Code + Copilot extension required
- Existing agent instructions and session context in a project can contaminate the creation process with conflicting rules
- Every project carries ADK boilerplate even after the agent/skill is created

**Target model (advanced tier):**
- smaqit-adk is installed globally once
- User runs `smaqit-adk create-agent`, `create-skill`, or `create-principle` from any project directory
- CLI opens a fresh, isolated LLM context with only the ADK framework and templates loaded
- CLI drives the interactive gathering and compilation flow
- CLI writes only the compiled output file into the user's project
- `validate` command runs the eval kit against a compiled agent or skill file
- `init` repurposed in Task 012 — drops only two L2-compiled standalone agents, no boilerplate

## Acceptance Criteria

- [ ] `smaqit-adk create-agent` runs interactively from any directory, gathers agent spec, and writes a compiled `.agent.md` into `.github/agents/` of the current project
- [ ] `smaqit-adk create-skill` runs interactively from any directory, gathers skill spec, and writes a compiled `SKILL.md` into `.github/skills/<skill-name>/` of the current project
- [ ] No framework files, templates, or Level agents are written into the user's project
- [ ] Creation runs in an isolated LLM context (only ADK framework + templates in context — no project agent instructions)
- [ ] smaqit-adk can be installed globally (not per-project)
- [x] Phase 0 research spike documents the chosen integration path — resolved 2026-03-29, Copilot Go SDK confirmed
- [ ] `smaqit-adk create-principle` runs interactively and writes a principle file into `.smaqit/framework/` (depends on Task 006)
- [ ] `smaqit-adk validate <file>` runs the eval kit against a compiled output file (depends on Task 010 Phase 3)

## Phases

### Phase 0 — Research Spike: Integration Path ✓ RESOLVED (2026-03-29)

**Chosen path:** `github.com/github/copilot-sdk/go` — `go get github.com/github/copilot-sdk/go`

**Viability confirmed:**
- Public repository, MIT license, active development (commit 3 days ago as of research date)
- Go SDK with `NewClient` → `CreateSession` → `Send` pattern — direct programmatic session driver
- `SystemMessage.Mode = "replace"` — load ADK framework + templates as standalone system prompt; no Copilot CLI persona contamination
- `OnUserInputRequest` — intercept agent questions; enables interactive terminal I/O routing between user and agent
- `ClientOptions.Cwd` — set working directory per invocation; output files written to user project dir
- `OnPermissionRequest` — control which tools the agent can invoke; lock down destructively for non-interactive contexts
- CLI embedding via `go:embed` + `cmd/bundler` — binary ships with Copilot CLI; no external install required
- Auth: `COPILOT_GITHUB_TOKEN` / `GH_TOKEN` env vars; also supports stored OAuth credentials from `copilot` CLI login

**Technical Preview caveat:** API may change in breaking ways. Pin SDK version in `go.mod`. Monitor release notes.

**Shared dependency with Task 010:** The same `github.com/github/copilot-sdk/go` package is used by the Task 010 eval runner. Both tasks should use the same pinned version in `installer/go.mod`.

---

### Phase 1 — Architecture Decision

**Decisions to codify before writing code:**

1. **CLI command surface:**
   - `smaqit-adk create-agent [--output <dir>]` — output defaults to `./.github/agents/`
   - `smaqit-adk create-skill [--output <dir>]` — output defaults to `./.github/skills/<name>/`
   - `smaqit-adk create-principle [--output <dir>]` — output defaults to `./.smaqit/framework/` (depends on Task 006)
   - `smaqit-adk validate <file>` — runs eval kit against a compiled agent or skill file (depends on Task 010 Phase 3)
   - `smaqit-adk init [dir]` — repurposed in Task 012; drops only two L2-compiled standalone agents
   - `smaqit-adk version`, `help`, `uninstall` — retained

2. **Context loading strategy:** How ADK framework + templates are injected as system context for the LLM session (embedded files read at startup, loaded as messages or system prompt)

3. **Output contract:** Naming conventions, file location resolution, overwrite behavior

4. **Isolation guarantee:** The create-* commands must not read or inject any files from the current project's `.github/` or `.smaqit/` into the LLM context

---

### Phase 2 — Implementation

**`create-agent` flow:**
1. Load ADK framework + templates into isolated LLM context
2. Run interactive gathering loop (equivalent to `smaqit.new-agent` skill: purpose, role, input, output, directives, tools)
3. Invoke L2 compilation step within same context
4. Write compiled `.agent.md` to output dir
5. Print confirmation with path

**`create-skill` flow:**
1. Load ADK framework + templates into isolated LLM context
2. Run interactive gathering loop (equivalent to `smaqit.new-skill` skill: name, description, steps, fragility, output, scope)
3. Invoke L2 compilation step within same context
4. Write compiled `SKILL.md` to output dir
5. Print confirmation with path

**Installer changes:**
- `init` repurposed in Task 012 — drops only `smaqit.create-agent.agent.md` and `smaqit.create-skill.agent.md`; no framework, no templates, no Level agents
- Update `install.sh` and README to reflect global installation story
- Add `validate` command wired to Task 010 eval kit (Phase 3 dependency)

**Skills retained:**
- `smaqit.new-agent` and `smaqit.new-skill` remain as the Copilot-chat alternative path for users in VS Code
- No changes to skill content

---

### Phase 3 — Update Documentation and README

- Update README: two-tier model (lite via `init`, advanced via CLI); global installation story; full command surface
- Update `cmdHelp()` text for all new commands
- Update `install.sh` global install steps
- No deprecation note on `init` — it is repurposed (lite tier, Task 012), not deprecated

## Notes

**Phase 0 resolved 2026-03-29** — Copilot Go SDK confirmed viable. See Phase 0 section above for full findings.

**Relationship to Task 012 (Lite Tier):**
- Task 012 handles `init` repurposing and compiles the two standalone agents (`smaqit.create-agent`, `smaqit.create-skill`)
- Task 012 can proceed independently of Task 011; no shared implementation dependencies
- Task 011 `create-agent` and `create-skill` CLI commands are the isolated-context equivalents of the Task 012 agents

**Relationship to Task 010 (Test Framework):**
- Task 010 Phase 3 (eval runner) uses the same `github.com/github/copilot-sdk/go` package
- Both tasks share the same SDK dependency in `installer/go.mod` — coordinate version pinning
- Task 010 Phase 3 is no longer blocked on the SDK research spike; it can proceed in parallel with Task 011 Phase 1
- Suggested order: Task 010 Phases 0–2 (no SDK required) can proceed independently; Task 010 Phase 3 and Task 011 Phase 1+ can proceed in parallel once the SDK dependency is added to the module

**Product identity:**
- Task 012 = lite tier: VS Code users, no global CLI required, immediate value from two compiled agents installed by `init`
- Task 011 = advanced tier: developer suite, isolated context, full `create-*` + eval surface
- Both delivered by the same `smaqit-adk` binary; `init` bridges them
