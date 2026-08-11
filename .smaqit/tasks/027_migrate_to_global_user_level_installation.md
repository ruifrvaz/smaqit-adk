# Migrate to Global User-Level Installation (Learned from smaqit-extensions)

**Status:** Not Started
**Created:** 2026-08-10

## Description

`smaqit-adk` currently installs its agents and skills into the *project* directory via `smaqit-adk lite` / `smaqit-adk advanced` — `.github/agents/smaqit.L0.agent.md`, `.github/agents/smaqit.L1.agent.md`, `.github/agents/smaqit.L2.agent.md`, `.github/skills/smaqit.new-principle/`, `.github/skills/smaqit.create-agent/`, `.github/skills/smaqit.create-skill/` (installer/main.go:101-110, 217-236). Every repository using smaqit-adk carries its own copy of whichever tier's agents/skills were installed, requiring re-init and re-sync per project.

The sibling project `smaqit-extensions` (github.com/ruifrvaz/smaqit-extensions) just completed and shipped this exact migration — Task 023, released v1.14.0 through v1.14.2 — moving agents/skills to global, user-level paths installed once per machine, with `init` reduced to project-local scaffolding only. This task ports that architecture to `smaqit-adk`.

**smaqit-adk's situation is not identical to smaqit-extensions' and requires its own design pass, not a direct copy:**
- **Single platform today.** Unlike smaqit-extensions (Copilot + Claude Code + Codex), smaqit-adk currently ships GitHub Copilot agents/skills only (`.github/agents/`, `.github/skills/` — no `.claude/` or `.codex/` references found anywhere in `installer/main.go`). Do not invent Claude/Codex support as part of this migration unless it's already planned elsewhere — scope this task to Copilot's global equivalent (`~/.copilot/agents/` for agents; determine the correct global skills path for Copilot-only skills, mirroring smaqit-extensions' `~/.agents/skills/` convention if a shared/cross-tool skills location makes sense, or a Copilot-specific one if not).
- **Tier system is a genuine open design question, not a detail.** `lite` and `advanced` currently install *different subsets* of agents/skills per project (`cmdLite`/`cmdAdvanced` at installer/main.go:174/196). Global installation raises a real question this task must resolve during planning, not assume an answer to: does the global install always ship the union of both tiers (harmless — unused agents/skills just aren't invoked), or does the machine-level install need its own tier selection mechanism? Do not default to either silently; surface this explicitly during `smaqit.task-plan` (or this project's planning skill) and get an explicit decision before implementing.
- **`smaqit-adk init` no longer exists as a command** (Task 017 replaced it with `lite`/`advanced` — see PLANNING.md Completed). Whatever project-facing scaffolding command remains after this migration (`lite`/`advanced`, or a new dedicated project-scaffold command) needs the same treatment smaqit-extensions gave `init`: it should scaffold project-local state only, not install agents/skills into the project.

## Design Decisions

- **Global paths — Copilot-only equivalent of smaqit-extensions' model:**
  - `~/.copilot/agents/` — GitHub Copilot custom agents (L0/L1/L2, tier-dependent — see open tier question below)
  - Skills global path — TBD during planning: either a shared `~/.agents/skills/` (matching smaqit-extensions' convention, useful if this machine also runs smaqit or smaqit-extensions skills) or a smaqit-adk-specific path. Resolve explicitly rather than defaulting.
  - Respect an equivalent of `COPILOT_HOME` for the agent root override.
- **No user-facing `install` subcommand.** smaqit-extensions initially shipped a new `install` subcommand (with `--scope`/`--agent` flags) as the mechanism for global installation, then had to walk it back one release later because it created three competing meanings of "install" (the shell installer script, the CLI subcommand, and the legacy per-project command). Go straight to the corrected design: `install.sh` downloads the binary and immediately triggers global installation via a hidden/internal flag (e.g. `--install-global`), not a subcommand a user is expected to type.
- **Command-line default must show help, not silently act.** An early version of the corrected smaqit-extensions CLI made bare `smaqit-extensions` (no args) scaffold the project silently — this was flagged as wrong. Ensure `smaqit-adk` with no arguments prints help; project scaffolding (if anything remains project-local) requires an explicit verb.
- **Tier selection at global-install time (open question, resolve during planning):** if the tier system survives this migration, decide whether tier selection happens at global-install time (`smaqit-adk` installer script prompts/flags for lite vs advanced) or whether the global install always ships everything and tier only gates what a *project* is told is "active" via its own local config. Do not assume; this is a real design fork specific to smaqit-adk that smaqit-extensions never had to solve.

## Implementation Steps

[TBD — to be resolved during planning/discovery. smaqit-adk's installer follows a broadly similar embed/copy shape to smaqit-extensions' pre-migration installer but is materially different in scope (single platform, tiered installs, no `init` command currently). Do not copy smaqit-extensions' or smaqit's implementation steps verbatim — re-derive against smaqit-adk's actual `installer/main.go`, `install.sh`, and tier system.]

Suggested high-level phases (adapt, do not copy):

1. Resolve the tier-selection-at-global-install open design question (Design Decisions) before writing any code.
2. Resolve the skills-global-path open question (shared `~/.agents/skills/` vs smaqit-adk-specific) before writing any code.
3. Add a global-path resolver for the (now Copilot-only, or however tier resolution decided) install targets.
4. Split whatever project-local scaffolding survives (if any) from global agent/skill installation in `installer/main.go`.
5. Wire `install.sh` to call the binary's internal global-install flag automatically after download. **Test this step for real** — smaqit-extensions shipped a broken `install.sh` (a `$target` variable referenced out of its defining function's scope, causing the global-install step to silently fail with "command not found") that passed its entire automated smoke-test suite because the suite invokes the binary directly, never through the actual shell installer. The bug was only caught by manually running `curl | bash` against a sandboxed `$HOME` and inspecting the resulting directory tree. Do the equivalent real end-to-end test here before considering this step done.
6. Update README, CHANGELOG, and any other documentation describing `lite`/`advanced` install behavior.
7. Cut a release and **verify with a real `curl | bash` install into a sandboxed `$HOME`** — not just the automated test suite — before considering this task complete.

## Known Issues Triage

[Populated by task-start via triage-issues, if this project has an equivalent skill. Do not edit manually.]

## Acceptance Criteria

- [ ] The tier-selection-at-global-install design question is explicitly resolved and documented (not silently defaulted) before implementation begins
- [ ] The skills-global-path design question (shared vs smaqit-adk-specific) is explicitly resolved and documented before implementation begins
- [ ] Agents and skills install to the resolved global path(s), not into any project directory, after running the installer script
- [ ] Whatever project-scaffolding command remains (successor to `lite`/`advanced` or a new dedicated command) creates only genuinely project-local state — verified by inspecting the project directory tree afterward, confirming no `.github/agents/` or `.github/skills/` appear there
- [ ] No user-facing `install` subcommand exists; global installation is triggered automatically by the installer script
- [ ] `smaqit-adk` with no arguments prints help
- [ ] A real `curl | bash` install against a sandboxed `$HOME` succeeds end-to-end (binary downloads, global install actually populates the expected directories) — verified by direct inspection, not solely by automated test suite passing
- [ ] Existing automated test suite passes
- [ ] `CHANGELOG.md` and `README.md` updated to describe the new installation model, including the tier-selection resolution

## Findings

[Populated on completion. Do not fill in manually before task is complete.]

**Implementation approach:**
- TBD

**Decisions made:**
- TBD

**Blockers encountered:**
- TBD

**Follow-up identified:**
- TBD

## Files to Create / Modify

| File | Action |
|------|--------|
| `installer/main.go` | Modify — global path resolver, split project/global install, tier-resolution decision |
| `install.sh` | Modify — trigger global install automatically after binary download |
| `README.md` | Modify — document global install model and tier resolution |
| `CHANGELOG.md` | Modify — record the change |
| Any skill/agent files referencing project-local `.github/agents/`, `.github/skills/` | Modify — update to reflect global paths |

## Notes

**Source of this task:** ported from `smaqit-extensions` Task 023 ("Global User-Level Installation with Agent-Specific Adapters"), which shipped as v1.14.0, then required two immediate patch releases:

- **v1.14.1** — fixed the deprecated per-project scaffold command re-delegating to the full per-project install path (reproducing the exact "install everything into the project" behavior the migration existed to eliminate) — caught only after a real user ran the real installer, not during automated testing.
- **v1.14.2** — fixed a broken `install.sh` (`$target` variable referenced out of function scope), meaning the *first* released version of the corrected installer downloaded the binary successfully but silently never installed anything globally. Caught by manually running `curl | bash` against a sandboxed `$HOME` and inspecting the resulting directory tree — the automated smoke-test suite did not catch this because it invokes the binary directly, not through the shell installer script.

**The core lesson driving this task's emphasis on real end-to-end verification:** automated test suites verify that *code paths* work when invoked directly. They do not verify that the actual user-facing entry point (the `curl | bash` command a real user runs) wires those code paths together correctly. Both smaqit-extensions patch releases were needed specifically because the shell-script integration layer was never exercised end-to-end before release. Budget time in this task specifically for a real, unscripted install run before considering it done.

**smaqit-adk-specific complications not present in smaqit-extensions**, both requiring an explicit decision during planning rather than an assumed default: (1) the lite/advanced tier system and what it means for a machine-level global install, and (2) whether smaqit-adk's skills should share smaqit-extensions' `~/.agents/skills/` convention or use their own path. Invoke `smaqit.task-plan` (or this project's planning skill, if named differently) before implementation to resolve both.
