> Historical R1 material, superseded by the [R2 specification](../../spec/README.md). Not the current language contract.

# R1 acceptance cases

These are design oracles, not passing compiler tests. R0's examples and prototype
tests do not establish R1 behavior. Compiler implementation remains paused while
the new surface is reviewed. Constructor cases use the
[restricted HKT proposal](restricted-hkt.md); the broader exploration's native
HKT examples are excluded, not silently counted as acceptance requirements.

The [parser-first revision](parser-first-revision.md) proposes replacing the data
helper surface. Its parser cases and rewritten calculator are additional design
obligations, not passing tests. These R1 oracles must be reconciled explicitly if
that candidate is adopted; existing host-harness results do not validate it.

The [evaluation contract](evaluation.md) and [declaration contract](declaration-emit.md)
define the staged behavior. [Semantic cases](semantic-cases.md) add individually
identified publication, validation, replay, native-body, emit and failure cases.

## A small language should pass small explanations

The introductory guide must explain type variables, type functions, type
application, loops, branches and patterns without requiring descriptor kinds,
native checker flags or a second type-level calling convention. Higher-kinded
examples should need only the distinction between F and F<T>. Advanced data and
reflection contracts remain available without dominating that explanation.

The following constructs are intentionally absent from R1: `data` declarations,
parenthesis-form typefn helpers, `const type`, automatic placeholder currying,
URI registration to pass a generic, and a user-facing extensional constructor
equality operator. Their absence must not be quietly reversed in implementation.

## Concrete and higher-kinded applications

Using the definitions in the guide and restricted proposal, plus:

```ts
typefn Twice<F<_>, T> { return F<F<T>>; }
type NativeStrings<T> = T extends string ? T : never;
```

| Input | Expected result |
| --- | --- |
| `Reverse<[1,2,3]>` | `[3,2,1]` |
| `Reverse<[]>` | `[]` |
| `Reverse<string[]>` | Open-sequence diagnostic |
| `Reverse<[string?]>` | Optional-element policy diagnostic |
| `Strings<"a"|1|"b">` | `"a"|"b"` |
| `Strings<never>` | `never` |
| `RouteId<"/users/42">` | `{id:"42"}` |
| `RouteId<"/posts/42">` | `never` |
| `MapTypes<[string,number],Promise>` | `[Promise<string>,Promise<number>]` |
| `Twice<Array,string>` | `string[][]` |
| `Twice<Promise,number>` | `Promise<Promise<number>>` |
| `Twice<Promise<string>,number>` | Kind error: supplied a Type where a constructor is required |
| `Twice<any,number>` | Kind error; any does not erase kinds |
| `AsyncArray<string>` after `type AsyncArray = Compose<Promise,Array>` | `Promise<string[]>` |
| `Apply<NativeStrings,"x"|1>` | `"x"`; native distribution preserved |
| `Apply<StringBox,"hello">` | `{text:"hello"}`; original bound checked at application |
| `Apply<StringBox,number>` | Native constraint diagnostic at F<number> |
| `Apply<Uppercase,"hello">` | `"HELLO"` |
| `Apply<UpperText,"hello"|1>` | `"HELLO"`; the explicit adapter filters nonstrings |

Bare Promise still cannot be the Type of a runtime variable. A bare defaulted
native alias keeps its existing default instantiation in an ordinary Type position.
The same generic name supplies a constructor when a constructor argument is
explicitly expected. Binary constructors, including ones with a trailing default,
do not silently fit a unary parameter. Explicit closures perform adaptation.

Reject constructor parameters outside the restricted kinds/boundary:

```ts
typefn Higher<H<F<_>>> {}      // Nested constructor input kind excluded.
typefn Bound<F<T extends string>> {} // Bounded constructor contract excluded.
interface Functor<F<_>> {}    // Native constructor parameter excluded.
type Factory<E> = WithError<E>; // Native generic alias must return a Type.
```

Also reject passing Compose itself to F<_>, constructor subtyping/equality,
constructor unions and chained `Compose<Promise,Array><string>` application.
Use a named binding for the last case. Native generic call signatures remain
ordinary Types and do not automatically become constructors.

Test capture snapshots separately from concrete application:

```ts
typefn Capture<T> {
  type F = typefn<X> { return Array<X>; };
  type Saved = typefn<X> { return F<X>; };
  F = Promise;
  return [Saved<T>, F<T>];
}
// Capture<string> = [string[], Promise<string>].
```

Tests must also reject mixed Type/constructor return paths, wrong result kinds,
constructor arity mismatch, constructor values in native union/
tuple/property-Type slots, and application of a nonconstructor.

## Generic and native-code behavior

* `Box<T>` exposes `.value` in a generic function before T is concrete.
* `Wrapped<T> = Wrap<Array,T>` publishes `{value:T[]}`. The ordinary generic
  unwrap example in the restricted proposal checks without knowing T.
* A closure returned by WithError<E> captures symbolic E without observing it;
  later substitution updates the immutable environment correctly.
* A known constructor may yield a blocked Type-valued computation. Native key
  operators must preserve that dependency rather than collapse it to never.
* A caller can construct known outer fields around a pending Type result. Pending
  checks survive discarded locals, projections, erased arguments and any unions.
  Known shape is not permission to drop validation.
* Select<"x"> yields Promise<"x">; Select<number> yields number[]. For unknown
  T, Select<T> defers the whole Type-valued call without publishing an abstract F.
* A factory that cannot select its constructor until T is known cannot publish
  an unresolved constructor alias into native scope. Selection/application can
  be moved inside a Type-returning typefn and deferred there.
* `TextOf<T> = Apply<StringBox,T>` defers when T is unresolved and unconstrained;
  TextOf<number> fails. Applying StringBox to a symbolic T bounded by string can
  publish its result after native validation. Concrete unknown fails the bound.
* Native dependent bounds are checked with actual arguments. A rejected input
  cannot be hidden by a native alias that erases the argument from its result.
* Constructor arguments work explicitly; there is no constructor inference from
  native metadata or structural resemblance. Normal Type inference can use
  ordinary published structure but cannot invert pending DSL computations.
* `Constant<T> = string` prevents treating equality of applications as injectivity.
  `Consume<T> = (value:T)=>void` prevents assuming every constructor is covariant.
* NoInfer preserves ordinary native Type inference behavior. No constructor-kind
  NoInfer extension exists in this profile because constructor inference is absent.
* Known native applications preserve constraints, defaults, conditional
  distribution, variance, alias identity and open interface augmentation.
* Native generic function/class/interface argument slots continue to take Types;
  unapplied constructors are rejected there. Constructor slots belong to typefn.
* Retry blocked inference/relations after substitution. Cache entries for unknown
  applications must not make a later valid concrete relation permanently false.
* An unannotated `type F = RequiredGeneric` reports missing arguments; it does not
  infer constructor kind by failure. Adding a default may make the RHS a native
  Type but cannot silently change a previously valid constructor binding's kind.
* Ordinary Types produced by different constructor declarations retain native
  structural compatibility. Constructor cache identity does not make outputs nominal.
* Generic runtime access to members comes from a published native Type's shape
  or ordinary native constraints. No output-bound promise is attached to F<_>.

The [stock compiler probes](../../probes/hkt-baseline.ts) reproduce the constraints
behind these cases. They validate observations of TypeScript 6.0.3, not DSL/HKT
implementation or the new deferred-check protocol.

## Data, reflection, matching and resources

The [numeric comparison examples](numeric-comparisons.md) add Q01–Q12 for literal
ordering, generic blocking, numeric-family/domain errors and explicit union/range
algorithms. The [source fixture](examples/numbers.typefn) uses Type ordering and
ordinary data arithmetic separately. These are not implementation test results.

The [string iteration/calculator example](string-iteration.md) adds S01–S17 and a
[source fixture](examples/calculator.typefn). String traversal with `for (type ... of ...)` uses UTF-16
units, snapshots its subject, rejects nonliteral/multi-member inputs, and waits
for unresolved arguments. Required tuple traversal retains its previous meaning.
Its host algorithm harness exercises the concrete parser helper body with exact
integers; it is not a DSL frontend, native-Type or VM conformance test.

The [HTTP route example](http-routes.md) adds cases H01–H26 and a
[self-contained source fixture](examples/http-routes.typefn). Check structural
method/route extraction separately from the single-literal validating parser.
Cover typed parameter fields, duplicate-name failure, symbolic input, correlated
unions, handler Types and pinned delimiters. These are design fixtures until the
compiler implements the syntax; they are not reported as executable tests.

Use the researched query-string and restricted JSON cases as the parser corpus,
rewritten with local ordinary functions and native-style data annotations. Preserve
the 15 query cases and 9 JSON-challenge cases as separate scopes; the latter rejects
numbers and is not a standards JSON claim. The R0 source examples are historical
design exercises, not current R1 syntax or executable tests.

Retain adversarial cases for empty captures, anchored matching, repeated delimiters,
pinned delimiters, lone surrogates, UTF-16 versus code points, guard scope and
blocked earlier match arms. Union correlation and ordering policies need explicit
negative tests. Exact optional properties, indexers, unique-symbol keys, signature
order/this/predicates and class identity require metadata round-trip cases.

Run recursive constructor closure/capture/application chains under budgets. Type
normalization is executable work too; kind correctness does not guarantee
termination. Cancellation, cycle errors and failed native validation must clean up
frames, closures, temporary native graphs and in-progress cache states. Measure
cold/warm applications and closure-cache growth, not just one concrete tuple loop.

## Integration and remaining design work

Declaration emit must preserve typefn signatures, constructor result kinds,
anonymous closure dependencies and resolved aliases naming constructors. Native
declarations cannot acquire F<_> parameters. Consume those declarations in a
separate program. Hover, rename,
completion and recovery must understand application heads and constructor slots.
Native projects without DSL/HKT syntax must retain behavior and measured performance.

The restricted proposal excludes constructor output contracts, native constructor
generic relations and constructor inference. The new contracts specify deferred
Type substitution, guarded relations and declaration recipes; these now require
implementation and validation against their semantic cases. Direct sharing of
user-written compile-time data helpers across modules without a parallel
declaration language remains a separate design question.
R1 supports lexical data helpers and reusable typefn constructors now at the
design level; a multi-file parser port must settle shared-helper packaging before
calling the surface mature. Fresh nominal enum generation and runtime-sized
generic signature generation are not current R1 promises. They must not be
implemented opportunistically under syntax that has not been designed.
