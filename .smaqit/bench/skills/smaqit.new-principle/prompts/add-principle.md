Add a new principle to this project's ADK framework.

If the Case brief lists a variant treatment named `skill`, read its SKILL.md
first and follow it exactly — do not invent your own approach instead. That skill
describes invoking the `smaqit.L0` agent as a subagent; L0 is intentionally
not registered in this isolated Bench workspace, so if the treatment table
lists `l0agent`, read those instructions and apply them inline. If the
treatment set is empty, proceed without ADK guidance and do not search
project-global or user-global ADK paths.

The project's framework files are already present, writable, at
`framework/*.md` relative to your working directory (this is the real
project tree, not the read-only staged input directories) — edit them there
directly, preserving the existing structure and heading conventions of
whichever file you judge most appropriate.

Add exactly this principle, verbatim, as a new bullet point:

"Fail Fast: An agent surfaces an error immediately upon detecting an invalid
precondition, rather than continuing and producing an unreliable result."

Follow the L0 agent's Directives when choosing where and how to add it:
principle/concept/mapping form only — no MUST/MUST NOT/SHOULD directive
language, no implementation details, no procedural steps. If your workflow
calls for asking a clarifying question, proceed with your best judgment
instead — there is no further input coming.
