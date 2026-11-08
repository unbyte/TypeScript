# Standard library and structural construction

Normative R2 chapter. Intrinsics are compiler-owned type-level symbols in the
implicit DSL prelude. Their qualified namespaces are `text`, `numeric`, `tuple`,
`reflect` and `types`; `fail`/`failAt` are unqualified aborting operations. A
typefn may not redeclare these roots in its DSL scope. They are not ambient
JavaScript objects and cannot execute host methods. Native modules may use these
names normally; an island uses native name resolution.

All arguments/results have Type kind, except constructor slots explicitly marked
below. There are no descriptor handles to forge. Descriptor records are ordinary
structural object Types containing other Types; lists are required finite tuple
Types. Builders validate them on every use. Intrinsics are not first-class
constructors in R2; wrap one in a typefn to pass it as F<_>.

Signatures below use `= value` only to document intrinsic defaults, not to add
user parameter-default syntax. Missing required/extra arguments are definition
errors. A known invalid operand diagnoses. Unresolved operands produce deferred
Type computations unless control/readiness requires blocking. No intrinsic
silently distributes a union. All retain guards from reached arguments.

## 1. Strings

S, needle and replacement operands below require single unbranded string literals.
Bounds/offsets/counts require nonnegative safe-integer number literals, at most
2^53−1. UTF-16 units are the indexing domain. There is no locale or implicit
Unicode normalization. Exact string results preserve lone surrogates.

| Intrinsic | Result and validation |
| --- | --- |
| `text.length<S>` | Number literal of the UTF-16 length |
| `text.peek<S, I>` | One-unit string literal; `""` at exactly length; I greater than length diagnoses |
| `text.slice<S, Start, End = text.length<S>>` | Half-open slice; require 0 ≤ Start ≤ End ≤ length; no clamping/negative indexing |
| `text.indexOf<S, Needle, Start = 0>` | First unit offset at/after Start, or -1; Start ≤ length; empty Needle returns Start |
| `text.includes<S, Needle>` | Boolean literal; empty Needle gives true |
| `text.startsWith<S, Prefix>` / `text.endsWith<S, Suffix>` | Boolean literal; empty affix gives true |
| `text.concat<Parts>` | Concatenate a finite tuple of string literals |
| `text.split<S, Delimiter>` | Tuple of parts, retaining empty parts; require nonempty Delimiter; nonoverlapping left-to-right split |
| `text.replaceAll<S, Needle, Replacement>` | Literal nonoverlapping left-to-right replacement; require nonempty Needle |
| `text.digit<S>` | One ASCII decimal digit maps to 0n–9n; every other literal string maps to -1n |
| `text.codeUnit<S, I>` | Integer number literal 0–65535; require I < length |
| `text.fromCodeUnit<N>` | One-unit string literal; require integer 0–65535 |

These operations deliberately do not inherit the coercion/clamping/regex overloads
of JavaScript methods. Pattern matching supplies structural template splitting;
these intrinsics require actual finite literal text. Native Uppercase/Lowercase
remain available with their native behavior and Unicode version, including on
symbolic Types; they are not substitutes for locale-aware text processing.

## 2. Numbers

Scalar arithmetic/ordering is in the language, not encoded through tuple lengths.
Number values use finite binary64; bigint values are exact integers with a size
budget. Numeric conversions are explicit and checked.

| Intrinsic | Result |
| --- | --- |
| `numeric.toNumber<Big>` | Bigint literal to finite number literal only if exactly representable; otherwise diagnose |
| `numeric.toBigint<N>` | Integral number literal to the exact bigint represented by that number; fractional input diagnoses |
| `numeric.isInteger<N>` | Boolean for an eligible number literal; bigint is a different family |
| `numeric.isSafeInteger<N>` | Boolean, integral and absolute value ≤ 2^53−1 |
| `numeric.floor<N>` / `numeric.ceil<N>` / `numeric.trunc<N>` | Finite number literal, canonical zero |
| `numeric.abs<N>` | Absolute value, preserving number/bigint family |
| `numeric.format<N>` | Decimal string literal: bigint without n, number uses native ECMAScript shortest round-trip decimal rendering |
| `numeric.parseInteger<S>` | Bigint literal for `[+-]?[0-9]+`; no whitespace, separators, radix prefixes or trailing text; otherwise diagnose |

Parsing a number directly as binary64 is outside the core library; an integer
parser can combine parseInteger and toNumber. `numeric.toBigint<9007199254740993>`
cannot recover the lost decimal unit: the native number literal has already
rounded. Use the bigint literal spelling or parse its string for exact input.

## 3. Tuple state

`tuple.length<T>` returns a required finite tuple's element count, without observing
its element Types. `tuple.slice<T,Start,End=tuple.length<T>>` requires that same
finite shape and strict half-open bounds; it preserves readonly and retained labels.
`[...A,...B]` handles common concatenation. Native indexed access handles selection;
it retains native diagnostics/behavior for invalid indices rather than adopting
the text.peek EOF convention. Optional/open tuples use reflection, not iteration.

Tuple/object state has immutable value semantics. Implementations may optimize
append/update patterns with persistent vectors or unshared transient buffers,
provided snapshots, native identity rules, guards and budgets are preserved.
No claim of constant-time append follows from the source spelling.

## 4. Classification and property reflection

`reflect.kind<T>` returns one of:

```text
"any", "never", "unknown", "stringLiteral", "numberLiteral",
"bigintLiteral", "booleanLiteral", "string", "number", "bigint", "boolean",
"null", "undefined", "void", "uniqueSymbol", "symbol", "tuple", "array",
"union", "intersection", "object"
```

Classification follows this order on native-normalized Types; whole boolean is
classified before its union representation, tuples before arrays. Native enum
Types and enum members are unsupported by reflect.kind and diagnose before
primitive/union classification; they are not implicitly unwrapped. Enum-specific
construction, classification and reflection APIs are deferred. Existing enum Types
can still be forwarded unchanged through ordinary native Type operations.
Callable Types are objects. Unresolved conditional/indexed/
mapped/native parameters wait unless their category is already stable under every
remaining substitution. An unsupported closed native category diagnoses rather
than being mislabeled object. `reflect.isAny<T>`, `reflect.isNever<T>` and
`reflect.isUnknown<T>` return booleans under the same readiness rule.

`reflect.properties<T>` enumerates public apparent properties of one known object
Type, including inherited public properties, excluding private/protected members.
It returns a required tuple of records:

```text
{ key: Key; value: DeclaredType; read: ReadType;
  optional: BooleanLiteral; readonly: BooleanLiteral }
```

Key is a normalized string literal or unique-symbol Type. Numeric property spellings
normalize to the native string key; index signatures are separate. `value` is the
declared write-side Type without the synthetic undefined introduced only by optional
reads; explicitly declared undefined remains. `read` uses current native read rules
and compiler options. Accessors expose their native read/write distinction only
when representable by this record; differing accessor read/write Types diagnose
and require forwarding the original Type. No silent collapse is permitted.

Order follows the native apparent-property enumeration for that checker snapshot.
It is not a portable order for semantically equivalent declarations; algorithms
requiring order must define a key ordering. There is no enumeration of possible
keys supplied by an index signature. Unknown key sets wait. Union/nonobject inputs
diagnose. Field Types can remain symbolic; reflection is shallow.

`reflect.indexers<T>` returns `{ key: KeyDomain; value: Type; readonly: Bool }`
records. KeyDomain may be a native string, number, symbol or supported template
index domain. Keep each native index info distinct; do not pretend an indexer is
a required field. `reflect.calls<T>` and `reflect.constructs<T>` return ordered
tuples of single-signature native callable/constructable Types, with the original
signature's origin, generic binders and identity retained. Overload order is native
resolution order, not union order.

Public property/signature queries are deliberately public views, not proofs that
rebuilding produces the original class/interface. The caller must choose whether
to forward T or construct a structural projection.

## 5. Dynamic objects and interface-shaped Types

`types.object<Properties, Indexers = [], Calls = [], Constructs = []>` constructs a
closed anonymous structural object Type. Property descriptors require exactly:

```text
{ key: Key; value: Type; optional: BooleanLiteral; readonly: BooleanLiteral }
```

The `read` field from reflection may additionally be present; it is derived output
and is ignored on construction. All other unrecognized descriptor fields diagnose
to catch misspellings. Structural intersections may supply the record, but each
control/metadata field must resolve to one eligible literal and must be required.
Property `value` may be any ordinary Type, including any/never or a symbolic Type.

Keys accept one string/number literal or unique-symbol Type, using native key
equivalence. Duplicate effective keys diagnose. Indexers have the schema above;
calls/constructs are tuples of native signatures of the matching category.
The builder validates property/indexer compatibility, overlapping index domains,
readonly flags and native signature legality. It neither silently overwrites
duplicate fields nor intersects incompatible declarations as a fallback.

```ts
typefn Field<Name, Value> {
  return types.object<[
    { key: Name; value: Value; optional: false; readonly: false }
  ]>;
}
```

This creates the shape an interface could describe. It does not declare a named,
open interface or create runtime objects. Public projections deliberately omit
private/protected/class identity and declaration merging. Retaining T or building
`T & Extension` preserves original identity as a native operand; reconstructing
public fields is an explicit structural transformation, not a nominal clone.

## 6. Tuples and arrays with metadata

`reflect.tuple<T>` requires a native tuple and returns:

```text
{ elements: [Element, ...]; readonly: BooleanLiteral }
Element = { value: Type; label: [] | [StringLiteral];
            mode: "required" | "optional" | "rest" | "variadic" }
```

Required/optional element values are declared Types, preserving explicit undefined.
Rest value is the entire array Type; variadic value is the entire tuple/tuple-parameter
Type. Labels are presentation metadata, not property keys. Native shape readiness
determines whether a descriptor can be returned; symbolic element values are allowed.
`types.tuple<Descriptor>` validates native element ordering, rest/variadic legality
and readonly, then constructs the corresponding tuple. This is how transformations
preserve optional/rest information that finite iteration intentionally rejects.

`reflect.array<T>` requires a non-tuple native array, returning
`{ element: Type; readonly: Bool }`. `types.array<Element, Readonly = false>` builds
the matching array Type. It does not inspect an arbitrary class merely because
that class has a numeric indexer.

## 7. Functions, constructors and overloads

For common cases, native islands are sufficient:

```ts
typefn Handler<Params, Result> {
  return type((params: Params) => Promise<Result>);
}
```

`types.function<Params, Result>` builds an ordinary call signature whose parameter
list is described by the tuple Type Params; labels/optional/rest metadata are
retained. `types.constructor<Params, Result, Abstract = false>` builds the analogous
construct signature; Result must satisfy the host's constructor result rules.
`types.overload<Calls, Constructs = []>` builds a callable object preserving order
and original signature handles. It is not an intersection-based reorder heuristic.

`reflect.isGeneric<Signature>` requires one call/construct signature and reports
whether it binds native type parameters. `reflect.signature<Signature>` accepts
one nongeneric signature and returns the record consumed by types.signature:

```text
{
  kind: "call" | "construct";
  parameters: [ { name: StringLiteral; value: Type;
                  mode: "required" | "optional" | "rest" }, ... ];
  this: [] | [Type]; returns: Type; abstract: BooleanLiteral;
  predicate: [] | [{ kind: "is" | "asserts";
                     target: NumberLiteral | "this"; narrowed: [] | [Type] }];
}
```

A rest parameter's value is its whole array/tuple Type. Target is a zero-based
parameter index or this. `is` requires one narrowed Type; `asserts` allows none
for a truthiness assertion. Native predicate compatibility is validated.
Construct signatures must have empty this/predicate fields; calls cannot be abstract.
Names must be native-valid parameter identifiers and unique; native optional/rest
ordering and parameter Type checks apply. `types.signature<Descriptor>` produces
a new ordinary signature, with ordinary function variance behavior, not a forged
method declaration origin. Unchanged native handles returned by reflect.calls/
constructs preserve native method-origin rules when passed to types.overload/object.

Generic signatures are preserved opaquely by reflect.calls/constructs and those
builders. reflect.signature rejects them. Native islands may reference or contain
native generic signatures using existing syntax; algorithmically synthesizing or
editing their binder lists is outside R2. No “last overload only” approximation
is permitted when all signatures were requested.

## 8. Recursive structural construction

`types.recursive<F<_>>` invokes F with a private Type reference Self and seals its
returned Type as the root. A definition that does not use Self is allowed and
returns that ordinary Type. `types.recursiveGroup<Count, F<_>>` invokes
F with a required tuple of Count private references and requires a tuple of Count
root definitions back. Count is a positive safe integer within allocation limits.
The result is the sealed root or tuple of roots. These are intrinsic constructor
slots, not native generic constructor parameters.

```ts
typefn ListNode<Item> {
  type Build = typefn<Self> {
    return { value: Item; next: Self | null };
  };
  return types.recursive<Build>;
}
```

Every cycle in the constructed graph MUST cross an object property, tuple/array
element or signature parameter/result edge. Bare Self, Self unions/intersections
without such a structural guard, and cycles through only aliases diagnose.
The adapter records structural edges through supported native constructors rather
than assuming every generic guards recursion. Reflection, equality, assignability,
branching or scalar operations on an unsealed reference diagnose. Direct native
construction may carry it only where the adapter can preserve and validate those
guarded edges; unsupported native operations diagnose instead of approximating.

Sealing is atomic after complete graph/guard validation. No unsealed reference can
escape through a closure, published Type or cache. Calls whose arguments/captures
contain unsealed references bypass ordinary result caches; their graph-local
scratch state is discarded at sealing or failure. If construction depends on an
unresolved shape, discard the graph and retain the original recursive-builder call
for replay. Symbolic leaf Types are allowed. Cancellation/failure discards the
whole group. This is structural recursion, never a new nominal declaration.
