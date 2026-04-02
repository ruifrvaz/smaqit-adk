#!/usr/bin/env bash
# Sets up a blank eval host workspace OUTSIDE the smaqit-adk repo tree.
#
# VS Code Copilot walks up from the open folder to find the git root and loads
# agents, instructions, and skills from there. The host workspace must be a
# separate git repo so that discovery stops at its own root, not smaqit-adk.
#
# Usage: bash tests/evals/host-setup.sh
# Then:  code ~/smaqit-eval-host   (open in a separate VS Code window)
# Then:  run evals from that VS Code's integrated terminal

HOST_DIR="${SMAQIT_EVAL_HOST:-$HOME/smaqit-eval-host}"

if [ -d "$HOST_DIR/.git" ]; then
  echo "Host workspace already exists at $HOST_DIR"
  echo "To open it: code $HOST_DIR"
  exit 0
fi

echo "Creating eval host workspace at $HOST_DIR ..."
mkdir -p "$HOST_DIR/.github/agents"
mkdir -p "$HOST_DIR/.github/skills"

cd "$HOST_DIR"
git init -q
git commit --allow-empty -q -m "init eval host"

echo ""
echo "Done. Open it in a separate VS Code window:"
echo ""
echo "  code $HOST_DIR"
echo ""
echo "Then from that window's integrated terminal, run evals:"
echo ""
echo "  cd $(dirname "$(dirname "$(cd "$(dirname "$0")/.." && pwd)")")/installer && make evals"
echo ""
echo "Or set SMAQIT_EVAL_HOST to use a different path."
