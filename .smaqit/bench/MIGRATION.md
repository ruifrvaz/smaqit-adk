# Legacy Eval Migration Map

Source: `tests/evals/{skills,agents}/**/*.json`, graded entirely by an LLM transcript judge (`tests/evals/runner/main.go`'s `grade()`). None of the 40 criteria across the 7 legacy case files were deterministically graded — every one was a YES/NO judgment against a conversation transcript. This document records, per legacy case, which criteria are retained as deterministic HarnessBench expectations and which are retired because they assert something about the conversation itself (question-asking order, confirmation gates, refusal wording) with no observable-artifact equivalent.

Retired criteria are not a loss of coverage in HarnessBench's model: Task 026's design decision is deterministic grading only, replacing an LLM judge with file/text/directory/command expectations — not replacing it with another hidden judge.

## `smaqit.create-agent`

Legacy: `001_gather_name_first.json`, `002_confirmation_before_compile.json`, `003_ambiguity_flagging.json`.

**Retained → `.smaqit/bench/skills/smaqit.create-agent/bench.yaml`:**
- Definition file is written to `.smaqit/definitions/agents/<name>.md` covering the required sections (file + content-presence expectation) — from 001.
- Compilation reaches `.github/agents/<name>.agent.md` (file-existence expectation) — from 001.
- A vague request still produces a definition file at `.smaqit/definitions/agents/<name>.md` (file-existence expectation) — from 003.
- At least one field in that definition file is suffixed with `[?]` and an inline uncertainty note (text/regex-presence expectation for `\[\?\]`) — from 003.

**Retired (conversational, no artifact equivalent):**
- "Asks for the agent name as its only interactive question" / "does not ask a follow-up question about X" (001) — turn-taking behavior.
- All of 002's criteria (proceeds without a confirmation summary, doesn't wait for "yes", reports the compiled path, lists `[?]` items in chat) — asserts what the assistant said and when, not an artifact.
- "Lists the `[?]`-annotated items in its final report" (003) — asserts the chat report, not the file (the file-level `[?]` presence check above covers the artifact side of this).
- Forbidden criteria in all three cases — asserts the assistant did *not* say/do something conversational; no artifact to check.

## `smaqit.create-skill`

Legacy: `001_full_gathering_flow.json`, `002_reject_first_person_description.json`.

**Mislabeled case:** `002_reject_first_person_description.json`'s filename and description claim it tests rejecting a first-person skill description, but every criterion in the file actually asserts validator-script gating (`scripts/validate-skill.go` must run before completion is reported) — identical in kind to one of 001's criteria. Per the resolved design decision, its actual content is migrated (folded into the single validator-gating expectation below) and the "reject first-person description" scenario is recorded here as never having been implemented — it is not re-created as new scope.

**Retained → `.smaqit/bench/skills/smaqit.create-skill/bench.yaml`:**
- Definition file is written to `.smaqit/definitions/skills/<name>.md` covering the required sections (file + content-presence expectation) — from 001.
- Compilation reaches `.github/skills/<name>/SKILL.md` (file-existence expectation) — from 001.
- The compiled skill's validator (`scripts/validate-skill.go`) runs and exits `0` before completion (command expectation) — from 001 and 002 combined (both asserted this; 002 had no other content).

**Retired (conversational):**
- "Asks for the skill name as its only interactive question" / "does not ask a follow-up question about X" (001).
- "Does not report the skill as complete before the validation step has run" (002) — ordering of the chat report relative to a tool call, not an artifact state.
- Forbidden criteria in both cases — assistant-said/didn't-say assertions.

## `smaqit.L2`

Legacy: `001_compile_base_agent.json`, `002_reject_unresolved_placeholders.json`.

**Retained → `.smaqit/bench/agents/smaqit.L2/bench.yaml`:**
- The compiled agent file contains no unresolved placeholders (`[DOMAIN]`, `[PREFIX]`, `[PHASE]`) — text/regex-absence expectation against the compiled file. Covers both 001's positive criterion and forbidden criterion (same check, framed twice in the legacy file) and 002's "does not write a file containing placeholders" / forbidden pair.
- The compiled agent file contains `MUST` and `MUST NOT` directive sections (text/section-presence expectation) — from 001.

**Retired (conversational or non-deterministic):**
- "Outputs the compiled agent content in the chat response" (001) — literally requires content to appear in chat.
- "The compiled agent file does not include principle explanations or rationale prose" (001) — a stylistic judgment without a clean deterministic equivalent; not migrated as a hard expectation. Could become a heuristic optional command grader later if it proves valuable, but that's out of this task's scope.
- "Refuses to compile and asks clarifying questions" (001 forbidden), "detects... reports which placeholders need to be resolved" (002), "refuses to compile" (002), "silently omits... without reporting" (002 forbidden) — all assert what the assistant said/did in conversation, not an artifact.

## Summary

| Target | Legacy cases | Legacy criteria (expected+forbidden) | Retained (deterministic) | Retired (conversational) |
|---|---|---|---|---|
| `smaqit.create-agent` | 3 | 18 | 4 | 14 |
| `smaqit.create-skill` | 2 (1 mislabeled, folded) | 9 | 3 | 6 |
| `smaqit.L2` | 2 | 13 | 2 (deduplicated from 4 overlapping assertions) | 11 |

None of the retired criteria are dropped silently — each is listed above with the reason it has no artifact-level equivalent, satisfying the acceptance criterion that every retained legacy scenario has a documented replacement or a recorded reason it was retired.
