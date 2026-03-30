# Task 011: Interactive CLI Product

**Status:** In Progress
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

- [x] `smaqit-adk create-agent` runs interactively from any directory, gathers agent spec, and writes a compiled `.agent.md` into `.github/agents/` of the current project — implemented 2026-03-29
- [x] `smaqit-adk create-skill` runs interactively from any directory, gathers skill spec, and writes a compiled `SKILL.md` into `.github/skills/<skill-name>/` of the current project — implemented 2026-03-29
- [x] No framework files, templates, or Level agents are written into the user's project — guaranteed by isolation contract in `cmdCreate`
- [x] Creation runs in an isolated LLM context (only ADK framework + templates in context — no project agent instructions) — `SystemMessage.Mode=replace` with embedded ADK artifacts only
- [x] smaqit-adk can be installed globally (not per-project) — `install.sh` installs to `~/.local/bin`
- [x] Phase 0 research spike documents the chosen integration path — resolved 2026-03-29, Copilot Go SDK confirmed
- [ ] `smaqit-adk create-principle` — **deferred to Task 013** (depends on Task 006: smaqit.new-principle skill)
- [ ] `smaqit-adk validate <file>` — **deferred to Task 013** (design decision required: eval criteria for user-compiled files)

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

### Phase 1 — Architecture Decision ✓ RESOLVED (2026-03-29)

**Decisions codified:**

1. **CLI command surface:**
   - `smaqit-adk create-agent [--output <dir>]` — output defaults to `./.github/agents/`
   - `smaqit-adk create-skill [--output <dir>]` — output defaults to `./.github/skills/<name>/`
   - `smaqit-adk create-principle [--output <dir>]` — output defaults to `./.smaqit/framework/` (depends on Task 006)
   - `smaqit-adk validate <file>` — runs eval kit against a compiled agent or skill file (depends on Task 010 Phase 3)
   - `smaqit-adk init [dir]` — repurposed in Task 012; drops only two L2-compiled standalone agents
   - `smaqit-adk version`, `help`, `uninstall` — retained

2. **Module strategy — Option A:** Add `github.com/github/copilot-sdk/go v0.2.0` to `installer/go.mod`. Single module, single binary. `tests/go.mod` remains a separate test module.

3. **Context loading strategy — Option C:** System prompt = compiled L2 agent (`agents/smaqit.L2.agent.md`) + relevant skill (`skills/smaqit.new-agent/SKILL.md` or `skills/smaqit.new-skill/SKILL.md`), concatenated. No raw framework files or templates loaded. Both files embedded via `go:embed`. If gathering behavior gaps are found in testing, targeted framework files may be added — not all of them.

4. **Timeout policy — Option D:** 15-minute `context.WithTimeout` at session level. Print elapsed time or spinner while `SendAndWait` is blocking. Ctrl-C cancels cleanly via context cancellation.

5. **Output contract:** Naming conventions, file location resolution, overwrite behavior (unchanged from original spec).

6. **Isolation guarantee:** The `create-*` commands load only embedded ADK artifacts into `SystemMessage.Content`. No files from the current project's `.github/` or `.smaqit/` are read or injected into the session context.

---

### Phase 2 — Implementation ✓ RESOLVED (2026-03-29)

**`create-agent` flow:**
1. Load `agents/smaqit.L2.agent.md` + `skills/smaqit.new-agent/SKILL.md` as system context (Mode: replace)
2. Run interactive gathering loop — `OnUserInputRequest` routes agent questions to terminal stdin; user answers fed back via callback
3. Compilation step runs in same session
4. Write compiled `.agent.md` to output dir (default: `./.github/agents/`)
5. Print confirmation with path

**`create-skill` flow:**
1. Load `agents/smaqit.L2.agent.md` + `skills/smaqit.new-skill/SKILL.md` as system context (Mode: replace)
2. Run interactive gathering loop — `OnUserInputRequest` routes agent questions to terminal stdin
3. Compilation step runs in same session
4. Write compiled `SKILL.md` to output dir (default: `./.github/skills/<name>/`)
5. Print confirmation with path

**Installer changes:**
- `init` repurposed in Task 012 — drops only `smaqit.create-agent.agent.md` and `smaqit.create-skill.agent.md`; no framework, no templates, no Level agents
- Update `install.sh` and README to reflect global installation story
- Add `validate` command wired to Task 010 eval kit (Phase 3 dependency)

**Skills retained:**
- `smaqit.new-agent` and `smaqit.new-skill` remain as the Copilot-chat alternative path for users in VS Code
- No changes to skill content

---

### Phase 3 — Update Documentation and README ✓ RESOLVED (2026-03-29)

- Updated README: two-tier model (lite via `init`, advanced via CLI); global installation story; full command surface
- Updated `cmdHelp()` text for all new commands
- No deprecation note on `init` — it is repurposed (lite tier, Task 012), not deprecated

## Notes

**Phase 0 resolved 2026-03-29** — Copilot Go SDK confirmed viable. See Phase 0 section above for full findings.

**Deferred 2026-03-29** — `create-principle` and `validate` deferred to Task 013. Core task (create-agent, create-skill, isolation, global install) is complete.

**Relationship to Task 012 (Lite Tier):**
- Task 012 handles `init` repurposing and compiles the two standalone agents (`smaqit.create-agent`, `smaqit.create-skill`)
- Task 012 can proceed independently of Task 011; no shared implementation dependencies
- Task 011 `create-agent` and `create-skill` CLI commands are the isolated-context equivalents of the Task 012 agents

**Relationship to Task 010 (Test Framework):**
- Task 010 Phase 3 (eval runner) completed 2026-03-29 — functional, 1/7 evals passing on last run
- Both tasks share `github.com/github/copilot-sdk/go v0.2.0` in `installer/go.mod`
- `validate` command is NOT blocked on Task 010 Phase 3 (already done); it is blocked on an open design question: what eval criteria apply to user-compiled files (not the same as the ADK's own agent/skill evals)

**Product identity:**
- Task 012 = lite tier: VS Code users, no global CLI required, immediate value from two compiled agents installed by `init`
- Task 011 = advanced tier: developer suite, isolated context, full `create-*` + eval surface
- Both delivered by the same `smaqit-adk` binary; `init` bridges them
