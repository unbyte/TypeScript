# Symbolic evaluation and VM contract

Normative R2 chapter. This is an execution contract, not an implemented bytecode
format or a proof about the existing native checker.

## 1. Observable law

For a program P, environment E and capture-avoiding native Type substitution σ:

```text
finish(substitute(evaluate(P, E), σ))  ≃  evaluate(P, substitute(E, σ))
```

Agreement means the same native Type behavior and semantic success/failure under
the same compiler/options/module graph, with sufficient resources. It does not
require the same display spelling, ordering of independent diagnostics or resource
cost. Native nominal identities must be retained. Intermediate substitutions
compose: T→U then U→string agrees with T→string. Successful native shape alone is
insufficient if one path silently drops a required validation.

## 2. Result model

```text
TypeResult  = (TypeTerm, Requirements)
TypeTerm    = NativeType
            | CallType(Code, FrozenEnvironment, Arguments)
            | OperationType(Tag, FrozenOperands, NativeBinders)
Requirement = CompleteReachedCall | ValidateReachedOperation
Execution   = Returned(TypeResult | ConstructorValue)
            | Blocked(Dependencies) | Failed(Diagnostic) | Cancelled
```

These are implementation concepts, not source-level kinds. Native Type fields and
arguments can contain TypeTerms. Native/VM results carrying requirements are
guarded. CallType/OperationType are dedicated checker representations, never fake
any/unknown/never parameters. Requirements are deduplicated by semantic identity
while retaining source/use provenance. They are not catchable values.

For an evaluation judgment `E,J ⊢ e ⇓ (t,R)`, the caller's journal becomes `J ∪ R`.
For a return, its result's requirements include the entire reached frame journal,
not merely dependencies reachable from t. Binding, rebinding, indexing and native
normalization cannot delete journal entries. Checks inside an uncalled closure or
unexecuted branch are not reached and add no entries.

## 3. Deferral versus control blocking

All ordinary Type-producing calls and operators may return deferred TypeTerms.
This includes numeric arithmetic, comparisons, string operations and reflection
when their results require unknown information. Native construction can embed
those terms while exposing an already known outer shape.

```ts
typefn Measured<S> {
  return { source: S; length: text.length<S> };
}
typefn DiscardLength<S> {
  type unused = text.length<S>;
  return string;
}
```

Measured<S> may expose source/length keys without reading S. DiscardLength<S>
may expose string but retains the string-literal eligibility check. Both fail
when S becomes number. R1's data-string extraction rule that blocked the whole
call is superseded: there is no separate concrete-data result category in R2.

A control decision must observe a ready value: if/while, logical short circuit,
iteration sequence, pattern selection and abort-message production. If needed
information is unresolved, that *observing call* blocks before choosing a path.
A Type-returning blocked call discards its mutable frame and produces CallType
for its original invocation, with the requirement to finish it. A caller may
embed this term; observing it in turn blocks that caller.

Constructor-returning calls cannot leave abstract constructors at the native
boundary. A constructor result must have known code, immutable captures and no
pending creation checks. Symbolic captured Types alone are allowed. A blocked
factory propagates until a Type-returning call can defer the entire computation.
A native constructor alias that cannot finish creation diagnoses. Selection and
application can be moved into one Type-returning function.

## 4. Readiness

Readiness is operation-specific, not “contains any free parameter.”

| Operation | Required behavior |
| --- | --- |
| `{ value: T }`, `[T,U]`, Array<T> | Publish known structure with symbolic leaves |
| `keyof {value:T}` | Return `"value"`; retain guards |
| `tuple.length<[T,U]>` and required tuple iteration | Return 2 / visit two Type values |
| Shallow reflection of `{value:T}` | Return known key/flags with value T |
| Native bound on T extends string | May use affirmative native proof to validate |
| `T is string` | Defer even if T extends string; a bound is not the observed argument |
| `T == T` for the same term | True, retaining requirements |
| Other comparison, scalar arithmetic, text.length<T> | Defer unresolved scalar operands; closed wrong domain diagnoses |
| Native conditional/mapped/template operation | Use lawful native symbolic nodes or retain an operation recipe |
| Union iteration with any reachable free parameter/deferred term | Wait under R2's conservative policy |
| Reflection with unknown keys/signature layout | Defer the query |
| `typeof runtimeName` in an island | Native static Type query only |

Native validation and value observation are separate protocols. A positive native
constraint proof can establish legality without revealing a scalar value. Failure
to prove a symbolic relation is not false. Concrete unknown/any/never are ordinary
native Types with defined operation domains; they are not unresolved binders.

Each blocked query records its unresolved binders/terms/native resolution tokens.
Stable shape/closedness summaries must be cycle-aware and must not eagerly expand
every latent member of a recursive native Type. An operation lacking both a result
and a genuine unresolved dependency is an unsupported-operation diagnostic, not
an endlessly rescheduled wait.

## 5. Requirements across native simplification

```ts
typefn Checked<T> { return Uppercase<T>; }
typefn Discard<T> { type unused = Checked<T>; return string; }
type Erase<X> = string;
type Hidden<T> = Erase<Discard<T>>;
type Absorbed<T> = any | Checked<T>;
```

Hidden<number> and Absorbed<number> fail. Native alias erasure, union absorption,
projection, literal unboxing and cache reuse must preserve guards from evaluated
operands. A result's free-variable summary includes dependencies used only by guards.
Actual native generic arguments validate before an erasing body can discard them.
No adapter may invent any to satisfy an unresolved constraint.

Native conditionals and mapped types keep their native laziness/distribution.
Do not execute calls in unselected branches merely to collect checks. Native
definition diagnostics remain native diagnostics. Passing an opaque interface
does not force all of its latent member computations; when a member is actually
resolved, its reached checks must enter the protocol.

A closed resolved application MUST finish its reached requirements before being
reported as successful at a native use, closed alias check or declaration boundary.
A pending application with genuine free dependencies remains a recipe instead.
This does not execute every uncalled typefn body or every latent native member;
it completes work already reached by the resolving query.

## 6. Replay and cycles

Substitution traverses Types, deferred operands, checks, constructor captures and
native bound environments by binder identity. It never rewrites names by spelling.
Use a fresh immutable environment. Reuse unchanged completed structure; retry a
blocked call only when relevant dependencies change. Retry from the call's entry,
with original capture snapshots after substitution, not today's mutable locals.
There is no serialized continuation or partially executed frame.

Reentering an active application with identical code, captures and arguments is a
cycle diagnostic. Recursion with changing arguments is allowed within limits.
Native guarded recursive Types and recursive builders use graph sealing, not
unbounded repeated execution of identical calls. Pattern/closure/native binders
must remain hygienic through replay and declaration reconstruction.

## 7. VM architecture and scalar representation

Use a handwritten, validated register VM. Compile each bound DSL body once per
checker snapshot. Bytecode instructions carry source spans and statically known
slot kinds; conceptual instruction families are load/move, construct/apply,
intrinsic/native operation, branch/jump, iterate, capture, return and abort.
Numeric opcode numbers and register allocation are implementation-private.

Only validated bytecode executes. Validation checks branch targets, slot
initialization/kinds, argument counts, capture layout and intrinsic dispatch.
No bytecode or source file can name host pointers, execute JS, call the network,
read files/time/randomness or mutate the native checker graph directly. All native
operations pass through one adapter that owns readiness, legality, checks,
identity, graph construction, cancellation and work accounting.

Eligible literal Types MAY be represented as tagged UTF-16 strings, binary64,
bigint or boolean values internally and reified lazily. This must preserve native
canonicalization, operand domains and all guards. Unboxing is not permission to
unwrap enums/brands or use data equality instead of native Type identity.
Descriptors/tuples may use persistent structural storage before reification.
Optimization must not change snapshots, observation order or validation reachability.

## 8. Caching

Keys include checker snapshot/native options, DSL ABI/code identity, frozen
captures, argument kinds/identities and guard dependencies. Entries distinguish
evaluating, blocked, returned-with-checks and complete. Constructor identity is
code plus capture identity, not extensional equality of outputs. Native handles
are checker-local; cross-snapshot result caching is outside R2.

Failed/cancelled results are not installed as successes. A completed child may
remain cached when its parent blocks; frames/unsealed builders cannot. Eviction
may change work but not successful semantic results with sufficient resources.
Cache hits reattach current use-site provenance instead of reusing the first
call's diagnostic location. Unchanged blocked requests reuse dependency state.

## 9. Budgets and cancellation

All nested VM and native work for a root request shares accounting. Initial R2
default limits are explicit implementation targets:

| Limit | Initial value |
| --- | --- |
| VM instructions | 1,000,000 |
| Active DSL call depth | 256 |
| DSL syntax nesting | 256 |
| DSL parser recognition work | 100 × token count + 10,000 rule visits |
| Live VM-owned allocation per root evaluation | 64 MiB |
| Single bigint magnitude | 1,048,576 bits |
| String work | 16,777,216 UTF-16 units examined/produced |
| Native adapter work | 1,000,000 graph visits/relation steps |
| Newly constructed native nodes per root | 250,000 |
| Snapshot DSL cache capacity | 64 MiB, eviction permitted |

These are quotas, not performance measurements or a promise that every program
below them is fast. Options may raise/lower positive limits; exact spelling is in
integration. Parser and checker service memory limits still apply outside the VM.
Count native work inside an expensive relation/normalization, not just one opcode
for the entire call. Charge string scans and bigint work in addition to dispatch;
each bigint limb processed charges native/primitive work against nativeWork.
Allocate only after quota checks. Cached retained memory counts against cache
capacity; new evaluation-owned buffers count against live allocation. Native
nodes remain subject to the checker memory policy as well as the construction cap.

Poll cancellation at least every 1,024 VM instructions or primitive/native work
units, and before/after adapter calls. Long native operations must poll internally.
Cancellation is not a source diagnostic. Exhaustion produces a focused resource
diagnostic naming the limit and bounded call trace; it never returns an approximate
Type or never. Cancelled/stale editor work publishes no result or declaration.

The counters and defaults are versioned service policy, not mathematical Type
semantics. Warm/cold runs may reach different operational limits. With sufficient
resources, cache policy cannot change the law in section 1. Measure actual latency,
allocation and native-project overhead before calling an implementation performant.

## 10. Diagnostics

Diagnostic categories are syntax, binding/kind, scalar/domain, native constraint,
descriptor construction, explicit fail, unresolved native use, constructor-not-ready,
recursive cycle, resource exhaustion and declaration ABI/nameability. Cancellation
is separate. Recovery error Types are tainted and cannot enter successful caches.

Point first to the failing specialization/use. Related spans identify the original
operation, constructor constraint and bounded call/capture trace. failAt also
includes the author's UTF-16 offset; it is not automatically a source-file offset.
Distinct applications get distinct use locations even when they share a cached
requirement. Native normalization must not reduce a failure to a missing-member
error on a fabricated unknown Type.
