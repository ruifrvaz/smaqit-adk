---
name: smaqit.sync
description: Reconciles registered agents and skills against their compiled output, detecting drift where a target was hand-edited outside the L0 → L1 → L2 compilation chain, and presenting each drifted entry for a recompile-from-source or keep-and-re-register resolution. The interactive counterpart to the non-interactive smaqit-adk sync CLI diagnostic, which reports drift but has no path to invoke an LLM for the semantic recompilation step.
metadata:
  version: "0.1.0"
compatibility: Requires the smaqit-adk binary on PATH for the registry subcommands, and sha256sum (or an equivalent SHA-256 utility) for hashing target files.
---

# Sync

## Steps

1. Enumerate registered entries:

   ```
   smaqit-adk registry get --scope base
   ```

   against the ADK-global registry. If a project-local `.smaqit/registry.json` exists, also run:

   ```
   smaqit-adk registry get --scope extension
   ```

   Treat a missing registry file as zero entries for that scope, not an error.

2. Detect drift per entry: for each entry, hash its primary target file (the `claude` path, or `codex` if `claude` is empty) with `sha256sum` and compare to the entry's `compiledHash`. A missing target file also counts as drift. Entries whose hash matches are current — do not report them further.

3. Present each drifted entry as a conflict, not an automatic action. For each drifted entry, show its name, scope, target path, and `lastCompiled` timestamp, then ask the user to choose:
   - **Recompile from source:** re-invoke `smaqit.L2` as a single-hop `Task` call from this skill's own turn, using the entry's recorded `sourceDefinition` and `sourceTemplates` paths to regenerate the target — this discards the hand edit.
   - **Keep the hand edit:** re-register the entry via `smaqit-adk registry put`, with the same name/target/source fields but the target's current hash, so it is no longer reported as drifted.

   Never pick one of these automatically. Never silently overwrite a hand edit.

4. Report: summarize total entries checked per scope, how many were current, how many were drifted, and the resolution chosen for each drifted entry.

## Output

- A per-scope drift report (checked / current / drifted counts)
- For each drifted entry, a recorded user resolution: recompiled via `smaqit.L2`, or kept and re-registered via `smaqit-adk registry put`
- No entry left drifted and unresolved without the user having been asked

## Scope

Reconciliation only. Does not create new registrations — that's `smaqit.L2`'s job on every compile — and does not modify the registry format. Does not run automatically: invoked explicitly by the user, or as a deliberate follow-up to `smaqit-adk sync` reporting drift in a non-interactive context (e.g. right after `install.sh`).

## Examples

**Input:** User runs `smaqit-adk sync` from a terminal, sees "1/3 entries drifted", then asks Claude to fix it.
**Output:** This skill enumerates the base registry, confirms the same single drifted entry, shows its target path and last-compiled time, and asks: recompile from source, or keep the hand edit? User picks recompile — the skill invokes `smaqit.L2` with the recorded definition/template paths, regenerates the target, and reports "1 entry recompiled from source."

## Gotchas

- This skill and the plain `smaqit-adk sync` CLI command are not the same thing: the CLI is a passive, non-interactive diagnostic (usable from a raw shell or during `install.sh`, where no LLM is available to do anything beyond reporting); this skill is the interactive, LLM-driven counterpart that actually resolves what the CLI can only report. Point users who ran `smaqit-adk sync` and saw drift to this skill next.
- Recompiling from a recorded source only works if `sourceDefinition`/`sourceTemplates` still exist and are unchanged — if the actual template/rules content itself has since moved on, this isn't the right tool; that's a live-cascade case (`smaqit.new-template`/`smaqit.new-rules`), not a drift-reconciliation case.

## Completion

- [ ] Registry entries enumerated for every applicable scope (base always; extension when `.smaqit/registry.json` exists)
- [ ] Every entry's current hash compared against its recorded `compiledHash`
- [ ] Every drifted entry presented to the user as a choice, never resolved automatically
- [ ] Chosen resolution applied (recompile via `smaqit.L2`, or re-register via `smaqit-adk registry put`) for each drifted entry the user addressed
- [ ] Final summary reported

## Failure Handling

| Situation | Action |
|-----------|--------|
| Required input not provided | Request the missing information before proceeding |
| Gathered input is ambiguous | Flag the ambiguity and ask for clarification |
| Subagent invocation fails | Report the failure with context; do not silently retry |
| Output artifact already exists | Confirm with user before overwriting |
| No registry file in any scope | Report "nothing registered yet" — not an error |
| A drifted entry's `sourceDefinition` or a `sourceTemplates` path no longer exists | Report as unrecoverable via recompile; offer only the keep-and-re-register resolution |
| `smaqit.L2` recompile fails | Report the failure; leave the entry drifted and unresolved; do not fall back to re-registering the hand edit without asking |
| `smaqit-adk` binary not on PATH | Instruct the user to run the smaqit-adk installer (`install.sh`) |
| User declines to resolve a drifted entry | Leave it drifted; report it as still-open in the final summary; do not re-prompt later in the same run |
