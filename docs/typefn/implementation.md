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
* Local named helpers and anonymous Type-returning closures now execute with frozen captures.
  Factories can return constructors inside DSL computations. Mixed Type/helper-constructor
  arguments execute in DSL and native calls, including native generic constructors.
  Nongeneric native constructor aliases and constructor captures in native
  islands execute. Full outer-native lexical dependencies still require work.
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
* Definition checking validates conservative return paths, finite value kinds and
  recursive result-kind dependencies. Module functions, local helpers and lambdas
  use the same DSL bound-expression evaluator. Bounds run at application time,
  in parameter order, and may reference preceding parameters and frozen outer
  bindings. A blocked proof suspends the call before checking later bounds.
  Native generic declarations retain native constraint semantics. Constructor
  parameters carry separate VM values through direct calls, closures and replay.
  Full outer-native lexical dependency capture
  remains incomplete. There is no claim of optimized append or
  memoized VM calls.

## Validation

The compiler-project harness has 303 fixtures covering native calibration, execution,
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
anchoring and resource exhaustion. Native-island patterns use the lexical bridge.
Structural predicates and union iteration
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
match. Finite result-kind inference and recursive-component validation are described below.

Result-kind annotations now survive parsing, printing, native traversal and the
public AST protocol. `: type` and `: typefn<...>` have explicit kind nodes; an empty
constructor slot list differs from a missing annotation. Factories may infer a constructor result kind or declare it explicitly. Constructor
results remain separate from ordinary native Type results.

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
and conservatively wait for symbolic captures. Module functions now use the same bound-expression and scope rules.

Editor tests cover native helper/capture definitions and rename, readonly capture
hover, and declaration-ordered helper completion with self-reference. Native islands now preserve constructor captures and direct named helper signatures.

Eligible named helpers can now be bound as constructor values (`type F = Box;`)
and applied through aliases. Constructor bindings may be reassigned only to the
same arity; Type/constructor kind changes and constructor operands in native Type
positions diagnose during definition checking, including dead branches. Captures
freeze the selected constructor, and generic substitution maps captured Types.
Module helper aliases retain declared input constraints and share deferred-call
identity with direct calls. Native generic names still require explicit constructor
context; a bare defaulted native alias retains its ordinary Type meaning.

Anonymous `typefn<T> { ... }` expressions now capture the lexical environment at
expression evaluation and produce fixed-arity Type-returning constructors. Their
native AST node owns its parameter/body scope without a synthetic declaration
name. Tests cover nested closures/helpers, shadowing, captured bounds, loop
snapshots, reassignment, native islands, blocked calls and expression replay.
The parser/printer, binder, AST serializer/transformer and editor capture scope
share this representation. An anonymous initializer cannot recursively refer to
its own still-uninitialized binding; named local helpers supply explicit self
recursion. Named combinators may accept constructor parameters; anonymous closures
remain Type-to-Type functions.

Value kinds are now checked separately from helper code identity. A checker-local
constraint graph infers Type results without execution, verifies all return kinds
(including unreachable returns), and resolves mutually referring functions before
checking their callers. Unresolved recursive result kinds and named bodies with
no value-returning path require annotations. Anonymous closures supply their
Type-returning constructor context. Declaration-order tests prevent callers from
silently repairing ambiguous callee signatures. Local-binding hover now distinguishes
`type` from `typefn<...>` using static kind information, without running a body.

Named factories now return constructors within DSL computations, including named
helpers, anonymous closures, captured local helpers and recursive factory calls.
Bindings retain inferred arity, and factories cannot themselves become first-class
constructors because their result kind is not Type. A constructor may capture
symbolic Types, but unresolved selection, bounds or creation checks block its
publication and defer the enclosing Type-producing call. No constructor enters a
native Type operand list. Nongeneric native constructor aliases now preserve this
readiness boundary.

Stored and passed Type results retain their reached checks as well as recording
them in the caller journal. This matters when a ready native shape with pending
checks is later passed to a factory: constructor creation must still wait. Tests
cover checks through local bindings, ignored factory results, input bounds,
recursive closures returned after their creation frame ends, two-stage native
substitution and distinct imported consumer/operation diagnostics. Factory and
constructor-binding hover use inferred static kinds without evaluating bodies.

Typefn parameters have dedicated native AST nodes containing an optional DSL
bound expression or constructor slot list. Native TypeScript type-parameter nodes
and their eager constraint caches are unchanged. Computed bounds now include
arithmetic, text operations, helper calls and native Type islands. Unused bounds
are bound and kind-checked but never executed. Later/self references diagnose even
when a same-named outer Type exists. Native constraint diagnostics are attributed
to the importing application and include the bound's source location. Editor
services preserve parameter identity, readonly status, constructor slot coloring,
kind hover and signature help.

Mixed calls now carry Type values and fixed-arity helper constructors without
encoding a constructor as a native Type. Direct named combinators accept either
kind and may return constructors; only Type-to-Type functions are first-class.
The native argument adapter uses declared parameter kinds, so `Apply<Box, T>`
and `Apply<Compose<Box, Box>, T>` are executable. A native constructor argument
must have a ready identity and completed creation checks; closures may capture
unresolved ordinary Types. Deferred calls retain mixed arguments and replay their
captured environments. Constraints still run at each actual application, rather
than being compared universally when passing a constructor. Editor queries omit
Type results for factories and retain ordinary symbol navigation.

Native aliases, interfaces and classes can now supply constructor identities in
explicit constructor slots, constructor-binding reassignment and annotated
constructor returns. Native names retain their original binders and lexical Type
captures. Application checks all declared slots (including defaulted slots), maps
dependent bounds using the full argument/capture environment, and constructs native
references or instantiates native alias bodies. Original private class identity,
mapped aliases and native intrinsics are retained. Ordinary bare defaulted aliases
still produce their default Type; bare nondefaulted generics still diagnose outside
constructor context. Tests include composition, tuple mapping, Pick bounds,
constructor-bound replay, captured native generic bounds, default arity, private
classes, intrinsics and editor queries/display. Nongeneric aliases that name
constructor *values*, such as `type AsyncArray = Compose<Promise, Array>`, use
the separate alias bridge described below.

Nongeneric native aliases may now name known constructor values. Their finite
kind and ready value are cached separately from native declared Types. Aliases
can be imported, chained, passed to constructor slots and applied in native or
DSL syntax. Factory selection and creation checks must finish at this native
boundary, while known closures may capture symbolic ordinary Types. Generic
native aliases cannot return constructors. Definition checking publishes known
alias signatures before compiling recursive helper bodies, and creation cycles
diagnose. A bare constructor remains invalid inside native Type operands.
Editor hover, completion, signature help and result hints preserve alias kind and
identity without creating a recovery Type. Tests cover the documented AsyncArray
composition, lexical captures, forward chains, imports, replay, bounds, native
island applications, invalid Type operands and creation readiness.

Native islands capture constructor values in immutable synthetic alias scopes,
separately from captured native Type aliases. A named local combinator retains
its full mixed-parameter signature for direct calls; it does not become an
eligible first-class constructor. Static island checking follows constructor
argument contexts, validates arity/kinds in dead code, and honors native mapped,
infer and signature binder shadowing. Captures of constructors with symbolic
Type environments conservatively defer the whole island and replay its frozen
snapshot. Native alias/module/factory applications work inside islands, including
parenthesized constructor arguments. Tests cover local factories/combinators,
constraints, self calls, snapshot reassignment, generic replay, dead-code errors
and native editor navigation before/after execution. Shadowed native binders do
not create phantom transitive DSL captures.

Typefns nested in native generic scopes now retain native lexical parameters as
implicit immutable environment inputs. A blocked zero-argument call therefore
has substitution dependencies even when it produced no Type-valued operand before
blocking. Native mappers specialize those environments before bounds/body execution;
sibling calls and returned closures inherit the captured binder identities. Direct
native applications share the native-constructor adapter for mapped outer bounds
and defaults. These cases cover DSL expressions and direct constructor applications;
ambient native island references and value queries still require their own bridge.

Native source evaluation deliberately clears the incidental VM invocation mapper.
Native alias declared-Type and constructor caches therefore retain their defining
environment; DSL consumers obtain a mapped immutable closure view. Query-order
regressions force specialization before native source declaration checks and
verify that two different outer substitutions cannot contaminate each other.
Tests also cover nested native generic functions, class methods and actual bound
failures after specialization. Native mapper backing storage is charged with the
invocation frame, and native constructor snapshots charge their temporary storage.

Deferred expression snapshots also retain implicit native environment slots, not
just registers named by explicit operands. Expression replay rebuilds its mapper
from that frozen snapshot using the same layout as call replay. Tests cover text
and tuple intrinsics, computed object keys, and a local value changed after the
original expression was captured. Mapper storage is charged before allocation.
