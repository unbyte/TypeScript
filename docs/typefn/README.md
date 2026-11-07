# Typefn

An imperative language for calculating TypeScript Types, embedded directly in
TypeScript source files. `type` binds a local Type; `typefn` organizes calculations.

```ts
typefn Reverse<Items> {
  type result = [];
  for (type item of Items) { result = [item, ...result]; }
  return result;
}
type Reversed = Reverse<["a", 2, true]>; // [true, 2, "a"]
```

**R2 is the design baseline. An experimental compiler/VM slice is implemented;
see [implementation status](implementation.md) for supported features and gaps.**
There is one computation language, with direct literal-Type arithmetic, string
operations, structural construction and restricted constructor composition.
No JavaScript helper language is embedded inside a typefn.

* [Current TypeScript port](typescript-port.md): repository baseline, integration
  decisions, and validation status for the Go-based monorepo.
* [User manual](manual.md): progressive guide, strings, HTTP bindings, calculator,
  dynamic Types, constructors and generic-input behavior.
* [Formal specification](spec/README.md): syntax, semantics, library, VM,
  compiler/language-server integration and conformance requirements.
* [Deferred calculations](deferred-evaluation.md): implemented native substitution,
  retained checks, editor behavior, and remaining generic-integration limits.
* [Where-clause investigation](where-constraints-investigation.md): Rust and
  TypeScript evidence, executable constraint probes, and the recommendation to
  retain `extends` while deferring a broader contract language.
* [Design/implementation audit](design-implementation-audit.md): verified fixes,
  specification coverage, remaining design-validation and implementation gates.
* Complete examples: [HTTP routes](examples/http-routes.typefn),
  [calculator](examples/calculator.typefn), [numbers](examples/numbers.typefn),
  [dynamic construction and composition](examples/construction.typefn).
* [Research](research.md), [source manifest](research-sources.json) and
  [TypeScript HKT constraint audit](typescript-hkt-audit.md).

The [R1 design](archive/r1/README.md), including its JavaScript-shaped calculator
and host-only algorithm harness, is historical. [R0](archive/r0/README.md) is also
archived. Neither supplies extra normative rules or proof of R2 implementation.
