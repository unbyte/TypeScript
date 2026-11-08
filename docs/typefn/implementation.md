# Implementation status

The `feat/typefn` branch has an **experimental implementation** of
language version 1. The R2 specification remains the design baseline; this is not
a claim of R2 conformance or production maturity. Native editor integration is
implemented for this slice; see [editor integration](editor-integration.md).
See [architecture](architecture.md) for ownership and integration boundaries,
[deferred evaluation](deferred-evaluation.md) for substitution and retained checks,
and [design decisions](design-notes.md) for unresolved constraint semantics.
`where` is not implemented syntax.
The examples in the manual describe the target language. Use the integration
fixtures to see what the compiler currently executes.

## Try the implemented slice

Build the fork with `go build -o /tmp/tsc-typefn ./tsc/cmd/tsc`, then run
`/tmp/tsc-typefn --project path/to/tsconfig.json`. Enable the DSL in `tsconfig.json`:

```json
{
  "compilerOptions": {
    "strict": true,
    "typefn": { "version": 1 }
  }
}
```

```ts
export typefn Reverse<T> {
  type result = [];
  for (type item of T) { result = [item, ...result]; }
  return result;
}

typefn Range<N> {
  type result = [];
  type i = 0;
  while (i < N) {
    result = [...result, i];
    i = i + 1;
  }
  return result;
}

type Result = Reverse<Range<4>>; // [3, 2, 1, 0]
```

Native Type assertions, assignments, type-only imports and re-exports consume the
results through the actual checker. JavaScript emit erases the declarations.
Declaration emit currently fails explicitly: source-recipe serialization must be
implemented before executable definitions can be published safely in `.d.ts` files.

## Implemented architecture

* The opt-in profile and parse-affecting limits participate in source parse options.
  Version and limit settings are validated by the native tsconfig parser.
* Dedicated generated AST nodes represent DSL syntax. A separate recursive-descent
  statement parser and Pratt expression parser handle bodies. Angle-list recognition
  uses bounded, memoized speculation. No body falls back to JavaScript parsing.
* Native binding registers declarations, Type parameters, mutable locals and iteration
  bindings in native lexical scopes. Editor queries and VM lowering share symbol
  identities. A checker-owned compiler resolves DSL slots and names, validates all branches,
  and lowers control flow into store, return, branch, jump and iterator instructions.
  Expression trees have resolved slots/symbols and are evaluated by the same VM.
* Registers hold immutable native Types or helper closures. Assignment changes a slot, and
  constructed Types retain their previous contents. Named calls share root budgets.
  Native generic applications use native instantiation and constraint checking.
* Compiled programs belong to one checker instance. No cross-checker Type handles,
  fabricated deferred type parameters, or global VM caches are used. Native checker
  node caches still apply. Pending terms are interned per checker; completed VM
  calls have no general result memoization layer yet.
  A lazily allocated `typefnState` owns these tables and the active root VM.
  Iterator slots are assigned while lowering and stored in each call frame; their
  storage participates in the live-frame estimate. Native-only operations avoid
  allocating DSL state, and disabled-profile checking skips DSL declaration scans.
* Dedicated native deferred Types retain immutable operands and captured locals.
  Native TypeMapper substitution replays calculations when arguments change.
  Per-call journals retain reached checks even when a result is discarded or
  absorbed by the supported native simplifications. See the
  [implemented boundary](deferred-evaluation.md) for details and limitations.
* The source printer preserves the implemented syntax. Round-trip tests include
  zero-argument applications, angle arguments containing predicates, mandatory
  semicolons, objects, tuples and native-island syntax.

## Supported execution

Module-level named functions with ordinary Type parameters; mutable `type` locals;
lexical block scopes; return, if/else, while, break/continue; required finite tuple and UTF-16 string
iteration; dependency-aware union iteration; tuple construction and finite
tuple spreads; object construction with string/number/unique-symbol keys and optional/readonly
properties; native indexing and `keyof`; native union/intersection construction;
number/bigint arithmetic and comparisons; literal boolean short-circuit operators;
statement-only `fail`/`failAt` with deferred messages and validated UTF-16 offsets;
native identity/assignability predicates with dependency summaries; calls to other named typefns and
ordinary native aliases/interfaces/classes, including native defaults and bounds.

Definitions reject missing names, wrong arity, duplicate/reserved bindings,
reads before local initialization, writes to immutable bindings and misplaced loop
control, including in unreachable branches. Runtime checks reject invalid scalar
families, division/remainder by zero, nonfinite arithmetic and unsupported tuple
shapes. Property keys use decoded names and native numeric normalization; equivalent
duplicate keys and nonfinite keys diagnose. Native recovery Types fail the concrete
calculation even when an intermediate value is discarded. Pending operations retain
checks for later specialization. Identical active applications diagnose cycles; changing-argument recursion
is bounded by call depth. Runtime fallthrough diagnoses.

The complete text library and string iteration preserve UTF-16/lone-surrogate
semantics under a shared string-work budget. Searches use linear prefix-table
matching, and split/replacement reserve output storage. The numeric library covers
explicit exact conversions, rounding, parsing/checks and formatting. Finite tuple
length/slicing preserve labels and readonly. Interpolation builds native template
Types and retains eligibility/resource checks; matching includes captures, guards,
computed pins, tuples/objects and opaque template suffixes. Native-island patterns now execute through the same lexical bridge.

## Important remaining work

* Native islands execute in checker-owned, freshly bound capture scopes. Symbolic
  DSL captures conservatively retain the whole island for replay. Publishing stable
  symbolic native shapes, reusing equivalent environments and complete native work
  accounting still require work.
* The calculator, numeric and HTTP-route examples execute directly from their
  documentation sources. They are milestones, not full conformance coverage.
* Local named Type-returning helpers now execute with frozen captures. Anonymous
  closures, first-class constructor bindings/slots/results and restricted HKT
  composition remain pending.
* Deferred scalar expressions, calls, control flow, captured construction, and
  native substitution/replay are implemented. Full requirement-bearing native
  relation proofs, scoped generic-body assumptions, latent member/signature
  dependencies, and exhaustive inference/replay laws remain pending. Guarded
  relations currently require the same canonical check set. Structural predicates
  and union traversal use native dependency summaries; exhaustive native forms
  and ordinary cast/assignment boundaries still need coverage. No constraint is
  enumerated as a value.
* Exhaustive tuple/property provenance,
  recursive structural sealing
  remain. Classification, arrays, symbol-key construction and property descriptors
  are implemented. Enum-specific support stays deferred.
* Declaration source recipes, advanced refactorings/auto-import completion, complete
  incremental dependency/replay tests and broad API factory compatibility remain pending.
  Native symbols now support definition/references/rename, hover, scoped completion,
  angle-call signature help, semantic coloring, result hints, outline/folding and
  formatting. Real editor-session tests cover imported-definition edits, cross-file
  navigation/rename, malformed producers, cancellation and profile toggling.
  The generated TypeScript API is type-checked. Runtime tests cover DSL serialization,
  transforming visitors and preservation of expression/statement kinds by updates.
* Resource accounting is preliminary: instruction dispatch, DSL call depth, explicit
  tuple/property construction, frame allocation and arithmetic bigint magnitude are
  checked. This is **not** complete accounting for all live buffers, native graph
  visits, native allocations, literal parsing or retained compilation caches.
  Explicit `nativeWork` or `cacheBytes` overrides diagnose on DSL
  execution until their accounting is implemented. Full native-operation
  cancellation and budget instrumentation remain required before calling this a hardened service boundary.
* Definition checking validates conservative return paths and still needs the complete
  kind/readiness analysis. There is no claim of optimized append or memoized VM calls.
  Parameter bounds currently use native Type grammar rather than full DSL expressions.
  Module-level bound checking can evaluate header applications at definition time,
  and an earlier blocked bound delays later checks. Module bounds accept later-parameter
  references even
  though the R2 text describes preceding-parameter scope. These behaviors need an
  explicit contract decision; see the [constraint decisions and probes](design-notes.md#input-constraints).
  Result-kind annotations are retained in the native AST and printer.
  Constructor execution, kind inference and recursive annotation validation remain pending.

## Validation

The compiler-project harness has 235 fixtures covering native calibration, execution,
definition errors, deferred checks, resource limits, imports and exact output.
Successful `Assert<...>` uses must resolve to literal `true`; negative fixtures check
complete diagnostic sets. The specification's full 141-case target matrix and broad
community workload ports remain outstanding. See [testing](testing.md) for commands,
coverage and the limits of the validation evidence.

### Pattern matching

Native AST, parser, printer and public API nodes preserve match arms, guards, all
pattern forms, and template capture/pin boundaries. Captures bind in native arm
scopes; pins and pattern Type expressions use the outer environment, and the body
has its own nested block scope. Definition, rename, completion and hover tests
cover those boundaries.

Execution now supports ordered arms and guards, literals, scalar `is` patterns,
immutable whole-value captures, required finite tuples with metadata-preserving
rest captures, and required public object properties. An undecidable earlier arm
blocks the call and replays at specialization. Arm-private registers cannot leak
captures to a later arm; reached checks survive a failed guard. Tests cover
continuation after an arm, loop control, symbolic specialization, private/optional
properties, static wildcard/key/capture errors and union-domain rejection.
Template patterns now use bounded UTF-16 decomposition, earliest delimiters without
backtracking, adjacent-unit captures and computed nonempty literal pins. Pins use
outer bindings and are validated before literal chunks are attempted. Native
opaque template holes remain opaque; final captures preserve a remaining symbolic
suffix. Free binders that must be inspected block and replay at specialization.
Tests cover header splitting, surrogate pairs, opaque holes, pin scope/errors,
anchoring and resource exhaustion. Native-island patterns remain pending. Structural predicates and union iteration
now inspect native argument/lexical-dependency summaries, retaining generic replay
without eagerly expanding recursive native properties. Tests cover recursive
interfaces, functions with local generic binders, inferred generic results, mapped
objects and failures reached through latent properties. Object patterns wait for
unknown mapped keys and exclude implicit Object/Function prototype augmentation.
Exhaustive native dependency and boundary conformance remains in the acceptance
tracker; these cases do not complete the full generic-evaluation contract.

### Classification

`reflect.kind`, `reflect.isAny`, `reflect.isNever` and `reflect.isUnknown` are
implemented. Classification distinguishes whole boolean before its native union,
and tuples before arrays. Symbolic object fields/array elements can retain a known
category; potentially collapsing unions and shape-changing mapped/variadic Types
wait. Tests cover all specified category tags, generic normalization, enum rejection
and the unsupported closed-category diagnostic (including infinite native templates,
for which the current category vocabulary has no tag). Prelude editor metadata is
shared with the intrinsic registry.

`types.array<Element, Readonly = false>` and `reflect.array<Array>` now build and
inspect native mutable/readonly arrays. Their descriptor is an ordinary native
record; symbolic element Types are preserved without blocking known array shape.
Tests prove generic length/element facts before specialization, round-trip readonly
arrays, replay symbolic flags, and reject tuples and numeric-indexer impostors.

Computed object keys accept unique-symbol Types and preserve the native key handle
through generic object instantiation. Tests use imported symbols, `keyof`, indexing,
optional/readonly fields, snapshot assignment and duplicate/broad-symbol rejection.

### Descriptor object construction

`types.object<Properties>` now constructs property-only structural objects from
required finite descriptor tuples. It validates required/public metadata, unknown
fields, native key equivalence, duplicate keys and literal optional/readonly flags;
`read` is accepted and ignored. Field values may remain symbolic. Tests cover
intersected descriptors, unique-symbol keys, generic key/flag replay, and both
exact-optional modes. Index descriptors now retain native key domains, value Types, readonly flags and
order. Validation follows native index applicability and assignability, including
numeric/string domains and template patterns. Symbolic compatibility requirements
remain attached to the published object shape and replay even after absorption.
`reflect.indexers` returns ordered descriptors, and generic object instantiation
retains indexers. Tests cover overlapping domains, invalid keys, duplicate domains,
readonly native writes, reflection, generic replay and retained failures.

Native signatures can now be forwarded through `reflect.calls`,
`reflect.constructs`, `types.overload`, and the object builder's call/construct
lists. The adapter retains original signature pointers/order, generic binders,
method variance and abstract flags. Queries expose public signatures;
`reflect.isGeneric` requires exactly one signature. TypeMapper replay follows
native signature provenance without evaluating latent generic return computations.
Tests exercise actual overload calls, generic inference, outer-Type specialization,
generic default constructors, method/function variance, abstract/private
constructors and category errors.

`types.function` and `types.constructor` now build ordinary native signatures
from parameter tuples, preserving native labels, optional/rest/variadic tails and
readonly input tuples. Native tuple-rest expansion controls call inference and
arity; fresh signatures have ordinary function variance. Generic parameter/result
leaves and the abstract flag replay independently. Tests include actual calls/new,
Parameters/ConstructorParameters, outer-Type specialization, native primitive
constructor results, and negative arity/variance/abstract cases.

`reflect.signature` now returns the nongeneric descriptor schema. It preserves
original parameter boundaries (including whole rest-tuple Types), names,
optional/rest modes, explicit this, returns, abstractness and all four native
predicate forms. Generic signatures diagnose rather than erasing their binders.
Tests cover symbolic outer-Type replay, initialized/optional/void parameters,
construct signatures, parameter/this predicates, assertions and generic rejection.
Parameter values use native parameter Types, including optional-read undefined;
the separate property/tuple declared-value contract remains unfinished.

`types.signature` now validates and builds the nongeneric descriptor schema.
Required/optional/rest ordering, identifier names, duplicate names, array-compatible
rest Types, call/construct fields, predicate targets and narrowed-Type compatibility
are checked. Predicate results use the host's boolean (`is`) or void (`asserts`)
Type. Native this-predicate rules are preserved. Symbolic compatibility checks
remain attached to the constructed callable shape and survive absorption.
Executable tests cover native call/new inference, this receivers, actual control-flow
narrowing/assertions, ordinary variance after rebuilding a method, deferred rest and
predicate validation, and invalid descriptor cases.

Native islands bind a cloned native subtree beneath immutable capture aliases,
using native symbol binding, checking and Type construction. Original source AST
caches are never rebound to invocation-local values. Inner native binders shadow
captures by symbol identity; captured nominal Types retain their original handles.
Native diagnostics are attributed to each application, with their operation spans
as related information. Tests cover mutable snapshots, mapped Types, conditional
infer, generic signatures, typeof, nominal identity, native patterns and symbolic
replay. Symbolic capture shape publication and environment caching remain pending.

Native-island reference names and generic arities are checked during lowering,
including unexecuted branches. Captured Types cannot be applied as constructors;
native typefn references require explicit applications. Tests cover missing value
names in typeof, missing native generic arguments and unapplied typefn names.

Tuple expression spreads use native variadic normalization. They preserve source
labels/optional/rest modes, support arrays and native union expansion, and recheck
unknown spread shapes after substitution. Tests compare results with ordinary
native tuple aliases and check multiple-rest/domain errors, including a deferred
failure whose value is absorbed by unknown. Tuple descriptor APIs are implemented below; exhaustive native-producer provenance
remains pending.

`reflect.properties` now returns ordered public-property descriptors, including
inherited fields, normalized numeric keys and unique-symbol keys. Declared optional
values remain distinct from read Types in both exact-optional modes. VM-created
optional fields retain immutable declared-value metadata across substitution;
unchanged instantiations reuse their property symbols. Native declarations,
identity mapped projections/key remapping, intersections and spread provenance
supply native declared values. Accessors with differing read/write Types diagnose.
Tests cover actual readonly projections, inherited/private/protected members,
optional methods/initializers, Pick/Partial/Readonly, explicit undefined, never,
generic replay and metadata-only changes. Exotic native optional-member provenance
(e.g. reverse mappings) still needs exhaustive coverage; unavailable provenance
diagnoses rather than silently stripping undefined.

Native tuple normalization now has an opt-in declared-value metadata channel.
Source annotations and variadic spreads retain optional declared values separately
from native read arguments. Immutable metadata views share native tuple targets and
semantic arguments, are interned per checker and participate in TypeMapper replay.
Unit tests cover distinct declared values with equal read Types, snapshot isolation,
reuse, spreads and optional-to-required normalization. Untracked native producers
retain an explicit unavailable-provenance marker until their producer adapters
are implemented and tested.

`reflect.tuple` and `types.tuple` now implement the descriptor schema for known
native tuple shapes. They preserve declared required/optional values, labels,
readonly, rest tails and variadic tuple expansion. Builders validate native ordering
and array/tuple domains; unknown shape/metadata retains the operation for replay.
Generated label declarations are reused per source operation/name/mode. Tests run
in both exact-optional modes and cover native round trips, metadata after spreads,
symbolic value/shape replay, optional never/undefined and invalid descriptors.
Native homomorphic tuple mappings now retain declared values through identity
projections (Partial/Readonly) and compute other templates before optional-read
wrapping. Tests cover both exact-optional modes, transformed templates, explicit
undefined, labels and generic replay. Native tuple slicing, inference slices,
readonly-to-mutable copies, normalization, widening and literal clones carry provenance
without changing native element read Types. Fresh native tuple inference records
its computed values before optional wrapping, removing only the internal missing
marker. Parameter tuple synthesis preserves originating rest-tuple declarations;
ordinary optional parameters/defaults retain their native undefined acceptance.
Tests cover Parameters/ConstructorParameters, prefixed rest tuples, generic replay,
reverse mappings, destructuring defaults and exact-optional array holes. Exhaustive
producer combinations and cross-declaration serialization remain acceptance work.

Public property, indexer and signature enumeration accepts the native `object`
keyword and returns its empty public surface, consistent with reflect.kind.
This does not treat a public structural projection as identity reconstruction.

Definition checking now rejects a reachable end of a typefn, including unused
definitions. It traverses lowered control flow without evaluating operands: loops
may execute zero times, both branch outcomes remain possible, and return/explicit
failure terminate paths. Final wildcard and whole-value capture patterns always
match. Return-kind inference and recursive-component annotation validation remain
separate pending work.

Result-kind annotations now survive parsing, printing, native traversal and the
public AST protocol. `: type` and `: typefn<...>` have explicit kind nodes; an empty
constructor slot list differs from a missing annotation. Constructor-result syntax
currently receives an explicit definition diagnostic until constructor values and
kind checking land. It does not silently execute as an ordinary Type function.

Local named helpers now lower into separate programs with immutable capture
snapshots. Captures may include earlier helpers; self-reference uses the current
closure without a mutable cycle. Calls and deferred expression snapshots retain
closure identities and substitute captured Types when native generic arguments
change. Constructor edges are traversed iteratively. Closure snapshots are interned
per checker, and new snapshot storage is charged to the current root budget.

Tests cover reassignment after capture, nested helpers, shared code with different
captures, self-recursion, iteration captures, shadowing, constraints, native Type
captures, generic replay, discarded failures, distinct importing callers and JS
erasure. Forward local-helper references, capture writes, invalid control flow and
wrong arity diagnose at definition time. Local-helper bounds may refer to preceding
Type parameters and frozen outer Types; they use immutable native-island evaluation
and conservatively wait for symbolic captures. Module-level bound timing/scope
remains separately tracked above.

Editor tests cover native helper/capture definitions and rename, readonly capture
hover, and declaration-ordered helper completion with self-reference. First-class
helper values, anonymous closures, constructor slots/results and constructor
captures in native islands still diagnose explicitly until their integration lands.
