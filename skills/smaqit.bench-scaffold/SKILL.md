---
name: smaqit.bench-scaffold
description: Authors a new `.smaqit/bench/` manifest for a skill or agent that doesn't have one yet — detects the project's skill/agent root generically, selects the harness (Codex CLI or Claude Code), drafts a self-contained single-shot prompt that stages the target artifact as a variant treatment, structurally validates the result, and delegates any optional live trial to `smaqit.bench-run`.
metadata:
  version: "0.3.0"
compatibility: Requires no additional runtime dependencies beyond the smaqit-adk binary for authoring and validation; an optional live trial requires the chosen harness CLI (codex, or claude with a credential), delegated to smaqit.bench-run.
---

# Bench Scaffold

## Steps

### 1. Detect the skill/agent root

Check `.agents/skills/` + `.claude/skills/` and `.claude/agents/`/`.codex/agents/` first — the project-local locations `smaqit.create-agent`/`smaqit.create-skill` write a consuming project's own custom artifacts to (smaqit-adk itself installs nothing into any project directory; only project-local custom agents/skills ever live there). If none exist, fall back to root `skills/`/`agents/` (smaqit-adk's own dev-repo convention, used by its dogfood suite). If neither resolves unambiguously, ask the user where the project's skills/agents live.

### 2. Select target

List targets under the detected root that have no `.smaqit/bench/{skills,agents}/<id>/bench.yaml`. If the chosen target already has a manifest, ask whether to add a case to it instead of creating a new one.

### 3. Select the harness

Ask which harness the manifest should exercise, recommending one from the target's platform behavior:

| Harness | Variant form | Recommend when |
|---------|--------------|----------------|
| Claude Code | `adapter: claude-code` | The target depends on Claude-Code-specific behavior (skill discovery and precedence, subagent `skills:` preload, `Task`/`SendMessage`, Claude tool names), or the user wants Claude Code results |
| Codex CLI | `process` + the reusable Codex block | The target is exercised through Codex |

If the user wants both, author one variant pair per harness in the same manifest (for example `codex-with-artifact`, `codex-baseline`, `claude-with-artifact`, `claude-baseline`) and note that a cross-harness winner compares harness *and* artifact together.

### 4. Understand the target

Read the target's `SKILL.md` or agent source `.md` file in full — its purpose, steps, and any subagent it invokes.

### 5. Draft the manifest

Write `.smaqit/bench/{skills,agents}/<id>/bench.yaml` (plus any `prompts/*.md` files it references) with:

**Case prompts** — self-contained, single-shot. Neither harness has a multi-turn relay in Bench: pre-supply any answer a human would normally give across turns, and tell the harness explicitly not to wait for further input. (Claude Code subagents never get `AskUserQuestion`; a headless main agent has no one to answer it.)

**Staging the target artifact (Codex / `process`)** — declare it on the with-artifact variant as a treatment:

```yaml
variants:
  - id: with-artifact
    adapter: process
    treatment:
      - id: skill
        source: ../../../../skills/<target-id>   # relative to this bench.yaml
    intendedDifferences:
      - Exposes the target skill as the variant treatment.
```

Bench copies treatments into the read-only `.smaqit-bench-input/` sidecar — **not** at the artifact's real project-relative path. The rendered Case brief appends a `# Variant treatment artifacts` table with each available ID and resolved absolute path. Prompt text is preserved verbatim, so the prompt **must** name the treatment ID and direct the harness to the path listed there, conditionally:

> "If the variant treatment table lists `skill`, read its SKILL.md first and follow it exactly. If the treatment set is empty, do not search global ADK paths."

Do not assume a project-relative path or prompt-body placeholder interpolation.

**Staging the target artifact (Claude Code)** — declare the same `treatment`, and also an `install` entry in `claudeCode` so Claude Code natively discovers it:

```yaml
    claudeCode:
      install:
        - {treatment: skill, as: skill}
```

`as: skill` takes a directory containing `SKILL.md`; `as: agent` takes a `.md` file with a frontmatter `name`; `as: command` takes a `.md` file. Bench installs it into the attempt's isolated config directory, so the prompt references the skill or agent by its name (for example "invoke the `<target-id>` skill with the Skill tool") instead of a sidecar path. Never stage it under the fixture's `.claude/` — a same-named definition there fails the run as shadowing.

**Without-artifact variant** — omit `treatment` (and `install`); the Case brief renders an explicit empty set. Never remove or copy evaluation artifacts with shell setup commands.

**When the target edits files, not just reads them** — declare a common writable Case fixture at a safe workspace-relative destination:

```yaml
fixture:
  source: ../../../../framework
  destination: framework
```

Use Case-level `prepare` only for deterministic common preparation after fixture copy and before the baseline snapshot. It is shared by every variant, can use `{workspace}` and `{caseId}` only, and must never encode a hidden treatment.

**Expectations** — prefer `command`-type checks over bare `text`-type for anything that might not exist; a `text` expectation crashes rather than failing gracefully against a missing file.

**Codex variants** — copy the reusable Codex process block verbatim from `.smaqit/bench/README.md`'s "Reusable Codex process-variant block" (pins `--sandbox danger-full-access`, `--skip-git-repo-check`, and an explicit `timeoutSeconds`). Do not hand-roll a variant of it. If the harness needs its own home or credentials, stage them under `{harnessHome}` (never `{workspace}`).

**Claude Code variants** — follow `.smaqit/bench/README.md`'s "Reusable Claude Code variant block":

- `adapter: claude-code` with a `claudeCode` block setting:
  - `permissionMode` — prefer `dontAsk` plus a minimal `allowedTools` list
  - `model` — pinned, for comparable runs
  - `maxTurns` — a bound
  - `credential: CLAUDE_CODE_OAUTH_TOKEN` (or `ANTHROPIC_API_KEY`)
  - `install`
- An explicit `execution.timeoutSeconds`.
- Never pass `--bare` or any adapter-owned flag in `arguments`; never put credentials or `HOME`/`CLAUDE_*`/`ANTHROPIC_*` in `environment`.
- A header comment showing the inline credential loader used to run it.

### 6. Structural validation

```bash
smaqit-adk bench validate .smaqit/bench/{skills,agents}/<id>/bench.yaml
```

Fix diagnostics before proceeding. Do not offer a live trial against a manifest that fails structural validation.

### 7. Optional live trial

Ask whether to run a live trial now. If yes, invoke `smaqit.bench-run` scoped to the new manifest — reuse its preflight/confirm/execute/diagnose logic (including Claude Code credential loading) rather than duplicating it. If a live trial surfaces a genuine `src/bench` engine limitation (not a known gotcha), report it precisely and recommend a follow-up task; never patch the engine inline from this skill.

### 8. Report

Report the new manifest path (and any `prompts/*.md` files), the harness(es) it targets, its structural validation status, and the live-trial result if one was run.

## Output

- `.smaqit/bench/{skills,agents}/<id>/bench.yaml` (+ any `prompts/*.md` it references)
- The harness(es) the manifest targets
- Structural validation status
- Live-trial result, if a trial was run
- No subagent invoked for authoring; a live trial delegates to `smaqit.bench-run`

## Scope

Authors manifests only. Does not execute a full suite run or reimplement `smaqit.bench-run`'s preflight/confirmation/diagnosis logic — any live trial delegates to it. Does not modify `src/bench` engine code to work around a discovered limitation; stops and reports instead. Generic multi-dimension scenario matrices (prompt, model, version, factorial) are out of scope.

## Examples

**Input:** User invokes `smaqit.bench-scaffold` in a project where `.claude/skills/my-skill/SKILL.md` exists with no corresponding `.smaqit/bench/skills/my-skill/bench.yaml`, and chooses Codex CLI.

**Output:** The skill detects `.claude/skills/` as the root, lists `my-skill` as an uncovered target, reads `my-skill/SKILL.md`, drafts `.smaqit/bench/skills/my-skill/bench.yaml` staging `my-skill` as the with-artifact treatment with conditional treatment-table wording in the prompt and the reusable Codex block, runs `bench validate` (passes), and asks whether to run a live trial via `smaqit.bench-run`.

**Input:** Same project, but the skill preloads into a subagent via `skills:` frontmatter, so the user picks Claude Code.

**Output:** The skill drafts two `adapter: claude-code` variants — `with-artifact` declaring `treatment: [{id: skill, source: ...}]` plus `install: [{treatment: skill, as: skill}]`, and a `baseline` with neither — using `permissionMode: dontAsk`, a minimal `allowedTools`, a pinned model, `maxTurns`, and `credential: CLAUDE_CODE_OAUTH_TOKEN`, with a prompt that names `my-skill` directly. `bench validate` passes and the skill offers a live trial via `smaqit.bench-run`.

## Gotchas

- `treatment` and shared `given` assets stage read-only in the Bench sidecar, never at the artifact's literal project-relative path — see Step 5. Use the correct Case-brief table in Codex prompts. Getting this wrong produces a manifest that structurally validates but fails every live run.
- For Claude Code, the sidecar copy is not enough: without an `install` entry the treatment is not discovered as a skill/agent/command. The globally installed personal copy of the same name never interferes, because Bench isolates `CLAUDE_CONFIG_DIR` per attempt.
- The reusable Codex process block and the Claude Code variant block in `.smaqit/bench/README.md` are the canonical, tested sources — copy them, don't reconstruct them from memory.

## Completion

- [ ] Skill/agent root detected (project-local `.agents/skills|.claude/skills|.claude/agents|.codex/agents` or root `skills|agents/`, or resolved by asking)
- [ ] Target selected; confirmed it has no existing manifest, or the user chose to extend one
- [ ] Harness selected (Codex CLI, Claude Code, or both)
- [ ] Target's `SKILL.md`/agent source `.md` file read in full before drafting
- [ ] Manifest stages the target artifact via the with-artifact variant's `treatment` (plus `install` for Claude Code), with prompt wording that names the treatment ID/table (Codex) or the installed name (Claude Code)
- [ ] Reusable Codex process block copied verbatim, or the Claude Code variant follows the README's block, not hand-rolled
- [ ] `bench validate` passes on the new manifest
- [ ] Live trial offered; if accepted, delegated to `smaqit.bench-run`

## Failure Handling

| Situation | Action |
|-----------|--------|
| Required input not provided | Request the missing information before proceeding |
| Gathered input is ambiguous | Flag the ambiguity and ask for clarification |
| Neither the project-local skill/agent locations nor root `skills\|agents/` resolves unambiguously | Ask the user where the project's skills/agents live |
| Chosen target already has a manifest | Ask whether to add a case to it instead of creating a new manifest |
| A manifest, `prompts/*.md` file, or other output artifact already exists at the write path | Confirm with the user before overwriting |
| The user can't decide on a harness | Recommend one from the target's platform behavior (Step 3) and proceed only on confirmation |
| Target's `SKILL.md`/agent source `.md` file is missing or unreadable | Stop; report the path that could not be read |
| Structural validation fails | Report diagnostics; fix the manifest before offering a live trial |
| User declines a live trial | Report the manifest as scaffolded-but-unverified; stop cleanly |
| Invoking `smaqit.bench-run` for the live trial fails outright | Report the failure with context; do not silently retry |
| Live trial (via `smaqit.bench-run`) surfaces a genuine `src/bench` limitation | Report it precisely with the reproducing case; recommend a follow-up task rather than patching the engine inline |
