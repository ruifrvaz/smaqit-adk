# Changelog

All notable changes to smaqit-adk will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-02-04

### Added

- Initial ADK extraction from smaqit monorepo
- Generic framework files (5): SMAQIT.md, AGENTS.md, TEMPLATES.md, ARTIFACTS.md, PROMPTS.md
- Generic agent templates (3): base-agent, specification-agent, implementation-agent
- Generic compilation rules (3): base, specification, implementation
- Level agents (3): L0 (principle curator), L1 (template compiler), L2 (agent compiler)
- new-agent prompt template for creating custom agents
- Self-contained installer with no external dependencies
- CLI commands: init, help, uninstall, version

### Philosophy

smaqit-adk is a **generic agent development toolkit**, not tied to any specific domain or layer model. It provides the compilation infrastructure for building custom agent orchestration systems.

The [smaqit product](https://github.com/ruifrvaz/smaqit) demonstrates one possible use case (five-layer specification system), but ADK users can create entirely different architectures.

[Unreleased]: https://github.com/ruifrvaz/smaqit-adk/compare/adk-v0.1.0...HEAD
[0.1.0]: https://github.com/ruifrvaz/smaqit-adk/releases/tag/adk-v0.1.0
