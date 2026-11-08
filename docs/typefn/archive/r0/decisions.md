# Decisions and reasons

## Failure review

The earlier work failed in process as well as coverage:

* It chose a small implementable syntax before establishing the workloads. That
  made missing features look like extensions to a correct core when the core's
  value model had not been settled.
* It treated concrete instantiation tests as a proxy for generic API usability.
  An opaque `Box<T>` that hides `.value` is unusable in ordinary generic code.
* It presented a few utility types as representative of schema, SQL, routing,
  parser and inference-heavy libraries. Their important contracts are different.
* It reused JavaScript AST positions for type expressions. This reduced initial
  edits but undermined typed visitors, tooling and the grammar's independence.
* It confused successful regression checks and one optimization with maturity.
* It repeatedly stopped at a proposed next phase despite an ongoing implementation
  request. Conversely, after being asked for deeper design, it kept advancing
  implementation before the design was coherent.

Corrective rule: requirements, semantic decisions, worked examples and conformance
criteria must precede the new frontend and VM. The old code is preserved as
experimental evidence; none of its implementation shortcuts is a requirement.

## Competing models

### A. Every value is a native TypeScript type

This makes `return {value: T}` compact and reuses native expressions everywhere.
However, parser cursors, integer arithmetic, ordered token lists and descriptors
then become types too. We would reproduce tuple arithmetic, distributive
conditionals and artificial object types for state. Rejected as the primary value
model; it does not follow that mutable Type bindings themselves should be rejected.

### B. Run ordinary JavaScript with a checker API

This provides good loops and libraries, but brings JavaScript coercions, prototypes,
dynamic property lookup, host effects and an oversized runtime. Allowing arbitrary
source/runtime calls would complicate dependency tracking and caching. A restricted
JavaScript interpreter could work, but its restriction boundary would still need
a new language specification. Rejected as an implicit execution model.

### C. A small typed imperative language with first-class Type values

Selected. Data is data; native types are opaque immutable values, held in mutable
or immutable bindings. A reflection library exposes deliberate semantic
descriptors instead of Go checker internals. The interpreter controls all effects.

The user's explicit preference is that `type` take the role of a variable
declaration for type calculations. Consequently, inside type functions:

```ts
typefn Box<T> {
  type result = T;
  result = { value: result };
  return result;
}
```

`type result` owns a variable containing a native Type handle; assigning does not
mutate that handle. `const type Result = ...` is the immutable counterpart. `let`
and `const` are retained for ordinary data. This deliberately gives a local `type`
declaration a different role from a top-level TypeScript alias.

There are two expression modes: Type and data. Type declarations, assignments to
Type variables and Type-returning bodies provide a Type context. Data contexts
embed native syntax with `type(...)`. This is a deliberate concession to concise
type calculations, not a hidden coercion or a decision made by running the VM.
The syntax specification defines how those contexts are determined.

## Surface decisions

| Question | R0 decision | Reason |
| --- | --- | --- |
| Mutable type variables | `type result = never` | User-selected role for `type` inside calculations |
| Immutable type variables | `const type R = NativeType` | Explicit counterpart without making handles mutable |
| Native type expressions | Direct in Type contexts; `type(...)` in data contexts | Concise calculations with explicit domain crossings |
| Computed literals | `literal(value)` | Makes conversion explicit |
| Type functions | Native entry `typefn F<T>`; data helper `typefn f(x: Kind): Kind` | Native use and ordinary compile-time calls have distinct adapters |
| Native calls in DSL | `F(T)` as a function value; `type(F<T>)` also allowed | No `<` versus comparison ambiguity in data expressions |
| Data type aliases | `data Name = Kind` | Does not overload native `type` with a second type system |
| Matching | `match (value) { pattern => { ... } }` | Explicit binding scope, no implicit result or fallthrough |
| String captures | Template patterns, `${name}`, `${_}`, `${=delimiter}` | Capture, discard and computed literal are distinguishable syntactically |
| Union processing | Explicit `members(T)`, `union(...)`, `intersection(...)` | Whole-type calls do not silently distribute |
| Higher-order functions | Ordinary function values and lexically captured lambdas | Avoids HKT encodings inside the DSL |
| Local data mutation | Value semantics with updates rooted in `let` | Saved values and cached results cannot change through aliases |
| Reflection | Native handles plus immutable descriptors | Does not lose identity by converting everything to fields |
| Unresolved operations | Suspend the native entry, never substitute a constraint | Keeps specialization sound |

`data` is a small but intentional addition. Reusing native `type` aliases to describe
interpreter records would entangle ordinary TypeScript assignability with the
interpreter's own kind checking. Omitting named data kinds would make parser
results and recursive syntax trees needlessly repetitive. Named kinds can be
structural, recursive records and tagged unions; they do not enter the native
type namespace as TypeScript object types.

## Capabilities and exclusions

R0 includes compile-time scalars/collections, local mutation, helpers/closures,
string and structural patterns, native type algebra, explicit union processing,
reflection/reification, nongeneric signature construction, recursive structural
types, nominal type-only enums, symbolic construction and native inference over
published structure. Existing generic signatures may be preserved opaquely.

R0 excludes generating new generic parameter lists or generic declarations,
runtime code generation, new mergeable interface declarations, runtime enum
objects, arbitrary host execution, observing mutable inference sessions, untyped
metaprogramming and arbitrary inverse inference. An interface-like object Type is
different from an `interface` declaration with module augmentation behavior.

No language can recover distinctions already erased from the supplied type or
represent every mathematical complement inside TypeScript's type algebra. Such
boundaries must be explicit failures, retained symbolic computations, or documented
approximations requested by the algorithm—not silent changes of meaning.

## Design-review method

For each primitive, ask: what kind is each argument, what information is observed,
what happens with free type parameters, what identity is preserved, what can fail,
what work is charged, and what editor location owns that failure? The other
specification files answer these questions. The workload file then checks whether
those answers compose into readable algorithms.
