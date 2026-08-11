# Bench Run and Scaffold Skills

**Status:** In Progress
**Created:** 2026-08-11
**Started:** 2026-08-11
**Mode:** Assisted

## Description

Two root-level, ADK-shipped skills installed by `smaqit-adk advanced` (alongside `smaqit.new-principle`), not by `smaqit-adk lite`: `smaqit.bench-run` automates running a project's `.smaqit/bench/` HarnessBench suite end-to-end (preflight → structural validate → live-execution confirmation → execute → report → diagnose-on-failure), and `smaqit.bench-scaffold` guides authoring a new benchmark manifest for a skill or agent that doesn't have one yet, working generically against any consuming project's skill/agent layout and delegating any live trial run back to `smaqit.bench-run` rather than duplicating its logic.

These are shipped product surface — installed into arbitrary consuming projects, not local dev tooling for this repo. That distinction drove two corrections during planning: target/artifact-root detection must be generic (a consuming project's skills/agents live at `.github/skills/`+`.github/agents/`, not smaqit-adk's own root `skills/`/`agents/`), and with-artifact staging must default to a `Given.Files`/`Given.Directories` input asset rather than the `smaqit-adk-dev lite {workspace}` Setup trick used to build Task 026's three manifests — that trick only works because this specific repo has its own dev binary in `installer/dist/`.

Task 026 shipped the underlying `bench suite validate|plan|run` engine capability and three real dogfood manifests (`smaqit.create-agent`, `smaqit.create-skill`, `smaqit.L2`), all live-verified against the authenticated Codex CLI. Three targets currently have no manifest: `smaqit.new-principle`, `smaqit.L0`, `smaqit.L1`.

## Design Decisions

- **Root-level ADK-shipped skills, not `.claude/skills/`:** `skills/smaqit.bench-run/SKILL.md` and `skills/smaqit.bench-scaffold/SKILL.md`, following the same shape as `skills/smaqit.create-agent/`, `skills/smaqit.create-skill/`, `skills/smaqit.new-principle/`. These are not this repo's own dev/project-management tooling (which lives under `.claude/skills/smaqit.task-*`, `smaqit.release-*`, etc.) — they ship to consumers.
- **Advanced tier only:** installed by `smaqit-adk advanced`, not `smaqit-adk lite`. Matches the existing precedent set by `new-principle` — a power-user/framework-maintenance capability, not lite's zero-config create-agent/create-skill framing.
- **Two skills split by workflow (run vs. author), not by Bench's own "variant" concept.**
- **Generic target/artifact-root detection:** `smaqit.bench-scaffold` checks `.github/skills/`+`.github/agents/` first (the standard location `lite`/`advanced` actually install into in a consuming project), falls back to root `skills/`+`agents/` (smaqit-adk's own dev repo convention), and asks the user if neither is unambiguous.
- **Default with-artifact staging via `Given.Files`/`Given.Directories`:** stage the target artifact directly at its conventional destination path (e.g. `.github/skills/<id>/SKILL.md`) as a case-level input asset. The `smaqit-adk-dev lite {workspace}` Setup-based full-install trick from Task 026's manifests is a documented special case for benchmarking smaqit-adk's own repo specifically, not the default a shipped skill can assume works elsewhere.
- **No CI/auto-confirm mode this task:** mirrors Task 026's explicit CI-wiring deferral. `smaqit.bench-run` always gates live execution with an interactive plain-chat confirmation (mirroring `smaqit.release.local`'s pattern, not `smaqit.release.pr`'s unattended one) — no auto-confirm marker exists for this yet, and none is being added here.
- **`smaqit.bench-scaffold` delegates live trials to `smaqit.bench-run`** rather than reimplementing execution/confirmation/diagnosis — same subagent-delegation shape as `smaqit.create-skill` invoking `smaqit.L2`.
- **A discovered Bench engine gap stops the scaffold flow, not a silent patch:** if a live trial reveals a genuine `src/bench` limitation (as `Command.Environment` did in Task 026), `smaqit.bench-scaffold` reports it precisely and recommends a follow-up task rather than patching the engine inline.

## Implementation Steps

**`smaqit.bench-run`**
1. Preflight: confirm `codex` is on PATH (stop with install/auth instructions if not — do not attempt to install it). If this project ships its own `installer/` (smaqit-adk's own repo), run `make build` there so `dist/smaqit-adk-dev`/`dist/validate-skill` are current; otherwise assume the already-installed `smaqit-adk` binary is what's on PATH.
2. Structural validation: `bench suite validate .smaqit/bench` (or a single manifest, if scoped) — stop and report diagnostics on failure, before any live spend.
3. Scope selection: whole suite vs. a single manifest/case/variant — ask if not specified by the caller.
4. Confirmation gate before live execution (plain chat, states manifest count and approximate time/quota cost) — always ask, no auto-confirm path.
5. Execute: `bench suite run .smaqit/bench` (or the narrowed `bench run` invocation), streaming plain events.
6. Report: per-manifest pass/fail/inconclusive outcome, `report.md` paths.
7. On failure, diagnose: read the failing run's grades, trace logs (`traces/harness.std{out,err}.log`), and submission tree; check known gotchas first (Snap/`go run` toolchain conflict, missing `--sandbox`/`--skip-git-repo-check` flags — see `.smaqit/bench/README.md`) before treating anything as a novel bug; offer `bench grade <experiment-dir>` to re-evaluate grading-only issues without spending more quota.

**`smaqit.bench-scaffold`** (depends on `smaqit.bench-run` existing, for step 12)
8. Detect the project's skill/agent root: check `.github/skills/`+`.github/agents/` first; fall back to root `skills/`+`agents/`; ask if ambiguous.
9. Select target: list targets under that root without a `.smaqit/bench/{skills,agents}/<id>/bench.yaml`; if the chosen target already has one, ask whether to add a case to it instead of creating a new manifest.
10. Understand the target: read its SKILL.md/agent.md in full — a judgment step the skill guides but doesn't mechanize, same as `smaqit.create-skill` guiding spec inference.
11. Draft the manifest: self-contained single-shot prompt(s) (Bench's process adapter is non-interactive, no multi-turn `ask_user` relay) with the conditional "if this project has an ADK skill/agent-authoring skill staged at `.github/skills/<id>/SKILL.md`, read it first and follow it exactly" phrasing as a MUST; command-type expectations (not bare `text`) for anything that might not exist; the reusable Codex block (`--sandbox danger-full-access --skip-git-repo-check`, explicit `timeoutSeconds`) copied verbatim from `.smaqit/bench/README.md`; with-artifact staging via `Given.Files`/`Given.Directories` by default.
12. Structural validation (`bench validate`), then offer an optional live trial by invoking `smaqit.bench-run` — reusing its confirm/execute/diagnose logic rather than duplicating it.
13. Report: new manifest path, validation status, live-trial result if run.

**Installer wiring**
14. `installer/main.go`: add `//go:embed` directives for both new skills; extend `cmdAdvanced()` with an install block mirroring the existing `smaqit.new-principle` one, writing to `.github/skills/smaqit.bench-run/` and `.github/skills/smaqit.bench-scaffold/`.
15. `installer/Makefile`'s `prepare` target: add `mkdir`/`cp` lines staging both new skills, mirroring the existing `new-principle`/`create-agent`/`create-skill` lines.
16. `.github/workflows/test-integration.yml`: extend the "Validate advanced-tier ADK structure" and "Verify lite-tier install boundaries" steps with assertions that both skills install under `advanced` and are absent under `lite`, mirroring the existing `new-principle` checks.
17. Update `.github/copilot-instructions.md`'s Skill Catalog table and the top-level `README.md`'s advanced-tier description to list both new skills.

## Known Issues Triage
**Triaged:** 2026-08-11
**Tools searched:** Codex CLI (openai/codex)
**Result:** Advisory

### Blocking Issues
None

### Advisory Issues
- [#33543 Problems with WSL Sandbox](https://github.com/openai/codex/issues/33543) — `openai/codex` — opened 2026-07-16 — bug, windows-os, sandbox, CLI. Attached report describes nested-child-process stalls specifically under the host-managed `workspace-write` sandbox on WSL2 (this machine's platform). Not applicable as filed: every existing and planned Bench manifest pins `--sandbox danger-full-access` (bypasses this sandbox entirely), per `.smaqit/bench/README.md`'s reusable Codex block. Re-check if `smaqit.bench-scaffold` is ever changed to default to a sandboxed mode instead.
- [#36570 exec: approvals_reviewer = "auto_review" silently defeats an explicit --sandbox](https://github.com/openai/codex/issues/36570) — `openai/codex` — opened 2026-08-02 — bug, sandbox, exec, CLI, config. Already known from Task 026; already mitigated by pinning `--sandbox danger-full-access` explicitly in every manifest.

### Historical (Closed)
- [#26723 Codex Desktop on Windows/WSL2 cannot run sandboxed commands in WSL projects](https://github.com/openai/codex/issues/26723) — `openai/codex` — closed (completed). Concerns the Windows Desktop app bridging into WSL2, not the CLI run natively inside a WSL2 shell (this project's actual usage) — different execution context, not directly relevant.

### Unresolvable Tools
None

## Acceptance Criteria

- [ ] `smaqit.bench-run` and `smaqit.bench-scaffold` exist under root `skills/`, pass `skills/smaqit.create-skill/scripts/validate-skill.go`, and are installed by `smaqit-adk advanced` but not `smaqit-adk lite`
- [ ] `smaqit.bench-run` automates preflight → structural validate → confirm → execute → report, never executes a live Codex run without explicit user confirmation, and stops cleanly (no live spend) when `codex` is missing/unauthenticated or a manifest fails structural validation
- [ ] On a failing run, `smaqit.bench-run` distinguishes infra/crash flakes (citing known gotchas) from genuine content failures, using actual grade messages and trace logs as evidence
- [ ] `smaqit.bench-scaffold` detects a project's skill/agent root generically (`.github/skills|agents/` first, root `skills|agents/` fallback) and produces a structurally-valid `bench.yaml` (verified via `bench validate`) for at least one of the three currently-uncovered smaqit-adk targets (`smaqit.new-principle`, `smaqit.L0`, `smaqit.L1`)
- [ ] Manifests produced by `smaqit.bench-scaffold` stage artifacts via `Given.Files`/`Given.Directories` by default (not a repo-specific dev-binary trick), apply the reusable Codex block by construction, and delegate any live trial run to `smaqit.bench-run`
- [ ] `installer/main.go`, `installer/Makefile`, and CI structure checks correctly gate both skills to advanced tier only, verified by installing `advanced` and `lite` into throwaway directories and checking presence/absence
- [ ] At least one live end-to-end demonstration: `smaqit.bench-run` executed against the existing `.smaqit/bench/` suite in this repo reports the real, current outcome

## Findings

[Populated by smaqit.task-complete. Do not fill in manually before task is complete.]

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
| `skills/smaqit.bench-run/SKILL.md` | Create |
| `skills/smaqit.bench-scaffold/SKILL.md` | Create |
| `installer/main.go` | Modify — embed directives + `cmdAdvanced()` install block |
| `installer/Makefile` | Modify — `prepare` target staging lines |
| `.github/workflows/test-integration.yml` | Modify — advanced/lite boundary assertions |
| `.github/copilot-instructions.md` | Modify — Skill Catalog table |
| `README.md` | Modify — advanced-tier description |
| `.smaqit/bench/skills/smaqit.new-principle/` or `.smaqit/bench/agents/smaqit.L0/` or `smaqit.L1/` | Create — first manifest produced via live demonstration of `smaqit.bench-scaffold` |

## Notes

Depends on Task 026 (HarnessBench Skill and Agent Evaluation Suite) being complete — it is (Completed 2026-08-10). `.smaqit/bench/README.md` and `.smaqit/bench/MIGRATION.md` are the canonical references for conventions and gotchas both new skills must apply, not reinvent. How new smaqit product skills currently get mirrored across `.claude/`, `.github/`, `.agents/`, `.codex/` was not investigated during planning (these two skills don't need that mirroring — they're root ADK-shipped, distributed via the installer instead) but is worth confirming during `task.start`'s research step if relevant precedent turns out to matter.
