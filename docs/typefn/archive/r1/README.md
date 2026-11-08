> Historical R1 material, superseded by the [R2 specification](../../spec/README.md). Not the current language contract.

# Typefn, simplified

**R1 design candidate. No implementation is claimed.** Work is on
`typefn-redesign`, based on the upstream compiler. The previous, more elaborate
[R0 specification](../r0/README.md) is archived rather than mixed into this
design.

**Surface revision in progress:** the JavaScript-shaped data helpers below are
being reconsidered. The [parser-first revision](parser-first-revision.md) proposes
one computation language using only type/typefn, with a
[complete calculator rewrite](examples/calculator-parser-first.typefn). It is a
candidate, not an implemented parser or a completed replacement specification.

The basic model is small: **`type` declares a type variable; `typefn` computes
types.** Ordinary TypeScript type expressions work inside the calculation.

```ts
typefn Reverse<T> {
  type result = [];
  for (type item of T) {
    result = [item, ...result];
  }
  return result;
}

type Reversed = Reverse<[1, 2, 3]>; // [3, 2, 1]
```

`type` bindings inside the function can change. Each assignment captures a value;
it does not mutate a previously constructed type or create a recursive alias.
Outside the body, ordinary TypeScript aliases retain their meaning.

## Branches, unions and strings

```ts
typefn Strings<T> {
  type result = never;
  for (type member in T) {
    if (member is string) {
      result = result | member;
    }
  }
  return result;
}
```

`of` visits a finite required tuple or a string literal's UTF-16 units in order.
`in` visits normalized union members without promising an order. `is` tests
assignability of the whole type. Calls and branches never distribute implicitly.

```ts
typefn RouteId<S> {
  match (S) {
    `/users/${id}` => { return { id: id }; }
    _ => { return never; }
  }
}
```

Patterns bind names in their arm. This matches the entire string Type; it does
not search for a substring or silently distribute a union.

For a complete example, see [HTTP binding strings](http-routes.md): extract the
method and route from `GET /users/:id<number>`, construct its parameter and handler
Types, and preserve method/route correlation across unions. The
[full source](examples/http-routes.typefn) includes validation and delimiter matching.

String literal Types can also be traversed directly:

```ts
typefn Characters<S> {
  type result = [];
  for (type character of S) {
    result = [...result, character];
  }
  return result;
}
type Units = Characters<"12+3">; // ["1", "2", "+", "3"]
```

The units are UTF-16 units. For lookahead and arithmetic, compare the
[earlier calculator](string-iteration.md) with the
[parser-first rewrite](examples/calculator-parser-first.typefn). The latter keeps
cursor arithmetic, helpers and results within the type/typefn computation model.

## Type constructors as arguments

`Array` is something that produces a type when given another type. `Array<string>`
is the resulting type. We can make that distinction part of the language:

```ts
typefn MapTypes<T, F<_>> {
  type result = [];
  for (type item of T) {
    result = [...result, F<item>];
  }
  return result;
}

type Promises = MapTypes<[string, number], Promise>;
// [Promise<string>, Promise<number>]
```

`F<_>` means a constructor accepting one type and returning a type. It is not a
placeholder Type such as any. Native generic aliases, interfaces/classes on their
type side, and compatible type functions can be passed directly. Passed
constructors must have ordinary Type parameters and a Type result. Their native
bounds are checked at each application: `Apply<Uppercase,"hello">` can succeed,
while applying Uppercase to number reports a constraint error.

Type functions can return constructors too:

```ts
typefn Compose<F<_>, G<_>> {
  return typefn<T> { return F<G<T>>; };
}

type AsyncArray = Compose<Promise, Array>;
type Names = AsyncArray<string>; // Promise<string[]>
```

The last alias is an extension: it names a constructor, so it must be applied
before use as a value's Type. This supports higher-kinded composition directly,
without URI registries or `this["args"]` encodings.

The agreed scope keeps these constructor parameters inside `typefn`.
Ordinary TypeScript functions, interfaces and aliases continue to have ordinary
Type parameters. A known constructor can still be applied to a symbolic Type:

```ts
typefn Wrap<F<_>, T> { return { value: F<T> }; }
type Wrapped<T> = Wrap<Array, T>; // { value: T[] }
```

There is no constructor inference, constructor subtyping or output-bound syntax.
Constructors passed as F accept ordinary Types and return an ordinary Type;
named combinators such as Compose can organize them and return such constructors.
See the [restricted proposal](restricted-hkt.md) for the limits and rationale.

## Ordinary compile-time data when needed

Numeric literal Types can be compared directly:

```ts
typefn Min<A, B> {
  if (A < B) { return A; }
  return B;
}
type Small = Min<3, 7>; // 3
```

See [numeric comparisons](numeric-comparisons.md) for clamping, explicit union
filtering and ranges. Unresolved numeric inputs wait; comparisons do not infer
a value from a broad number Type.

Parsers and metadata transformations still need strings, counters and lists.
They use familiar `let`, `const` and local `function` declarations. They are
optional tools, not a second form of `typefn` or a new `data` declaration language.

```ts
typefn Repeat<T, N> {
  const count = literal.integer(N);
  if (count < 0) { fail("Expected a nonnegative count"); }
  type result = [];
  for (let i = 0; i < count; i += 1) {
    result = [...result, T];
  }
  return result;
}
```

Reflection and construction are library operations:

```ts
typefn ReadonlyShape<T> {
  let shape = reflect.object(T);
  for (let i = 0; i < shape.fields.length; i += 1) {
    shape.fields[i].readonly = true;
  }
  return object(shape);
}
```

This example changes named fields; it preserves the descriptor's indexers and
signatures. A policy that also makes indexers readonly must edit those explicitly.
For simple cases, ordinary mapped-type syntax is still available.

## The important semantic rule

The evaluator can construct `{value:T}` while T is still generic. It must wait
before making a decision that depends on T. A constraint is not a concrete input.
That allows normal generic code to see a constructed object's fields without
making the VM guess the outcome of an unresolved branch.

The restricted HKT proposal keeps constructors in the DSL. Publishing a native
Type requires known constructor identities, while ordinary Type arguments may
remain symbolic. Unresolved observations defer the Type-valued DSL computation.
This avoids native constructor subtyping and output-bound requirements; core
symbolic evaluation and checker integration still require implementation work.

A call can publish an object's known fields while another field's Type is still
pending. Validation is retained even when a later calculation discards the result.
Specializing a generic computation must agree with running it directly, including
whether it fails. Emitted declarations preserve the code needed to reproduce this.

## Further detail

* [Language and execution rules](design.md): the compact specification and what
  was removed from R0.
* [Restricted higher-kinded types](restricted-hkt.md): the agreed smaller
  scope, useful examples, exclusions and compiler boundary.
* [Symbolic evaluation](evaluation.md): publication, pending checks, replay,
  native generic bodies, diagnostics and caching.
* [Declaration emit](declaration-emit.md): executable source recipes, captures,
  ABI compatibility, dependencies and declaration-only consumption.
* [Semantic cases](semantic-cases.md): concrete expected outcomes for those rules;
  these are design oracles, not passing implementation tests.
* [HTTP route examples](http-routes.md): string literal/template matching, typed
  path parameters, dynamic object/function Types and correlated endpoint unions.
* [String iteration and calculator](string-iteration.md): character Types and a
  complete exact-integer expression parser with precedence and diagnostics.
* [Numeric comparisons and VM capabilities](numeric-comparisons.md): literal
  ordering, arithmetic/data boundaries, ranges and library directions.
* [Broader HKT exploration](../../higher-kinded-types.md): the earlier, more extensive
  proposal, including native signatures and bounds outside the restricted scope.
* [Why TypeScript has no native HKTs yet](../../typescript-hkt-audit.md): maintainer
  discussions, reproducible constraints and corrections to our proposal.
* [Acceptance cases](acceptance.md): examples and counterexamples that must pass
  before implementation claims are made.
* [Research](../../research.md) and [source fingerprints](../../research-sources.json): 39
  inspected files across 17 project families, including fp-ts, hotscript and Effect
  HKT encodings.
