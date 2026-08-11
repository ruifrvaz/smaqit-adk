# Project Compendium

## Architecture

**Where do skills and agents actually read templates and framework files from at runtime?**

`.smaqit/templates/` and `.smaqit/framework/` — not the top-level `templates/` and `framework/` directories. The top-level directories are the ADK source repo's own pre-install locations; `smaqit-adk lite`/`advanced` install them into `.smaqit/` in a consuming project, and every current skill/agent that reads them (`smaqit.L2.agent.md`'s Input section, `smaqit.create-agent`'s Failure Handling table) documents the `.smaqit/`-prefixed path as the one it expects. Any tool that provisions a workspace for testing purposes must copy to `.smaqit/templates`/`.smaqit/framework`, not the top level, or reads will silently fail.

---

**What is Bench, and where does it live?**

Bench (`smaqit-adk bench`) is smaqit-adk's local process-evaluation engine. Its implementation is the independent Go module under `src/bench/` and `src/benchcli/`; `installer/` packages the resulting command but does not own the engine. It uses strict, plan-first manifests (`bench.yaml`) and a generic local `process` adapter, so a configured CLI harness such as Codex can receive declared prompts, specs, files, directories, and images while deterministic expectations and hidden graders evaluate frozen outputs. One variant is a plain evaluation; two or more variants (e.g. with-artifact vs. without-artifact) produce a comparison. `smaqit-adk bench suite validate|plan|run <directory>` discovers every `bench.yaml` under a directory tree and drives each through the same pipeline, aggregating pass/fail/error counts across manifests.

Do not call this "HarnessBench" — that was only ever early-planning shorthand and is not the product name.

---

## Bench Manifests

**Where does smaqit-adk's own dogfood benchmark suite live, and how is it organized?**

`.smaqit/bench/` — this is smaqit-adk using its own shipped Bench engine to benchmark its own skills and agents. It is project-local data, not ADK-shipped source: it is not installed into a consuming project by `smaqit-adk lite`/`advanced`, the same way `.smaqit/tasks/` or `.smaqit/compendium.md` are local project state rather than shipped artifacts. Skill benchmarks live at `.smaqit/bench/skills/<skill-id>/bench.yaml`; agent benchmarks (flat files, no natural directory of their own) live at `.smaqit/bench/agents/<agent-id>/bench.yaml`. Run output goes under `.smaqit/bench/runs/`, gitignored. See `.smaqit/bench/README.md` for the full layout convention and `.smaqit/bench/MIGRATION.md` for how the legacy eval suite's scenarios map onto it.

---

**How does a Bench manifest compare a skill/agent's behavior with vs. without the artifact staged?**

One Case, two Variants — not two separate Cases. Both variants share the same `expect` block; they're differentiated by staging: the with-artifact variant's `setup` (or a case-level input asset) makes the skill/agent file available in the disposable workspace, and the without-artifact variant either omits that staging or removes it before the harness runs. This lets Bench's native comparison logic (grouped by variant ID) produce a win/tie/inconclusive outcome with no custom aggregation code. Staging the file alone is not sufficient for the harness to actually use it — a live-verified gotcha is that a generic prompt (e.g. "create a skill") can let a coding-agent harness fall back to its own built-in tooling instead of reading the staged file, and file-discovery habits like `rg --files` hide dotfiles (`.github/`, `.smaqit/`) by default. Prompts must explicitly and conditionally point at the staged path (e.g. "if this project has an ADK skill-authoring skill staged at `.github/skills/<id>/SKILL.md`, read it first and follow it exactly").

---

**Why does a Bench `command`-type expectation or grader need an explicit `environment` block, and what's the `go run` gotcha?**

`Setup` commands, command-type graders, and command-type expectations run with a completely empty environment by default — not even `PATH` or `HOME` — unless the manifest sets `command.environment.inherit`/`.set`. Most POSIX tools (`sh`, `grep`, `test`, `rm`) don't need one, but anything that does (like `go run`) will fail without it. On a Snap-packaged Go toolchain (`/snap/bin/go -> /usr/bin/snap`, common on Ubuntu), `go run` fails even with a correct environment, because Bench's process-group isolation (needed for reliable timeout/kill handling) collides with Snap's own confinement locking. The reliable fix for a Go-based grader is to pre-compile it into a plain binary at build time and point the command grader at that binary, rather than invoking `go run` during grading.

---

## Git & Release Workflow

**Why does `git push` get rejected for touching `.github/workflows/*.yml`, and how do I fix it?**

GitHub blocks any push that creates or modifies a file under `.github/workflows/` unless the pushing credential has workflow-write permission — this applies even to single-line changes. The fix depends on the credential type: a **classic** personal access token needs the `workflow` OAuth scope, addable via `gh auth refresh -s workflow`. A **fine-grained** personal access token (identifiable by its `github_pat_...` prefix, visible via `gh auth status`) needs the "Workflows" repository permission (Read and write) added directly on the token itself at GitHub's personal-access-token settings page — `gh auth refresh -s workflow` does not apply to fine-grained tokens. No local re-authentication is needed after granting the permission; the existing token starts working on the next push once GitHub applies the change.
