---
status: PR Open
created: "2026-08-24"
mode: Assisted
started: "2026-08-24"
pr: 3
---

# Installer Get-Started Message Omits Bench Usage

## Description

`install.sh`'s post-install "Get started" summary (printed at the end of a successful install) tells
the user to say "create a new agent"/"create a new skill" and to run `smaqit-adk --help`, but never
mentions that `smaqit-adk bench` exists — a core shipped capability (local process evaluation and
benchmarking). A new user reading only the install output has no hint that bench exists unless they
separately dig into `--help` output or the README's Quick Start section.

## Issue Triage Context

**Mode:** Skip
**Technologies:** Bash (install.sh)
**Platforms/Environments:** None
**Features/Integrations:** install.sh distribution, smaqit-adk bench
**Versions/Constraints:** Current adk version 3.0.0

## Design Decisions

- **Scope limited to install.sh's own printed output.** `installer/main.go`'s `cmdHelp()` already
  documents bench in full, and README's Quick Start section already has a full bench walkthrough
  further down — neither needs a duplicated change for this narrow gap.

## Implementation Steps

1. Add one line to `install.sh`'s `main()` "Get started" echo block referencing `smaqit-adk bench`,
   alongside the existing create-agent/create-skill/`--help` lines.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [x] `install.sh`'s "Get started" echo block mentions `smaqit-adk bench` and points at how to learn
      more (e.g. `smaqit-adk bench --help` or `smaqit-adk --help`)
- [x] The addition is a single short line, consistent in style with the existing "Get started" block —
      no restructuring of unrelated `install.sh` content

## Findings

[Populated by smaqit.task-complete. Do not fill in manually before task is complete.]

**Implementation approach:**
- Added one `echo` line to `install.sh`'s `main()` "Get started" block, between the existing create-agent/create-skill line and the `--help` line, pointing at `smaqit-adk bench validate|plan|run`

**Decisions made:**
- Scope stayed limited to `install.sh`'s own printed output, as planned — `installer/main.go`'s `cmdHelp()` and README's Quick Start section already cover bench and didn't need duplicated changes

**Blockers encountered:**
- None

**Follow-up identified:**
- None

## Files to Create / Modify

| File | Action |
|------|--------|
| `install.sh` | Modify — add one line to the "Get started" echo block mentioning `smaqit-adk bench` |

## Notes

- Found by the user immediately after Task 034's live cross-device verification of the new `gh`-based
  install one-liner.
