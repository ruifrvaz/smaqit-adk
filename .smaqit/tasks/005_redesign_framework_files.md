# Redesign Framework Files

**Status:** Not Started
**Created:** 2026-02-27
**Depends on:** Task 004 (AGENTS.md must be enriched before other files align to it)

## Description

The `framework/` directory contains five files that L0 is responsible for maintaining. Three of them — `SMAQIT.md`, `ARTIFACTS.md`, and most of `TEMPLATES.md` — describe the **smaQit product** domain model, not the ADK itself. They define product concepts (Business/Functional/Stack/Infrastructure/Coverage specification layers, phase workflows, prompt files, BUS/FUN/STK/INF/COV requirement identifiers, traceability matrices, spec lifecycle states, coverage translation) that belong in a smaQit product extension, not in the ADK framework that ships to all consumers.

L0's role is to maintain framework principles for the ADK compilation chain. The current files make L0 responsible for product-domain principles it has no business owning, and they create confusion about what the ADK is versus what the smaQit product built with it is.

**Files and their current status:**

| File | Status | Problem |
|------|--------|---------|
| `AGENTS.md` | ✅ Good | Accurately describes ADK Level agents and invocation model |
| `SKILLS.md` | ✅ Good | Updated this session — covers skill principles and structure |
| `SMAQIT.md` | 🔴 Replace | Entirely describes smaQit product principles (layers, phases, spec-driven development). Not ADK content. |
| `ARTIFACTS.md` | 🔴 Replace | Describes smaQit product artifact rules (spec IDs, lifecycle states, traceability, coverage). Not ADK content. |
| `TEMPLATES.md` | 🟡 Partial | Agent template section is ADK-relevant. Specification and prompt template sections are product content. |

## What the ADK framework should contain

The ADK framework principles should define how the ADK itself works:

- **Compilation chain** — L0 → L1 → L2, what each level does, what it reads and produces
- **Agent model** — What an agent is in the ADK context, the three template types, foundational behaviors (already in AGENTS.md)
- **Skill model** — Progressive disclosure, description quality, instruction-only content (already in SKILLS.md)
- **Template model** — How ADK templates work, placeholder conventions, hierarchy (foundation → extension), compilation rules
- **Artifact model** — What ADK artifacts are (agent files, compilation logs, definition files — not smaQit product specs)
- **Extension model** — How consumers extend the ADK (L0/L1/L2 routes, product agents, custom skills)

## Acceptance Criteria

- [ ] `SMAQIT.md` replaced with ADK-scoped content: describes the smaqit-adk compilation philosophy and principles (not the smaQit product model)
- [ ] `ARTIFACTS.md` replaced with ADK-scoped content: describes ADK artifact types (agent files, definition files, compilation logs, template files) and their rules
- [ ] `TEMPLATES.md` scoped to ADK template model only: agent template structure, placeholder conventions, compilation hierarchy — removes spec and prompt template sections (product content)
- [ ] No framework file references smaQit product-domain concepts (BUS/FUN/STK/INF/COV layers, prompt files, spec lifecycle states)
- [ ] All principles follow the `### Principle Name` / bold one-liner / explanation paragraph pattern established in the current `SMAQIT.md` and applied to `SKILLS.md`
- [ ] L0, L1, L2 agent files updated if their `## Framework Reference` sections point to removed content
- [ ] README and wiki updated if they reference framework file content that changed
- [ ] Installer synced via `make build`

## Notes

- The smaQit product content in the current files is not lost — it belongs in a smaQit product extension (e.g., `.github/` of a smaQit product consumer). The ADK's role is to provide the compilation machinery, not to prescribe the product domain.
- `AGENTS.md` and `SKILLS.md` are good models to follow for tone, scope, and principle formatting.
- This task does NOT include creating new smaQit product extension files — only removing product content from the ADK framework layer.
