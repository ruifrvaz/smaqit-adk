# Repair Dogfood Bench Manifests Broken by Task 027

**Status:** Not Started
**Created:** 2026-08-13

## Description

Task 027's removal of `lite` and the Copilot `.agent.md` format broke 4 dogfood Bench manifests under `.smaqit/bench/` — one fails structural validation outright, three would fail live execution. Migrate all 4 onto the generic `given.files`/`given.directories` + `{input:<id>}` staging pattern the README already documents as canonical (rather than the special-cased `lite`-install trick three of them used), update expectations for dual Claude/Codex output, and re-verify structurally and live.

The 3 manifests using the `smaqit-adk-dev lite {workspace}` trick (`agents/smaqit.L2`, `skills/smaqit.create-agent`, `skills/smaqit.create-skill`) were always a documented special case — `.smaqit/bench/README.md`'s own canonical with/without-artifact example already shows the generic `given.directories` + `{input:<id>}` pattern instead. `smaqit.new-principle`'s manifest already follows that canonical pattern and needs only a one-line path fix. Since `lite` (a project-scoped install) has no replacement — global install only ever targets `$HOME` — the right fix isn't swapping the command, it's bringing all 3 manifests onto the pattern the other one and the README already establish.

A second, real design consideration: `create-agent`/`create-skill`'s skills now say "invoke `smaqit.L2` as a native subagent." A Bench manifest that did a real global install to make `smaqit.L2` spawnable would mutate the real machine's `~/.codex/agents/`/`~/.claude/agents/` on every run, cutting against Bench's disposable-workspace philosophy (and needs real credentials to authenticate, confirmed the hard way during Task 027). The resolution: reuse the read-and-follow-inline fallback Task 027 already designed for Copilot — stage L2's body alongside the routing skill, and tell Codex explicitly that L2 isn't registered as a spawnable agent in this sandbox, so it should read `{input:l2}` directly and apply it. `new-principle`'s manifest already does exactly this for L0, just with a stale file path — proof the pattern works, not something new to invent. Task 027's own live `codex exec` test already verified the *real* global-registration-and-spawn mechanism separately; this task doesn't need to re-prove that.

## Design Decisions

- **Unify all 4 manifests onto one staging philosophy** (`given.files`/`given.directories` + explicit `{input:<id>}` prompting) — no manifest performs a real global install or touches the developer machine's actual `~/.claude` or `~/.codex`.
- **L2 is never spawned as a real registered custom agent inside Bench tests.** Its body is staged and read inline instead, reusing the Copilot-fallback pattern from Task 027. The real spawn mechanism is considered separately verified (Task 027's live `codex exec` test), not this task's job to re-prove.
- **Both Claude `.md` and Codex `.toml` outputs get checked** in every `expect` block that previously checked one Copilot file, since L2 now always produces both.
- **In-scope minor fix:** `new-principle`'s prompt currently claims `codex exec` has "no subagent-invocation capability" — Task 027's live test proved this false (spawn_agent works non-interactively). Correct this while touching the file.

## Implementation Steps

1. Fix `smaqit.new-principle`'s stale `l0agent` source path (`agents/smaqit.L0.agent.md` → `agents/smaqit.L0.md`); correct the "no subagent-invocation capability" claim in its prompt while there.
2. Migrate `agents/smaqit.L2/bench.yaml` off the `lite`-install trick onto `given.files` staging (the base-agent template + base rules, mirroring how `new-principle` stages `framework/`); update both cases' `expect` blocks from `agents/*.agent.md` to `.claude/agents/*.md` + `.codex/agents/*.toml`.
3. Update `agents/smaqit.L2/prompts/compile-base-agent.md` and `prompts/reject-unresolved-placeholders.md` to reference `{input:<id>}` explicitly for staged guidance, and name both new output paths in the "compile to" instruction.
4. Migrate `skills/smaqit.create-agent/bench.yaml` off the `lite`-install trick onto `given.directories` staging `skills/smaqit.create-agent` (matches the README's own canonical example) plus `agents/smaqit.L2.md` for the inline-fallback pattern; update `expect` to `.claude/agents/qa-helper.md` + `.codex/agents/qa-helper.toml`.
5. Update `skills/smaqit.create-agent`'s inline prompt (`given.prompt.text`) to reference `{input:skill}/SKILL.md` instead of the stale `.github/skills/...` path, and add the L2 inline-fallback instruction (mirroring `new-principle`'s existing wording for L0).
6. Migrate `skills/smaqit.create-skill/bench.yaml` the same way: `given.directories` for `skills/smaqit.create-skill` + `agents/smaqit.L2.md`; `expect` updated to `.agents/skills/smaqit.new-principle/SKILL.md` (and/or `.claude/skills/...`); validator command path updated to match.
7. Update `skills/smaqit.create-skill`'s inline prompt with the same `{input:skill}` + L2 inline-fallback treatment as step 5.
8. Run `smaqit-adk bench suite validate .smaqit/bench` — fix any remaining structural issues until `valid=true` for all 4 manifests.
9. Run `smaqit-adk bench suite run .smaqit/bench` (live, authenticated `codex exec`, real cost) — confirm pass/fail/inconclusive per manifest; diagnose any failures against known Bench gotchas (`.smaqit/bench/README.md`) before treating them as new bugs.
10. Update `.smaqit/compendium.md`'s "known stale as of Task 027" note (under "Where does smaqit-adk's own dogfood benchmark suite live") to reflect the fix.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] `smaqit-adk bench suite validate .smaqit/bench` reports `valid=true` for all 4 manifests
- [ ] No manifest references `.github/`, the `lite` command, or `.agent.md` output anywhere
- [ ] Live `smaqit-adk bench suite run .smaqit/bench` executes without harness-level errors (pass/fail/inconclusive results are acceptable outcomes; crashes or timeouts are not)
- [ ] `.smaqit/compendium.md`'s "known stale" note about these manifests is updated to reflect the fix

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
| `.smaqit/bench/agents/smaqit.L2/bench.yaml` | Modify — migrate off `lite` to `given.files` staging, dual-output expectations |
| `.smaqit/bench/agents/smaqit.L2/prompts/compile-base-agent.md` | Modify — `{input:<id>}` references, dual output paths |
| `.smaqit/bench/agents/smaqit.L2/prompts/reject-unresolved-placeholders.md` | Modify — dual output paths |
| `.smaqit/bench/skills/smaqit.create-agent/bench.yaml` | Modify — migrate off `lite` to `given.directories` staging + L2 inline-fallback, dual-output expectations |
| `.smaqit/bench/skills/smaqit.create-skill/bench.yaml` | Modify — same treatment, plus validator command path |
| `.smaqit/bench/skills/smaqit.new-principle/bench.yaml` | Modify — fix stale `l0agent` source path |
| `.smaqit/compendium.md` | Modify — update the Task-027-stale note once fixed |

## Notes

Source: follow-up item #2 from Task 027's own Findings (`.smaqit/tasks/027_migrate_to_global_user_level_installation.md`), planned via `smaqit.task-plan` in the same session Task 027 completed. Discovery for this plan was done by direct file reads (all 4 manifests, their prompts, and `.smaqit/bench/README.md`) rather than Explore subagents — the area is small and context was already fresh from implementing Task 027 itself.
