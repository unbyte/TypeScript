# Implementation acceptance tracker

The specification is the completion boundary. This tracker does not narrow it.
A passing subset, a clean build, or an editor demonstration is not language
completion. Each implementation commit must include relevant executable evidence;
update this tracker and implementation.md when a capability changes status.

Status: **partial** means useful behavior exists but the full specified contract
is not satisfied; **pending** means no executable implementation meets the contract.
Tests must assert actual native Types and complete diagnostics, not printed comments.

| Contract | Status | Remaining acceptance evidence |
| --- | --- | --- |
| Core bindings/control/scalars | partial | Complete definition-time kind/return-path validation and scalar edge cases |
| Strings and string iteration | partial | All text intrinsics/iteration implemented; tests cover UTF-16, bounds, symbolic replay and budgets; declaration-only coverage awaits recipes |
| Numeric and tuple library | partial | Numeric and finite-tuple library implemented with conversion/bounds/symbolic tests; declaration-only coverage awaits recipes |
| Explicit failure | partial | fail/failAt, statement-only use, deferred operands and producer/consumer spans tested; declaration-only cases await recipes |
| Templates and match | partial | Template AST/parser/printer/API and evaluation implemented; match scopes, captures/guards, tuple/object and template/pin execution added; native-island patterns and general structural relations remain |
| Local helpers and closures | pending | Lexical capture snapshots, constructor identity, recursive captures and scope diagnostics |
| Restricted constructors/HKT | pending | Fixed-arity slots, passing/returning constructors and kind errors |
| Native islands | partial | Executable native Types using frozen local environments; binding and replay |
| Union and structural predicates | partial | Object/discriminated unions, graph readiness and native relation fidelity |
| Reflection | pending | Classification, properties/indexers, tuple/array/signature metadata and nominal boundaries |
| Structural builders | partial | Descriptors, symbol keys, general tuple spreads, callable/constructable Types and overload order |
| Recursive structural graphs | pending | Single/group sealing, guarded edges, invalid cycles and cancellation |
| Generic deferral/inference | partial | Latent member dependencies, generic-body requirements and exhaustive substitution boundaries |
| Constraints/result kinds | partial | Resolve documented scope/timing discrepancies; retain and check result kinds |
| Declaration recipes | pending | Emit/consume standalone declarations, ABI checks, captures, source maps and dependency invalidation |
| Editor and public API | partial | Services for all syntax, prelude completion, refactorings, auto-imports and API compatibility |
| Resources/cancellation | partial | All specified work/allocation/cache budgets and cancellation within expensive native work |
| Caching/performance | partial | Completed-call and blocked-result reuse, eviction/provenance, workload measurements and native overhead |
| Community workloads | pending | Ports paired with concrete, symbolic, negative, metadata, declaration and resource cases |

The detailed required cases are in [conformance](spec/conformance.md), with
workload families in its community traceability matrix. The calculator and HTTP
route programs are mandatory milestones, not substitutes for that matrix.

Deliberately excluded: enum-specific APIs, new generic declaration synthesis,
unrestricted HKT, runtime/I/O/regex and inspecting an active inference session.
`where` is unresolved research, not adopted syntax; native extends is the current
constraint contract. See [design decisions](design-notes.md).

## Executable milestones

* Calculator and numeric example modules are read directly by the integration
  harness (`example-calculator`, `example-numbers`). Assertions cover exact Types,
  precedence, associativity, unary signs, whitespace, bigint precision and generic
  specialization. Separate calculator fixtures check empty input, missing operands
  and parentheses, division by zero, nonintegral division and trailing input at the
  importing application. This completes these examples, not the language matrix.
