# Changelog

All notable changes to smaqit-adk will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/ruifrvaz/smaqit-adk/compare/adk-v0.3.1...HEAD
[0.3.1]: https://github.com/ruifrvaz/smaqit-adk/compare/adk-v0.3.0...adk-v0.3.1
[0.3.0]: https://github.com/ruifrvaz/smaqit-adk/compare/adk-v0.2.0...adk-v0.3.0
[0.2.0]: https://github.com/ruifrvaz/smaqit-adk/compare/adk-v0.1.0...adk-v0.2.0
[0.1.0]: https://github.com/ruifrvaz/smaqit-adk/releases/tag/adk-v0.1.0
