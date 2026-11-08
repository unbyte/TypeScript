# Native compiler and editor integration

This document describes implemented behavior on `feat/typefn`. The complete
language remains experimental; see [implementation status](implementation.md)
for execution gaps and [the specification](spec/integration.md) for the target.

## Using the fork in an editor

Build the compiler and extension from this checkout:

```sh
go build -o built/local/tsc ./tsc/cmd/tsc
npm run extension:build
```

Run the extension in a VS Code Extension Development Host with `packages/vscode-typescript` as
its extension development directory. In development mode it discovers
`built/local/tsc`. The repository's `.vscode/launch.template.json` includes an
extension-host launch template.
A separately installed native TypeScript extension can select `built/local` with
`js/ts.tsdk.path`; check that the selected executable is this fork.

Enable the language in the project's `tsconfig.json`:

```json
{
  "compilerOptions": {
    "strict": true,
    "typefn": { "version": 1 }
  }
}
```

Enable the native server and optional result overlays in editor settings:

```json
{
  "js/ts.experimental.useTsgo": true,
  "js/ts.inlayHints.typefnResults.enabled": true,
  "editor.semanticHighlighting.enabled": true
}
```

A [runnable example project](examples/editor/README.md) is included.

The result-hint setting is declared by this fork's extension. LSP clients using
the server directly can provide `includeInlayTypefnResultHints: true` in the
native user preferences. Result hints default to off.

```ts
// scale.ts
export typefn Twice<N> {
  type result = N;
  result = result + N;
  return result;
}

// consumer.ts
import type { Twice } from "./scale";
type Answer = Twice<21>;
const answer: Answer = 42;
```

Hover on the application shows the declaration signature and `Twice<21> = 42`.
An enabled result hint displays ` = 42` after the application. Assignment checking,
imports, references, and rename use the native TypeScript program.
Generic applications can now remain pending and replay under native substitution.
Pending expressions display their calculation; known shapes with outstanding
checks display a `pending validation<...>` label. These applications do not receive
completed-result overlays. See [deferred calculations](deferred-evaluation.md).

## Implemented services

| Service | Behavior |
| --- | --- |
| Binding | Declarations, parameters, mutable locals, nested blocks, and iteration bindings have native symbols and scopes. |
| Definition, references, rename | Resolve symbol identities, including shadowed locals and imported declarations across files. An iteration binding belongs to its body; its iterable resolves in the enclosing scope. |
| Hover | Shows `typefn Twice<N>: type`, declaration documentation through the native hover path, and native application results. Mutable locals show `type result: type (mutable binding)`; iteration variables show their binding kind. |
| Completion | Offers visible Type-space symbols and DSL keywords. Namespace member completion filters out value-only exports. The compiler prelude supplies intrinsic namespaces and members. Native type islands use the native completion path. |
| Signature help | Supports native applications and DSL angle calls, including unfinished `F<`, commas, and nested argument delimiters. Names must resolve to a typefn or an implemented prelude intrinsic. |
| Semantic highlighting | Classifies bound Type names, declarations, assignments, and immutable iteration bindings. Native scanner tokens supply DSL keywords, operators, strings, and numbers. |
| Intrinsic help | Built-in calls expose parameter/default signatures in hover and signature help, including incomplete angle calls. Intrinsic namespace/member names receive semantic coloring. |
| Result hints | Opt-in computed results at native TypeScript type applications, including nested applications. |
| Outline and folding | Lists typefn declarations and local bindings; folds DSL blocks. |
| Formatting | Uses the native formatter for spacing, indentation, and comments. Preserves required DSL semicolons even with the JavaScript semicolon-removal preference. |
| Diagnostics and edits | Use the native project snapshot and checker pool. Imported-definition edits and profile changes invalidate the affected results and syntax. |

Syntax coloring here is delivered through the server's semantic-token protocol.
It requires a running forked server and a client/theme with semantic highlighting.
There is no separate TextMate grammar for coloring a disconnected editor. Multiline
literal tokens remain with the editor's existing lexical grammar.

The implemented local hover reports the binding's kind. A generic function can run
with different inputs and mutate a local many times; this implementation does not
assign that local a fictitious single computed value. Flow-sensitive symbolic
local previews and execution tracing remain future work.

## Compiler boundaries

The integration shares the native source file, AST, binder, symbol tables, Types,
project lifetime, and language-service request infrastructure:

1. The opt-in native parser creates dedicated DSL nodes directly in the source AST.
   Bodies retain their small DSL grammar. Ordinary TypeScript syntax continues to
   use its existing parser.
2. The native binder installs DSL declarations and lexical bindings. VM lowering
   keeps these symbol identities alongside its frame slots. References can resolve
   before a function has been compiled or evaluated.
3. Native type-reference checking invokes the checker-owned VM. The VM consumes
   and returns native Type handles, so assignments and downstream native generic
   operations use the actual checker relations.
4. Editor providers query those same symbols. A dedicated checker query obtains
   application results with the request cancellation context. No per-invocation
   values are written onto shared AST nodes or binder symbols.
5. Canceled computations mark their checker unusable. The editor pool disposes it;
   a later request acquires a fresh checker. Result hints omit recovery/error Types.
   Recovered consumer syntax and malformed producer files are not evaluated.
   Semantic checking also refuses to lower syntax-invalid source files. Native
   operation failures propagate through discarded concrete VM values, so they do
   not acquire successful result overlays.
6. Native AST visitors traverse operators in source order. The API serializer keeps
   its schema field ordering, including the compact operand lists. Remote API
   operator traversal exposes source order independently of that wire layout.
   Generated transforming visitors dispatch every DSL kind and factory updates
   retain each node's kind; runtime tests cover local and remote AST transformations.

Key implementation files:

- `tsc/internal/ast/typefn.go`, `tsc/internal/binder/binder.go`: syntax traversal and scopes.
- `tsc/internal/checker/typefn_program.go`, `typefn_vm.go`, `typefn_expression.go`,
  `typefn_native.go`, `typefn_values.go`, and `typefn_operators.go`: lowering,
  frame execution, native applications, construction, and scalar operations.
- `tsc/internal/checker/typefn_state.go`, `typefn_services.go`: checker-local
  ownership, signature descriptions, and cancelable result queries.
- `tsc/internal/checker/typefn_deferred.go`, `tsc/internal/checker/typefn_display.go`:
  immutable deferred terms, substitution/replay, guards, and pending Type display.
- `tsc/internal/ls/typefn_symbols.go`, `typefn_completion.go`, `typefn_tokens.go`
  and the native LS providers: completion, signature help,
  scanner coloring, hover, result hints, navigation, outline, and folding.
- `tsc/internal/format`: native formatting context and indentation rules.
- `tsc/internal/project/checkerpool.go`: existing checker ownership and cancellation
  disposal, used by the new queries.

## Validation and limits

`tsc/internal/ls/typefn_test.go` exercises language-service APIs before semantic checking,
shadowing, iteration scopes, namespace completion, unfinished calls, hover/results,
UTF-16 positions, and token modifiers. `typefn_format_test.go` reparses formatted
source and checks idempotence across arithmetic, control flow, tuple/object
construction, calls, and native-island syntax.
Incomplete-buffer tests include semantic diagnostics and parser-limit recovery;
token-navigation tests visit every position in several recovered files.
`typefn_recovery_test.go` checks that native failures suppress successful hover
and inlay results even when the failing intermediate calculation is discarded.
`typefn_deferred_test.go` checks pending display, suppression of pending overlays,
native alias specialization, and stable Type counts across unchanged queries.

`tsc/internal/project/typefn_test.go` uses a real editor session and pool. It checks
cross-file navigation/references/rename, changed and restored imported results,
malformed imported bodies, profile-cache isolation, and deterministic cancellation
followed by a successful request, including cancellation during deferred replay.
`tsc/internal/lsp/server_typefn_test.go` also sends
serialized LSP requests to verify hover, definition, and result-hint preference
routing. Encoder tests exercise DSL wire round-trips.

See [testing](testing.md) for validation commands and results. Automated service,
protocol and extension-build checks do not constitute a graphical VS Code session.
The warm hover benchmark includes lookup, source mapping and Markdown creation;
it does not measure cold evaluation or whole-project latency:

```sh
go test ./tsc/internal/ls -run '^$' -bench BenchmarkTypefnHover -benchmem
```

The implemented slice does not include string algorithms, HKT execution, native-island
execution, or declaration-file recipes. Deferred execution now has a tested
[implementation boundary](deferred-evaluation.md). The integration does not
claim DSL-specific auto-import completion, extract/move refactorings, all code
fixes, project-reference/source-map coverage, or complete API factory compatibility.
The broader conformance suite and complete native-operation resource accounting
remain required before describing the language as mature.
See [implementation status](implementation.md) for the remaining specification gaps.
