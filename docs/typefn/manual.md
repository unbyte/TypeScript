# Typefn user manual

R2 language design. The examples describe the target language; an experimental
parser/VM subset is now available. Check [implementation status](implementation.md)
before trying an example. The authoritative rules are in the
[specification](spec/README.md).

Typefn lets you calculate Types with local variables, loops and reusable functions.
You still consume the result as an ordinary TypeScript generic Type.

The implemented slice includes native editor support for navigation, rename, hover,
completion, signature help, semantic highlighting and optional result hints. See
[editor setup and behavior](editor-integration.md) to use the fork in VS Code.

## 1. Your first type function

```ts
typefn Reverse<Items> {
  type result = [];
  for (type item of Items) {
    result = [item, ...result];
  }
  return result;
}

type Reversed = Reverse<["a", 2, true]>;
// [true, 2, "a"]
```

Inside a typefn, `type` declares a local you can reassign. Outside it, `type` keeps
its ordinary TypeScript alias meaning. Typefn parameters are immutable. Each
assignment creates a new value; it does not alter a previously built Type.

```ts
typefn Snapshot<T> {
  type current = T;
  type saved = { value: current };
  current = number;
  return [saved, current];
}
type Saved = Snapshot<string>; // [{ value: string }, number]
```

Use semicolons and braces. There is no let/const, JavaScript function, arrow
expression or runtime function call inside a typefn. Helpers are typefns too.

## 2. Calculate with literal Types

```ts
typefn Clamp<Value, Low, High> {
  if (Low > High) { fail<"Lower bound exceeds upper bound">; }
  if (Value < Low) { return Low; }
  if (Value > High) { return High; }
  return Value;
}
type Clamped = Clamp<120, 0, 100>; // 100

typefn Double<N> { return N * 2; }
type Answer = Double<21>; // 42
```

These inputs are literal Types, so there is an actual value to calculate with.
Double<number> diagnoses: broad number is not one number. Unions do not distribute
through arithmetic automatically. Comparisons return true/false literal Types.

Numbers follow finite JavaScript number arithmetic. Use bigint when integers must
remain exact beyond number precision:

```ts
typefn Next<N> { return N + 1n; }
type Exact = Next<9007199254740992n>; // 9007199254740993n
```

Do not mix number and bigint. numeric.toNumber converts a bigint only when the
result is exactly representable; numeric.toBigint requires an integral number.
Bigint is optional, not a requirement for ordinary type calculations.
Division by zero and nonfinite results diagnose. Bigint division truncates toward
zero; a calculator can explicitly reject nonintegral division.

`A == B` compares native Type identity. `A is B` tests assignability of the whole
Type. Neither compares constructors. `|` and `&` build Type unions/intersections;
they are not integer bitwise operators. Conditions accept only true/false Types.

## 3. Helpers and composition

```ts
typefn Box<T> { return { value: T }; }
typefn Twice<T> { return Box<Box<T>>; }
type Nested = Twice<string>; // { value: { value: string } }
```

Helpers may be module-level or local. A local helper snapshots its outer bindings
when its declaration executes. It cannot change them. Pass state in and return
new state explicitly, often as a tuple. Recursive functions need a result-kind
annotation when their kind cannot otherwise be resolved:

```ts
typefn CountDown<N>: type {
  if (N <= 0) { return []; }
  return [N, ...CountDown<N - 1>];
}
```

The annotation says “returns a Type,” not a particular TypeScript result shape.
Recursive execution has depth/work limits; loops are preferable for long scans.
Module-level helpers can be imported with normal type-only imports:

```ts
// transforms.ts
export typefn Box<T> { return { value: T }; }
// consumer.ts
import type { Box } from "./transforms.js";
type Wrapped = Box<string>;
```

Enable the [implemented subset](implementation.md) through
`compilerOptions.typefn: { "version": 1 }`; full R2 support remains a target.
It lives directly in .ts/.tsx/.mts/.cts
files. Documentation's .typefn listings are examples, not a separate module format.
Typefns and type-only imports emit no JavaScript.

## 4. Unions and tuples

`of` visits tuple elements or string characters in order. `in` visits union members:

```ts
typefn OnlyStrings<T> {
  type result = never;
  for (type member in T) {
    if (member is string) { result = result | member; }
  }
  return result;
}
type Names = OnlyStrings<"Ada" | 42 | "Lin">; // "Ada" | "Lin"
```

Union order is unspecified. Never has no members; any and unknown each have one.
Native normalization may absorb members before iteration sees them.
No operation distributes implicitly, so you control which alternatives stay paired:

```ts
typefn Labelled<T> {
  type result = never;
  for (type member in T) {
    result = result | { original: member; wrapped: [member] };
  }
  return result;
}
// Labelled<"a" | 1> keeps "a" paired with ["a"], and 1 with [1].
```

Tuple iteration requires a finite required tuple, including readonly tuples.
Arrays, optional/rest tuples and unresolved lengths need explicit reflection or
another policy; the loop does not guess a length. The sequence is snapshotted
before the first iteration, so reassigning its source does not change traversal.

## 5. Strings and patterns

Extract the method and route with one anchored match:

```ts
typefn SplitBinding<S> {
  match (S) {
    `${method} ${route}` => {
      return { method: method; route: route };
    }
    _ => { return never; }
  }
}
type Parts = SplitBinding<"GET /users/:id">;
type Method = Parts["method"]; // "GET"
type Route = Parts["route"];   // "/users/:id"
```

Captures are Types. Arms are tried in order; a final `_` is required. Delimiters
match literally, using the first occurrence without backtracking. A match covers
the entire input. Use a pinned expression for a calculated delimiter:

```ts
typefn SplitOnce<S, Separator> {
  match (S) {
    `${left}${=Separator}${right}` => { return [left, right]; }
    _ => { return [S]; }
  }
}
type Pair = SplitOnce<"a::b::c", "::">; // ["a", "b::c"]
```

A pinned separator must be a nonempty string literal. It is never a regular
expression. A single final capture can retain a native template Type:
SplitBinding<`GET /${string}`> returns method GET and route `/${string}`. Splitting
through an unknown template hole is not guessed. Broad string does not guarantee
a space and therefore does not match this pattern. An unresolved parameter waits.

Iterate a literal string directly for character algorithms:

```ts
typefn Characters<S> {
  type result = [];
  for (type character of S) { result = [...result, character]; }
  return result;
}
type Units = Characters<"12+3">; // ["1", "2", "+", "3"]
```

Characters are UTF-16 units. An astral character occupies two units, matching
text.length/peek/slice offsets. Adjacent pattern captures also consume units.
This is not Unicode grapheme iteration.

## 6. Write a parser with a cursor

```ts
typefn ReadInteger<Text, Start> {
  type cursor = Start;
  type value = 0n;
  while (cursor < text.length<Text>) {
    type digit = text.digit<text.peek<Text, cursor>>;
    if (digit < 0n) { break; }
    value = value * 10n + digit;
    cursor = cursor + 1;
  }
  if (cursor == Start) { failAt<"Expected a digit", cursor>; }
  return [value, cursor];
}
type Parsed = ReadInteger<"123x", 0>; // [123n, 3]
```

The tuple carries the parsed value and next position. text.peek returns an empty
string at exactly EOF; offsets beyond EOF diagnose. text.digit returns -1n for
anything except one ASCII digit. The [complete calculator](examples/calculator.typefn)
uses this approach for precedence, parentheses, unary signs and exact division:

```ts
type A = Calc<"2 + 3 * 4">;   // 14
type B = Calc<"(2 + 3) * 4">; // 20
type C = CalcBig<"9007199254740992 + 1">; // 9007199254740993n
```

Its grammar is decimal integers, ASCII space/tab/CR/LF, unary signs, binary
addition/subtraction/multiplication/division,
and parentheses. Leading zeros are decimal. It consumes the entire input and
rejects zero divisors and nonintegral division. It is not a JavaScript expression
evaluator or a floating-point calculator.

For recoverable parsing, return a tagged object Type rather than aborting:
`{ ok: false; offset: cursor; message: "Expected digit" }`. The caller can inspect
its fields or use object patterns. Returning never is also allowed if that is
your chosen API contract. Budget failures are diagnostics, never parser no-match.

## 7. HTTP bindings into parameter and handler Types

The [full HTTP example](examples/http-routes.typefn) composes extraction, validation,
segment matching and dynamic field construction:

```ts
type Endpoint = HttpBinding<"GET /users/:id<number>/posts/:slug">;
// {
//   method: "GET";
//   route: "/users/:id<number>/posts/:slug";
//   params: { id: number; slug: string };
// }
type GetHandler = Handler<"GET /users/:id<number>">;
// (params: { id: number }) => Promise<void>
```

The example accepts uppercase GET/POST/PUT/PATCH/DELETE/HEAD/OPTIONS, one separator
space and an absolute path. Parameters use ASCII identifier names with optional
`<string>`, `<number>` or `<boolean>` annotations. Duplicate names, empty segments,
nonroot trailing slashes, whitespace, query/fragment text and wildcard parameters
diagnose. Percent escapes remain literal; there is no URL decoding. A runtime
router must actually implement the declared conversions; typefn installs nothing.

HttpBindings explicitly maps a union and preserves method/route/params correlation.
One invalid member fails the calculation rather than being silently filtered out.

## 8. Construct native Types and inspect metadata

Common construction is direct:

```ts
typefn Entry<Key, Value> { return { key: Key; value: Value }; }
typefn Pair<A, B> { return [A, B]; }
```

Use `type(...)` for native Type syntax such as functions or mapped types:

```ts
typefn HandlerOf<Params, Result> {
  return type((params: Params) => Promise<Result>);
}
typefn Optional<T> { return type({ [K in keyof T]?: T[K] }); }
```

This is a native **Type** island, not JavaScript execution. Native type expressions
can use current bindings. Compute arithmetic in a local first if it is needed
inside a native generic argument.

Native mapped keys must still be provably assignable to `PropertyKey`. A deferred
calculation with a known key shape supplies that proof. If branching leaves its
result opaque, you can bound it explicitly:

```ts
type Remap<T> = { [K in keyof T as ComputeKey<K> & PropertyKey]: T[K] };
```

Here `ComputeKey` is your key calculation. The intersection bounds the result;
it does not suppress failures when the calculation runs for a concrete key.

For calculated field names/modifiers, build property descriptors:

```ts
typefn Field<Name, Value> {
  return types.object<[
    { key: Name; value: Value; optional: false; readonly: false }
  ]>;
}
type Id = Field<"id", number>; // { id: number }
```

Descriptors are object Types and their lists are tuple Types. No extra data
language is needed. reflect.properties<T> returns key, declared value, read Type,
optional and readonly. Preserve declared value and optionality separately:
an absent optional property differs from a required property containing undefined.
Indexers and call/construct signatures have their own reflection operations.

The [construction examples](examples/construction.typefn) include ReadonlyPublic,
which preserves public optionality, indexers and signatures while making fields
readonly, and EventHandlers, which builds a callback per discriminated event.
These are public structural transformations. Rebuilding public fields does not
clone a class's private identity or create an open interface declaration.

types.function constructs a function from a parameter tuple and result Type;
types.overload combines ordered signature handles. Generic native signatures can
be forwarded unchanged. Building new generic binder lists is outside this version.

Enum-specific APIs are deferred. Existing TypeScript enum Types can still be
passed through unchanged; this version does not construct or specially reflect them.

## 9. Recursive Types

For genuinely recursive shapes, use a guarded builder:

```ts
typefn ListNode<Item> {
  type Build = typefn<Self> {
    return { value: Item; next: Self | null };
  };
  return types.recursive<Build>;
}
```

Self can be placed inside the new structure, but cannot be inspected before the
structure is sealed. Returning bare Self is an invalid unguarded cycle. Ordinary
reassignment never implicitly ties a recursive knot.

## 10. Pass and return type constructors

Array is a constructor; Array<string> is a Type. `F<_>` accepts a constructor
with one Type input and one Type result:

```ts
typefn MapTypes<Items, F<_>> {
  type result = [];
  for (type item of Items) { result = [...result, F<item>]; }
  return result;
}
type Promises = MapTypes<[string, number], Promise>;
```

Compose returns a constructor closure:

```ts
typefn Compose<F<_>, G<_>> {
  return typefn<T> { return F<G<T>>; };
}
type AsyncArray = Compose<Promise, Array>;
type Names = AsyncArray<string>; // Promise<string[]>
```

Closures also express partial application and argument reordering. Use F<_,_> for
two inputs. Native bounds remain attached and are checked at each actual call:
passing Uppercase is allowed; applying it to number fails.

Constructor arguments are explicit. The compiler does not infer F from F<T>,
compare arbitrary constructor domains, or add F<_> to native interfaces/functions.
Bind a returned constructor before applying it; `Compose<F,G><T>` is not syntax.
This restricted model supports reusable calculations while keeping native generic
checking within ordinary Type arguments.

## 11. Generic inputs and errors

Implementation note: native deferred calculations and retained checks are now
available for a tested subset. See [the implementation boundary](deferred-evaluation.md)
for executable examples and the remaining generic-signature limitations. This
section continues to describe the full target contract.

```ts
typefn Box<T> { return { value: T }; }
type Wrapped<T> = Box<T>;
function unwrap<T>(x: Wrapped<T>): T { return x.value; }
```

This works by exposing a known object shape even before T is known. A parser
cannot read an unknown string parameter, and a loop cannot enumerate an unknown
union. Such computations wait for substitution; the compiler does not use a
constraint as though it were the actual input. Native inference sees available
structure but does not reverse arbitrary imperative calculations.

Reached validation is never discarded just because its result is unused:

```ts
typefn Checked<T> {
  type unused = Uppercase<T>;
  return string;
}
// Checked<number> fails even though its visible result would be string.
```

A generic runtime function is checked once. If its body needs an unresolved
requirement, put that requirement in its written signature or prove it with a
native bound. It cannot silently add checks that only run when future callers
specialize the body. [Integration](spec/integration.md) gives the precise rule.

Loops and recursion share finite work/memory limits with native operations.
Exhaustion reports a diagnostic; cancellation ends stale editor work. Neither
means that the requested Type is never. Exported declarations retain executable
typefn recipes and require this fork; stock TypeScript cannot consume the syntax.

## 12. Syntax and API quick reference

| Need | Form |
| --- | --- |
| Local binding / update | `type x = T;` / `x = U;` |
| Helper / application | `typefn F<T> { ... }` / `F<T>` |
| Type constructor parameter | `F<_>` or `F<_,_>` |
| Ordered iteration / union iteration | `for (type x of T)` / `for (type x in T)` |
| Whole-Type test / identity | `A is B` / `A == B` |
| Predicate as call argument | `Choose<(A < B)>` |
| Native complex Type | `type((x: T) => U)` |
| String capture / computed delimiter | `${name}` / `${=Delimiter}` in a pattern |
| Field access | `record["field"]`; dots qualify symbols/intrinsics |
| Explicit diagnostic | `fail<"message">;` or `failAt<"message", offset>;` |

Angle calls and comparisons coexist through the specified syntactic rule. Keep
predicate arguments parenthesized; arithmetic arguments such as F<N + 1> need no
extra parentheses. Whitespace does not change a call into a comparison.

See the [library reference](spec/library.md) for complete string/numeric APIs and
descriptor schemas, and [conformance](spec/conformance.md) for edge cases and
implementation requirements.
