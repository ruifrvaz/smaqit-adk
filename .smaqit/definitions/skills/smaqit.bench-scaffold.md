---
name: smaqit.bench-scaffold
tier: advanced
---

# Skill Definition: smaqit.bench-scaffold

## Identity

- **Name:** `smaqit.bench-scaffold`
- **Description (draft, L2 may refine):** Authors a new `.smaqit/bench/` manifest for a skill or agent that doesn't have one yet — detects the project's skill/agent root generically, drafts a self-contained single-shot Codex prompt that stages the target artifact as a declared Bench input, structurally validates the result, and delegates any optional live trial to `smaqit.bench-run`.
- **Version:** 0.1.0

## Steps (with fragility)

1. **Detect the skill/agent root** (medium fragility). Check `.github/skills/` + `.github/agents/` first — the standard `smaqit-adk lite`/`advanced` install location in a consuming project. Fall back to root `skills/`/`agents/` (this repo's own dev-repo convention) only when `.github/skills|agents/` doesn't exist. If neither resolves unambiguously, ask the user where the project's skills/agents live.
2. **Select target** (medium fragility). List targets under the detected root that have no `.smaqit/bench/{skills,agents}/<id>/bench.yaml`. If the chosen target already has a manifest, ask whether to add a case to it instead of creating a new one.
3. **Understand the target** (low fragility — judgment call, not mechanized). Read the target's `SKILL.md` or `.agent.md` in full — its purpose, steps, and any subagent it invokes.
4. **Draft the manifest** (high fragility — exact mechanics matter and getting them wrong produces a structurally-invalid or silently-broken manifest):
   - Self-contained, single-shot case prompts. Bench's `process` adapter is non-interactive with no multi-turn `ask_user` relay — pre-supply any answer a human would give across turns and tell the harness not to wait for further input.
   - Stage the target artifact as a case-level `given.files` (single-file agent) or `given.directories` (skill directory) input asset, e.g. `directories: [{id: skill, source: ../../../../skills/<target-id>}]` (path relative to the new `bench.yaml`).
   - Bench copies declared inputs into a **read-only** `.smaqit-bench-input/` directory beside the harness's actual working tree — never at the artifact's real project-relative path inside the workspace itself. The harness only reaches staged content through the resolved `{input:<id>}` absolute-path placeholder. The prompt MUST reference this placeholder explicitly, e.g.: "The project's ADK skill-authoring skill is staged at `{input:skill}/SKILL.md` — read it first and follow it exactly." Never phrase this as "if staged at `.github/skills/<id>/SKILL.md`" — that phrasing only works for the repo-specific `smaqit-adk-dev lite {workspace}` Setup trick used by this repo's own three original dogfood manifests, which install directly into the workspace root instead of a declared input asset.
   - When the target *edits* files rather than just reading them (e.g. a principle-authoring skill editing `framework/*.md`), add a `setup:` step that copies the staged read-only input into the workspace at its conventional writable path before the harness runs, e.g. `setup: [{executable: sh, arguments: ["-c", "mkdir -p {workspace}/framework && cp {input:framework-src}/*.md {workspace}/framework/"]}]`.
   - For a with/without-artifact comparison, the without-artifact variant never references `{input:<id>}` in its prompt, and its `setup` removes the staged copy: `setup: [{executable: rm, arguments: ["-rf", "{input:skill}"]}]`.
   - Prefer `command`-type expectations over bare `text`-type for anything that might not exist — a `text` expectation crashes rather than failing gracefully against a missing file.
   - Copy the reusable Codex process block verbatim from `.smaqit/bench/README.md` (pins `--sandbox danger-full-access`, `--skip-git-repo-check`, an explicit `timeoutSeconds`). Never hand-roll a variant of it.
5. **Structural validation** (high fragility). Run `smaqit-adk bench validate .smaqit/bench/{skills,agents}/<id>/bench.yaml`. Fix diagnostics before proceeding. Do not offer a live trial against a manifest that fails structural validation.
6. **Optional live trial** (medium fragility). Ask whether to run a live trial now. If yes, invoke `smaqit.bench-run` scoped to the new manifest — reuse its preflight/confirm/execute/diagnose logic rather than duplicating it. If a live trial surfaces a genuine `src/bench` engine limitation (not a known gotcha), report it precisely and recommend a follow-up task; never patch the engine inline from this skill.
7. **Report** (low fragility). Report the new manifest path (and any `prompts/*.md` files), its structural validation status, and the live-trial result if one was run.

## Output

- `.smaqit/bench/{skills,agents}/<id>/bench.yaml` (+ any `prompts/*.md` it references)
- Structural validation status
- Live-trial result, if a trial was run
- No subagent invoked for authoring; a live trial delegates to `smaqit.bench-run`

## Scope

Authors manifests only. Does not execute a full suite run or reimplement `smaqit.bench-run`'s preflight/confirmation/diagnosis logic — any live trial delegates to it. Does not modify `src/bench` engine code to work around a discovered limitation; stops and reports instead.

## Completion

- [ ] Skill/agent root detected (`.github/skills|agents/` or root `skills|agents/`, or resolved by asking)
- [ ] Target selected; confirmed it has no existing manifest, or the user chose to extend one
- [ ] Target's `SKILL.md`/`.agent.md` read in full before drafting
- [ ] Manifest stages the target artifact via `given.files`/`given.directories`, with the prompt referencing `{input:<id>}` explicitly
- [ ] Reusable Codex process block copied verbatim, not hand-rolled
- [ ] `bench validate` passes on the new manifest
- [ ] Live trial offered; if accepted, delegated to `smaqit.bench-run`

## Failure Scenarios

| Situation | Response |
|-----------|----------|
| Neither `.github/skills\|agents/` nor root `skills\|agents/` resolves unambiguously | Ask the user where the project's skills/agents live |
| Chosen target already has a manifest | Ask whether to add a case to it instead of creating a new manifest |
| Structural validation fails | Report diagnostics; fix the manifest before offering a live trial |
| User declines a live trial | Report the manifest as scaffolded-but-unverified; stop cleanly |
| Live trial (via `smaqit.bench-run`) surfaces a genuine `src/bench` limitation | Report it precisely with the reproducing case; recommend a follow-up task |
| Target's `SKILL.md`/`.agent.md` is missing or unreadable | Stop; report the path that could not be read |

## Gotchas

- `given.files`/`given.directories` stage into a read-only directory beside the workspace, never at the artifact's literal project-relative path — see Step 4. This is the single most consequential mechanic in this skill; getting it wrong produces a manifest that structurally validates but fails every live run.
- The reusable Codex process block in `.smaqit/bench/README.md` is the canonical, tested source — copy it, don't reconstruct it from memory.

## Examples

**Input:** User invokes `smaqit.bench-scaffold` in a project where `.github/skills/my-skill/SKILL.md` exists with no corresponding `.smaqit/bench/skills/my-skill/bench.yaml`.
**Output:** Skill detects `.github/skills/` as the root, lists `my-skill` as an uncovered target, reads `my-skill/SKILL.md`, drafts `.smaqit/bench/skills/my-skill/bench.yaml` staging `my-skill`'s directory via `given.directories` with an explicit `{input:skill}` reference in the prompt and the reusable Codex block, runs `bench validate` (passes), and asks whether to run a live trial via `smaqit.bench-run`.

## Compatibility

Requires no additional runtime dependencies beyond the `smaqit-adk` binary for authoring/validation; an optional live trial requires `codex` on `PATH`, delegated to `smaqit.bench-run`.
