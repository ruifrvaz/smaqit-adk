# Advanced-Tier Behavioral Evals

**Status:** Obsolete — Superseded by Task 026
**Created:** 2026-04-05

## Description

> **Superseded (2026-08-10):** Task 026 replaces this Copilot SDK-specific eval work with the HarnessBench skill and agent evaluation suite. Do not author or repair additional Copilot JSON evals here.

Rewrite and extend the behavioral eval suite for the advanced-tier artifacts. The existing 7 evals (`smaqit.new-agent`, `smaqit.new-skill`, `smaqit.L2`) were written before Task 018 (Level Skills Completion) and need to be reviewed, aligned, and extended to cover the new skills (`smaqit.new-principle`, `smaqit.new-template`, `smaqit.new-rules`) and updated L0/L1 agents.

## Scope

| Artifact | Type | Current eval coverage | Target coverage |
|---|---|---|---|
| `smaqit.L2` agent | Level agent | 2 evals (compile, reject placeholders) | Review + retain or replace |
| `smaqit.new-agent` skill | Advanced skill | 3 evals | Review + retain or replace |
| `smaqit.new-skill` skill | Advanced skill | 2 evals | Review + retain or replace |
| `smaqit.L0` agent | Level agent | none | ≥1 eval (principle curation) |
| `smaqit.L1` agent | Level agent | none | ≥1 eval (template compilation) |
| `smaqit.new-principle` skill | Advanced skill | none | ≥2 evals |
| `smaqit.new-template` skill | Advanced skill | none | ≥1 eval |
| `smaqit.new-rules` skill | Advanced skill | none | ≥1 eval |

## Deliverables

- Review all 7 existing evals: update, retire, or retain each
- New eval files for `smaqit.L0`, `smaqit.L1`, `smaqit.new-principle`, `smaqit.new-template`, `smaqit.new-rules`
- All evals pass `make -C installer evals`

## Acceptance Criteria

- [ ] All 7 existing evals reviewed and confirmed current or replaced
- [ ] Eval files authored for `smaqit.L0` agent (≥1 eval)
- [ ] Eval files authored for `smaqit.L1` agent (≥1 eval)
- [ ] Eval files authored for `smaqit.new-principle` skill (≥2 evals)
- [ ] Eval files authored for `smaqit.new-template` skill (≥1 eval)
- [ ] Eval files authored for `smaqit.new-rules` skill (≥1 eval)
- [ ] All evals pass `make -C installer evals` (exit 0, no FAIL)

## Dependencies

- Task 018 (Level Skills Completion) — `smaqit.new-principle`, `smaqit.new-template`, `smaqit.new-rules` must exist before evals can be written for them
- Task 019 (Cross-Level Compilation) — optional; `smaqit.compile.*` skills can be covered in a follow-on eval task if needed

## Notes

- Existing evals live in `tests/evals/agents/smaqit.L2/` and `tests/evals/skills/smaqit.new-agent/` and `tests/evals/skills/smaqit.new-skill/`
- Review for staleness: the `smaqit.new-agent` evals simulate a full gathering flow that may have changed with Task 018
- Do not delete existing evals until their replacement is confirmed passing
- This task is blocked until Task 018 delivers stable L0/L1 artifact definitions
