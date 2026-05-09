# Skill Format Spec Alignment

**Status:** Not Started  
**Created:** 2026-05-09

## Description

Align the ADK's skill compilation chain with the agentskills.io specification and best practices. The current chain produces skills that are structurally valid but missing several high-value sections that the spec recommends and that meaningfully improve skill quality.

Three artifacts require changes:
1. `skills/smaqit.create-skill/SKILL.md` — the gathering + compilation skill itself
2. `templates/skills/base-skill.template.md` — the template L2 compiles against
3. `templates/skills/compiled/skill.rules.md` — the L1-compiled directives L2 follows

After editing root artifacts, run `cd installer && make build` to sync changes into the installer. All tests must pass.

---

## Background: Gap Analysis

This analysis was performed against https://agentskills.io/specification and https://agentskills.io/skill-creation/best-practices.md.

### Gap 1 — Description field: phrasing and trigger scope (HIGH)

**Spec says:**
- Use imperative phrasing: "Use this skill when..." not "This skill does..."
- Focus on user intent, not implementation mechanics
- "Err on the side of being pushy" — list contexts including indirect ones where the user doesn't name the domain directly
- Be specific about what the output is

**Current `smaqit.create-skill` description:**
```
Creates a new skill for this project. Use when the user asks to create, define, or build a new skill.
```

**Problems:**
- Declarative first sentence
- Only triggers on explicit "create a skill" wording; misses indirect intent ("I need to package this workflow as a reusable command", "make this a skill")
- Doesn't say the output is a compiled `SKILL.md` file
- Doesn't mention it compiles via L2

### Gap 2 — Definition file spec: missing `gotchas` section (HIGH)

**Spec says:**
> "The highest-value content in many skills is a list of gotchas — environment-specific facts that defy reasonable assumptions. These aren't general advice but concrete corrections to mistakes the agent will make without being told otherwise."

> "Keep gotchas in SKILL.md where the agent reads them before encountering the situation."

**Current definition file spec in `smaqit.create-skill`:**
- Name and description
- Steps
- Output
- Scope
- Completion criteria
- Failure handling

**Problem:** No `gotchas` section. The failure handling table covers what to do when things go wrong — but gotchas are proactive: non-obvious environment facts the agent needs *before* executing, not reactive error handling.

### Gap 3 — Definition file spec: missing `examples` section (MEDIUM)

**Spec says:** "Recommended sections: Step-by-step instructions, Examples of inputs and outputs, Common edge cases."

**Current:** No examples slot in the definition file spec. L2 is never instructed to include concrete input/output examples in compiled skills.

### Gap 4 — `allowed-tools` and `compatibility` frontmatter never surfaced (MEDIUM)

**Spec:** Both are valid optional frontmatter fields. `allowed-tools` is a space-separated string of pre-approved tools. `compatibility` declares environment requirements.

**Current:** The definition file spec, base template, and rules file make no mention of either field. Skills that require specific tools (e.g., `Bash(git:*)`) have no mechanism to declare them.

### Gap 5 — Progressive disclosure: no guidance on `references/` (MEDIUM)

**Spec says:**
> "Keep your main SKILL.md under 500 lines. Move detailed reference material to separate files. The key is telling the agent *when* to load each file."
> "Structure large skills with progressive disclosure — tell the agent when to load each file."

**Current:** The `skill.rules.md` mentions that reference content lives in subdirectories and must be referenced from `SKILL.md`, but L2 is given no guidance on when to apply this — no threshold triggers, no instruction to split content proactively when the skill definition is large.

### Gap 6 — Description field directives in `skill.rules.md` (MEDIUM)

**Current rules say:** description "Must be written in third person. Must explain what the skill does and when to use it."

**Spec says:** Use imperative phrasing ("Use this skill when..."), focus on user intent not implementation. These two are in tension — the spec favors second-person imperative directed at the agent ("Use when..."), while the current rules mandate third person.

The distinction matters: "Use this skill when the user wants to package a workflow as a reusable command" (imperative/agent-facing) outperforms "Packages a workflow as a reusable command" (declarative/third-person) for triggering reliability. The rules should be updated to match spec best practices.

---

## Changes Required

### 1. `skills/smaqit.create-skill/SKILL.md`

**a. Description field — rewrite**

Current:
```
Creates a new skill for this project. Use when the user asks to create, define, or build a new skill.
```

Rewrite to imperative, broader triggers, output-explicit:
```
Use when the user wants to create, define, build, or package a new skill — including when they ask to turn a workflow into a reusable command, wrap domain knowledge into a slash-command, or describe a repeatable procedure they want Copilot to follow. Gathers name and purpose, infers a complete specification, writes a definition file, and invokes smaqit.L2 to compile a SKILL.md file.
```

**b. Step 3 — expand definition file spec**

Add these sections to the list of required definition file content:

- **Gotchas** — environment-specific facts the agent must know before executing; non-obvious corrections to mistakes it would make without being told. Distinct from failure handling (which is reactive). Include any project conventions, unexpected API behaviors, naming quirks, or platform constraints that apply to this skill.
- **Examples** — at least one concrete example of what triggers the skill and what it produces. Input: a representative user request. Output: the artifact or response produced.
- **Allowed tools** (optional) — if the skill requires specific tools to run (e.g., git, bash scripts), list them as `allowed-tools` values using the format `Bash(git:*)`, `Read`, etc.
- **Compatibility** (optional) — if the skill has environment requirements (specific agent product, system packages, network access), note them here.

**c. Step 4 — add progressive disclosure guidance for L2**

Append to the compile instruction sent to L2:
```
If the compiled skill body would exceed 400 lines, move detailed reference content to a `references/` subdirectory and link from SKILL.md with explicit load conditions ("Read references/[file].md if [condition]"). The main SKILL.md body must remain under 400 lines after extraction.
```

### 2. `templates/skills/base-skill.template.md`

Add two new sections and two optional frontmatter fields:

**Frontmatter additions (optional):**
```yaml
compatibility: [COMPATIBILITY]
allowed-tools: [ALLOWED_TOOLS]
```
These should be omitted (not rendered as empty lines) when the definition file specifies they are not needed.

**New body sections to add after `## Scope` and before `## Completion`:**

```markdown
## Examples

[EXAMPLES_CONTENT]

## Gotchas

[GOTCHAS_CONTENT]
```

Full updated template structure:
```
---
name: [SKILL_NAME]
description: [SKILL_DESCRIPTION]
metadata:
  version: "[SKILL_VERSION]"
---

# [SKILL_TITLE]

## Purpose

[PURPOSE_CONTENT]

## Steps

[STEPS_CONTENT]

## Output

[OUTPUT_CONTENT]

## Scope

[SCOPE_CONTENT]

## Examples

[EXAMPLES_CONTENT]

## Gotchas

[GOTCHAS_CONTENT]

## Completion

[COMPLETION_CONTENT]

## Failure Handling

[FAILURE_HANDLING_CONTENT]
```

Note: `compatibility` and `allowed-tools` are optional frontmatter fields. L2 MUST include them only when the definition file specifies a value. When omitted, they must not appear in the compiled output (no empty field lines).

### 3. `templates/skills/compiled/skill.rules.md`

**a. Update `description` field directive**

Change from:
> Must be written in third person. Must explain what the skill does and when to use it.

Change to:
> Must use imperative phrasing directed at the agent: "Use this skill when..." or "Use when...". Must describe user intent (what the user is trying to achieve) not implementation mechanics. Must state what the skill produces. Should cover indirect trigger contexts — situations where the skill is relevant even if the user doesn't use the exact domain keyword. Maximum 1024 characters.

**b. Add new placeholder entries to the Placeholder Catalog:**

| Placeholder | Description |
|-------------|-------------|
| `[EXAMPLES_CONTENT]` | At least one concrete example: a representative triggering request and the output produced. |
| `[GOTCHAS_CONTENT]` | Environment-specific facts the agent must know before executing. Non-obvious corrections to mistakes it would make without instruction. Proactive, not reactive. If none apply, write "None identified." |
| `[COMPATIBILITY]` | Optional. Environment requirements for this skill (agent product, system packages, network). Omit from compiled output if not specified. |
| `[ALLOWED_TOOLS]` | Optional. Space-separated pre-approved tools (e.g., `Bash(git:*) Read`). Omit from compiled output if not specified. |

**c. Add progressive disclosure directive**

Add a new directive section after the existing Loading Stages table:

```
**Progressive disclosure directive:** When compiling a skill whose body would exceed 400 lines, Agent-L2 MUST extract detailed reference content into a `references/` subdirectory and link from `SKILL.md` with explicit load conditions. Example: "Read `references/api-errors.md` if the API returns a non-200 status." The main `SKILL.md` body must remain under 400 lines after extraction. Reference files longer than 100 lines MUST include a table of contents at the top.
```

**d. Update `[SKILL_DESCRIPTION]` placeholder description**

Current:
> Activation signal. Maximum 1024 characters. Must be written in third person. Must explain what the skill does and when to use it.

Update to:
> Activation signal. Maximum 1024 characters. Must use imperative phrasing ("Use when..." or "Use this skill when..."). Must describe the user's intent and what the skill produces. Should include indirect trigger contexts. Must not describe internal implementation mechanics.

---

## Acceptance Criteria

- [ ] `smaqit.create-skill/SKILL.md` description is imperative, mentions compiled `SKILL.md` output, and covers indirect trigger contexts
- [ ] `smaqit.create-skill/SKILL.md` Step 3 definition file spec includes `gotchas` and `examples` sections
- [ ] `smaqit.create-skill/SKILL.md` Step 3 definition file spec mentions optional `allowed-tools` and `compatibility` fields
- [ ] `smaqit.create-skill/SKILL.md` Step 4 compile instruction includes progressive disclosure guidance for long skills
- [ ] `base-skill.template.md` includes `## Examples` and `## Gotchas` sections with correct placeholder names
- [ ] `base-skill.template.md` includes optional `compatibility` and `allowed-tools` frontmatter fields with L2 omission note
- [ ] `skill.rules.md` placeholder catalog includes `[EXAMPLES_CONTENT]`, `[GOTCHAS_CONTENT]`, `[COMPATIBILITY]`, `[ALLOWED_TOOLS]`
- [ ] `skill.rules.md` description field directive updated to imperative phrasing + intent-focused language
- [ ] `skill.rules.md` includes progressive disclosure directive with 400-line threshold
- [ ] `cd installer && make build` exits 0
- [ ] `cd tests && go test ./...` exits 0 (all structural and unit tests pass)
- [ ] The ADK's own skills (`smaqit.create-skill`, `smaqit.create-agent`, `smaqit.new-principle`) are NOT retroactively modified to add the new sections — this task only changes the template and compilation chain going forward

## Notes

- Do NOT modify the `.github/skills/` or `.github/agents/` files — those are smaqit product extensions, not ADK artifacts
- Do NOT manually copy files into `installer/framework/`, `installer/skills/`, `installer/agents/`, or `installer/templates/` — `make prepare` (run automatically by `make build`) handles all syncing
- The `name` field dot convention (e.g., `smaqit.create-skill`) is a known intentional divergence from the agentskills.io spec; do not change it
- The structural tests (`tests/structural/skills_test.go`) check for required sections by heading name — if new sections are added to the template, verify the test does not fail on existing shipped skills that lack those sections (they are optional additions, not required by the test)
- If the structural test checks section headings strictly, update it to treat `## Examples` and `## Gotchas` as optional (not required) for existing skills
