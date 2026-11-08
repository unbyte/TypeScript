# Testing and validation

Validation date: 2026-11-08. Toolchain: Go 1.27.0 and Node 24.22.1 on Linux/amd64.
These checks were run on the implementation preserved at
[`7fe54fd97`](https://github.com/unbyte/TypeScript/commit/7fe54fd975b95531092485468ac5677ed3a453fc).
The documentation-only history rewrite preserves that implementation exactly.
The implementation remains experimental; a green integration suite is not full
language-spec conformance or an exhaustive native compiler regression run.

* Passed: 65 isolated compiler-project fixtures, including iterator-frame accounting
  and native declaration output with the profile both off and on. The harness checks
  complete diagnostics, exact emit, and literal-true assertions, with native positive/negative calibration.
* Passed: full suites for the 17 packages listed below, including native Type layout,
  casts, composite cache allocations, project invalidation, cancellation, and mapped
  projections.
* Passed: 11 selected upstream compiler baselines for conditional, mapped, union,
  template literal, and generic-signature types (including their ordinary emit,
  type/symbol, union-ordering and source-parent checks).
* Passed: forced AST, compiler-option/API/enum, preference and sync API regeneration;
  no tracked-file drift.
* Passed: 27 option-generator tests, script type checks, API/test project type checks,
  5 DSL AST traversal/factory tests, the TypeFlags test, and 28 API encoder tests.
* Passed: async and generated sync RPC profile tests against the built compiler,
  including the Go-encoded DSL subtree and disabled-profile rejection.
* Passed: targeted race suites for the harness, parser, language service, project
  sessions and LSP server, rerun after the declaration-scan guard.
* Passed: desktop/web extension build, rebuilt CLI editor example, 13 DSL constraint
  observation cases plus their native baseline, and the JavaScript compiler baseline
  probe. These probes preserve existing observations; they do not implement `where`.
* Passed: Go 1.27 formatting for every changed Go file and `git diff --check`.

Reproducible native checks from the repository root:

```sh
go test ./tsc/internal/typefntest ./tsc/internal/tsoptions ./tsc/internal/parser ./tsc/internal/printer ./tsc/internal/ast ./tsc/internal/astnav ./tsc/internal/api/encoder ./tsc/internal/binder ./tsc/internal/checker ./tsc/internal/compiler ./tsc/internal/ls ./tsc/internal/format ./tsc/internal/ls/lsutil ./tsc/internal/project ./tsc/internal/lsp ./tsc/internal/core ./tools/gen-proto -count=1
go test -race ./tsc/internal/typefntest ./tsc/internal/parser ./tsc/internal/ls ./tsc/internal/project ./tsc/internal/lsp -run 'TestIntegration|Typefn' -count=1
go test ./tsc/internal/testrunner -run '^TestLocal/(conditionalTypes[12]|mappedTypes[12]|unionTypeReduction2?|templateLiteralTypes[12]|genericCallWithGenericSignatureArguments[23]?)\.ts' -count=1
go build -o built/local/tsc ./tsc/cmd/tsc
built/local/tsc --project docs/typefn/examples/editor/tsconfig.json
python3 docs/typefn/probes/check-where-native.py built/local/tsc
node docs/typefn/probes/check-where-baseline.mjs
```

Generated/API/editor checks:

```sh
node_modules/.bin/hereby generate:ast --force
node_modules/.bin/hereby generate:compileroptions --force
node_modules/.bin/hereby generate:preferences --force
node_modules/.bin/hereby generate:sync --force
node tools/scripts/tsc/options.test.ts
node_modules/.bin/hereby check:scripts
node_modules/.bin/tsc -b packages/typescript-wasip1-wasm
node_modules/.bin/tsc -b packages/typescript/test --force
node --conditions @typescript/source packages/typescript/test/typefn_ast.test.ts
node --conditions @typescript/source packages/typescript/test/typefn_flags.test.ts
node --conditions @typescript/source packages/typescript/test/encoder.test.ts
node --conditions @typescript/source --test-name-pattern='typefn profile' packages/typescript/test/async/api.test.ts
node --conditions @typescript/source --test-name-pattern='typefn profile' packages/typescript/test/sync/api.test.ts
npm run extension:build
```

## Performance scope

`BenchmarkTypefnProject` measures cold compiler projects, including configuration,
bundled libraries and native checking. `BenchmarkTypefnHover` measures a warm editor
request. Neither isolates VM throughput, and no speedup or large-project performance
claim follows from the small historical samples.

```sh
go test ./tsc/internal/typefntest -run '^$' -bench BenchmarkTypefnProject -benchtime=3x -count=3
go test ./tsc/internal/ls -run '^$' -bench BenchmarkTypefnHover -benchmem
```

The full upstream corpus, target conformance matrix, large community workload ports
and a graphical editor session are not covered by the reported checks. Full native
work, retained-cache and memory accounting also remain incomplete.
