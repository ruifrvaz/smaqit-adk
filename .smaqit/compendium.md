# Project Compendium

## Architecture

**Where do skills and agents actually read templates and framework files from at runtime?**

`.smaqit/templates/` and `.smaqit/framework/` — not the top-level `templates/` and `framework/` directories. The top-level directories are the ADK source repo's own pre-install locations; `smaqit-adk lite`/`advanced` install them into `.smaqit/` in a consuming project, and every current skill/agent that reads them (`smaqit.L2.agent.md`'s Input section, `smaqit.create-agent`'s Failure Handling table) documents the `.smaqit/`-prefixed path as the one it expects. A tool that provisions a workspace for testing (e.g. `tests/evals/runner/main.go`'s `setupWorkspace()`) must copy to `.smaqit/templates`/`.smaqit/framework`, not the top level, or reads will silently fail.

---

**Where will HarnessBench (the harness A/B benchmarking tool, Task 023) live in the codebase?**

As a `bench` subcommand of the existing `smaqit-adk` binary (`installer/bench/`), not a separate module/binary. This was a deliberate choice against the alternative of a dependency-free separate module — the `smaqit-adk` installer currently has zero external Go dependencies and every `curl | bash` user downloads it, so adding `bench` brings in `copilot-sdk/go` and a YAML parser. The trade-off was made explicitly, favoring one distributable and discoverability over installer size. Phase 1 scope is deliberately small: variants, repetitions, deterministic graders, statistics, and winner selection, driving the Copilot SDK in-process — no external harness process adapters, no worker-process boundary, no git-diff metrics yet.

---

## Eval Testing

**Why did `tests/evals/runner/main.go` leak `copilot --headless` processes across a run?**

Every `copilot.NewClient(...)` call spawns and owns a real OS subprocess; the Go SDK's own package documentation requires `defer client.Stop()` immediately after creation. The runner created a `copilot.Client` twice per eval file — once for the eval session itself, once per graded criterion inside `grade()` — and never called `.Stop()` at either site, so every session's process stayed alive indefinitely. A run with several evals and several criteria each could leak 20+ processes, which in turn caused resource contention severe enough to make later sessions in the same run time out or hang. Fixed by adding `defer client.Stop()` at both `NewClient()` call sites.

---
