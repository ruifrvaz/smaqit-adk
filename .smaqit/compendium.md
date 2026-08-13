# Project Compendium

## Architecture

**Where do skills and agents actually read templates and framework files from at runtime?**

`~/.agents/smaqit-adk/templates/` and `~/.agents/smaqit-adk/framework/` — global, not per-project. The top-level `templates/`/`framework/` directories in this repo are the ADK source repo's own pre-install locations; `install.sh` installs them globally via a hidden `--install-global` flag, and every current skill/agent that reads them (`agents/smaqit.L2.md`'s Input section, `smaqit.create-agent`'s Failure Handling table) documents the `~/.agents/smaqit-adk/`-prefixed path as the one it expects. There is no project-local copy. Any tool that provisions a HOME for testing purposes must copy to `~/.agents/smaqit-adk/templates`/`~/.agents/smaqit-adk/framework`, or reads will silently fail.

---

**What platforms does smaqit-adk compile for, and where does each global install target land?**

Claude Code and Codex CLI are the two authored, compiled targets; GitHub Copilot is not — it has no dedicated compiled agent file for `smaqit.L0`/`L1`/`L2` or for any `smaqit.create-agent`-produced custom agent. Copilot compatibility comes entirely from two standards it already reads natively: root `AGENTS.md` (canonical instructions) and the shared `~/.agents/skills/` path. A single global install (triggered by `install.sh`) writes: `smaqit.L0`/`L1`/`L2` to `~/.claude/agents/` (as `smaqit-L0.md` etc., dash-separated) and `~/.codex/agents/` (as `smaqit.L0.toml` etc., dot-separated, matching each platform's own naming convention); all 5 ADK skills to `~/.agents/skills/` (shared — read natively by both Copilot and Codex) and `~/.claude/skills/` (Claude-specific, identical content); templates/framework to `~/.agents/smaqit-adk/`. `CLAUDE_CONFIG_DIR`/`CODEX_HOME` env vars override the respective roots if set. No smaqit-adk artifact is ever written into a project directory — a user-created custom agent (via `smaqit.create-agent`) lands project-locally instead, at `.claude/agents/[name].md` and `.codex/agents/[name].toml`; a user-created custom skill (via `smaqit.create-skill`) lands at `.agents/skills/[name]/SKILL.md` and `.claude/skills/[name]/SKILL.md`.

When a skill needs to invoke `smaqit.L0`/`L1`/`L2`, the routing is platform-conditional: a native subagent/custom-agent call on Claude Code or Codex CLI, or — on Copilot, which has no compiled file to call — reading the Claude-format body directly (e.g. `~/.claude/agents/smaqit-L2.md`) and following it inline for the current turn.

---

**Does sandboxing `$HOME` for isolated testing work for live Codex CLI verification?**

Only for testing the *installer's* file-placement behavior, not for testing anything that needs to actually talk to the Codex API. Codex CLI's stored authentication lives under `$HOME/.codex/` (or `$CODEX_HOME` if set) alongside its config and registered custom agents — sandboxing `HOME` to an isolated temp directory to keep tests from touching the real machine also strips away that auth, so any live `codex exec` call fails with `401 Unauthorized` before it ever reaches the behavior under test. Verifying real Codex behavior (e.g. whether a globally-registered `~/.codex/agents/*.toml` custom agent actually spawns correctly) requires either running against the real, unsandboxed `$HOME` (and cleaning up afterward), or pre-populating a sandboxed `$CODEX_HOME` with real credentials before the test. Codex CLI does genuinely support spawning a custom agent registered this way via `codex exec` non-interactively (confirmed live, `codex-cli 0.147.0`) — the agent's `developer_instructions` load correctly and it responds as itself.

---

**What is Bench, and where does it live?**

Bench (`smaqit-adk bench`) is smaqit-adk's local process-evaluation engine. Its implementation is the independent Go module under `src/bench/` and `src/benchcli/`; `installer/` packages the resulting command but does not own the engine. It uses strict, plan-first manifests (`bench.yaml`) and a generic local `process` adapter, so a configured CLI harness such as Codex can receive declared prompts, specs, files, directories, and images while deterministic expectations and hidden graders evaluate frozen outputs. One variant is a plain evaluation; two or more variants (e.g. with-artifact vs. without-artifact) produce a comparison. `smaqit-adk bench suite validate|plan|run <directory>` discovers every `bench.yaml` under a directory tree and drives each through the same pipeline, aggregating pass/fail/error counts across manifests.

Do not call this "HarnessBench" — that was only ever early-planning shorthand and is not the product name.

---

## Bench Manifests

**Where does smaqit-adk's own dogfood benchmark suite live, and how is it organized?**

`.smaqit/bench/` — this is smaqit-adk using its own shipped Bench engine to benchmark its own skills and agents. It is project-local data, not ADK-shipped source: it is not installed into a consuming project by the global installer, the same way `.smaqit/tasks/` or `.smaqit/compendium.md` are local project state rather than shipped artifacts. Skill benchmarks live at `.smaqit/bench/skills/<skill-id>/bench.yaml`; agent benchmarks (flat files, no natural directory of their own) live at `.smaqit/bench/agents/<agent-id>/bench.yaml`. Run output goes under `.smaqit/bench/runs/`, gitignored. See `.smaqit/bench/README.md` for the full layout convention and `.smaqit/bench/MIGRATION.md` for how the legacy eval suite's scenarios map onto it.

None of the dogfood manifests spawn a Level agent as a real registered custom agent — doing so would require a real global install and real Codex/Claude credentials on every run, defeating the disposable-workspace model. Instead, a Level agent's body is staged as a read-only input (`given.files`/`given.directories`) and the prompt tells the harness to read it directly and apply its instructions inline, the same fallback mechanism used for GitHub Copilot's routing (see `.smaqit/tasks/029_repair_dogfood_bench_manifests.md` if a manifest's staging approach is unclear).

---

**How does a Bench manifest compare a skill/agent's behavior with vs. without the artifact staged?**

One Case, two Variants — not two separate Cases. Both variants share the same `expect` block; they're differentiated by staging: the with-artifact variant's `setup` (or a case-level input asset) makes the skill/agent file available to the harness, and the without-artifact variant either omits that staging or removes it before the harness runs. This lets Bench's native comparison logic (grouped by variant ID) produce a win/tie/inconclusive outcome with no custom aggregation code. Staging the file alone is not sufficient for the harness to actually use it — a live-verified gotcha is that a generic prompt (e.g. "create a skill") can let a coding-agent harness fall back to its own built-in tooling instead of reading the staged file. Prompts must reference the staged artifact explicitly and unambiguously — see "How do `given.files`/`given.directories` stage content, and where does the harness actually see it?" for the correct phrasing and why file-discovery-based phrasing (e.g. "if staged at a project-relative path") is unreliable outside smaqit-adk's own repo.

---

**How do `given.files`/`given.directories` stage content, and where does the harness actually see it?**

Declared inputs (`given.files`, `given.directories`, `given.specs`, `given.images`) are copied into a **read-only** directory (`.smaqit-bench-input/`) beside the harness's actual working directory, not merged into the working tree itself — even when an explicit `destination` is set, that path is still resolved relative to the read-only input root, never the workspace root. The harness only ever reaches staged content through the resolved `{input:<id>}` absolute-path placeholder (available in prompts, `setup` commands, and process arguments). A prompt that says "if staged at a project-relative path, read it" is not reliable in general. The generic, portable pattern — the one every current dogfood manifest follows — is to reference `{input:<id>}` explicitly in the prompt (e.g. "the skill is staged at `{input:skill}/SKILL.md` — read it first and follow it exactly"), and file-discovery habits like `rg --files` hiding dotfiles never come into play.

If a case needs the harness to *edit* the staged content rather than just read it (e.g. a principle-authoring skill editing framework files in place), staging alone is not enough — add a `setup:` step that copies the staged read-only input into the workspace at its real, writable path first (e.g. `mkdir -p {workspace}/framework && cp {input:framework-src}/*.md {workspace}/framework/`), then have the prompt reference the writable copy at its normal relative path, not `{input:<id>}`.

---

**Why does a Bench `command`-type expectation or grader need an explicit `environment` block, and what's the `go run` gotcha?**

`Setup` commands, command-type graders, and command-type expectations run with a completely empty environment by default — not even `PATH` or `HOME` — unless the manifest sets `command.environment.inherit`/`.set`. Most POSIX tools (`sh`, `grep`, `test`, `rm`) don't need one, but anything that does (like `go run`) will fail without it. On a Snap-packaged Go toolchain (`/snap/bin/go -> /usr/bin/snap`, common on Ubuntu), `go run` fails even with a correct environment, because Bench's process-group isolation (needed for reliable timeout/kill handling) collides with Snap's own confinement locking. The reliable fix for a Go-based grader is to pre-compile it into a plain binary at build time and point the command grader at that binary, rather than invoking `go run` during grading.

---

## Skill Authoring

**How are new ADK-shipped skills (root `skills/`) actually authored, and does `smaqit.new-skill` exist?**

`smaqit.new-skill` does not exist — it's referenced in `README.md`'s "ADK Source (Expert Use)" section as the contributor-facing tool for this, but no such file exists anywhere in the repo or at any installed location. The working path is manual: write a specification to `.smaqit/definitions/skills/[name].md` (the same shape `smaqit.create-skill` would produce — identity, steps with fragility levels, output, scope, completion, failure handling, gotchas, examples), then invoke `smaqit.L2` to compile it, reading `agents/smaqit.L2.md`, `templates/skills/base-skill.template.md`, and `templates/skills/compiled/skill.rules.md`. When compiling inside smaqit-adk's own source repo (as opposed to compiling for a separate consuming project), L2's own "For Skills" procedure writes the result to root `skills/[name]/SKILL.md` — this is L2's native, default output path per its own agent file, not a workaround. `smaqit.create-skill`'s project-local output paths (`.agents/skills/[name]/SKILL.md`, `.claude/skills/[name]/SKILL.md`) are only its own prompt-level override of that default, used specifically because it compiles skills *for a different, consuming project*; they don't apply when the ADK is compiling its own shipped skills for itself. Hand-writing a `SKILL.md` file directly and only validating it with `scripts/validate-skill.go` is not equivalent to this — passing structural validation does not mean the skill was produced through the ADK's own compilation chain.

---

**What does `skills/smaqit.create-skill/scripts/validate-skill.go` actually check, and is it worth keeping?**

It enforces the ADK's SKILL.md format rules: frontmatter (`name` pattern, `description` length/anti-pattern checks — no first-person, no "Use when..." gating language), required sections (Steps, Output, Scope, Completion, Failure Handling), a 500-line body cap, no unresolved `[PLACEHOLDER]`-style tokens, and a well-formed Failure Handling table (header + separator + ≥2 data rows). It mirrors the same rules enforced by `tests/structural/skills_test.go`, so it's a genuinely shared source of truth rather than a duplicate check, and it's load-bearing in three separate places: it gates `smaqit.create-skill`'s own compile step, it works standalone as a general-purpose linter for any `SKILL.md`, and its pre-compiled form (`installer/dist/validate-skill`, built specifically to avoid the Snap/`go run` toolchain conflict) is the actual pass/fail grader inside `.smaqit/bench/skills/smaqit.create-skill/bench.yaml`. A parallel "validate-bench.go" for `bench.yaml` manifests is not similarly justified: `smaqit-adk bench validate` already is the engine's own authoritative manifest loader/validator (schema fields, ID uniqueness, safe paths, oracle isolation, placeholder resolution) — the structural gap `validate-skill.go` exists to fill for skills simply doesn't exist for manifests. The narrower thing a bespoke bench linter could catch (missing `--sandbox`/`--skip-git-repo-check` flags, `{input:<id>}` misuse, `type: text` where `type: command` would fail more gracefully) is currently handled as documented convention in `smaqit.bench-scaffold`'s own instructions rather than code, pending evidence that mistakes repeat enough to justify automating the check.

---

## Git & Release Workflow

**Why does `git push` get rejected for touching `.github/workflows/*.yml`, and how do I fix it?**

GitHub blocks any push that creates or modifies a file under `.github/workflows/` unless the pushing credential has workflow-write permission — this applies even to single-line changes. The fix depends on the credential type: a **classic** personal access token needs the `workflow` OAuth scope, addable via `gh auth refresh -s workflow`. A **fine-grained** personal access token (identifiable by its `github_pat_...` prefix, visible via `gh auth status`) needs the "Workflows" repository permission (Read and write) added directly on the token itself at GitHub's personal-access-token settings page — `gh auth refresh -s workflow` does not apply to fine-grained tokens. No local re-authentication is needed after granting the permission; the existing token starts working on the next push once GitHub applies the change.
