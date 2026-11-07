> Historical R1 material, superseded by the [R2 specification](../../spec/README.md). Not the current language contract.

# Symbolic evaluation and deferred checks

R1 normative design candidate within the [agreed restricted HKT scope](restricted-hkt.md).
This settles the execution decisions below; it is not a claim of implementation,
metatheoretic proof or measured performance. No new programmer-facing syntax is
introduced. Internal terms in this document are compiler concepts.

## 1. The observable contract

A typefn computes with immutable Type values and locally mutable bindings. Knowing
a Type does not require knowing every generic argument inside it. `{value:T}`
already has a known field; a string parser cannot yet read an unknown T.

The required specialization law is:

```text
Evaluate a call after substituting its arguments
  agrees with
Substitute and finish its previously computed symbolic result and pending checks.
```

Agreement means the same successful native Type behavior and the same validity
or failure, under the same compiler/options/module graph and sufficient resources.
It does not require identical pretty-printing, diagnostic ordering when multiple
requirements fail, or resource counts. Native identity-sensitive Types retain
their original declaration identities; equal displayed text is insufficient.

This law rules out replacing an unknown argument with its constraint, silently
dropping a failed computation, and permanently caching an unresolved relation as
false. It also permits a result containing symbolic native Type parameters without
executing branches that depend on their eventual values.

## 2. Definition checking versus execution

Every body is parsed, bound and kind-checked before execution. This checks names,
scopes, fixed arity, Type/data categories, return kinds, assignment rules and
helper-data typing, including in unreachable branches. Malformed editor buffers
never enter the VM. Recursive return-kind inference needs an annotation.

Argument-dependent native Type legality is checked when the corresponding DSL
operation is reached. A known-false branch does not execute its native applications
or `fail` statements. An unresolved condition does not execute either arm. This
is different from checking every native type expression in every branch eagerly.
Native TypeScript declarations outside a typefn keep their normal definition checks.

Native Type-valued expressions inside the DSL use native construction semantics,
with DSL calls participating in the staged protocol below. Direct native generic
applications and calls through F check the same original bounds at their execution
point. No invalid generic application is made valid merely by hiding it behind F.

## 3. Result representation

The conceptual representation is:

```text
TypeResult = { value: TypeTerm, checks: ordered set<Requirement> }
TypeTerm   = NativeType
           | CallType(known code, frozen environment, arguments)
           | OperationType(native operation, frozen operands and binders)
Requirement = finish a reached call or validate a reached native operation

TypefnExecution = Returned(TypeResult | ConstructorValue)
                | Blocked(dependencies)
                | Failed(diagnostic)
                | Cancelled
```

Native Types may contain TypeTerms in their fields, tuple elements or other Type
positions. CallType and OperationType are dedicated checker representations, not
any, unknown, never or an unconstrained native type parameter. OperationType has a
fixed native-operation tag, not an arbitrary suspended VM continuation. It retains
all information needed to repeat native semantics, including conditional binders
and distribution behavior.

A TypeResult with outstanding checks is internally **guarded**. Its known shape
can participate in generic checking under those requirements, but its checks are
not erased by taking a property, normalizing a union or printing a declaration.
This wrapper/metadata is not a new kind visible to the programmer. The compiler
must preserve it through all native paths that can carry the result. A global
side table disconnected from Type substitution is not an adequate implementation.

Local data helpers use the same protocol for their journal: a returned data value
also returns its reached checks to the caller. A helper cannot hide a pending Type
call by returning an unrelated integer. Data itself is concrete; no symbolic data
value or resumable helper frame crosses the native boundary.

### Type-returning calls

Evaluate arguments in source order and validate the callee's input bounds. If its
body returns, collect its result and all pending checks reached on that execution
path. If it blocks, discard its mutable frame and unsealed builders and produce
CallType for its original application, with a requirement to finish that call.
The caller can carry this Type value without knowing its final shape.

A nested blocked Type-returning call does not automatically block construction
of an enclosing object:

```ts
typefn Upper<T> { return Uppercase<T>; }
typefn Envelope<T> {
  return { tag: "text"; value: Upper<T> };
}
```

For unconstrained T, Envelope<T> has the known outer fields and a pending value
Type/check. Envelope<"hello"> completes to `{tag:"text";value:"HELLO"}`.
Envelope<number> fails, even if a later consumer only asks for `tag`.

When an observation needs the unfinished result, it blocks the *observing* call.
For example, reading `literal.text(Upper<T>)` cannot manufacture a data string.
The resulting CallType replays that observing call from its entry after substitution.

### Constructor-returning calls

There is no abstract constructor term at the native boundary. A constructor
result must have known code, immutable captures and no unresolved checks from
creating it. Captured ordinary Type parameters are allowed; they are values, not
pending work. Creating a closure does not execute its body.

If a constructor factory blocks, the suspension propagates through callers until
a Type-returning call can retain it as CallType. A native constructor alias with
no such enclosing call gets a constructor-not-ready diagnostic. Keep selection
and application in a Type-returning typefn. No fallback to a native F parameter,
constraint, arbitrary constructor or constructor-kind unknown is allowed.

## 4. Pending work cannot be erased

```ts
typefn Checked<T> { return Uppercase<T>; }

typefn Discard<T> {
  type unused = Checked<T>;
  return string;
}
```

Discard<T> may publish the string shape, but retains the requirement to complete
Checked<T>. Discard<number> is invalid both directly and after substituting number
into Discard<T>. Rebinding `unused`, returning a constant or dropping a descriptor
field must not lose that requirement.

The VM therefore keeps a frame-local journal of checks from **reached** operations,
separate from the values still present in local slots. A successful return combines
that journal with checks on its returned value. Unexecuted branches and uncalled
closure bodies add no checks. A blocked outer call retains its original recipe;
its partial journal/frame is not exposed as a resumable continuation.

The same rule applies at the native boundary:

```ts
type Erase<X> = string;
type Hidden<T> = Erase<Discard<T>>;
type Absorbed<T> = any | Checked<T>;
```

Hidden<number> and Absorbed<number> both fail. Type operations preserve checks
from evaluated operands even when the native value normalizes to string or any.
They may normalize the *shape* but must keep the guard. Native conditional and
mapped types retain their native evaluation/distribution rules; this does not
eagerly execute DSL calls in branches that those rules have not selected.

Native generics with invalid actual arguments must validate those arguments
before their body erases them. A pending constraint becomes a check, not an
unchecked native instantiation. Passing actual any retains native bound behavior;
the adapter must never invent any to discharge a pending check.

Conditional branches and mapped-type bodies are retained as scoped lazy recipes
where native evaluation requires that. Do not evaluate both conditional branches
just to assemble their checks. A pending operator retains the requirement to finish
its selected computation; it does not assert that every possible branch executes.
This does not suppress stock TypeScript's existing static declaration diagnostics.

Checks apply to reached computations, not every latent member of every referenced
native Type. Passing an opaque recursive interface Type does not recursively force
all its members. Native demand boundaries for those member definitions remain;
when a member computation is actually resolved, its checks must be retained.
The protocol must not turn ordinary recursive Type graphs into eager VM traversals.

This preservation is part of core DSL integration, not extra HKT polymorphism.
Without it, symbolic evaluation of imperative computations changes their meaning.

## 5. Readiness is specific to an operation

The adapter distinguishes a stable answer, a closed failure and an unresolved
dependency. Having free Type parameters somewhere in a graph is not by itself
a reason to block every operation.

| Operation | Ready with symbolic inputs | Must wait / fail |
| --- | --- | --- |
| Build `{value:T}`, `[T,U]`, `(x:T)=>T` | Known shape with symbolic Type leaves | Illegal structure fails; unresolved computed keys need deferral |
| Apply known Array to T | Produces native T[] | Unknown constructor identity cannot enter native application |
| Native bound `T extends string` | Native checker establishes it from T's actual bound | Unconstrained parameter waits; closed number/unknown fails |
| Whole-Type `T is string` | Closed Type observation | Unresolved T waits even if constrained to string; validation and observation are separate |
| `T == T` | Same immutable handle is reflexive; checks still retained | Distinct symbolic operands do not become equal merely by compatible bounds |
| Numeric Type ordering `<`, `<=`, `>`, `>=` | Two finite number literals or two bigint literals | Unresolved operands wait; broad, multi-member or mixed-family operands diagnose; do not fold T<T before scalar validation |
| `keyof {value:T}` | `"value"`; keys do not depend on T | `keyof Checked<T>` retains an operation recipe until its operand resolves |
| Shallow object reflection | Known fields/flags; field Type handles may remain symbolic | Unknown field set, indexer/signature layout or required metadata waits |
| Required tuple iteration of `[T,U]` | Two elements, with symbolic Type values | Unknown length/mode waits; closed optional/open tuples fail the current policy |
| String literal iteration | UTF-16 units of one unbranded literal, each a Type | Unresolved subject waits; broad strings, infinite templates and unions fail the sequence requirement |
| Union iteration | A normalized member list proved stable under pending substitution | Unresolved member identity/absorption waits; never is empty, any/unknown singleton |
| String extraction/matching | Closed literal/template subject under the specified match rules | An unresolved parameter waits, even if it has a string constraint |

For union iteration, the initial algorithm is conservative: a reachable free
Type parameter or deferred term in the subject causes a wait. Thus a literal
union is iterable, but even a known-discriminant union with generic fields waits.
Native union *construction* still accepts symbolic leaves. Removing this traversal
restriction later requires a proof of stable normalization; it is not an assumed
ability of the current checker. Unordered member traversal remains unspecified.

Type-producing native operators use native symbolic nodes when those nodes preserve
dependencies correctly; otherwise they retain OperationType. The required operator
families are native instantiation, keyof, indexed access, conditional/mapped types,
templates/string mappings, unions/intersections and validated structural builders.
No operator may receive a fake ordinary-Type approximation of an unresolved term.
Descriptor/data-producing queries cannot return symbolic booleans, arrays or text;
they block the observing call instead.

Native constraint validation may use affirmative native proofs under declared
bounds. Observations do not guess an actual argument from those bounds. The initial
observation rule permits identity reflexivity and explicit structural independence
above, not general quantifier reasoning over all possible arguments.

## 6. Substitution and replay

Substitution traverses result Types, guards, deferred operands, ordinary Type
arguments and captured Types in closures/data descriptors. It respects native
generic binders and pattern/helper scopes; capturing a Type parameter named T
does not license replacing every identifier spelled T.

1. Substitute into a fresh immutable environment; do not alter the original result.
2. Reuse completed native structure where valid. Reevaluate changed pending
   requirements and operation recipes; do not repeat an unchanged blocked request.
3. If a call can advance, restart its original body from the entry. Captures remain
   the original snapshots with substitution applied, not the values of today's locals.
4. Publish a completed result, a smaller guarded/symbolic result, a diagnostic, or
   cancellation. Keep remaining dependencies until a later specialization.

Intermediate substitutions must compose: T→U followed by U→"x" agrees with
T→"x". A closure snapshot made before reassignment keeps the old value. A caller's
inference variables cannot be inspected while the native inference session is
still mutating them; retry only at a stable native substitution/query boundary.

Dependencies identify unresolved binders, deferred terms or native resolution
tokens. A blocked operation with no unresolved dependency is an unsupported/invalid
operation diagnostic, not an endlessly rescheduled computation. Reentry into the
same actively evaluating application is a cycle diagnostic. Native guarded recursive
Type construction is separate from recursive execution of the same VM call.

## 7. Native checking, relations and inference

Published shape uses native relations and ordinary inference. Unknown shape does
not supply arbitrary members, indexers or inverse inference candidates. NoInfer
continues to suppress ordinary native candidates. Relation caches must distinguish
a proven answer from a blocked query and retain the guards needed by a proven
shape relation. Syntactically identical deferred recipes are reflexive under the
same environment; they still have to satisfy their pending validity checks.

A cached positive shape relation records its required guards. Reusing it in a
different generic context must discharge or match those guards there; a proof
under one function's signature assumptions is not an unconditional cached true.
Source locations are separate from semantic recipe identity, so matching the same
requirement at two call sites does not lose either site's diagnostic provenance.

A native use requiring a relation that remains unproved receives a focused
unresolved-computation diagnostic. It is not silently accepted or permanently
cached as false. A deferred type operator may retain that dependency for later
substitution. Generic runtime function bodies are not delayed for rechecking at
every concrete call.

### Generic bodies are checked once

Native generic type aliases can describe partial computations such as TextOf<T> =
Apply<StringBox,T>. A local unused type alias also describes a recipe, rather than
executing the alias at runtime. Actual closed specializations must drain their
checks, even if native simplification erased the visible dependency.

Written native function/signature parameter and result Types can carry pending
checks. A generic body may reason under those explicit signature requirements,
just as its callers must satisfy them. Matching an existing requirement is by
canonical recipe identity after substitution, not by theorem proving. At a
concrete call/signature specialization the checker validates these requirements
even if an argument is any or the caller ignores the return value.

A generic body cannot introduce new unproved local checks and rely on its callers
to run the body later. At a value operation, a local check must be discharged from
known Types/bounds or already be required by the written signature. Otherwise
report it at the generic definition. Do not infer hidden constraints or synthesize
runtime-function effects to make such a body legal.

```ts
typefn Text<T> { return Uppercase<T>; }

function good<T extends string>(x: Text<T>): Text<T> {
  return x;
}

function declared<T>(x: Text<T>): Text<T> {
  return x; // Identity under the signature's explicit requirement.
}
// declared<number>(...): constraint error during signature specialization.

function bad<T>(x: any) {
  const y: Discard<T> = x; // Definition error: new unproved validity check.
  return y;
}
```

Returning an already valid constructed object can still infer an ordinary native
signature. Only unresolved additional validity conditions require an explicit
contract/bound. Assignability of function values must not drop source requirements:
a concrete specialization checks them first; a symbolic comparison must match
them to the target signature's requirements or prove them. A conversion to a
signature that silently accepts invalid specializations is rejected. This is an
obligation of partial Type computations, not native constructor-kind subtyping.

These checks concern statically known signatures. Native any/casts can erase a
callable Type just as in ordinary TypeScript; the DSL does not reconstruct that
signature or add runtime enforcement. Supplying an any argument to a still-known
signature, however, does not erase the signature's own specialization checks.

## 8. Diagnostics, cancellation and caching

| Category | Timing and meaning |
| --- | --- |
| Syntax/name/kind | Definition checking, before execution, including unused branches |
| Input/operation validation | A reached native operation is invalid for known inputs |
| Explicit failure | A reached `fail(message)`; never converted into never or match failure |
| Unresolved use | A native operation needs information or validity unavailable in its generic context |
| Constructor not ready | Native constructor alias cannot publish the factory's unfinished result |
| Cycle/resource | Execution cannot finish within the service rules; not an approximate Type |
| Cancellation | No source error; discard stale query work |

An application diagnostic points first to the current failing specialization/use.
Related locations identify the original typefn operation, supplied native constructor
and bound, and a bounded call/capture trace. Errors discovered after substitution
use the new application's provenance, not the first cached generic use's location.
Deduplicate by requirement identity and use site; identical failures at different
user applications still get their own locations. Editor recovery error Types are
tainted and cannot enter successful VM caches or emitted declarations.

Cache keys include checker snapshot/options, ABI/code identity, frozen captures,
argument kinds and argument identities, including pending checks. Cache entries
and free-variable summaries must include dependencies used only by guards. A
string-shaped result depending on Checked<T> still needs substitution of T.
Cache entries
distinguish evaluating, returned-with-checks, blocked and complete. Never install
a result from a failed/cancelled evaluation. A cached successful child may remain
valid after its parent blocks; no mutable frame or unsealed builder may remain.

All nested native/VM work shares budget and cancellation accounting. Warm caches
can reduce cost; resource failure is operational, not a mathematical Type result.
Given sufficient resources, cache eviction cannot change success values or drop
semantic checks. Edit invalidation changes the checker snapshot; native handles
and results never cross it. Cross-snapshot persistent result caching is excluded.

## 9. Concrete integration points

The pinned Go baseline's `instantiateTypeWorker` in
[checker.go](../../../../internal/checker/checker.go) explicitly switches over native
Type categories. It needs deferred/guard substitution before falling through to
the original Type. `getUnionTypeWorker` can return any/unknown early, so guards
must be extracted and reattached before that simplification. `checkTypeArgumentConstraints`
currently returns a boolean; the DSL adapter must retain pending-versus-failed
information rather than treating that boolean as its whole readiness protocol.

Native interfaces/classes still receive only Type arguments. Their Type graphs
may now contain guarded or deferred Type computations, as they already contain
symbolic native operators. Constructor values stay in the DSL environment. The
inference/relater, nodebuilder, declaration transformer and language server must
all preserve the protocol; these inspected hooks are a starting map, not an
exhaustive code-change list. [Semantic cases](semantic-cases.md) define the expected
observations. [Declaration emit](declaration-emit.md) preserves the executable recipes.
