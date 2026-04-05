# Full Compilation Chain CLI (L0→L1→L2)

**Status:** Deferred  
**Created:** 2026-03-31  
**Deferred:** 2026-04-05  
**Reason:** CLI work paused in favor of VS Code-native skill approach. Cross-level compilation will be handled by `smaqit.compile` skills (Task 019), which supersedes this on the VS Code surface. CLI chain may be revisited for CI/CD use cases independently.

## Description

Design and implement a `smaqit-adk compile` command (or `--full` flag on `create-agent`) that runs the complete ADK compilation chain: L0 (principle) → L1 (rules) → L2 (agent). Each phase runs as a focused SDK session; outputs are file-chained in Go and injected as context into the next session. All ADK artifacts (L0, L1, L2 agents, framework files, templates) live in the binary — nothing is written to the user's project except the final compilation artifacts.

## Motivation

The ADK's core value proposition is the compilation chain. VS Code users can do this conversationally by switching agent contexts, but the CLI provides a programmatic, repeatable path for teams and CI workflows. The user experience goal: one command that walks the user through principle definition, rule compilation, and agent output in a single terminal session.

## Architecture

### Session chain

```
Phase 1 — L0 session
  system:  embedded(smaqit.L0.agent.md) + embedded(framework/*.md)
  purpose: gather a new principle from the user; write it to disk
  output:  .smaqit/framework/[name].md

Phase 2 — L1 session
  system:  embedded(smaqit.L1.agent.md) + embedded(templates/**) + content of phase 1 output (injected by Go)
  purpose: compile the new principle into template directives
  output:  .smaqit/templates/compiled/[name].rules.md

Phase 3 — L2 session
  system:  embedded(smaqit.L2.agent.md) + content of phase 1+2 outputs (injected by Go)
  purpose: gather agent spec; compile using new rules; write agent file
  output:  .github/agents/[name].agent.md
```

**Key constraint:** ADK infrastructure (agents, framework, templates) stays in the binary. The user's project only receives phase outputs.

### Phase hand-off

Go reads each phase's output file after `SendAndWait` completes, then injects the content as additional user/system context for the next session. No subagent tool calls needed.

### Entry point options (decide during design)

Option A: `smaqit-adk compile` — dedicated command for the full chain  
Option B: `smaqit-adk create-agent --full` — flag on existing command  
Option C: `smaqit-adk create-agent` always runs full chain; `--lite` for the current self-contained path  

## Acceptance Criteria

- [ ] Design decision on entry point documented (Option A/B/C or hybrid)
- [ ] Phase 1 (L0 session) gathers a principle and writes `.smaqit/framework/[name].md`
- [ ] Phase 2 (L1 session) receives phase 1 output and writes `.smaqit/templates/compiled/[name].rules.md`
- [ ] Phase 3 (L2 session) receives phase 1+2 outputs and writes `.github/agents/[name].agent.md`
- [ ] User can run the full chain without any ADK source files present in their project
- [ ] User can also run `create-agent` (lite path) without going through L0/L1 phases
- [ ] CLI prints clear phase headers so the user knows which phase they're in
- [ ] `make build` passes cleanly
- [ ] End-to-end manual test passes (principle → rules → agent, all files written correctly)

## Dependencies

- Task 006 (smaqit.new-principle skill) is related but not blocking: this task designs the CLI chain, which may inform or supersede Task 006's VS Code skill approach
- Task 013 (create-principle CLI command) overlaps with Phase 1 of this task — consider merging or sequencing

## Notes

- Each phase should suppress the progress ticker and have no timeout (same decisions as Task 014)
- Phase outputs are intermediate artifacts owned by the user's project — `.smaqit/` is the right location (already used for definitions and logs by L2)
- The `--output` flag on existing `create-agent` sets the final agent output dir; keep this for the full chain too
- Eval targets: each phase independently testable with existing eval runner infrastructure
