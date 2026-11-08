# Typefn language specification — R2

Design baseline dated 2026-11-06. Language profile and declaration ABI: **1**.
R2 is an internal design revision label, not a released language version. Earlier
drafts established no public ABI or compatibility obligation.
This specification defines the target language for the personal TypeScript fork. It does **not** claim that the compiler, VM or language server implements it.
Earlier drafts are historical, not an additional source of requirements.
See [implementation status](../implementation.md) for the supported subset.

## Reading and authority

| Document | Contract |
| --- | --- |
| [Syntax](syntax.md) | Lexing, complete core productions, precedence, native Type islands and disambiguation |
| [Semantics](semantics.md) | Bindings, Types, constructors, control, scalar operations, patterns and identity |
| [Standard library](library.md) | Strings, numbers, reflection, structural builders and recursive construction |
| [Evaluation and VM](evaluation.md) | Symbolic computation, guards, substitution, execution, caches and resources |
| [TypeScript integration](integration.md) | Native checking/inference, files, modules, declarations and editor behavior |
| [Conformance](conformance.md) | Required outcomes, community workload coverage and release gates |

These six chapters are normative. MUST, MUST NOT and MAY describe implementation
requirements, prohibitions and permitted choices. Examples illustrate the rules;
their expected results are not reports of executed tests. The [manual](../manual.md)
is informative. Native Type semantics are supplied by the pinned fork baseline
and its compiler options, except for explicit extensions documented here.

The current fork is based on Microsoft TypeScript commit
`f7973164a5f079a2d8b88d0fdb39347222549124`. Earlier research used a different native
baseline. A compiler version change is a different native semantic profile;
compatibility requires conformance checks, not similar printed Types. The current
implementation does not yet satisfy the full target conformance matrix.

## Language in one paragraph

`type` declares a locally mutable binding to an immutable Type or constructor.
`typefn` computes with those values using ordinary sequential control flow.
Literal Types support scalar computation; tuples and object Types hold state and
metadata. Helpers use the same typefn declaration and application model. Native
types remain first-class operands. Complex native syntax is delimited by
`type(...)`. There is no JavaScript sublanguage or second data-function ABI.

## Decisions closed in this baseline

* Keep `F<T>`, using the syntactic ordered-choice rule in the syntax chapter.
  No semantic parsing, whitespace-sensitive calls or ASI.
* Use Type values for scalar results and reflection descriptors. There is no
  source-level int, opaque descriptor kind, let/const or ordinary function inside
  a typefn. Constructors are the sole other value kind.
* Keep number and bigint literal families. Bigint is optional exact integer
  arithmetic; number is finite binary64. No implicit conversion between them.
* Support named helpers across modules, local helpers and constructor closures.
  First-class constructors take ordinary Types and return an ordinary Type.
* Publish known native structure with pending validation. Scalar-producing
  operations can also be deferred; control decisions requiring their values wait.
* Keep native generic signatures and nominal identities by reference. Structural
  builders do not generate native declarations, private brands or fresh enums.
* Provide explicit public projection, overload preservation and guarded recursive
  structural construction. Do not claim arbitrary objects round-trip structurally.
* Specify pure deterministic VM execution with source-based declaration recipes,
  resource diagnostics, cancellation and checker-local caches.

## Deliberate limits

No new native generic declaration construction, native constructor parameters,
constructor inference/subtyping, runtime execution, I/O, regular expressions,
arbitrary inverse inference, or inspection of an active inference session.
Enum-specific construction, classification and reflection are deferred, including
structural enum builders. Existing native enum Types retain their native behavior
when forwarded. Open interfaces remain existing native declarations, not
dynamically generated augmentation points.

The research goal is to express complicated community computations more clearly,
not to promise that every mathematical set or every TypeScript inference behavior
can be reproduced by an imperative program. Coverage and these boundaries are
tracked in the [community workload matrix](conformance.md#9-community-workload-traceability).

## Design and implementation status

The syntax, semantic domains, core library boundary, VM protocol and integration
contracts are specified here. Changes to them require an explicit spec revision.
Implementation and testing are incomplete; the manual's algorithm examples do
not establish grammar or execution conformance.
Resource defaults below are initial implementation targets, not benchmark results.
