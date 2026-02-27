# Distill AGENTS-old into AGENTS.md

**Status:** Not Started
**Created:** 2026-02-27

## Description

`framework/AGENTS-old.md` is the original smaQit product AGENTS file. It is heavily contaminated with product-domain content (specification agents for BUS/FUN/STK layers, implementation agents for develop/deploy/validate phases, `smaqit plan --phase` CLI commands, spec frontmatter tracking, cross-layer consolidation). However, it contains several ADK-level principles that are either missing from `framework/AGENTS.md` or stated more weakly.

This task distills the valuable ADK-level content from AGENTS-old into AGENTS.md, then removes the old file. It is a prerequisite for Task 005 (framework redesign) because the enriched AGENTS.md is the model that other framework files should align with.

**What to distill (confirmed ADK-level content):**

| Principle | Current state in AGENTS.md | What AGENTS-old adds |
|-----------|---------------------------|----------------------|
| Boundary Enforcement | "Agents decline out-of-scope work with clear redirection" — one line | 3-step procedure: (1) stop immediately, don't plan or create todos, (2) state current scope + correct agent, (3) suggest next step |
| Self-Validation Loop | "Agents validate their output before finishing" — one sentence | Explicit numbered loop: produce → check → iterate if unmet → declare complete → flag blocker if impossible |
| Quality Boundary | Missing | Stop when: all criteria met OR blocking issue OR clarification required. MUST NOT iterate indefinitely or lower standards. |
| Failure Modes | Missing | Situation/action table: ambiguous input, conflicting requirements, missing upstream, impossible requirement |
| Agent-Skill invocation | Missing | When agents detect ambiguity or complexity, invoke the appropriate skill rather than implementing inline |
| Assumption flagging | "Agents do not invent requirements" | Add: agents flag assumptions explicitly when clarification is unavailable |
| Role architecture pattern | Missing | Compiled agents' Role section must contain: identity statement + goal + context, 3–4 sentences max |
| Tool descriptions | Listed without description | Full description of what each tool does |
| Extensibility through templates | Covered but briefly | "What must remain invariant → base template. What varies by role → extension templates." Cleaner articulation. |

**What to explicitly exclude (product-domain contamination):**
- Layer/phase names (BUS/FUN/STK/INF/COV, develop/deploy/validate)
- Specification agent and implementation agent sections
- `smaqit plan --phase` CLI commands
- Spec frontmatter lifecycle tracking
- Cross-layer consolidation and phase orchestration
- `.github/prompts/smaqit.[layer].prompt.md` references
- Specification/implementation agent mapping tables
- `smaqit.business`, `smaqit.functional` etc. naming examples

## Acceptance Criteria

- [ ] All 9 distilled principles above are incorporated into `framework/AGENTS.md`
- [ ] Each new or enriched principle follows the `### Name` / bold one-liner / explanation paragraph pattern from SMAQIT.md
- [ ] No product-domain content (layer names, phase names, prompt file paths, CLI commands) is introduced
- [ ] `framework/AGENTS-old.md` is deleted after distillation
- [ ] AGENTS.md remains coherent — no contradictions between existing content and newly added content
- [ ] Installer sync deferred to `make build` (no manual copy)

## Notes

- The Role architecture pattern (identity + goal + context) is relevant to L2 compilation guidance — once distilled into AGENTS.md, it may also be worth referencing in the L2 agent or the base agent template. Assess during execution.
- AGENTS-old.md should be deleted (not archived) once distillation is complete — it is a historical artifact, not a shipped file.
