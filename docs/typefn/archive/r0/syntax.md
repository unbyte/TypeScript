# Syntax and binding specification

R0 candidate. This file defines the surface independently of the old compiler
experiment. `Type` with a capital T denotes the interpreter kind of a native
TypeScript type handle. Lowercase `type` is a binding keyword or an explicit
native-expression boundary.

## Declarations and the TypeScript boundary

```ts
export typefn Box<T> {
  return { value: T };
}

export typefn box(value: Type): Type {
  return { value: value };
}

data Parsed = { value: Type; rest: string };
data ParseError = { offset: int; expected: string };

typefn parse(text: string): Result<Parsed, ParseError> {
  // Ordinary compile-time data calculation.
  return err({ offset: 0, expected: "a value" });
}
```

There are two forms of the same callable declaration:

* **Native entry:** `typefn Name<TypeParameters> { ... }`. Parameters and their
  constraints/defaults use native TypeScript syntax. Arguments have kind Type;
  the result has kind Type. A zero-parameter entry spells `typefn Name<>` to
  distinguish it from a data helper. In native TypeScript it is used as `Name`
  when no arguments are required, otherwise `Name<T, U>`.
* **Helper:** `typefn name(parameters): Kind { ... }`. Parameters have interpreter
  kinds. It is callable only inside the DSL with parentheses. Return annotations
  are required in R0, including on recursive helpers. A helper whose return kind
  is Type writes `: Type` literally; a data alias cannot alias Type alone.

A native entry also has a compile-time function value: `Box(T)` is equivalent to
`type(Box<T>)` inside a data expression. It validates the same generic constraints
and defaults. It is not a runtime function. A helper has no native generic-type
application adapter; `parse<string>` in ordinary TypeScript is an error.

`export`, `declare`, type imports/reexports and enclosing native namespaces apply
to top-level declarations. A `declare typefn` still contains its body: execution
must remain available to consumers. Overloaded typefn declarations are not allowed.
Use explicit dispatch or function values. Duplicate declarations cannot merge.

Helpers and `data` aliases may be nested in a body. Nested native entries are
excluded in R0; use a helper and capture Type values instead. All helper calls
are statically kind-checked. Recursion is allowed subject to runtime limits.

## Bindings

```ts
typefn Snap<T> {
  type current = T;
  const type saved = { value: current };
  current = number;
  return [saved, current];
}
// Snap<string> = [{value: string}, number].
```

`type name = TypeExpression` introduces a mutable local Type binding, initialized
once when reached. `const type name = TypeExpression` introduces an immutable one.
Local Type bindings cannot declare generic parameters. Native entry parameters,
helper parameters, pattern captures and `for (const ...)` bindings are immutable.

`let`/`const` bind ordinary data. Their initializers and optional annotations use
data syntax. A direct Type-valued binding must use `type` or `const type`; a data
record, list or closure may contain Type handles. This makes a local binding's
expression mode explicit without querying native assignability.

```ts
typefn FieldNames<T> {
  const shape = reflect.object(T);
  let names: List<Type> = [];
  for (const field of shape.fields) { names.push(field.key); }
  return union(names);
}
```

Names have lexical scope. Reads before initialization are errors; there is no
temporal initialization by lazy aliases. A local initializer cannot refer to the
binding it is introducing. Recursive Types use the recursive constructor API.
Nested scopes may shadow outer names; duplicate declarations in one scope fail.
Assignments to parameters, captures, nonlocals or `const` bindings fail. Function
closures capture snapshots at creation, not writable access to outer locals.

Top-level ordinary TypeScript `type` aliases keep their existing semantics.
Inside native type syntax, mapped/infer binders keep native scope and shadowing.
They do not introduce imperative variables.

## Two expression modes

**Type mode** occurs after a Type binding's `=`, on assignment to a local Type
binding, and after `return` in a native entry or helper explicitly returning Type.
It includes the full native type-expression grammar, plus calls returning Type
and projections of Type-valued descriptor fields:

```ts
type R = A | B;
R = { readonly value?: R };
type M = { [K in keyof T]: T[K] };
type C = T extends readonly [infer H, ...infer Rest] ? H : never;
type F = (...args: Args) => Result;
type O = object({ fields: fields });
type P = field.type;
return Box<T>;
```

The explicit `type(TypeExpression)` form is also legal in Type mode, although
usually redundant. `object(...)` and other ordinary calls in Type mode parse their arguments in data
mode. Their result must be Type. A qualified name such as `field.type` can refer
to a Type-valued data projection or a native qualified type name; binding resolves
which. A projection yielding ordinary data is a kind error, never an implicit
literal conversion. Native indexed access remains native indexed access: `T[K]`
requires K to be a Type handle. It is not data-list indexing.

**Data mode** occurs in data bindings, conditions, loop iterables, call arguments,
data-helper returns and ordinary collection operations. It has literals, records,
lists, arithmetic, calls, projections and lambdas. Native syntax is embedded as
`type(TypeExpression)`. A Type variable can be passed directly as a value.

```ts
const enabled = true;             // boolean data
const field = { key: literal("x"), type: T, optional: false, readonly: true };
const keyText = "x";              // string data
type Key = literal(keyText);      // native literal Type
const sameAsNever = same(T, type(never));
const boxed = { type: type({ value: T }) }; // data record containing a Type
```

There is no implicit conversion of string data to literal Type, boolean data to
boolean-literal Type, or a data record to an object Type. `type(...)` parses native
syntax; `literal(value)` converts computed data. They are different operations.

### Parsing assignment contexts

A simple assignment root is a previously declared lexical binding. Its binding
form (`type`, `const type`, `let`, `const`) determines RHS mode; this information
is syntactic and does not depend on TypeScript inference or VM execution. A
parser may track these lexical forms or defer RHS elaboration to the DSL binder,
but must retain one dedicated, source-positioned DSL syntax representation. It
must not guess from a computed value or rewrite text into generated aliases.

An assignment through a data path always uses data mode, even when the field has
kind Type: `shape.fields[i].type = type(Promise<T>)`. Assignments to properties of
native Type handles (`T[Key] = ...`) are **not** in R0. Use descriptor edits and
`object(...)`; these make modifier and identity policies explicit. This differs
from the archived experiment's ad hoc property mutation.

Unknown assignment targets receive a binding diagnostic; parser recovery must not
need to evaluate either interpretation. Native type syntax inside a data lambda's
return always uses `type(...)`, even when contextual typing expects Type. This
keeps lambdas independent of their eventual call sites.

## Statements and compact grammar

This grammar treats `NativeType` and native type parameter lists as references
to the host TypeScript grammar. `TypeExpr` extends NativeType with Type-valued
calls and descriptor projections as described above. `DataExpr` is the restricted
expression grammar described below. Assignment names are classified by lexical
binding form, not arbitrary type inference.

```text
native-entry := modifiers? 'typefn' Name '<' NativeTypeParameters? '>' block
helper       := modifiers? 'typefn' Name '(' parameters? ')' ':' Kind block
data-alias   := modifiers? 'data' Name '=' Kind ';'
parameters   := Name ':' Kind [ '=' DataExpr ] (',' parameters)?
block        := '{' statement* '}'
statement    := ['const'] 'type' Name '=' TypeExpr ';'
              | ('let' | 'const') Name [':' Kind] '=' DataExpr ';'
              | TypeLocalName '=' TypeExpr ';'
              | DataPath assignment-op DataExpr ';'
              | DataCall ';'
              | 'return' ReturnExpr ';'
              | 'if' '(' DataExpr ')' block ['else' (block | if-statement)]
              | 'while' '(' DataExpr ')' block
              | 'for' '(' 'const' Name 'of' DataExpr ')' block
              | 'for' '(' let-init ';' DataExpr ';' data-update ')' block
              | match-statement
              | 'break' ';' | 'continue' ';'
              | helper | data-alias | block | ';'
match-statement := 'match' '(' SubjectExpr ')' '{' arm+ '}'
arm          := Pattern ['if' DataExpr] '=>' block [',' | ';']
ReturnExpr   := TypeExpr in Type-returning bodies; DataExpr otherwise
SubjectExpr  := DataExpr  // A Type binding is a valid data expression.
assignment-op := '=' | '+=' | '-=' | '*=' | '%='
DataPath     := DataLocalName ('.' Name | '[' DataExpr ']')*
```

Semicolons are required in the canonical form. R0 permits the host's line-break
semicolon insertion for simple statements, with the same restrictions around
`return`; tools emit explicit semicolons. Braces are mandatory for control flow.
No labels, exceptions, generators, async functions, dynamic eval, JavaScript
`new`, classes, prototype operations or runtime statements occur inside the DSL.
`break`/`continue` target the nearest loop, including from within match arms.
Falling off a function is an error unless its result kind is `void`.

### Data operators

Precedence follows this order, tightest first: member/index/call; unary `! - + ~`;
`* / %`; `+ -`; shifts; `< <= > >=`; `== !=`; bitwise `& ^ |`; `&&`; `||`; `?:`.
Operands are evaluated left to right. Logical and conditional operators short
circuit. Assignment is a statement, not an expression. Increment/decrement and
comma expressions are excluded. There is no JavaScript coercion or truthiness.

`==` on data performs structural equality for equal comparable kinds. Type handles
and functions are not data-comparable; use `same` for native identity, and explicit
keys for functions. `is` is not an R0 operator: `assignable(A, B)` and `same(A, B)`
make the queried relation clear and work naturally as function values. Native
conditional types remain available inside Type expressions.

`+` supports equal numeric kinds and string concatenation. Integer division uses
`math.div(a, b)` (truncation toward zero); `/` is for number data. Mixing int and
number requires an explicit checked conversion. Integer `%` satisfies
`a == math.div(a,b)*b + a%b`. Division by zero and nonfinite number results fail.

## Compile-time kinds

Kinds are checked statically by the DSL checker, not native TS assignability:

```text
Kind := Type | boolean | int | number | string | void
      | literal-data-kind
      | List<Kind> | Map<KeyKind, Kind> | Set<KeyKind>
      | Option<Kind> | Result<Kind, Kind>
      | '{' (Name ['?'] ':' Kind ';')* '}'
      | '(' (Name ':' Kind ',')* ')' '=>' Kind
      | Kind '|' Kind
      | NamedDataKind
```

`int` is an arbitrary-precision signed integer, bounded by the evaluation's bit
budget. Integer source literals have kind int; fractional/exponent literals have
kind number, which uses finite IEEE-754 binary64. Checked `number(int)` and
`int(number)` conversions require exact representability. `number` data retains
negative zero. `literal(-0.0)` follows native TypeScript's literal normalization.
Native numeric/bigint literal syntax is always available in Type mode.

Strings are immutable UTF-16 code-unit sequences, including isolated surrogates.
`length`, `slice`, `at` and template captures count code units. `codePoints` and
`fromCodePoints` are explicit alternatives. This choice follows JavaScript string
and historical TS template-literal behavior, not Go rune indexing.

Records are closed structural data values. Optional fields are absent or present;
absence is not an `undefined` data value. Read an optional field through `get`,
returning Option. Record spread requires known fields; later fields replace earlier
ones. Lists are ordered and homogeneous; data tuples are represented by records
or lists in R0. Native tuples remain separate Type values. Tagged data unions use
disjoint literal tags and are narrowed by match/boolean tag tests. They do not
have TypeScript's arbitrary conditional/mapped type algebra.

`data` aliases can be mutually recursive only through records, lists or tagged
variants. Naked cycles such as `data A = B; data B = A` fail. Values are finite;
data cycles and self-containing mutable lists are not supported. Kind parameters
on user helpers are not in R0; built-in List/Map/Result operations are kind-generic.
Functions on Type values are enough for higher-order type composition without
introducing native generic signature construction.

Map/Set keys are strings, booleans, ints, finite numbers or explicit `TypeKey`
tokens obtained from `identity(T)`. TypeKey has equality but no ordering or textual
conversion. It is local to a checker snapshot. `members(T)` returns TypeSet, a
special immutable set whose keys are native constituent identities.

Map preserves insertion order. Updating an existing key retains its position;
deleting and reinserting appends it. Map equality/cache keys include the entry
sequence. Set traversal is unspecified like TypeSet, and Set equality ignores
order. No program may assume that semantic equality of unordered sets implies a
particular iteration order.

Empty collections require an annotation or a contextual kind. There is no data
`any`, implicit unknown escape, heterogeneous untyped list, or unchecked cast.

### Kind checking and local inference

Type is one opaque interpreter kind: two Type values do not have different data
kinds merely because their native Types differ. Native relations are explicit
operations. An ordinary data function cannot turn kind checking into a native
conditional/mapped type computation.

Scalar literal data kinds are subkinds of their primitive kind. `let` scalar
initializers widen to their primitive kind unless annotated; immutable constants
may retain literal kinds. A contextual record/variant annotation retains literal
discriminators. Records are closed: assignment requires the declared field set,
respecting optional fields; extra fields require an explicit projection or a
different declared kind. Pattern record matching may inspect a subset without
changing this construction rule.

List, Option and Result are covariant in their element kinds because values and
parameters have snapshot semantics. Assigning a List to a widened local cannot
give that local write access to the original list. Map key kinds must agree; its
value kind is covariant. Functions are contravariant in parameter kinds and
covariant in results with compatible required/default arities. TypeKey and opaque
descriptor tokens are nominal interpreter kinds with no structural fabrication.

Local inference uses declared/contextual kinds and finite least upper bounds of
scalar literals, records and explicit tagged unions. It does not infer a universal
kind-polymorphic helper from different call sites. If no permitted common kind
exists, require an annotation or reject. There is no implicit fallback to any.
Tag comparisons and record/variant patterns narrow data unions; mutation of a
root invalidates refinements of paths through that root. Captured snapshots retain
their refinements. Native Type constraints never participate in data kind inference.

## Functions and imports

Lambda syntax is `(x: Kind, ...) => DataExpr` or `(x: Kind, ...) => { ... }`.
Parameter annotations may be omitted when a declared higher-order parameter gives
an unambiguous kind. Lambda blocks use data-mode returns; Type expressions use
`type(...)`. Lambdas capture immutable snapshots. Recursion uses a named helper;
there is no mutable self-capture trick.

Native entries, helpers and data aliases may be exported and imported with native
type-only import syntax. The binder records which namespace/kind each export
belongs to. Only native entries can appear as ordinary TypeScript generic types.
Native aliases are not automatically callable data functions: wrap an alias in a
native entry if it needs to be passed as a higher-order argument. Runtime values
are never executed; `typeof` inside a native type expression can inspect their
static Type according to normal TypeScript rules.

Top-level helper names are mutually visible, including for recursion. They have
no executable module initialization. Nested helpers form declaration groups whose
references to other helpers are mutually visible; captures must already be
initialized at the group's declaration point. Calling an uninitialized local
function value remains an error. A group captures one immutable environment.

## Core data library

These operations complete the data vocabulary used by the examples. They are
typed interpreter intrinsics, not calls into the JavaScript host:

* Lists: `length`, `at(i)`/`[i]`, `slice(start,end)`, `map(f)`, `filter(f)`,
  `some(f)`, `every(f)`, `join(separator)` for string lists. Updates on mutable
  roots: `push`, `pop` (Option result), `insert`, `remove`, `reverse`, indexed set.
* Maps: contextually typed `map()`, `get(key)` returning Option, `has`, `entries()`
  yielding `{key,value}` records, `keys`, `values`, `size`; updates `set`, `delete`.
* Sets: contextually typed `set()`, `has`, `size`; updates `add`, `delete`.
  TypeSet is returned by `members` and supports size/iteration; `union` and
  `intersection` accept a TypeSet or a List<Type>.
* Strings: `length`, `at(i)`/`[i]`, `slice(start,end)`, `startsWith`, `endsWith`,
  `includes`, `indexOf` (int, -1 when absent), literal `split`, `trimStart(chars)`.
  Split uses nonoverlapping literal delimiters and retains empty segments; an
  empty delimiter splits UTF-16 units. Trim removes only units in its argument.
* Unicode: `text.codeUnitAt(s,i)`, `text.fromCodeUnits(list)`, `text.codePoints(s)`,
  `text.fromCodePoints(list)`. Invalid units/points diagnose; lone surrogates are
  retained as their unit value by codePoints rather than replaced with U+FFFD.
* Numeric: checked conversions `int`/`number`, `math.div`, `math.abs`, `math.min`,
  `math.max`, and bounded `math.parseInt(text,radix)` returning Result.
* Variants: `some(value)`, `none`, `ok(value)`, `err(value)`. Missing type parameters
  on constructors are supplied by contextual kinds; there is no unchecked unwrap.

Indices use int, start-inclusive/end-exclusive ranges, and no implicit negative
index wrapping. Out-of-bounds access and invalid slice ranges diagnose. List and
string `slice(start)` defaults end to length. Collection callbacks are closures
with snapshot captures. Mutation callbacks cannot mutate an enclosing local.
`fail(string)` is a control-bottom intrinsic usable at any return kind.
