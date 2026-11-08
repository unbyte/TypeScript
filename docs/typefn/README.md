# Typefn

An imperative language for calculating TypeScript Types inside TypeScript files.
`type` declares a mutable local binding; `typefn` organizes calculations.

```ts
typefn Reverse<Items> {
  type result = [];
  for (type item of Items) { result = [item, ...result]; }
  return result;
}
type Reversed = Reverse<["a", 2, true]>; // [true, 2, "a"]
```

The compiler implements an experimental subset of the target language. Start with
[implementation status](implementation.md) for supported features and limitations.
The manual and specification also describe features that are not implemented yet.

* [Acceptance tracker](acceptance.md): remaining contracts and completion evidence.
* [Manual](manual.md): language guide and target-language examples.
* [Specification](spec/README.md): syntax, semantics, library, evaluation,
  TypeScript integration and required conformance cases.
* [Architecture](architecture.md): native compiler integration and ownership.
* [Deferred evaluation](deferred-evaluation.md): execution timing, substitution,
  retained checks and current limits.
* [Editor integration](editor-integration.md): setup and supported services.
* [Design decisions](design-notes.md): constructor restrictions and unresolved
  input-constraint semantics.
* [Testing](testing.md): reproducible checks and the scope of validation.

The [editor project](examples/editor/README.md) runs with the implemented compiler.
The [HTTP routes](examples/http-routes.typefn), [calculator](examples/calculator.typefn),
[numbers](examples/numbers.typefn) and [construction](examples/construction.typefn)
files illustrate the target design; they are not executable conformance fixtures.

Superseded drafts, full research reports and historical validation journals remain
in the [preserved snapshot](https://github.com/unbyte/TypeScript/tree/7fe54fd975b95531092485468ac5677ed3a453fc/docs/typefn) on
`archive/typefn-before-doc-cleanup`.
