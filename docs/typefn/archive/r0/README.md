> Historical R0 draft. Superseded by [the current design](../../README.md).

# Typefn: independent language redesign

Design revision: **R0**, 2026-11-06. This is a specification candidate, not an
implemented or validated language. No DSL compiler implementation is present on
this branch. Examples below describe the candidate language; stock TypeScript
cannot parse them.

Branch `typefn-redesign` starts at upstream TypeScript-Go commit
`89d5d5b2849a0db0957065889ca58536fa6d2e4a`. The previous experiment is preserved on
`typefn-language`, ending in archival WIP commit `700eab37`. Its implementation,
syntax and AST shortcuts are not the baseline for this design.

The objective is an imperative compile-time language whose values include native
TypeScript types. A type function can inspect and construct types without encoding
control flow as conditional aliases or integers as tuple lengths. Its exported
type-valued entry points are used as normal generic types:

```ts
typefn Box<T> {
  return { value: T };
}

type Boxed = Box<string>;

typefn Reverse<T> {
  let shape = reflect.tuple(T);
  shape.elements.reverse();
  return tuple(shape);
}
```

Inside a type function, `type` declares a mutable Type variable. Type initializers,
assignments and returns use native type expression syntax. `let` and `const` bind
ordinary compile-time data. In a data expression, `"x"` is a string and `type("x")`
is a native literal Type; `literal(value)` converts computed data into a Type.
This distinction keeps parser cursors and intermediate records out of TypeScript's
type algebra while keeping common type compositions concise.

```ts
typefn Wrap<T> {
  type result = T;
  result = { value: result };
  return result;
}
```

Assignments capture type values rather than redefine recursive aliases. Outside
type-function bodies, ordinary TypeScript `type` declarations keep their meaning.

## Specification map

1. [Research](../../research.md) and [source manifest](../../research-sources.json): community
   algorithms, edge cases and provenance; 38 inspected files across 17 families.
2. [Design rationale](decisions.md): failed assumptions, alternatives, selected
   model and deliberately excluded capabilities.
3. [Syntax](syntax.md): declarations, expression boundaries, grammar, bindings and
   compile-time data kinds.
4. [Semantics](semantics.md): evaluation, patterns, unions, symbolic types,
   inference and errors.
5. [Type operations](type-operations.md): reflection, construction, signatures,
   enums, recursive types and identity.
6. [Runtime](runtime.md): interpreter, staging, caches, limits and cancellation.
7. [TypeScript integration](integration.md): parser, checker, declarations,
   language server and compatibility.
8. [Worked workloads](workloads.md): complete candidate-language examples and
   their expected behavior, plus boundaries requiring further evidence.
9. [Conformance](conformance.md): design consistency and implementation gates.

R0 chooses one coherent design instead of treating several incompatible sketches
as one language. “Normative” below means required by this candidate, not already
implemented. The specification can be revised when examples expose a problem;
implementation is not allowed to quietly redefine it.

The broad community-coverage objective remains a testable target, not a claim
that every library has been ported. In particular, arbitrary imperative programs
cannot be inverted for generic inference, and constructing new polymorphic
TypeScript declarations is outside the agreed initial scope.
