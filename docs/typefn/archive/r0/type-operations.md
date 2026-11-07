# Native type operations and construction

R0 candidate. Names below are DSL prelude functions/kinds. They are not injected
into ordinary TypeScript programs or backed by JavaScript runtime objects. Calls
return native Type handles or finite immutable interpreter data. Every operation
can internally suspend or fail as described in the semantic specification.

## Core operations

| Operation | Result and contract |
| --- | --- |
| `literal(value)` | Native string, boolean or finite number literal Type; int must fit exactly in a native number |
| `literal.bigint(int)` | Native bigint literal Type without conversion through binary64 |
| `literal.value(T)` | Data scalar of a single unbranded literal Type; broad types/unions/enums reject; unresolved literal blocks |
| `literal.integer(T)` | Exact int from a single integral native number/bigint literal; rejects fractional/nonliteral Types |
| `same(A,B)` | Native identity relation, with symbolic blocking rules |
| `assignable(A,B)` | Native assignability, with symbolic blocking rules |
| `identity(T)` | Opaque equality-only TypeKey; blocks if substitution can change the identity being observed |
| `members(T)` | Normalized TypeSet; unresolved normalization blocks |
| `union(types)` | Native normalized union; empty means never |
| `intersection(types)` | Native normalized intersection; empty means unknown |
| `reflect.kind(T)` | Stable semantic classification; never a diagnostic/internal flag name |

`literal.value` returns number data for number literals, int for bigint literals,
string for string literals, boolean for boolean literals. Since its result is a
closed tagged data union, callers match its data kind or use the checked typed
forms `literal.text`, `literal.boolean`, `literal.number`, `literal.integer`.
`literal.text(T)` requires a single unbranded string literal. `type(null)` and
`type(undefined)` supply native nullish Types directly; they are not ordinary
DSL null/undefined data values. Option represents absence in data.

Type identity tokens are intended for cycle guards/maps, not serialization,
sorting or constructing nominal identities. Display APIs exist in diagnostics
and editor inspectors, not as semantic conversion functions.

`reflect.kind` returns one of: any, unknown, never, string, number, bigint, boolean,
null, undefined, void, symbol, enum, tuple, array, object, union, intersection.
Literal Types report their primitive family; literal extraction is separate.
Callable/constructable Types are objects with signatures, not a mutually exclusive
kind. An enum is classified before its underlying primitive/union. An unreduced
intersection reports intersection; callers can ask for its effective object shape
or deliberately inspect constituents. This classification observes native semantic
structure, not the spelling of an alias. Free parameters block classification.

## Object descriptors

The following are interpreter data kinds, with optional fields represented by
Option at the API boundary. The notation is descriptive; named prelude kinds are
available without declaring them in each program.

```text
Field = {
  key: Type;                 // one string/number literal or unique symbol Type
  type: Type;                // declared value Type, absence excluded
  readType: Type;            // derived: Type observed when reading the field
  optional: boolean;
  readonly: boolean;
}
Indexer = { key: Type; type: Type; readonly: boolean }
ObjectShape = {
  fields: List<Field>;
  calls: List<Signature>;
  constructs: List<Signature>;
  indexers: List<Indexer>;
  origin: Option<ObjectOrigin>;
}
```

`reflect.object(T)` requires one object-like Type, including a callable object or
an effective native intersection of objects. It rejects any, unknown, never and
unions. It includes inherited public fields, all call/construct signatures and
indexers. An unresolved property set blocks. Field Types may remain symbolic.
Reflection is shallow: it never recursively visits every field's Type.

`readType` reflects optional-property reading under the current native compiler
options. `type` plus `optional` preserve the declared/write distinction under
`exactOptionalPropertyTypes`. `object` ignores a supplied readType and derives it;
changing readType cannot change the constructed Type. The same rule applies to
other documented derived descriptor fields. Required undefined and absent optional
are never collapsed into one field state.

Object field order is not semantic. Reflection provides a reproducible source/
native-member order within one snapshot for diagnostics, not a portable property
ordering algorithm. Keys retain unique-symbol identity and native numeric/string
key equivalence. A broad string/number indexer is not enumerated as a property.

`object({fields, calls?, constructs?, indexers?, origin?})` constructs a new native
object Type. Omitted collection fields default to empty, origin defaults to none.
Fields require key/type/optional/readonly; the shorthand helper
`field(key, type, {optional:false, readonly:false})` supplies false defaults and
derives readType. Constructors reject duplicate effective keys, incompatible
overlapping indexers, fields incompatible with applicable indexers, malformed
signature descriptors and inaccessible/private identity fabrication. Indexer
domains use native legal string/number/symbol/template domains and compatibility.

`reflect.object` records an opaque ObjectOrigin if the source has identity-bearing
members or semantics that a public structural descriptor cannot reproduce. An
unchanged round trip may return the original Type. Altering such a descriptor
while retaining its origin is rejected in R0; removing the origin is also rejected
when it would silently discard restricted members. To deliberately project a
public shape, call `reflect.public(T)`; that API returns an origin-free public
descriptor and explicitly drops private/protected identity. To retain identity
while adding constraints, use a native intersection with T.

Reflection never promises to create a named, open `interface` declaration or
module-augmentable registry. Existing native interfaces/aliases can be preserved
as handles. A structural object constructor does not manufacture a new declaration.

### Property lookup

`reflect.field(T, key)` returns Option<Field> for a named public member of a known
object shape. It does not synthesize a declared property just because an indexer
permits that key. `index(T,key)` returns the native indexed-access Type, including
native symbolic indexed access if its validation is universally valid. It follows
type-level indexed access, not runtime expression `noUncheckedIndexedAccess`
insertion; `readType` and actual native expression checking retain their respective
rules. Missing keys, uncertain structure and optional fields are distinct cases.

## Tuples, arrays and signatures

```text
Element = {
  type: Type;
  label: Option<string>;
  mode: "required" | "optional" | "rest" | "variadic";
}
TupleShape = { elements: List<Element>; readonly: boolean }
ArrayShape = { element: Type; readonly: boolean }
Parameter = {
  name: Option<string>;
  type: Type;
  optional: boolean;
  rest: boolean;
}
ConcreteSignature = {
  kind: "concrete";
  parameters: List<Parameter>;
  returns: Type;
  this: Option<Parameter>;
  predicate: Option<Predicate>;
  abstract: boolean;
}
Signature = ConcreteSignature | OpaqueGenericSignature
```

`reflect.tuple(T)` requires a single native tuple. A tuple can have optional,
rest/variadic and readonly metadata; the API does not force it into a finite list
of guaranteed values. A fixed tuple containing symbolic element Types can be
reflected. If future substitution changes the segment layout, reflection blocks
unless the variadic segment itself can be retained as a segment descriptor.

`tuple(shape)` validates native tuple element order, optional/rest constraints,
labels and readonly state. A rest element stores the repeated element Type;
a variadic element stores the tuple/array Type to splice. These cannot be confused.
`reflect.array(T)`/`array(shape)` distinguish mutable/readonly arrays from tuples.
Data List indexing or native tuple pattern matching cannot replace these APIs.

`element(T, {label?, mode?})` creates an Element descriptor with default label none
and mode required. `parameter(T, {name?, optional?, rest?})` creates a Parameter
with no name and false flags by default. These shorthands construct data; native
validation still occurs when the enclosing Type is constructed.

`reflect.calls(T)` and `reflect.constructs(T)` return all native signatures in
their effective overload order. They do not use the common conditional-infer trick
that observes only the last overload. `callable(signature)` and
`constructable(signature)` create one callable/constructable Type. `overload(list)`
creates an ordered call set; `constructors(list)` creates an ordered construct set.
An empty overload set is rejected. An object may contain both lists and properties.

Concrete signatures retain parameter names, optional/rest information, explicit
this, return Type, type predicates/assertion predicates and constructor abstractness.
Predicate targets are a valid parameter index or this; asserted Types must pass
native predicate validation. Rest parameters must be arrays/tuples. Constructor
and call-only fields cannot be used in the wrong signature kind. Method bivariance
and other origin-specific native behavior must be recorded as an opaque signature
origin when ordinary reconstruction would change it; it cannot silently disappear.

Generic signatures are opaque values carrying their native identity and captured
substitution environment. They may be copied and installed unchanged. Their
parameters/returns cannot be edited by pretending their bound parameters are free
Type handles. Reflection explicitly reports `kind: "generic"`; algorithms must
choose to preserve, skip or reject such signatures. Dynamic generic-signature
construction is outside R0, as requested.

Native syntax can express the common concrete forms directly:

```ts
typefn Callback<Args, Output> { return (...args: Args) => Output; }
typefn Constructor<Args, Instance> { return new (...args: Args) => Instance; }
```

These still validate symbolic/rest obligations. A syntax shortcut must not bypass
the descriptor constructor's semantic checks.

## String Type reflection

`reflect.string(T, {brands:"reject"|"strip"})` returns a native string-base Type,
Option<string> for an exact literal, and known literal/template segments. It
requires a recognizable string family. Default is reject. Explicit strip can
discard a recognizable intersection brand to expose the primitive base; it does
not erase arbitrary object identities or guess a base for an ambiguous intersection.
Enums require their own explicit underlying-value API. Unknown/free structure
blocks. `literal.text` never silently invokes stripping.

Template descriptors can retain symbolic Type holes. Constructing a template uses
native interpolation-domain checking and native union expansion. Broad string and
infinite template patterns are not finite lists of strings. A string parser must
request an exact literal or define a deliberate fallback.

## Enum construction

Two distinct operations are available:

* `enumValues(members)` creates a structural union of supplied literal values.
  It does not introduce nominal identity.
* `enumType(key, members)` creates a native type-only nominal enum. Members are a
  nonempty ordered list of `{name:string, value:string|int|number}` with unique
  names and finite exactly representable values. Values are explicit; automatic
  runtime numeric initialization is not part of the builder.

Neither operation emits a JavaScript enum object or declares a value namespace.
`reflect.enum(T)` exposes underlying scalar values and native nominal member Types.
Native enum assignability remains that of the host compiler version. Duplicate
values follow native enum semantics; they must not be assigned invented brands
solely to force distinctness. `enumValues` is appropriate when structural literal
compatibility is desired.

Nominal identity is keyed by the nearest native entry's declaration identity and
argument identities, the lexical enum-construction operation, and an explicit
data key. It is not keyed by evaluation order, counters, diagnostic display names
or the caller's source location. Calling the same entry twice with the same
arguments yields the same enum identity. A key reused for different definitions
within that owner is a diagnostic. A helper invoked in a loop uses distinct keys
when it needs distinct enums. The runtime must register these definitions
transactionally so a failed evaluation does not publish partial identities.

Declaration serialization preserves the entry, dependencies and logical operation
identity. Stock-TypeScript export of these computed nominal Types requires
materialization into named declarations and is not silently approximated by a
literal union. This is a required integration test before enum support is enabled.

## Recursive type construction

```ts
typefn Tree<Value> {
  return recursive((self: Type) => type({
    value: Value;
    children: readonly self[];
  }));
}
```

`recursive(builder)` allocates a private placeholder Type, passes it to a pure
builder and seals the returned structural Type as the recursive root.
`recursive.group(names, builder)` supports mutually recursive roots; builder
receives an immutable record of the named placeholder Types and must return
exactly one definition per name. Names must be unique data strings known when
the group is allocated.

Unsealed placeholders may occur in construction positions but cannot be reflected,
compared, returned independently, used as cache keys or sent to native operations
that demand completed members. Such operations fail with an unsealed-handle
diagnostic; they do not observe empty objects. Construction involving other
unresolved native parameters can suspend the entry and discards the group.

The completed graph must have guarded cycles: every cycle crosses an object,
array, tuple or signature constructor boundary. `recursive(self => self)` and
unions/intersections of only unguarded placeholders fail. Sealing is atomic and
publishes native graph nodes only after all definitions and validations succeed.
The resulting graph is structurally recursive; it does not acquire class-private
or nominal enum identity. Native recursive Type operations remain bounded and
use cycle-aware traversals.

## Why these APIs are not simply compiler reflection

The VM cannot expose a pointer to the checker's Type struct and call that a
language. Descriptors distinguish declared versus read types, field versus indexer,
array versus tuple, call versus construct signature, open generic versus concrete
signature and structural versus nominal identity. Each distinction is required by
the researched workloads. Any native representation not faithfully expressible
through an editable descriptor must be preserved opaquely or rejected, never
silently flattened.
