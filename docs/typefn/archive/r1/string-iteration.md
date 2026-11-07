> Historical R1 material, superseded by the [R2 specification](../../spec/README.md). Not the current language contract.

# String iteration and a calculator parser

R1 design extension: `for (type item of T)` accepts either a finite required tuple
or a single string literal Type. The [complete example](examples/calculator.typefn)
shows character Types and an integer expression parser. The compiler has not yet
implemented this syntax.

The JavaScript-shaped calculator helpers on this page are the earlier R1 design.
The [parser-first revision](parser-first-revision.md) and its
[calculator source](examples/calculator-parser-first.typefn) replace that helper
surface in a new candidate. The host harness below validates only the earlier
algorithm; it does not parse or execute the new candidate.

## Iterate character Types directly

```ts
typefn Characters<S> {
  type result = [];
  for (type character of S) {
    result = [...result, character];
  }
  return result;
}

type A = Characters<"12+3">; // ["1", "2", "+", "3"]
type B = Characters<"">;    // []
```

Each iteration binds an immutable string literal Type. `type` keeps its role as
a Type binding; no char keyword, recursive conditional type or string-to-tuple
intrinsic is required. The iterator snapshots the subject once. Reassigning a
local Type variable used as the subject does not change the remaining sequence.

The units are UTF-16 code units, matching this DSL's template-pattern splitting,
data-string length, indexing and slicing. Thus Characters<"💡"> has two surrogate
literal elements. Lone surrogates are retained. This differs from JavaScript's
runtime string `for...of`, which combines surrogate pairs; the new rule concerns
`for (type ...)`, not that runtime statement. Unicode scalar/grapheme processing
requires an explicit conversion/policy, not a silent change of iteration units.

| Type subject | `for (type item of subject)` |
| --- | --- |
| `"abc"` | Three one-unit literal Types |
| `""` | Zero iterations |
| `["ab", number]` | Two Type elements; tuple contents are not flattened |
| `"a" | "b"` | Diagnostic: select a union member explicitly first |
| broad string, an infinite template Type, any, unknown, never | Not a known finite sequence; diagnostic |
| Unresolved S, including S extends string | Wait; do not enumerate the constraint |

String eligibility is the same as literal.text: one unbranded string literal after
native normalization. A template that normalizes to one literal qualifies; a
finite union still needs explicit member iteration. String enum members and branded
intersections are not implicitly converted to their primitive text. Existing tuple
rules for optional/open/variadic elements remain unchanged.

`for (type member in S)` still means union traversal. For the Type "abc" it visits
one member, not three characters. In particular, never is an empty union but not
an empty string or tuple. To map a union of strings, nest character traversal
inside explicit union traversal and keep the outputs correlated.

Pending checks carried by a subject remain attached to the result, even for an
empty string or an early break. Observing an unresolved string blocks the current
call under the [evaluation contract](evaluation.md). Iteration never expands an
infinite string pattern into guessed characters.

## Use data and a cursor for a parser

A parser benefits from lookahead and a cursor more than a character-Type tuple.
Convert the input once with literal.text, use ordinary compile-time functions and
local variables, then construct a native literal Type from the answer:

```ts
type A = Calc<"2 + 3 * 4">;        // 14
type B = Calc<"(2 + 3) * 4">;      // 20
type C = Calc<"-(10 - 4) / 3">;    // -2
type D = Calc<"8 - 3 - 2">;        // 3

type Large = CalcBig<"9007199254740992 + 1">;
// 9007199254740993n
```

The full source contains the parser, not just a hypothetical parseExpression
builtin. It uses precedence climbing: binary + and - bind at 10, * and / at 20,
and unary signs at 30. A recursive right operand receives the binary precedence
plus one, making subtraction and division left-associative. Parentheses start a
nested expression at precedence zero.

Each helper returns `{value:int; next:int}`. The essential call sequence is:

```ts
const right = expression(text, position + 1, power + 1);
// Apply the operator to value and right.value.
position = right.next;
```

The cursor is passed and returned explicitly. Helpers do not mutate an outer
`let position`, which would contradict R1's snapshot captures. Ordinary named
helper self-recursion uses its lexical code binding and the shared call-depth
budget, not a mutable captured function slot. The helpers in this example capture
no mutable outer parser state.

CalcBig ends with:

```ts
const text = literal.text(S);
const result = expression(text, 0, 0);
const end = skipSpace(text, result.next);
if (end != text.length) { error("Unexpected token", end); }
return literal.bigint(result.value);
```

Requiring complete input consumption prevents `2 + 3 garbage` from silently
producing 5. Failed parses report a message and a zero-based UTF-16 offset.
No runtime eval, JavaScript Function call or host callback evaluates the input.

## Arithmetic and grammar are explicit

The example grammar is:

```text
expression := term (('+' | '-') term)*
term       := unary (('*' | '/') unary)*
unary      := ('+' | '-') unary | primary
primary    := digits | '(' expression ')'
digits     := ASCII digit+
```

Space, tab, CR and LF are ignored between tokens. Leading zeros are decimal;
repeated unary signs are allowed. Decimal fractions, exponents, hexadecimal input,
numeric separators, implicit multiplication and exponentiation are excluded.

This is an **exact integer calculator**. Before division, it checks for zero and
requires zero remainder. `7 / 2` is an error, not 3 or 3.5. That is the example's
policy; a rational or floating-point calculator could choose a different data
representation without changing string iteration.

All parsing/evaluation arithmetic uses DSL int. This refinement specifies int
division as truncation toward zero and remainder with the dividend's sign, with
division/remainder by zero diagnosed. The example checks exact divisibility first,
so it never depends on discarded fractional values. Integer growth is bounded by
the VM's bit/work budgets, not silently rounded to machine integers.

CalcBig returns a native bigint literal Type. Calc explicitly converts that result:

```ts
typefn Calc<S> {
  const value = literal.integer(CalcBig<S>);
  return literal(value);
}
```

literal(int) requires a finite, exactly representable native number. Therefore
Calc<"9007199254740992 + 1"> diagnoses instead of returning a rounded number;
CalcBig preserves the answer. Exactly representable integers outside the safe
integer range are not automatically rejected solely for exceeding that range.
This conversion policy is distinct from the parser's arbitrary-precision arithmetic.

## Data intrinsics used by the example

These are pure operations in the interpreter whitelist, not permission to invoke
arbitrary JavaScript string/number methods:

| Operation | Data contract |
| --- | --- |
| `text.length` | UTF-16 unit count as int |
| `text.slice(start, end?)` | UTF-16 substring; int indices, ordinary slice clamping/negative-index rules |
| `text.includes(part)` | Literal UTF-16 substring membership |
| `text.indexOf(part)` | First UTF-16 offset as int, or -1 |
| `integer.toString()` | Exact base-10 text, no bigint suffix or implicit numeric rounding |

String concatenation in the diagnostic combines strings; the offset conversion
is explicit. The digit loop calls indexOf only on a nonempty, one-unit substring,
so indexOf("") cannot accidentally become the digit zero at end of input.

Cursor scanning avoids allocating a native Type per character. The loop itself
advances monotonically and does not backtrack. Big-integer arithmetic is not
constant-cost, and deeply nested parentheses/signs consume helper call depth.
All string, integer and native construction work must be charged and cancellable.
Collecting a tuple in Characters may cost more than merely iterating its input;
no linear-time claim is made for repeated native tuple spreading.

## Symbolic inputs and error behavior

Calc<S> waits when S is an unresolved native parameter. A string constraint does
not provide its characters. Closed broad string and `1 + ${number}` are not single
literals and fail extraction. To process a union, explicitly iterate its members;
neither Calc nor Characters distributes automatically.

Syntax/arithmetic failures use fail, so pending checks cannot be hidden by discarding
the calculator result or wrapping it in an any union. An application diagnostic
retains the original binding and parser operation; the offset identifies the
location inside the literal input. Cancellation and resource exhaustion remain
distinct from invalid arithmetic or syntax.

## Checks and remaining implementation boundary

Run the [algorithm harness](examples/check-calculator-example.mjs):

```sh
node docs/typefn/examples/check-calculator-example.mjs
```

It extracts the **actual CalcBig helper body** from the source example, erases its
annotations with the installed stock TypeScript transpiler, and adapts integer
literals to JavaScript BigInt and string operations to a small UTF-16 wrapper.
The arithmetic parser is not duplicated. Expected generated results are computed
from expression trees, independently of parsing their printed text.

On the installed TypeScript 6.0.3 / Node 24 environment, **19 fixed results,
18 error/offset cases and 200 generated expressions passed**. This validates the
concrete data algorithm under that adapter. It does not validate the DSL frontend,
data kind checker, snapshot implementation, Type iteration, symbolic checking,
native number-Type conversion, VM resource accounting, or declaration emit.
The one-second host timeout is a harness safeguard, not the VM's resource contract.

Additional DSL acceptance cases remain design oracles:

| ID | Input/action | Required result |
| --- | --- | --- |
| S01 | Characters<"12+3"> | ["1","2","+","3"] |
| S02 | Characters<""> | [] |
| S03 | Characters<"💡"> and a lone surrogate literal | Two surrogate literal Types / one preserved surrogate |
| S04 | Character iteration over a local Type binding changed in the loop body | Original subject snapshot, not the replacement |
| S05 | Characters<"a"|"b">, string, any, unknown or never | Sequence diagnostic; no implicit distribution |
| S06 | Characters<S> with unresolved S | Pending; specialization to a literal replays the loop |
| S07 | Iterate `["ab",number]` with `for (type item of ...)` | Two elements; string tuple element is not flattened |
| S08 | Calc<"2 + 3 * 4"> / Calc<"(2 + 3) * 4"> | Native number literal Types 14 / 20 |
| S09 | Calc<"8 - 3 - 2"> / Calc<"16 / 4 / 2"> | 3 / 2; left associativity |
| S10 | Calc<"7 / 2"> / Calc<"6 / 0"> | Non-integral division / division-by-zero diagnostics |
| S11 | Calc<"1 + (2 * 3"> / Calc<"2 + 3 garbage"> | Missing ')' at offset 10 / unexpected token at offset 6 |
| S12 | CalcBig<"9007199254740992 + 1"> | Native bigint literal Type 9007199254740993n |
| S13 | Same input to Calc | Exact native-number representation diagnostic |
| S14 | Calc<S> unresolved / Calc<string> / infinite template | Pending / literal-extraction error / literal-extraction error |
| S15 | Discard a pending Calc<S> result, then substitute invalid syntax | Failure remains attached to the resulting Type |
| S16 | Emit/import Calc/CalcBig with producer sources absent | Helpers, exact arithmetic, offsets and failures preserved |
| S17 | Very deep nesting, very long input or large integer intermediates | Defined resource diagnostics/cancellation; no host stack panic or silent truncation |
