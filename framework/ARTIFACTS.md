# Artifacts

Artifacts are the outputs produced by agents. This document establishes the principles and invariants that govern them.

## Core Principles

### Artifact Types

**Agents produce two kinds of artifacts: specification artifacts that declare what must be true, and implementation artifacts that make it true.**

The distinction is architectural: specification agents produce specification artifacts; implementation agents consume them and produce implementation artifacts. The two types are not symmetric — a specification is the contract, an implementation is the fulfillment of that contract.

**Invariants:**
- Specification artifacts are declarative — they state conditions without prescribing how those conditions are met.
- Implementation artifacts are imperative — they exist to satisfy the conditions declared by specification artifacts.
- An artifact is one or the other; nothing is both.

---

### Specification Self-Containment

**A specification is self-contained when another agent can implement or validate against it without requiring additional context.**

Self-containment is the quality that makes a specification usable as a contract. A specification that requires the reader to consult other documents, infer unstated conditions, or apply domain knowledge not present in the artifact is not a complete specification. Self-containment is the bar a specification must clear before it can be used downstream.

**Invariants:**
- A complete specification contains everything needed to implement or validate the conditions it asserts.
- A specification that requires external context to be actionable is not complete.
- Scope boundaries are explicitly stated within the specification.

---

### Acceptance Conditions

**Every specification carries acceptance conditions — verifiable criteria that must be satisfied for the specification to be considered fulfilled.**

Acceptance conditions make specifications actionable. A condition that cannot be verified cannot be used to determine whether a specification has been satisfied. Conditions that cannot be automatically verified are not excluded — they are identified as such, with a direction toward resolution.

**Invariants:**
- Every acceptance condition is measurable (quantifiable outcome), observable (externally verifiable), and unambiguous (single interpretation).
- Each condition carries a unique identifier that remains stable for the lifetime of the specification — identifiers are never reused or renamed after assignment.
- Conditions that cannot be automatically verified are explicitly identified as such; a path toward measurable alternatives or a resolution disposition is included.
- When an acceptance condition changes, any prior verification status for that condition is invalidated.

---

### Artifact Traceability

**Artifacts reference their sources. Implementation artifacts reference the specifications they satisfy. Each acceptance condition traces to a source requirement.**

Traceability makes the system auditable. An implementation without reference to its specification cannot be verified against it. A condition that cannot be traced to a source requirement cannot be validated for scope.

**Invariants:**
- Implementation artifacts reference the specification artifacts they satisfy.
- Each verifiable acceptance condition can be mapped to one or more test definitions.
- References use paths within the artifact store; no external references are required to resolve them.

---

### Implementation Dimensions

**Every implementation exists across three dimensions: behavior derived from specifications (invariant), structure derived from domain standards (consistent), and internal design (free).**

The Anchoring Principle: Implementations comply with domain-appropriate standards for their stack, while satisfying spec-defined behavior. Two compliant implementations may differ internally, but must be structurally recognizable and behaviorally equivalent.

The Test Independence Principle: Test artifacts exist independently of agent execution. Tests run in any environment with the appropriate runtime, enabling verification outside any particular workflow.

**Invariants:**
- Behavior is fully verifiable — spec-defined behavior must be satisfied exactly and can be verified by test.
- Structure is consistently verifiable — implementations follow domain-appropriate standards and are recognizable to practitioners.
- Internals are free — variable names, helper functions, and internal patterns may vary; they are not independently required to be verified.
- Test artifacts are independently executable — they do not depend on the agent runtime that produced them.

---

### Artifact Completeness

**An artifact is complete when it can be used by its downstream consumer without modification, clarification, or additional context.**

Completeness has different definitions for each artifact type, but the underlying principle is the same: a complete artifact requires nothing external to be acted upon.

**Invariants:**
- A complete specification has all template sections filled, all conditions uniquely identified and verifiable (or explicitly flagged with a resolution path), all references valid, and scope explicitly stated.
- A complete implementation satisfies all acceptance conditions in the specifications it references, follows domain-appropriate standards, documents traceability, and adds no unspecified behavior.
- An artifact that retains unresolved placeholders, omits required references, or leaves conditions unevaluated is not complete.
