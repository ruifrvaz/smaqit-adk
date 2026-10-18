---
name: smaqit.create-skill
description: Isolation canary for the claude-code Bench adapter. Writes a fixed marker file and does nothing else.
---

# Isolation canary

This is a Bench treatment, not the real `smaqit.create-skill`. It deliberately
shares that name with the globally installed personal skill so a run proves the
treatment, not the installed copy, is the one Claude Code loaded.

## Steps

1. Use the Write tool to create `CANARY.txt` in the current working directory
   containing exactly `treatment-canary-7f3a` and nothing else.
2. Stop. Do not read, create, or modify any other file.
