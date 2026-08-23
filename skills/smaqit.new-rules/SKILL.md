---
name: smaqit.new-rules
description: Adds or refines a Level 1 compiled rules file in templates/agents/compiled/ or templates/skills/compiled/ — delegates gathering, validation, and authoring to smaqit.L1, then cascades the change by recompiling every registered dependent agent or skill via smaqit.L2.
metadata:
  version: "0.1.0"
compatibility: Requires the smaqit-adk binary on PATH for the registry subcommands used in Step 2.
---

# New Rules

## Steps

1. Invoke `smaqit.L1`:
   - **On Claude Code or Codex CLI:** invoke it as a native subagent/custom-agent call.
   - **On GitHub Copilot** (no dedicated compiled `smaqit.L1` file exists): read `~/.claude/agents/smaqit-L1.md`'s body directly and follow its instructions inline for this turn, as if it had been invoked as a subagent.

   Pass the instruction: "The user wants to add or refine a Level 1 compiled rules file (not a template structure file)." Wait for `smaqit.L1` to confirm which rules file it wrote, and record that path for the next step.

2. Query the registry for dependents of the changed rules file:

   ```
   smaqit-adk registry get --scope base --depends-on <changed-rules-path>
   ```

   If a project-local `.smaqit/registry.json` also exists, also run:

   ```
   smaqit-adk registry get --scope extension --depends-on <changed-rules-path>
   ```

   Treat a missing registry file as zero dependents, not an error.

3. For every dependent entry returned, invoke `smaqit.L2` as a single-hop `Task` call to recompile that agent or skill against the updated rules file — mirror `smaqit.development`'s orchestration of its Business/Functional/Stack spec agents: each `Task` call originates from this skill's own turn, not from inside `smaqit.L1`, which has no `Task` tool grant to chain into `smaqit.L2` directly.

4. Report which rules file changed and which dependents (if any) were recompiled, or that no dependents were registered.

## Output

- The changed rules file (`templates/agents/compiled/*.rules.md` or `templates/skills/compiled/*.rules.md`), added or updated by the `smaqit.L1` subagent
- Zero or more dependent agents or skills recompiled via `smaqit.L2`, in the same session
- A cascade summary reported to the user: which rules file changed, how many dependents were found, and how many were recompiled

## Scope

This skill triggers rules authoring and cascade only. It does not gather or validate rules content directly — all gathering, validation, and file authoring happen inside the `smaqit.L1` subagent. It does not modify the registry format itself — only reads it to resolve dependents.

Out of scope:
- Template structure authoring — use the `smaqit.new-template` skill
- Framework principle authoring — use the `smaqit.new-principle` skill

## Examples

**Input:** User says "add a MUST directive to the base rules requiring output validation before completion."
**Output:** `smaqit.L1` updates `templates/agents/compiled/base.rules.md` and reports the path. This skill queries the registry, finds every registered agent compiled from base rules as a dependent, and invokes `smaqit.L2` once per dependent to recompile. Reports: "Rules updated; 3 dependents recompiled."

## Gotchas

- A freshly installed machine has no registry file yet — this is expected, not an error; the cascade step degrades to "zero dependents" cleanly.
- The cascade always runs `smaqit.L2` as a sibling `Task` call from this skill's own turn, never as a nested call from inside `smaqit.L1` — `smaqit.L1` has no `Task` tool grant and this skill does not add one.

## Completion

- [ ] `smaqit.L1` subagent invoked and confirms the rules file added or updated
- [ ] Registry queried for dependents of the changed rules file (both ADK-global and project-local, when present)
- [ ] Each returned dependent recompiled via a single-hop `smaqit.L2` invocation
- [ ] Cascade result reported to the user

## Failure Handling

| Situation | Action |
|-----------|--------|
| Required input not provided | Request the missing information before proceeding |
| Gathered input is ambiguous | Flag the ambiguity and ask for clarification |
| Subagent invocation fails | Report the failure with context; do not silently retry |
| Output artifact already exists | Confirm with user before overwriting |
| `smaqit.L1` agent not installed | Instruct the user to run the smaqit-adk installer (`install.sh`), which installs `smaqit.L1` globally |
| Registry file missing or unreadable | Treat as zero dependents; report the rules file changed with no cascade; do not block |
| A dependent's `smaqit.L2` recompile fails | Report which dependent failed and why; leave other successful recompiles in place; do not roll back |
| User wants to author templates or principles instead | Stop; redirect to `smaqit.new-template` or `smaqit.new-principle` |
| User abandons the request midway | The `smaqit.L1` subagent handles this case; this skill does not intervene |
