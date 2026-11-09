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
| Strings and string iteration | partial | All text intrinsics/iteration implemented; tests cover UTF-16, bounds, symbolic replay and budgets; calculator/route declaration consumers pass; exhaustive declaration-only coverage remains |
| Numeric and tuple library | partial | Numeric and finite-tuple library implemented with conversion/bounds/symbolic tests; calculator/route declaration consumers pass; exhaustive declaration-only coverage remains |
| Explicit failure | partial | fail/failAt, statement-only use, deferred operands and producer/consumer spans tested; repeated direct/native-alias failures preserve distinct root sites in source and fresh declaration consumers. Mapped/instantiated-member recovery Types and native relation proofs reached through failures are discarded after stack unwinding, with query-order, assignment and declaration-consumer coverage. Substitution, conditional-root and generic-alias publication also retain failures inside containers/distribution/unknown absorption. Other native-cache failure retention and complete recipe coverage remain |
| Templates and match | partial | Template AST/parser/printer/API and evaluation implemented; match scopes, captures/guards, tuple/object and template/pin execution added; native-island patterns implemented; exhaustive native shape/dependency cases remain |
| Local helpers and closures | partial | Named Type-returning helpers, frozen captures, recursion, generic replay and editor scope tests implemented; first-class helper bindings, arity-preserving reassignment and anonymous closures and constructor results implemented; nongeneric native aliases execute; native islands capture constructors and named combinators; full outer-native lexical dependencies and declaration recipes remain |
| Restricted constructors/HKT | partial | Helper constructors/factory results and finite-kind checks implemented; mixed Type/helper/native-constructor inputs execute in DSL and native calls, preserving native bounds and captures; nongeneric native aliases execute with readiness checks; native islands capture constructors and named combinators; full outer-native lexical dependencies and declaration recipes remain |
| Native islands | partial | Frozen native capture scopes and whole-island symbolic replay implemented; stable symbolic shape publication, environment reuse and full dependency coverage remain |
| Union and structural predicates | partial | Native object/function/mapped/recursive cases now tested; exhaustive dependency summaries, ordinary cast/assignment boundaries and native relation fidelity remain |
| Reflection | partial | Classification, arrays, indexers and ordered signature handles implemented/tested; nongeneric signature descriptors and public properties implemented; tuple descriptors implemented; exhaustive native-producer optional provenance/nominal boundaries remain |
| Structural builders | partial | Property/indexer descriptors, symbol keys, arrays and native signature forwarding implemented; tuple-based functions/constructors implemented; full signature descriptors and native tuple spreads implemented; tuple descriptors implemented; native optional-producer provenance remains |
| Recursive structural graphs | partial | Arenas, guarded sealing, groups, nested builders and property/indexer/tuple/signature descriptors execute; ordered native overloads and discarded checks are preserved. Remaining: broader native adapters, exhaustive symbolic metadata, declaration recipes and cancellation publication tests |
| Generic deferral/inference | partial | Native lexical invocation environments replay by binder identity; Ambient island parameters, aliases/defaults and value queries now replay; polymorphic `this` and receiver queries now replay; contextual literal checking handles opaque/guarded base constraints without recursion or discharging requirements; latent member dependencies, generic-body requirements and exhaustive substitution boundaries remain |
| Constraints/result kinds | partial | Result-kind inference/annotations and recursive component validation implemented; computed bounds now share application-time evaluation and preceding-parameter scope; full constructor contexts remain |
| Declaration recipes | partial | Written recipes and constructor aliases emit with ABI 1; fresh-checker consumers test calculators/routes, frozen captures, namespaces, private helpers and native identities. Inferred call-return recipes retain native generic substitution, selected inference defaults, single-signature bounds/parameter checks, private cross-module helpers, identifier/single-return forwarding, native specialization expressions, inferred/constructed callable returns and consumer augmentation; ordinary/forced/builder atomic suppression is tested, including malformed DSL. Remaining: inferred-recipe lifting, complete dependency-closure failure handling, declaration-map navigation, redirects and invalidation |
| Editor and public API | partial | Services for all syntax, prelude completion, refactorings, auto-imports and API compatibility |
| Resources/cancellation | partial | Bigint inputs share preallocation magnitude checks across literals, arithmetic and conversions, including native arguments. Native relations, substitution, conditional evaluation and union/intersection collection/sorting poll cancellation internally, unwind and suppress interrupted cache publication. Complete work/allocation/cache budgets and remaining native-operation cancellation are still required |
| Caching/performance | partial | Component-local tuple environments bound recursive alias growth; paired profile-0 compiler-fixture measurements find no clear regression from native polling. Completed-call/blocked-result reuse, eviction/provenance, enabled-profile and broader native/workload measurements remain |
| Community workloads | partial | type-fest finite-literal word scanner and React Hook Form field-path traversal compared with native algorithms, including symbolic/options, metadata/nominal, negative, declaration, augmentation and resource cases. Public Words normalization and the remaining workload families remain; see [executed workloads](workloads.md) |

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
