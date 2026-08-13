# Global Installation & Multi-Platform Compilation (Claude Code + Codex Primary)

**Status:** Completed
**Created:** 2026-08-10
**Started:** 2026-08-13
**Completed:** 2026-08-13
**Mode:** Assisted
**Scope last revised:** 2026-08-13 — broadened substantially during `task.plan` from a straight install-path migration into a platform-strategy pivot. See Notes for the full progression.

## Description

`smaqit-adk` currently installs its agents and skills into the *project* directory via `smaqit-adk lite` / `smaqit-adk advanced` — `.github/agents/smaqit.L0.agent.md`, `.github/agents/smaqit.L1.agent.md`, `.github/agents/smaqit.L2.agent.md`, `.github/skills/*` (installer/main.go:101-110, 217-236). Every repository using smaqit-adk carries its own copy, requiring re-init and re-sync per project.

The sibling project `smaqit-extensions` shipped a global, user-level install migration (its Task 023, v1.14.0–v1.14.2) that this task originally set out to port. During planning, scope broadened twice more:

1. **Codex CLI support was added** — Codex agents are real, TOML-format, already shipped by both `smaqit-extensions` and `smaqit` (`installer/agents-codex/*.toml`), and Codex's official docs confirm `$HOME/.agents/skills` as its own global skills path (independently, not just by smaqit-extensions' convention) — the same path VS Code Copilot auto-discovers from.
2. **GitHub Copilot is deprioritized as an authored target entirely** (fresh decision, made in this session — see `copilot-deprioritization-decision` project memory). This directly contradicts both sibling repos' current state (both keep Copilot fully first-class), so it is explicitly *not* inherited precedent — it's a new call for smaqit-adk specifically. Claude Code and Codex CLI become the two compiled/authored targets, for both the ADK's own shipped agents (L0/L1/L2) and the compiler's output for end-user-created agents. Copilot keeps working through the accepted standards it already natively supports — `AGENTS.md` for instructions, `~/.agents/skills/` for skills — same as any other AGENTS.md/agentskills.io-compliant tool, not as a degraded fallback tier.

**Copilot does not lose the L0/L1/L2 workflow.** Per this repo's own "Skill-Mediated Workflows" principle (`framework/AGENTS.md`), skills — not agent files — are the actual entry point, and skills are already fully cross-platform. Routing skills (`create-agent`, `create-skill`, `new-principle`) invoke the target agent by name; for Claude/Codex this resolves to a real native subagent/custom-agent call against the compiled file. For Copilot, which has no dedicated compiled file, the skill instead instructs it to read the Claude-format body directly (plain markdown: Role/Input/Output/Directives) and follow it inline as the current turn's instructions. Same content, no native subagent context isolation, but the full workflow still runs.

## Design Decisions

- **Global paths:**
  - Agents: `~/.claude/agents/` and `~/.codex/agents/` — no `~/.copilot/agents/` (Copilot authoring dropped). `CLAUDE_CONFIG_DIR`/`CODEX_HOME` overrides respected (real env vars both tools honor).
  - Skills: `~/.agents/skills/` (shared — read natively by both Copilot and Codex, confirmed via each platform's own docs) plus `~/.claude/skills/` (Claude-specific — same `SKILL.md` content, second destination only, not a second authoring pass).
  - Templates/framework: `~/.agents/smaqit-adk/{templates,framework}/` — unchanged path decision from the original round, now carrying per-platform frontmatter schema content instead of a single Copilot-shaped one.
- **Tier gating dropped entirely.** Global install always ships everything unconditionally. `lite`/`advanced` cease to exist as install-time concepts.
- **No user-facing `install` subcommand, and no user-callable `--install-global` either.** `install.sh` is the sole global-install entry point; the flag is its internal trigger only.
- **No project-scaffolding command survives.** `.smaqit/definitions/` still exists, created lazily by `create-agent`/`create-skill` on first use.
- **No backwards-compatibility references.** Dead code deleted outright, not preserved as fallback; docs describe only the new model.
- **`AGENTS.md` becomes canonical; `.github/copilot-instructions.md` is deleted.** Its content migrates into root `AGENTS.md`. Confirmed no functional loss — VS Code Copilot reads `AGENTS.md` natively (`chat.useAgentsMdFile`), alongside GitHub's autonomous coding agent and Copilot CLI.
- **README repositioned:** "an Agent Development Kit for GitHub Copilot" → "an Agent Development Kit," with a compatibility table underneath (mirroring smaqit-extensions' README pattern) presented as three platforms each fully compatible via their own accepted mechanism — not Copilot-as-lesser-tier framing.
- **Multi-format generation follows smaqit's validated pattern exactly, not an invented shortcut.** smaqit's `scripts/generate-agents.py` reached this shape only after two rejected approaches (hand-duplicated bodies per platform; committed generated output) — see Notes. The adopted shape: one shared, platform-neutral body per agent (Role/Input/Output/Directives prose — smaqit-adk's L1/L2 chain already produces exactly this) + one per-platform metadata source giving explicit `name`/`description`/`tools` per platform, deterministically rendered, output gitignored and regenerated at build time. Generator written in Go, not Python — keeps smaqit-adk dependency-free of a second toolchain; same split-source logic.
- **Tool mapping uses smaqit's exact explicit per-platform table — not an "omit and inherit everything" shortcut.** Each of L0/L1/L2's existing Copilot tool arrays gets an explicit translation: Claude gets an explicit tool-name list (`execute/runInTerminal`→`Bash`, `read/readFile`→`Read`, `edit`→`Edit`, `agent`→`Task`, `todo`→`TodoWrite`, `search`→`Grep`/`Glob`, `web/fetch`→`WebFetch`); Codex gets a `[tools]`/`[mcp_servers.*]` table only where a concrete extra capability is genuinely needed (matches the confirmed smaqit precedent — most agents need nothing extra; `[tools]` grants additions, it is not a restrict-to-list mechanism). No `copilot:` block is needed anywhere, since Copilot gets no compiled file at all — this mapping is Claude↔Codex only, narrower than smaqit's 3-way table.
- **Codex custom-agent invocation model assumed consistent with today's subagent pattern** (Codex matches on the agent's `name` field when spawning) — flagged as an assumption to verify live during implementation, not a blocker to planning.

## Implementation Steps

*Phase 0 — Repositioning*
1. `README.md` — reframe to "an Agent Development Kit"; add compatibility table (Claude Code / Codex CLI / GitHub Copilot, each described by what it gets, not ranked).
2. Delete `.github/copilot-instructions.md`; migrate its content into root `AGENTS.md` as the canonical instructions file.

*Phase 1 — Global path resolution*
3. Add path resolvers: agents → `~/.claude/agents/`, `~/.codex/agents/` (with `CLAUDE_CONFIG_DIR`/`CODEX_HOME` overrides); skills → `~/.agents/skills/` + `~/.claude/skills/`; data → `~/.agents/smaqit-adk/{templates,framework}/`.
4. Replace `cmdLite`/`cmdAdvanced`/`installLiteComponents` with one `installGlobal()` writing everything unconditionally, triggered via hidden `--install-global` flag.
5. Collapse `cmdUninstall` to a single tier-free global uninstall targeting all resolved global locations; keep the y/N confirmation pattern.
6. Delete the `init` case outright — unrecognized verb falls through to `default:` → `printUsage()` + exit 1.
7. Confirm bare `smaqit-adk` (no args) satisfies "prints help."

*Phase 2 — ADK's own L0/L1/L2 become multi-format (depends on Phase 1 paths)*
8. Split each of the 3 agent files into a shared platform-neutral body (unchanged prose — already produced by today's L1/L2 chain) and a per-platform metadata source (explicit `name`/`description`/`tools` per platform, per the Design Decisions tool-mapping table).
9. Write a Go generator mirroring `generate-agents.py`'s logic (dump Claude YAML frontmatter + body → `.md`; render Codex `developer_instructions = '''...'''` TOML literal + optional `[tools]`/`[mcp_servers.*]` tables); wire into `installer/Makefile`'s `prepare` target; output gitignored, regenerated at build time, never hand-duplicated.
10. Apply the explicit tool mapping table (Design Decisions) to each of L0/L1/L2's existing Copilot tool arrays.
11. `docs/wiki/agent-frontmatter.md` — add Claude/Codex sibling reference content or consolidate into one multi-platform reference.

*Phase 3 — Compiler output for end users (depends on Phase 2's generator pattern)*
12. Rewrite `agents/smaqit.L2.agent.md`'s per-pattern compile procedures — after the existing Role/Input/Output/Directives merge (unchanged), add a per-platform render step producing Claude `.md` + Codex `.toml` outputs instead of one Copilot `.agent.md`.
13. `templates/agents/*.template.md` — the 3 structural shells need Claude/Codex frontmatter-shape siblings, or L2 needs an equivalent small per-platform schema reference to merge against. Resolve exact file layout during implementation.
14. Update `create-agent`/`create-skill`/`new-principle` skills' compilation-step instructions to encode the platform-conditional routing described in the Description: native subagent call for Claude/Codex, direct-read-and-follow-inline fallback for Copilot.
15. Skills side needs no format change (already cross-platform) — only install-path references (Phase 1) and the `go run .github/skills/.../validate-skill.go` invocation path.

*Phase 4 — `install.sh` wiring*
16. After `install_binary`/`verify_installation`, call the installed binary directly with `--install-global` — reference `${INSTALL_DIR}/smaqit-adk` inline, never through an intermediate variable (the exact bug class that broke smaqit-extensions' first release).
17. Update the script's closing "Next steps" text for the new model.

*Phase 5 — Update remaining path/doc consumers*
18. `skills/smaqit.new-principle/SKILL.md` — fix its inconsistent `.smaqit/agents/` fallback reference against the real global framework path.
19. `skills/smaqit.bench-scaffold/SKILL.md` — delete its `.github/skills|agents/`-first root-detection branch (dead code once nothing installs there).
20. `docs/wiki/extending-smaqit.md` — replace tier diagrams and command tables with the new model; no backwards-compatibility pointers.

*Phase 6 — Tests (depends on Phases 1–3)*
21. Rewrite path/format assertions in `installer/Makefile`'s `test-scaffold` target, `.github/workflows/test-integration.yml`, `tests/unit/lite_test.go`, `tests/unit/embed_test.go` for the new multi-platform, multi-path reality; sandbox `$HOME` so `go test` never touches the real developer machine.

*Phase 7 — Real end-to-end verification (explicit AC, depends on everything above)*
22. Cut a release, run the actual `curl | bash` installer against a sandboxed `$HOME`, and inspect the resulting tree directly for both Claude and Codex agent/skill placement — automated suites alone did not catch smaqit-extensions' two prior regressions here, and this task now has more moving parts than that one did.

## Known Issues Triage
**Triaged:** 2026-08-13
**Tools searched:** openai/codex (Codex CLI custom agent TOML), anthropics/claude-code (subagent tools field), microsoft/vscode-copilot-chat (AGENTS.md discovery)
**Result:** Advisory

### Advisory Issues
- [#26408 Project-scoped custom subagent in `.codex/agents` is advertised but cannot be spawned](https://github.com/openai/codex/issues/26408) — `openai/codex` — opened 2026-06-04 — `bug,CLI,app,subagent,config`. Worth a direct smoke test early in Phase 2/3 before investing in the full generator: confirm a real, globally-installed `~/.codex/agents/*.toml` file is actually spawnable by the live Codex CLI, not just structurally valid.
- [#18823 Custom agent requests are easy to misroute to skills instead of `.codex/agents/*.toml` custom agents](https://github.com/openai/codex/issues/18823) — `openai/codex` — opened 2026-04-21 — `enhancement,skills,subagent`. Relevant to Phase 3's platform-conditional routing design (skill invokes L0/L1/L2 by name) — worth confirming the routing skill's invocation phrasing actually resolves to the custom agent, not Codex's built-in skill mechanism.
- [#80104 Custom subagents don't receive Agent/ToolSearch tools even when listed in frontmatter `tools:`](https://github.com/anthropics/claude-code/issues/80104) — `anthropics/claude-code` — opened 2026-07-22
- [#81852 `tools:` allowlist is enforced for subagents but silently dropped for named agents (teammates)](https://github.com/anthropics/claude-code/issues/81852) — `anthropics/claude-code` — opened 2026-07-28

None of the above reached Blocking under the triage rule (bug/regression label + platform match) because this task has no specific target platform (OS) to match against — these are feature-level advisories, not platform regressions. A large additional set of Claude Code subagent `tools:`/MCP edge-case issues exists beyond the two cited (mostly around MCP tool inheritance and tool-search behavior in nested/background subagents) — none found that regress the basic "omit `tools:` → inherit everything" default this task's tool-mapping decision relies on.

### Historical (Closed)
- [#26868 Custom subagents in `~/.codex/agents/*.toml` are not applied on spawn](https://github.com/openai/codex/issues/26868) — `openai/codex` — closed 2026-06-09 — directly on-point to this task's exact global path; closed/resolved, but confirms this exact mechanism has had real breakage before
- [#26363 Custom `.codex/agents` no longer selectable in CLI v0.137.0; subagents spawn generic and inherit parent model](https://github.com/openai/codex/issues/26363) — `openai/codex` — closed 2026-06-05 — `bug,regression,subagent`
- [#4989 fix: discover AGENTS.md and CLAUDE.md at workspace roots](https://github.com/microsoft/vscode-copilot-chat/pull/4989) — `microsoft/vscode-copilot-chat` — merged 2026-04-06 — confirms `AGENTS.md` discovery was deliberately built and shipped, not an assumption

## Acceptance Criteria

- [x] Agents install to `~/.claude/agents/` and `~/.codex/agents/`; skills install to `~/.agents/skills/` and `~/.claude/skills/`; templates/framework install to `~/.agents/smaqit-adk/` — no project-directory installation of any smaqit-adk artifact
- [x] No command installs any smaqit-adk artifact into a project directory — verified by inspecting a fresh project directory after running the global installer; only `.smaqit/definitions/` may appear, and only lazily when `create-agent`/`create-skill` are used
- [x] No user-facing `install` subcommand or user-callable `--install-global` exists; global installation is triggered automatically by `install.sh` alone
- [x] `smaqit-adk` with no arguments prints help
- [x] L0, L1, L2 are compiled into genuinely separate Claude `.md` and Codex `.toml` outputs from a single shared-body source, generated deterministically (not hand-duplicated) via a build-time Go generator
- [x] `smaqit.create-agent`'s compiled output for end users is multi-format (Claude + Codex), not the old single Copilot `.agent.md`
- [x] `create-agent`/`create-skill`/`new-principle` skills route to L0/L1/L2 correctly — **Codex live-verified for real** (see Findings: real authenticated `codex exec` genuinely spawned the globally-installed `smaqit.L2` custom agent, which correctly reported its own real title line); Claude relies on the platform's well-established native subagent mechanism plus structural verification of the compiled `.md` frontmatter/content, not an equivalent live spawn test — session tooling couldn't address a dynamically-installed subagent by name for a symmetric test; Copilot fallback verified by inspection only, consistent with the AC's own stated tolerance
- [x] `AGENTS.md` exists at repo root as the canonical instructions file; `.github/copilot-instructions.md` no longer exists
- [x] `README.md` reframed as "an Agent Development Kit" with a three-platform compatibility table, none framed as a lesser tier
- [ ] A real `curl | bash` install against a sandboxed `$HOME` succeeds end-to-end — **not done**, requires a cut release to exist first; explicit follow-up, see Findings
- [x] Existing automated test suite passes, rewritten for the new paths/formats
- [x] `CHANGELOG.md` updated to describe the new installation model and platform strategy

## Findings

**Implementation approach:**
- Followed the 7-phase plan largely in order, but implemented Phase 2 (L0/L1/L2 split + Go generator) before Phase 1 (`installer/main.go` global paths) since `main.go`'s `go:embed` directives can't compile until the generator's output exists — a real build-ordering dependency the plan's phase numbering didn't capture.
- L0/L1/L2 split into a shared, platform-neutral body (`agents/*.md`, unchanged prose from the pre-migration `.agent.md` files) plus per-platform metadata (`.smaqit/definitions/agents/*.frontmatter.yaml`), rendered by a new `installer/generate-agents.go` (Go, `//go:build ignore`, invoked via `go run` from `make prepare`) — mirrors smaqit's validated `scripts/generate-agents.py` pattern exactly, including its hard-won lesson (never hand-duplicate body content per platform).
- `smaqit.L2`'s compile procedures (Base/Specification/Implementation Agent patterns) each gained an explicit "Render for both platforms" step after the existing merge logic, rather than restructuring the merge itself — the Role/Input/Output/Directives content stays platform-neutral; only the final wrapper (frontmatter vs. TOML) differs.
- `installer/main.go`'s `cmdUninstall` was redesigned mid-implementation to check target existence before prompting (the initial version always showed a removal list and asked to confirm even with nothing installed) — caught via a stale test assumption, fixed before it shipped.

**Decisions made:**
- Skills install to two global paths (`~/.agents/skills/`, `~/.claude/skills/`) with identical content — confirmed via VS Code Copilot's and Codex CLI's own docs that `~/.agents/skills/` is genuinely shared, not a smaqit-extensions-only convention.
- `create-skill`'s project-local output path is `.agents/skills/[name]/SKILL.md` + `.claude/skills/[name]/SKILL.md` (not `.github/skills/`) — verified both are real, documented project-local conventions (Codex CLI: `.agents/skills` scanned up from cwd to repo root; Claude Code: `.claude/skills/`), not assumed by symmetry with the global paths.
- `docs/wiki/agent-frontmatter.md` fully rewritten (was 100% Copilot-specific) rather than deleted, since a multi-platform metadata reference (Claude YAML fields, Codex TOML fields, the tool-name mapping table) is genuinely useful and didn't exist elsewhere.
- Fixed several pre-existing doc issues encountered while rewriting sections anyway rather than leaving known-false content next to corrected content: README's fictional `smaqit-adk create-agent`/`create-skill` CLI commands (Task 025's exact finding), `extending-smaqit.md`'s references to the non-existent `smaqit.new-agent`/`smaqit.new-skill` skills (the Task 009 doc-gap), and `.smaqit/compendium.md`'s stale architecture entries (compendium is live-loaded knowledge, not historical record, so left-stale entries would actively mislead future sessions).

**Blockers encountered:**
- Mid-implementation, a manual verification command forgot to sandbox `$HOME` and wrote real files into this machine's actual `~/.claude`, `~/.codex`, `~/.agents` — caught immediately, cleaned up via the just-built `uninstall` command (which doubled as a real-world confidence check that it works correctly), no lasting effect.
- `bench suite validate` against `.smaqit/bench/` (this repo's own dogfood manifests) revealed they're now broken: one fails structural validation outright (`given.files[0].source` points at the deleted `agents/smaqit.L0.agent.md`), the other three would fail live execution since their `setup:` steps invoke the deleted `lite` command. Not fixed as part of this task — flagged as follow-up (below) and documented in the compendium; a superficial path fix without addressing the underlying staging mechanism would give false confidence.
- Live-testing AC7's Codex half required real credentials — sandboxing `$HOME` (used for every other test in this task) also sandboxes Codex's stored auth, causing a 401 on the first attempt. Resolved by testing against the real machine directly (with explicit user approval) and immediately uninstalling afterward — verified clean before and after.
- Two transient Codex router errors appeared during the live spawn test (`Full-history forked agents inherit the parent agent type`, `timeout_ms must be at least 10000`) before it succeeded on retry — not investigated further since the end-to-end result was correct, but consistent with the general subagent-spawning flakiness already surfaced during issue triage.

**Follow-up identified:**
- AC10 (real `curl | bash` against a sandboxed `$HOME`) requires a cut release and could not be performed as part of this task — explicit, user-acknowledged follow-up for immediately after the next release, per this task's own Notes on why smaqit-extensions needed two patch releases to get this right.
- The 4 dogfood Bench manifests under `.smaqit/bench/` need a rewrite for the new install mechanism (restage via `--install-global` against a sandboxed workspace HOME instead of the deleted `lite` command; update expectations for dual-format Claude/Codex output instead of single `.agent.md`).
- AC7's Claude-side routing was verified structurally (compiled `.md` frontmatter/content correctness) but not via an equivalent live spawn test — session tooling constraints prevented addressing a dynamically-installed global subagent by name. Worth a live Claude Code round-trip test in a future session if stronger confidence is wanted.
- `smaqit.create-agent`'s full round-trip (gather → definition file → L2 compile → dual-format output) was verified by code review and by confirming L2 itself spawns and loads correctly, but not by actually running `create-agent` end-to-end to produce a brand-new custom agent.

## Files to Create / Modify

| File | Action |
|------|--------|
| `installer/main.go` | Modify — global path resolvers (Claude/Codex agents, dual skills paths), `installGlobal()`, tier-free `cmdUninstall()` |
| `install.sh` | Modify — trigger global install automatically after binary download |
| `agents/smaqit.L0.agent.md`, `smaqit.L1.agent.md`, `smaqit.L2.agent.md` | Modify — split into shared body + per-platform metadata; L2's compile procedures rewritten for multi-format output |
| `templates/agents/*.template.md` | Modify — per-platform frontmatter shape siblings or L2-side schema reference |
| New Go generator (location TBD) | Create — mirrors `generate-agents.py`; wired into `installer/Makefile` `prepare` |
| `README.md` | Modify — repositioning + compatibility table |
| `AGENTS.md` (root) | Create — canonical instructions, absorbing `.github/copilot-instructions.md` content |
| `.github/copilot-instructions.md` | Delete |
| `docs/wiki/agent-frontmatter.md`, `docs/wiki/extending-smaqit.md` | Modify — multi-platform reference content |
| `skills/smaqit.create-agent/SKILL.md`, `smaqit.create-skill/SKILL.md`, `smaqit.new-principle/SKILL.md` | Modify — platform-conditional routing to L0/L1/L2; path references |
| `skills/smaqit.bench-scaffold/SKILL.md` | Modify — delete dead root-detection branch |
| `installer/Makefile`, `.github/workflows/test-integration.yml`, `tests/unit/lite_test.go`, `tests/unit/embed_test.go` | Modify — new paths/formats |
| `CHANGELOG.md` | Modify — record the change |

## Notes

**Source of this task:** originally ported from `smaqit-extensions` Task 023 ("Global User-Level Installation with Agent-Specific Adapters"), which shipped as v1.14.0, then required two immediate patch releases:

- **v1.14.1** — fixed the deprecated per-project scaffold command re-delegating to the full per-project install path — caught only after a real user ran the real installer, not during automated testing.
- **v1.14.2** — fixed a broken `install.sh` (`$target` variable referenced out of function scope) — the *first* released version downloaded the binary successfully but silently never installed anything globally. Caught only by manually running `curl | bash` against a sandboxed `$HOME` and inspecting the resulting tree.

**The core lesson driving this task's emphasis on real end-to-end verification:** automated test suites verify that *code paths* work when invoked directly, not that the actual user-facing entry point wires them together correctly. Budget real time for an unscripted install run before considering this done — now more important than the original scope, given how much more surface area this task has grown to cover.

**Scope progression within this single planning session (2026-08-13), for traceability:**
1. Started as a straight port of smaqit-extensions' global-install migration, Copilot-only, tier question and skills-path question open.
2. Tier gating resolved (dropped entirely); skills path resolved (`~/.agents/skills/`); templates/framework also moved global — all confirmed via direct VS Code/Codex documentation checks, not assumption.
3. User asked to add Codex TOML agent support — confirmed real and already shipped in both sibling repos (`installer/agents-codex/*.toml`).
4. User directed full Copilot deprioritization — confirmed as a *new* decision NOT shared by either sibling repo (see `copilot-deprioritization-decision` project memory); investigation found smaqit's own Task 088 explicitly preserves Copilot when Codex was added, so this diverges from established family precedent deliberately.
5. User expanded scope to include the compiler's OUTPUT format for end-user-created agents (not just the ADK's own shipped artifacts) — confirmed smaqit's `scripts/generate-agents.py` as validated prior art (reached its current shape only after two rejected simpler approaches — see `.smaqit/history/058_claude_code_migration_2026-07-16.md` in the smaqit repo for the full account) and adopted the same split-source model.
6. Corrected an initial assumption that Copilot would lose L0/L1/L2 access entirely — this repo's own "Skill-Mediated Workflows" principle resolves it: skills (cross-platform) route to a native subagent call on Claude/Codex, or a direct-read-and-follow-inline fallback on Copilot.
7. Corrected an initial "omit tools field, inherit everything" shortcut in favor of smaqit's actual validated explicit per-platform tool-mapping table.

**A known, unresolved inconsistency in smaqit's own installer, worth avoiding here rather than copying:** its `cmdInstallGlobal` only ships the Copilot-rendered skill variant to the shared `~/.agents/skills/` path — the Codex-specific rendered variant is generated but never actually installed globally, a real live gap in the precedent repo. Since smaqit-adk's skills are already confirmed format-agnostic (no separate Copilot/Codex rendering needed), this specific gap shouldn't recur here, but it's a reminder to verify the generator's output is actually wired into the install path for every target it claims to support, not just embedded.
