# smaqit-adk's own HarnessBench suite

This directory is smaqit-adk **dogfooding** its own `smaqit-adk bench` engine against its own skills and agents. It is not ADK-shipped product source — the engine itself lives in `src/bench`/`src/benchcli` and ships in the binary. This directory is this repo's local data, in the same sense `.smaqit/tasks/` or `.smaqit/compendium.md` are: state a project keeps about itself, not something `smaqit-adk lite`/`advanced` writes into a consumer's project.

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

Skill suites live under `skills/<skill-id>/`; agent suites (flat files, no natural directory of their own) live under `agents/<agent-id>/`, matching the ID used in `skills/<skill-id>/SKILL.md` or `agents/<agent-id>.agent.md` at the repo root. One `bench.yaml` per target unless a target has genuinely distinct scenarios worth splitting into multiple manifests.

## Case naming

Each `Case.ID` in a manifest names the scenario, not the target (the target is implied by the manifest's directory). Prefer a short, present-tense description of what's being exercised, e.g. `gather-and-compile`, `ambiguity-flagging`, `validator-gating`. `MIGRATION.md` records which legacy `NNN_scenario_name.json` file each case descends from, so history stays traceable without carrying the old numeric prefixes forward.

## With-artifact / without-artifact comparison

Per Task 026's design decision, this is **one Case, two Variants**, not two Cases. Stage the target skill/agent as a case-level input (visible to both variants), then use the *without-artifact* variant's `setup` to remove it before the harness runs — never referencing it in that variant's prompt/arguments either. This lets Bench's native `compare.go` (grouped by `VariantID`) produce the win/tie/inconclusive comparison directly, with no custom aggregation code:

```yaml
cases:
  - id: gather-and-compile
    given:
      prompt:
        file: prompt.md
      directories:
        - id: skill
          source: ../../../../skills/smaqit.create-agent   # relative to this bench.yaml
    expect:
      - id: definition-file-written
        type: file
        actual: "file:.smaqit/definitions/agents/qa-helper.md"
        operator: exists

variants:
  - id: with-artifact
    adapter: process
    process: *codex-exec        # see the reusable block below
  - id: without-artifact
    adapter: process
    process: *codex-exec
    setup:
      - executable: rm
        arguments: ["-rf", "{input:skill}"]
    intendedDifferences:
      - Does not stage or reference the target skill; establishes the no-artifact baseline.
```

(YAML anchors like `&codex-exec`/`*codex-exec` only work *within* a single manifest file — Bench's loader rejects multi-document files and has no cross-file include mechanism. Anchor within one manifest if it has multiple process variants; across manifests, copy the block below verbatim.)

## Reusable Codex process-variant block

There's no templating in Bench for this — every manifest that drives `codex exec` copies this block and adjusts only the prompt/arguments specific to its case. Two things matter for correctness, both surfaced by issue triage on this task (`openai/codex` open issues, see the task file's Known Issues Triage):

- **Pin a non-interactive sandbox/approval mode explicitly.** Bench's `process` adapter has no PTY and no approval-relay — if `codex exec` waits on an approval prompt, the run hangs until Bench's own timeout kills it (`openai/codex#36570`: `approvals_reviewer` can silently defeat an explicit `--sandbox` flag). Don't rely on defaults.
- **Pass `--skip-git-repo-check`.** Confirmed live: Codex refuses to run at all ("Not inside a trusted directory") outside a Git repository, and Bench's disposable workspace is a plain temp directory, never a repo.
- **Set an explicit `timeoutSeconds`** on `execution` rather than trusting the 300s default — `codex exec` has a known stall report (`openai/codex#28476`). A bounded timeout turns a stall into a clean `timedOut` status instead of an indefinite hang.
- **Point the harness at the staged artifact explicitly.** Staging a skill/agent file into the workspace is not enough on its own — confirmed live, Codex's own repo-exploration habits (e.g. `rg --files`, which hides dotfiles by default) can miss content under `.github/`/`.smaqit/`, and generic wording like "create a skill" can collide with Codex's own built-in skill-authoring feature. Prompts should say e.g. "if this project has an ADK skill-authoring skill staged at `.github/skills/<id>/SKILL.md`, read it first and follow it exactly" — phrased conditionally so the same prompt is honest for the without-artifact variant too, where nothing is staged.

## Command graders and Snap-packaged toolchains

A `command`-type expectation or grader (including `Setup`) runs with **no environment at all** unless the manifest sets `command.environment.inherit`/`.set` — not even `PATH` or `HOME`. Most Unix tools (`sh`, `grep`, `test`, `rm`) don't need one. `go run` does, and on a Snap-packaged Go toolchain (`/snap/bin/go -> /usr/bin/snap`, common on Ubuntu) it fails even with a correct environment: Bench's process-group isolation (needed for reliable timeout/kill handling) collides with Snap's own confinement locking (`error: race condition detected, snap-run can only retry once`), confirmed live. Where a grader needs a Go tool, prefer pre-compiling it once at build time (see `installer/Makefile`'s `build` target building `dist/validate-skill`) and pointing the command grader at the compiled binary — it needs no environment and never touches Snap.

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
    - "{taskFile}"
  inputMode: argument
```

## Hidden oracles and graders

Oracle values (`golden`/`valueFile` targets, grader `assets`) live outside any staged input path — Bench enforces this at manifest-validation time (an oracle file cannot resolve inside `given.fixture`/`specs`/`files`/`directories`/`images`). Keep oracle files as siblings of the `bench.yaml` that references them, not inside a fixture directory that gets copied into the harness's workspace.

## Run output

`bench run` (and the suite-level equivalent) write experiment output under each manifest's `output.directory`. Point every dogfood manifest at a path under `.smaqit/bench/runs/<skill-or-agent-id>/` so all evidence lands in one gitignored place — see `.gitignore` in this directory. This supersedes `tests/evals/runs/`'s role; historical snapshots there are not migrated forward (they're an artifact of the deleted Copilot-SDK runner, not restaged evidence).
