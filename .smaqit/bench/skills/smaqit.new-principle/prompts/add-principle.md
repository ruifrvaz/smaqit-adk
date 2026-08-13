Add a new principle to this project's ADK framework.

If this project has an ADK principle-authoring skill staged at
`{input:skill}/SKILL.md`, read it first and follow it exactly — do not
invent your own approach instead. That skill in turn describes invoking the
`smaqit.L0` agent as a subagent; in this sandbox `smaqit.L0` is not
registered as a spawnable custom agent, so instead read the L0 agent's
own instructions directly at `{input:l0agent}` and apply them yourself.

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
