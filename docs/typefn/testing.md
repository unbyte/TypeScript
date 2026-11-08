# Testing and validation

Validation through [`3175dc0e3`](https://github.com/unbyte/TypeScript/commit/3175dc0e3),
2026-11-08. Toolchain: Go 1.27.0 and Node 24.22.1 on Linux/amd64.
The implementation remains incomplete against the language specification; these
results establish the tested boundaries below.

* Passed: 219 isolated compiler-project fixtures. The harness checks complete
  diagnostics and requires successful native `Assert<...>` applications to resolve
  to literal true. Calculator, numeric and HTTP-route documentation sources execute
  directly; reflection/builders, native islands, symbolic replay, imported failures
  and definition-time return paths have executable coverage.
* Passed during this integration: full parser, printer, AST, Go API encoder, checker,
  DSL harness and language-service suites. Subsequent annotation hover changes passed
  the targeted language-service suite.
* Passed: 11 selected upstream compiler baselines for conditional, mapped, union,
  template literal and generic-signature Types, including native emit and type/symbol
  baselines. These were rerun after the result-kind AST change.
* Passed: AST, enum and sync API regeneration. Generated enum values agree with Go.
* Passed: API/test project type checks, eight DSL AST traversal/factory/protocol tests
  and 28 API encoder tests.
* Passed: async and generated sync RPC profile tests against the rebuilt compiler.
  Both read the new result-kind node from the Go-encoded AST, preserve computed
  native results and reject the disabled profile.
* Passed: rebuilt CLI against the editor example project; changed Go source formatting
  and `git diff --check`.

Earlier race, project-invalidation, extension-build and wider package checks are
recorded in the [historical validation record](https://github.com/unbyte/TypeScript/blob/7fe54fd975b95531092485468ac5677ed3a453fc/docs/typefn/testing.md).
They are not reported as rerun for this integration. Full conformance, declaration
recipes, broad community ports, resource accounting and performance acceptance
remain tracked in [acceptance](acceptance.md).

Reproducible native checks from the repository root (the first command is broader
than the latest seven-package run):

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
