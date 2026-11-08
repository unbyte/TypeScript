# Evaluation and semantic specification

R0 candidate. The interpreter's values are ordinary compile-time data, function
values and immutable native Type handles. Native generic parameters may occur
inside Type values. They are not ordinary unknown data values.

## State and effects

An evaluation has lexical environments, local variables, a call stack, a resource
account and a native checker snapshot. Expressions evaluate left to right.
Assignments replace values in local bindings. No operation mutates an existing
native Type, native symbol, AST node or shared descriptor.

Data collections have **value semantics**. Saving a list/record creates a snapshot;
mutating either binding cannot change the other. An implementation may use
persistent structures or copy-on-write with ownership checks. This differs from
JavaScript reference aliasing and is intentional.

```ts
let a: List<int> = [1];
const saved = a;
a.push(2);
// saved == [1], a == [1, 2]
```

Mutation methods are permitted only on paths rooted in a local `let` binding.
They update that root; a method call on a temporary, parameter, `const`, capture
or function return is rejected. `list.push(x)`, `reverse`, `set` and `delete` are
typed update operations, not dynamically looked-up JavaScript methods. Compound
updates evaluate the path and RHS once, then publish the new root atomically.
Records cannot acquire undeclared data fields through assignment; use a Map or
construct a new record. Reflected descriptor fields may be changed in a local
copy, but the native source type changes only by explicitly constructing a result.

Type assignment also has snapshot semantics. A native lazy object or signature
must capture an immutable environment, not the interpreter's mutable register
array. A Type-valued field projected from data can be passed directly to a type
constructor. Reifying it does not retain mutable builder storage.

There are no host effects. Calls can read native type information in the current
checker snapshot, allocate local data, construct native types and report errors.
They cannot read a clock, filesystem, network, environment, randomness, mutable
global state or execute runtime JavaScript. Diagnostic formatting is not a source
of semantic data: programs cannot branch on a displayed type string.

## Native type expressions

Native union, intersection, conditional, mapped, indexed-access, template, object,
tuple, function, constructor, operator and literal expressions retain native
TypeScript meanings. The fork remains responsible for their semantics; the VM
does not invent a second assignability engine. Captured interpreter Type bindings
are substituted by handle in an immutable evaluation environment.

Native syntax and name binding are checked in all branches. Value-dependent
checks—such as whether a particular key exists or a computed argument satisfies
a generic constraint—run when that expression executes. An unexecuted branch
cannot fail because its operation is invalid for the current arguments. This is
essential for ordinary imperative guards. A missing name or malformed type
expression remains a static error even in an unreachable branch.

Type expressions may preserve native recursive aliases and existing generic
signatures. Dynamically introducing new native generic parameter lists is not
part of this revision. Native expression forms that declare a generic signature
are therefore rejected inside constructions; referencing/copying an already
declared generic signature is allowed.

## Unions and special types

Calls and branches do not distribute implicitly. `members(T)` explicitly returns
the normalized constituent TypeSet. `never` gives the empty set; any nonunion
gives a singleton. `boolean` follows the native checker's normalized union of
boolean literal constituents when exposed by that checker. `any` and `unknown`
are singleton special Types, not iterators over possible runtime values.

`union(collection)` uses native normalization and maps an empty collection to
never. `intersection(collection)` maps an empty collection to unknown. These
operations do not restore absorbed constituents such as `"x"` in `string | "x"`.
An algorithm requiring correlation must traverse each member and rebuild that
member before combining results. Nested unions are not automatically expanded
into Cartesian products.

TypeSet iteration order is unspecified. It is stable for a particular set value
within an evaluation, but not promised across compiler versions, equivalent
spellings or different checker snapshots. Order-sensitive output from such a loop
is permitted but nonportable; the language does not pretend to prove commutativity.
`ordered(set, key)` provides a portable list only when key returns unique, totally
ordered scalar keys. Duplicate keys are an error; Type IDs and display strings
are not implicit tie-breakers. Union/intersection accumulation and independent
field maps are the intended unordered operations.

Relations return boolean data when decidable:

| Query on closed Types | never | unknown | any |
| --- | --- | --- | --- |
| `assignable(T, type(string))` | true | false | true |
| `same(T, type(never))` | true | false | false |
| `reflect.kind(T)` | `"never"` | `"unknown"` | `"any"` |
| `members(T)` | empty | singleton | singleton |

`same` means the native identity relation, not a proof of every mathematical
equivalence. Mutual assignability can be written as two `assignable` calls but
does not replace identity. A predicate consuming Type does not return a Type
literal boolean. Native conditional expressions retain their separate behavior
around any, never and distribution.

## Symbolic execution and suspension

Native generic parameters remain symbolic Type values. Ordinary data does not
have an implicit symbolic-number, symbolic-list or symbolic-boolean form. Data
records may contain symbolic Types when their data structure is already known.

Operations are divided by what they observe:

1. **Structure-preserving construction** can proceed with symbolic Types. Examples
   are `{value:T}`, `(x:T)=>T`, a tuple of known length containing T, or a record
   descriptor with a known key and T as its value.
2. **Observation of known structure** can proceed when substitution cannot alter
   that observation. A literal object `{value:T}` has a known field name even if
   its value is symbolic. A required tuple `[T,U]` has two known positions.
3. **Observation that depends on unresolved information** suspends. Examples are
   the property set of T, normalized members of `T | string`, template captures
   from T, or a relation whose answer can change after substitution.

For an observation, the checker adapter returns `Ready(value)`, `Blocked` or
`Failure(diagnostic)`. Blocked is an internal interpreter signal, not any, unknown,
never, false, an empty collection or a catchable user result. It propagates to
the enclosing native entry. The entry discards unpublished local state and
returns a deferred native application of its original closure and arguments.
Instantiation reevaluates that application with substituted arguments.

R0 uses whole-entry restart after suspension. It does not promise continuations
of partially executed mutable frames or automatic merging of branches. Calls that
finish can publish native symbolic structure immediately:

```ts
typefn Box<T> { return { value: T }; }
function unbox<T>(box: Box<T>): T { return box.value; }
```

The second line must work before T is concrete. In contrast:

```ts
typefn Choice<T> {
  if (assignable(T, type(string))) { return { text: T }; }
  return { other: T };
}
```

`Choice<T>` suspends when the relation depends on T. It must not select the second
branch just because native assignability cannot prove the first. A constraint is
not the argument. It cannot stand in for T during enumeration or branching.

For R0, relation queries conservatively block when either operand contains free
outer parameters, except an explicitly proven substitution-invariant identity
such as comparing the same immutable handle with itself. Bound parameters inside
an existing generic signature are not free outer parameters. Future stronger
proofs must be specified and tested; they cannot use a failed native proof as
evidence that a relation is false.

### Validation versus observation

Constructors must validate native constraints even during symbolic construction.
A native verification judgment may prove that a construction is valid for every
permitted substitution—for example an element T constrained to string in a
string-valued index signature. Such a proof can use constraints without treating
them as actual values. The adapter returns Verified, Rejected or Pending.

If a necessary validation is Pending, construction suspends the entry. R0 does
not publish an unchecked Type and hope that later instantiation happens to check
it. This avoids lost obligations when a native alias erases its argument:

```ts
type Checked<T extends string> = number;
typefn F<T> { return Checked<T>; }
// F<number> must diagnose the constraint even though Checked's result is number.
```

Simple symbolic records must remain usable; validation that observes only their
known syntax cannot spuriously demand that all field Types be closed. This is a
specific conformance requirement, not an optimization to add later.

`reflect.kind(T)` does not expose `"unresolved"`, `"parameter"` or `"deferred"`
as ordinary classifications a program can branch on. That would make an answer
depend on whether the checker queried before or after inference. Such states are
available only in debugger/diagnostic output.

## Inference boundaries

Native inference sees the native symbolic structure that an entry publishes.
It can infer T through `Box<T>` just as it can through `{value:T}`. Existing
NoInfer barriers, variance annotations, unique symbols and native signatures must
survive publication. This is the main benefit of living inside the checker.

The VM cannot read or mutate a half-complete inference session, choose candidates,
or run an arbitrary program backwards. A deferred computation is opaque to
inverse inference unless an ordinary native relation exposes more structure.
Inference order must not become observable to a DSL program. Generic constraints
can validate a candidate; they do not prescribe how to invert a transformation.

Native conditional `infer` is available in type expressions. A separate general
solver/query language is not in R0. Structural match patterns capture parts that
their specified decompositions can expose; they are not arbitrary inverse calls.

## Matching

`match` is a statement. It evaluates its subject once and tries arms in source
order. The first successful pattern whose guard is true executes. Each arm has
its own scope. There is no fallthrough and no implicit block result. `return`
returns from the surrounding function. Captures are immutable; duplicate capture
names in a pattern fail statically. `_` discards. An existing name is never
implicitly a value pattern.

Every match must have a final unguarded `_` arm, except when the compiler proves
exhaustiveness for a finite closed data union/boolean/Option/Result. Type-domain
matches require `_` in R0. A wildcard can return never deliberately, raise an
error, continue a loop or implement a fallback. Silent no-op matching is excluded.

Patterns have these forms:

```text
Pattern := '_' | Capture
         | scalar-literal
         | template-pattern
         | '[' Pattern* [',' '...' Capture] ']'
         | '{' (Name ':' Pattern | Name)* '}'
         | 'type' '(' NativeType ')'   // exact native identity pattern
         | 'is' NativeType            // whole-Type assignability pattern
         | 'some' '(' Pattern ')' | 'none'
         | 'ok' '(' Pattern ')' | 'err' '(' Pattern ')'
```

A scalar pattern compares data values when the subject is data. On a Type subject,
it compares the corresponding native literal Type using native identity. Record
patterns on data require the named fields; extra fields are permitted. On Type
subjects they require one object with those public **required** fields; captures
are field Types. Optional fields do not match a required-field pattern. Reflection
is the explicit way to process optionality and indexers.

A list pattern decomposes ordered data lists or native finite required tuples.
A native rest/optional tuple or open array does not match this tuple pattern;
use reflection for its metadata. A leading sequence consumes positions; a trailing
rest capture receives the remaining data list or native tuple, respectively.
Readonly and labels are preserved on a captured native remainder. No rest capture
can occur before a fixed element in R0 patterns.

Type matching does not implicitly distribute unions or inspect an existential
runtime value. `match (type(string)) { "x" => ... }` does not succeed because x
is a possible string. An unresolved subject blocks if its future structure can
change the answer. A wildcard can still match without observing structure; an
earlier blocked arm prevents falling through to it.

Guards use data booleans and can read captures. They cannot mutate enclosing
state. Calls in guards are pure with respect to the caller; local computations
inside such helpers are allowed. A blocked guard suspends the entry. Captures
from failed patterns/guards never leak into another arm or the enclosing scope.

### String patterns

```ts
match (text) {
  `/users/${id}` => { /* id */ }
  `${head}${tail}` => { /* first UTF-16 unit and the remaining text */ }
  _ => { /* fallback */ }
}
```

Template patterns have literal text and three interpolation forms:

* `${name}` captures a string (data subject) or string Type (Type subject).
* `${_}` consumes the same part without introducing a name.
* `${=expression}` inserts a computed **data string** delimiter. It introduces
  no capture. The expression is evaluated once when this arm is attempted and
  cannot mutate caller state. A nonstring is a kind error; unresolved evaluation
  blocks rather than guessing a delimiter.

Matching is anchored at both ends. Literal delimiters are not regexes. Prefix
and suffix must match. Between captures, the earliest delimiter occurrence is
selected; the matcher does not backtrack to find another decomposition. Adjacent
captures give one UTF-16 code unit to each nonfinal capture; the final capture
receives the rest. A capture before a nonempty delimiter may be empty. A sole
capture accepts an empty string. Empty pins become ordinary empty literal parts
and use the same adjacent-capture rule after coalescing literal fragments.

Examples: `${a}:${b}` on `x:y:z` binds x and y:z; on `:` binds two empty strings.
`${a}${b}` on the empty string fails; on one code unit it binds that unit and an
empty string. `${a}${b}` on an astral character splits its surrogate pair. Use
`text.codePoints` when a code-point parser is desired. No normalization of Unicode
or case folding is implicit.

For closed native template-literal Types, a successful decomposition returns
native string Types, possibly infinite patterns such as `${number}`. Broad string
matches only a sole unrestricted capture or wildcard, not a specific prefix.
Patterns must cover the entire native string Type, not merely some of its possible
values. A Type union requires explicit member processing. Type intersections with
brands require an explicit `reflect.string`/normalization policy; the matcher
does not silently discard a brand to force a match. A nonstring Type is a normal
pattern mismatch, not an evaluation error. The Type pattern adapter must implement
the same UTF-16 contract even if a native helper currently uses Go code points.

## Errors and termination

There are three distinct outcomes: a returned value, a suspended application,
or an evaluation diagnostic. Native never is a successful Type value. User
parsers can return `Result`/`Option` data to express ordinary failure; those are
not compiler diagnostics. `fail(message)` raises a source-located evaluation
diagnostic and aborts the enclosing native entry. No catch mechanism is provided.

Static name/kind/syntax errors invalidate the declaration before VM execution.
Constructor validation, out-of-bounds indexing, checked conversion failures and
resource exhaustion are evaluation diagnostics. Errors include the operation
span and a bounded application trace. Cancellation is separate and publishes
neither a result nor a diagnostic for an obsolete document version.

Reentering the same running closure with the same arguments is a cycle error.
Explicit recursive type constructors provide guarded type cycles without
recursively executing the same function forever. Other recursion and loops share
the native entry's resource account. Exhaustion never silently truncates a type,
returns never, or substitutes any.
