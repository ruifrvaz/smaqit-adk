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

**Which comparison shapes can Bench express today, and what must stay controlled?**

Bench executes `Cases × Variants × repetitions`; a single Variant is an evaluation, while named Variants in one manifest are the candidates in a comparison. Artifact A/B uses variant treatments. Prompt A/B keeps one shared Case prompt and supplies each candidate prompt as a same-ID variant treatment. Model A/B and harness/version A/B differ through each Variant's explicit process executable, arguments, or non-secret configured environment. A factorial comparison is one explicitly named Variant per candidate tuple, such as `model-a__prompt-b`.

Keep the Case prompt, fixture, preparation, shared inputs, expectations, graders, budgets, and non-candidate environment fixed. A suite aggregates independent manifests but never establishes a winner across them; place every claimed comparison in one manifest. Bench currently ranks compound Variant IDs, rather than validating factor completeness or calculating per-factor effects. See `docs/wiki/benchmarking.md` for the canonical scenario matrix once it is added.

---

**Why does a Bench `command`-type expectation or grader need an explicit `environment` block, and what's the `go run` gotcha?**

`Setup` commands, command-type graders, and command-type expectations run with a completely empty environment by default — not even `PATH` or `HOME` — unless the manifest sets `command.environment.inherit`/`.set`. Most POSIX tools (`sh`, `grep`, `test`, `rm`) don't need one, but anything that does (like `go run`) will fail without it. On a Snap-packaged Go toolchain (`/snap/bin/go -> /usr/bin/snap`, common on Ubuntu), `go run` fails even with a correct environment, because Bench's process-group isolation (needed for reliable timeout/kill handling) collides with Snap's own confinement locking. The reliable fix for a Go-based grader is to pre-compile it into a plain binary at build time and point the command grader at that binary, rather than invoking `go run` during grading.

---

**Does a Bench harness need to clean up its own workspace, and what happens if it leaves behind a non-regular file (socket, FIFO)?**

No manual cleanup belongs in a manifest — `src/bench/run.go`'s `runAttempt` unconditionally `defer`s a full removal of the entire ephemeral workspace (`removeWorkspace`, an `os.RemoveAll`-shaped call) after every run, regardless of pass, fail, or timeout. A manifest's own `process`/`prepare` commands should never add their own teardown logic; it's redundant at best.

**Updated by Task 037 (2026-10-18, folding in Task 035):** both former gaps are fixed. The freeze step records a socket, FIFO, or device the harness left behind (for example `codex exec`'s own IPC socket) as a `special` entry in `final-tree.json` — never digested, counted as `specialFiles`, and excluded from the frozen submission — instead of crashing with a raw `ENXIO` (or, for a FIFO, blocking forever on `open()`). Cleanup is always reported: removing the workspace and the per-attempt harness home is attempted on every exit path, and any failure lands in `result.json`'s `cleanup`, `report.md`'s "Cleanup failures" section, the `attempt.completed` event, and a CLI warning, and marks the attempt not passing, even when an earlier phase already failed. Harness-owned state (home/config directories, copied credentials) belongs in `{harnessHome}`, a private directory outside the workspace that is never frozen — never in `{workspace}`.

---

**How does Bench run Claude Code, and why can't a plain `process` variant do it safely?**

Through `adapter: claude-code` (Task 037). Claude Code discovers skills, agents, and commands only from a project's `.claude/` and its config directory, and personal skills override project skills of the same name — so on a machine with smaqit's skills installed globally, a non-isolated run silently tests the installed copy instead of the treatment. The adapter gives every attempt an isolated `HOME` and `CLAUDE_CONFIG_DIR` inside the harness home, installs declared treatments there (`install: [{treatment: <id>, as: skill|agent|command}]`), fails the attempt on a same-named definition in the fixture's or an ancestor's `.claude/` or a managed location, and asserts every installed treatment appears in the harness's `system/init` event. It runs `claude -p --output-format stream-json --verbose` with the Case brief on stdin, captures the transcript into `traces/transcript.jsonl`, parses token/cost/tool-call usage, records the harness version in the plan and report, and treats a final `is_error: true` as failure even on exit 0. Exactly one credential variable (`CLAUDE_CODE_OAUTH_TOKEN` or `ANTHROPIC_API_KEY`) passes through; every other `CLAUDE_*`/`ANTHROPIC_*` variable — including the invoking Claude Code session's own `CLAUDE_CODE_SESSION_ID`/messaging socket — is absent, and a post-attempt canary scan redacts and fails on any echoed credential. For live runs on Linux/WSL, the invoking user's own login token is loaded inline, never via Vault: `CLAUDE_CODE_OAUTH_TOKEN="$(jq -r .claudeAiOauth.accessToken ~/.claude/.credentials.json)" smaqit-adk bench run …` — live-verified working with an empty config dir (Claude Code 2.1.296). `--bare` can't be used because it disables the discovery treatments depend on. `.smaqit/bench/scenarios/claude-code-isolation/` is the live proof.

---

## Registry Propagation

**How does `smaqit.new-principle`'s live cascade actually propagate a principle change?**

`smaqit.new-principle` owns the entire cascade itself — it is not agent-to-agent chaining. After invoking `smaqit.L0` to land the principle (base scope: `framework/*.md`; extension scope: a consuming project's `.smaqit/framework/*.md`), the skill invokes `smaqit.L1` as a single-hop `Task` call to judge whether any `templates/agents/*.template.md`, `templates/agents/compiled/*.rules.md`, or `templates/skills/` equivalent needs updating — "no update needed" is a valid, common outcome that ends the cascade there. Only if `smaqit.L1` actually touches a file does the skill query the registry (`smaqit-adk registry get --scope base|extension --depends-on <path>`) for dependents, then invoke `smaqit.L2` once per dependent — again a single-hop `Task` call from the skill's own turn — to recompile and re-register it. Neither `smaqit.L0` nor `smaqit.L1` holds a `Task` tool grant; no 2-hop agent-to-agent subagent chain exists anywhere in smaqit-adk or the sibling `smaqit` product, so every hop originates from the orchestrating skill, mirroring `smaqit.development`'s Business/Functional/Stack orchestration pattern. The registry only has edges from an L1 template/rules file to the L2 output compiled from it — never a direct edge from an L0 framework file — which is why L1's judgment step must run before the registry lookup, not be skipped in favor of it.

---

**What is the registry, and what are its two scopes?**

A Terraform-statefile-style JSON file (structs defined in `installer/registry.go`) tracking every ADK-compiled agent/skill's target path(s), source definition file, source template/rules dependencies, and a compiled-content hash. Two scopes exist: `base` (`~/.agents/smaqit-adk/registry.json`) tracks smaqit-adk's own `smaqit.L0`/`L1`/`L2` and its shipped skills against the base framework; `extension` (`.smaqit/registry.json` in a consuming project) tracks that project's own `smaqit.create-agent`/`create-skill` output. `smaqit.L2` writes an entry uniformly on every compile via the hidden `smaqit-adk registry put` subcommand, regardless of which skill invoked it — this is the only registration path; no other skill duplicates it. `smaqit-adk sync` (the CLI) can only report drift — a plain binary has no way to invoke an LLM for the semantic recompile a drifted entry needs. `smaqit.sync` (the skill) is the interactive counterpart that actually resolves it: it shells out to the same registry CLI, then lets the user choose per drifted entry to recompile from the recorded source or keep a hand edit and re-register it.

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
