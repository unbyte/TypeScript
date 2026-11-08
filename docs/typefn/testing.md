# Testing and validation

Latest verification recorded on 2026-11-08, including mixed constructor arguments, native
constructors, native island capture bridges and native lexical invocation environments. Toolchain: Go 1.27.0 and Node 24.22.1 on Linux/amd64;
the rebuilt fork reports `Version 7.1.0-dev`. The implementation remains incomplete
against the language specification; these checks establish the boundaries below.

* Passed: 341 isolated compiler-project fixtures (full harness plus added mutual-group and environment-projection cases). The harness checks complete
  diagnostics and requires successful native `Assert<...>` applications to resolve
  to literal true. Calculator, numeric and HTTP-route documentation sources execute
  directly. Recent cases cover tuple provenance, closures/factories, static kinds,
  creation-check readiness, computed bounds, mixed constructor arguments, native
  constructor identity, dependent bounds, lexical capture substitution, constructor
  aliases, forward chains, imports, creation cycles, native island constructors and
  recursive builders and groups, guard rejection and scratch-cache isolation,
  native lexical replay with query-order isolation, ambient island alias defaults,
  `typeof` queries, mapped generic constraints and Type/value namespace separation.
* Passed: full parser, binder, printer, Go API encoder, checker, DSL harness and
  language-service suites. Checker, harness and language-service suites were rerun
  after recursive environment compaction; optional-provenance assertions pass
  in both exact-optional modes, and the added generic-body dependency assertion passes.
* Passed: 11 selected upstream compiler baselines for conditional, mapped, union,
  template literal and generic-signature Types, including native emit and type/symbol
  baselines, rerun after native constructor aliases.
* Passed: AST, enum and sync API regeneration. Generated enum values agree with Go.
* Passed: API/test project type checks, ten DSL AST traversal/factory/protocol tests
  and 28 API encoder tests.
* Passed: async and generated sync RPC profile tests against the rebuilt compiler.
  Both read parameter, bound, lambda and result-kind nodes from the Go-encoded AST,
  preserve computed native results and reject the disabled profile.
* Passed: rebuilt CLI against the editor example project and the constraint probe
  script (13 fork cases and the native baseline). Unused bounds no longer execute,
  and later-parameter references diagnose. This does not add `where` syntax.
* Passed: changed Go source formatting and `git diff --check`.

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

Recursive construction scaling was measured with:

```sh
PATH=/workspace/toolchains/go/bin:$PATH GOCACHE=/workspace/.cache/go-build GOPATH=/workspace/.cache/go go test ./tsc/internal/typefntest -run '^TestIntegration/recursive-(independent|mutual)-roots$' -bench '^BenchmarkTypefnProject/recursive-' -benchtime=3x -count=1
```

Linux/amd64, AMD EPYC 9V74, Go 1.27.0, three iterations per case:

| Project | Time/op | Bytes/op | Allocs/op |
| --- | ---: | ---: | ---: |
| 256 independent roots | 161.5 ms | 62,112,813 | 305,852 |
| 256 mutually recursive roots | 178.2 ms | 62,091,901 | 302,855 |

These cold-project figures include parsing, libraries and checking; they are not
isolated VM throughput. Both fixtures also pass their explicit 100,000 native-node
budget. The independent case failed that budget before environment compaction.
