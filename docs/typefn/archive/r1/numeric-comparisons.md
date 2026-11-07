> Historical R1 material, superseded by the [R2 specification](../../spec/README.md). Not the current language contract.

# Numeric comparisons and the VM's useful core

R1 design extension. Numeric ordering is natural for the VM: data integers/numbers
already support comparisons in parser loops. This adds the same familiar ordering
notation for numeric literal **Types**. It remains an observation of Type values,
not a new native TypeScript numeric-refinement system.

```ts
typefn Min<A, B> {
  if (A < B) { return A; }
  return B;
}

type A = Min<3, 7>;       // 3
type B = Min<3.5, 2.25>;  // 2.25
```

The [complete examples](examples/numbers.typefn) include Less, Min, Clamp, union
filtering, integer addition and finite range generation. None is claimed as an
implemented compiler feature.

## Keep the meanings separate

| Expression | Meaning |
| --- | --- |
| `A is B` | Native Type assignability |
| `A == B`, `A != B` | Native Type identity/equality rules already specified |
| `A < B`, `A <= B`, `A > B`, `A >= B` | Order the scalar values of numeric literal Types |
| `a < b`, where a/b are numeric data | Ordinary VM data comparison |

The new Type predicates produce boolean data, so they work in if conditions,
guards or boolean data bindings. To return a boolean Type, use `return literal(A < B)`
or the explicit true/false returns shown in Less. There is no implicit conversion
of arbitrary boolean data into a native Type in a return statement.

When one operand establishes a Type-expression context, an adjacent literal uses
that context: `A < 10` compares against the native number literal Type 10. In
`count < 10`, count is an int data binding, so 10 is an int data literal. Mixing
a Type handle and a data variable requires explicit extraction or construction,
such as `literal.integer(A) < count`. No coercion decides between the two.

The frontend resolves this from lexical/static expression categories, not the
VM's result. Generic application lookahead remains syntactic: `F<T>` is an
application, whereas `A < B` is an ordering predicate. Comparison chains are not
range notation: write `Low <= Value && Value <= High`; do not introduce a Python-
style interpretation of `Low <= Value <= High`. Complex operands can be explicitly
extracted to data when the Type-expression spelling would be ambiguous.

## Operand domain and readiness

Both operands must be single, unbranded finite number literal Types, or both must
be bigint literal Types. Number literals use their existing native binary64 values;
bigint ordering is exact. Aliases resolving to eligible literals are fine. Broad
number/bigint, any, unknown, never, multi-member unions, enum members and branded
intersections are not scalar inputs to this operation.

There is no implicit number/bigint mixing. For integral operands, explicitly
extracting both with literal.integer gives exact int data for comparison. Fractional
number values cannot pass that integer extractor. Numeric strings are strings;
parse them deliberately rather than allowing JavaScript-style coercion.

An unresolved operand waits, including a Type parameter constrained to number.
Do not substitute the bound or fold `T < T` to false before establishing that T
is an eligible scalar. Closed invalid operands diagnose. Pending checks on eligible
operand shapes remain attached even if the numeric comparison itself is decidable.
No symbolic boolean is guessed for an if branch.

Native number Types have already undergone native literal parsing. The VM cannot
recover integer digits lost before a Type reached it. Use bigint literal Types
and exact int data for integers that need arbitrary precision. Producing a number
Type through literal(int) still requires exact, finite representation.

## Unions remain explicit

```ts
typefn Below<Values, Limit> {
  type result = never;
  for (type value in Values) {
    if (value < Limit) { result = result | value; }
  }
  return result;
}

type Small = Below<2 | 8 | 15, 10>; // 2 | 8
```

`Less<1|2,3>` does not silently mean all members or any member. Those are different
algorithms. The same applies to Cartesian comparisons between two unions; the
program must choose the quantification and output explicitly. Native union
normalization still happens before iteration.

## Arithmetic uses the existing data bridge

```ts
typefn AddIntegers<A, B> {
  const a = literal.integer(A);
  const b = literal.integer(B);
  return literal(a + b);
}

type Sum = AddIntegers<17, 25>; // 42
```

This keeps arithmetic, parser state, counters and collections in one ordinary
data language. No new Type-level meaning for `+`, `-`, `*` or `/` is added here.
In particular, native `A | B` and `A & B` remain union/intersection, not integer
bit operations. literal.bigint produces exact bigint results when desired.

The range example deliberately produces number literal Types, even if its integer
inputs were extracted from bigint literals. It can fail if a produced value is
not exactly representable or if the union exceeds resource limits. It is not an
unbounded interval Type or an instruction to silently approximate the result by
number. Changing the output constructor to literal.bigint gives a different API.

## A small set of primitives enables many libraries

| Building block | Algorithms it enables |
| --- | --- |
| Exact arithmetic, numeric ordering and loops | Ranges, bounded dimensions, counters and arithmetic evaluators |
| String access, matching, functions and cursor records | Route grammars, query parsers, calculators and DSL parsers |
| Data arrays/records/maps/sets and iteration | Token streams, deduplication, grouping and sorting with an explicit comparator |
| Type reflection and metadata-preserving builders | Schema transformations, field filtering, object/tuple/function-Type construction |
| Explicit union traversal | Filtering, correlation-preserving transformations and chosen all/any checks |
| Constructor arguments and closures | Reusable transformations and partial application within the agreed HKT scope |

These are library directions enabled by the design, not newly implemented APIs or
claims that every collection intrinsic is already specified. Keep new algorithm
features primarily in ordinary functions and the prelude; add language syntax
only when a primitive cannot express the intended operation clearly.

Running inside the checker supplies native Types and native operations. It does
not supply the value of an unresolved parameter, make arbitrary type programs
invertible, or permit inspection of a half-completed inference session. The
existing deferral, validation, capture, declaration and resource rules continue
to apply. Fresh nominal type generation and fabricated generic binders are not
added by this comparison feature.

## Required design cases

| ID | Case | Expected |
| --- | --- | --- |
| Q01 | Less<-3,2> | true |
| Q02 | Min<3.5,2.25> | 2.25 |
| Q03 | Clamp<120,0,100> | 100 |
| Q04 | Clamp<5,10,0> | Invalid-bound diagnostic |
| Q05 | Below<2|8|15,10> | 2|8 |
| Q06 | Less<1|2,3>, Less<number,3>, Less<never,3> | Scalar-operand diagnostics |
| Q07 | Less<1,2n> | Numeric-family diagnostic; no implicit conversion |
| Q08 | Less<9007199254740992n,9007199254740993n> | true, without binary64 conversion |
| Q09 | Less<T,T> with unresolved T | Pending, not a folded false that hides invalid operands |
| Q10 | `A <= B`, `A >= B` for equal eligible literal values | true; strict comparisons false |
| Q11 | RangeNumbers<0,4> / RangeNumbers<4,4> | 0|1|2|3 / never |
| Q12 | Numeric observation of an eligible literal shape carrying a pending check | Keep the check through control flow and returned result |

These cases are design oracles. The calculator's host harness exercises numeric
data operations; it does not test these proposed native Type predicates.
