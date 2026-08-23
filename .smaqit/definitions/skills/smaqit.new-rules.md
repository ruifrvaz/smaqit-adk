---
name: smaqit.new-rules
tier: advanced
---

# Skill Definition: smaqit.new-rules

## Identity

- **Name:** `smaqit.new-rules`
- **Description (draft, L2 may refine):** Adds or refines a Level 1 compiled rules file (`templates/agents/compiled/*.rules.md` or `templates/skills/compiled/*.rules.md`) — delegates gathering, validation, and authoring to `smaqit.L1`, then cascades the change to every registered dependent.
- **Version:** 0.1.0

## Steps (with fragility)

1. **Invoke `smaqit.L1`** (low fragility). On Claude Code or Codex CLI, invoke it as a native subagent/custom-agent call. On GitHub Copilot (no dedicated compiled `smaqit.L1` file exists), read `~/.claude/agents/smaqit-L1.md`'s body directly and follow its instructions inline for this turn. Pass the instruction: "The user wants to add or refine a Level 1 compiled rules file (not a template structure file)." Wait for `smaqit.L1` to confirm which rules file it wrote and report back that path.
2. **Query the registry for dependents** (high fragility — the cascade only reaches what's actually registered, and a wrong scope silently skips real dependents). Run `smaqit-adk registry get --scope base --depends-on <changed-rules-path>` against the ADK-global registry. If a project-local `.smaqit/registry.json` also exists, also run `smaqit-adk registry get --scope extension --depends-on <changed-rules-path>` against it. Treat a missing registry file as zero dependents, not an error.
3. **Cascade to each dependent** (medium fragility). For every dependent entry returned, invoke `smaqit.L2` as a single-hop `Task` call to recompile that agent or skill against the updated rules — mirror `smaqit.development`'s orchestration of its Business/Functional/Stack spec agents: each `Task` call originates from this skill's own turn, not from inside `smaqit.L1`. Do not grant `smaqit.L0`/`smaqit.L1` a `Task` tool to chain into `smaqit.L2` directly.
4. **Report** (low fragility). State which rules file changed and which dependents (if any) were recompiled, or that no dependents were registered.

## Output

- Rules file added or updated in `templates/agents/compiled/*.rules.md` or `templates/skills/compiled/*.rules.md`
- Zero or more dependent agents/skills recompiled via `smaqit.L2`, in the same session
- A cascade summary: rules file changed, dependents found, dependents recompiled

## Scope

Compiled rules changes only. Does not author template structure (`smaqit.new-template`) and does not author L0 principles (`smaqit.new-principle`). Does not modify the registry format itself — only reads it to resolve dependents.

## Completion

- [ ] `smaqit.L1` subagent invoked and confirms the rules file added or updated
- [ ] Registry queried for dependents of the changed rules file (both ADK-global and project-local, when present)
- [ ] Each returned dependent recompiled via a single-hop `smaqit.L2` invocation
- [ ] Cascade result reported to the user

## Failure Scenarios

| Situation | Action |
|-----------|--------|
| `smaqit.L1` agent not installed | Instruct the user to run the smaqit-adk installer (`install.sh`), which installs `smaqit.L1` globally |
| Subagent invocation fails | Report the failure with context; do not silently retry |
| Registry file missing or unreadable | Treat as zero dependents; report the rules file changed with no cascade; do not block |
| A dependent's `smaqit.L2` recompile fails | Report which dependent failed and why; leave other successful recompiles in place; do not roll back |
| User wants to author templates or principles instead | Stop; redirect to `smaqit.new-template` or `smaqit.new-principle` |
| User abandons the request midway | The `smaqit.L1` subagent handles this case; this skill does not intervene |

## Gotchas

- A freshly installed machine has no registry file yet — this is expected, not an error; the cascade step must degrade to "zero dependents" cleanly.
- The cascade always runs `smaqit.L2` as a sibling `Task` call from this skill's own turn, never as a nested call from inside `smaqit.L1` — `smaqit.L1` has no `Task` tool grant and this skill does not add one.

## Examples

**Input:** User says "add a MUST directive to the base rules requiring output validation before completion."
**Output:** `smaqit.L1` updates `templates/agents/compiled/base.rules.md` and reports the path. This skill queries the registry, finds every registered agent compiled from base rules as a dependent, and invokes `smaqit.L2` once per dependent to recompile. Reports: "Rules updated; 3 dependents recompiled."

## Compatibility

Requires the `smaqit-adk` binary on `PATH` for the registry subcommands used in Step 2.
