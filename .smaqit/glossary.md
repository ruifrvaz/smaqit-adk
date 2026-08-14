# Project Glossary

## Bench

**Bench Case**

One evaluation scenario: a fixed user goal, common project fixture, shared inputs, expectations, and graders. A Case is the unit that is repeated and evaluated; it is unrelated to a smaqit tracked Task.

---

**Case brief**

The rendered instruction delivered to a harness for one Case attempt. It contains the raw Case prompt plus tables of the shared inputs and variant treatments actually available for that attempt. The raw prompt is not rewritten or placeholder-expanded.

---

**Fixture**

Common writable project material copied into each isolated Bench execution workspace before a Case attempt. It establishes the same editable starting state for every variant. A fixture may contain any project files that the Case must edit; the Case prompt is not a fixture, and a skill or agent is normally a treatment when it is the thing being compared (or a given input when it is shared reference material).

---

**Given input**

Common read-only source material made available to every variant, such as a specification, image, or reference document. Given inputs are not the artifact being compared.

---

**Harness**

The local process or adapter that receives a Case brief and executes an attempt in its isolated workspace. A harness may invoke a model CLI or API, an agent runner, or a custom script; it is not the model itself. Its executable, version, arguments, and environment can be part of a Variant’s process configuration.

---

**Treatment**

The variant-only artifact or configuration under evaluation, staged read-only in Bench's excluded sidecar. Examples include a skill version, custom-agent body, system instruction, retrieval corpus, or harness configuration.

---

**Variant**

One alternative way to run the same Bench Case. Variants share the Case's prompt, fixture, expectations, and graders; they differ only in the declared treatment and/or harness process configuration. Bench compares their repeated results to report a winner, tie, or inconclusive outcome.

---
