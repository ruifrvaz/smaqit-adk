# Compilation Log: smaqit.sync

**Compilation timestamp:** 2026-08-20
**Compiled by:** Agent-L2 (smaqit.L2)
**Pattern:** Pattern 4 — Skill Compilation (3-way merge)

## L1 Sources Read

- `agents/smaqit.L2.md` — role, directives, "For Skills" compilation procedure (steps 1-9, including the worktree's Step 9 "Register" instruction), "Section-Level Compilation" guidance
- `templates/skills/base-skill.template.md` — structural template with `[PLACEHOLDER]` markers (frontmatter + `## Steps`/`## Output`/`## Scope`/`## Examples`/`## Gotchas`/`## Completion`/`## Failure Handling`)
- `templates/skills/compiled/skill.rules.md` — placeholder catalog, Degrees of Freedom directive, Conciseness directive, Reference Structure directive, Base Failure Handling Pattern table, "Compilation Guidance for Agent-L2" (steps 1-9)

All three sources were read from the local worktree copies (`templates/...` relative to the worktree root), per the case instruction that this repository's own `templates/` tree is the ADK source of truth being shipped — not the globally-installed `~/.agents/smaqit-adk/` copies.

## Definition File

- `.smaqit/definitions/skills/smaqit.sync.md` (tier: advanced) — Identity, Steps (with fragility annotations), Output, Scope, Completion, Failure Scenarios, Gotchas, Examples, Compatibility. All required sections present (identity, steps with fragility levels, output, scope, completion, failure handling); no gaps requiring a stop-and-request.

## Sibling Review

Skimmed `skills/smaqit.new-template/SKILL.md` (and, for style corroboration, `skills/smaqit.new-rules/SKILL.md` and `skills/smaqit.new-principle/SKILL.md`, all compiled earlier the same day in this worktree) for this family's established tone and structural conventions only — not for step content, which was compiled strictly from `smaqit.sync`'s own definition file. Conventions carried forward:

- Frontmatter: plain-scalar `description` (no backticks), `metadata.version` as a quoted string, `compatibility` as a flat top-level field with backticks dropped, placed immediately after `metadata`.
- `# <Title>` derived by dropping the `smaqit.` prefix and title-casing the remainder (`new-template` → "New Template", `new-rules` → "New Rules"; here, `sync` → "Sync").
- Section order matching the template literally: Steps, Output, Scope, Examples, Gotchas, Completion, Failure Handling.
- Steps rendered as a flat numbered list (`1.` … `4.`) with bold sub-labels and fenced code blocks for literal commands — the convention used by `smaqit.new-rules` and `smaqit.new-principle` (as opposed to `smaqit.new-template`'s `### N. Title` heading variant, which the `smaqit.new-template` log itself flagged as an unresolved stylistic divergence, not a mandate).
- Failure Handling table: insert the 4-row base pattern from skill.rules.md as foundation rows, then append the definition's own failure scenarios, deduping only rows that are textually identical in both situation and action wording (per the precedent set in the `smaqit.new-template` log, which deduped exactly one such row and kept everything else).

## Merge Process Summary

- `[SKILL_NAME]` → `smaqit.sync` (from Identity)
- `[SKILL_DESCRIPTION]` → definition's draft description rewritten into the compiled form: "Reconciles registered agents and skills against their compiled output, detecting drift where a target was hand-edited outside the L0 → L1 → L2 compilation chain, and presenting each drifted entry for a recompile-from-source or keep-and-re-register resolution. The interactive counterpart to the non-interactive smaqit-adk sync CLI diagnostic, which reports drift but has no path to invoke an LLM for the semantic recompilation step." Two changes from the draft: (1) the draft's "lets the user choose to recompile ... or keep the hand edit and re-register it" was reworded to "presenting each drifted entry for a recompile-from-source or keep-and-re-register resolution" — "lets the user" reads as the same conversational/user-centric gating pattern the rules file explicitly bans ("Allows the user to..."), so it was replaced with a capability-oriented, present-tense phrasing; (2) the draft's tight `L0→L1→L2` arrow notation was respaced to `L0 → L1 → L2` to match the only existing in-repo precedent for this notation (`README.md`), since no shipped `SKILL.md` uses an arrow at all and the README's spaced form was the closest available convention. Verified: capability-oriented, present-tense, declarative; states what the skill produces; no "Use when...", no first person, no can/may/might/will. Backticks dropped from `smaqit-adk sync` and `LLM` in the frontmatter description to match the plain-scalar YAML style established by the three same-day sibling compilations.
- `[SKILL_VERSION]` → `"0.1.0"` (from Identity — the definition's stated version used directly, since the "default 1.0.0" fallback in skill.rules.md's Compilation Guidance only applies when a version is unspecified)
- `[SKILL_TITLE]` → "Sync"
- `[STEPS_CONTENT]` → definition's 4 steps, degrees-of-freedom applied per fragility marking: Steps 1-3 (all marked high fragility in the definition) rendered with maximal literalness — the exact `smaqit-adk registry get --scope base|extension` commands preserved verbatim in fenced code blocks; the drift-detection rule (hash the `claude` path, or `codex` if `claude` is empty; missing target file also counts as drift) carried through without paraphrase; and Step 3's two resolution options (recompile vs. keep-and-re-register) kept as an explicit bulleted choice with the "never pick automatically / never silently overwrite" constraint preserved as its own sentence, since this is the one behavior the definition itself calls out as the skill's entire reason to exist. Step 4 (low fragility) compressed to a single prose sentence. Parenthetical fragility labels themselves were stripped (compile-time signal for Agent-L2, not runtime instruction for the executing agent).
- `[OUTPUT_CONTENT]` → definition Output section, carried through verbatim as three bullets (already concise, no trimming needed)
- `[SCOPE_CONTENT]` → definition Scope section, condensed to one paragraph naming the boundary against `smaqit.L2` (registration) and the registry format, plus the non-automatic-invocation condition
- `[EXAMPLES_CONTENT]` → definition's single example, carried through verbatim (already concrete: a specific CLI output, a specific user choice, a specific reported result)
- `[GOTCHAS_CONTENT]` → definition's 2 gotchas, verbatim (already concise, no trimming needed)
- `[COMPLETION_CONTENT]` → definition Completion checklist, verbatim
- `[FAILURE_HANDLING_CONTENT]` → see Failure Handling Merge below
- `compatibility` frontmatter → included (definition specifies: "Requires the `smaqit-adk` binary on `PATH` for the registry subcommands, and `sha256sum` (or an equivalent SHA-256 utility) for hashing target files."), backticks dropped to match plain-scalar YAML style, flat top-level field placed after `metadata`
- `allowed-tools` frontmatter → omitted; definition does not specify a value

## Failure Handling Merge

Base Failure Handling Pattern (4 rows, from skill.rules.md) merged against the definition's 5 Failure Scenarios rows:

- Checked each definition row against each base row for textual identity (situation wording AND action wording both matching, the bar the `smaqit.new-template` precedent used). None matched: the closest pair — base's generic "Subagent invocation fails" / "Report the failure with context; do not silently retry" versus the definition's "`smaqit.L2` recompile fails" / "Report the failure; leave the entry drifted and unresolved; do not fall back to re-registering the hand edit without asking" — differs in both situation specificity and action detail (the definition's action encodes this skill's core no-silent-overwrite guarantee, which the generic base row does not). Kept both rather than deduping, since collapsing them would lose the skill-specific constraint.
- Base's "Output artifact already exists" / "Confirm with user before overwriting" also has no exact definition counterpart, but remains genuinely applicable as a foundation row: it is the general form of the specific guarantee Step 3 already implements (never silently overwrite a hand edit) — kept per skill.rules.md's literal "insert the following as the foundation" instruction, matching the same judgment call the `smaqit.new-template` log made for this identical row.
- All 5 definition rows (no registry file in any scope; a drifted entry's source paths gone; `smaqit.L2` recompile fails; `smaqit-adk` binary missing; user declines to resolve) had no base overlap — carried through unchanged.

Result: 9-row single coherent table (4 base + 5 definition, 0 dedupes), header `Situation | Action` (definition used `Situation | Response`, normalized to `Action` to match the base table's header, mirroring the same normalization the `smaqit.new-template` log recorded).

## Conciseness Filter

Removed the definition's inline parenthetical fragility labels ("(high fragility — ...)", "(low fragility)") from the compiled Steps — compile-time signal for Agent-L2, not runtime instruction for the executing agent. No other trims were needed: the definition's prose was already concise, and every remaining sentence (the exact registry commands, the hash-target-selection rule, the two-option conflict presentation, the never-automatic/never-silent constraint) carries information the executing agent could not safely infer on its own.

## Progressive Disclosure

Not needed — compiled body is 77 lines, well under the 400-line threshold. No `references/` subdirectory created.

## Registration (Step 9 of the worktree's `agents/smaqit.L2.md`)

Not performed. The worktree's `agents/smaqit.L2.md` (uncommitted, part of this same task) adds a Step 9 instructing a `smaqit-adk registry put` shell-out after every successful compile. Confirmed the globally installed `smaqit-adk` binary on `PATH` (`~/.local/bin/smaqit-adk`, v2.0.1) has no `registry` subcommand (`smaqit-adk registry --help` returns the top-level usage listing only `bench`, `help`, `uninstall`, `version` — no `registry` entry): only the worktree-local, uncommitted `installer/registry.go`/`installer/registry_cli.go` implement it, and that installer has not been rebuilt/reinstalled as part of this compilation. All three same-day sibling compilations (`smaqit.new-template`, `smaqit.new-rules`, `smaqit.new-principle`) recorded the identical finding and likewise skipped registration. Followed the same precedent here, and matched the explicit scope of this task's own instructions (confirm definition, 3-way merge, conciseness filter, validate placeholders/nesting — no mention of registry registration).

## Validation Checklist

- [x] No unresolved `[PLACEHOLDER]`-style uppercase-bracket tokens (`grep -noE '\[[A-Z_]+\]'` against the compiled file → no matches)
- [x] Frontmatter `name` field exactly `smaqit.sync`
- [x] Description is capability-oriented, declarative, present-tense; no "Use when...", no first person, no can/may/might/will; reworded away from the definition draft's "lets the user" gating phrasing (see Merge Process Summary)
- [x] All required sections present, in template order: `## Steps`, `## Output`, `## Scope`, `## Examples`, `## Gotchas`, `## Completion`, `## Failure Handling`
- [x] `compatibility` frontmatter field present (sourced from definition); `allowed-tools` omitted (not specified)
- [x] No nested reference chains — this skill's body references no file bundled within its own skill directory at all (no `SKILL.md` → file, let alone `SKILL.md` → file → file); it references `smaqit.L2` and `smaqit-adk registry put`/`get` as external agent/CLI invocations, not as files read for instruction
- [x] Body under 400 lines (77 lines)
- [x] Third-party validator (`validate-skill.go`, installed globally) passes with zero violations; re-ran against all 8 skills in `skills/` (the 7 pre-existing plus this one) as a sanity check — all pass

## Decisions Made

1. Reworded the `[SKILL_DESCRIPTION]`'s "lets the user choose..." clause to avoid the conversational/user-centric gating pattern the rules file bans, and respaced the draft's tight `L0→L1→L2` notation to the README's established `L0 → L1 → L2` spacing, since no shipped `SKILL.md` uses an arrow character at all.
2. Chose the flat numbered-list Steps style (matching `smaqit.new-rules`/`smaqit.new-principle`) over the `### N. Title` heading style used by `smaqit.new-template`, following the majority convention among the three same-day siblings rather than the specific one suggested for skimming.
3. Kept the "Subagent invocation fails" base row alongside the definition's more specific "`smaqit.L2` recompile fails" row rather than deduping, since the two carry meaningfully different action guidance (generic "do not silently retry" vs. this skill's specific "leave drifted, do not fall back to re-registering without asking") — a stricter application of the `smaqit.new-template` precedent's own textual-identity bar for deduping.
4. Did not perform the worktree `agents/smaqit.L2.md`'s Step 9 "Register" instruction, since the registry feature it depends on is uncommitted, worktree-local, and not present in the installed `smaqit-adk` binary — matching the identical precedent set by all three same-day sibling compilations, and consistent with this task's own explicitly scoped instructions.
