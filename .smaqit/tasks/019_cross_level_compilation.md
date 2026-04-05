# Cross-Level Compilation (smaqit.compile)

**Status:** Not Started  
**Created:** 2026-04-05

## Description

Design and author a set of `smaqit.compile` skills that run the full L0 → L1 → L2 compilation chain within VS Code, using sequential subagent invocations. When a new principle is added or a template rule changes, this chain ensures the change cascades down to existing agents. Each phase's output files serve as the next phase's input context.

This is the VS Code-native successor to the CLI session-chaining approach from Task 015 (deferred).

## Motivation

Without a cross-level skill, a user who adds a new principle via `smaqit.new-principle` must manually know to then invoke `smaqit.new-rules` (L1) and then re-run `smaqit.new-agent` (L2) to recompile affected agents. The `smaqit.compile` skill set makes that cascade explicit and guided.

## Scope

Three compile entry points (naming confirmed during design):

| Skill | Chain | Use when |
|-------|-------|---------|
| `smaqit.compile.principle` | L0 → L1 → L2 | New principle that must propagate to templates and agents |
| `smaqit.compile.template` | L1 → L2 | New or changed template that must propagate to agents |
| `smaqit.compile.agent` | L2 only | Recompiling an existing agent against updated rules (no principle/template change) |

## Architecture

Each phase runs as an isolated subagent invocation. The skill collects the output path from each phase before invoking the next, passing file paths as context:

```
smaqit.compile.principle:
  Phase 1: invoke smaqit.L0 → output: framework/[name].md (or update to existing file)
  Phase 2: invoke smaqit.L1 with Phase 1 output → output: templates/.../[name].rules.md
  Phase 3: invoke smaqit.L2 with Phase 1+2 outputs → output: .github/agents/[name].agent.md
```

## Acceptance Criteria

- [ ] Skill naming convention decided and documented (`smaqit.compile.*` or alternative)
- [ ] `smaqit.compile.principle` chains L0 → L1 → L2 with output file hand-off between phases
- [ ] `smaqit.compile.template` chains L1 → L2
- [ ] `smaqit.compile.agent` invokes L2 only
- [ ] Each phase is an isolated subagent invocation — no context bleed between phases
- [ ] User receives clear phase progress indicators within the skill
- [ ] All compile skills shipped by `smaqit-adk advanced`
- [ ] `make build` passes cleanly

## Dependencies

- Task 018 (Level Skills Completion) — all Level agents must have definition file input patterns before a chain can be composed from them

## Notes

- This supersedes the VS Code surface of Task 015 (CLI chain in Go); the skill approach keeps the chain in agent space
- The CLI chain (Task 015) may still be worth implementing for CI/CD use cases — that is a separate decision
- Compile skills are cross-cutting: they depend on all three Level agents being fully operational with definition file patterns (ensured by Task 018)
- `smaqit.compile.agent` has no Level skill dependency — it only needs L2, which already has a definition file pattern
