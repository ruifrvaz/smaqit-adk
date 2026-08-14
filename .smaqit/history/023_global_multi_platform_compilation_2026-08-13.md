# Global Multi-Platform Compilation

**Date:** 2026-08-13
**Session focus:** Plan, implement, live-verify, and complete Task 027 — a straight install-path migration that ballooned during planning into a full platform-strategy pivot (global installation, GitHub Copilot dropped as an authored target, Claude Code + Codex CLI added as compiled platforms). Plan and create a direct follow-on task (029) for the dogfood Bench manifests it broke.
**Tasks completed/referenced:** 027 completed (planned, started, and finished in this session); 029 created (planning only, not started).

## Actions Taken

- Ran `session.start`: surfaced Task 027 as the user's explicit top priority (per prior session's memory), plus Tasks 025/018/019 as lower-priority active work.
- Ran `task.plan 027`. What began as a straightforward port of smaqit-extensions' global-install migration (Copilot-only, tier/skills-path questions open) grew through several distinct rounds of user-directed scope expansion, each investigated and confirmed rather than assumed:
  1. Resolved the original open questions: tier gating dropped entirely, skills path resolved to `~/.agents/skills/`, templates/framework also moved global — all confirmed via direct VS Code/Codex documentation checks.
  2. Added Codex CLI TOML custom-agent support after confirming it's real and already shipped in both sibling repos (`smaqit-extensions`, `smaqit`).
  3. User directed full GitHub Copilot deprioritization. Investigated and confirmed this is a *new* decision, not inherited precedent — both sibling repos keep Copilot fully first-class; `smaqit`'s own Task 088 explicitly preserves Copilot when Codex was added. Saved as a dedicated project memory given how sharply it diverges from established family precedent.
  4. Corrected an initial assumption that Copilot would lose the L0/L1/L2 workflow entirely — resolved via this repo's own "Skill-Mediated Workflows" principle: skills (cross-platform) route via native subagent call on Claude/Codex, or read-and-follow-inline fallback on Copilot.
  5. User expanded scope to include the compiler's OUTPUT format for end-user-created agents, not just the ADK's own shipped artifacts — adopted `smaqit`'s validated `generate-agents.py` split-source pattern (shared body + per-platform metadata) after confirming it was arrived at only after two rejected simpler approaches.
  6. Corrected an "omit tools field, inherit everything" shortcut in favor of `smaqit`'s actual validated explicit per-platform tool-mapping table.
- Ran `task.start 027`: created the task branch/worktree, ran issue triage (Result: Advisory — most notable, `openai/codex#26408`/`#26868` around `.codex/agents/*.toml` spawn reliability, directly informing later live-verification choices).
- Implemented Task 027 end-to-end: global path resolution in `installer/main.go` (replacing `lite`/`advanced`/`init` with a hidden `--install-global`); split L0/L1/L2 into shared body + per-platform metadata rendered by a new Go generator (`installer/generate-agents.go`, mirroring `smaqit`'s pattern); rewrote `smaqit.L2`'s compile procedures and the `create-agent`/`create-skill`/`new-principle` skills for platform-conditional routing and dual-format output; repositioned `README.md`, made `AGENTS.md` canonical (deleted `.github/copilot-instructions.md`); rewrote `docs/wiki/agent-frontmatter.md` and `extending-smaqit.md`; deleted `bench-scaffold`'s dead `.github/`-first detection branch; rewrote the test suite (`global_install_test.go` replacing `lite_test.go`, `embed_test.go`, `init_test.go`, the CI workflow, the `test-scaffold` Makefile target) with every test sandboxing `$HOME`.
- Live-verified repeatedly against sandboxed `$HOME`: install/uninstall cycles, idempotency, real TOML/YAML parseability, and — critically — that `uninstall` is surgical (planted fake competing-product content first, confirmed it survived untouched).
- Made and immediately caught a real mistake: one manual verification command forgot to sandbox `$HOME` and wrote real files into this machine's actual `~/.claude`/`~/.codex`/`~/.agents`; cleaned up immediately via the just-built `uninstall` command, which doubled as a real-world confidence check.
- On explicit request ("complete task"), audited all 12 acceptance criteria honestly rather than checking boxes reflexively: found 2 explicitly required live/release-based proof that hadn't been done. Surfaced this directly to the user rather than silently completing. For AC7 (live routing verification), got explicit approval to test against the real machine; a real authenticated `codex exec` successfully spawned the globally-installed `smaqit.L2` custom agent, which correctly reported its own real title line — then cleaned up and reverified the machine was clean again. AC10 (real `curl | bash` from a cut release) was left explicitly unmet — cutting a release is a deliberate action outside a task's scope, documented as follow-up instead.
- Completed Task 027: Findings written honestly (including the deferred ACs), merged to `main` (`--no-ff`, no conflicts), worktree/branch cleaned up, `PLANNING.md` updated — including a note that Task 025 (README's fictional CLI commands) looks resolved as a side effect of the README rewrite, without unilaterally marking a different task complete.
- Ran `task.plan` (Mode A) for the dogfood Bench manifests Task 027 broke. Did Discovery via direct file reads (not Explore subagents, since the area was small and context was already fresh) — found the 3 `lite`-dependent manifests were always a documented special case relative to `.smaqit/bench/README.md`'s own canonical `given.directories`/`{input:<id>}` pattern, and resolved a real design question (how to test skill→L2 routing without mutating the real machine's global Codex/Claude config) by reusing the Copilot inline-fallback pattern Task 027 already built. Created Task 029 on approval.

## Problems Solved

- Task 027's own file's Design Decisions and Implementation Steps were fully rewritten three separate times across the session as scope kept expanding — each rewrite preserved a "scope progression" log in the task file's Notes for traceability, rather than silently overwriting prior reasoning.
- The Go generator's TOML output initially failed its own safety check (a literal `'''` sequence in L2's own documentation of the TOML format broke the delimiter) — caught by the generator's own guard before it could ship broken output, fixed by rewording rather than disabling the check.
- `cmdUninstall`'s initial design always showed a removal list and prompted even when nothing was installed — caught via a stale test assumption, redesigned to check existence first.
- `bench suite validate` against `.smaqit/bench/` surfaced that Task 027 broke its own dogfood suite (1 structural failure, 3 live-execution failures) — not silently ignored; documented in the compendium as known-stale and turned into Task 029.

## Decisions Made

- GitHub Copilot deprioritized as an authored compilation target — a fresh, session-specific decision for `smaqit-adk` only, explicitly not shared by `smaqit`/`smaqit-extensions` (saved to memory to prevent future conflation).
- Multi-format generation follows `smaqit`'s validated split-source pattern exactly (shared body + per-platform metadata → deterministic render), implemented in Go rather than Python to avoid a second toolchain dependency.
- No user-facing `install` subcommand and no user-callable `--install-global` — `install.sh` is the sole global-install entry point, matching the corrected design smaqit-extensions arrived at only after shipping and reverting a worse one.
- Bench manifests must never perform a real global install to test routing — L2 gets staged and read inline instead, reusing the Copilot-fallback mechanism, keeping the dogfood suite side-effect-free on the real machine.
- Task 027 completed with one AC (real `curl | bash` post-release) explicitly left unmet and documented rather than blocked on or silently faked — cutting a release is a separate, deliberate action.

## Files Modified

- `installer/main.go`, `installer/generate-agents.go` (new), `installer/Makefile`, `.gitignore` — global path resolution, `installGlobal()`, tier-free `cmdUninstall()`, the multi-format generator wiring.
- `agents/smaqit.L0.md`, `smaqit.L1.md`, `smaqit.L2.md` (renamed from `.agent.md`) — shared platform-neutral bodies; `.smaqit/definitions/agents/*.frontmatter.yaml` (new) — per-platform metadata.
- `README.md`, `AGENTS.md` (new, root), `.github/copilot-instructions.md` (deleted) — repositioning and canonical instructions.
- `docs/wiki/agent-frontmatter.md`, `docs/wiki/extending-smaqit.md`, `.smaqit/compendium.md` — multi-platform reference content, stale entries corrected.
- `skills/smaqit.create-agent/SKILL.md`, `smaqit.create-skill/SKILL.md`, `smaqit.new-principle/SKILL.md`, `smaqit.bench-scaffold/SKILL.md` — platform-conditional routing, path fixes, dead branch removed.
- `install.sh` — `--install-global` trigger wiring.
- `tests/unit/global_install_test.go` (new, replaces deleted `lite_test.go`), `embed_test.go`, `init_test.go`, `.github/workflows/test-integration.yml` — full test rewrite, all `$HOME`-sandboxed.
- `CHANGELOG.md` — Unreleased entry documenting the breaking change.
- `.smaqit/tasks/027_migrate_to_global_user_level_installation.md`, `.smaqit/tasks/029_repair_dogfood_bench_manifests.md` (new), `.smaqit/tasks/PLANNING.md` — task lifecycle state for both tasks.

## Next Steps

- **Task 029 (Repair Dogfood Bench Manifests) is ready to start** — full per-manifest breakdown already worked out in the task file.
- Task 025 (README's fictional CLI commands) is likely already resolved as a side effect of Task 027's README rewrite — worth a quick verification pass and a formal `task.complete 025` rather than independent rework.
- The real `curl | bash` end-to-end install test (Task 027's AC10) still needs to happen immediately after the next release is cut.
- Consider a follow-up live Claude Code subagent spawn test to close the one remaining asymmetry in Task 027's live verification (Codex was tested for real; Claude relied on structural verification of the compiled file plus the platform's well-established native subagent mechanism).

## Session Metrics

- Tasks completed: 1 (027, planned/started/completed same session, largest single task this session has seen — 30 files changed)
- Tasks created: 1 (029, planning only)
- Real, live external verifications: 1 authenticated `codex exec` call successfully spawning a globally-installed custom agent (with explicit user approval, fully reversed afterward)
- Real mistakes made and self-caught: 1 (unsandboxed `$HOME` write to the real machine, cleaned up immediately)
- Acceptance criteria: 11/12 met and verified for Task 027; 1 explicitly and transparently left unmet (requires a cut release)
- Scope-expansion rounds during planning: 5 distinct, user-directed, each independently investigated rather than assumed
