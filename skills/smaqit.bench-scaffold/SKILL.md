---
name: smaqit.bench-scaffold
description: Authors a new `.smaqit/bench/` manifest for a skill or agent that doesn't have one yet — detects the project's skill/agent root generically, drafts a self-contained single-shot Codex prompt that stages the target artifact as a declared Bench input, structurally validates the result, and delegates any optional live trial to `smaqit.bench-run`.
metadata:
  version: "0.1.0"
compatibility: Requires no additional runtime dependencies beyond the smaqit-adk binary for authoring/validation; an optional live trial requires codex on PATH, delegated to smaqit.bench-run.
---

# Bench Scaffold

Guides authoring a new `.smaqit/bench/{skills,agents}/<id>/bench.yaml` manifest for a skill or agent lacking one, working generically against any consuming project's layout.

## Steps

### 1. Detect the skill/agent root

Check `.github/skills/` + `.github/agents/` first — the standard install location for `smaqit-adk lite`/`advanced` in a consuming project. If neither exists, fall back to root `skills/`/`agents/` (smaqit-adk's own dev-repo convention). If neither location resolves unambiguously, ask the user where the project's skills/agents live.

### 2. Select target

List targets under the detected root that have no `.smaqit/bench/{skills,agents}/<id>/bench.yaml`. If the user's chosen target already has a manifest, ask whether to add a case to it instead of creating a new one.

### 3. Understand the target

Read the target's `SKILL.md` or `.agent.md` in full — its purpose, steps, and any subagent it invokes.

### 4. Draft the manifest

Write `.smaqit/bench/{skills,agents}/<id>/bench.yaml` (plus any `prompts/*.md` files it references) with:

**Case prompts** — self-contained, single-shot. Bench's `process` adapter is non-interactive with no multi-turn `ask_user` relay: pre-supply any answer a human would normally give across turns, and tell the harness explicitly not to wait for further input.

**Staging the target artifact** — declare it as a case-level `given.files` (single-file agent) or `given.directories` (skill directory) input asset:

```yaml
given:
  prompt:
    file: prompts/case-name.md
  directories:
    - id: skill
      source: ../../../../skills/<target-id>   # relative to this bench.yaml
```

Bench copies declared inputs into a read-only `.smaqit-bench-input/` directory beside the actual harness workspace — **not** at the artifact's real project-relative path (`.github/skills/<id>/SKILL.md` inside the workspace itself). The harness only ever sees the staged copy through the resolved `{input:<id>}` absolute-path placeholder. The prompt **must** reference that placeholder explicitly:

> "The project's ADK skill-authoring skill is staged at `{input:skill}/SKILL.md` — read it first and follow it exactly. Do not invent your own approach instead."

Do not phrase this as "if staged at `.github/skills/<id>/SKILL.md`, read it" — that phrasing only works for the three manifests Task 026 built for smaqit-adk's own repo, which stage via a `setup:` command running the actual ADK installer (`smaqit-adk-dev lite {workspace}`) directly against the workspace root instead of a declared input asset. That trick is a documented special case, not something a shipped skill can assume works in an arbitrary consuming project — this skill's manifests default to `given.files`/`given.directories` plus explicit `{input:<id>}` prompting instead.

**Without-artifact variant** (only when the manifest is doing a with/without comparison, not a single-variant evaluation) — never reference `{input:<id>}` in its prompt, and its `setup` removes the staged copy before the harness runs:

```yaml
setup:
  - executable: rm
    arguments: ["-rf", "{input:skill}"]
```

**When the target edits files, not just reads them** — `given.files`/`given.directories` stage into a read-only area outside the harness's actual working tree; the harness only ever sees them through `{input:<id>}`, and cannot write back into them. If the case needs a writable starting tree (e.g. framework files a principle-authoring skill edits in place), add a `setup:` step that copies the staged read-only input into the workspace at its conventional path before the harness runs:

```yaml
setup:
  - executable: sh
    arguments: ["-c", "mkdir -p {workspace}/framework && cp {input:framework-src}/*.md {workspace}/framework/"]
```

The prompt then tells the harness the writable copy is at its normal relative path (`framework/*.md`), distinct from the read-only skill/agent guidance it reaches via `{input:<id>}`.

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

**Input:** User invokes `smaqit.bench-scaffold` in a project where `.github/skills/my-skill/SKILL.md` exists with no corresponding `.smaqit/bench/skills/my-skill/bench.yaml`.

**Output:** The skill detects `.github/skills/` as the root, lists `my-skill` as an uncovered target, reads `my-skill/SKILL.md`, drafts `.smaqit/bench/skills/my-skill/bench.yaml` staging `my-skill`'s directory via `given.directories` with an explicit `{input:skill}` reference in the prompt and the reusable Codex block, runs `bench validate` (passes), and asks whether to run a live trial via `smaqit.bench-run`.

## Gotchas

- `given.files`/`given.directories` stage into a read-only directory beside the workspace, never at the artifact's literal project-relative path — see Step 4. Getting this wrong produces a manifest that structurally validates but fails every live run.
- The reusable Codex process block in `.smaqit/bench/README.md` is the canonical, tested source — copy it, don't reconstruct it from memory.

## Completion

- [ ] Skill/agent root detected (`.github/skills|agents/` or root `skills|agents/`, or resolved by asking)
- [ ] Target selected; confirmed it has no existing manifest, or the user chose to extend one
- [ ] Target's `SKILL.md`/`.agent.md` read in full before drafting
- [ ] Manifest stages the target artifact via `given.files`/`given.directories`, with the prompt referencing `{input:<id>}` explicitly
- [ ] Reusable Codex process block copied verbatim, not hand-rolled
- [ ] `bench validate` passes on the new manifest
- [ ] Live trial offered; if accepted, delegated to `smaqit.bench-run`

## Failure Handling

| Situation | Action |
|-----------|--------|
| Neither `.github/skills\|agents/` nor root `skills\|agents/` resolves unambiguously | Ask the user where the project's skills/agents live |
| Chosen target already has a manifest | Ask whether to add a case to it instead of creating a new manifest |
| A manifest, `prompts/*.md` file, or other output artifact already exists at the write path | Confirm with the user before overwriting |
| Structural validation fails | Report diagnostics; fix the manifest before offering a live trial |
| User declines a live trial | Report the manifest as scaffolded-but-unverified; stop cleanly |
| Invoking `smaqit.bench-run` for the live trial fails outright | Report the failure with context; do not silently retry |
| Live trial (via `smaqit.bench-run`) surfaces a genuine `src/bench` limitation | Report it precisely with the reproducing case; recommend a follow-up task rather than patching the engine inline |
| Target's `SKILL.md`/`.agent.md` is missing or unreadable | Stop; report the path that could not be read |
