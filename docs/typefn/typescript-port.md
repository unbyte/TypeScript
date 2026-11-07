# Integration into the current TypeScript repository

## Baselines and scope

This port starts at `unbyte/TypeScript` main
`f7973164a5f079a2d8b88d0fdb39347222549124`. The implementation being ported is
`typefn-redesign` at `62d8a815` in the former `microsoft/typescript-go` checkout,
whose upstream baseline was `89d5d5b2849a0db0957065889ca58536fa6d2e4a`.
The original branch and its individual development commits are preserved there.
The new branch starts from the current Go-based TypeScript monorepo; it does not
replace main with the old repository or assume the histories are interchangeable.

This is an architectural port of the implemented slice, not a claim that every
feature in the target language specification is implemented. The implementation
status and deferred-evaluation limitations still apply. In particular, this work
does not add `where`, HKT, the full string library, or declaration recipes.

## Integration decisions

### Native syntax, binding, and evaluation

Keep dedicated native AST nodes and the bounded DSL parser. Normal TypeScript
parser recovery, native lexical scopes, and the checker own those nodes. Keep
bytecode execution inside the checker with native Type handles; do not lower the
DSL to JavaScript or to generated conditional-type source.

Keep deferred calculations in native type instantiation and relation checking.
A VM call depending on an unknown input must not run speculatively using its
constraint. Reached checks must survive simplifications and replay when native
substitution supplies arguments. Existing immutable term interning remains
checker-local; completed VM results are not globally memoized. New upstream
checker accessors and temporary-type-set pools must remain intact.

### Options are generated from one model

Register the opt-in profile in `tools/scripts/tsc/options.ts`, with its structured
shape in the option generator model. Generate Go options, public TypeScript API
options, JSON schemas, cloning/equality, serialization, and incremental option
comparisons through the repository's existing generators. Resource limit names
are shared between generated Go structs, API interfaces, and schemas.

Use the native configuration validation hook for profile diagnostics. Do not
restore the old custom `ParseCompilerOptions` signature or copy old generated
option declarations. Value equality for profiles is structural, so separately
parsed equivalent configurations do not invalidate caches by pointer identity.
Changes affect semantic diagnostics, emit, and build information. Parse-affecting
profile fields remain in `SourceFileParseOptions`, the native source-file cache
key. Tests must cover switching the profile off and back on and changing limits.

### Editor preferences and mapped documents

Add result-hint preferences to `tools/userPreferences.schema.json`; generate both
Go and API preference types. Keep the extension's user setting connected to that
preference. The new language server can project a document into multiple scripts:
DSL hover and signature help must participate inside the projection loop and use
native location conversion instead of assuming one unmapped script per document.

### Public AST API and wire protocol

Regenerate AST, factories, predicates, encoder/decoder, and enum exports from the
new metadata sources. Use `@typescript/typescript` public entry points rather than
the former `@typescript/native-preview/unstable` paths. The new API has iterator
child traversal as well as callbacks and transforming visitors. All three must
visit DSL operators in source order while retaining the stable wire field order.

### Repository layout

| Former path | Current path |
| --- | --- |
| `internal/` | `tsc/internal/` |
| `_packages/native-preview/` | `packages/typescript/` |
| `_extension/` | `packages/vscode-typescript/` |
| `_scripts/` | `tools/scripts/tsc/` |
| `cmd/tsgo` | `tsc/cmd/tsc` |
| `_submodules/TypeScript/tests/` | `tsc/testdata/` |

The current workspace uses Go 1.27 and the repository's internal JSON package.
Old audit records describe validation of the old checkout; they are not evidence
that this port passes. Current-port validation will be recorded below separately.

## Validation gates

Before publication: regenerate from sources; compile the native compiler and API;
run the compiler-project DSL fixtures, parser/recovery and codec tests; exercise
native inference and repeated failed deferred calls; exercise editor sessions,
profile invalidation, and serialized LSP requests; test public AST visitors and
flags; build the editor extension; run representative upstream compiler tests.
Add focused coverage for changed integration boundaries rather than merely
accepting changed error baselines. Record commands, results, and limitations here.

Status: port in progress; current-repository validation is not complete.
