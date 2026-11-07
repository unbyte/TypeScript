# Semantics

Normative R2 chapter. “Type” is a native TypeScript type value, possibly symbolic,
guarded or deferred as defined in [evaluation](evaluation.md). Lowercase `type`
is syntax. A string literal Type is itself a usable compile-time string operand;
there is no extraction into a separate JavaScript value domain.

## 1. Values, kinds and environments

```text
Kind        ::= Type | Constructor(n)
Value       ::= TypeResult | ConstructorValue
Constructor ::= known native generic | eligible typefn code + frozen captures
Signature   ::= ordered parameter kinds + one result kind
Environment ::= lexical binding identity -> immutable Value
Frame       ::= environment + local slots + reached-check journal
```

Constructor(n) accepts exactly n ordinary Types and returns an ordinary Type.
It has no recursively nested input/output kinds or subtyping relation. A named
typefn may have a mixture of Type/constructor parameters and may return a constructor,
but only Type-to-Type functions are eligible first-class constructor values.
Constructor values cannot occur as native object fields, tuple elements or unions.

A Type binding has a fixed kind but no fixed narrower Type: `type x = 1; x = "a";`
is valid. Every assignment replaces the local slot; it cannot mutate an already
built Type, capture or another frame. Parameters, pattern/iteration bindings,
function names and captures are immutable. Capturing a local reads its value then,
not its slot. Assigning a field/index is outside the grammar.

Duplicate declarations, read-before-initialization and assignments to immutable
bindings diagnose during definition checking. Nested blocks may shadow names;
local initializer self-reference diagnoses, rather than creating a recursive alias.
Native declarations outside the DSL retain native scoping rules.

Module-level functions are available by lexical binding, including mutual recursion.
A local named typefn becomes available at its declaration and snapshots its free
bindings there. Its self-reference resolves its own code, not a mutable slot.
References to later local helpers are rejected; use module helpers for mutual
recursion. Anonymous typefns capture at expression evaluation. Capturing alone
does not execute a body or inspect the Types in its environment.

## 2. Definition checking and calls

All bodies are parsed, bound and kind-checked, including unreachable branches.
Definition checking validates names, arity, kinds, scopes, legal control transfers,
and return-kind consistency. Scalar eligibility, dependent native constraints and
descriptor validity are checked when their operations execute. No unreachable
branch executes merely to validate its eventual Type operands.

Typefn parameters have no defaults, rest or implicit arguments. Type parameters
may have `extends` bounds depending on preceding Type parameters; bounds are Type
expressions evaluated in that environment. Constructor parameters specify arity
only. Named functions may infer their result kind without evaluation; a recursive
call-graph component must have sufficient explicit result-kind annotations to
resolve every member. All returns must agree in kind/constructor arity. A path
that reaches the end diagnoses; a path that aborts has no result kind.
If a body has no value-returning path, it needs an explicit return-kind annotation
unless an enclosing constructor slot already supplies that kind. Definition
checking uses conservative control flow; it does not run loops to prove termination
or infer a kind.

Calls evaluate arguments left to right, check kinds/bounds, then execute the callee.
Passing native `Array`, `Promise`, an alias or a typefn to an explicit constructor
slot supplies its constructor identity. Declared native bounds are validated at
each application, not universally compared against the receiving slot.

```ts
typefn Apply<F<_>, T> { return F<T>; }
typefn Compose<F<_>, G<_>> {
  return typefn<T> { return F<G<T>>; };
}
type AsyncArray = Compose<Promise, Array>;
type Names = AsyncArray<string>; // Promise<string[]>
```

Constructor application supplies all declared slots, including slots with native
defaults. Direct named native generic applications retain native default behavior.
Bare generics never become constructors because native instantiation failed:
`type F = Array;` diagnoses. A constructor slot, an existing constructor binding,
an explicit constructor return kind, a closure or a factory supplies kind context.
A bare defaulted native alias in an ordinary Type position remains its default
instantiation. This rule prevents added defaults silently changing a binding's kind.

A bare eligible named typefn already denotes a constructor; `type F = Box;` can
therefore bind it without a wrapper. Named combinators with constructor inputs or
results can be called by name, but cannot be stored as first-class constructors.
Native generic names need the explicit kind context described above because they
also have native Type-position/default semantics.

Only a nongeneric native alias may name a constructor, and only when its identity
and creation checks are ready. Generic factories use typefn. No native class,
interface, alias or runtime function can declare `F<_>` parameters. Ordinary
native generic arguments never contain constructor values. Known closures may
capture unresolved ordinary Types; no abstract constructor crosses that boundary.

## 3. Construction and native identity

Simple object, tuple and template syntax constructs native Types. `|`, `&`, `keyof`
and indexed access use native semantics. A computed field key must be one string
or number literal or unique-symbol Type; unresolved keys defer construction.
Duplicate effective keys diagnose, including numeric/string spellings of one key.
An object expression is structural and does not merge declarations.

`type(NativeType)` binds the host Type expression in the current immutable lexical
environment and applies native semantics. The native grammar itself is unchanged:
conditional `infer`, mapped types, `typeof`, function types and existing native
generic signatures work inside the island. A type query reads a runtime symbol's
static Type, never its runtime value. Native conditional distribution remains native;
DSL control flow never acquires implicit distribution.

Construction preserves native private/unique-symbol/enum identities when their
Type handles are retained. Native normalized unions can absorb members and
intersections can reduce to never. Type equality is the host's identity relation,
not spelling, mutual assignability or pointer equality alone. An implementation
must not clone nominal declarations by reconstructing printed shapes.

Every reached operand's validation survives normalization/projection. Constructing
`any | Checked<T>` cannot erase a failed Checked specialization.

## 4. Scalar operations

An eligible scalar is one unbranded native literal of the required family after
native normalization. Enum members and branded intersections are not silently
unwrapped. Broad primitives, multi-member unions, any, unknown and never are not
scalar operands. An unresolved Type defers; its bound is not a sample value.

| Operation | Domain and result |
| --- | --- |
| `+ - * / %` | Two number literals or two bigint literals, same family; result in that family |
| Unary `+ -` | One numeric literal; preserve family |
| `< <= > >=` | Two same-family numeric literals; boolean literal Type |
| `== !=` | Two ordinary Types; native identity / its negation; boolean literal Type |
| `is` | Two ordinary Types; whole-Type native assignability; boolean literal Type |
| `!` | One boolean literal Type; negation |
| `&& \|\|` | Boolean literal Types, left-to-right short-circuit |

Strings concatenate through text.concat or templates, not `+`. There is no
truthiness, numeric/string coercion, bitwise integer operation or mixed-family
arithmetic. Constructor equality and ordering are not defined. For `any`/`never`,
`is` follows the native relation rather than inventing set-theoretic behavior;
algorithms can classify these Types explicitly before testing assignability.

Number arithmetic uses correctly rounded binary64 operations, round-to-nearest
ties-to-even, with no fused reassociation. Results must be finite. Both numeric
families diagnose a zero divisor for `/` and `%`. Number remainder follows
truncating-quotient remainder. Bigint division truncates toward zero and remainder
has the dividend's sign. Bigint operations are exact within the resource budget.
Native number Type identity canonicalizes negative zero to zero; this occurs at
each result, even if the VM keeps unboxed scalars. NaN/infinity cannot be Types
produced by these operations. Number rounding is otherwise ordinary binary64;
exact integer computations should use bigint.

```ts
typefn Double<N> { return N * 2; } // Double<21> is 42.
typefn Exact<N> { return N * 2n; } // Exact<21n> is 42n.
```

Operand eligibility is checked before arithmetic simplifications: `0 * T` and
`T < T` cannot ignore an unknown or invalid scalar. Identity `T == T` can return
true from the same immutable Type term while retaining its checks. Other symbolic
identity and assignability observations conservatively defer, except where the
evaluation chapter explicitly permits structural independence.

## 5. Sequential control and iteration

Statements execute in source order. `return` ends the current call. `if` and
`while` require exactly true or false literal Types; a pending condition blocks
the observing call before either branch executes. A known invalid condition
diagnoses. No branch narrowing changes a parameter or local's Type value;
patterns introduce captured values, and bounds are checked independently.

An iteration evaluates and snapshots its subject once:

* `for (type item of T)` visits a finite tuple's required elements or a single
  string literal's UTF-16 units in order. Readonly tuples are allowed. Optional,
  open/rest/variadic tuples, arrays, broad strings and unions diagnose. Elements
  may themselves be symbolic Types. Empty tuple/string produces no iterations.
* `for (type item in T)` visits normalized union members with unspecified order.
  Never produces zero iterations; any and unknown each produce one. For R2, any
  reachable free native Type parameter or deferred term makes union traversal
  wait, even within a known discriminant's field. No constraint is enumerated.

Reassigning the original subject local does not alter the snapshot. Loop bindings
are immutable and fresh each iteration. `break`/`continue` target the nearest
lexical loop; a match arm is not a loop. Neither may cross a function boundary.
Union order must not be used as a portable sort order. A tuple built from it may
have any member order, including across compiler versions; semantic validity
and the accumulated checks must be independent of traversal shortcuts.

## 6. Pattern matching

Evaluate the subject once, then try arms in order. Patterns do not distribute a
union. Match has three semantic outcomes: match with captures, definite no-match,
or pending; domain violations are diagnostics. Pending in an earlier arm blocks
the call rather than falling through. A successful pattern evaluates its guard;
false tries the next arm, true executes its block. A guard may call typefns but
cannot mutate outer state. Reached checks survive a false guard. A selected arm
does not fall through; execution resumes after match if its block does not return.

| Pattern | Meaning |
| --- | --- |
| `_` | Match any Type without observing its shape |
| `name` | Capture the whole Type immutably |
| Literal | Native identity with that literal Type |
| `type(T)` | Native identity with the bound native Type expression |
| `is T` | Whole-Type assignability |
| Tuple | Required finite tuple decomposition; exact length without rest; rest preserves labels/readonly |
| Object | Required public properties must exist; extra properties allowed; recurse on each property's declared Type |
| Template | Anchored deterministic string decomposition below |

Tuple/object patterns do not project union alternatives. A union subject for
either diagnoses; use explicit iteration or an `is` arm. A known non-tuple or
nonobject subject is a no-match. Optional/open tuple layouts and optional object
properties are no-match for required structural patterns. No index signature
alone guarantees that a named required property exists. Generic signatures and
private members are not exposed by object patterns.

Duplicate effective object-pattern keys diagnose. Object patterns visit their
fields in source order, retaining checks from every reached subpattern. A failed
subpattern does not publish its captures to later arms.

### String templates

1. A subject is one unbranded string literal, native template Type, or broad string.
   Unresolved binders wait. A union or branded string subject diagnoses; known
   nonstring Types do not match. Never is a nonstring here, not an empty input.
2. Match the entire subject. Literal chunks (including decoded escapes) are exact.
   Pinned `${=E}` chunks require a nonempty single string literal Type; evaluate
   pins in source order when the arm is attempted. They are literal delimiters,
   never regexes. Invalid pins diagnose, even if a later chunk would not match.
3. On a literal subject, the capture preceding a nonempty delimiter receives the
   prefix ending at the earliest occurrence of that delimiter. Consume it and
   continue. Never backtrack if a later chunk fails. Anchored initial/final chunks
   still have to match; a final literal delimiter is not a last-occurrence search.
4. Adjacent captures give one UTF-16 unit to each nonfinal capture and the remainder
   to the last. Empty chunks do not separate captures. A sole capture matches empty
   input. `${a}${b}` requires at least one unit. `${_}` consumes identically but
   does not bind a name. Literal tails may leave captures empty.
5. For a normalized template subject, preserve its holes as opaque native Type
   fragments. A final capture can take the entire remaining fragment sequence.
   Prefix tests and delimiter searches may consume known literal fragments only;
   do not search through an opaque hole or take one unknown UTF-16 unit from it.
   If the required split is not determined by that rule, this *closed pattern
   Type* is a no-match. This is structural matching, not existential matching over
   possible strings. Broad string is one opaque hole: a sole capture matches it.
   A free binder is different and waits rather than being classified no-match
   when its contents are needed. A final capture may retain a symbolic template
   suffix without observing it; for example, a known `GET ` prefix can be removed
   even when the remaining template contains a free Type parameter.

```ts
typefn Binding<S> {
  match (S) {
    `${method} ${route}` => { return [method, route]; }
    _ => { return never; }
  }
}
// Binding<"GET /users"> -> ["GET", "/users"]
// Binding<`GET /${string}`> -> ["GET", `/${string}`]
// Binding<string> -> never; Binding<S> with unresolved S waits.
```

Captures are Type values, not host strings. Template expressions (construction)
use native template interpolation semantics; template patterns use the rules above.
The latter intentionally specify UTF-16 units independently of native conditional
template inference, whose Unicode behavior can differ in the pinned Go compiler.

## 7. Failure and effects

`fail<Message>` aborts with a diagnostic. `failAt<Message, Offset>` additionally
records a nonnegative safe-integer UTF-16 offset in the current input, supplied by
the parser author. These aborting intrinsics are permitted only as application statements;
their operands must be ready before failure. They have no successful return value.

Returning never is successful computation of the empty union. Parser recovery can
use object Types such as `{ ok: false; offset: 3; message: "Expected digit" }`.
Neither form catches a diagnostic, resource exhaustion or cancellation. There is
no exception-catching facility. All effects are local bindings plus internal
validation/resource accounting; the language has no runtime/environment access.
