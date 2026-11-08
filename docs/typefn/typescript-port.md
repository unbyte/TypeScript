# Integration into the current TypeScript repository

This note records the original port. The subsequent [`feat/typefn` review series](review-series.md)
rebuilds that implementation as focused commits and records the current validation.

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
checker accessors and temporary-type-set pools remain intact. Deferred terms use
the native concrete type tag and checked casts, with a compile-time zero-offset
assertion and the existing cast/GC test matrix. They add no field to the native
Type header. Dynamically built object properties use the native compact SymbolMap
rather than reintroducing the old Go map representation.

### Options are generated from one model

Register the opt-in profile in `tools/scripts/tsc/options.ts`, with its structured
shape in the option generator model. Generate Go options, public TypeScript API
options, JSON schemas, cloning/equality, serialization, and incremental option
comparisons through the repository's existing generators. Resource limit names
are shared between generated Go structs, API interfaces, and schemas.

Use the native configuration validation hook for profile diagnostics, with source
locations in tsconfig. The JSON API validates the same shape and nested limits.
A null profile follows the native convention of an unset option; a present object
must select version 1. Do not
restore the old custom `ParseCompilerOptions` signature or copy old generated
option declarations. Value equality for profiles is structural, so separately
parsed equivalent configurations do not invalidate caches by pointer identity.
Changes affect semantic diagnostics, emit, and build information. Parse-affecting
profile fields remain in `SourceFileParseOptions`, the native source-file cache
key. Tests cover switching the profile off and back on, changing syntax limits
through watched-file events, and independently created API programs using
different profiles for the same source.

### Editor preferences and mapped documents

Add result-hint preferences to `tools/userPreferences.schema.json`; generate both
Go and API preference types. Keep the extension's user setting connected to that
preference. The new language server can project a document into multiple scripts:
DSL hover and signature help must participate inside the projection loop and use
native location conversion instead of assuming one unmapped script per document.
Tests exercise both supplemental-only source and duplicate projections, checking
computed hover results, original hover coordinates, and DSL signature help.

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

## Current-port validation

Validated on Linux/amd64 with Go 1.27.0 and Node 24.22.1. The executable reports
`Version 7.1.0-dev`. The native implementation is committed as `7c3eed203`;
configuration/generator integration is `c2becbe6b` and the preserved design record
is `a446d89fe`.

| Gate | Result |
| --- | --- |
| Full native package suites listed below | Passed, including all 62 isolated DSL compiler fixtures |
| Targeted race suites | Passed for fixtures, parser, language service, project/session, and serialized LSP |
| Native type representation | Cast/GC matrix and Type header size tests passed |
| Native composite caches | Zero-allocation cache-hit and temporary-set reuse tests passed |
| Profile integration | Config and raw API validation, structural equality, build-info inclusion, watched-config profile/limit changes passed |
| Mapped documents | Supplemental-only and duplicate projections: computed hover, original coordinates, signature help passed |
| Public API | Type-check passed; synchronous and asynchronous RPC programs computed `Twice<21>` as `42` and isolated enabled/disabled profiles |
| AST/wire protocol | DSL visitors, iterators and early termination, remote round trips, flag checks, and native encoder suite passed |
| Generator checks | 27 option/schema tests, Go protocol generator tests, script type-check passed |
| Reproducibility | Forced AST, options/API, preferences, and sync generation produced no source changes |
| Native baseline coverage | Selected conditional, mapped, union, template-literal, and generic-inference compiler baselines passed |
| User tools | Native executable and desktop/web extension built; editor example compiled |
| Constraint investigation | Native baseline and all 13 DSL probes reproduced the documented observations |

Native package suites, from the repository root:

```sh
go test ./tsc/internal/typefntest ./tsc/internal/tsoptions \
  ./tsc/internal/parser ./tsc/internal/printer ./tsc/internal/ast \
  ./tsc/internal/astnav ./tsc/internal/api/encoder ./tsc/internal/checker \
  ./tsc/internal/compiler ./tsc/internal/ls ./tsc/internal/format \
  ./tsc/internal/ls/lsutil ./tsc/internal/project ./tsc/internal/lsp

go test -race ./tsc/internal/typefntest ./tsc/internal/parser \
  ./tsc/internal/ls ./tsc/internal/project ./tsc/internal/lsp \
  -run 'TestIntegration|Typefn' -count=1

go test ./tsc/internal/testrunner \
  -run '^TestLocal/(conditionalTypes[12]|mappedTypes[12]|unionTypeReduction2?|templateLiteralTypes[12]|genericCallWithGenericSignatureArguments[23]?)\.ts' -count=1

go test ./tools/gen-proto
node tools/scripts/tsc/options.test.ts
node_modules/.bin/hereby check:scripts
node_modules/.bin/tsc -b packages/typescript-wasip1-wasm
node_modules/.bin/tsc -b packages/typescript/test --force

node --conditions @typescript/source --test \
  packages/typescript/test/typefn_ast.test.ts \
  packages/typescript/test/typefn_flags.test.ts \
  packages/typescript/test/encoder.test.ts

go build -o built/local/tsc ./tsc/cmd/tsc
node --conditions @typescript/source --test-name-pattern='typefn profile' \
  packages/typescript/test/async/api.test.ts
node --conditions @typescript/source --test-name-pattern='typefn profile' \
  packages/typescript/test/sync/api.test.ts
built/local/tsc --project docs/typefn/examples/editor
python docs/typefn/probes/check-where-native.py built/local/tsc
npm run extension:build
```

Regeneration was checked with `--force` on `generate:ast`,
`generate:compileroptions` (which also generates enums and the API protocol),
`generate:preferences`, and `generate:sync`. The checker stringer was regenerated
as well. Generated outputs remain reproducible from their native sources of truth.

The port exposed and corrected missing remote iterator ordering, protocol type
imports, raw-API object validation, bare-checker profile lookup, and obsolete
host/path/client contracts in the test harnesses. The config-version fixture now
requires the diagnostic at the invalid profile object in `tsconfig.json`.

This is validation of the implemented experimental slice, not certification of
the entire target specification or the full upstream compiler baseline corpus.
The allocation checks protect existing native cache behavior; they do not establish
end-to-end throughput parity with upstream. Known language and deferred-constraint
limitations remain recorded in the implementation status and constraint investigation.
