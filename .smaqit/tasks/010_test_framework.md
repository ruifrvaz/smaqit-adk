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
  README.md
```

**Eval JSON format** (Anthropic standard + `skill_file` extension field):

```json
{
  "skill": "smaqit.new-agent",
  "skill_file": "skills/smaqit.new-agent/SKILL.md",
  "query": "I need to create a QA agent for this project",
  "expected_behavior": [
    "Asks for the agent name before proceeding",
    "Asks for a description under 80 characters",
    "Asks which tools the agent needs from the allowed list",
    "Does not generate an agent file without completing all 8 gathering sections"
  ]
}
```

**Eval runner:** `installer/cmd/eval-runner/main.go`
- Parses eval JSON files from the `evals/` path passed as argument
- Builds a synthetic system prompt injecting the skill file content
- Sends query to Anthropic API (reads `ANTHROPIC_API_KEY` from env)
- Grades each `expected_behavior` criterion via a second LLM call: "Did the response satisfy: [criterion]? Answer Yes or No."
- Outputs pass/fail per criterion and overall result
- Dependency: `github.com/anthropics/anthropic-sdk-go` added to `go.mod`

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
- [ ] Phase 3: `make evals ANTHROPIC_API_KEY=...` runs without crashing and outputs per-criterion results
- [ ] Phase 3: At least 3 eval files exist for `smaqit.new-agent` and 2 for `smaqit.new-skill`

## Design Decisions

- **All Go tests in `installer/`** — single Go module, no new module; root artifacts accessed via `../` relative paths
- **Eval runner as separate binary** (`cmd/eval-runner/`) — not mixed into `main.go`; runnable without building the full CLI
- **Behavioral evals graded by a second LLM call** — natural language criteria require semantic grading; no manual scoring
- **`evals/` at repo root** — eval definitions are ADK artifacts, not build artifacts; not under `installer/`
- **Framework file tests are heuristic** — MUST/SHOULD detection is reliable; "no file paths" is best-effort, not a hard gate
- **Embed bug fix is Phase 0 prerequisite** — installer unit tests must be able to assert correct post-`init` state including `smaqit.new-skill`
