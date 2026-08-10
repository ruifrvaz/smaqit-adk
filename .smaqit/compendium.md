# Project Compendium

## Architecture

**Where do skills and agents actually read templates and framework files from at runtime?**

`.smaqit/templates/` and `.smaqit/framework/` — not the top-level `templates/` and `framework/` directories. The top-level directories are the ADK source repo's own pre-install locations; `smaqit-adk lite`/`advanced` install them into `.smaqit/` in a consuming project, and every current skill/agent that reads them (`smaqit.L2.agent.md`'s Input section, `smaqit.create-agent`'s Failure Handling table) documents the `.smaqit/`-prefixed path as the one it expects. A tool that provisions a workspace for testing (e.g. `tests/evals/runner/main.go`'s `setupWorkspace()`) must copy to `.smaqit/templates`/`.smaqit/framework`, not the top level, or reads will silently fail.

---

**Where does HarnessBench live, and how will it evaluate skills and agents?**

HarnessBench is the `smaqit-adk bench` subcommand. Its implementation is the independent Go source module under `src/bench/` and `src/benchcli/`; `installer/` packages the resulting command but does not own the engine. Bench uses strict, plan-first manifests and a generic local `process` adapter, so a configured CLI harness such as Codex can receive declared prompts, specs, files, directories, and images while deterministic expectations and hidden graders evaluate frozen outputs.

The planned repository convention colocates skill benchmarks under `skills/<skill-id>/evals/` and agent benchmarks under `agents/evals/<agent-id>/`. Those suites will replace the legacy Copilot SDK JSON eval runner; until native discovery is demonstrated, a benchmark explicitly makes the tested skill or agent available to the configured harness and compares it with an artifact-absent baseline.

---

## Eval Testing

**Why did `tests/evals/runner/main.go` leak `copilot --headless` processes across a run?**

Every `copilot.NewClient(...)` call spawns and owns a real OS subprocess; the Go SDK's own package documentation requires `defer client.Stop()` immediately after creation. The runner created a `copilot.Client` twice per eval file — once for the eval session itself, once per graded criterion inside `grade()` — and never called `.Stop()` at either site, so every session's process stayed alive indefinitely. A run with several evals and several criteria each could leak 20+ processes, which in turn caused resource contention severe enough to make later sessions in the same run time out or hang. Fixed by adding `defer client.Stop()` at both `NewClient()` call sites.

---
