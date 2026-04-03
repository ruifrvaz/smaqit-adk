# Behavioral Evaluations

This directory contains Layer 3 of the smaqit-adk test framework: behavioral evaluations that drive live Copilot SDK sessions against ADK artifacts and grade the results.

## Structure

```
evals/
  skills/
    smaqit.new-agent/   ← eval files for the new-agent skill
    smaqit.new-skill/   ← eval files for the new-skill skill
  agents/
    smaqit.L2/          ← eval files for the L2 compiler agent
  runner/
    main.go             ← eval runner binary
  README.md             ← this file
```

## Running

```sh
cd installer && make evals
```

`make evals` auto-detects a token via `gh auth token`. An explicit token is required to prevent
eval sessions from routing through VS Code and loading smaqit-adk's workspace context.

> **Note:** Classic PATs are rejected by the Copilot agent API with a 400 error. Use the output of
> `gh auth token`, the `GITHUB_TOKEN` secret in CI, or a fine-grained PAT with the `copilot` scope.

### Why a token is required

The Copilot CLI (without a token) reads auth from the user's shared XDG config dirs. That token
belongs to whichever VS Code window last refreshed it — typically smaqit-adk, the active project.
There is no per-window token isolation; the shared config always wins. Switching VS Code windows does
not help.

With an explicit token the SDK passes `--no-auto-login` to the CLI and never touches VS Code or XDG
config. Sessions are fully isolated from any open VS Code instance.

## Auth

The runner checks, in order:

1. `GH_TOKEN` — set automatically by `make evals` via `gh auth token`
2. `GITHUB_TOKEN` — auto-provisioned by GitHub Actions in CI

If neither is set and `gh auth token` fails, the runner exits with an error.

## Eval file format

```json
{
  "type": "skill",
  "artifact_file": "skills/smaqit.new-agent/SKILL.md",
  "description": "Human-readable description of what this eval verifies",
  "turns": [
    { "user_input": "Initial message sent via session.SendAndWait" },
    { "user_input": "Answer fed via OnUserInputRequest callback", "trigger": "ask_user" }
  ],
  "expected_behavior": [
    "Criterion that must be observed in the transcript"
  ],
  "forbidden_behavior": [
    "Criterion that must NOT be observed in the transcript"
  ]
}
```

### Turn routing

| Turn | Routing |
|------|---------|
| No `trigger` | Sent via `session.SendAndWait` in order |
| `"trigger": "ask_user"` | Pre-loaded into a queue; popped by `OnUserInputRequest` as the agent calls `ask_user` |

### Grading

After the session completes, each criterion is evaluated by a second Copilot session that reads the full transcript and answers YES or NO. Expected behavior criteria must be satisfied; forbidden behavior criteria must not be present.

## Adding evals

1. Create a `.json` file in the appropriate subdirectory.
2. Set `artifact_file` to the path of the skill or agent file relative to the repo root.
3. Write `turns` to drive the conversation — use `"trigger": "ask_user"` for responses to `ask_user` calls.
4. Write clear, testable `expected_behavior` and `forbidden_behavior` criteria.

The runner discovers all `.json` files under `evals/` (excluding the `runner/` directory) without configuration.
