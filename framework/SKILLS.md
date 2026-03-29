# Skills

This document establishes the behavioral principles and invariants that govern skills. These principles describe how skills must be built and behave — not what skills exist or how they are packaged.

## Core Principles

### Instructions, Not Data

**Skills contain instructions for agents — what to gather, how to validate, how to compile. They are not repositories for user requirements or accumulated input.**

A skill tells an agent how to execute a workflow, not what the workflow's inputs are. When user requirements are embedded in a skill, two distinct concerns merge: the instruction format and the data format. This conflation makes the skill change with every execution, breaks reproducibility across runs, and produces a version-controlled file that diverges between projects. A skill that stores user data is no longer a read-only instruction artifact.

**Invariants:**
- Skills contain gathering steps, validation rules, and compilation guidance — no user requirements or accumulated input.
- A skill file is unchanged between executions.
- User requirements gathered during execution live in conversation context or in a separately defined compilation artifact, not in the skill.

---

### Progressive Disclosure

**Agents load only the skill name and description at discovery — the full body loads only when the skill is activated.**

Discovery across all installed skills is cheap because only the name and description load at that stage. The full body — which can carry detailed multi-step instructions — loads only when the agent determines the skill is relevant. The cost of a large skill library is therefore bounded by description size, not body size. The description carries the full weight of the activation decision; the body carries the full weight of the execution.

**Invariants:**
- At discovery, only the skill name and description are loaded.
- The full skill body is read only when the skill is activated for the current task.
- The description is the sole signal used to determine whether a skill matches.

---

### Instruction-Only Content

**Skill bodies contain only gathering steps, validation rules, and compilation guidance. Skills do not accumulate requirements, states, or execution records.**

The skill body is a procedure, not a record. When the skill body grows to include execution state or gathered context, it has become a hybrid artifact that mixes instruction with data. The skill body must remain stable and read-only — identical before and after every execution.

**Invariants:**
- The skill body grows only when the procedure itself changes, not when the skill is executed.
- Execution-time data is never written into the skill body.
- An artifact that changes between executions is not a skill body.

---

### Context-Driven Input

**Requirements gathered during skill execution live in conversation context. If persistence is needed, the agent deposits gathered input into a separately defined artifact — not into the skill.**

The skill is a shared, version-controlled instruction. Making it the destination for gathered requirements means it leaves each execution in a different state. When persistence of gathered input is needed, the skill instructs the agent to write a designated external artifact, preserving the skill as identical before and after every run.

**Invariants:**
- Gathered input is not written into the skill file.
- When persistence is required, the skill instructs the agent to write a designated external artifact.
- The skill is identical at the start and end of every execution.

---

### Description-Driven Activation

**Skill descriptions are the sole activation signal. A description must explain what the skill does and when to use it — not just name it.**

At discovery, the agent reads only the name and description to determine whether a skill matches the current task. A label gives the agent a name but no context to evaluate fit. An explanation gives the agent enough context to distinguish the skill from adjacent ones and to decline false positives. The description is a decision surface, not a title — its length is determined by the precision needed to avoid false matches, not by a brevity target.

**Invariants:**
- A skill description explains what the skill does and when to invoke it — it does not merely label the skill.
- The description is precise enough to distinguish the skill from adjacent skills that share trigger words.
- Description length is determined by the precision needed to avoid false positives, not by a brevity convention.
