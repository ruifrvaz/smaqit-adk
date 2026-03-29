# Templates

This document establishes the principles and invariants that govern templates. A template is the structural contract that defines what a compiled output must contain — every required section, its purpose, and its order.

## Core Principles

### Template as Compilation Surface

**Templates define what is architecturally invariant and what varies by role — compilation fills in the variation while preserving the invariant.**

A template captures the structure every compiled agent must have: sections, their order, and behavioral shape. It marks every location where role- or domain-specific content belongs as a placeholder. Compilation is the act of resolving those placeholders without altering the structure. A template is not a guideline; it is the required shape of the compiled output.

**Invariants:**
- Every section that must appear in a compiled agent is present in the template, even if its content is a placeholder.
- Every value that varies by role or domain is expressed as a placeholder — not hardcoded in the template.
- A compiled agent that omits a required section or retains unresolved placeholders is not valid.

### Placeholder Convention

**Placeholders are the compilation interface — marked locations where the compiler resolves generic structure into domain-specific content.**

All placeholders use a consistent format: screaming-case inside square brackets (`[PLACEHOLDER]`). This allows the compiler to locate and resolve every placeholder unambiguously. The complete set of placeholders a template defines is that template's compilation contract. The meaning of each placeholder is documented in the corresponding compilation rules file — not in the template itself.

**Invariants:**
- Every placeholder follows the `[SCREAMING_CASE]` format inside square brackets.
- Every placeholder in a template has a defined meaning in the template's compilation rules file.
- A valid compiled agent contains no unresolved placeholders.

### Template Hierarchy

**Templates organize into a foundation that all agents share and extensions that add role-specific behavior without duplicating the foundation.**

The foundation template captures what must remain invariant across every compiled agent: bounded scope, self-validation, traceable references, fail-fast behaviors, template-constrained output. Extension templates inherit the full foundation and add sections for role-specific concerns. Extensions never redefine or duplicate foundation sections. Changes to the foundation propagate to all agent types; role-specific changes remain isolated.

**Invariants:**
- A compiled agent's foundation structure is identical regardless of which extension template was used.
- An extension template adds sections to the foundation; it does not override or duplicate foundation sections.
- What differentiates agents by role lives exclusively in extension templates.

### Section Structure as Behavioral Contract

**Template sections define what a compiled agent must contain; their presence, purpose, and order are part of the structural contract.**

Template sections are not organizational suggestions — they are the behavioral map of a compiled agent. Each section captures a distinct concern (role, input, output, directives, scope, completion, failure handling). Compilation fills sections; it does not add new sections beyond what the template defines or omit required ones. This predictability makes compiled agents structurally consistent across runs and domains.

**Invariants:**
- Every template section appears in the compiled agent.
- The compiler does not add sections not defined in the template.
- The compiler does not omit required sections.
- Section order in compiled agents matches the template order.

### Extension Inheritance

**Extensions inherit the full foundation — they are strictly additive.**

Compiling with an extension template means all foundation content is included first, then extension-specific content is appended. Extensions do not weaken, override, or selectively apply foundation behaviors. Every compiled agent, regardless of role, carries the complete foundation.

**Invariants:**
- All base directives appear before extension directives in the compiled agent.
- Extension directives do not contradict or remove base directives.
- The boundary between base and extension content is recognizable in every compiled agent.
