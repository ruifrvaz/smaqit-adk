---
status: PR Open
created: "2026-08-18"
mode: Assisted
started: "2026-08-19"
pr: 1
---

# Registry-Driven Principle Propagation

## Description

Build a registry (Terraform-statefile style) of ADK-owned agents/skills, plus a framework
**extension** model, so principle and template changes automatically cascade through L0 → L1 → L2
to every registered dependent — replacing the manual `smaqit.compile.*` chain proposed in
superseded Task 019. smaqit-adk's own `framework/*.md` remains the single, domain-agnostic **base**
principle set. Any consuming project may additionally maintain its own **extension** at
`.smaqit/framework/*.md` in that project's own repo — product-specific principles that extend the
base rather than duplicating it. `smaqit.new-principle` becomes scope-aware (base when curating
inside smaqit-adk itself, extension when curating inside a consuming project) and enforces
non-duplication against the base when authoring extension principles. Registered agents/skills
inherit cascades from both the shared base and, where applicable, their own project's extension.
Also closes Task 018's two remaining gaps (L0 definition-file input pattern, missing L1 skill entry
points) as a prerequisite step, since the cascade cannot run L0→L1→L2 end-to-end without them.

## Issue Triage Context

**Mode:** Skip
**Technologies:** Go (installer/main.go), Bash (install.sh), Markdown/SKILL.md, YAML (`.smaqit/definitions/`), JSON (proposed registry format)
**Platforms/Environments:** Claude Code (`~/.claude/agents/`, `~/.claude/skills/`), Codex CLI (`~/.codex/agents/`, `~/.agents/skills/`); GitHub Copilot is AGENTS.md-only, not an authored compile target (per Task 027)
**Features/Integrations:** `smaqit.L0`/`L1`/`L2` subagent chain, `smaqit.create-agent`/`smaqit.create-skill`, `smaqit.new-principle`, `installer/main.go` CLI (`bench`, `uninstall`, `version`, proposed `sync`)
**Versions/Constraints:** Current adk version 2.0.1; must not break existing `--install-global` / `uninstall` flows, which currently re-derive "what's installed" from the binary's own embedded file list with zero persisted state

## Design Decisions

- **Supersession, not extension (of the task, not to be confused with framework extensions below):** This task supersedes Task 019 rather than rewriting it in place or shipping it first — "automatic propagation" obsoletes 019's manual-chain premise entirely.
- **Framework extension model, not a merge (course correction, 2026-08-19):** The originally-planned "two independent, disjoint registries" design is revised. There is exactly one base principle set — smaqit-adk's own `framework/*.md` — never duplicated elsewhere. A consuming project's product-domain principles are not merged into smaqit-adk's repo (that would recontaminate the domain-agnostic core, reversing Task 001 and Task 054) and are not a wholly separate parallel set either. They live in that project's own repo at `.smaqit/framework/*.md`, explicitly modeled as an **extension** of the base: additive, not restating what the base already covers. `smaqit.new-principle` must detect which scope it is operating in (base, when run inside smaqit-adk itself; extension, when run inside a consuming project) and route the write accordingly.
- **Deduplication is enforced going forward, not a one-time cleanup:** When `smaqit.new-principle` authors or refines an extension principle, it must check the new content against the base framework's existing principles and flag/reject anything that merely restates a base principle rather than adding genuinely project-specific content. The corollary one-time migration — moving an existing project's own root-level `framework/*.md` into `.smaqit/framework/*.md` and deduplicating it one-by-one against the base — is a separate, project-side follow-up (see Notes), not a file this task itself changes.
- **Registry pairs with the extension:** The per-project registry (`.smaqit/registry.json` in any consuming project) is paired with that same project's `.smaqit/framework/*.md` extension (where one exists) and tracks that project's own `create-agent`/`create-skill`-generated agents/skills. The ADK-global registry (`~/.agents/smaqit-adk/registry.json`) tracks `smaqit.L0/L1/L2` + the 5 shipped skills against the base framework. Cascade direction follows inheritance: an edit to the base propagates to every registered project's dependents (the base is inherited everywhere); an edit to one project's own extension propagates only to that project's own registered dependents.
- **Registry scope excludes hand-authored product agents:** Only resources created via `smaqit.create-agent`/`smaqit.create-skill` (plus the ADK's own L0/L1/L2 and shipped skills) are registered. smaqit product's hand-authored agents (`agents/business.md` etc. in the `smaqit` repo, compiled via its own `scripts/generate-agents.py`) are explicitly out of scope — a separate, unrelated pipeline this task does not touch.
- **Propagation trigger, cascade owned by the skill layer (course correction, 2026-08-19):** Live cascade is owned by `smaqit.new-principle`/`new-template`/`new-rules`, not by agent-to-agent chaining. Discovery confirmed no 2-hop subagent chain (agent invokes agent invokes agent) exists anywhere in this repo or the sibling `smaqit` product, and only `smaqit.L2` currently holds the `Task` tool grant. Instead, the skill reads the registry itself after its Level agent returns and loops single-hop `Task` calls to L1 then L2 for each registered dependent, in the same session — mirroring the proven `smaqit.development` → Business/Functional/Stack orchestration pattern. No `Task` tool grant is added to `agents/smaqit.L0.md`/`smaqit.L1.md` or their `.smaqit/definitions/agents/*.frontmatter.yaml` metadata. Plus a `smaqit-adk sync` CLI subcommand for on-demand/install-time reconciliation of drift.
- **Registry I/O via a shared Go package (course correction, 2026-08-19):** `installer/registry.go`, modeled directly on `src/bench/plan.go`'s `writeJSONAtomic`/`ReadPlan` (atomic write, `DisallowUnknownFields`, exact `SchemaVersion` check) and `src/bench/hash.go`'s `digestFile`. Exposed via hidden `smaqit-adk registry get`/`registry put` subcommands (undocumented, alongside the existing `--install-global`) that agents shell out to via Bash (already granted); `smaqit-adk sync` uses the same package directly. Avoids duplicate hash logic between agent-side shell `sha256sum` and Go `sha256` that could silently diverge and produce false drift results.
- **Uniform registration in L2, not per-caller (course correction, 2026-08-19):** `smaqit.L2` writes its own registry entry via `smaqit-adk registry put` immediately after every compile, regardless of which skill invoked it (`create-agent`, `create-skill`, or a cascade-triggered recompile). Supersedes the original plan to duplicate registration logic inside `smaqit.create-agent`/`smaqit.create-skill` individually — one write path, not two.
- **Scope-detection heuristic (assumption, 2026-08-19):** `smaqit.L0` detects base vs. extension scope by checking whether the invoking repo has smaqit-adk's own source layout (`agents/smaqit.L0.md` + `installer/main.go` present at repo root) — if absent, the invoking context is a consuming project and the write targets `.smaqit/framework/*.md`. Open to revision during implementation if a more robust signal is found.
- **Registry format (proposed, open to revision during implementation):** JSON, keyed by target file path, recording source definition file, source template/rules dependency, and a compiled-content hash for drift detection.
- **Task 018 folded in:** Its two remaining ACs (L0 definition-file pattern, `smaqit.new-template`/`smaqit.new-rules` skills) become this task's Implementation Step 3 rather than staying a separate blocking task.

## Implementation Steps

1. **Registry foundation** — Implement `installer/registry.go` (shared Go package: registry struct, atomic JSON read/write, sha256 hashing) and hidden `smaqit-adk registry get`/`registry put` subcommands in `installer/main.go`, modeled on `src/bench/plan.go`/`hash.go`. Each entry records: target file path(s) per platform, source definition file, source template/rules file(s) it was compiled from, and a last-compiled hash/timestamp.
2. **Uniform registration in L2** — Update `agents/smaqit.L2.md` to shell out to `smaqit-adk registry put` immediately after every compile, regardless of caller — no changes needed to `smaqit.create-agent`/`smaqit.create-skill` themselves.
3. **Close Task 018's gaps** (depends on step 1):
   - Add a definition-file input pattern to `smaqit.L0` (`.smaqit/definitions/principles/[name].md`), matching L2's existing pattern.
   - Author `smaqit.new-template` and `smaqit.new-rules` skills (L1 entry points), currently missing entirely.
4. **Framework extension support** (depends on step 3) — Teach `smaqit.L0`/`smaqit.new-principle` to detect scope: base (invoked inside smaqit-adk's own repo/install, targeting `framework/*.md`) vs. extension (invoked inside a consuming project, targeting `.smaqit/framework/*.md`). When authoring or refining an extension principle, diff the proposed content against the base framework's existing principles and flag/reject duplication rather than silently accepting a restatement.
5. **Live cascade owned by the skill layer** (depends on steps 1–4) — Wire `smaqit.new-principle`, `smaqit.new-template`, and `smaqit.new-rules` (the skills themselves, not L0/L1) so that after their Level agent confirms a successful edit, the skill queries the registry for every downstream dependent and loops single-hop `Task` calls to L1 then L2 for each one, in the same session — mirroring `smaqit.development`'s orchestration pattern. A base edit cascades to every registered project's dependents; an extension edit cascades only to that project's own registered dependents.
6. **Reconciliation CLI** (depends on step 1) — Add a `smaqit-adk sync` subcommand to `installer/main.go` that walks the relevant registry/registries, recompiles anything whose source hash changed since last compile, and reports drift (registered agents/skills hand-edited outside the chain). Runs standalone or as part of `--install-global`.
7. **Scope validation** (depends on step 5) — Verify the base/extension model end-to-end: a base-framework change cascades to every registered project; a project's own extension change cascades only to that project; and an attempt to add an extension principle duplicating base content is flagged/rejected rather than silently written.
8. **Cleanup and migration** (parallel with steps 6–7) — Mark Tasks 018 and 019 Abandoned in `PLANNING.md`, pointing to this task; retroactively register the ADK's existing L0/L1/L2 + 5 shipped skills into the global registry with a baseline hash (verify this does not fire false-positive drift against in-flight work, e.g. Task 031); update `docs/wiki/extending-smaqit.md`'s "Framework changes don't propagate" entry, which currently prescribes the exact manual steps this task eliminates, to also document the base/extension model.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [x] Registry file format implemented and documented, one per scope (ADK-global, per-project)
- [x] `smaqit.L2` auto-registers every compiled agent/skill in the registry immediately after compiling, regardless of caller
- [x] `smaqit.L0` accepts a definition-file input pattern (`.smaqit/definitions/principles/[name].md`)
- [x] `smaqit.new-template` and `smaqit.new-rules` skills exist and invoke `smaqit.L1`
- [x] `smaqit.new-principle` correctly detects base vs. extension scope and routes writes to `framework/*.md` or `.smaqit/framework/*.md` accordingly
- [x] `smaqit.new-principle` flags/rejects an extension principle that duplicates existing base content
- [x] Editing a principle automatically cascades to every registered dependent in the same session, correctly scoped
- [x] `smaqit-adk sync` subcommand recompiles drifted registered entries and reports conflicts rather than silently overwriting manual edits
- [x] A base-framework edit cascades to every registered project; an extension edit cascades only to its own project; verified to never cross
- [x] Tasks 018 and 019 marked Abandoned in `PLANNING.md`, pointing to this task
- [x] `make build` passes cleanly

## Findings

**Implementation approach:**
- Registry as a shared Go package (`installer/registry.go` + `registry_cli.go`): atomic JSON read/write and sha256 hashing modeled on `src/bench/plan.go`/`hash.go`, exposed via hidden `smaqit-adk registry get`/`put` subcommands agents shell out to via Bash
- `smaqit.L2` registers uniformly on every compile (Base/Spec/Impl Agent and Skill patterns all gained a final "Register" step) instead of duplicating registration logic per calling skill
- `smaqit.L0` gained scope detection (base vs. extension, via a repo-layout heuristic), a definition-file input pattern, and duplicate-content rejection for extension principles
- Live cascade (`smaqit.new-principle`/`new-template`/`new-rules`) owned entirely by the skill layer: each skill invokes its Level agent, then loops single-hop `Task` calls to L1 then L2 for every registered dependent the registry returns — no nested agent-to-agent chaining
- All new/changed root-shipped skills (`smaqit.new-template`, `smaqit.new-rules`, the `smaqit.new-principle` revision, and `smaqit.sync`) compiled via a real `smaqit.L2` subagent from `.smaqit/definitions/skills/*.md`, not hand-written, per this repo's established dogfooding convention
- Live-verified the full L0 → L1 → registry → L2 chain end-to-end via a new Bench scenario (`.smaqit/bench/scenarios/registry-cascade/`) run against the authenticated Codex CLI: confirmed by direct trace-log inspection that a principle landed, `skill.rules.md` was updated, the registered dependent (`smaqit.sync`) was genuinely recompiled with the new content, and the registry hash updated

**Decisions made:**
- Cascade orchestration lives in the skill layer, not granted to L0/L1 via a `Task` tool — confirmed by direct source inspection that no 2-hop subagent chain exists anywhere in this repo or the sibling `smaqit` product; `smaqit.development`'s Business/Functional/Stack orchestration is always exactly one hop
- Registry I/O centralized in Go rather than per-agent shell/JSON manipulation, avoiding hash divergence between shell `sha256sum` and Go `sha256`
- `smaqit-adk sync` (the CLI) stays report-only by design — a plain binary has no path to invoke an LLM for the semantic L1/L2 recompile a drifted entry needs. Added `smaqit.sync` (a skill, not originally planned in this task's file) as the actual interactive resolution path: it shells out to the same registry CLI, then uses its own `Task` access to recompile from source or re-register a kept hand edit, on explicit per-entry user choice — this addition came directly from the user questioning why the CLI alone was expected to "recompile," and closes what would otherwise have been a real gap between the AC's wording and what a plain Go binary can do
- `smaqit.new-principle`'s cascade sequencing is L0 → L1 (judgment: does this principle change actually require a template/rules update?) → registry lookup → L2, not L0 → registry lookup directly — the registry only has L1-template→L2-output edges, never an L0-principle→L2 edge, so skipping the L1 judgment step would make the registry lookup silently return zero dependents every time

**Blockers encountered:**
- A manual verification pass ran `--install-global` against the WIP dev binary without sandboxing `HOME`, briefly overwriting the real machine-wide install (`~/.claude/agents`, `~/.codex/agents`, `~/.agents/skills`, `~/.claude/skills`, `~/.agents/smaqit-adk`). Caught immediately, restored from the real production binary, and removed the resulting orphaned WIP-only skill directories the restore doesn't auto-clean
- The live Bench trial surfaced two genuine `src/bench` engine bugs, not implementation bugs in this task: the post-harness "freeze" step crashes on a non-regular file (Codex CLI's own IPC socket, created as a side effect of running `codex exec` with a workspace-local `HOME`) instead of handling it gracefully, and — more seriously — a resulting workspace-cleanup failure was silently swallowed with no visible warning, leaking a workspace containing copied Codex credentials to disk. Found via manual inspection after the run reported failure, cleaned up by hand, filed as Task 035 rather than fixed inline

**Follow-up identified:**
- Task 035 (filed): fix Bench's freeze-step crash on non-regular files and the silently-swallowed cleanup-error bug
- The pre-existing dogfood Bench case at `.smaqit/bench/skills/smaqit.new-principle/` only exercises the skill's Step 1 (L0) in isolation and doesn't cover the new cascade steps this task added — worth folding into or superseding with the new `scenarios/registry-cascade/` manifest in a future task
- Two ACs rest on implementation correctness and code-level reasoning rather than an independent live trial: extension-scope duplicate-principle rejection, and cross-project cascade isolation (structurally guaranteed by separate registry files, but not exercised via an actual second consuming project in a live run) — worth a dedicated live trial if stronger proof is wanted later
- `smaqit` product repo's own migration (relocating its 7 root `framework/*.md` files to `.smaqit/framework/*.md` and deduplicating against the base) remains a separate, smaqit-repo-side follow-up this task deliberately does not touch

## Files to Create / Modify

| File | Action |
|------|--------|
| `installer/registry.go` | Create — shared registry struct, atomic JSON read/write, hash helpers (models `src/bench/plan.go`/`hash.go`) |
| `agents/smaqit.L0.md` | Modify — add definition-file input pattern, report changed file on completion for the calling skill's cascade query (no `Task` grant added) |
| `agents/smaqit.L1.md` | Modify — report compiled output on completion (no `Task` grant added) |
| `agents/smaqit.L2.md` | Modify — shell out to `smaqit-adk registry put` after every compile (uniform registration) |
| `skills/smaqit.new-principle/SKILL.md` | Modify — own the live cascade: query registry, loop single-hop `Task` calls to L1 then L2 per dependent |
| `skills/smaqit.new-template/SKILL.md` | Create — new L1 entry point, same cascade-owning shape |
| `skills/smaqit.new-rules/SKILL.md` | Create — new L1 entry point, same cascade-owning shape |
| `installer/main.go` | Modify — hidden `registry get`/`put` subcommands, public `sync` subcommand |
| `docs/wiki/extending-smaqit.md` | Modify — replace manual-cascade troubleshooting entry |
| `.smaqit/tasks/PLANNING.md` | Modify — abandon 018 and 019, add 032 |
| `.smaqit/tasks/018_level_skills_completion.md` | Modify — mark Abandoned, point to 032 |
| `.smaqit/tasks/019_cross_level_compilation.md` | Modify — mark Abandoned, point to 032 |

## Notes

- Planned in a Claude Code session working in the sibling `smaqit` product repo, not persisted there — this task file is the durable record; the planning session is not recoverable.
- Origin: the user added a new product-domain principle ("No Grandfathering") to `smaqit`'s own `framework/SMAQIT.md` via `smaqit.new-principle` → `smaqit.L0`, then asked how to propagate it to smaqit's agents/skills. Investigation found no compiler wired to smaqit's own product agents at all (they're hand-authored, compiled via `scripts/generate-agents.py`, unrelated to L0/L1/L2) — confirming this task's locked scope decision that hand-authored product agents stay out of the registry. That specific "No Grandfathering" principle **remains unpropagated to smaqit's own agents/skills** after this task ships; propagating it is a separate, smaqit-product-side concern this task does not solve.
- Landmine surfaced during Discovery, now resolved by design: Task 018's notes state `smaqit.new-principle` "targets the ADK framework only — product-domain principles belong in the product extension, not the ADK." Implementation Step 4 (Framework extension support) is that product-extension mechanism — it did not exist before this task.
- **Course correction (2026-08-19):** the original plan called for two fully independent, disjoint registries/principle-sets (ADK-global vs. per-project, never relating to each other). Revised to a base/extension inheritance model instead — see the Design Decisions above. Rejected outright: physically merging a consuming project's product-specific content into smaqit-adk's own `framework/*.md` (would recontaminate the domain-agnostic core smaqit-adk was deliberately extracted to preserve — see Task 001 "Clean L2 Agent Contamination" and Task 054 "Level Agent Extraction Cleanup").
- **Reference migration, not part of this task's file changes:** the `smaqit` product repo (the ADK's own reference consumer) currently keeps 7 `framework/*.md` files at its repo root — confirmed, by direct diff and a targeted grep for ADK/Level-agent vocabulary, to be substantively independent product content, not a stale copy of smaqit-adk's 5 base files. Once this task ships, smaqit's own migration is: relocate `framework/*.md` → `.smaqit/framework/*.md`, then run the one-by-one dedup pass Design Decisions describes against smaqit-adk's base. That migration happens in the `smaqit` repo and should be tracked as its own task there when picked up — this task only needs to build the mechanism that makes it possible.
- Supersedes Task 019 (Cross-Level Compilation) — manual chain premise obsoleted by automatic cascade.
- Folds in Task 018's (Level Skills Completion) remaining two ACs as this task's Implementation Step 3.
- **Course correction (2026-08-19, task-plan session):** Discovery (two Explore agents) confirmed no 2-hop agent-to-agent subagent chain exists anywhere in this repo or the sibling `smaqit` product, and only `smaqit.L2` holds the `Task` tool grant — `smaqit.development`'s orchestration of Business/Functional/Stack is always exactly one hop from an orchestrating context, never nested. Revised cascade ownership from agent-to-agent chaining to the skill layer, registry I/O from per-agent direct file manipulation to a shared Go package/hidden CLI subcommand, and registration from per-skill duplication to a single uniform write inside `smaqit.L2` — see Design Decisions above.
