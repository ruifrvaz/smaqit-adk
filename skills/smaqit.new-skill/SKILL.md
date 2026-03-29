---
name: smaqit.new-skill
description: Guides the creation of a new skill for this project. Use this skill when the user wants to define and compile a new skill — the skill gathers the skill's purpose, steps, output, scope, completion criteria, and failure handling through an interactive interview, then writes a definition file and invokes Agent-L1 to compile the skill file.
metadata:
  version: "0.2.0"
---

# New Skill Creation

## Steps

Gather the following specifications from the user in order. Request each section systematically — do not infer values without asking.

### 1. Skill Identity

**Name**
- Ask: "What is the skill name? Use the format `smaqit.[name]` — lowercase, hyphens allowed (e.g., `smaqit.release-analysis`, `smaqit.new-principle`)"
- Used in: directory name (`skills/[name]/`), frontmatter `name` field

**Description**
- Ask: "What is the skill description? Write in third person. Include what the skill does and when to use it. (e.g., 'Guides creation of X. Use when the user wants to Y or when Z.')"
- Constraint: third person only; must include both what and when; maximum 1024 characters
- Used in: frontmatter `description` field — this is the sole discovery signal

**Version**
- Default to `"0.2.0"` unless the user specifies otherwise.

### 2. Skill Purpose

**What it does**
- Ask: "What workflow does this skill guide? What does it produce at the end?"

**What triggers it**
- Ask: "What user request or situation should activate this skill?"

### 3. Steps

For each step the skill will contain:
- Ask: "Describe each step the agent should follow when running this skill. For each step, also tell me: is this step high-fragility (exact sequence required), medium (preferred pattern, some variation ok), or low-fragility (many valid approaches)?"
- Collect all steps before moving on.

Use fragility levels to determine how L1 will write each step:
- **High** → exact commands or literal sequences
- **Medium** → templates or pseudocode
- **Low** → prose guidance

### 4. Output

- Ask: "What artifact does this skill produce? Where does it live?"
- Ask: "Does this skill invoke a subagent to compile the output? If so, which one? (e.g., `smaqit.L1`, `smaqit.L2`, `smaqit.L0`)"

### 5. Scope

- Ask: "What does this skill explicitly NOT handle? Are there redirections to other skills or agents?"

### 6. Failure Handling

- Ask: "What failure scenarios should this skill handle? Provide as situation / action pairs."
- Example pairs:
  - User provides incomplete step description → "Request the missing detail before proceeding"
  - Output artifact already exists → "Confirm with user before overwriting"

## Validation

After gathering all specifications:

1. **Validate description** — Must be written in third person. Must include both what the skill does and when to invoke it. Reject first-person or second-person phrasing.
2. **Validate completeness** — All 6 sections must have content. Request missing information before proceeding.
3. **Validate steps** — Each step must have a fragility level. Reject steps without one.
4. **Confirm** — Present a summary of gathered specifications and ask the user to confirm before compiling.

## Compilation

Once specifications are confirmed:

1. Write a definition file to `.smaqit/definitions/skills/[name].md` containing all gathered specifications in the format below
2. Use the `agent` tool to invoke `smaqit.L1` as a subagent, passing the definition file path as context
3. Agent-L1 will read the definition file and compile the skill:
   - `templates/skills/base-skill.template.md` — structure
   - `templates/skills/compiled/skill.rules.md` — compilation directives
   - `.smaqit/definitions/skills/[name].md` — gathered specifications
4. Output: `skills/[name]/SKILL.md` created by L1

### Definition File Format

```markdown
# Skill Definition: [name]

**Created:** YYYY-MM-DD
**Skill:** smaqit.new-skill

## Identity

- **Name:** [name]
- **Description:** [description — third person, what + when]
- **Version:** [version]

## Purpose

- **What it does:** [workflow the skill guides and artifact it produces]
- **What triggers it:** [user request or situation that activates the skill]

## Steps

| Step | Description | Fragility |
|------|-------------|-----------|
| 1 | [step description] | [High / Medium / Low] |
| 2 | [step description] | [High / Medium / Low] |

## Output

- **Artifact:** [path to output file]
- **Subagent:** [smaqit.L1 / smaqit.L2 / smaqit.L0 / none]

## Scope

[what this skill does NOT handle; redirections]

## Failure Handling

| Situation | Action |
|-----------|--------|
| [situation] | [action] |
```

## Notes

- The definition file at `.smaqit/definitions/skills/[name].md` is the auditable record of what was requested
- L1 applies the conciseness filter and degrees-of-freedom rules when writing the skill body — the gathered steps are inputs, not verbatim output
- After the skill is compiled, the user must add it to `installer/Makefile` prepare target to include it in ADK distribution
