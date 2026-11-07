> Historical R1 material, superseded by the [R2 specification](../../spec/README.md). Not the current language contract.

# Declaration emit and cross-module execution

R1 normative design candidate for the [restricted HKT scope](restricted-hkt.md).
It complements the [evaluation contract](evaluation.md). These rules specify
source/declaration behavior before implementation; the examples are required
outputs or equivalent outputs, not a demonstration of a working emitter.

## 1. Preserve the computation that defines the public type

Fork declarations can contain typefn bodies. Emission preserves the code and
dependencies needed to reproduce an exported computation, including pending
validation. It does not serialize VM frames, native Type pointers, caches, quotas
already consumed, or a dump of the current closure heap.

```ts
// library.ts
typefn WithError<E> {
  return typefn<A> {
    return { ok: true; value: A } | { ok: false; error: E };
  };
}

export type Result = WithError<Error>;
export type Message = Result<string>;
```

An emitted declaration can preserve this recipe:

```ts
// @typefn-abi: 1
typefn WithError<E> {
  return typefn<A> {
    return { ok: true; value: A } | { ok: false; error: E };
  };
}
export type Result = WithError<Error>;
export type Message = Result<string>;
```

WithError is module-local, not a new public export. Its retained body is required
to recreate Result's closure. Bodies and ordinary lexical data helpers remain
executable compile-time definitions in declaration files; they produce no JS.
Factories, aliases and consumers use ordinary type-only import/export resolution.

An exported named typefn retains its declaration, signature, source-equivalent
body and referenced dependencies. The emitter may rename private helper symbols
hygienically, but must not alter binding, capture time, validation order or native
declaration identity. Generic native declarations do not gain constructor slots.

## 2. ABI header and compatibility

The canonical declaration header is `// @typefn-abi: 1`, before nontrivia source.
It is a semantic pragma in this fork. Emit it in each generated declaration file
containing DSL definitions or constructor-valued declarations. Whitespace may be
normalized; formatters/emission must preserve the directive. It introduces no
network fetch, executable JavaScript import or external code loader.

When consuming a declaration file containing this syntax, a missing, conflicting
or unsupported ABI header is a diagnostic before any VM execution. Ordinary
declaration files without DSL syntax need no header. Source .ts files use the
fork's configured DSL version; users do not need to annotate every source file.

ABI 1 covers grammar, interpreter-value representation, prelude intrinsic contracts
and staged evaluation behavior. It does not claim identical native TypeScript
behavior across arbitrary compiler versions/options. Round-trip comparisons use
the same pinned fork build, native options, libraries and module augmentations.
Changes to DSL semantics require an explicit ABI compatibility decision.

There is no stock-TypeScript declaration fallback in this design. The normal
emitter must not replace executable definitions or unresolved Types with any,
unknown, an approximation or only the currently visible shape. A future explicit
materialization command is a separate feature, not silently selected by flags.

## 3. Captures: replay source semantics, not current slot values

Each constructor/helper closure captures immutable values at its creation point.
Emission preserves the containing typefn body and reconstructs that moment by
replaying it. This handles mutated locals, native Type references, data records,
exact integers, UTF-16 strings and lexical helper closures without inventing a
general object serializer for opaque compiler handles.

```ts
export typefn Save<T> {
  type current = T;
  type saved = typefn<X> { return [current, X]; };
  current = number;
  return saved;
}
export type Saved = Save<string>;
export type Pair = Saved<boolean>; // [string, boolean], not [number, boolean].
```

The emitted Save body preserves the assignment order. Printing a lambda that reads
the final `current` would be incorrect. Moving the lambda to module scope without
binding its captured value would also be incorrect. Source recipes preserve the
same rule for data-helper captures and descriptor values.

Native symbols retain their module/declaration identity through imports and
source/declaration redirects. A private class or unique symbol cannot be recreated
from its printed structural shape. Module augmentation is resolved in the consuming
program normally; results may legitimately change when its native environment
changes. No cached producer Type graph is authoritative in the consumer.

## 4. Inferred signatures and local constructor aliases

An inferred exported native Type may reference a local DSL calculation with no
directly printable module-level name. The emitter creates a private Type-returning
typefn containing the original recipe. Its explicit ordinary Type parameters
represent free native binders; captured constructors are recreated from known
definitions/factory calls. The helper does not add an F parameter to a native
function, and does not require chained constructor application syntax.

For example, suppose an exported native function's return Type originates from
a lexical alias `C<string>`, where `C = WithError<T>` and T is that function's
native Type parameter. A valid declaration shape is:

```ts
// @typefn-abi: 1
// WithError's dependency declaration is retained as above.
typefn __typefn_export_0<T> {
  type C = WithError<T>;
  return C<string>;
}
export declare function message<T>(error: T): __typefn_export_0<T>;
```

The helper must have enough provenance to reconstruct every captured value at its
original creation point. It must not read a runtime parameter's value: native T
is a Type handle, while `error` is a runtime value and is unavailable to the VM.
Captured native Type expressions such as a nameable `typeof` reference remain
native Type dependencies, not permission to execute that reference.

Generated names are deterministic for a given source dependency graph, collision
checked against real declarations, and mapped to the original source locations.
Different source files are not allowed to merge unrelated closure identities by
sharing a generated name. There is no required stable bytecode or private helper
name across source edits; public source/declaration symbol redirects must agree
within a build graph.

If the compiler cannot reconstruct a source recipe with nameable dependencies,
emission diagnoses the escaping dependency. It does not guess a structural
replacement, clone nominal identity, drop validation or fall back to any. The
diagnostic requests a nameable typefn/alias or an explicit exportable annotation.
This is a defined failure mode, not a promise that every anonymous native Type is
serializable. Native accessibility/nameability checks still apply.

## 5. Guards and deferred computations survive emission

```ts
typefn Checked<T> { return Uppercase<T>; }
export typefn Discard<T> {
  type unused = Checked<T>;
  return string;
}
export type Hidden<T> = Discard<T>;
```

Hidden<T> must not emit as `type Hidden<T> = string`. The dependency on Checked
is executable validation even though the visible result shape is string. Emit
Discard's body, Checked's body and the original application. After a separate
consumer specializes Hidden<number>, the same bound failure must occur.

OperationType and CallType are not printed as internal native-type tags. Emit the
source/native operator recipe or a private helper that reconstructs it. Known
payloads with guards retain those guards by emitting their computation. Native
generic signatures retain the explicit Type expressions that carry their pending
requirements; emitting only their apparent shape changes their accepted calls.

Default declaration emission therefore prefers the originating DSL recipe even
for currently concrete results. There is no unrequested constant-folding export
optimization that drops a dependency on an open interface, constructor identity
or a validation call. Any later optimization must prove it preserves checks and
consumer-visible native dependencies, not just printed Type equality.

## 6. Dependency collection and module boundaries

Collect dependencies transitively from signatures, bodies, captured source recipes,
native Type expressions, local helper annotations, all branch bodies and static
bindings. A dependency in a currently unchosen branch still matters when a consumer
uses different arguments. This is lexical dependency collection, not tracing only
one successful VM execution.

Keep module-local helper definitions needed by exports. Preserve external references
with type-only imports or qualified native names using normal module resolution.
Do not inline a foreign class, enum, unique-symbol declaration or native generic
body into a fresh nominal identity. A reference to an inaccessible external symbol
is a nameability diagnostic. Re-exported typefn symbols preserve the defining
module's code identity rather than creating independent functions by textual copy.

Mutually recursive named typefn definitions can be emitted as an ordered dependency
component; they need explicit result-kind anchors where inference is cyclic. At
execution, an actual repeated application is still subject to cycle/depth checks.
Declaration import cycles do not license evaluating a partially bound module.

User-defined data helpers remain lexical to a typefn. Their source and captures
travel with that body. This settles how existing helpers are emitted; it does not
add a new cross-module data-function language or solve the separate ergonomics
question of sharing such helpers directly between parser libraries.

## 7. Compiler, CLI and editor behavior

Declaration emission uses the same binder, kind checker and evaluation protocol
as normal checking. It must not force a generic computation merely to obtain a
pretty-printed result. Known invalid closed applications diagnose; unresolved
generic applications emit their recipes. Source errors must not produce a forged
successful DSL definition, including when ordinary emit-on-error is enabled.

With noEmitOnError, honor the native whole-program policy. Otherwise native outputs
can follow native policy, but a declaration output whose DSL dependency closure is
invalid must be withheld with a diagnostic. Never publish a truncated body, omitted
guard or fabricated return Type. Output for an affected declaration file is atomic.
Incomplete/cancelled editor work publishes no declaration file for that snapshot.

`skipLibCheck` does not authorize executing malformed or ABI-incompatible DSL code.
The frontend/bytecode safety checks needed to execute an imported body remain
mandatory. Checking native declarations unrelated to DSL execution follows normal
options. Compilation of validated bodies can be cached within their snapshot.

Hover may show known shape plus a concise pending-call explanation. It must not
present a guarded Type as an unconditional completed result. Go-to-definition,
rename and declaration maps follow original bindings and generated-helper mappings.
Application errors in a consumer point to that consumer's use with related locations
in the defining declaration/source map. Internal helper names are secondary detail.

The Go baseline's [declaration transformer](../../../../internal/transformers/declarations/transform.go)
and [checker nodebuilder](../../../../internal/checker/nodebuilder.go) need explicit DSL
recipe handling before normal apparent-Type serialization. Type-only dependency
tracking, [declaration tracking](../../../../internal/transformers/declarations/tracker.go)
and source maps must include body references. This is a concrete entry-point map,
not evidence that every generated visitor/API path has already been implemented.

## 8. Required round-trip checks

Compare source consumption and declaration-only consumption in separate checker
instances, with identical native settings. Delete producer source from the consumer
fixture so it cannot accidentally follow a source redirect. Also test redirects
separately. The corpus must include:

* Named functions, constructor aliases, returned closures and snapshot captures.
* Symbolic captured Types, renamed binders and lifted local aliases.
* Native constraints, conditional distribution and defaulted native declarations.
* Guard-only dependencies: unused locals, erased native alias arguments, any unions.
* Closed failures with correct consumer provenance and generic pending calls.
* Non-exported local helpers, transitive type-only imports and module augmentation.
* Private/unique-symbol identity, inaccessible dependencies and generated-name clashes.
* ABI mismatch/missing header, invalid declaration bodies, skipLibCheck and cancellation.

These are acceptance requirements. No declaration-only round-trip test is reported
as passing until the fork actually supports the syntax and semantic protocol.
