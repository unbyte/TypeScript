# Design decisions and open constraints

The [specification](spec/README.md) defines the target contract. This note preserves
current rationale and unresolved questions; it does not add syntax or claim that
unimplemented features work.

## Constructor composition

The design restricts constructor values to known arity, ordinary Type inputs and an
ordinary Type result. It excludes native constructor parameters, constructor
inference/subtyping and arbitrary new native generic declaration construction.
Existing generic signatures and nominal identities remain native handles. The
implementation has not yet enabled constructor slots, closures or local typefns.

These restrictions avoid treating arity alone as a complete kind: bounds, binder
scope, variance and input/output relationships also affect composition. TypeScript's
[HKT issue #1213](https://github.com/microsoft/TypeScript/issues/1213) discusses these
constraints; it is not evidence that unrestricted HKT is impossible or that this
restricted design is already implemented. The probes in
`probes/hkt-baseline.ts` and `probes/check-hkt-baseline.mjs` record native behavior.

## Input constraints

Keep native `extends` bounds for now. A constraint proves eligibility; it does not
supply a literal value to execute, enumerate a union, or authorize an inverse
inference rule. Assignability, class ancestry and equality are distinct contracts.
In particular, special `any`/`never` behavior and conditional-Type distribution
cannot be replaced with an ordinary logical conjunction.

Rust's [where clauses](https://rust-lang.github.io/rfcs/0135-where.html),
[trait bounds](https://doc.rust-lang.org/reference/trait-bounds.html) and
[parameter environments](https://rustc-dev-guide.rust-lang.org/typing-parameter-envs.html)
provide useful context, but importing the syntax does not import Rust's solver.
TypeScript constraints already participate in inference; relevant open problems
include [lower bounds](https://github.com/microsoft/TypeScript/issues/14520),
[generic narrowing](https://github.com/microsoft/TypeScript/issues/13995) and
[circular constraints](https://github.com/microsoft/TypeScript/issues/51011).

Three existing behaviors need a contract decision before adding `where`:

* An earlier unresolved parameter bound delays checking later bounds. Parameter
  order can therefore affect diagnostic timing.
* Checking a header bound can evaluate an application at declaration time, even
  when the surrounding typefn is never called.
* Later-parameter references in bounds are accepted, while the target specification
  describes preceding-parameter scope.

The executable observations are `probes/where-baseline.ts`,
`probes/check-where-baseline.mjs` and `probes/check-where-native.py`.
[Testing](testing.md) gives their commands. They record current behavior, including
known mismatches; they are not `where` conformance tests.

A possible future `where` limited to one upper bound per declared parameter would
need equivalence to inline bounds, fixed parameter-order checks, defined scope and
replay, and matching inference/diagnostics. General lower-bound, equality, trait or
proof clauses require a separate solver contract. None is enabled by this note.

## Evidence and workload coverage

The [community workload matrix](spec/conformance.md#9-community-workload-traceability)
retains concrete obligations for merge/serialization, paths, parsers, schema and
query libraries, correlated unions and constructor composition. Passing a small
algorithm example does not establish compatibility with those libraries.

The full investigations, inspected-source manifest and superseded proposals remain
in the [preserved research snapshot](https://github.com/unbyte/TypeScript/tree/7fe54fd975b95531092485468ac5677ed3a453fc/docs/typefn). Keep new conformance evidence in
executable fixtures and current contract decisions in the specification, rather than
reintroducing a parallel historical document set here.
