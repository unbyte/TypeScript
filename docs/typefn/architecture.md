# Compiler architecture

The fork is based on Microsoft TypeScript main at
`f7973164a5f079a2d8b88d0fdb39347222549124`. Typefn uses the native AST, binder,
Type identities and language-service pipeline. It does not lower bodies into
conditional-Type text or run a separate structural type checker.

## Front end and configuration

`compilerOptions.typefn.version: 1` enables the language. Shared option metadata
in `tools/scripts/tsc` generates native options, API types and JSON schemas. The
profile and parse-affecting limits participate in source-file cache identity.

`tools/scripts/tsc/ast.json` owns syntax metadata and generated Go/API representation.
The bounded statement/Pratt parser lives in `tsc/internal/parser/typefn.go`.
Dedicated nodes preserve source positions and recovery information. Native visitors
visit operands and operators in source order; the encoder keeps schema field order
on the wire. Generated API traversal exposes source order again to clients.

The binder installs declarations and locals in native Type-space scopes, without a
JavaScript control-flow graph. Lowering validates every branch and resolves lexical
slots before execution. Recovered syntax is available to editor services but never
becomes executable VM code.

## Checker and VM ownership

All modules below are private to `tsc/internal/checker`:

| Module | Responsibility |
| --- | --- |
| `typefn_program.go` | Definition validation, resolved bindings and lowered control flow |
| `typefn_vm.go` | Root budgets, cancellation, recursion checks and call frames |
| `typefn_expression.go` | Expression evaluation and structural construction |
| `typefn_native.go` | Native applications, defaults and parameter bounds |
| `typefn_values.go`, `typefn_operators.go` | Scalar observation, iteration domains and operators |
| `typefn_readiness.go` | Cycle-aware native argument/lexical-dependency summaries for structural observation |
| `typefn_pattern_compile.go`, `typefn_pattern.go`, `typefn_pattern_template.go` | Arm lowering, structural matching and bounded UTF-16 decomposition |
| `typefn_types.go`, `typefn_deferred.go` | Immutable pending terms, retained checks and substitution |
| `typefn_signatures.go`, `typefn_signature_reflection.go`, `typefn_signature_builder.go`, `typefn_function.go` | Native signature forwarding, metadata, validation and construction |
| `typefn_state.go` | Lazy checker-owned tables and active root VM |
| `typefn_display.go`, `typefn_services.go` | Pending display, signatures and cancelable result queries |

Registers hold native Type identities. Assignment changes a local slot, not a Type.
Lowering assigns dense iterator indices; each call owns its iterator storage, and
non-iterating programs allocate no iterator backing storage. Re-entry replaces its
own slot while iteration retains the original immutable subject.

The live-frame estimate includes register/argument pointers, iterator storage and a
bookkeeping allowance. It does not account for all Go heap or native checker work.
Instruction, call-depth, construction and bigint limits share one evaluation root;
nested calls cannot reset them. Unsupported resource-limit overrides diagnose.

One lazy `typefnState` owns lowered programs, interned terms, constructed-object and
syntax-dependency tables, and the active root VM. These belong to a checker snapshot.
No Type handles or mutable frames cross snapshots. Canceled checkers are discarded
by the native checker pool. There is no general completed-call memoization cache.

## Native checking boundaries

Immutable deferred terms use the native tagged Type representation without enlarging
its header. Native substitution replays changed operands. Reached checks survive
supported native simplifications and cannot be replaced by an unconditional successful
result. Failed shared signature results remain replayable at each use site.
See [deferred evaluation](deferred-evaluation.md) for the implemented boundaries.

Disabled-profile checking bypasses DSL declaration scans and extra constraint-node
queries. Native operand processing avoids obligation scans when no DSL state exists;
unguarded shapes bypass journal deduplication. Native declaration diagnostics also
skip whole-AST DSL scans when the profile is disabled. Native allocation tests cover
the existing composite-cache fast paths.

## Emit, editor and API

The source printer round-trips DSL nodes. JavaScript emit erases declarations;
declaration emit diagnoses until executable source recipes can be serialized.
Normal imports, namespaces, aliases and symbol identities supply composition.

Editor providers use native symbols, source mappings and scanner tokens. Completed
results may appear in hover and opt-in hints; pending values and recovered errors do
not publish successful-result overlays. The generated async/sync APIs expose the
same syntax and native Types. See [editor integration](editor-integration.md).

[Testing](testing.md) records fixture, project-invalidation, cancellation, mapped
projection, protocol and generated-file checks. [Implementation status](implementation.md)
separates these implemented boundaries from the target specification.
