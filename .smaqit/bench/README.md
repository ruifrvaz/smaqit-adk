# smaqit-adk's own HarnessBench suite

This directory is smaqit-adk **dogfooding** its own `smaqit-adk bench` engine against its own skills and agents. It is not ADK-shipped product source — the engine itself lives in `src/bench`/`src/benchcli` and ships in the binary. This directory is this repo's local data, in the same sense `.smaqit/tasks/` or `.smaqit/compendium.md` are: state a project keeps about itself, not something the global installer writes into a consumer's project.

See [`MIGRATION.md`](MIGRATION.md) for how the legacy Copilot-SDK eval suite's scenarios map onto the manifests here.

## Layout

```
.smaqit/bench/
├── skills/
│   └── <skill-id>/
│       └── bench.yaml       # + any fixture/oracle files it references
├── agents/
│   └── <agent-id>/
│       └── bench.yaml
├── runs/                    # generated experiment output — gitignored, see .gitignore
├── MIGRATION.md
└── README.md
```

Skill suites live under `skills/<skill-id>/`; agent suites (flat files, no natural directory of their own) live under `agents/<agent-id>/`, matching the ID used in `skills/<skill-id>/SKILL.md` or `agents/<agent-id>.md` at the repo root. One `bench.yaml` per target unless a target has genuinely distinct scenarios worth splitting into multiple manifests.

## Bench vocabulary

Bench uses **Case** for an evaluation scenario, **Prompt** for the author-supplied `given.prompt`, and **Case brief** for the rendered prompt plus shared-input and variant-treatment paths delivered to a harness. A smaqit **Task** remains a tracked work item and is not Bench terminology. Process manifests use `{brief}` for the rendered Case brief and `{briefFile}` for its read-only `brief.md` file.

## Case naming

Each `Case.ID` in a manifest names the scenario, not the target (the target is implied by the manifest's directory). Prefer a short, present-tense description of what's being exercised, e.g. `gather-and-compile`, `ambiguity-flagging`, `validator-gating`. `MIGRATION.md` records which legacy `NNN_scenario_name.json` file each case descends from, so history stays traceable without carrying the old numeric prefixes forward.

## With-artifact / without-artifact comparison

This is **one Case, two Variants**, not two Cases. Put the target skill/agent and its supporting sources in the with-artifact variant's `treatment`; leave the baseline treatment empty. Bench stages treatments read-only after the common writable fixture and Case preparation are complete. The Case brief renders each variant's actual treatment table, so no removal command or prompt mutation is needed:

```yaml
cases:
  - id: gather-and-compile
    given:
      prompt:
        file: prompt.md
    expect:
      - id: definition-file-written
        type: file
        actual: "file:.smaqit/definitions/agents/qa-helper.md"
        operator: exists

variants:
  - id: with-artifact
    adapter: process
    treatment:
      - id: skill
        source: ../../../../skills/smaqit.create-agent
    process: *codex-exec        # see the reusable block below
    intendedDifferences:
      - Exposes the target skill as the variant treatment.
  - id: without-artifact
    adapter: process
    process: *codex-exec
    intendedDifferences:
      - Uses an explicitly empty treatment set.
```

(YAML anchors like `&codex-exec`/`*codex-exec` only work *within* a single manifest file — Bench's loader rejects multi-document files and has no cross-file include mechanism. Anchor within one manifest if it has multiple process variants; across manifests, copy the block below verbatim.)

## Reusable Codex process-variant block

There's no templating in Bench for this — every manifest that drives `codex exec` copies this block and adjusts only the prompt/arguments specific to its case. Four things matter for correctness; the Codex execution caveats were surfaced by issue triage on this task (`openai/codex` open issues, see the task file's Known Issues Triage):

- **Pin a non-interactive sandbox/approval mode explicitly.** Bench's `process` adapter has no PTY and no approval-relay — if `codex exec` waits on an approval prompt, the run hangs until Bench's own timeout kills it (`openai/codex#36570`: `approvals_reviewer` can silently defeat an explicit `--sandbox` flag). Don't rely on defaults.
- **Pass `--skip-git-repo-check`.** Confirmed live: Codex refuses to run at all ("Not inside a trusted directory") outside a Git repository, and Bench's disposable workspace is a plain temp directory, never a repo.
- **Set an explicit `timeoutSeconds`** on `execution` rather than trusting the 300s default — `codex exec` has a known stall report (`openai/codex#28476`). A bounded timeout turns a stall into a clean `timedOut` status instead of an indefinite hang.
- **Point the harness at the staged artifact explicitly.** The Case brief's `# Variant treatment artifacts` table gives each available treatment ID and resolved path. Prompts should say, for example, "if the treatment table lists `skill`, read its SKILL.md first and follow it exactly." The baseline brief explicitly says `None`, so the same prompt remains honest without filesystem mutation.

## Case data planes

Use each plane for one purpose:

- `fixture` is a writable starting project tree shared by every variant; `destination` places it at a safe workspace-relative path.
- Case-level `prepare` performs common deterministic preparation before the baseline snapshot. It can use only `{workspace}` and `{caseId}`.
- `given` inputs are shared, read-only resources available to every variant.
- Variant `treatment` artifacts are read-only resources available only to that variant. They require `intendedDifferences` and are the correct way to model with/without comparisons.

The complete `.smaqit-bench-input/` sidecar is excluded from repository metrics and frozen submissions. Never use preparation to encode a treatment.

## Command graders and Snap-packaged toolchains

A `command`-type expectation, grader, or Case preparation command runs with **no environment at all** unless the manifest sets `command.environment.inherit`/`.set` — not even `PATH` or `HOME`. Most Unix tools (`sh`, `grep`, `test`) don't need one. `go run` does, and on a Snap-packaged Go toolchain (`/snap/bin/go -> /usr/bin/snap`, common on Ubuntu) it fails even with a correct environment: Bench's process-group isolation (needed for reliable timeout/kill handling) collides with Snap's own confinement locking (`error: race condition detected, snap-run can only retry once`), confirmed live. Where a grader needs a Go tool, prefer pre-compiling it once at build time (see `installer/Makefile`'s `build` target building `dist/validate-skill`) and pointing the command grader at the compiled binary — it needs no environment and never touches Snap.

```yaml
process:
  executable: codex
  arguments:
    - exec
    - --sandbox
    - danger-full-access      # non-interactive; confirmed against codex-cli 0.147.0
    - --skip-git-repo-check   # required: Bench's disposable workspace is a plain temp dir, never a git repo, and codex refuses to run outside one without this
    - --cd
    - "{workspace}"
    - "{briefFile}"
  inputMode: argument
```

## Hidden oracles and graders

Oracle values (`golden`/`valueFile` targets, grader `assets`) live outside any staged input path — Bench enforces this at manifest-validation time (an oracle file cannot resolve inside `given.fixture`/`specs`/`files`/`directories`/`images`). Keep oracle files as siblings of the `bench.yaml` that references them, not inside a fixture directory that gets copied into the harness's workspace.

## Run output

`bench run` (and the suite-level equivalent) write experiment output under each manifest's `output.directory`. Point every dogfood manifest at a path under `.smaqit/bench/runs/<skill-or-agent-id>/` so all evidence lands in one gitignored place — see `.gitignore` in this directory. This supersedes `tests/evals/runs/`'s role; historical snapshots there are not migrated forward (they're an artifact of the deleted Copilot-SDK runner, not restaged evidence).
