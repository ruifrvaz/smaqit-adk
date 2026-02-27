# Migrate Prompts to Skills

**Status:** Completed  
**Created:** 2026-02-27  
**Completed:** 2026-02-27

## Description

GitHub Copilot prompts (`.github/prompts/`) are being deprecated in favour of Agent Skills (agentskills.io format). ADK must migrate its single prompt artefact and all framework/agent/installer references to the skills format.

The ADK's prompt (`smaqit.new-agent.prompt.md`) maps cleanly to a skill — it was already instruction-only, never a user data store. The deeper conceptual question is what happens to the **"prompts as input records"** philosophy described in `framework/PROMPTS.md`: the model where users write requirements into prompt files and agents read them as input. Skills are instruction sets, not data stores. This philosophy must be reconsidered.

## Key Design Decision: Input Model

This is the most important decision to resolve before implementation begins.

### Current model (prompts as input records)
Users write requirements into `.github/prompts/[domain].prompt.md`. Agents read the file contents as their primary source of requirements. The filled prompt file is committed alongside specs as an auditable input record.

### Skills constraint
Skills contain instructions for agents — not user data. The `SKILL.md` body is written by the skill author (the ADK), not by end users at runtime. Storing user requirements inside a skill file would conflate instructions with data.

### Proposed replacement model
Two separate concerns:

1. **Skill (instruction)** — `smaqit.new-agent/SKILL.md` instructs the agent on HOW to gather user input (what questions to ask, what to validate, what to compile). The skill is read-only for end users.

2. **User input (context only)** — Requirements are gathered interactively by the agent and held in conversation context. They are NOT written to a file as part of the skill flow. If auditability is needed, the agent documents gathered inputs in a compilation log (already the case for `smaqit.new-agent`).

**Consequence for downstream users (smaQit product):** The pattern of `smaqit.business.prompt.md` as a user-editable input record is dropped. Product agents that previously read from prompt files would instead gather requirements interactively or define their own non-skill input pattern (e.g., user-managed files in a `requirements/` directory that agents read — distinct from skills). This is a smaQit product concern, not ADK's to solve.

**ADK's responsibility:** Update the framework to describe the skills-based model, drop the "input record" principle, and be explicit that downstream users who need persistent requirement storage must design their own input pattern.

---

## Acceptance Criteria

### 1. File Migration: prompt → skill

- [ ] `prompts/smaqit.new-agent.prompt.md` removed
- [ ] `smaqit.new-agent` skill created at `.github/skills/smaqit.new-agent/SKILL.md` (following agentskills.io spec)
- [ ] Skill frontmatter: `name: smaqit.new-agent`, `description`, `metadata.version`
- [ ] Skill body contains instruction content from old prompt (questions to ask, gather steps, validation rules, compilation guidance)
- [ ] Skill body does NOT contain HTML comment examples in the old prompt style — those were for user guidance in a fill-in form; skills are pure instructions
- [ ] `installer/prompts/` directory and all prompt files removed
- [ ] `installer/skills/smaqit.new-agent/SKILL.md` added (synced copy for embedding)

### 2. Framework: PROMPTS.md → SKILLS.md

- [ ] `framework/PROMPTS.md` renamed to `framework/SKILLS.md`
- [ ] Content rewritten to describe the skills-based architecture: progressive disclosure, instruction-only model, context-driven user input
- [ ] "Prompts as input records" principle explicitly dropped, with a note on what replaces it (interactive gathering + compilation log)
- [ ] Skills taxonomy defined: workflow skills (session, task, release), domain skills (new-agent), and how downstream users should design their own input patterns
- [ ] `installer/framework/PROMPTS.md` renamed to `installer/framework/SKILLS.md` with same content

### 3. Agent Updates

#### L0 Agent (`agents/smaqit.L0.agent.md`)
- [ ] Input source list: `framework/PROMPTS.md` → `framework/SKILLS.md`
- [ ] Framework file inventory updated (PROMPTS.md → SKILLS.md)
- [ ] Level boundary examples: any prompt-as-input-record examples updated to reflect skill-based input model
- [ ] L0 principle updated: "Skill-Driven Input" replaces "Prompt-Driven Input" at the concept level

#### L1 Agent (`agents/smaqit.L1.agent.md`)
- [ ] Compiled directive updated: `MUST read from .github/prompts/[agent].prompt.md` → new skills-aware equivalent (e.g., `MUST gather requirements interactively following the smaqit.new-agent skill` or `MUST read from [user-defined input location]` depending on design decision)
- [ ] L1 examples and form distinctions updated to reflect new input pattern
- [ ] `[AGENT].prompt.md` placeholder replaced with skills-appropriate equivalent

#### L2 Agent (`agents/smaqit.L2.agent.md`)
- [ ] Input sources: `smaqit.new-agent.prompt.md` path references updated to skill path
- [ ] Compilation log remains the record of user-provided input (no change to this pattern)
- [ ] Any instruction to "follow structure from `.github/prompts/smaqit.new-agent.prompt.md`" updated to "activate the `smaqit.new-agent` skill"

#### Installer copies (sync after agent updates)
- [ ] `installer/agents/smaqit.L0.agent.md` synced
- [ ] `installer/agents/smaqit.L1.agent.md` synced
- [ ] `installer/agents/smaqit.L2.agent.md` synced

### 4. Installer: `main.go` and `Makefile`

- [ ] `main.go` embed directive: `//go:embed prompts/*.md` removed
- [ ] `main.go` embed directive added for skills (e.g., `//go:embed skills/smaqit.new-agent/SKILL.md`)
- [ ] `copyEmbeddedDir` call for prompts removed
- [ ] Skills copy logic added: scaffold `smaqit.new-agent` skill to `.github/skills/smaqit.new-agent/SKILL.md`
- [ ] Created directory list updated: `.github/prompts` → `.github/skills/smaqit.new-agent`
- [ ] Success message updated: "Copied prompts (new-agent template)" → "Copied skills (new-agent skill)"
- [ ] Next-steps message updated: no longer instructs user to fill `.github/prompts/smaqit.new-agent.prompt.md`; instead directs to invoke `/smaqit.L2` and interact with the new-agent skill
- [ ] Uninstall logic updated: removes `.github/skills/` instead of `.github/prompts/`
- [ ] `Makefile`: `prepare` target updated to copy `../skills/` instead of `../prompts/`; `clean` removes `skills/` not `prompts/`

### 5. Documentation

- [ ] `README.md`:
  - Directory tree updated: `prompts/` → `skills/smaqit.new-agent/`
  - Quick start step 2: "Edit `.github/prompts/smaqit.new-agent.prompt.md`" → "Invoke `/smaqit.L2` — the new-agent skill guides you interactively"
  - Framework file list: `PROMPTS.md` → `SKILLS.md`
- [ ] `docs/wiki/extending-smaqit.md`:
  - ADK project structure diagram updated
  - "Agent creation prompt" references replaced with "Agent creation skill"
  - Workflow description updated to reflect interactive skill-driven approach

### 6. Specification Compilation Rules

- [ ] `templates/agents/compiled/specification.rules.md`: Review for any prompt-file input directives; update to skill-based or interactive input model if present
- [ ] `templates/agents/compiled/implementation.rules.md`: Same review
- [ ] `templates/agents/compiled/base.rules.md`: Same review
- [ ] `installer/templates/agents/compiled/` copies synced

### 7. Validation

- [ ] `make clean build test` passes with no prompt references in scaffolded project
- [ ] Scaffolded project contains `.github/skills/smaqit.new-agent/SKILL.md`, NOT `.github/prompts/`
- [ ] Invoking `/smaqit.L2` in a scaffolded project activates the new-agent skill correctly
- [ ] No broken references to `PROMPTS.md` or `.github/prompts/` remain in any ADK file

---

## Scope Notes

**In scope (ADK):**
- The `smaqit.new-agent` skill (one-to-one replacement for `smaqit.new-agent.prompt.md`)
- Framework principle document (`PROMPTS.md` → `SKILLS.md`)
- All level agents (L0, L1, L2)
- Installer and build system
- README and wiki documentation

**Out of scope (smaQit product):**
- How smaQit's layer/phase agents (business, functional, stack, etc.) adapt their input model is a smaQit product decision, not ADK's
- Defining a universal "user requirements file" pattern — ADK can note the gap, but should not prescribe a product-specific solution

**Out of scope (skill registry / distribution):**
- Publishing to any skills registry
- CI/CD skill validation step (can be added later)

## Notes

The `smaqit.new-agent` skill is structurally the cleanest part of this migration. The prompt file was already instruction-only — L2 gathered user input interactively and documented it in the compilation log, never storing it back in the prompt file. The skill format formalises this intent.

The harder work is untangling the `Prompt-Driven Input` concept from `framework/AGENTS.md` and `framework/PROMPTS.md`, and updating L1's directive examples — since L1 currently compiles `MUST read from .github/prompts/[agent].prompt.md` as a representative directive for downstream agents. Once skills replace prompts as the input mechanism, this pattern needs a principled replacement.

**Recommended L1 replacement pattern (open for discussion):**
- Keep `MUST read from [user-defined input path]` as a generic directive at L1, leaving the concrete path as a user decision (not prescribed by ADK)
- Skills instruct agents HOW to gather input; where to persist it (if at all) is a downstream product decision
