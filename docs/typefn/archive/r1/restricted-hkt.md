> Historical R1 material, superseded by the [R2 specification](../../spec/README.md). Not the current language contract.

# Restricted HKT design direction

Agreed scope for the next R1 design iteration, following user feedback on
2026-11-06. This narrows the
[broader exploration](../../higher-kinded-types.md) in response to the
[TypeScript constraint audit](../../typescript-hkt-audit.md). Agreement selects the
design direction; it does not establish a complete specification or an implemented
feature. Where the two disagree, this document defines the current
scope; the broader document remains a record of possible future capabilities.

The rule to teach is: **type functions may take and return type constructors;
ordinary TypeScript generics still take ordinary types.**

The [evaluation contract](evaluation.md) specifies symbolic publication, deferred
checks, replay under substitution and diagnostics. The [declaration contract](declaration-emit.md)
specifies source recipes, captures and ABI behavior. These remain implementation
and validation obligations; they do not reopen native constructor parameters or
constructor inference. [Semantic cases](semantic-cases.md) make their decisions
reviewable without claiming a working compiler.

## The useful part stays small

```ts
typefn Apply<F<_>, T> {
  return F<T>;
}

typefn Compose<F<_>, G<_>> {
  return typefn<T> { return F<G<T>>; };
}

type AsyncArray = Compose<Promise, Array>;
type Names = AsyncArray<string>; // Promise<string[]>
```

`F<_>` accepts a constructor with one ordinary Type argument and a Type
result. `F<_, _>` accepts two. `_` is an anonymous parameter slot, not a partial
application hole. All arguments are explicit. A returned closure captures the
current bindings by value, including constructors and symbolic ordinary Types.

Partial application uses those same closures:

```ts
type Outcome<A, E> =
  | { ok: true; value: A }
  | { ok: false; error: E };

typefn WithError<E> {
  return typefn<A> { return Outcome<A, E>; };
}

type Result = WithError<Error>;
type Message = Result<string>;
```

No extra placeholder syntax or new application operator is needed. Use a binding
for a returned constructor; chained `Compose<F,G><T>` syntax is outside this scope.

## Five restrictions

| Restriction | Work it removes | What remains useful |
| --- | --- | --- |
| Constructor parameters only in `typefn` | Native generic constructor slots, their variance and inference rules | Pass native Array/Promise/aliases into DSL computations |
| Passed constructors accept ordinary Types and return an ordinary Type | Recursively nested constructor-parameter kinds and constructor subtyping | Unary/binary/n-ary transformations, closures and composition |
| Validate declared bounds at application, not by comparing constructor domains | Universal domain implication and dependent-bound matching between constructors | Bounded native operations pass directly; each actual call must satisfy their bounds |
| Constructor arguments are explicit | Recovering a constructor from an applied Type, constructor-aware NoInfer | Ordinary Type inference from already published native structure |
| No abstract constructor escapes into native checking | Native `F<T>` relations, output bounds and hypothetical constructor member access | Known constructors may still be applied to symbolic ordinary Types |

These restrictions remove specific obligations identified in the audit. Restricting
only arity to one would not remove them, and would unnecessarily hinder tuples,
error/result transformations and other multi-argument operations.

## Which functions can be passed?

A constructor slot accepts a native generic reference, an eligible named typefn,
or a typefn closure with exactly the required arity. Each input slot has ordinary
Type kind, and the result kind must be Type. Declared bounds are retained for
checking each application, not compared with a universally quantified receiving
domain. Defaults do not remove parameter slots. This is a contract about how to
call a compile-time operation, not a promise that it succeeds for every Type.

Applications of constructor values supply all their fixed slots. Direct named
native generic uses retain native default-argument behavior.

The DSL therefore needs a flat kind description for parameters and results:

```text
Kind = Type | Constructor(arity)
```

Constructor(arity) always describes ordinary Type inputs and a Type result.
A named typefn has a parameter-kind list and one result kind; it does not need
the broader proposal's recursively nested constructor-kind calculus.

The named combinator Compose can take constructors and return a constructor, but
it cannot itself be supplied as F<_>. It has the wrong argument/result kinds.
Likewise, a parameter such as `H<F<_>>` is excluded. Composition works by directly
calling named combinators and passing the resulting ordinary Type-to-Type closure.
First-class functions that themselves consume or return constructors are deferred.

Kind checking happens without running a body. All return paths must have the same
kind, including constructor arity. Recursive inference needs an annotation:
`: type`, `: typefn<_>`, or a constructor annotation with another fixed slot count.
The return-kind annotation is optional for acyclic, inferable cases.

Having ordinary Type arguments means an application has the right kind, not that
an arbitrary program terminates or succeeds. Native bounds or a typefn's validation
can reject an input; a computation can exhaust its budget. No totality guarantee,
result bound or constructor equality is inferred.

### Bounded native operations still work

```ts
type StringBox<T extends string> = { text: T };
type Text = StringBox<"hello">; // Ordinary native application remains valid.
type Good = Apply<StringBox, "hello">; // { text: "hello" }
type Bad = Apply<StringBox, number>;   // Native bound fails at F<number>.
type Upper = Apply<Uppercase, "hello">; // "HELLO"
```

F<_> describes arity and argument/result kinds only. Applying the supplied
constructor checks its original constraints before instantiating its result,
even if its body would erase the argument. Dependent native bounds, such as
Pick's `K extends keyof T`, are checked under the actual preceding arguments by
the native adapter. The DSL does not prove compatibility between two bound schemas.

With symbolic ordinary arguments, a bound proved by the native checker permits
evaluation. A violation on closed arguments produces a diagnostic. If a bound
cannot yet be established, retain its deferred application/operation and required
check as specified in evaluation.md. A caller can build known outer structure
around a pending Type; an observation needing its result must wait. Failure to
prove the bound is not permission to ignore it. Concrete unknown is an ordinary
Type, not an unresolved type parameter. No sampling with any is allowed.

This deliberately permits a deferred alias such as
`type TextOf<T> = Apply<StringBox,T>`, which can later fail at TextOf<number>.
It is a partial compile-time computation, just like a parser typefn that reports
invalid input. Ordinary native generic declarations retain their existing bound
checking rules; this staged behavior belongs to typefn applications. Diagnostics
must trace the application back to both the supplied constructor and the DSL call.

An explicit adapter can choose an input policy:

```ts
typefn UpperText<T> {
  return Uppercase<Extract<T, string>>;
}

type FilteredUpper = Apply<UpperText, "hello" | 1>; // "HELLO"
```

UpperText intentionally filters nonstrings. This is not a claim that Uppercase
itself accepts nonstrings. Adapters must preserve or deliberately specify their
validation/filtering semantics. Ordinary Type parameter bounds on typefns remain
available and are checked when those functions are applied, including via F.
Constructor parameters have no universal input/output-bound syntax in this scope.

## Symbolic Types remain supported

Known constructor identity does not mean concrete arguments:

```ts
typefn Wrap<F<_>, T> { return { value: F<T> }; }
type Wrapped<T> = Wrap<Array, T>; // { value: T[] }

function unwrap<T>(x: Wrapped<T>): T[] {
  return x.value;
}
```

The evaluator knows F is Array and can publish its native structure while T is
still generic. Only the DSL declaration has an F parameter; Wrapped and unwrap
have ordinary native Type parameters. The native checker sees a Type result,
not an abstract constructor parameter.

Constructor closures can capture an unresolved ordinary Type. For example,
WithError<E> can produce a known closure containing symbolic E without inspecting
E. Applying it can construct the Outcome union normally. Substitution traverses
captures without replacing E by its constraint.

Selection that needs an unknown Type must wait:

```ts
typefn Select<T> {
  type F = typefn<X> { return Array<X>; };
  if (T is string) {
    F = Promise;
  }
  return F<T>;
}

type Selected<T> = Select<T>;
// Select<"x"> -> Promise<"x">; Select<number> -> number[].
```

For unresolved T, retain the Type-valued call Select<T> for later reevaluation.
Do not publish F as an unknown native constructor or guess a branch. The static
kind checker can check the body with an abstract F, but native result publication
requires that any used constructor has become a known declaration/closure.
Unused abstract bindings do not prevent publication of an independent result.

A factory that must inspect an unknown Type before returning a constructor blocks
inside the DSL in the same way. It must not publish an unresolved constructor alias
into native scope. Move the selection and application into a Type-returning typefn,
as Select does. This restriction does not forbid dynamic constructor selection
once its dependencies are available.

## Native boundary and inference

Reject these declarations in this profile:

```ts
interface Functor<F<_>> { /* ... */ }
function map<F<_>, A, B>(/* ... */) { /* ... */ }
type Container<F<_>, T> = F<T>;
```

Existing native HKT encodings still work as ordinary TypeScript. Replacing their
runtime-polymorphic APIs with direct F parameters is outside the restricted scope.
This is a real limitation, not an equivalent spelling of full native HKTs.

An ordinary generic alias must still return a Type. A nongeneric alias can name
a resolved constructor (`type AsyncArray = Compose<Promise,Array>`). Use typefn
for a generic constructor factory; `type Factory<E> = WithError<E>` is excluded.
A nongeneric constructor alias may capture enclosing ordinary Type parameters
provided its constructor identity is resolvable without observing those Types.
It is not a runtime value or the Type of one.

Constructor creation must also finish its own validation checks before publishing
the alias. Merely capturing an ordinary symbolic Type parameter creates no check;
executing an unfinished computation while preparing the capture does.

Bare native generic/default rules from the audit remain: no constructor lifting
by a failed arity check. `type F = Array` reports missing arguments; use a typefn
literal or constructor-valued computation. Assignment to a known constructor
binding, or an explicit constructor argument slot, provides the necessary context.

No inference discovers F, and constructor-kind NoInfer is unnecessary. Once a
known constructor publishes ordinary structure, native inference may use that
structure. Pending DSL computations remain opaque to inverse inference. They
must not fabricate results or force a runtime generic body to be checked only
after concrete calls exist. Native conditional aliases keep native distribution;
imperative typefn branches still test the whole Type unless explicitly iterated.

## Implementation consequences and remaining work

Constructor arguments live in the DSL's signatures, environments and cache keys.
They do not enter ordinary interface/class/signature type-argument vectors. Resolve
a known native constructor, then instantiate it using ordinary native Type
arguments. This avoids the broad proposal's requirement to add constructor kinds
to every native generic argument and variance path.

The native checker still needs the core DSL integration: Type-valued deferred
computations, substitution through captured Types, published native structure,
diagnostics, relation-cycle handling, declaration emit and cancellation. A deferred
Type-valued call carries known code and its environment; it is not a native
constructor variable. This is a substantial reduction in scope, not a claim that
the remaining implementation is trivial or already designed completely.

Compiler/editor declarations preserve typefn bodies and constructor aliases with
their captures. No native F<_> declarations are emitted. Fork-specific declarations
remain necessary; rejecting native HKT parameters does not make the DSL parseable
by stock TypeScript. The previous shared-helper packaging question also remains.

Native operations and closure evaluation still share budgets. Conditional aliases
and VM loops can be expensive even with explicit, known constructors. Performance
must be measured against unmodified native projects and the community workload
corpus; no speedup or completeness claim follows from these restrictions.

## Workload fit and limits

| Investigated need | Restricted scope |
| --- | --- |
| Map/filter/fold Types and construct objects, signatures, unions | Retained in the core DSL |
| String parsers, route types, schema transformations | Unchanged by the HKT restrictions |
| Hotscript-style reusable transformations and composition | Direct constructor application covers this role; not a full API port |
| Effect-style A/E/R transformations | Fixed arities and closures can bind/reorder ordinary Type slots |
| fp-ts-style generic runtime Functor<F> APIs | Keep existing encoding; direct native constructor parameters excluded |
| Passing Compose itself to a higher-order constructor combinator | Excluded; call it directly and pass its returned constructor |
| Passing a bounded constructor | Retained with native bound validation at each actual application |
| Universal bounded-constructor contracts | Excluded; application success is not guaranteed for every argument |
| Reverse inference, constructor subtyping, output bounds | Excluded |

These are design mappings, not passing ports of those libraries. The proposal
retains expressive type computations without promising to replace every native
HKT encoding. Acceptance examples are recorded in [acceptance.md](acceptance.md).
