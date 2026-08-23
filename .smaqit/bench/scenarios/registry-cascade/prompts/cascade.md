Add a principle to this project's ADK framework, then carry out the full
registry-driven propagation cascade this project defines for that change.
This is a non-interactive run: make every decision yourself and do not wait
for further input.

## Step 1 — smaqit.new-principle / smaqit.L0

If the variant treatment table lists `skill`, read its SKILL.md first and
follow it exactly. That skill describes invoking `smaqit.L0` as a subagent;
no native subagent spawning is available in this environment, so if the
treatment table lists `l0agent`, read those instructions and apply them
inline instead, as if L0 had been invoked.

The project's framework files are already present, writable, at
`framework/*.md` relative to your working directory — edit them there
directly.

Add exactly this principle, verbatim, as a new bullet point in whichever
file you judge most appropriate (framework/SKILLS.md is the best fit — it
already covers skill-format principles):

"Terminal Failure Row: Every skill's failure handling table ends with a
catch-all row for errors not otherwise covered, directing the agent to stop
and report rather than guess."

Follow L0's own directives when choosing where and how to add it: principle
form only — no MUST/MUST NOT/SHOULD directive language, no implementation
details, no procedural steps.

## Step 2 — smaqit.L1 review

This specific principle change requires a compiled-rules update: it is not
a "no update needed" case. If the treatment table lists `l1agent`, read
`agents/smaqit.L1.md`'s instructions and apply them inline, as if L1 had
been invoked as a subagent, to update
`templates/skills/compiled/skill.rules.md`.

Update its "Base Failure Handling Pattern" section: append exactly this row
as the new final row of the table shown there —

| Unrecoverable error not covered above | Stop and report the failure verbatim; do not guess a resolution |

— and add one line to the "Compilation Directives" section instructing
Agent-L2 to include this catch-all row as the final row of every compiled
skill's own Failure Handling table.

## Step 3 — registry lookup

The `smaqit-adk` binary is on `PATH` in this environment (a hidden
`registry` subcommand — undocumented, not in `smaqit-adk help`). Run:

```
smaqit-adk registry get --scope base --depends-on templates/skills/compiled/skill.rules.md
```

This returns the registered entries that depend on the file you just
updated. There is exactly one entry pre-registered in this environment:
`smaqit.sync`, whose compiled output lives at `skills/smaqit.sync/SKILL.md`.

## Step 4 — smaqit.L2 recompile

If the treatment table lists `l2agent`, read `agents/smaqit.L2.md`'s
instructions and apply them inline, as if L2 had been invoked as a
subagent, to recompile `smaqit.sync`:

- Read its definition file at `.smaqit/definitions/skills/smaqit.sync.md`
- Read `templates/skills/base-skill.template.md` and the now-updated
  `templates/skills/compiled/skill.rules.md`
- Recompile the full skill from these three sources (a real 3-way merge,
  not a patch) and overwrite `skills/smaqit.sync/SKILL.md` with the result
  — its Failure Handling table must end with the new catch-all row from
  Step 2
- Register the recompiled result by running:

```
smaqit-adk registry put --scope base --name smaqit.sync \
  --claude skills/smaqit.sync/SKILL.md \
  --source-definition .smaqit/definitions/skills/smaqit.sync.md \
  --source-templates templates/skills/base-skill.template.md,templates/skills/compiled/skill.rules.md \
  --hash-of skills/smaqit.sync/SKILL.md
```

## Step 5 — report

State plainly: which framework file received the new principle, that
`skill.rules.md` was updated, that `smaqit.sync` was recompiled, and that
the registry was updated. If the treatment set is empty, proceed without
any ADK guidance and do not search project-global or user-global ADK
paths — report that no cascade was possible instead.
