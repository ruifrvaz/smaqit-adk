---
name: smaqit.bench-scaffold
description: Authors a new `.smaqit/bench/` manifest for a skill or agent that doesn't have one yet — detects the project's skill/agent root generically, drafts a self-contained single-shot Codex prompt that stages the target artifact as a variant treatment, structurally validates the result, and delegates any optional live trial to `smaqit.bench-run`.
metadata:
  version: "0.2.0"
---

# Bench Scaffold

Guides authoring a new `.smaqit/bench/{skills,agents}/<id>/bench.yaml` manifest for a skill or agent lacking one, working generically against any consuming project's layout.

Structural authoring and validation require only the smaqit-adk binary. An optional live trial also requires an authenticated `codex` executable on `PATH` and is delegated to `smaqit.bench-run`.

## Steps

### 1. Detect the skill/agent root

Check `.agents/skills/` + `.claude/skills/` and `.claude/agents/`/`.codex/agents/` first — the project-local locations `smaqit.create-agent`/`smaqit.create-skill` write a consuming project's own custom artifacts to (smaqit-adk itself installs nothing into any project directory; only project-local custom agents/skills ever live there). If none exist, fall back to root `skills/`/`agents/` (smaqit-adk's own dev-repo convention, used by its dogfood suite). If neither location resolves unambiguously, ask the user where the project's skills/agents live.

### 2. Select target

List targets under the detected root that have no `.smaqit/bench/{skills,agents}/<id>/bench.yaml`. If the user's chosen target already has a manifest, ask whether to add a case to it instead of creating a new one.

### 3. Understand the target

Read the target's `SKILL.md` or agent source `.md` file in full — its purpose, steps, and any subagent it invokes.

### 4. Draft the manifest

Write `.smaqit/bench/{skills,agents}/<id>/bench.yaml` (plus any `prompts/*.md` files it references) with:

**Case prompts** — self-contained, single-shot. Bench's `process` adapter is non-interactive with no multi-turn `ask_user` relay: pre-supply any answer a human would normally give across turns, and tell the harness explicitly not to wait for further input.

**Staging the target artifact** — for a with/without comparison, declare it on the with-artifact variant as a treatment:

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

Bench copies treatments into the read-only `.smaqit-bench-input/` sidecar — **not** at the artifact's real project-relative path. The rendered Case brief appends a `# Variant treatment artifacts` table with each available ID and resolved absolute path. Prompt text is preserved verbatim, so the prompt **must** name the treatment ID and direct the harness to the path listed there:

> "If the variant treatment table lists `skill`, read its SKILL.md first and follow it exactly. If the treatment set is empty, do not search global ADK paths."

Do not assume a project-relative artifact path or prompt-body placeholder interpolation. Use the Case brief's treatment table and conditional wording so the same raw prompt is accurate for both variants.

**Without-artifact variant** — omit `treatment`; the Case brief will render an explicit empty set. Never remove or copy evaluation artifacts with shell setup commands.

**When the target edits files, not just reads them** — declare a common writable fixture and place it at its conventional workspace-relative destination:

```yaml
fixture:
  source: ../../../../framework
  destination: framework
```

Use Case-level `prepare` only for deterministic common preparation that must run after fixture copy and before the baseline snapshot. Preparation is shared by every variant and can use `{workspace}` and `{caseId}` only; it must never encode a hidden treatment.

**Expectations** — prefer `command`-type checks over bare `text`-type for anything that might not exist; a `text` expectation crashes rather than failing gracefully against a missing file.

**Reusable Codex process block** — copy verbatim from `.smaqit/bench/README.md`'s "Reusable Codex process-variant block" (pins `--sandbox danger-full-access`, `--skip-git-repo-check`, and an explicit `timeoutSeconds`). Do not hand-roll a variant of it.

### 5. Structural validation

```bash
smaqit-adk bench validate .smaqit/bench/{skills,agents}/<id>/bench.yaml
```

Fix diagnostics before proceeding. Do not offer a live trial against a manifest that fails structural validation.

### 6. Optional live trial

Ask whether to run a live trial now. If yes, invoke `smaqit.bench-run`, scoped to the new manifest — reuse its preflight/confirm/execute/diagnose logic rather than duplicating it here. If a live trial surfaces a genuine `src/bench` engine limitation (not a known gotcha), report it precisely and recommend a follow-up task; do not patch the engine inline from this skill.

### 7. Report

Report the new manifest path (and any `prompts/*.md` files), its structural validation status, and the live-trial result if one was run.

## Output

- `.smaqit/bench/{skills,agents}/<id>/bench.yaml` (+ any `prompts/*.md` it references)
- Structural validation status
- Live-trial result, if a trial was run
- No subagent invoked for authoring; a live trial delegates to `smaqit.bench-run`

## Scope

Authors manifests only. Does not execute a full suite run or reimplement `smaqit.bench-run`'s preflight/confirmation/diagnosis logic — any live trial delegates to it. Does not modify `src/bench` engine code to work around a discovered limitation; stops and reports instead.

## Examples

**Input:** User invokes `smaqit.bench-scaffold` in a project where `.claude/skills/my-skill/SKILL.md` exists with no corresponding `.smaqit/bench/skills/my-skill/bench.yaml`.

**Output:** The skill detects `.claude/skills/` as the root, lists `my-skill` as an uncovered target, reads `my-skill/SKILL.md`, drafts `.smaqit/bench/skills/my-skill/bench.yaml` staging `my-skill` as the with-artifact treatment with conditional treatment-table wording in the prompt and the reusable Codex block, runs `bench validate` (passes), and asks whether to run a live trial via `smaqit.bench-run`.

## Gotchas

- `treatment` and shared `given` assets stage read-only in the Bench sidecar, never at the artifact's literal project-relative path — see Step 4. Use the correct Case-brief table in prompts. Getting this wrong produces a manifest that structurally validates but fails every live run.
- The reusable Codex process block in `.smaqit/bench/README.md` is the canonical, tested source — copy it, don't reconstruct it from memory.

## Completion

- [ ] Skill/agent root detected (project-local `.agents/skills|.claude/skills|.claude/agents|.codex/agents` or root `skills|agents/`, or resolved by asking)
- [ ] Target selected; confirmed it has no existing manifest, or the user chose to extend one
- [ ] Target's `SKILL.md`/agent source `.md` file read in full before drafting
- [ ] Manifest stages the target artifact via the with-artifact variant's `treatment`, with conditional prompt wording that names the treatment ID and table
- [ ] Reusable Codex process block copied verbatim, not hand-rolled
- [ ] `bench validate` passes on the new manifest
- [ ] Live trial offered; if accepted, delegated to `smaqit.bench-run`

## Failure Handling

| Situation | Action |
|-----------|--------|
| Neither the project-local skill/agent locations nor root `skills\|agents/` resolves unambiguously | Ask the user where the project's skills/agents live |
| Chosen target already has a manifest | Ask whether to add a case to it instead of creating a new manifest |
| A manifest, `prompts/*.md` file, or other output artifact already exists at the write path | Confirm with the user before overwriting |
| Structural validation fails | Report diagnostics; fix the manifest before offering a live trial |
| User declines a live trial | Report the manifest as scaffolded-but-unverified; stop cleanly |
| Invoking `smaqit.bench-run` for the live trial fails outright | Report the failure with context; do not silently retry |
| Live trial (via `smaqit.bench-run`) surfaces a genuine `src/bench` limitation | Report it precisely with the reproducing case; recommend a follow-up task rather than patching the engine inline |
| Target's `SKILL.md`/agent source `.md` file is missing or unreadable | Stop; report the path that could not be read |
