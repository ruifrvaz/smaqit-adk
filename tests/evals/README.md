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
# Local development (gh CLI)
cd installer && GH_TOKEN=$(gh auth token) make evals

# CI
cd installer && make evals COPILOT_GITHUB_TOKEN=<token>
```

## Auth

An explicit token is **required**. Set one of:

1. `COPILOT_GITHUB_TOKEN` environment variable
2. `GH_TOKEN` environment variable
3. `GITHUB_TOKEN` environment variable

**Why a token is required:** Without an explicit token, the Copilot Go SDK falls back to `UseLoggedInUser`, which routes sessions through the VS Code Copilot extension. This injects the currently open VS Code workspace (smaqit-adk) as context into every eval session, contaminating the isolated scaffold. With an explicit token, the SDK talks directly to the GitHub Copilot API and VS Code is not involved.

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
