# Changelog

All notable changes to smaqit-adk will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.7.3] - 2026-05-17

### Added

- `smaqit.project-compendium` skill — manages a live Q&A knowledge manifest at `.smaqit/compendium.md`; supports list, fetch, upsert, and remove operations
- `smaqit.project-glossary` skill — manages a per-project glossary at `.smaqit/glossary.md`; supports list, fetch, upsert, and remove operations
- `smaqit.project-init` skill — bootstraps a new smaqit project by generating a structured `.github/copilot-instructions.md` from a template
- `smaqit.project-recap` skill — generates a live project dashboard from the current codebase state and writes it to `.smaqit/project-recap.md`
- `smaqit.project-research` skill — builds and maintains a documentation topology map; discovers section-level docs across GitHub, official docs, ReadTheDocs, pkg.go.dev, npm, PyPI, and more
- `smaqit.session-recap` skill — summarizes session progress as a structured table of accomplished and pending steps
- `smaqit.utils.read-pdf` skill — PDF content extraction utility for skills and agents that need to process PDF files
- `smaqit.utils.triage-issues` skill — pre-implementation gate that searches upstream GitHub repositories for open bugs and regressions relevant to a task; classifies results as Blocking, Advisory, Historical, or Clear
- `.smaqit/templates/` directory — ships PLANNING template, `copilot-instructions` template, and task template for project bootstrapping

### Changed

- `smaqit.release.local`, `smaqit.release.pr`, and `smaqit.user-testing` agent definitions updated to reflect current skill and workflow conventions
- `smaqit.session-finish`, `smaqit.session-title`, and `smaqit.session-recap` skills enriched with `scripts/recap.py` helper for structured session documentation
- `smaqit.task-create` skill updated with `assets/TASK_TEMPLATE.md` for consistent task scaffolding
- `smaqit.task-complete`, `smaqit.task-start`, `smaqit.task-list`, `smaqit.test-start`, `smaqit.session-start` skill definitions updated with refined rules and references
- `smaqit.release-analysis`, `smaqit.release-git-pr`, `smaqit.release-prepare-files` skill definitions updated with improved boundary-commit reconciliation guidance

### Fixed

- Release workflow consolidated: replaced `release.yml` with the improved `post-merge-release.yml`; eliminates duplicate trigger paths and ensures single authoritative post-merge automation

## [0.7.2] - 2026-05-17

### Changed

- Installer version bumped to 0.7.2

## [0.7.1] - 2026-05-15

### Added

- Post-merge release workflow (`.github/workflows/post-merge-release.yml`) — automatically creates git tag, builds multi-platform binaries, and publishes GitHub Release when a release PR is merged to `main`

## [0.7.0] - 2026-05-15

### Changed

- `smaqit.create-skill` description rewritten to imperative phrasing with broader trigger scope and explicit output description; covers indirect intent ("package a workflow as a reusable command", "wrap domain knowledge into a slash-command")
- `smaqit.create-skill` Step 3 definition file spec now includes `gotchas` (proactive environment facts), `examples` (concrete triggering request + output), and optional `allowed-tools` / `compatibility` sections
- `smaqit.create-skill` Step 4 compile instruction now includes progressive disclosure guidance: skills exceeding 400 lines must extract reference content to a `references/` subdirectory
- `base-skill.template.md` updated with `## Examples` and `## Gotchas` sections and optional `compatibility` / `allowed-tools` frontmatter fields (omitted when not specified)
- `skill.rules.md` description field directive updated to imperative/intent-focused phrasing; placeholder catalog extended with `[EXAMPLES_CONTENT]`, `[GOTCHAS_CONTENT]`, `[COMPATIBILITY]`, `[ALLOWED_TOOLS]`; progressive disclosure directive added with 400-line threshold

## [0.6.0] - 2026-04-16

### Added

- `smaqit.new-principle` skill — advanced-tier entry point for adding or refining principles in the ADK framework; installed by `smaqit-adk advanced`
- Inference-first creation flow for `smaqit.create-agent` and `smaqit.create-skill`: scans repo, infers full specification from name and purpose, writes a definition file to `.smaqit/definitions/`, invokes `smaqit.L2` to compile — no draft, no confirmation step

### Changed

- `smaqit-adk lite` now installs `smaqit.L2` agent + templates (`.smaqit/templates/`) + `smaqit.create-agent` / `smaqit.create-skill` skills; no longer installs compiled product agents
- `smaqit-adk advanced` now installs lite tier as a superset, then adds L0/L1 agents + framework + `smaqit.new-principle`
- `smaqit-adk uninstall lite` now removes entire `.smaqit/` directory (templates + user definition files)
- `smaqit-adk uninstall advanced` removes advanced-only components (L0/L1, new-principle, `.smaqit/framework/`) while preserving lite tier
- `smaqit.create-agent` and `smaqit.create-skill` skills rewritten to inference-first pattern (v2.0.0)
- Copilot SDK dependency removed from installer; `go.mod` cleaned up

### Removed

- `smaqit.create-agent` and `smaqit.create-skill` compiled agents (L2 is now the compiler)
- `smaqit.new-agent` and `smaqit.new-skill` skills (replaced by rewritten `smaqit.create-*` skills)

## [0.5.0] - 2026-04-05

### Added

- `smaqit-adk lite` — new CLI subcommand; installs lite-tier artifacts (2 agents + 2 routing skills) into `.github/`
- `smaqit-adk advanced` — new CLI subcommand; installs full ADK into `.smaqit/` (Level agents L0/L1/L2, framework files, templates, and advanced skills)

### Changed

- `smaqit-adk init` deprecated with migration message directing users to `smaqit-adk lite`
- CI workflow updated to test both `lite` and `advanced` subcommands independently
- README Quick Start updated to use `smaqit-adk lite`
- `install.sh` next steps updated to reference `smaqit-adk lite`
- `smaqit.create-agent` and `smaqit.create-skill` skill error tables: stale `smaqit-adk init` references updated to `smaqit-adk lite`
- ADK wiki structure section: corrected lite-tier output tree; added advanced-tier output tree

## [0.4.0] - 2026-04-03

### Added

- `smaqit.create-agent` routing skill — lite-tier entry point installed by `init`; activates via natural language ("create a new agent") or `/smaqit.create-agent` slash command; delegates to the `smaqit.create-agent` agent as a subagent
- `smaqit.create-skill` routing skill — lite-tier entry point installed by `init`; activates via natural language ("create a new skill") or `/smaqit.create-skill` slash command; delegates to the `smaqit.create-skill` agent as a subagent

### Changed

- `smaqit-adk init` now installs 4 files into `.github/`: 2 agents (`smaqit.create-agent`, `smaqit.create-skill`) + 2 routing skills (`smaqit.create-agent/SKILL.md`, `smaqit.create-skill/SKILL.md`)
- `smaqit-adk uninstall` now removes routing skill files and directories in addition to agents
- README Quick Start updated — natural language entry point ("say 'create a new agent'") is now the primary UX, replacing direct agent context switch
- `install.sh` next steps updated to reflect natural language invocation

## [0.3.2] - 2026-04-02

### Fixed

- `create-agent` / `create-skill` CLI: wrong agent context — was using `smaqit.L2 + smaqit.new-agent` skill (which invokes L2 as a subagent, unsupported in CLI sessions); now uses self-contained `smaqit.create-agent` / `smaqit.create-skill` agents
- `create-agent` / `create-skill` CLI: removed 15-minute session timeout; interactive sessions are human-paced and exit via Ctrl-C only

### Changed

- `smaqit.create-agent` and `smaqit.create-skill`: agents now scan the project repository before asking questions — reads existing agents, skills, README, and config files to infer defaults; asks only name and description/purpose explicitly; presents a full draft for one confirmation pass before compiling
- `smaqit.create-agent` and `smaqit.create-skill`: added `read` and `search` tools to frontmatter to support repo scanning
- Makefile eval target: auto-detects GitHub token via `gh auth token`; explicit `GH_TOKEN` override still supported

### Removed

- `agents/qa-helper.agent.md` — test artifact not part of the ADK agent catalog

## [0.3.1] - 2026-03-30

### Fixed

- `create-agent` / `create-skill` CLI: agent questions were never displayed — `OnUserInputRequest` was ignoring `req.Question`; user saw only `>` with no context
- Progress ticker no longer prints `[working... Xs]` while stdin is blocking for user input

## [0.3.0] - 2026-03-30

### Added

- `smaqit-adk create-agent` — interactive CLI command; gathers agent specs via Copilot SDK in an isolated LLM context and writes a compiled `.agent.md` into `.github/agents/`
- `smaqit-adk create-skill` — interactive CLI command; gathers skill specs via Copilot SDK in an isolated LLM context and writes a compiled `SKILL.md` into `.github/skills/<name>/`
- Copilot SDK integration (`github.com/github/copilot-sdk/go`) — enables programmatic Copilot sessions from the CLI
- Eval runner under `tests/evals/runner/` — drives Copilot SDK evaluation sessions from the command line with workspace isolation and grading
- 7 evals across `smaqit.L2` and `smaqit.new-agent` / `smaqit.new-skill` skills

### Changed

- `installer/main.go` refactored to include `cmdCreate` driving full interactive `create-agent` / `create-skill` sessions
- README updated with advanced-tier CLI documentation

### Removed

- `HANDOVER.md` removed

## [0.2.0] - 2026-03-29

### Added

- `smaqit.create-agent` — self-contained lite-tier agent that gathers specs interactively and compiles `.agent.md` files; installed by `init`
- `smaqit.create-skill` — self-contained lite-tier agent that gathers specs interactively and compiles `SKILL.md` files; installed by `init`
- `smaqit.new-agent` skill — advanced-tier creation skill with definition file output and L2 subagent invocation
- `smaqit.new-skill` skill — advanced-tier creation skill with definition file output and L2 subagent invocation
- Skill compilation layer: `templates/skills/`, `skill.rules.md`, L2 extended for skill compilation
- Go-based test framework under `tests/` with unit and structural suites

### Changed

- `smaqit-adk init` now installs only `smaqit.create-agent` and `smaqit.create-skill` into `.github/agents/` — no framework files, templates, or skills distributed
- Framework `PROMPTS.md` replaced by `SKILLS.md`; L0 principles rewritten to behavioral-only
- Skill compilation ownership corrected from L1 to L2
- README fully rewritten for lite-tier model

### Removed

- `prompts/smaqit.new-agent.prompt.md` (migrated to `skills/smaqit.new-agent/SKILL.md`)
- `framework/PROMPTS.md` (replaced by `framework/SKILLS.md`)

## [0.1.0] - 2026-02-04

### Added

- Initial ADK extraction from smaQit monorepo
- Generic framework files (5): SMAQIT.md, AGENTS.md, TEMPLATES.md, ARTIFACTS.md, PROMPTS.md
- Generic agent templates (3): base-agent, specification-agent, implementation-agent
- Generic compilation rules (3): base, specification, implementation
- Level agents (3): L0 (principle curator), L1 (template compiler), L2 (agent compiler)
- new-agent prompt template for creating custom agents
- Self-contained installer with no external dependencies
- CLI commands: init, help, uninstall, version

### Philosophy

smaqit-adk is a **generic agent development toolkit**, not tied to any specific domain or layer model. It provides the compilation infrastructure for building custom agent orchestration systems.

The [smaQit product](https://github.com/ruifrvaz/smaqit) demonstrates one possible use case (five-layer specification system), but ADK users can create entirely different architectures.

[Unreleased]: https://github.com/ruifrvaz/smaqit-adk/compare/adk-v0.7.3...HEAD
[0.7.3]: https://github.com/ruifrvaz/smaqit-adk/compare/adk-v0.7.2...adk-v0.7.3
[0.7.2]: https://github.com/ruifrvaz/smaqit-adk/compare/adk-v0.7.1...adk-v0.7.2
[0.7.1]: https://github.com/ruifrvaz/smaqit-adk/compare/adk-v0.7.0...adk-v0.7.1
[0.7.0]: https://github.com/ruifrvaz/smaqit-adk/compare/adk-v0.6.0...adk-v0.7.0
[0.6.0]: https://github.com/ruifrvaz/smaqit-adk/compare/adk-v0.5.0...adk-v0.6.0
[0.5.0]: https://github.com/ruifrvaz/smaqit-adk/compare/adk-v0.4.0...adk-v0.5.0
[0.4.0]: https://github.com/ruifrvaz/smaqit-adk/compare/adk-v0.3.2...adk-v0.4.0
[0.3.2]: https://github.com/ruifrvaz/smaqit-adk/compare/adk-v0.3.1...adk-v0.3.2
[0.3.1]: https://github.com/ruifrvaz/smaqit-adk/compare/adk-v0.3.0...adk-v0.3.1
[0.3.0]: https://github.com/ruifrvaz/smaqit-adk/compare/adk-v0.2.0...adk-v0.3.0
[0.2.0]: https://github.com/ruifrvaz/smaqit-adk/compare/adk-v0.1.0...adk-v0.2.0
[0.1.0]: https://github.com/ruifrvaz/smaqit-adk/releases/tag/adk-v0.1.0
