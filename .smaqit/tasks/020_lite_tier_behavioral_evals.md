# Lite-Tier Behavioral Evals

**Status:** Obsolete — Superseded by Task 026
**Created:** 2026-04-05

## Description

> **Superseded (2026-08-10):** Task 026 replaces this Copilot SDK-specific eval work with the HarnessBench skill and agent evaluation suite. Do not author or repair additional Copilot JSON evals here.

Write behavioral eval files for the lite-tier artifacts: `smaqit.create-agent` (skill + agent) and `smaqit.create-skill` (skill + agent). These are the two routing entries installed by `smaqit-adk lite`. No advanced-tier artifacts are involved.

This establishes the lite eval track as a parallel, independent eval suite that can be run and maintained separately from the advanced-tier evals.

## Scope

| Artifact | Type | Eval coverage |
|---|---|---|
| `smaqit.create-agent` skill | Routing skill | Invokes create-agent agent as subagent; produces `.github/agents/[name].agent.md` |
| `smaqit.create-skill` skill | Routing skill | Invokes create-skill agent as subagent; produces `.github/skills/[name]/SKILL.md` |
| `smaqit.create-agent` agent | Product agent | Gathers spec interactively; writes compiled agent file with no unresolved placeholders |
| `smaqit.create-skill` agent | Product agent | Gathers spec interactively; writes compiled skill file with correct structure |

## Deliverables

New eval files under:
- `tests/evals/skills/smaqit.create-agent/` — at least 2 evals (e.g. happy path, rejection of incomplete input)
- `tests/evals/skills/smaqit.create-skill/` — at least 2 evals
- `tests/evals/agents/smaqit.create-agent/` — at least 1 eval (output quality)
- `tests/evals/agents/smaqit.create-skill/` — at least 1 eval (output quality)

Each eval file follows the existing JSON format (`type`, `artifact_file`, `description`, `turns`, `expected_behavior`, `forbidden_behavior`).

## Acceptance Criteria

- [ ] Eval files authored for `smaqit.create-agent` skill (≥2 evals)
- [ ] Eval files authored for `smaqit.create-skill` skill (≥2 evals)
- [ ] Eval files authored for `smaqit.create-agent` agent (≥1 eval)
- [ ] Eval files authored for `smaqit.create-skill` agent (≥1 eval)
- [ ] All new evals pass `make -C installer evals` (exit 0, no FAIL)
- [ ] Eval runner report written to `tests/evals/runs/<timestamp>/`

## Dependencies

- Task 017 (CLI Tier Subcommands) — completed; lite artifacts are stable

## Notes

- Eval file format reference: `tests/evals/agents/smaqit.L2/001_compile_base_agent.json`
- Routing skills are thin: their evals test that the subagent is invoked and that the output lands in the correct location
- Agent evals test output quality: correct file structure, no unresolved placeholders, correct YAML frontmatter
- This task is independent of Task 021 — can proceed and be run at any time
- **2026-08-05 (Task 024):** the pre-existing `smaqit.new-agent`/`smaqit.new-skill` evals were broken (pointed at renamed, nonexistent artifact files) and have been repaired and repointed to `tests/evals/skills/smaqit.create-agent/` (3 evals) and `tests/evals/skills/smaqit.create-skill/` (2 evals), rewritten to test the current inference-first (single-question) flow instead of the old turn-by-turn gathering flow. This likely already satisfies this task's "≥2 evals" deliverable for both `skill` rows. The two `agent`-type rows (output-quality evals for `tests/evals/agents/smaqit.create-agent/` and `tests/evals/agents/smaqit.create-skill/`) remain entirely unaddressed — that is the real remaining gap for this task. Note also: as of 2026-08-05, running any eval in this repo fails with a Copilot Agent API authorization error unrelated to eval content — see Task 024's Findings for details before assuming a new eval is broken.
