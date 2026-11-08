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
| Core bindings/control/scalars | partial | Conservative return-path and finite-kind validation implemented; constructor contexts and scalar edge cases remain |
| Strings and string iteration | partial | All text intrinsics/iteration implemented; tests cover UTF-16, bounds, symbolic replay and budgets; declaration-only coverage awaits recipes |
| Numeric and tuple library | partial | Numeric and finite-tuple library implemented with conversion/bounds/symbolic tests; declaration-only coverage awaits recipes |
| Explicit failure | partial | fail/failAt, statement-only use, deferred operands and producer/consumer spans tested; declaration-only cases await recipes |
| Templates and match | partial | Template AST/parser/printer/API and evaluation implemented; match scopes, captures/guards, tuple/object and template/pin execution added; native-island patterns implemented; exhaustive native shape/dependency cases remain |
| Local helpers and closures | partial | Named Type-returning helpers, frozen captures, recursion, generic replay and editor scope tests implemented; first-class helper bindings, arity-preserving reassignment and anonymous closures and constructor results implemented; native alias/island bridges remain |
| Restricted constructors/HKT | partial | Helper constructors/factory results and finite-kind checks implemented; constructor input slots, native generic constructor values and nongeneric native aliases remain |
| Native islands | partial | Frozen native capture scopes and whole-island symbolic replay implemented; stable symbolic shape publication, environment reuse and full dependency coverage remain |
| Union and structural predicates | partial | Native object/function/mapped/recursive cases now tested; exhaustive dependency summaries, ordinary cast/assignment boundaries and native relation fidelity remain |
| Reflection | partial | Classification, arrays, indexers and ordered signature handles implemented/tested; nongeneric signature descriptors and public properties implemented; tuple descriptors implemented; exhaustive native-producer optional provenance/nominal boundaries remain |
| Structural builders | partial | Property/indexer descriptors, symbol keys, arrays and native signature forwarding implemented; tuple-based functions/constructors implemented; full signature descriptors and native tuple spreads implemented; tuple descriptors implemented; native optional-producer provenance remains |
| Recursive structural graphs | pending | Single/group sealing, guarded edges, invalid cycles and cancellation |
| Generic deferral/inference | partial | Latent member dependencies, generic-body requirements and exhaustive substitution boundaries |
| Constraints/result kinds | partial | Result-kind inference/annotations and recursive component validation implemented; full constructor contexts and documented bound scope/timing discrepancies remain |
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

* HTTP-route example source now executes directly (`example-http-routes`). Native
  mapped parameter objects and handler signatures are asserted alongside route
  splitting, typed parameters, unions and generic specialization. Native-island
  negative fixtures check missing names, initialization, indexing, constraints and
  distinct importing-application diagnostics.
