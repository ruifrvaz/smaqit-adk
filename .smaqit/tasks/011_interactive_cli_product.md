# Task 011: Interactive CLI Product

**Status:** Not Started
**Created:** 2026-03-29

## Description

Redesign smaqit-adk from a per-project scaffolder into a globally installed, standalone interactive CLI. Users install smaqit-adk once and invoke it from any project whenever they need to create a new agent or skill. The CLI drives the interactive creation flow using the Copilot API and writes only the compiled output file into the user's project. No framework boilerplate is dropped into the project.

**Problem with current model:**
- `smaqit-adk init` requires per-project installation of framework files, templates, and Level agents
- Creation workflows live inside Copilot chat — VS Code + Copilot extension required
- Existing agent instructions and session context in a project can contaminate the creation process with conflicting rules
- Every project carries ADK boilerplate even after the agent/skill is created

**Target model:**
- smaqit-adk is installed globally once
- User runs `smaqit-adk create-agent` or `smaqit-adk create-skill` from any project directory
- CLI opens a fresh, isolated LLM context with only the ADK framework and templates loaded
- CLI drives the interactive gathering and compilation flow (equivalent to the current skill flow)
- CLI writes only the compiled output (`*.agent.md` or `SKILL.md`) into the user's project
- User closes the CLI and continues working — no residual ADK files in the project

## Acceptance Criteria

- [ ] `smaqit-adk create-agent` runs interactively from any directory, gathers agent spec, and writes a compiled `.agent.md` into `.github/agents/` of the current project
- [ ] `smaqit-adk create-skill` runs interactively from any directory, gathers skill spec, and writes a compiled `SKILL.md` into `.github/skills/<skill-name>/` of the current project
- [ ] No framework files, templates, or Level agents are written into the user's project
- [ ] Creation runs in an isolated LLM context (only ADK framework + templates in context — no project agent instructions)
- [ ] smaqit-adk can be installed globally (not per-project)
- [ ] Phase 0 research spike documents the chosen integration path (copilot-sdk, copilot-cli, or alternative) and its auth model

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
   - `smaqit-adk init [dir]` — fate TBD (retained, removed, or repurposed — open decision for Phase 1)
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
3. Invoke L1 compilation step within same context
4. Write compiled `SKILL.md` to output dir
5. Print confirmation with path

**Installer changes:**
- `init` fate is an open decision — may be retained, removed, or repurposed depending on Phase 1 architecture conclusions
- Update `install.sh` and README to reflect global installation story

**Skills retained:**
- `smaqit.new-agent` and `smaqit.new-skill` remain as the Copilot-chat alternative path for users in VS Code
- No changes to skill content

---

### Phase 3 — Update Documentation and README

- Update README: global installation story, new command surface, no per-project init required
- Update `cmdHelp()` text
- Update `install.sh` global install steps
- Deprecation note on `init` as boilerplate scaffolder

## Notes

**Phase 0 resolved 2026-03-29** — Copilot Go SDK confirmed viable. See Phase 0 section above for full findings.

**Relationship to Task 010 (Test Framework):**
- Task 010 Phase 3 (eval runner) uses the same `github.com/github/copilot-sdk/go` package
- Both tasks share the same SDK dependency in `installer/go.mod` — coordinate version pinning
- Task 010 Phase 3 is no longer blocked on the SDK research spike; it can proceed in parallel with Task 011 Phase 1
- Suggested order: Task 010 Phases 0–2 (no SDK required) can proceed independently; Task 010 Phase 3 and Task 011 Phase 1+ can proceed in parallel once the SDK dependency is added to the module
