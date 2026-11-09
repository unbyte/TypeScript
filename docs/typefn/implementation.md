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
Declaration emit preserves written typefn bodies, constructor aliases and their
lexical dependencies with ABI 1. Inferred escaping recipes, full dependency-failure
handling and declaration source maps still need further implementation and validation.

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
  The [community workload record](workloads.md) additionally covers executed
  type-fest word scanning and React Hook Form field-path traversal, including
  native comparisons, declaration consumers and resource failures.
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
  enumerated as a value. Native base-constraint resolution treats opaque deferred
  computations as unconstrained, preventing contextual literal checking from
  recursively following the same term. A guard exposes its known shape for
  contextual typing while retaining its validation journal for relations and
  specialization; this does not reverse-infer opaque computations.
* Exhaustive tuple/property provenance and complete recursive builder adapters
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
  checked. Bigint operand conversion shares a decimal-length preflight and exact
  magnitude check across literals, arithmetic, formatting and numeric conversion,
  including values supplied by native aliases/parameters. Arithmetic results now
  use the same native-node allocation check as numeric intrinsics. This is **not**
  complete accounting for all live buffers, native graph
  visits, native allocations, literal parsing or retained compilation caches.
  Explicit `nativeWork` or `cacheBytes` overrides diagnose on DSL
  execution until their accounting is implemented. Native relation entry/recursive
  steps and property comparisons now poll cancellation within the VM operation,
  suppress interrupted relation-cache publication and invalidate the canceled
  query snapshot. Substitution and conditional-type evaluation also poll internally;
  interrupted substitution unwinds its mapper stack and skips active-mapper and
  conditional-instantiation cache publication. Union/intersection member collection
  and union sorting/deduplication also poll; incomplete private sets are discarded
  before normalizer cache publication. This covers these traversals, not every
  native adapter or bigint primitive. Full native-operation
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

The compiler-project harness has 341 fixtures covering native calibration, execution,
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
ambient native island references and value queries use the captured-scope bridge
described below.

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

Native islands also freeze ambient native parameters, scoped aliases with defaults,
scoped typefns, and value roots of `typeof` queries. Native generic constraints are
checked with their original binders and mapped outer environments. Type and value
captures have separate namespaces even when their spelling is shared. Native
signature-local binders remain local. Source declaration caches stay unspecialized;
query-order tests cover aliases and value annotations first forced during a
specialized call. Hover preserves source meaning and generic Types. Polymorphic
`this`, stable symbolic shape publication and complete latent dependencies remain.

Recursive-builder integration starts with a separate arena validator. Guarded
edges are removed and the remaining reachable graph must be acyclic; merely
finding a guarded edge in each strongly connected component would accept invalid
subcycles. Traversal is iterative, checks every root and edge, and shares the
caller's cancellation/work gate. Tests include a 100,000-node chain, mixed guarded
and unguarded cycles, unreachable scratch nodes and cancellation in both passes.
This validator is used by the recursive intrinsic integration below.

`types.recursive` and `types.recursiveGroup` now invoke one-slot constructors in a
private arena. Object and tuple expressions, `types.array`, union/intersection
assembly, groups, nested builders and frozen helper captures participate. Native
normalization never receives an unsealed handle; scratch closures bypass the
ordinary constructor cache. Failed or blocked construction discards the arena and
replays the original builder expression, including implicit lambda capture reads.
After guard validation, generated native recursive aliases carry an explicit leaf
environment, preserving native generic substitution through lazy recursive members.
Self-free definitions retain their ordinary native result. Tests assert native
indexing, bidirectional assignability, distinct substitutions, group iteration,
invalid cycles, forbidden observations, finite kinds and allocation failures.
Broader native constructor adapters, declaration recipes and cancellation publication
coverage remain follow-up work; unsupported observations diagnose.

Recursive construction also supports `types.function`/`types.constructor`, including
recursive parameter/result Types, abstract constructors and mixed recursive groups.
Sealing emits ordinary native signatures; variance and signature inference remain
native. The original global Array and ReadonlyArray constructors have structural
adapters (including constructor arguments and direct DSL calls). Identity checks
reject a user alias that merely shares a built-in name. Tests cover native call/
construct observations, lexical substitution and invalid parameter/flag metadata.

Tuple descriptors can now carry recursive references as required/optional elements,
rest-array elements and variadic tuple payloads. Metadata remains separately
validated: inspecting Self as a label, mode, bound or flag diagnoses. Known private
tuple spreads normalize before native sealing, avoiding native alias cycles caused
by otherwise transparent spread aliases. Required construction tuples support
length/slice and group iteration without inspecting the references they contain.
Tests cover labels, readonly/rest/optional metadata, generic replay, invalid modes
and slice bounds, with both exact-optional modes. The declared optional value is
asserted separately from its read Type. Native sealing checks alias resolution
before accessing instantiation caches and attributes generated diagnostics to the
builder application.

Every reached nested recursive builder contributes validation roots, including a
builder whose result is later discarded. Invalid nested cycles therefore cannot
be erased by returning a constant or an unrelated structure. After validation,
only the output-reachable graph is materialized: discarded valid builders do not
introduce phantom native generic parameters. Regressions cover both discarded
failures and a generic-body reflection assertion that resolves without depending
on the unused builder's symbolic leaf.

Closures created inside a recursive arena charge storage to that arena rather than
the ordinary retained constructor cache. Their live charge is released together
with the arena on sealing, failure or deferral. A regression performs 100 separate
recursive constructions under a 4 KiB live budget; the same case previously
exhausted the budget by retaining charges for discarded private closures.

An iterative component planner now groups mutually dependent graph nodes in
source-to-sink order, including structural edges. This supplies the dependency
order for narrowing native alias environments below. Tests compare component membership against an
independent transitive-closure oracle, verify dependency order, and exercise a
100,000-node chain and cancellation without partial output.

Native recursive aliases now use one tuple environment per dependency component.
Nodes in a cycle pass that environment unchanged; edges to other components
project only the required slots. Dependency planning is iterative, deterministic
and charged to the invocation's live/work budgets. This avoids repeating the
entire group's leaf list at every alias and recursive edge. Both 256 independent
roots and a 256-root mutual cycle fit a 100,000-node budget; the independent case
previously exceeded it. Tests also verify distinct substitutions across projected
child environments and a generic-body assertion for a child with no dependencies.

Recursive object descriptors now preserve required/optional and readonly properties,
including unique-symbol keys. Synthetic key bindings retain the original native
symbol identity; string/number effective-key duplicates still diagnose. Both
exact-optional modes verify declared values separately from read values, distinct
symbol keys, generic replay and the ignored derived `read` descriptor field.

Recursive object indexers now carry private value payloads through sealing while
validating key domains and metadata immediately. Native string, number, symbol and
template-pattern index domains retain identity. Property/indexer and overlapping
domain checks execute on sealed native Types, including reached intermediate
objects whose results are discarded. Symbolic failed proofs defer and replay;
independent check components do not add dependencies to the returned shape. Tests
cover native indexing/reflection, overlap compatibility, generic substitution,
duplicate/invalid domains and discarded failures, including an ordinary result.

Recursive `types.object` and `types.overload` now forward ordered call/construct
handles through a narrow native anonymous-member resolver. Each sealed object
resolves its immutable signature sources under the native instance mapper, rather
than instantiating an incomplete raw-environment signature list. This preserves
native generic binders, duplicate overloads, method variance and abstract constructor
metadata. Signature-only free dependencies participate in readiness checks.
Tests cover nested single-signature forwarding, object/indexer combinations,
recursive parameters/results, native inference, query-order substitution and
invalid categories/counts. No signature is approximated by an intersection.

Signature descriptor processing now separates metadata parsing, native Type checks
and signature creation through one immutable plan. Optionality is added only when
creating native parameters or checking predicate compatibility; parameter names,
modes, receiver and predicate metadata stay independent of Type payloads. This
refactor retains all existing descriptor diagnostics and control-flow behavior,
and provides a shared validation path for recursive payloads.

Recursive signature descriptors now use that plan for parameters, explicit receivers,
returns and predicate narrowing Types. Metadata checks happen before sealing;
rest/predicate Type checks run after sealing and survive discarded results. Native
signature syntax supplies optional/rest behavior, ordinary function variance,
abstract construction and control-flow predicates. Both exact-optional modes test
reflection, `this`/parameter predicates, truthiness assertions, native narrowing,
generic replay and recursive results. Negative cases cover private metadata,
illegal rest/result/narrowing Types and discarded/symbolic failed checks.

Native islands now capture ambient `this` Types and `typeof this` roots in their
original native scope. Cloned Type nodes carry the mapped receiver; query roots
use private value bindings so member lookup and generic query arguments remain
native. Predicate `this is` targets remain syntax. Native definition diagnostics
invalidate the compiled program even when the host returns recovery `any`.
Regressions cover generic and derived classes, explicit function receivers, static
value queries, generic member instantiation, native-invalid contexts and editor
hovers. The original positive fixture produced nine spurious namespace/receiver
diagnostics; the recovery regression previously executed an invalid definition.

The integration harness now supports a declaration-only consumer in a fresh checker
and filesystem. It transfers emitted declaration bytes, never producer ASTs or
sources. Native controls test nominal/generic behavior, unavailable producer-source
imports and emit-on-error. The same path now tests generated DSL recipes.

Authored declaration files now enforce ABI 1 before DSL execution. The parser
checks definitions; the checker also validates constructor-valued aliases and
cached programs at consumer use sites. Missing, conflicting and unsupported
headers fail with `skipLibCheck` enabled. Consumer diagnostics retain producer
provenance, including hover-before-diagnostics requests. Native-only declaration
files need no header. Generated recipes use a printer marker that survives comment
removal and does not leak into unrelated files.

Written typefn declarations now emit source-equivalent bodies through the native
declaration transformer. Lexical dependency tracking retains private helpers,
imports, native nominal classes and unique-symbol declarations, including references
in unselected branches and native islands. Source recipes preserve closure capture
timing, namespace scope and constructor-valued aliases. Fresh consumers test the
calculator and HTTP-route modules directly, generic replay, private-helper visibility,
comment removal and aliases emitted in a separate module from their factory.
Invalid compiled recipes suppress the affected declaration output while independent
native files can emit. Pending display-only Types cannot be serialized as invented
declaration syntax. Recoverable call-return recipes now reuse their source annotation;
other escaping pending terms still require a nameable source annotation/helper. Complete inferred-recipe lifting, dependency closure validation
(including stripInternal), source maps, redirects and cancellation remain open.

DSL printing now enters the native source-map/comment pipeline for each node.
A decoded-mapping regression checks local-binding and return-statement positions
inside an emitted body, including the generated-line shift from the ABI pragma.
Declaration-map navigation, redirects and multi-file mapping scenarios still need
integration coverage.

Recipe dependency records now retain both the written import/name and its resolved
declarations. `stripInternal` rejects direct dependencies it would remove, including
private/exported helpers and imports from another source module. Namespace imports
track distinct target members; already-authored declaration dependencies remain
available and are not treated as newly stripped source.

Validation for emitted definitions and constructor aliases now follows transitive
source dependencies with an iterative, cycle-safe walk. A stripped helper behind a
native alias or in another source module blocks the importing recipe too; cross-file
failures include producer provenance. Inspection does not consume or enqueue the
other module's native visibility work. Explicitly annotated native declarations
contribute their type surface rather than their runtime bodies/default initializers.
Tests cover alias factories, indirect aliases, cross-module failures, recursive
native alias graphs and this native/DSL boundary. Ordinary native references now
validate reachable recipes too, with cycle-safe positive/negative reachability caches
and a native-only stripInternal control. Existing declaration files and referenced
inputs are traversed without treating them as newly stripped source. Malformed
dependency files are rejected before inspecting their recovery trees, and authored
invalid recipes cannot escape validation through skipLibCheck. Bundled native-library
declarations terminate the walk; separate user augmentations remain dependencies.
Generated references now participate through the native symbol tracker. It records
referenced symbols and source locations while node building holds the checker lock;
the transformer validates their closures after serialization releases that lock.
Per-file deduplication preserves deterministic first-reference diagnostics. Tests cover
inferred variable, property, return and default exports, forced/builder suppression
and an unchanged native-only consumer.

The harness exercises ordinary emission, forced declaration emission and incremental
builder signatures against invalid definitions, malformed DSL syntax and transitive
stripped dependencies. Even modes that bypass native noEmitOnError gates preserve
the per-file DSL failure boundary; independent native declarations still emit.

Inferred variables and properties initialized by native calls now recover a DSL-bearing
return annotation and the resolved signature's generic mapper before native value
serialization. The recipe belongs to the source expression, never to a shared literal
Type. Native node copying substitutes binder identities, retains accessible names and
imports, and schedules private helpers through the existing visibility tracker. A
strict-copy internal flag participates in node-builder cache keys and prevents native
recovery from silently materializing an inaccessible recipe. Native widening/type
compatibility is checked before source reuse; explicit user annotations retain priority.

Fresh declaration consumers augment an open interface and observe the recalculated
property count. Tests cover concrete and generic calls, mapped binders, generic class
fields, nested object properties and public cross-module helpers. They also retain
unrelated literal/native annotations unchanged and withhold stripped or unnameable
recipes in ordinary, forced and builder emit. Remaining inferred-recipe work includes
arbitrary value/return flow, inferred value-argument origins, native call requirements
for overloaded/non-nameable callees, and hygienic lifting of local helpers. This is not complete inferred declaration emit.

Explicit native type arguments now carry their original syntax into inferred recipes,
including calls whose return annotation is only a native parameter or tuple. A scoped
argument environment maps native parameter identities to caller syntax. Defaults
selected by a partial explicit argument list remain in the callee's scope and can
refer to preceding bindings; explicit arguments restore the caller's mapper and
argument environment. Cycle protection prevents malformed defaults from recursing
through source copying. No metadata is attached to shared native Types.

Native Type-identity serialization caching is disabled inside strict recipe copying:
two argument recipes can have identical current results and distinct consumer behavior.
Tests cover repeated equal-result calls, dependent/chained defaults, generic class
binders and augmentation. A value inferred to the same Type as its parameter's default
is deliberately not treated as use of that default.

Native inference now records whether its final argument came from the parameter's
default, including constraint fallback and inference-context cloning/reset. Only the
successful overload's facts are collected. They are published per call alongside the
cached native signature, not attached to the shared instantiated signature. Nested
resolutions retain ownership of their published facts; transient or failed resolutions
cannot overwrite them. Editor cache-reset scopes restore provenance with native
signatures and value Types. Disabled-profile calls allocate no DSL provenance state.

The recipe copier uses those facts to retain defaults actually selected by inference.
Tests deliberately share one native signature between a defaulted call and a call
with an inferred value of the same Type. Fresh consumers with augmentation distinguish
their output; completion and signature-help queries before/after semantic checking
preserve that distinction. Separate publication tests cover nested stack growth,
abandoned resolution, early error caching and restoration. This retains default
origins; it does not yet replay all native call constraints or inferred value origins.


For a nameable single-signature native callee, inferred declarations now project the
return Type from its instantiated value Type. The generated conditional infers both
parameters and return, retaining generic constraints and parameter-only DSL checks.
Source arguments/defaults still use the scoped recipe environment. Referencing the
original callable also keeps private return helpers in their defining module, without
exporting or cloning them. Native symbol tracking retains the callable/import and
validates its recipe dependency closure.

Fresh consumers augment native interfaces to invalidate an argument bound, a
DSL-computed bound, or a generic parameter-only calculation. The latter reports at
the consumer use even with skipLibCheck; ordinary native bound diagnostics follow
native skipLibCheck policy. Private cross-module return helpers now have a positive
consumer test. Inaccessible local helpers still suppress ordinary/forced/builder
output. Multiple overloads cannot use this conditional projection (native conditional
inference would select the last overload); cases requiring call validation fail
rather than silently publishing an unchecked result. Full overload/receiver handling,
value-origin tracking and hygienic local lifting remain incomplete.


Native identifier forwarding now retains a value query when the referenced
variable's declared Type matches the inferred export. Recipe reachability follows
value references in unannotated native initializers/bodies, using the value namespace;
explicit annotations still terminate inference dependencies. The existing cycle-safe
reachability cache and visibility tracker retain alias chains and imports. This does
not store provenance on shared literal Types or reinterpret control-flow narrowing
as a declaration Type.

Fresh consumers cover local alias chains, imported annotated/inferred values,
object properties, class fields, getters, arrows and single-return functions. Native
pseudotype/checker fallbacks preserve the recipe after verifying the inferred Type;
inferred predicates retain native handling. Unrelated literal values
and explicit widened annotations remain native. A stripped forwarded value is
rejected in ordinary, forced and builder output. Arbitrary expression/value flow,
local lifting and inference-origin reconstruction remain separate obligations.


Native specialization expressions (`read<Open>` without a call) now retain their
written type arguments in a native value query. Strict source copying keeps computed
arguments and partial-list defaults executable in the consumer, including generic
class fields. Calls through those already-specialized values project their existing
callable Type; they do not apply the original generic parameter list a second time.
A direct parenthesized specialization uses the caller's mapper/scope when copying
its arguments. Dependency tracking distinguishes native value specializations from
heritage Type references.

Consumer tests cover direct and stored specializations, defaults, generic class
fields, private imported return helpers and an argument bound invalidated by interface
augmentation. A separate specialization with the same current native Type but a
fixed argument remains fixed. Overloaded/non-identifier callee projection and
arbitrary inferred value origins remain incomplete.


Call-return recipes no longer require a written return annotation. A single native
callee with an inferred return uses the native instantiated return Type for the
producer equivalence check, applying only the enclosing builder mapper. The emitted
value query retains the helper's own inferred recipe. Nongeneric constructed callable
signatures also participate: when no source declaration exists, selection requires
signature identity, not equality of two missing declaration pointers. Native inferred
predicates keep their existing serialization path.

Consumer tests cover inferred generic functions/arrows, generic class fields and
functions produced by types.function. An augmented input that invalidates a concrete
constructed callable diagnoses in the emitted declaration. That declaration semantic
diagnostic follows native skipLibCheck policy; the separate ABI/frontend execution
safety checks remain mandatory. This differs from generic parameter checks instantiated
at a consumer use, which diagnose there even with skipLibCheck. Multi-return/async
flow, overload/receiver projection and local lifting remain incomplete.


Community acceptance now includes an executable type-fest finite-literal word-scanner
port and its original native conditional algorithm. Direct and fresh declaration-only
programs compare 464 cases each, alongside fixed expectations, generic options,
explicit union distribution, tuple metadata/reconstruction, domain failures and
resource cases. The imperative scanner uses UTF-16 offsets to avoid growing-prefix
concatenation: a 512-character word passes a 4,096-unit string-work limit that the
concatenating port exceeded. The benchmark includes both algorithms and project
setup; it does not isolate VM speed. Public Words normalization and the other
community families remain incomplete; [workload scope](workloads.md) records that boundary.
