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
| `typefn_value.go`, `typefn_closure.go` | VM value/Type boundary, immutable helper captures, interning and substitution |
| `typefn_program.go`, `typefn_flow.go` | Definition validation, resolved bindings, lowered control flow and conservative return paths |
| `typefn_vm.go` | Root budgets, cancellation, recursion checks and call frames |
| `typefn_expression.go` | Expression evaluation and structural construction |
| `typefn_island.go` | Immutable native capture scopes, native checking and application diagnostics |
| `typefn_native.go` | Native applications, defaults and parameter bounds |
| `typefn_values.go`, `typefn_operators.go` | Scalar observation, iteration domains and operators |
| `typefn_readiness.go` | Cycle-aware native argument/lexical-dependency summaries for structural observation |
| `typefn_pattern_compile.go`, `typefn_pattern.go`, `typefn_pattern_template.go` | Arm lowering, structural matching and bounded UTF-16 decomposition |
| `typefn_types.go`, `typefn_deferred.go` | Immutable pending terms, retained checks and substitution |
| `typefn_signatures.go`, `typefn_signature_reflection.go`, `typefn_signature_builder.go`, `typefn_function.go` | Native signature forwarding, metadata, validation and construction |
| `typefn_tuple_metadata.go`, `typefn_tuple_descriptors.go` | Tuple descriptors and declared-value provenance through native normalization, mappings, inference and copies |
| `typefn_properties.go` | Public property projection and declared/optional-read provenance |
| `typefn_state.go` | Lazy checker-owned tables and active root VM |
| `typefn_display.go`, `typefn_services.go` | Pending display, signatures and cancelable result queries |
| `typefn_abi.go`, `typefn_emit.go` | Declaration ABI validation, lexical recipe dependencies and emit validation |

Registers hold explicit VM values containing a native Type or helper closure.
Native operations check that boundary before using operands. Assignment
changes a local slot, not a Type.
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

## Constructor values and integration boundary

`Constructor(n)` stays separate from native Type handles. Current local helper
closures contain eligible code and frozen Type/helper captures. They are not new
TypeFlags cases, invented native parameters, or object/tuple fields. Native
operators and structural builders accept only the Type alternative.

Local self-reference names the same code and closure environment without a mutable
capture cycle. Snapshots are interned per checker; mutating a caller register never
changes a published capture. Native substitution maps captured Types and traverses
constructor edges iteratively. Capturing alone does not expand native members or
execute a body. New snapshot storage is charged to the current root; complete
retained-cache accounting/eviction remains pending.

Pending Type results retain call recipes with closure identities and ordinary Type
arguments. Deferred expressions retain immutable VM-value snapshots. Their keys,
substitution and diagnostics include captured values; constructors never travel in
native `[]*Type` operand lists. This supports named and anonymous closures, constructor bindings and factory
results, including symbolic replay. Native generic constructor values and
constructor input slots remain pending.

Definition checking assigns binding/result kinds independently of execution,
propagates constructor arities through applications, and solves recursive result
dependencies with finite kind constraints and explicit annotations. Dependency
components finalize before callers can constrain their application results. Native Type positions reject
constructor values. The separate native alias bridge is restricted to nongeneric
aliases whose constructor identity and creation checks are ready, as required by
the specification. This bridge also remains necessary for helper-constructor captures in native
islands, which currently diagnose rather than losing their environments.

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
declaration emit preserves written recipes and retains lexical dependencies through
native declaration visibility. ABI and invalid-recipe markers live in the emit
context, so comment removal cannot remove the ABI and forced emit cannot publish
a recipe marked invalid. Inferred-recipe lifting remains incomplete.
Normal imports, namespaces, aliases and symbol identities supply composition.

Editor providers use native symbols, source mappings and scanner tokens. Completed
results may appear in hover and opt-in hints; pending values and recovered errors do
not publish successful-result overlays. The generated async/sync APIs expose the
same syntax and native Types. See [editor integration](editor-integration.md).

[Testing](testing.md) records fixture, project-invalidation, cancellation, mapped
projection, protocol and generated-file checks. [Implementation status](implementation.md)
separates these implemented boundaries from the target specification.

## Native island cache isolation

An island's native subtree is cloned and bound beneath a checker-owned scope whose
alias symbols hold the frozen native capture Types. The subtree keeps original
source positions and an outward link to the original lexical scope. Inner mapped,
infer and signature binders use native symbols; they never resolve by textual
replacement. Original parsed nodes and their checker caches are untouched.

For now, a reachable unresolved DSL capture retains the original island and capture
snapshot as a deferred expression. This is conservative and costs shape information;
stable symbolic native shape publication is still an acceptance item. Executions
also allocate fresh native subtrees rather than reusing completed environment keys.
Neither strategy is presented as the finished performance architecture.

Factory calls use the VM-value boundary and may return a constructor. A factory
with pending creation checks blocks its enclosing Type computation; it does not
publish a constructor with hidden requirements or encode one as a native Type.
Type-valued bindings and arguments retain their own checks in addition to the
frame journal, so a native shape cannot bypass factory readiness after storage.

DSL signatures use `TypeFunctionParameter` nodes, distinct from native
`TypeParameterDeclaration`. An ordinary parameter has a Type kind and an optional
DSL bound expression; a constructor parameter records an exact slot count. This
keeps DSL bounds out of native declaration-time constraint evaluation and permits
the full TypeArgument expression grammar without changing native generics.
Programs lower bounds against preceding parameter slots and frozen captures, then
validate them after arguments/captures are installed in the call frame. Header
names and kinds are checked without executing the bound. Constructor parameter
syntax is retained while its mixed-value call adapter remains pending.
