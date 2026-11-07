> Investigation/history, not a current language contract. The [R2 specification](spec/README.md) supersedes earlier design recommendations; the [manual](manual.md) uses the current syntax.

# Community investigation, revision 2

Evidence collected on 2026-11-06. The [source manifest](research-sources.json)
records 39 source/test files from 17 project families, their upstream URLs and
SHA-256 content fingerprints. Branch URLs are mutable; fingerprints identify the
bytes inspected, not an immutable upstream commit. Large modules (notably Effect,
Zod and XState) were sampled around the relevant contracts. This is not a claim
to have audited whole libraries or run their upstream suites. Local downloaded
snapshots are in `/workspace/typefn-research`; they are not vendored in the fork.

## What the initial investigation missed

The first version selected algorithms that fit an existing small VM and then
used successful concrete examples as evidence for the language. That is backwards.
It missed generic API ergonomics, directional inference, optional metadata,
overload order, identity, parser failure semantics and representation costs.
Passing compiler regressions demonstrates that an extension did not break those
tests; it does not establish that the extension meets these requirements.

The design now has three separate obligations: express the algorithm clearly,
preserve its type contracts, and run it within measured resource limits. Each
library investigation below supplies adversarial cases for those obligations.

## Objects: merge, serialization, paths, brands

### type-fest MergeDeep and ts-toolbelt Merge

Inspected `MergeDeepRecordProperty`, `DoMergeDeepRecord`,
`MergeDeepIndexSignatures`, tuple/rest helpers and merge tests in type-fest;
object/list merge dispatch in ts-toolbelt. These are policy-driven transformations,
not recursively overwriting a map.

* A source property of `any` must be handled before an `undefined` test: ordinary
  assignability otherwise classifies it as only undefined.
* Optional absence and an explicit undefined value are different. A descriptor
  needs both the declared type and optional flag; a read type alone loses data.
* A shared string indexer does not guarantee that either source has a particular
  property. Its merged value must account for destination-only, source-only and
  combined values. Explicit fields and indexers need separate processing.
* Numeric and string property spellings can denote the same runtime key. Key
  identity cannot be implemented by displaying a type and using that text.
* Array replacement, concatenation, tuple pairing, optional elements and variadic
  tails are separate policies. A finite tuple loop cannot represent all of them.
* Source comments explain that combining mapped branches can exceed comparison
  depth. Fewer lines and a prettier displayed result do not imply cheaper types.

Required fixtures: optional versus required undefined, any, overlapping indexers,
numeric keys, readonly input, empty arrays, optional/rest tuples, union members,
and generic merge use inside a function. Do not label a simple object overwrite
implementation as MergeDeep compatibility.

### type-fest Jsonify and Tagged

Inspected `Jsonify`, `JsonifyList`, `ToJsonResult`, `WithoutToJson`, flavor removal,
and tag container definitions. Serialization is contextual: undefined, functions
and symbols disappear in object fields but become null in array positions.
Optional tuples can imply several possible lengths. The library distinguishes
calling `toJSON` on the original value from calling it again on its result.

An enum retains native identity even when its runtime serialization is a string.
Primitive brands and object brands cannot all be removed by the same structural
clone. Interface openness affects compatibility with index signatures. TypeScript
also cannot distinguish every runtime number, so no type DSL can recover NaN
information absent from the input type.

Design consequence: transformations need context parameters, opaque identity
handles and explicit loss policies. Reflection must not imply that all objects
can be reconstructed faithfully from public property names.

### React Hook Form Path

Inspected `PathImpl`, `PathInternal`, `Path`, `ArrayPathInternal`, `PathValueImpl`.
The inspected implementation has an explicit depth of 9, browser/native leaf
types, tuple-specific keys, and a mutual-type comparison for traversal history.
Its public distribution step deliberately processes union members independently
so that an assignable member does not contaminate another member's visited set.
`PathValueImpl` propagates undefined along optional paths.

Design consequence: recursion tracking belongs to an algorithm's state, not a
universal VM rule that silently truncates every object traversal. The VM's safety
budget must report failure; an algorithm's depth option can deliberately produce
a partial result. These have different contracts.

## Pattern matching: ts-pattern

Inspected `DeepExclude`, `DistributeMatchingUnions`, `FindUnions`, `Distribute`,
`FindSelected`, `InvertPattern`, and representative exhaustiveness tests.

Subtracting `["a", 1]` from `["a" | "b", 1 | 2]` requires revealing correlated
combinations. A plain `Exclude` on the enclosing tuple does not do it. ts-pattern
first discovers relevant union positions, distributes selected combinations, then
rebuilds the enclosing structure. The inspected search has a depth cutoff.

Arrays, maps and sets are not finite tuples: matching one element does not remove
the element type from every other position. Negative patterns can also describe
sets TypeScript cannot represent (`number` except the literal 10). Match selection
bindings have additional collision and aggregation rules. Inverting a pattern
for accepting input and inverting it for exhaustiveness subtraction are separate
operations.

Consequences:

1. String decomposition is a useful first pattern family, not a replacement for
   ts-pattern's structural analysis.
2. Union distribution must be explicit, correlation-preserving and budgeted.
3. An inability to represent a subtraction must not be reported as an empty set.
4. A native checker relation is not an all-purpose theorem prover. There are
   operations with a third result: not yet decidable or not representable.

## Strings and parsers

### ArkType

Inspected runtime/type-level parser pairs in `parser/string.ts`, the static
reduction state, and unenclosed operands. This is a particularly useful comparison:
the runtime code uses state and loops where the type implementation threads state
through conditional recursion. State includes roots, operators, groups, bounds,
finalizers and scope resolution. Errors and completion candidates are outputs of
the parser, not merely `never` at the first unrecognized character.

The language should make a cursor, local records/lists, lookahead, result/error
variants and full-input consumption ordinary operations. A template match alone
does not supply an ergonomic lexer or parser. Fixed literal delimiters are useful;
dynamic delimiters and character classes need deliberate APIs and ambiguity rules.

### gql.tada

Inspected tokenizer, parser, selection, variable and API types. Parsing threads an
unconsumed input through records such as `{out, in}`. Lists, fields, directives and
arguments reuse this state. Constant versus variable contexts differ. Whitespace
and comments are skipped; source comments document inlining choices for compiler
performance. Parsing and schema-directed selection are distinct phases.

Variable requiredness depends on both default values and GraphQL NON_NULL;
list nullability and element nullability are distinct. Fragment selection adds
schema lookup and composition. The type-level parser is purpose-built, not a
claim to preserve every property of a runtime GraphQL AST.

Design consequence: compile-time data should hold parser state and syntax trees;
native Type handles should enter at semantic analysis/construction. Encoding every
cursor integer as a tuple length merely relocates the original problem.

### Type challenges, TanStack Router and type-fest Words

The query-string challenge has 15 tests. Its contract preserves insertion order
of distinct duplicate values, deduplicates repeats, and represents bare keys as
true. It is not URLSearchParams: decoding, plus signs and malformed escapes are
outside those tests. Empty segments need an explicit extra policy.

The JSON challenge has 9 tests and deliberately rejects numbers (`[1]` is never).
It checks nested arrays/objects, booleans/null and escaped control characters.
Passing it must never be described as implementing standard JSON.

TanStack Router's `link.ts` uses required/optional/rest parameter sets, brace
forms, bracket escapes, dollar parameters and splats. A demonstration of
`/users/:id` does not cover that grammar. Route resolution also depends on the
source route and uses inference controls to keep one argument authoritative.

`Words` has acronym, number, case-change and punctuation policies, strips certain
brands, handles infinite string types and intentionally controls option-union
distribution. A two-capture split is not enough for case conversion parity.

The native Go checker consumes Unicode code points for adjacent template
inference holes; the JavaScript compiler historically consumes UTF-16 units.
The DSL currently follows this fork's native decomposition and tests astral
characters and lone surrogates. This difference must be documented, not hidden
under a claim of full JavaScript-TypeScript compatibility.

## Inference-heavy APIs

### Kysely and Drizzle

Inspected selection parsing and column roles in Kysely, sample selection tests,
and general/Postgres selection types in Drizzle. A column can have different
select, insert and update types. Selection aliases change keys while retaining
their value types. Dynamic references can produce optional selections. Kysely's
`DrainOuterGeneric` is evidence of representation and instantiation costs.

Drizzle carries join-nullability maps, selection mode, excluded fluent methods
and higher-kinded result types. An outer join may make an entire nested object
nullable; making each of its fields nullable is a different result. Generic
callbacks need access to symbolic table fields before final arguments are known.

Design consequence: symbolic object and signature construction is a core feature.
An opaque `F<T>` that only evaluates at concrete call sites breaks these APIs even
if every isolated concrete result is correct. Tests must include real generic
function bodies, contextual callbacks, rejected selections and chained inference.

### tRPC, Zod, TypeBox and Effect

Inspected tRPC parser adapters and procedure builder, Zod's `$ZodTypeInternals`,
TypeBox 0.34.41 static decoding, and selected Effect Schema contracts.

* tRPC distinguishes parser input and output across several schema ecosystems.
  A resolver receives parsed input; output validation consumes another input
  direction. Context overwrites and unset markers add state.
* Zod has separate input/output parameters and variance annotations. A schema is
  not simply the shape of its parsed result.
* TypeBox `Static` supplies parameters through an intersection and indexed access.
  Static decoding handles transform outputs, optional/readonly wrappers,
  references, recursive/module schemas and function/constructor types. Dispatch
  order (for example enum before union) is semantically significant.
* Effect separates Type, Encoded and context requirements and uses NoInfer in
  directional operations. Erasing a context parameter changes the API contract.

Design consequence: expose native type handles and preserve symbolic relationships.
Do not introduce an intrinsic named `infer` that guesses an inverse of arbitrary
imperative code. Native structural inference should see published symbolic
structure; deferred computations should remain deferred. Explicit pattern
inference can use the native solver with bounded, local inference variables.

### XState

Inspected event indexing, inference controls and correlated state/actor types in
`types.ts`. `IndexByProp` remaps discriminated members. `DoNotInfer` and `LowInfer`
show that inference priority is part of public usability, not an implementation
detail. Candidate collection and validation are different phases.

Design consequence: no observation of mutable, half-complete inference contexts
from the VM. An evaluation result must not depend on which checker query happened
first. Native NoInfer barriers must survive type construction and instantiation.

## Higher-order composition

Inspected hotscript `Fn`, `Apply`, `Call` and argument machinery, fp-ts HKT
registries, and ts-toolbelt Curry. Hotscript's `this`/argument encodings exist to
represent callable type-level functions. Named calls alone do not replace that
abstraction. fp-ts also relies on open module-augmented registries; freezing them
into a closed reflected object would change their meaning.

Curry handles gaps, required positions, labels and optional/rest parameters.
Some outputs are generic call signatures. The user's permission to skip creating
generic types places an explicit boundary on full Curry parity; it does not
justify silently dropping polymorphism. Existing generic signatures should be
preserved opaquely. Higher-order DSL functions are a separate feature from
constructing native TypeScript generic declarations.

## Revised priorities and evidence gates

| Requirement | Why needed | Evidence required before claiming support |
| --- | --- | --- |
| Symbolic construction | SQL/schema/fluent APIs | Generic property access, calls, forward inference and delayed constraints |
| String matching | Routes and parser primitives | Scope, anchoring, empty captures, Unicode, unions, malformed patterns, emitted declarations |
| Compile-time data | Parsers, arithmetic, descriptors | Exact arithmetic, cursor progress, local mutation, no type/data coercion |
| Reflection/reification | Merge, paths, schema transforms | Optional/readonly/indexer/signature round trips with exact-optional enabled |
| Explicit union algorithms | Correlation and exhaustiveness | Nested products, any/never/unknown, normalization and bounded explosion |
| Directed inference | Schema and fluent APIs | NoInfer preservation, callback context, inference order independence |
| Higher-order composition | hotscript and HKT | Arity/kind errors, closures, stable cache identity, open native registries |
| Robust integration | Personal compiler and LS use | Imports, .d.ts round trips, rename/hover, cancellation, edit invalidation, malformed input |
| Performance | Large real programs | Cold/warm time, allocation, cache size, adversarial resource failures, native baselines |

The archived VM is an experimental implementation on `typefn-language`; the
independent `typefn-redesign` branch contains the new specification and no DSL
implementation. Completed tests must be recorded separately from designed or
investigated capabilities. Neither a source count nor a compiler-regression pass
is a maturity score. The acceptance corpus must pair each transformation with
positive, negative, symbolic and resource cases.


## R1 follow-up: native higher-kinded constructors

Revisited fp-ts HKT and hotscript Core, inspected Effect's complete
`packages/effect/src/HKT.ts`, and examined the pinned native checker's
`getTypeFromTypeAliasReference`/`getTypeAliasInstantiation` argument and cache paths.
The new Effect snapshot is recorded in the manifest. No upstream suites were run.

fp-ts uses URI registries and arity-specific Kind aliases; hotscript uses callable
interfaces and this/rawArgs; Effect uses TypeLambda slots with distinct
input/output/target relationships. All demonstrate the need to pass and apply
constructors. None establishes that arbitrary constructor inference is possible.
The native checker instantiates known declarations using Type argument vectors;
that is useful infrastructure but is not already first-class constructor support.

R1 therefore proposes explicit constructor parameters (`F<_>`), constructor
closures and symbolic applications in both the DSL and native generic declarations.
It preserves domain constraints and inference barriers, and does not assume
injectivity or covariance. The [higher-kinded specification](higher-kinded-types.md)
records those rules and the required representation changes. The simplified
[current guide](README.md) supersedes R0's larger surface; historical syntax and
workload sketches are kept in `archive/r0`.

## TypeScript's missing HKTs: constraint audit

The [dedicated audit](typescript-hkt-audit.md) reviews #1213 and #55280, official
higher-order inference and instantiation-expression PRs, the maintainer distinction
between HKTs and call types, the native inference FAQ and the Go checker's mapper,
inference and variance paths. It distinguishes historical maintainer interest from
community proposals; no definitive impossibility/rejection rationale was found in
the reviewed evidence. The source scope and incomplete comment retrieval are
recorded there rather than counted as additional library source files.

Reproducible [stock TypeScript probes](probes/hkt-baseline.ts) demonstrate alias
structure versus preserved inference evidence, non-injectivity, non-covariance,
input/output constraints, generic callable versus constructor application, overload
extraction, defaults and native conditional distribution. The runner checks 11
observations and six diagnostic sites on installed TypeScript 6.0.3.

This audit corrected R1's default-dependent implicit constructor lifting and
overstated constructor inference, and identifies output contracts and native
constructor-aware relations/variance as unfinished design. It adds no DSL compiler
implementation and makes no performance or upstream-suite claims.

## Restricted HKT follow-up

The [restricted proposal](restricted-hkt.md) responds to the audit by keeping
constructor polymorphism inside typefn, admitting only fixed-arity constructors
from ordinary Types to an ordinary Type, and requiring explicit constructor
arguments. Bounds are validated at each application, rather than comparing entire
constructor domains. Named combinators and returned closures retain composition
and partial application. Constructors do not enter native generic argument slots.

This removes output contracts, native constructor variance and constructor
inference from the immediate requirements. Symbolic ordinary Types, deferred
Type-valued computations and core checker integration remain. The proposal does
not replace native runtime-polymorphic HKT APIs such as Functor<F>; existing
encodings stay usable. The earlier broad HKT document remains an exploration.
No additional external investigation, library port or implementation test is
claimed for this design-only refinement.

## Symbolic execution and declaration contract follow-up

Inspected the pinned native checker's instantiateTypeWorker, union normalization
and type-argument constraint paths, and located the checker nodebuilder and
declaration transformer/tracker. Native union normalization can immediately return
any/unknown; the native substitution switch also assumes its known Type categories.
These provide concrete places where silently dropping deferred validation or
unrecognized Type terms would break specialization. This is an integration map,
not an exhaustive implementation audit or a new upstream-suite run.

The [evaluation contract](evaluation.md) makes specialization consistency the
governing rule. It specifies known-shape publication, pending checks independent
of local slot liveness, guarded native simplification, replay with frozen captures,
and conservative generic-body/signature requirements. The [declaration contract](declaration-emit.md)
preserves executable recipes and validation instead of serializing current shapes
or VM/native pointers. [Semantic cases](semantic-cases.md) cover the decisions.
Neither a written contract nor these design oracles establish a working compiler.

## R2 design reconciliation

The [R2 specification](spec/README.md) adopts the parser-first, single type/typefn
model. This refinement adds no new external source investigation or library-port
results. Earlier recommendations for separate concrete data and native Type handles
are superseded at the source level: literal Types, tuples and object Types now
represent scalar/parser/descriptor state, with permitted internal unboxing.
Performance of that representation still requires measurement.

R2 closes the helper-sharing question with ordinary exported typefns; defines
angle-call disambiguation and explicit native Type islands; keeps the agreed
restricted constructor scope; specifies Type-valued reflection/builders; and
updates guarded evaluation and declaration ABI to that source model. String
iteration/pattern matching uses explicit UTF-16 rules, independent of the pinned
Go checker's native template-inference behavior. Native islands retain native
semantics. Enum-specific APIs, including structural builders and reflection, are
deferred following the user's scope reduction. Algorithmic generic-binder
construction remains excluded; existing native signatures/identities can still
be preserved.

The [conformance matrix](spec/conformance.md) traces the investigated families to
facilities, limitations and required evidence. A formal design baseline is not
proof of parser conformance, runtime performance or full community compatibility.
