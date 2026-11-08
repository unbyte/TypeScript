> Historical R1 material, superseded by the [R2 specification](../../spec/README.md). Not the current language contract.

# Parser-first revision: one computation language

Revision candidate prompted by the objection to JavaScript-shaped helpers inside
typefn. This is a substantive alternative to R1's separate data-function surface,
not an approved replacement for all R1 semantics. The agreed restricted HKT scope
is retained. The [calculator rewrite](examples/calculator-parser-first.typefn) makes
the candidate concrete; it is not executable by the current compiler.

## What went wrong

The earlier calculator used local function declarations, let/const, primitive data
annotations, methods, and literal extraction/construction inside a Type-oriented
outer function. R1 intended only a restricted subset, not the whole JavaScript
language. Nevertheless, the example made two expression domains and two user
function models central to an ordinary parser task.

The function keyword alone is not the difficult part of parsing JavaScript. Its
larger costs include contextual grammar, ASI, arrow/function ambiguities, regex
literals versus division, destructuring and overlapping expression/type forms.
We should specify a small DSL grammar directly instead of relying on the reader
or implementer to infer which JavaScript constructs were meant to work.

## The proposed source model

* `type` binds a mutable Type value; parameters and loop/pattern captures remain immutable.
* `typefn` is the only user-defined function form, at module or local scope.
* `F<...>` applies helpers, native constructors and DSL intrinsics. There are no
  user data functions called with parentheses.
* Literal Types support scalar operations directly. Strings, numeric literals,
  boolean literals, tuples and object Types suffice for the parser's state/results.
* The VM may keep scalar literal Types unboxed internally and materialize native
  handles when needed. That is a representation choice, not a second source language.

For example:

```ts
typefn ReadInteger<Text, Start> {
  type cursor = Start;
  type value = 0n;
  type end = text.length<Text>;

  while (cursor < end) {
    type character = text.peek<Text, cursor>;
    type digit = text.digit<character>;
    if (digit < 0n) { break; }
    value = value * 10n + digit;
    cursor = cursor + 1;
  }

  if (cursor == Start) { fail<"Expected a decimal digit">; }
  return [value, cursor];
}

type Parsed = ReadInteger<"123x", 0>; // [123n, 3]
```

The complete rewrite uses the same precedence-climbing algorithm as the earlier
calculator. Its helpers are reusable typefns, return native tuples and pass their
cursor explicitly. They need no embedded function grammar, mutable outer cursor,
data-record return annotation or literal.text/literal.integer bridge.

## A deliberately bounded grammar

Use recursive descent for declarations/statements/patterns and a Pratt parser for
expressions. This is an implementation plan, not a proof that every grammar conflict
has already been removed. Require semicolons and braced control flow; omit ASI,
JavaScript function/arrow expressions, regex literals, destructuring declarations,
async/generators, prototypes/accessors, new, this and runtime statements.

An outline of the parser boundary is:

```text
declaration  := typefn name '<' parameters '>' return-kind? block
statement    := type name '=' expression ';'
              | assignment ';' | call ';' | return expression ';'
              | if '(' expression ')' block (else block)?
              | while '(' expression ')' block
              | for-header block | match-statement
              | break ';' | continue ';' | declaration | block

expression   := literal | name | '(' expression ')'
              | tuple | object | anonymous-typefn
              | application | index | prefix | binary | native-type-island
application  := qualified-name '<' arguments '>'
native-type-island := type '(' native-TypeScript-type ')'
```

This is an outline, not an executable full grammar. Each expression production
has explicit precedence/follow sets; pattern templates have their own restricted
capture grammar. Qualification names a declaration/intrinsic, not arbitrary host
method dispatch. There is no JavaScript parser fallback for an unrecognized body.

Simple literals, tuples, objects, applications and indexing belong to the DSL
expression grammar. Complex native type syntax uses an explicit island, for example:

```ts
typefn Callback<T> {
  return type((value: T) => Promise<T>);
}
```

That island invokes the existing compiler's Type parser, not its JavaScript
expression parser. Native type binding/checking and immutable environment capture
still apply. Conditional/mapped/generic function Types do not have to be reimplemented
as competing productions in the DSL's scalar-expression parser.

### Angle brackets and comparisons need an actual rule

Keeping F<...> means the parser still has to distinguish application from `<`.
The candidate rule is syntactic application lookahead after a qualified name,
using a restricted argument-expression mode. At argument-list depth zero,
comparisons and logical &&/|| are excluded; arithmetic, Type union/intersection,
tuples, indexing and nested applications remain available. Parentheses can enter
the full expression grammar for a deliberately computed predicate argument.

```ts
Expression<Text, cursor + 1, power + 1> // Application with arithmetic arguments.
A < B && C > D                        // Comparisons, not an argument list.
Choose<(A < B)>                       // Explicit predicate argument.
```

`>` closes the current argument list in that restricted mode. Nested application
closers must be split contextually; shift operators are not introduced. Chained
comparison syntax is rejected rather than given a new interpretation. Failed
lookahead must restore token state without consulting symbols or running the VM.
Parser conformance must check precedence, recovery, nested closers and bounded
lookahead work before this rule is called settled. If it is not robust, revisit
the call spelling explicitly rather than silently using semantic parsing.

Application lookahead also needs a valid expression-follow token after its closing
bracket. For example, `A < B > C` cannot become application `A<B>` followed by an
unrelated identifier; it is rejected as a comparison chain. Whitespace does not
change the interpretation. A bounded or memoized recognition strategy must avoid
repeatedly scanning the same nested candidate suffixes. This requirement is not
satisfied merely by choosing recursive descent.

The proposed binary precedence, from tightest to loosest, is `* / %`, `+ -`,
Type intersection `&`, Type union `|`, comparisons (`is == != < <= > >=`), `&&`,
then `||`. Application/indexing bind more tightly than unary operators, which
bind more tightly than binary operators. Comparisons do not chain; grouping is
required. Thus `character is "+" | "-"` tests membership in the union. Logical
operators short-circuit and require boolean literal Types; assignment is a
statement, not an expression. This is a DSL precedence table, not JavaScript's.

### Parser conformance obligations

These are expected outcomes for a future parser, not tests reported as passing.
Every fragment below appears inside a typefn body unless noted otherwise.

| ID | Fragment or workload | Required outcome |
| --- | --- | --- |
| P01 | `return Expression<Text, cursor + 1, power + 1>;` | One application with three arguments |
| P02 | `return A < B && C > D;` | Two comparisons joined by logical conjunction |
| P03 | `return Choose<(A < B)>;` | One application with a grouped predicate argument |
| P04 | `return Outer<Inner<T>>;` | Two applications; both closing brackets consumed correctly |
| P05 | `return A < B > C;` | Reject chained comparison; do not accept a partial application |
| P06 | `return character is "+" \| "-";` | One assignability test whose right operand is a union |
| P07 | `return type((value: T) => Promise<T>);` | Explicit native Type island; no DSL arrow expression |
| P08 | `type cursor = 0` followed by a newline and `return cursor;` | Diagnose missing semicolon; no ASI |
| P09 | `function f() {}`, `type f = x => x;`, `type [a,b] = T;` | Reject unsupported function, arrow and destructuring forms |
| P10 | `return /a+/;` and `return A / B;` | Reject regex literal; accept scalar division |
| P11 | Rename or remove the declarations of A, B and C in P01–P06 | Identical syntactic interpretation; binding errors occur later |
| P12 | Deep nested angle applications and near-matching unclosed lists | Measured bounded work, configured depth limit, local diagnostic recovery |
| P13 | `type X = F<A + 1>;` outside a typefn | Native Type syntax diagnostic; do not enable DSL expressions globally |
| P14 | `return F<T>;` and `return F < T >;` | Same application parse; spacing is not semantic |

Template-pattern lexing, native-island termination and malformed-input recovery
also need cases before the grammar can be considered complete. Patterns and
native Type islands must switch scanner modes explicitly, including when a string
or template contains delimiter characters.

Outside typefn bodies, native TypeScript Type syntax remains the boundary:
`type Result = Calc<"2 + 3">` is a normal-looking generic Type application.
This proposal does not add arbitrary scalar-expression arguments to every native
TypeScript generic application outside the DSL.

## Scalar semantics the rewrite additionally needs

This is more than replacing the word function:

| Area | Proposed change from R1 |
| --- | --- |
| Arithmetic | `+ - * / %` operate on same-family scalar literal Types; number operations use finite binary64, bigint operations exact bounded arithmetic |
| Conditions | Require a boolean literal Type; comparisons produce those Types directly; no truthiness |
| State/results | Type bindings and tuples/objects, rather than JavaScript-shaped data locals and records |
| Native escapes | type(...) is an explicit native-type grammar island, not a Type/data conversion |
| Failure | fail<Message> is an aborting intrinsic application, not the never Type |
| Helpers | Named/module-local typefns with the same application syntax and capture rules as other typefns |

Union/intersection `|`/`&` retain Type meanings, not integer bitwise meanings.
No mixed number/bigint coercion is introduced. Broad or invalid scalar Types
diagnose; unresolved inputs wait. Any scalar implementation must preserve the
existing guards, substitution law, error distinction, cancellation and budgets.
Internal unboxing must not change native Type equality or drop validation metadata.

The example's proposed prelude contracts are small and explicit:

| Intrinsic | Contract |
| --- | --- |
| text.length<Text> | One unbranded string literal to its UTF-16 length number literal |
| text.peek<Text,Offset> | One-unit string literal; empty string at exactly EOF; other out-of-range/nonintegral offsets diagnose |
| text.digit<Character> | ASCII digit to bigint literal 0n–9n; other literal text to -1n |
| numeric.toNumber<Big> | Bigint literal to an exactly representable finite number literal, otherwise diagnostic |
| fail<Message> | Abort with a string-literal diagnostic message and normal call/source provenance |

The sketch retains exact integer division checks, but does not yet reproduce the
old calculator's explicit offset text in every error. A structured diagnostic
intrinsic can address that without reviving a JavaScript data-function language.

## Status and remaining work

The old calculator and its host harness are historical evidence for its algorithm,
not evidence for this grammar. Its 237 host checks cannot be counted as tests of
this rewrite. The rewrite has not been parsed, type-checked or executed.

Before adopting this as the full language, complete the expression/pattern grammar,
resolve metadata/collection APIs under the Type-value model, and update declaration
emit/ABI rules for the changed source language. Native type islands and angle-call
ambiguities need focused parser tests. These are real design obligations, not
details hidden by claiming that a tiny JavaScript subset is easy to support.

The intended direction is one small computation language, preserving the user's
type/typefn model and the agreed restrictions on native HKT integration.
