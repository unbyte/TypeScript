> Historical R1 material, superseded by the [R2 specification](../../spec/README.md). Not the current language contract.

# R1 language and execution rules

R1 replaces R0's surface design. This is a specification candidate, not an
implemented language. The short [guide](README.md) is the intended starting point;
this file defines the rules behind its examples.

The data-helper surface described here is under revision. See the
[parser-first proposal](parser-first-revision.md) for a single type/typefn
computation language. Its direct literal-Type arithmetic and boolean-Type
conditions are proposed changes to these rules, not already compatible additions.

The [restricted HKT design](restricted-hkt.md) defines the agreed
constructor scope. The broader HKT document is an exploration, not an additional
set of requirements for this profile.

## Why R0 was difficult to understand

R0 introduced a separate data-kind language, a `data` keyword, two typefn declaration
forms, two calling conventions for one type function, special immutable Type
bindings and a large descriptor vocabulary before demonstrating everyday use.
It also used parsing/binding details to explain ordinary assignments. Those are
too many concepts to require at the entry point.

R1 keeps the necessary distinction between Types and ordinary data, but removes
the extra public kind language and the second form of typefn. Type functions use
angle-bracket application everywhere. Ordinary data helpers use ordinary function
syntax. Reflection is a library used when needed, not the mandatory spelling of
object, function and tuple construction.

| R0 surface | R1 choice |
| --- | --- |
| `typefn F<T>` and `typefn f(x: Kind): Kind` | One typefn form; local `function` for data helpers |
| `F(T)` and `F<T>` for one type function | `F<T>` consistently |
| New `data` aliases and List/Result kind syntax | Familiar TypeScript data annotations and records/arrays |
| `const type` | Omitted; local `type` is mutable, Type values themselves are immutable |
| Native generics require wrappers to become function values | Compatible native constructors pass directly |
| Higher-order support primarily through data closures | Constructor parameters and returned typefn closures |
| Flat exhaustive API introduction | Small language guide, separate advanced contracts |

Removing notation does not remove optional-property distinctions, generic blocking,
identity or resource limits. Those remain obligations of the implementation.

## Bindings and functions

```ts
typefn Box<T> { return { value: T }; }
typefn WrapTwice<T> { return Box<Box<T>>; }
```

`typefn Name<Parameters> { Statements }` declares a compile-time function. Parameters
normally accept native Type values. Constructor parameters use `F<_>` or another
fixed number of ordinary Type slots, as specified in the restricted proposal.
Only typefn declarations may bind constructor parameters. Parameters are immutable.
A function may return a Type or a constructor; all its return paths must have one statically
compatible kind. That kind is inferred without executing the body. Recursive
return-kind inference must have an anchor or an explicit annotation.

Inside the body, `type name = Expression` introduces a mutable type-level binding.
It can contain a native Type or a constructor, but its inferred kind is fixed:
rebinding a Type variable to a constructor is a kind error. Reads before
initialization, duplicate local declarations and assignment to parameters fail.
Nested scopes may shadow outer names. A self-reference in a local initializer
is not a recursive alias; recursive structural construction is explicit.

```ts
typefn Snapshot<T> {
  type current = T;
  type saved = { value: current };
  current = number;
  return [saved, current];
}
```

`Snapshot<string>` is `[{value:string},number]`. Lazy native types retain immutable
environment snapshots, never pointers into mutable VM slots. Returned constructor
closures also capture values at creation, including captured local constructors.

Top-level TypeScript aliases remain declarative. Their RHS may additionally denote
a constructor, as in `type AsyncArray = Compose<Promise,Array>`. Such an alias has
constructor kind and cannot serve directly as the Type of a runtime value. Such
aliases must be nongeneric and resolve a known constructor; generic factories use
typefn. Ordinary aliases whose RHS is a Type are unchanged. The treatment of
generic defaults is specified in the restricted proposal: a missing native type
argument never implicitly converts a bare generic to a constructor. An unannotated binding
uses a typefn literal, an existing constructor value or a constructor-valued
application to establish constructor kind. This prevents an upstream added default
from silently changing the binding's kind.

## Expressions and control flow

Type initializers, assignments to type bindings and returns from typefn use native
type syntax, extended with constructor values/application and Type-returning
library calls. Native conditional infer, mapped types, indexing, unions,
intersections, templates and concrete signatures remain available.

Data bindings/assignments, local ordinary-function bodies and library call
arguments use data syntax. A Type binding can be passed as a handle. `type(...)`
embeds a native expression when data syntax cannot express it; `literal(value)`
creates a native literal Type from computed data. There is no conversion of an
arbitrary data object to a native object Type by resemblance.

```ts
const name = "id";
type Key = literal(name);
let field = { key: Key, type: T, optional: false, readonly: true };
field.type = type(Promise<T>);
```

Library calls such as object/tuple accept data descriptors and return Types.
Calls to typefn use `<...>` even within data expressions: there they produce an
opaque Type/constructor handle. Parsing a type-argument application uses syntactic
lookahead like TypeScript's instantiation-expression grammar, never VM evaluation.
Data helpers use `helper(...)` and cannot call runtime JavaScript functions.

Conditions accept boolean data, plus whole-Type tests `A is B`, `A == B`, `A != B`.
Numeric literal Types additionally support ordering predicates `<`, `<=`, `>`
and `>=`, returning boolean data. Both operands must be eligible literals of the
same numeric family; unresolved values wait and broad/union operands are not
silently quantified. [Numeric comparisons](numeric-comparisons.md) specifies the
domain, literal contexts and explicit data conversions.
`is` uses native assignability; equality uses native identity, not mutual
assignability or extensional constructor equivalence. Constructor equality is
not a user operation. In a Type comparison both operands are type expressions;
in a data comparison both are data expressions. The binder classifies an operand
rooted in a Type binding; explicit `same(A,B)`/`assignable(A,B)` library calls
handle descriptor projections or other mixed syntactic positions. No truthiness
or implicit Type/data boolean conversion is allowed.

The frontend uses dedicated DSL nodes/cover expressions and lexical binding forms
to elaborate these contexts. It must not replace malformed bodies with an ordinary
JavaScript AST or execute a checker query to decide the grammar. Ambiguous nested
Type expressions in data syntax use explicit `type(...)`.

Statements are bindings, assignment, return, blocks, if/else, while, for, match,
break, continue and local helper/function declarations. Braces are required for
control flow. Semicolon insertion follows TypeScript; canonical output includes
semicolons. No exception handling, labels, generators, async or runtime statements
occur inside typefn. Reaching its end without a result diagnoses.

* `for (type item of T)` snapshots either a finite tuple's required elements or
  one unbranded string literal's UTF-16 units, in order. String elements are
  one-unit string literal Types. It rejects optional/variadic/open tuples,
  broad/infinite string Types and unions instead of guessing or distributing.
  An unresolved subject waits; empty tuple/string gives zero iterations.
* `for (type member in T)` snapshots normalized union members. Never gives zero
  iterations; any/unknown are singleton special Types. Order is unspecified.
* `for (const item of data)` and numeric `for (let i=...; ...; ...)` process data.

Type iteration bindings and pattern captures are immutable. `break`/`continue`
target the nearest loop, including from match arms. Union accumulation is explicit;
normalization can absorb members before the program sees them. Building an ordered
tuple from union traversal does not give a portable ordering guarantee.

## String and structural match

The [HTTP binding example](http-routes.md) applies these rules to method/route
extraction and typed path parameters; [its source](examples/http-routes.typefn)
also demonstrates computed literal delimiters and explicit union processing.

```ts
match (S) {
  `${head}:${tail}` => { return [head, tail]; }
  _ => { return never; }
}
```

Subject evaluation happens once. Arms are tried in order, with no fallthrough.
Captures are immutable and local to the arm. Duplicate captures are errors. A
final wildcard is required for Type matches; finite discriminated data matches
may prove exhaustiveness. Guards are boolean and cannot mutate caller state.
An unresolved earlier arm blocks instead of falling through to a later wildcard.

Templates match the whole string with literal delimiters, using the earliest
delimiter occurrence without backtracking. `${name}` captures, `${_}` discards;
`${=separator}` inserts a computed data-string delimiter. Adjacent captures give
one UTF-16 unit to each nonfinal capture and the remainder to the last. A sole
capture accepts empty input; two adjacent captures require at least one unit.
Use explicit code-point iteration for Unicode scalar algorithms.

On string data, captures are strings. On a native string Type, captures are Types,
possibly template Types such as `${number}`. Broad string matches an unrestricted
single capture but no specific prefix. A union is processed explicitly; a match
does not claim that all possible runtime values are one literal. Branded string
intersections require explicit normalization and retain their policy boundary.

List patterns `[head, ...tail]` decompose data arrays or finite required native
tuples; rest captures preserve native tuple readonly/label metadata. Object
patterns `{key: value}` capture required public fields and allow extra fields.
Optional/open/native signature structure is handled by reflection. Literal
patterns compare corresponding data values or literal Types. `type(T)` patterns
use native identity; `is T` patterns use whole-Type assignability. Data tagged
record patterns support parser-result narrowing.

## Supporting data code

`let`, `const`, arithmetic, records, arrays and local `function` declarations are
a deterministic TypeScript-like subset. Annotations use familiar TypeScript
notation, including `Type` for the opaque checker handle and imported prelude
descriptor types. There is no separate user-facing `data`, Kind, List or Result
language. Tagged records express results in the usual style:

```ts
typefn Parse<S> {
  function skipSpace(text: string, position: int): int {
    while (position < text.length && " \t\r\n".includes(text[position])) {
      position += 1;
    }
    return position;
  }
  const source = literal.text(S);
  const start = skipSpace(source, 0);
  // This small entry demonstrates a data helper, not a complete parser.
  return literal(source.slice(start));
}
```

Ordinary data-helper parameters are local writable copies, as in ordinary function
code; typefn parameters remain immutable. Helpers/lambdas capture outer values by
snapshot and cannot mutate captured bindings. Lists/records have value semantics:
updating one local never changes a saved copy. Mutable operations require a local
root; no mutation through native handles, shared state or host objects is possible.

Data type annotations are checked with native TypeScript's familiar type vocabulary
against a restricted interpreter-value ABI. They are not permission to execute
arbitrary values of an annotated native type. The ABI permits finite records,
arrays, tagged unions, scalars, Maps/Sets, closures and opaque Type/constructor/
descriptor handles. Any, unchecked casts, arbitrary classes/prototypes, accessors,
host callbacks and ambient runtime calls are rejected. Runtime intrinsic dispatch
also checks handle tags, so a structural lookalike cannot forge a native handle.

Integer literals have exact `int` arithmetic, with a bit budget. `number` is finite
binary64, with checked conversion between the two. Strings use UTF-16, including
lone surrogates. No coercion, implicit narrowing to machine integers or regex
backtracking is introduced. Native numeric Type literals keep native semantics.
This is a deliberate small difference from JavaScript, not a full JS interpreter.

Int division truncates toward zero; remainder has the dividend's sign. Zero
divisors diagnose. The [calculator example](string-iteration.md) checks exact
divisibility before division and documents the pure string/int intrinsics it
uses, including UTF-16 indexOf and exact decimal int.toString. Named helper
self-recursion follows the lexical code binding and shares the call-depth budget;
it does not require a mutable captured function slot.

For R1, user-defined data helpers are lexical to a typefn; native type functions
and the prelude remain the cross-module composition mechanism. Packaging shared
user data helpers is an explicit remaining design question. Do not revive R0's
second typefn form silently. This restriction must be revisited with a multi-file
parser library before claiming broad library ergonomics.

## Native type construction

Use native syntax for common construction. Reflection provides a descriptor when
an algorithm needs metadata rather than just the value Type:

* Fields keep key identity, declared Type, derived read Type, optional and readonly.
* Objects keep fields, indexers and ordered call/construct signatures separately.
* Tuples keep position, label, required/optional/rest/variadic mode and readonly.
* Signatures keep parameter metadata, this, returns, predicates and abstractness.

Reflecting is shallow. Constructing validates duplicate effective keys, indexers,
optional/rest ordering, predicates and identity restrictions using native rules.
Reading an optional property is not the same as its declared Type plus absence.
Generic native signatures can be preserved opaquely; modifying their bound
parameter lists is not implied by supporting constructor-valued type functions.

An unchanged restricted descriptor can round-trip to its original Type. Editing
it cannot forge/drop class-private identity silently. Explicit public projection
or native intersection must state that intention. Interfaces remain native open
declarations; a structural object builder does not synthesize module augmentation.

Recursive structural Types use a guarded, atomic recursive builder. Unsealed
handles cannot be reflected, compared or cached. Enum-like literal unions use
ordinary union construction. Fresh nominal enum generation is deferred from the
R1 surface until declaration identity and serialization are settled; native enum
Types and their members can already be consumed/preserved. R0's generative enum
proposal remains historical, not a silently promised R1 feature.

## Symbolic evaluation, errors and inference

The [evaluation contract](evaluation.md) defines publication, pending checks,
substitution/replay, generic-body requirements and diagnostics. The rules below
are its summary; the [semantic cases](semantic-cases.md) supply concrete oracles.

The VM distinguishes a completed value, blocked observation, diagnostic and
cancellation. None is encoded by substituting any, unknown, never, false or an
empty shape. User parser failure can be a tagged data result or an explicit never
return. Resource failure is a diagnostic, not parser failure.

Known structure can be built with symbolic Type leaves: `{value:T}`, `[T,U]` or
`(x:T)=>T`. A nested Type-returning call can leave a deferred field Type while its
caller publishes known outer structure. Observations requiring unknown keys,
union membership, string parts or a dependent relation block the observing call.
That call discards its mutable frame and retains its original application for
substitution/reexecution. Constructor selection that blocks propagates to a
Type-returning call; abstract constructors never enter native Type argument slots.

Returned Types retain checks from all reached operations, including discarded
locals and operands erased by native simplification. Known shape does not imply
completed validation. Generic runtime bodies are checked once: they may use written
signature requirements, but cannot acquire hidden caller obligations from new
unproved local computations. The evaluation contract specifies this boundary.

Constraint validation is distinct from observing a value. A universally valid
native construction may be proved from bounds. Pending validation retains a
deferred operation/check instead of an unchecked native instantiation. A native
alias erasing its argument must not erase the argument's failed constraint.
TypeScript's failure to prove a symbolic relation is not proof that it is false.
Union iteration initially waits if its subject contains free Type parameters or
deferred terms; native union construction itself still permits symbolic leaves.

Native inference can see published symbolic structure, including Box<T>'s value
field. NoInfer and native variance survive. Arbitrary type programs cannot be
run backwards, and the VM cannot inspect a half-completed inference session.
The restricted HKT proposal specifies the constructor boundary. Constructor
arguments are explicit and restricted to ordinary Type inputs/results. Declared
native/typefn input bounds are checked on each application, not compared as
constructor domains. Constructor arguments are not inferred from native metadata
or outputs. Constructor subtyping, output bounds and native constructor parameters
are excluded. The broader design problems in
the [TypeScript HKT audit](../../typescript-hkt-audit.md) therefore do not all become
implementation requirements for this profile.

## Runtime and integration obligations

Use a handwritten typed register VM, dedicated AST kinds and source spans. Native
type operations go through one adapter that owns readiness, validation, budgets,
cancellation, source diagnostics and immutable substitution snapshots. Do not
represent constructor handles or deferred computations as ordinary Type parameters
with an unknown constraint.

Compile a declaration once per checker snapshot. Cache successful immutable results
by code identity, captured values, tagged Type/constructor arguments and compiler
options. Pending checks are part of the result's identity and substitution graph;
returned-with-checks and complete entries are distinct. Constructor captures and
return kinds participate in the key. A running
identical application is a cycle diagnostic; canceled/failed evaluations leave no
successful entry. Cache eviction cannot change successful semantic results.

All nested work shares budgets for instructions, stack depth, data memory, integer
size, string work, native graph construction and native structural operations.
One native operation can be expensive, so instruction fuel alone is insufficient.
Cancellation must interrupt native work too. Quota defaults require measurements;
no performance claim follows from this design.

The CLI and language server share parser, binder, evaluator and ABI. Type functions
erase from JS. Fork-specific declarations preserve executable bodies, constructor
aliases/constraints, closure dependencies and a versioned language ABI. Ordinary
TS parsers do not acquire this grammar automatically. No stock-compatible export
may silently replace an unresolved program with any or a guessed structural Type.
The [declaration contract](declaration-emit.md) specifies the ABI pragma, source
recipes, private dependency retention, lifted captures, guard preservation and
atomic failure behavior. Declaration-only consumption must use a separate checker.

Constructor parameters extend typefn declarations only. Native generic argument
vectors continue to contain ordinary Types; DSL environments/caches carry their
own constructor handles. Generated AST visitors, rename/hover/completion, declaration
consumers, incremental invalidation, canceled edit states and unchanged native projects are
all required acceptance areas. The archived prototype validates none of R1.
