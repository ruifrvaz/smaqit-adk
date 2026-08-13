---
name: smaqit.create-skill
description: Generates a compiled SKILL.md file from a name and project context — infers a complete specification covering steps, output, scope, failure handling, and examples; writes a definition file; and invokes smaqit.L2 to produce the final skill artifact. Applies to new skill creation, workflow packaging, domain knowledge encapsulation, and repeatable procedure authoring.
metadata:
  version: "3.0.0"
allowed-tools: Bash
---

# Create Skill

## Steps

### 1. Gather

Ask the user for the skill **name** in a single message (lowercase, hyphens allowed, e.g., `my-review`). The description will be inferred from the name and scanned context.

### 2. Scan

Before writing anything, read:
- All existing files in `.agents/skills/` and `.claude/skills/` — for patterns and conventions already used in this project
- Project README — for domain and conventions
- Any project manifests that describe workflows or user-facing operations

Also extract any relevant detail the user has already provided in the conversation.

### 3. Infer and write definition file

Using the name and scanned context, infer a complete skill specification. Do not ask further questions.

Write the inferred specification to `.smaqit/definitions/skills/[name].md`. Create the directory if it does not exist.

For any field where the correct value is genuinely ambiguous, suffix the value with `[?]` and add a brief inline note.

The definition file must cover:
- Name and description
- Steps (what the skill does, in sequence)
- Output (what the skill produces)
- Scope (what is out of scope)
- Completion criteria (testable, checkbox-style)
- Failure handling (likely failure modes and responses)
- Gotchas — environment-specific facts the agent must know before executing; non-obvious corrections to mistakes it would make without being told. Distinct from failure handling (which is reactive). Include any project conventions, unexpected API behaviors, naming quirks, or platform constraints that apply to this skill.
- Examples — at least one concrete example of what triggers the skill and what it produces. Input: a representative user request. Output: the artifact or response produced.
- Allowed tools (optional) — if the skill requires specific tools to run (e.g., git, bash scripts), list them as `allowed-tools` values using the format `Bash(git:*)`, `Read`, etc.
- Compatibility (optional) — if the skill has environment requirements (specific agent product, system packages, network access), note them here.

### 4. Compile

Invoke `smaqit.L2`:
- **On Claude Code or Codex CLI:** invoke it as a native subagent/custom-agent call.
- **On GitHub Copilot** (no dedicated compiled `smaqit.L2` file exists): read `~/.claude/agents/smaqit-L2.md`'s body directly and follow its instructions inline for this turn, as if it had been invoked as a subagent.

Pass the instruction:
> "Compile the skill definition at `.smaqit/definitions/skills/[name].md`. Write the compiled skill to `.agents/skills/[name]/SKILL.md` and `.claude/skills/[name]/SKILL.md` — identical content in both, since `SKILL.md` is already a cross-platform format; this covers Codex CLI (`.agents/skills/`) and Claude Code (`.claude/skills/`) project-local discovery. After compilation, list any fields annotated with `[?]` and suggest a resolution for each. If the compiled skill body would exceed 400 lines, move detailed reference content to a `references/` subdirectory and link from SKILL.md with explicit load conditions ("Read references/[file].md if [condition]"). The main SKILL.md body must remain under 400 lines after extraction."

### 5. Validate

Read `scripts/validate-skill.go` before running this step.

Run the validation script (installed globally alongside this skill) against the compiled skill (written project-locally):

```bash
go run ~/.agents/skills/smaqit.create-skill/scripts/validate-skill.go .claude/skills/[name]/SKILL.md
```

If violations are reported:
1. Surface the violations to the user.
2. Automatically analyse each violation and apply the minimal fix to the definition file (`.smaqit/definitions/skills/[name].md`) that resolves it — do not ask the user to fix them manually.
3. Re-run Step 4 (Compile) to regenerate the compiled skill from the updated definition.
4. Re-run the validation script.
5. Repeat steps 2–4 up to **3 times** in total. If violations still remain after the third attempt, stop, surface the remaining violations to the user, and ask them to update the definition file before re-invoking the skill.

Do not proceed to Step 6 while violations remain.

### 6. Report

After L2 completes, report to the user:
- Paths of both compiled skill files
- Any `[?]`-annotated items and L2's suggested resolutions
- How to adjust: edit `.smaqit/definitions/skills/[name].md` and re-invoke `/smaqit.create-skill`, or switch to `smaqit.L2` directly

## Output

- `.smaqit/definitions/skills/[name].md` — inferred specification (scaffolding)
- `.agents/skills/[name]/SKILL.md` — compiled skill file, Codex CLI project-local discovery (source of truth)
- `.claude/skills/[name]/SKILL.md` — compiled skill file, Claude Code project-local discovery (identical content, source of truth)

## Scope

Does not create agents, framework files, or templates.

## Completion

- [ ] Name obtained from user
- [ ] Repository scanned for context
- [ ] Definition file written to `.smaqit/definitions/skills/[name].md`
- [ ] `smaqit.L2` invoked and compilation completed
- [ ] Compiled skill exists at `.agents/skills/[name]/SKILL.md` and `.claude/skills/[name]/SKILL.md`
- [ ] Validation script passes with no violations

## Failure Handling

| Situation | Action |
|-----------|--------|
| Name not provided | Request before proceeding |
| `~/.agents/smaqit-adk/templates/` not present | Inform the user that ADK templates are required — run the smaqit-adk installer (`install.sh`) first |
| Output artifact already exists | Report the conflict; do not overwrite without user confirmation |
| L2 invocation fails | Report the error and include the path to the definition file so the user can inspect or correct it |
| Validation script reports violations | Surface violations to the user; auto-fix the definition file and re-run Steps 4–5 up to 3 times; if violations persist after 3 attempts, ask the user to update the definition file before re-invoking |
