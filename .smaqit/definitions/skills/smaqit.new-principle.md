---
name: smaqit.new-principle
tier: advanced
---

# Skill Definition: smaqit.new-principle

## Identity

- **Name:** `smaqit.new-principle`
- **Description (draft, L2 may refine):** Adds or refines a principle in an ADK framework file, then cascades the change through `smaqit.L1` (template/rules review) and `smaqit.L2` (dependent recompilation) so every registered dependent stays current automatically.
- **Version:** 0.3.0 (revision of the existing 0.2.0 skill — adds the live-cascade steps; Step 1 and the overall purpose are unchanged)

## Steps (with fragility)

1. **Invoke `smaqit.L0`** (low fragility — unchanged from the existing skill). On Claude Code or Codex CLI, invoke it as a native subagent/custom-agent call. On GitHub Copilot (no dedicated compiled `smaqit.L0` file exists), read `~/.claude/agents/smaqit-L0.md`'s body directly and follow its instructions inline for this turn. Pass the instruction: "The user wants to add or refine a framework principle." Capture `smaqit.L0`'s reported scope (base/extension) and the exact framework file path it changed.
2. **Invoke `smaqit.L1` to review template/rules impact** (medium fragility — a judgment call `smaqit.L1` makes, not this skill). A single-hop `Task` call from this skill's own turn, not chained from inside `smaqit.L0` — `smaqit.L0` has no `Task` tool grant and this skill does not add one. Pass: "Review whether any `templates/agents/*.template.md`, `templates/agents/compiled/*.rules.md`, `templates/skills/*.template.md`, or `templates/skills/compiled/*.rules.md` file needs updating to reflect this principle change: [summarize the principle diff from Step 1]." `smaqit.L1` may determine no update is needed — that is a valid, common outcome, not a failure. Capture the exact path(s) of any file `smaqit.L1` actually wrote.
3. **Query the registry for dependents** (high fragility — a wrong scope or a skipped project-local check silently drops real dependents from the cascade). For each template/rules path `smaqit.L1` touched in Step 2, run `smaqit-adk registry get --scope base --depends-on <path>` against the ADK-global registry, and `smaqit-adk registry get --scope extension --depends-on <path>` against a project-local `.smaqit/registry.json` when one exists. Treat a missing registry file as zero dependents, not an error.
4. **Cascade to each dependent** (medium fragility). For every dependent entry returned, invoke `smaqit.L2` as a single-hop `Task` call to recompile it — mirror `smaqit.development`'s orchestration of its Business/Functional/Stack spec agents. Skip this step entirely if Step 2 found no template/rules update was needed.
5. **Report** (low fragility). State the principle change and its scope, whether `smaqit.L1` updated any template/rules file, and which dependents (if any) were recompiled.

## Output

- A principle added or updated in the appropriate framework file (base or extension scope), produced by `smaqit.L0`
- Zero or more template/rules files updated by `smaqit.L1`, when the principle change warrants it
- Zero or more dependent agents/skills recompiled via `smaqit.L2`, in the same session
- A cascade summary covering all of the above

## Scope

This skill triggers principle authoring and owns the resulting live cascade. `smaqit.L0` performs all principle gathering, validation, and framework editing; `smaqit.L1` performs all template/rules judgment; `smaqit.L2` performs all recompilation. This skill's own logic is limited to sequencing those three calls and querying the registry between them — it never edits framework, template, rules, or compiled agent/skill content directly.

Out of scope:
- Template-only or rules-only authoring with no principle change — use `smaqit.new-template` or `smaqit.new-rules` directly
- Agent or skill creation — use `smaqit.create-agent` or `smaqit.create-skill`

## Completion

- [ ] `smaqit.L0` subagent invoked; confirms principle added or updated, with scope and file path reported
- [ ] `smaqit.L1` subagent invoked to review template/rules impact; confirms which file(s), if any, it updated
- [ ] Registry queried for dependents of every template/rules file `smaqit.L1` touched (both ADK-global and project-local, when present)
- [ ] Each returned dependent recompiled via a single-hop `smaqit.L2` invocation
- [ ] Full cascade result reported to the user

## Failure Scenarios

| Situation | Response |
|-----------|----------|
| `smaqit.L0` agent not installed | Instruct the user to run the smaqit-adk installer (`install.sh`), which installs `smaqit.L0` globally to `~/.claude/agents/` and `~/.codex/agents/` |
| Subagent invocation fails (any of L0/L1/L2) | Report the failure with context; do not silently retry |
| `smaqit.L0` rejects an extension principle as duplicating base content | Report the rejection and the base principle it duplicates; stop — do not attempt the cascade |
| `smaqit.L1` finds no template/rules update needed | Not a failure — report the principle change alone and skip Steps 3–4 |
| Registry file missing or unreadable | Treat as zero dependents for that scope; continue, do not block |
| A dependent's `smaqit.L2` recompile fails | Report which dependent failed and why; leave other successful recompiles in place; do not roll back |
| User wants to author templates or rules instead | Stop; redirect to the appropriate skill |
| User abandons creation midway | The subagent handles this case; this skill does not intervene |

## Gotchas

- A principle change does not always require a template/rules update — `smaqit.L1`'s "no update needed" outcome in Step 2 is common and must short-circuit Steps 3–4 cleanly, not be treated as an empty/failed cascade.
- The registry has no direct edge from an L0 framework file to an L2-compiled resource — only from an L1 template/rules file to the resources compiled from it. This is why Step 2 (asking `smaqit.L1` what it touched) must run *before* Step 3's registry lookup, not be skipped in favor of looking up the framework path directly.

## Examples

**Input:** User says "add a principle that agents must fail fast on ambiguous input rather than guessing."
**Output:** `smaqit.L0` adds the principle to `framework/SMAQIT.md` (base scope). `smaqit.L1` determines `templates/agents/compiled/base.rules.md` already covers this via its fail-fast directive and makes no change — cascade skipped. Skill reports: "Principle added (base scope); no template/rules update needed."

## Compatibility

Requires the `smaqit-adk` binary on `PATH` for the registry subcommands used in Step 3.
