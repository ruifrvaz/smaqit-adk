# Compilation Log: smaqit.new-principle

**Compilation timestamp:** 2026-08-20
**Compiled by:** Agent-L2 (smaqit.L2)
**Pattern:** Pattern 4 — Skill Compilation (3-way merge)
**Nature of change:** Revision (0.2.0 → 0.3.0) of an existing shipped skill — Step 1 (invoke `smaqit.L0`) is functionally unchanged aside from a new capture instruction; Steps 2-5 (live cascade through `smaqit.L1` and `smaqit.L2`, registry query) are new.

## L1 Sources Read

- `agents/smaqit.L2.md` (worktree-local copy) — role, directives, "For Skills" compilation procedure (steps 1-9, including the worktree's new Step 9 "Register" instruction), "Section-Level Compilation" guidance
- `templates/skills/base-skill.template.md` — structural template with `[PLACEHOLDER]` markers (frontmatter + `## Steps`/`## Output`/`## Scope`/`## Examples`/`## Gotchas`/`## Completion`/`## Failure Handling`)
- `templates/skills/compiled/skill.rules.md` — placeholder catalog, Degrees of Freedom directive, Conciseness directive, Reference Structure directive, Base Failure Handling Pattern table, "Compilation Guidance for Agent-L2" (steps 1-9)

All three sources were read from the local worktree copies (`templates/...`, `agents/smaqit.L2.md` relative to the worktree root), per the case instruction that this repository's own `templates/`/`agents/` trees are the ADK source of truth being shipped — not the globally-installed `~/.agents/smaqit-adk/` copies.

## Definition File

- `.smaqit/definitions/skills/smaqit.new-principle.md` (tier: advanced) — Identity (explicitly documents this is a 0.2.0 → 0.3.0 revision), Steps (5, with fragility annotations), Output, Scope, Completion, Failure Scenarios (8 rows), Gotchas (2), Examples (1), Compatibility. All required sections present (identity, steps with fragility levels, output, scope, completion, failure handling); no gaps requiring a stop-and-request.

## Context Consulted (not a compilation input)

- Existing `skills/smaqit.new-principle/SKILL.md` (pre-revision, v0.2.0) — read to understand the skill being revised. Its Step 1 wording (the `smaqit.L0` invocation, including the Claude/Codex-vs-Copilot platform bullets) was carried forward essentially verbatim per the definition's own note that "Step 1 and the overall purpose are unchanged." Its terse, flat-numbered-list style (no `### N. Title` subheadings) and its Failure Handling table convention (skill-specific rows only, no injected base-pattern rows) were preserved as the target style for the revision, per the case instruction.
- `skills/smaqit.new-rules/SKILL.md` and its compilation log, and `.smaqit/definitions/skills/smaqit.new-rules.md` — read as the analogous sibling compilation (same live-cascade family: invoke L-agent → query registry → recompile dependents via `smaqit.L2`). Used to confirm step-fragility rendering conventions (literal `smaqit-adk registry get` commands for the high-fragility registry step; directive prose, not pseudocode, for the medium-fragility cascade step) and to confirm that platform-specific Claude/Codex/Copilot invocation bullets are only written into a step when the *source definition* itself specifies them for that step — `smaqit.new-principle`'s Step 2 (invoking `smaqit.L1`) does not specify Copilot fallback wording the way `smaqit.new-rules`'s Step 1 (invoking `smaqit.L1` as its *primary* agent) does, so no such bullets were invented for Step 2 here.

## Merge Process Summary

- `[SKILL_NAME]` → `smaqit.new-principle` (from Identity; unchanged from existing skill)
- `[SKILL_DESCRIPTION]` → definition's draft description used verbatim: "Adds or refines a principle in an ADK framework file, then cascades the change through smaqit.L1 (template/rules review) and smaqit.L2 (dependent recompilation) so every registered dependent stays current automatically." Verified capability-oriented, present-tense, declarative, states what the skill produces, no conversational/user-centric gating, no first person, no can/may/might/will. Backticks around agent names dropped to match the plain-scalar frontmatter style already established by the existing compiled skill and by `smaqit.new-rules`.
- `[SKILL_VERSION]` → `"0.3.0"` (from Identity — revision bump from the existing `"0.2.0"`)
- `[SKILL_TITLE]` → "New Principle" (unchanged from existing skill)
- `[STEPS_CONTENT]` → definition's 5 steps, degrees-of-freedom applied per fragility marking:
  - Step 1 (low fragility, unchanged) — carried forward from the existing skill's Step 1 near-verbatim (Claude/Codex vs. Copilot invocation bullets preserved exactly), with the definition's new closing sentence appended: capture `smaqit.L0`'s reported scope and the exact framework file path it changed, since Step 2 needs both.
  - Step 2 (medium fragility — a judgment call `smaqit.L1` makes, not this skill) — rendered as directive prose: the single-hop `Task` call constraint, the exact quoted message to pass (kept verbatim, including the four candidate file globs), the "no update needed is valid, not a failure" clarification, and the path-capture instruction that gates Steps 3-4.
  - Step 3 (high fragility — wrong scope or a skipped project-local check silently drops real dependents) — rendered with the literal `smaqit-adk registry get --scope base --depends-on <path>` and `--scope extension --depends-on <path>` commands preserved verbatim, no paraphrasing of the ADK-global-vs-project-local distinction or the "missing registry file = zero dependents" rule.
  - Step 4 (medium fragility) — kept as directive prose describing the pattern (single-hop `Task` call, mirrors `smaqit.development`'s orchestration) and the explicit skip condition, without inventing an unstated "not chained from `smaqit.L1`" clarification the source definition does not include for this particular step (unlike Step 2, which does state the analogous L0→L1 constraint explicitly).
  - Step 5 (low fragility) — single prose reporting instruction.
- `[OUTPUT_CONTENT]` → definition Output section, used verbatim — already states each artifact, which subagent produced it, and that the cascade summary covers all of it.
- `[SCOPE_CONTENT]` → definition Scope section, used verbatim, including both "Out of scope" redirects (`smaqit.new-template`/`smaqit.new-rules` for non-principle authoring; `smaqit.create-agent`/`smaqit.create-skill` for agent/skill creation).
- `[EXAMPLES_CONTENT]` → definition's single example, verbatim. (Not enumerated as a fill step in this worktree's `agents/smaqit.L2.md` Step 5 bullet list — an apparent gap versus `skill.rules.md`'s own placeholder catalog and Compilation Guidance, which both define `[EXAMPLES_CONTENT]`/`[GOTCHAS_CONTENT]`. Filled anyway per the template's own structural requirement that "all sections present" and per precedent from both sibling compilations, which filled these sections despite the same gap in the agent file's Step 5 text.)
- `[GOTCHAS_CONTENT]` → definition's 2 gotchas, verbatim (already concise; minor punctuation normalization only — em-dash-joined clauses split into plain sentences where the source used "Steps 3–4" en-dash notation, normalized to ASCII "3-4" for consistency with the rest of the compiled body).
- `[COMPLETION_CONTENT]` → definition Completion checklist, verbatim (5 items).
- `[FAILURE_HANDLING_CONTENT]` → see Failure Handling Merge below.
- `compatibility` frontmatter → included (definition specifies: "Requires the `smaqit-adk` binary on `PATH` for the registry subcommands used in Step 3."), backticks dropped to match plain-scalar YAML style, flat top-level field placed after `metadata` (matching `smaqit.new-rules`' same-day compiled form).
- `allowed-tools` frontmatter → omitted; definition does not specify a value.

## Failure Handling Merge

Per the case's explicit steer: favor conciseness over completeness for this skill, consistent with its existing terse style, rather than performing the full base-pattern merge-with-dedupe that both `smaqit.new-rules` and `smaqit.new-template` performed the same day.

Compared skill.rules.md's 4-row Base Failure Handling Pattern against the definition's 8 Failure Scenarios rows:

- Base "Subagent invocation fails" overlaps the definition's own "Subagent invocation fails (any of L0/L1/L2)" row — identical action text ("Report the failure with context; do not silently retry"), and the definition's row is already the more specific, more useful version (names all three subagents in the cascade). No separate base row added; the definition's row stands alone.
- Base "Required input not provided," "Gathered input is ambiguous," and "Output artifact already exists" have no meaningful skill-specific counterpart in this skill's own failure surface (this skill sequences three subagent calls and a registry query; it does not itself gather ambiguous input from the user or produce an overwritable artifact — `smaqit.L0`/`smaqit.L1`/`smaqit.L2` own those concerns internally, and each already has its own compiled Failure Handling for them). Omitted, matching the existing (pre-revision) `smaqit.new-principle` skill's own established convention of a skill-specific-only Failure Handling table.

Result: the definition's 8 rows, used verbatim, no base-pattern rows injected. This intentionally diverges from the `smaqit.new-rules`/`smaqit.new-template` precedent (full merge) in favor of the older, still-live `smaqit.new-principle` precedent (terse, skill-specific-only) — per this task's explicit instruction to use judgment favoring this skill family's existing convention over literal completeness.

## Conciseness Filter

- Did not invent a "not chained from `smaqit.L1`" architecture clarification for Step 4 that the source definition does not state for that step (present only for Step 2's L0→L1 relationship in the source) — avoids adding content beyond what's traceable to the definition.
- Normalized the definition's en-dash "Steps 3–4" notation to ASCII "Steps 3-4" throughout the compiled body for consistency with the existing skill's plain-ASCII style.
- No other trims were needed; the definition's prose was already concise and every remaining sentence (exact registry commands, the quoted `smaqit.L1` message, the `smaqit.development` orchestration analogy, the scope/path-capture instructions that gate later steps) carries information the executing agent could not safely infer.

## Progressive Disclosure

Not needed — compiled body is 83 lines, well under the 400-line threshold. No `references/` subdirectory created.

## Registration (Step 9 of the worktree's `agents/smaqit.L2.md`)

Not performed. The worktree's `agents/smaqit.L2.md` (uncommitted, part of this same task) adds a new Step 9 instructing a `smaqit-adk registry put` shell-out after every successful compile. The globally installed `smaqit-adk` binary on `PATH` (`~/.local/bin/smaqit-adk`, v2.0.1) predates this feature and has no `registry` subcommand — only the worktree-local, uncommitted `installer/registry.go`/`installer/registry_cli.go` implement it, and that installer has not been rebuilt/reinstalled as part of this compilation task. Both sibling same-day compilations (`smaqit.new-rules`, `smaqit.new-template`) likewise performed no registration step and recorded none in their logs. Followed the same precedent here rather than building and invoking a work-in-progress binary outside this task's stated scope (recompile one skill per the explicit checklist: confirm definition, 3-way merge, conciseness filter, validate placeholders/nesting).

## Validation Checklist

- [x] No unresolved `[PLACEHOLDER]`-style uppercase-bracket tokens (`grep -nE '\[[A-Z_]+\]' skills/smaqit.new-principle/SKILL.md` → no matches)
- [x] Frontmatter `name` field exactly `smaqit.new-principle`
- [x] Description is capability-oriented, declarative, present-tense; no "Use when...", no first person, no can/may/might/will
- [x] All required sections present, in template order: `## Steps`, `## Output`, `## Scope`, `## Examples`, `## Gotchas`, `## Completion`, `## Failure Handling`
- [x] `compatibility` frontmatter field present (sourced from definition); `allowed-tools` omitted (not specified)
- [x] No nested reference chains — skill body references `~/.claude/agents/smaqit-L0.md` (an external agent file, read inline for one turn on Copilot only) but does not reference any file bundled within its own skill directory, so no `SKILL.md` → file → file chain exists
- [x] Body under 400 lines (83 lines)
- [x] Third-party validator (`~/.agents/skills/smaqit.create-skill/scripts/validate-skill.go`, run via `go run`) passes with zero violations
- [x] Re-ran the validator against all 7 skills in `skills/` as a sanity check — all pass, no regressions introduced

## Decisions Made

1. Diverged from the same-day `smaqit.new-rules`/`smaqit.new-template` precedent of a full base-pattern Failure Handling merge, instead keeping this skill's table to its 8 definition-sourced rows only — per this task's explicit instruction to favor conciseness and this skill family's own existing (pre-revision) terse-table convention over skill.rules.md's literal "insert the base pattern as the foundation" instruction.
2. Did not add Claude/Codex/Copilot platform-invocation bullets to Step 2 (`smaqit.L1` invocation) even though the analogous primary-invocation step in the sibling `smaqit.new-rules` skill has them — the `smaqit.new-principle` definition only specifies that bullet structure for Step 1's `smaqit.L0` invocation, not for Step 2, so none was invented for Step 2.
3. Filled `[EXAMPLES_CONTENT]` and `[GOTCHAS_CONTENT]` despite this worktree's `agents/smaqit.L2.md` Step 5 bullet list not explicitly enumerating those two placeholders (an apparent gap against `skill.rules.md`'s own placeholder catalog and Compilation Guidance, and against the base template's own section structure) — followed the template's structural requirement and both sibling compilations' precedent rather than the narrower literal bullet list.
4. Did not perform the worktree `agents/smaqit.L2.md`'s new Step 9 "Register" instruction, since the registry feature it depends on is uncommitted, worktree-local, and not present in the installed `smaqit-adk` binary — matching the same-day precedent set by both sibling compilations, which also skipped it without comment.
