# Task 010: Test Framework

**Status:** In Progress
**Created:** 2026-03-03
**Updated:** 2026-03-29

## Objective

Build a three-layer test framework covering installer binary correctness, structural validation of all ADK artifacts (skills, agents, templates, rules, framework files), and behavioral evaluation of skills via LLM.

## Phases

### Phase 0 — Prerequisites ✅ *(completed 2026-03-29)*

**Fix embed bugs in `installer/main.go`:** *(resolved)*
- Changed `//go:embed skills/smaqit.new-agent/SKILL.md` → `//go:embed skills` (directory embed); added `.github/skills/smaqit.new-skill` to `cmdInit` dirs list
- Added `//go:embed templates/skills` directive; added `.smaqit/templates/skills` to `cmdInit` dirs and corresponding `copyEmbeddedDir` call

**Skill template compliance** *(resolved):* Both shipped skills (`smaqit.new-agent`, `smaqit.new-skill`) had section naming drift from `base-skill.template.md`. All 6 required sections (Purpose, Steps, Output, Scope, Completion, Failure Handling) are now present in both files.

**Makefile prepare idempotency** *(resolved):* `prepare` target now clears each destination directory with `rm -rf` before copying, preventing stale artifact drift.

---

### Phase 1 — Layer 2: Installer Unit Tests ✅ *(completed 2026-03-29)*

New files: `tests/unit/init_test.go`, `tests/unit/embed_test.go`

**Approach:** Black-box — build installer binary, invoke via `exec.Command`; no internal package imports.

**Binary location:** Tests read the binary path from `SMAQIT_ADK_BIN` env var (set by `make test`). A `TestMain` in the package fails fast with a clear error if the var is unset.

| Test | What it checks |
|------|----------------|
| `TestCmdInit` | Run binary `init` against `t.TempDir()`; assert all expected dirs and files exist (including `smaqit.new-skill`) |
| `TestCmdInit_Idempotent` | Run `init` twice; second run must not error or corrupt state |
| `TestCmdInit_AlreadyExists` | Run `init` when already initialized; assert graceful, non-destructive behavior |
| `TestCmdUninstall` | Run `init` then `uninstall`; assert target directories are removed |
| `TestCmdUninstall_NotInitialized` | Run `uninstall` without prior `init`; assert non-fatal exit |
| `TestCmdVersion` | Run `version`; assert non-empty version string on stdout |
| `TestEmbedCompleteness` | Run `init`; walk installed files; assert all expected entries present (skills, agents, framework, templates) |
| `TestEmbedContentMatchesSource` | For each installed file, read corresponding source from `../../`; assert byte-for-byte equality (catches prepare/embed drift) |

---

### Phase 2 — Layer 1: Structural Validation Tests ✅ *(completed 2026-03-29)*

New files: `tests/structural/skills_test.go`, `tests/structural/templates_test.go`, `tests/structural/agents_test.go`, `tests/structural/framework_test.go`

**Approach:** Read root artifacts directly via relative paths (`../../`) from `tests/`; no binary required.

**Implemented — skill tests** (`skills_test.go`):
- `TestSkillFrontmatter` — `name` regex `^[a-z][a-z0-9.-]*$`; `description` ≤ 1024 chars, third-person check (rejects `"I "` / `"You can"`), when-signal check (`"Use when"` or `"when the user"`)
- `TestSkillRequiredSections` — all 6 sections present (Purpose, Steps, Output, Scope, Completion, Failure Handling); code-block-aware
- `TestSkillBodyLength` — body ≤ 500 lines
- `TestSkillNoUnresolvedPlaceholders` — no `[SCREAMING_CASE]` tokens outside fenced code blocks
- `TestSkillFailureHandlingTable` — ≥4 pipe-rows (header + separator + 2 data rows)

**Implemented — template tests** (`templates_test.go`):
- `TestTemplatePlaceholdersDefinedInRules` — every `[PLACEHOLDER]` in a template is defined in its rules file(s); extension templates check both `base.rules.md` + their own rules file

**Implemented — agent tests** (`agents_test.go`):
- `TestAgentFrontmatter` — `name` present; `tools` field non-empty
- `TestAgentRequiredSections` — Role, Input, Output, Directives, Completion Criteria present; code-block-aware
- `TestAgentCompletionCriteria` — at least one `- [ ]` item; code-block-aware
- `TestAgentDirectivesHasMust` — `### MUST` subsection present; code-block-aware

**Implemented — framework tests** (`framework_test.go`):
- `TestFrameworkNoDirectiveLanguage` — no lines containing `MUST`, `MUST NOT`, or `SHOULD` keywords

**Not implemented (deferred):**
- Bidirectional template ↔ rules cross-reference (rules → template direction)
- Rules file structural checks (frontmatter fields, section presence)
- All-agents scope (currently tests all `*.agent.md`, not just `smaqit.L*`)
- `description` what-signal check (what-signal is implicitly covered by third-person + when-signal; skipped as low value)

---

### Phase 3 — Layer 3: Behavioral Evaluations

**SDK:** `github.com/github/copilot-sdk/go` (public, technical preview, active — last commit 3 days ago)

**SDK research findings (resolved 2026-03-29):**
- `ClientOptions.Cwd` — sets CLI working directory; this is the isolated workspace root
- `SessionConfig.SystemMessage` with `Mode: "replace"` — replaces the full system prompt with the artifact under test
- `OnUserInputRequest` — intercepts agent `ask_user` calls; drives scripted conversation turns from eval definitions
- `OnPermissionRequest` — controls tool permissions per eval: deny shell, allow file read/write within temp dir only
- `session.GetMessages()` — retrieves full conversation history after a run for grading
- CLI bundling via `go:embed` + `cmd/bundler` — eval runner binary ships with embedded CLI; no external install required for users
- Auth: `COPILOT_GITHUB_TOKEN` / `GH_TOKEN` / `GITHUB_TOKEN` env vars — CI-friendly
- Note: Technical Preview — breaking API changes possible; pin to a specific SDK release in `go.mod`

**Isolated workspace model:**
Each eval run creates a fresh `t.TempDir()` (or `os.MkdirTemp`) containing a standardized scaffold:
```
<eval-workspace>/
  .github/
    agents/    ← empty; agent output written here during agent evals
    skills/    ← empty; skill output written here during skill evals
```
The `ClientOptions.Cwd` is set to this temp dir. The artifact under test is injected as the full system prompt (`Mode: "replace"`). No smaqit-adk repo content, no user project files, no cross-contamination.

**Eval scope:** Both skills and compiled agents.
- **Skill evals** — inject `SKILL.md` as system prompt; drive scripted `OnUserInputRequest` turns; verify gathering flow behavior and output artifact
- **Agent evals** — inject `.agent.md` as system prompt; give the agent a task; verify it behaves per its directives

**Directory structure:**

```
tests/evals/
  skills/
    smaqit.new-agent/
      001_create_qa_agent.json
      002_create_doc_helper.json
      003_invalid_directive_form.json
    smaqit.new-skill/
      001_create_new_principle_skill.json
      002_invalid_description_person.json
  agents/
    smaqit.L2/
      001_compile_base_agent.json
      002_reject_unresolved_placeholders.json
  runner/
    main.go
  README.md
```

**Eval JSON format:**

```json
{
  "type": "skill",
  "artifact_file": "skills/smaqit.new-agent/SKILL.md",
  "description": "Verify gathering flow: agent name asked before proceeding",
  "turns": [
    { "user_input": "I need to create a QA agent for this project" },
    { "user_input": "qa-helper", "trigger": "ask_user" },
    { "user_input": "Answers questions about QA processes", "trigger": "ask_user" }
  ],
  "expected_behavior": [
    "Asks for the agent name before any other section",
    "Asks for a description under 80 characters",
    "Does not write any output file before all gathering sections are complete"
  ],
  "forbidden_behavior": [
    "Writes an agent file before completing gathering"
  ]
}
```

- `type` — `"skill"` or `"agent"`
- `artifact_file` — path to the `SKILL.md` or `.agent.md` relative to repo root; injected as system prompt via `Mode: "replace"`
- `turns` — scripted conversation; `trigger: "ask_user"` entries are fed via `OnUserInputRequest`; others via `session.Send`
- `expected_behavior` — natural language criteria graded by a second Copilot session
- `forbidden_behavior` — criteria that must NOT be present; graded as inverted pass

**Eval runner:** `tests/evals/runner/main.go`
1. For each eval JSON file in the provided `evals/` path:
   a. Create a temp dir, write standardized workspace scaffold
   b. `copilot.NewClient(&copilot.ClientOptions{Cwd: tempDir})` — start client
   c. `client.CreateSession` with `SystemMessage.Mode = "replace"`, `SystemMessage.Content = artifact file content`
   d. Register `OnUserInputRequest` that pops from the `turns` queue
   e. Register `OnPermissionRequest` that denies shell, approves read/write within temp dir
   f. Drive `turns` via `session.Send` in order
   g. Collect `session.GetMessages()` after `session.idle`
   h. Grade: start a second Copilot session; for each criterion, send "Did this conversation satisfy: [criterion]? Answer YES or NO with one sentence reason." — collect answer
2. Output per-criterion pass/fail and overall result per eval file
3. Exit non-zero if any eval fails (CI gate)

**Dependency:** `github.com/github/copilot-sdk/go` added to `tests/go.mod` (module `github.com/ruifrvaz/smaqit-adk/tests`, matching installer naming convention); SDK version pinned; isolated from installer binary.
**Auth requirement:** `COPILOT_GITHUB_TOKEN` (or `GH_TOKEN`) in environment — required for eval runner to run; runner prints a clear error if missing.

---

### Phase 4 — Makefile Wiring

Extends `installer/Makefile`. Current `test` target renamed to `test-scaffold`; three new targets added.

```makefile
REPO_ROOT = ..

# Preserved (renamed): smoke-test installation scaffold
.PHONY: test-scaffold
test-scaffold: build
	... (existing test target body)

# Layer 1 + 2: black-box unit tests + structural validation
.PHONY: test
test: build
	@SMAQIT_ADK_BIN=$(CURDIR)/dist/$(DEV_BINARY_NAME) \
	 cd $(REPO_ROOT)/tests && go test ./unit/... ./structural/...

# Layer 3: behavioral evaluations (requires COPILOT_GITHUB_TOKEN)
.PHONY: evals
evals: build
	@cd $(REPO_ROOT)/tests && go run ./evals/runner/... $(REPO_ROOT)/tests/evals/

# All layers
.PHONY: test-all
test-all: test evals
```

---

## Acceptance Criteria

- [x] Phase 0: `smaqit.new-skill` embedded and installed by `cmdInit`
- [x] Phase 1: `make test` passes with all installer unit tests green
- [x] Phase 2: `make test` passes with all structural validation tests green
- [x] Phase 2: Intentionally breaking a skill description (add "I ") causes a test failure
- [x] Phase 2: Adding `[UNRESOLVED]` to a skill body causes a test failure
- [x] Phase 2: Removing a placeholder from a template without updating the catalog causes a test failure
- [ ] Phase 3: `make evals COPILOT_GITHUB_TOKEN=...` runs without crashing and outputs per-criterion results
- [ ] Phase 3: At least 3 eval files exist for `smaqit.new-agent` and 2 for `smaqit.new-skill`
- [ ] Phase 3: At least 2 eval files exist for `smaqit.L2` (agent compilation behavior)

## Design Decisions

- **Root-level `tests/` module** (`github.com/ruifrvaz/smaqit-adk/tests`, matching installer naming convention) — separate `go.mod` isolates Copilot SDK dependency from the installer binary; root artifacts accessed via `../../` relative paths
- **Black-box testing for unit tests** — installer binary built by the test suite, invoked via `exec.Command`; no internal package imports; tests validate observable CLI behavior and filesystem outcomes only
- **Eval runner as separate binary** (`tests/evals/runner/`) — not mixed into `main.go`; runnable without building the full CLI
- **Copilot Go SDK for eval runtime** — `github.com/github/copilot-sdk/go`; SDK is public (technical preview); chosen over Anthropic API because both Task 010 and Task 011 require the same Copilot session driver; no API key duplication
- **Isolated workspace = temp dir + `ClientOptions.Cwd`** — clean room via OS temp dir; standardized `.github/` scaffold inside; the CLI sees only what we put there
- **System prompt injection via `Mode: "replace"`** — the artifact under test (SKILL.md or .agent.md) becomes the full system prompt; no Copilot CLI persona, no contaminating instructions
- **Multi-turn scripted conversations** — `OnUserInputRequest` drives `ask_user` triggers from eval JSON `turns`; enables end-to-end gathering flow validation
- **Grader = second Copilot session** — natural language criteria require semantic grading; a second session evaluates each criterion against the collected message history
- **`tests/evals/` for eval definitions** — eval definitions live inside the `tests/` module alongside the runner; co-located with what uses them; not under `installer/`
- **Both skills and agents evaluated** — skills validate gathering flows; agents validate directive compliance and behavioral boundaries
- **Framework file tests are heuristic** — MUST/SHOULD detection is reliable; "no file paths" is best-effort, not a hard gate
- **Embed bug fix is Phase 0 prerequisite** — installer unit tests must be able to assert correct post-`init` state including `smaqit.new-skill`
- **SDK version pinned** — technical preview; pin to specific release in `go.mod` to prevent silent breaking changes
