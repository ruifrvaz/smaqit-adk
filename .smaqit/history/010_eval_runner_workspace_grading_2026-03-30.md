# Eval Runner Workspace and Grading

**Date:** 2026-03-30
**Session Focus:** Fixing eval runner bugs — workspace placement, grader file visibility, and per-eval timeout
**Tasks Completed:** —
**Tasks Referenced:** —

---

## Actions Taken

- Changed eval workspace from random `os.MkdirTemp` to `runs/<id>/workspaces/<eval-name>/` — files written by agents are now preserved and inspectable after each run
- Removed `defer os.RemoveAll` so run directories are fully self-contained
- Added `collectWorkspaceFiles` function that walks the workspace post-session, appends agent-written file contents to the grading transcript (excludes seeded setup files: `templates/`, `framework/`, `smaqit.L2.agent.md`)
- Raised per-eval timeout from 5 → 10 minutes to accommodate the 14-turn `smaqit.new-agent/001` eval

---

## Problems Solved

- **L2/001 FAIL**: Grader only saw the chat transcript; L2 wrote files to disk without echoing content in chat. Fixed by appending workspace file contents to the transcript before grading.
- **new-agent/001 ERROR**: `context deadline exceeded` — 14 ask_user exchanges plus continuation loop exceeded 5 minutes. Fixed by raising timeout to 10 minutes.
- **Workspace in temp dir**: Files created by agents were discarded after each run, making failures undebuggable. Fixed by placing workspaces inside the run directory.

---

## Decisions Made

- `collectWorkspaceFiles` skips entire subtrees (`templates/`, `framework/`) and the seeded L2 agent file — only agent-written files are appended to the transcript
- Workspace path uses `/` → `__` sanitization of the relative eval path for safe directory naming
- Workspaces are not cleaned up — intentional for post-run inspection

---

## Files Modified

- `tests/evals/runner/main.go` — workspace placement, `collectWorkspaceFiles`, timeout increase

---

## Next Steps

- Run full eval suite with `GH_TOKEN=$(gh auth token) go run ./evals/runner/... ./evals/` from outside VS Code to validate all 7 evals pass
- Investigate L2/001 further if L2 still only claims to write rather than actually writing

---

## Session Metrics

- Duration: ~1 session continuation
- Files modified: 1
- Functions added: 1 (`collectWorkspaceFiles`)
- Bugs fixed: 3 (workspace placement, grader visibility, timeout)
- Evals passing at session end: 5/7 (pre-fix run)
