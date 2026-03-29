# smaQit Framework

This document establishes principles that apply across all concepts in an agentic framework — agents, skills, templates, and artifacts. These are system-level properties that must hold for a well-functioning agentic system, not properties of any single concept.

## Core Principles

### Explicit Over Implicit

**Every decision, assumption, boundary, and requirement is stated. Nothing in a well-functioning agentic system remains implicit.**

Implicit content compounds silently. An assumption that travels through a system without being named becomes invisible — to other agents, to downstream consumers, and to the humans responsible for the system. Each transition between stages is an opportunity for implicit content to diverge from intent. Making things explicit at every stage closes those gaps before they propagate.

**Invariants:**
- Assumptions are stated, not embedded.
- Boundaries are defined, not implied by convention.
- Requirements are explicit at the point they are consumed, not assumed from context.
- When something is unclear, it is named as unclear — not resolved silently.

---

### Single Source of Truth

**Each piece of information exists in exactly one authoritative location. Other locations reference the source rather than replicating it.**

Replication creates divergence. When the same information exists in multiple places, those places drift — through edits, updates, and omissions. Each diverged copy becomes a competing source of truth, and the system's behavior becomes dependent on which copy any given stage happens to consult. Single sourcing eliminates that class of inconsistency: there is always one place to update, one place to verify, and one place that is correct.

**Invariants:**
- Information is defined once and referenced elsewhere.
- When two locations contain equivalent content, one is the source and the other is a reference or derivation — never two independent definitions.
- A system that requires reconciling multiple copies of the same information has a sourcing problem.

---

### Traceability

**Every output in the system can be traced back to its origin. An output that cannot be traced to an input is either an invention or an error.**

Traceability is what makes a system auditable and its outputs trustworthy. Without it, a change to a requirement cannot be reliably propagated to the outputs that depended on it, and an output cannot be verified against what it was meant to fulfill. Traceability is not a documentation practice — it is a structural property that must be designed in at every stage, because it cannot be reconstructed after the fact.

**Invariants:**
- Every output references the input it was derived from.
- A system element with no traceable origin is not a valid output.
- Changes to a source either propagate to all outputs that depend on it, or those dependencies are explicitly invalidated.

---

### Composability

**Systems are built from single-responsibility parts that compose, not from monoliths that do everything.**

A part with one responsibility can be replaced, verified, and reused independently. A part with many responsibilities cannot. Composability is not a style preference — it is what makes a system maintainable when requirements change and trustworthy when parts are combined. Composition requires that each part has a clearly bounded role and interacts with other parts only through explicit interfaces.

**Invariants:**
- Each part of the system has exactly one responsibility.
- Parts interact through explicit interfaces, not through shared state or implied convention.
- A part that cannot be used or verified independently is not composable.

---

### Validate Behavior, Not Reproduction

**The measure of correctness is whether the system behaves as specified — not whether it produces identical output across runs.**

Agentic systems are non-deterministic. The same inputs produce outputs that differ in form while remaining equivalent in behavior. Enforcing identical reproduction fights this property at significant cost for no gain. The correct target is behavioral equivalence: outputs that satisfy the same specification, meet the same acceptance conditions, and fulfill the same requirements — regardless of surface differences.

**Invariants:**
- Correctness is defined by specified behavior, not by output identity.
- Acceptance conditions are the basis for evaluating output, not comparison to a reference output.
- Two outputs that satisfy the same specification are equivalent, regardless of how they differ in form.
