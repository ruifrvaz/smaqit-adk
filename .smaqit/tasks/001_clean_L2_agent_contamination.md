# Clean L2 Agent Contamination

**Status:** Completed  
**Created:** 2026-02-27  
**Completed:** 2026-02-27

## Description

Remove smaQit product-specific contamination from `agents/smaqit.L2.agent.md`. The L2 agent was partially cleaned during initial ADK extraction (only a 1-line rename was applied), leaving its body full of references to smaQit's 5-layer + 3-phase model. This work fully generalized the agent to reflect the ADK's domain-agnostic compilation model.

## Acceptance Criteria

- [x] No references to non-existent layer/phase rules files (business.rules.md, functional.rules.md, etc.)
- [x] No references to non-existent product agents (smaqit.business.agent.md, etc.)
- [x] Hardcoded Placeholder Resolution tables replaced with generic guidance
- [x] Compilation Architecture describes 3-way/4-way merge with optional domain/phase rules
- [x] All examples use generic domains (security, billing, compliance) instead of smaQit layers
- [x] Form Distinctions and Self-Containment Validation examples are generic
- [x] `[LAYER]`/`[LAYER_PREFIX]`/`[LAYER_NAME]` placeholders replaced with `[DOMAIN]`/`[PREFIX]`/`[DOMAIN_NAME]`
- [x] Installer copy (`installer/agents/smaqit.L2.agent.md`) synced
- [x] Build and installation test passes (`make clean build test`)

## Notes

Assessed as "Option A: Full Generic Rewrite" — removing all layer/phase-specific references rather than minimal removal, to align with ADK's "generic by design" philosophy. Approximately 30+ targeted replacements across the file. L0 and L1 agents were already clean; this completes the level agent cleanup begun during the initial ADK extraction.
