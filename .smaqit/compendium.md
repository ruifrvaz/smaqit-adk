# Project Compendium

## Architecture

**Where do skills and agents actually read templates and framework files from at runtime?**

`~/.agents/smaqit-adk/templates/` and `~/.agents/smaqit-adk/framework/` — global, not per-project, as of Task 027 (2026-08-13). The top-level `templates/`/`framework/` directories in this repo are the ADK source repo's own pre-install locations; `install.sh` installs them globally via `--install-global`, and every current skill/agent that reads them (`agents/smaqit.L2.md`'s Input section, `smaqit.create-agent`'s Failure Handling table) documents the `~/.agents/smaqit-adk/`-prefixed path as the one it expects. This superseded the prior per-project `.smaqit/templates/`/`.smaqit/framework/` model entirely — there is no project-local copy anymore. Any tool that provisions a HOME for testing purposes must copy to `~/.agents/smaqit-adk/templates`/`~/.agents/smaqit-adk/framework`, or reads will silently fail.

---

**What is Bench, and where does it live?**

Bench (`smaqit-adk bench`) is smaqit-adk's local process-evaluation engine. Its implementation is the independent Go module under `src/bench/` and `src/benchcli/`; `installer/` packages the resulting command but does not own the engine. It uses strict, plan-first manifests (`bench.yaml`) and a generic local `process` adapter, so a configured CLI harness such as Codex can receive declared prompts, specs, files, directories, and images while deterministic expectations and hidden graders evaluate frozen outputs. A case's raw `given.prompt` is rendered with its declared-input locations into a **case brief**, keeping Bench terminology distinct from smaqit's first-class Task lifecycle. One variant is a plain evaluation; two or more variants (e.g. with-artifact vs. without-artifact) produce a comparison. `smaqit-adk bench suite validate|plan|run <directory>` discovers every `bench.yaml` under a directory tree and drives each through the same pipeline, aggregating pass/fail/error counts across manifests.

Do not call this "HarnessBench" — that was only ever early-planning shorthand and is not the product name.

---

## Bench Manifests

**Where does smaqit-adk's own dogfood benchmark suite live, and how is it organized?**

`.smaqit/bench/` — this is smaqit-adk using its own shipped Bench engine to benchmark its own skills and agents. It is project-local data, not ADK-shipped source: it is not installed into a consuming project by the global installer, the same way `.smaqit/tasks/` or `.smaqit/compendium.md` are local project state rather than shipped artifacts. Skill benchmarks live at `.smaqit/bench/skills/<skill-id>/bench.yaml`; agent benchmarks (flat files, no natural directory of their own) live at `.smaqit/bench/agents/<agent-id>/bench.yaml`. Run output goes under `.smaqit/bench/runs/`, gitignored. See `.smaqit/bench/README.md` for the full layout convention and `.smaqit/bench/MIGRATION.md` for how the legacy eval suite's scenarios map onto it.

**Updated by Task 029 (2026-08-14):** all four dogfood manifests use schema version 2, case briefs (`{briefFile}`), common writable fixtures where needed, and variant-only read-only treatments for the skill/agent guidance under test. They assert the current Claude `.md` and Codex `.toml` outputs (and both cross-platform skill copies); the complete live suite finished with zero harness errors or timeouts. These resources are controlled evaluation fixtures, not an ADK installation.

---

**How does a Bench manifest compare a skill/agent's behavior with vs. without the artifact staged?**

One Case, two Variants — not two separate Cases. Both variants share the same `expect` block; the with-artifact variant declares a `treatment`, while the baseline declares none. Bench stages only that variant's treatment in its read-only sidecar, so no setup command needs to copy, delete, or chmod guidance. This lets Bench's native comparison logic (grouped by variant ID) produce a win/tie/inconclusive outcome with no custom aggregation code. Staging the artifact alone is not sufficient for the harness to actually use it — a live-verified gotcha is that a generic prompt (e.g. "create a skill") can let a coding-agent harness fall back to its own built-in tooling instead of reading the treatment. Prompts must name the treatment ID explicitly and conditionally in the Case brief.

---

**How do `given.files`/`given.directories` stage content, and where does the harness actually see it?**

Shared declared inputs (`given.files`, `given.directories`, `given.specs`, `given.images`) are copied under Bench's read-only, submission-excluded sidecar. A variant-only `treatment` is staged alongside them only for that variant. The Case brief presents separate tables with resolved paths for shared inputs and treatments; raw `given.prompt` text is deliberately preserved, so it must name the relevant ID and direct the harness to read the path listed in the appropriate table. `{input:<id>}` and `{treatment:<id>}` expand only in permitted process-command arguments, not in prompt text.

If a case needs the harness to *edit* common source material (for example, a principle-authoring skill modifying framework files), declare a Case `fixture` with a safe workspace-relative `destination`. Bench copies it before the baseline snapshot and makes it writable for every variant. Case-level `prepare` may perform deterministic, common preparation before sidecar staging, but it cannot encode a variant treatment.

---

**Why does a Bench `command`-type expectation or grader need an explicit `environment` block, and what's the `go run` gotcha?**

`Setup` commands, command-type graders, and command-type expectations run with a completely empty environment by default — not even `PATH` or `HOME` — unless the manifest sets `command.environment.inherit`/`.set`. Most POSIX tools (`sh`, `grep`, `test`, `rm`) don't need one, but anything that does (like `go run`) will fail without it. On a Snap-packaged Go toolchain (`/snap/bin/go -> /usr/bin/snap`, common on Ubuntu), `go run` fails even with a correct environment, because Bench's process-group isolation (needed for reliable timeout/kill handling) collides with Snap's own confinement locking. The reliable fix for a Go-based grader is to pre-compile it into a plain binary at build time and point the command grader at that binary, rather than invoking `go run` during grading.

---

## Skill Authoring

**How are new ADK-shipped skills (root `skills/`) actually authored, and does `smaqit.new-skill` exist?**

`smaqit.new-skill` does not exist — it's referenced in `README.md`'s "ADK Source (Expert Use)" section as the contributor-facing tool for this, but no such file exists anywhere in the repo or at any installed location. The working path is manual: write a specification to `.smaqit/definitions/skills/[name].md` (the same shape `smaqit.create-skill` would produce — identity, steps with fragility levels, output, scope, completion, failure handling, gotchas, examples), then invoke `smaqit.L2` as a subagent to compile it, reading `agents/smaqit.L2.md`, `templates/skills/base-skill.template.md`, and `templates/skills/compiled/skill.rules.md`. When compiling inside smaqit-adk's own source repo (as opposed to compiling for a separate consuming project), L2 writes the result to root `skills/[name]/SKILL.md`; for a consumer project it writes identical `.agents/skills/[name]/SKILL.md` and `.claude/skills/[name]/SKILL.md` copies. Hand-writing a `SKILL.md` file directly and only validating it with `scripts/validate-skill.go` is not equivalent to this — passing structural validation does not mean the skill was produced through the ADK's own compilation chain.

---

**What does `skills/smaqit.create-skill/scripts/validate-skill.go` actually check, and is it worth keeping?**

It enforces the ADK's SKILL.md format rules: frontmatter (`name` pattern, `description` length/anti-pattern checks — no first-person, no "Use when..." gating language), required sections (Steps, Output, Scope, Completion, Failure Handling), a 500-line body cap, no unresolved `[PLACEHOLDER]`-style tokens, and a well-formed Failure Handling table (header + separator + ≥2 data rows). It mirrors the same rules enforced by `tests/structural/skills_test.go`, so it's a genuinely shared source of truth rather than a duplicate check, and it's load-bearing in three separate places: it gates `smaqit.create-skill`'s own compile step, it works standalone as a general-purpose linter for any `SKILL.md`, and its pre-compiled form (`installer/dist/validate-skill`, built specifically to avoid the Snap/`go run` toolchain conflict) is the actual pass/fail grader inside `.smaqit/bench/skills/smaqit.create-skill/bench.yaml`. A parallel "validate-bench.go" for `bench.yaml` manifests is not similarly justified: `smaqit-adk bench validate` already is the engine's own authoritative manifest loader/validator (schema fields, ID uniqueness, safe paths, oracle isolation, placeholder resolution) — the structural gap `validate-skill.go` exists to fill for skills simply doesn't exist for manifests. The narrower thing a bespoke bench linter could catch (missing `--sandbox`/`--skip-git-repo-check` flags, `{input:<id>}` misuse, `type: text` where `type: command` would fail more gracefully) is currently handled as documented convention in `smaqit.bench-scaffold`'s own instructions rather than code, pending evidence that mistakes repeat enough to justify automating the check.

---

## Git & Release Workflow

**Why does `git push` get rejected for touching `.github/workflows/*.yml`, and how do I fix it?**

GitHub blocks any push that creates or modifies a file under `.github/workflows/` unless the pushing credential has workflow-write permission — this applies even to single-line changes. The fix depends on the credential type: a **classic** personal access token needs the `workflow` OAuth scope, addable via `gh auth refresh -s workflow`. A **fine-grained** personal access token (identifiable by its `github_pat_...` prefix, visible via `gh auth status`) needs the "Workflows" repository permission (Read and write) added directly on the token itself at GitHub's personal-access-token settings page — `gh auth refresh -s workflow` does not apply to fine-grained tokens. No local re-authentication is needed after granting the permission; the existing token starts working on the next push once GitHub applies the change.
