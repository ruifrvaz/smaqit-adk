# Create smaqit.new-principle Skill

**Status:** In Progress
**Created:** 2026-02-27

## Description

Create a new ADK skill `smaqit.new-principle` that guides the addition of a new principle to the framework. The skill gathers the principle's details interactively, then invokes Agent-L0 as a subagent to write the principle into the appropriate framework file following the established structure and conventions.

This skill mirrors the pattern of `smaqit.new-agent` (gather → write definition → invoke specialist agent) but targets L0 instead of L2, and the output is a framework principle rather than an agent file.

**Why this skill is needed:** Framework principles are currently added ad hoc — this session added `Description-Driven Activation` to SKILLS.md manually, with the formatting pattern inferred from SMAQIT.md. A skill makes the process repeatable, enforces the `### Name` / bold one-liner / explanation paragraph pattern, ensures the principle is placed in the correct framework file, and catches duplicates or conflicts with existing principles before writing.

**Skill description (to apply Description-Driven Activation principle):**
> "Adds a new principle to the ADK framework. Use this skill when defining a new framework-level rule or guideline — the skill gathers the principle name, one-line statement, explanation, and target framework file, validates it against existing principles, then invokes Agent-L0 to write it."

## Acceptance Criteria

- [ ] `skills/smaqit.new-principle/SKILL.md` created at root (shipped ADK artifact)
- [ ] Description follows Description-Driven Activation principle: explanatory, answers "this skill does..." and "use this skill when...", no bare keyword lists
- [ ] Gathering steps collect: principle name, one-line bold statement, explanation paragraph(s), target framework file, and rationale for why it is a framework-level principle
- [ ] Validation step checks: principle name not already present in target file, statement is a single bold imperative sentence, explanation supports the statement without contradicting existing principles
- [ ] Compilation step writes a definition file to `.smaqit/definitions/principles/[name].md` containing gathered specs, then invokes `smaqit.L0` as a subagent passing the definition file path
- [ ] Definition file format documented in skill body
- [ ] Notes section states: this skill targets the ADK framework only — product-domain principles belong in the product extension, not the ADK
- [ ] Installer synced via `make build` (no manual copy)

## Prerequisites

Before creating the skill:

1. **Revisit `agents/smaqit.L0.agent.md` description** — Apply the Description-Driven Activation principle established this session. L0's description is the activation trigger for expert users and subagent invocation; it must explain what L0 does and when to use it, not just label it. Assess and rewrite if needed before building the skill that invokes it.
2. **Assess whether L0 needs a definition file input pattern** — L0 currently has no instruction for reading a `.smaqit/definitions/principles/[name].md` file. This skill requires that pattern. Update L0 as needed (same pattern applied to L2 in Task 003).

## Notes

- Follow the same gather → definition file → subagent pattern as `smaqit.new-agent`
- The definition file path convention should extend `.smaqit/definitions/` — use `.smaqit/definitions/principles/[name].md` for namespace consistency
- Target framework file is user-specified but constrained to `framework/AGENTS.md`, `framework/SKILLS.md`, `framework/TEMPLATES.md`, etc. — not product files
