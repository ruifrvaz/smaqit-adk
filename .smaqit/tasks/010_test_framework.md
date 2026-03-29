# Task 010: Test Framework

**Status:** Not Started
**Created:** 2026-03-03

## Objective

Build a three-layer test framework covering installer binary correctness, structural validation of all ADK artifacts (skills, agents, templates, rules, framework files), and behavioral evaluation of skills via LLM.

## Phases

### Phase 0 — Fix Embed Bug (prerequisite)

`installer/main.go` uses a single-file `go:embed` that only captures `skills/smaqit.new-agent/SKILL.md`. The `smaqit.new-skill` skill is copied by `make prepare` but never embedded in the binary, and `cmdInit` never installs it.

**Changes:**
- `installer/main.go` — change embed directive from single file to `//go:embed skills` (embed the directory); add `.github/skills/smaqit.new-skill` entry to `cmdInit` dirs/files list
- Must be done before writing installer unit tests that assert correct post-`init` state

---

### Phase 1 — Layer 2: Installer Unit Tests

New file: `installer/main_test.go`

| Test | What it checks |
|------|----------------|
| `TestCmdInit` | Run `cmdInit` against `t.TempDir()`; assert all expected dirs and files exist (including `smaqit.new-skill`) |
| `TestCmdInit_Idempotent` | Run `cmdInit` twice; second run must not error or corrupt state |
| `TestCmdUninstall` | Run `cmdInit` then `cmdUninstall`; assert target directories are removed |
| `TestEmbedCompleteness` | Iterate embedded FS; cross-check against Makefile `prepare` target's intent (skills, agents, framework, templates all present) |
| `TestEmbedContentMatchesSource` | For each embedded file, read corresponding source from `../`; assert byte-for-byte equality (catches prepare/embed drift) |

---

### Phase 2 — Layer 1: Structural Validation Tests

New file: `installer/structural_test.go`

All tests use `os.ReadFile` and path walks on `../` (root artifacts). No LLM required.

**Skill tests** (all `skills/*/SKILL.md`):
- Frontmatter parses as valid YAML
- `name` ≤ 64 chars, matches `^[a-z][a-z0-9-]*$`, no reserved words ("anthropic", "claude")
- `description` ≤ 1024 chars, non-empty, written in third person (reject "I " / "You can" at sentence start)
- `description` contains what-signal and when-signal (keyword: "Use when" or "when the user")
- Body ≤ 500 lines
- All 6 required sections present (Purpose, Steps, Output, Scope, Completion, Failure Handling)
- No unresolved `[SCREAMING_CASE]` placeholders in body
- Base failure handling pattern rows present in Failure Handling section (4 required rows)

**Template tests** (all `templates/**/*.template.md`):
- All `[PLACEHOLDER]` tokens in a template are listed in the corresponding `compiled/*.rules.md` placeholder catalog
- Every placeholder in the catalog appears in the template (bidirectional cross-reference)

**Rules file tests** (all `templates/**/*.rules.md`):
- Frontmatter present and includes `type`, `target`, `sources`, `created`
- Source L0 Principles table present
- Placeholder Catalog section present and non-empty
- L1 Directive Compilation section present
- Compilation Guidance section present
- MUST section contains no negative directives (lines matching `^- MUST NOT` absent from MUST blocks)

**Framework file tests** (all `framework/*.md`):
- No MUST/MUST NOT/SHOULD directive lines (heuristic: `^- MUST|^- SHOULD`)
- No explicit file path references suggesting specificity leakage (heuristic, not strict gate)

**Level agent tests** (all `agents/smaqit.L*.agent.md`):
- Valid YAML frontmatter
- `tools` field is a non-empty list
- Required sections present: Role, Input, Output, Directives, Scope Boundaries, Completion Criteria, Failure Handling
- Failure Handling section contains a markdown table
- Completion Criteria section contains at least one `- [ ]` item

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
evals/
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

**Eval runner:** `installer/cmd/eval-runner/main.go`
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

**Dependency:** `github.com/github/copilot-sdk/go` added to `installer/go.mod`; SDK version pinned.
**Auth requirement:** `COPILOT_GITHUB_TOKEN` (or `GH_TOKEN`) in environment — required for eval runner to run; runner prints a clear error if missing.

---

### Phase 4 — Makefile Wiring

```makefile
# Rename existing test target
test-scaffold: build    # (preserved, renamed from test)
  ...

# New targets
test: prepare           # Runs go test ./... (Layers 1 + 2)
  cd installer && go test ./...

evals: build            # Runs eval runner (Layer 3, requires API key)
  cd installer && go run ./cmd/eval-runner/... $(REPO_ROOT)/evals/

test-all: test evals    # All layers
```

---

## Acceptance Criteria

- [ ] Phase 0: `smaqit.new-skill` embedded and installed by `cmdInit`
- [ ] Phase 1: `make test` passes with all installer unit tests green
- [ ] Phase 2: `make test` passes with all structural validation tests green
- [ ] Phase 2: Intentionally breaking a skill description (add "I ") causes a test failure
- [ ] Phase 2: Adding `[UNRESOLVED]` to a skill body causes a test failure
- [ ] Phase 2: Removing a placeholder from a template without updating the catalog causes a test failure
- [ ] Phase 3: `make evals COPILOT_GITHUB_TOKEN=...` runs without crashing and outputs per-criterion results
- [ ] Phase 3: At least 3 eval files exist for `smaqit.new-agent` and 2 for `smaqit.new-skill`
- [ ] Phase 3: At least 2 eval files exist for `smaqit.L2` (agent compilation behavior)

## Design Decisions

- **All Go tests in `installer/`** — single Go module, no new module; root artifacts accessed via `../` relative paths
- **Eval runner as separate binary** (`cmd/eval-runner/`) — not mixed into `main.go`; runnable without building the full CLI
- **Copilot Go SDK for eval runtime** — `github.com/github/copilot-sdk/go`; SDK is public (technical preview); chosen over Anthropic API because both Task 010 and Task 011 require the same Copilot session driver; no API key duplication
- **Isolated workspace = temp dir + `ClientOptions.Cwd`** — clean room via OS temp dir; standardized `.github/` scaffold inside; the CLI sees only what we put there
- **System prompt injection via `Mode: "replace"`** — the artifact under test (SKILL.md or .agent.md) becomes the full system prompt; no Copilot CLI persona, no contaminating instructions
- **Multi-turn scripted conversations** — `OnUserInputRequest` drives `ask_user` triggers from eval JSON `turns`; enables end-to-end gathering flow validation
- **Grader = second Copilot session** — natural language criteria require semantic grading; a second session evaluates each criterion against the collected message history
- **`evals/` at repo root** — eval definitions are ADK artifacts, not build artifacts; not under `installer/`
- **Both skills and agents evaluated** — skills validate gathering flows; agents validate directive compliance and behavioral boundaries
- **Framework file tests are heuristic** — MUST/SHOULD detection is reliable; "no file paths" is best-effort, not a hard gate
- **Embed bug fix is Phase 0 prerequisite** — installer unit tests must be able to assert correct post-`init` state including `smaqit.new-skill`
- **SDK version pinned** — technical preview; pin to specific release in `go.mod` to prevent silent breaking changes
