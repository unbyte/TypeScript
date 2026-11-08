# Interpreter and runtime contract

R0 candidate. The implementation target is a handwritten, typed register VM
embedded in each native checker. This file specifies obligations before choosing
Go representations or reusing existing checker helpers.

## Stages

1. Parse dedicated DSL syntax with exact source spans. Native type expressions
   remain native AST islands. No source concatenation or generated-alias lowering.
2. Bind lexical Type/data/function/kind names and imported/native references.
3. Check kinds, return modes, captures, control flow, pattern scopes and intrinsic
   argument contracts. Compile-time data code is not dynamically typed JavaScript.
4. Lower a valid declaration to typed instructions with a source map. Do not run
   malformed or kind-invalid programs, including during editor recovery.
5. Evaluate on demand with native Type arguments or helper data values.
6. Publish a sealed native Type, a dedicated deferred application, or diagnostics.

Compilation is once per declaration body/environment schema per checker snapshot.
Native type-expression validation involving argument values is delayed to its
execution point. The compiler must distinguish static name/kind failures from
argument-dependent failures instead of eagerly evaluating all branches.

## Values, frames and operations

The VM has tagged immutable values: boolean, int, number, UTF-16 string, Type,
record, List, Map, Set, tagged variant, function closure, opaque descriptor/token
and void. Builder optimization is an implementation detail, not a mutable value
that can escape. A closure contains code identity and immutable captured values.

A frame contains a program counter, typed local slots, argument slots, iterator
snapshots, return destination and the current source operation. `type` and `let`
bindings use writable slots of different kinds; `const` slots are write-once.
Branch/pattern captures live in child lexical ranges and cannot be read later.
Every collection loop iterates a snapshot of its input, so modifying a local
collection in the body does not extend that loop's iterator.

Instruction families:

| Family | Responsibility |
| --- | --- |
| Load/store/move | Typed slots and snapshots; const enforcement |
| Scalar | Checked arithmetic, comparisons, strings and conversions |
| Data | Record/list/map/set construction, projection and local path updates |
| Control | Conditional branches, jumps, loop initialization/advance |
| Pattern | Ordered matching, temporary captures, pure guards and failure edges |
| Call | Helper/native-entry/lambda invocation, returns and depth checks |
| Native | Type expression instantiation, relation queries and reflection |
| Construct | Validated native objects, tuples, signatures, enums and recursive graphs |
| Suspend/fail | Propagation to entry boundary without publishing partial state |

The bytecode validator checks instruction operands, slot kinds, initialized reads,
branch destinations and lexical capture ranges before execution. A verifier bug
must not let untrusted source become a Go panic or arbitrary pointer operation.
Intrinsic dispatch uses a versioned whitelist, not arbitrary method names on Go
objects. Calls to runtime TypeScript symbols cannot reach this dispatch.

## Native adapter

All native operations pass through one audited adapter. It owns:

* operation-specific symbolic readiness, rather than a single “all arguments
  must be closed” predicate;
* native constraint/grammar validation in the actual argument environment;
* error translation and source provenance;
* resource and cancellation checks before/during expensive work;
* immutable substitution environments for lazy native structures;
* identity-preserving construction and descriptor round trips.

Native Type handles are valid only in their checker snapshot. No pointer is
stored in process-global state, persisted bytecode, external AST payloads or
cross-checker caches. Reflection returns descriptors, not exposed internal structs.
Native origin restrictions and incomplete recursive handles are checked at every
adapter entry, not only at public constructors.

## Symbolic results

A completed result may be a native symbolic Type containing ordinary native type
parameters. It is published immediately, so member access and inference see its
known structure. A blocked operation produces a **DeferredTypeFunction** native
type kind containing the entry identity, captured environment, arguments and
diagnostic provenance. It is not represented by an ordinary type parameter with
an unknown constraint.

The checker must explicitly handle this kind in instantiation, apparent/base
types, relations, inference, variance, keyof/indexed access, normalization,
serialization and display. Instantiation substitutes arguments and retries the
entry. Unresolved relations must remain unresolved; they cannot be cached as a
permanent false relation. Deferred types carry no claim to arbitrary invertibility.

Whole-entry restart is the initial operational model. Partial builders and
temporary enum/recursive registrations are discarded on Blocked or Failure.
Published symbolic return structures cannot retain those transactions. A future
continuation optimization must preserve the restart semantics and snapshot rules.

## Memoization and identity

A successful application cache key contains:

* declaration/closure code identity and DSL ABI version;
* immutable captured environment;
* normalized native argument identities and canonical data argument values;
* relevant native compiler options and the checker snapshot identity.

List/record/Map/Set keys hash their values with collision verification. Set equality
is order-independent; List and insertion-ordered Map equality are ordered. Function values use code
identity plus capture values, never an address that can be reused after collection.
Opaque native tokens are valid only within their snapshot. Values containing
unsealed placeholders cannot be cache keys.

Entry states are evaluating, completed, or deferred. Reentering evaluating with
the same key raises a cycle diagnostic with a bounded trace. Native recursive
type graphs are built using explicit constructors rather than bypassing this rule.
Deferred entries store restart applications, not incomplete frames. An error or
cancellation clears in-progress state and publishes no successful cache value.
Failures are not memoized as never/error Type results. Cached error *templates*
may be a later optimization only if they reproduce each application's provenance.

Native and DSL caches are checker-local. Dropping a checker snapshot drops its
handles and results. Incremental dependency reuse may retain parsed/verified
code by content hash, but rebinds native references and creates new Type handles.
Persistent result caching is outside R0; it needs an identity/serialization format
that cannot mistake equal displayed text for equal native Types.

Cache capacity is bounded. Eviction can affect performance, not the successful
value returned. Compiler query order and thread scheduling must not affect a
successful result except the explicitly unspecified ordering of TypeSet traversal.
Nominal enum identity is derived from its specified owner/key, not cache residency.

## Resource policy

Resource limits are part of the execution service, not approximations to types.
The candidate defaults below are provisional deployment values to be calibrated
by the conformance workload. Their existence and failure behavior are mandatory.

| Resource per outer native entry | Candidate default |
| --- | ---: |
| VM instruction fuel | 1,000,000 |
| Helper/native-entry call depth | 256 |
| Aggregate inspected/produced string units | 16,000,000 UTF-16 units |
| Integer magnitude | 4,096 bits per integer |
| VM-owned live allocation | 64 MiB |
| New native result/temporary graph nodes | 100,000 |
| Native structural work units | 1,000,000 |
| Diagnostic call sites | 12 plus omitted-frame count |

Nested calls share the same outer account. A cached child result still incurs a
call and result-access charge, but not fictitious allocation. Resource failures
are operational: a warm computation can cost less than a cold one. This is
documented rather than pretending that a fuel limit is a mathematical property
of a type. Given sufficient resources, successful results must agree. Failed or
canceled requests do not poison caches.

One VM instruction may trigger substantial native work. Therefore fuel alone is
insufficient. Native union cross products, mapped/member expansion, inference,
relations and graph construction must charge adapter/native traversal work and
check the cancellation token. Reusing an existing graph node does not allocate
it again; creating a large temporary that is later discarded still counts toward
work and live allocation. Constructors check prospective lengths and integer
sizes before allocation or multiplication whenever possible.

String search uses bounded linear-time primitives (or an equivalent work-accounted
algorithm), including pinned delimiters. Repeated suffix slicing uses shared
immutable views where safe; retained backing buffers count toward live memory.
Regex engines and potentially catastrophic backtracking are not part of R0.
Unicode code-point conversion is explicit and charged by inspected UTF-16 units.

No algorithm-specific silent truncation is implemented by these limits. A path
utility may itself choose a depth option; the VM reports exhaustion separately.
Diagnostics identify the exhausted resource, observed count, configured limit,
operation and root application. Users can adjust project limits; the implementation
must not recommend raising them to hide an evident nonterminating program.

## Cancellation, concurrency and determinism

VM dispatch polls cancellation at bounded intervals and at native-operation
boundaries. Long native/string loops also poll internally. A language-server
request's deadline can cancel evaluation independently of semantic fuel. A
canceled operation unwinds cleanup for frames, temporary graphs, registrations
and cache states. It cannot publish results or diagnostics to a newer document
version.

Each checker owns its VM state. Parallel program checking uses independent
checkers or the native checker's existing synchronization rules; no global
mutable execution stack or shared native-pointer memo table is introduced.
The native scheduler controls aggregate process memory when many requests run;
per-entry accounting is not a substitute for a process-wide capacity policy.

Performance must be measured separately for parsing, binding, bytecode compilation,
cold evaluation, memoized evaluation, native operation work and publication. A
fast cached hover is not evidence of cheap uncached type computation.
