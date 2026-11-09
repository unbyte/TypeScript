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
* [Executed community workloads](workloads.md): compared algorithms, evidence and limits.
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
The [calculator](examples/calculator.typefn) and [numbers](examples/numbers.typefn)
are executable modules: the integration harness reads these exact files as TypeScript
and checks their native results, including generic specialization and invalid inputs.
Copy them to `.ts` files to use them in an enabled project.
[HTTP routes](examples/http-routes.typefn) also executes in the harness. Calculator
and route tests additionally consume their emitted declarations in fresh checkers.
[Construction](examples/construction.typefn) remains a target-design example.

Superseded drafts, full research reports and historical validation journals remain
in the [preserved snapshot](https://github.com/unbyte/TypeScript/tree/7fe54fd975b95531092485468ac5677ed3a453fc/docs/typefn) on
`archive/typefn-before-doc-cleanup`.
