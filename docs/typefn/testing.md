# Testing and validation

Latest verification recorded on 2026-11-09, including mixed constructor arguments, native
constructors, native island capture bridges and native lexical invocation environments. Toolchain: Go 1.27.0 and Node 24.22.1 on Linux/amd64;
the rebuilt fork reports `Version 7.1.0-dev`. The implementation remains incomplete
against the language specification; these checks establish the boundaries below.

* Passed: 470 compiler fixtures (full harness, including written-recipe producer/consumer cases). The harness checks complete
  diagnostics and requires successful native `Assert<...>` applications to resolve
  to literal true. Calculator, numeric and HTTP-route documentation sources execute
  directly. Recent cases cover tuple provenance, closures/factories, static kinds,
  creation-check readiness, computed bounds, mixed constructor arguments, native
  constructor identity, dependent bounds, lexical capture substitution, constructor
  aliases, forward chains, imports, creation cycles, native island constructors and
  recursive builders and groups, guarded validation, property/indexer/tuple/signature
  descriptors, overload forwarding, native predicate narrowing and scratch-cache isolation,
  native lexical replay with query-order isolation, ambient island alias defaults,
  `typeof` queries, mapped generic constraints and Type/value namespace separation.
  Authored declaration recipes test ABI headers, constructor captures and skipLibCheck;
  editor regressions check consumer diagnostic provenance in both query orders.
  Generated declarations additionally exercise frozen captures, private helper visibility,
  namespace scope, native identities and the calculator/route source modules.
  Direct stripped-dependency failures cover private/exported helpers, imports and
  distinct namespace members, with authored declaration dependencies as a control.
  Transitive cases cover constructor aliases, native aliases, module boundaries and
  cycles; an explicitly typed native runtime body is not treated as an emitted recipe.
  Ordinary native aliases and return annotations also validate reachable DSL recipes;
  native-only stripInternal behavior is unchanged. Reachability tests alternate native
  and DSL-bearing cycles to verify that negative cache entries cannot hide recipes.
  Authored declaration bridges are traversed, malformed native dependencies diagnose
  before recovery-tree inspection, and skipLibCheck cannot hide invalid recipes.
  Inferred variable/property/return/default-export references validate generated
  dependencies too; a native-only consumer verifies the non-DSL control.
  Inferred call recipes are consumed with and without interface augmentation,
  including generic substitutions, mapped binders, generic class fields, nested
  object properties and public/private cross-module helpers. Stripped and inaccessible local helpers suppress
  output, including forced/builder modes; unrelated literals and explicit native
  annotations retain their native behavior. Explicit computed arguments and chained
  defaults after a partial explicit argument list preserve their source syntax;
  identical producer results, generic class binders and an inferred-value/default
  collision are checked in fresh augmented and non-augmented consumers. Native
  inference-selected defaults preserve their provenance separately from shared
  signatures; completion/signature-help queries run before and after semantic
  checking, and nested/abandoned resolution publication is tested. Single-signature call
  projections retain native argument bounds, DSL-computed bounds and parameter-only
  calculations; augmented consumers exercise normal and skipLibCheck policies.
  Value forwarding covers local alias chains, imported annotated/inferred values,
  properties, single-return functions, arrows and getters. A same-name Type/value
  pair checks namespace resolution; stripped forwarded values suppress all emit modes.
  Stored/direct native specializations preserve computed arguments and defaults;
  generic class fields, imported private helpers and consumer-invalidated bounds
  are exercised alongside a fixed specialization with the same producer Type.
  Inferred native returns and constructed callable signatures now retain consumer
  recipes, with generic function/arrow/class cases and concrete declaration failure
  checks under normal and skipLibCheck policies.
  Deferred contextual literal regression tests cover generic input inference and
  forwarding. Checker tests verify that opaque terms supply no constraint, guarded
  shapes retain their checks, and cached constraints do not prevent specialization.
  Bigint input-budget cases cover native arguments to conversion/formatting,
  comparisons and unary operations, preflight rejection of long decimal operands,
  signed/zero boundary values and declaration consumers with skipLibCheck. The full
  harness, checker and language-service suites pass after the shared operand checks.
  Deterministic cancellation tests stop inside native comparisons, mapped-property
  substitution, distributive conditionals, union collection/sorting and intersection
  collection on 4,000-member inputs. Result-query and semantic-diagnostic paths
  suppress publication and invalidate the checker; fresh checkers produce the exact
  expected native Types. Full harness/checker/LS suites pass with these hooks. Other
  native-operation families and complete quota accounting remain pending.
  Repeated native-alias failures retain distinct outer application spans through
  direct calls, deferred aliases and nested islands, with a successful sibling
  control. Source imports and fresh declaration consumers (normal/skipLibCheck)
  pass, as do the full harness, checker and language-service suites.
* Passed: type-fest literal word-scanner workload: 464 native-Type comparisons in
  direct and declaration-only programs, fixed expected results, generic options,
  explicit unions, tuple metadata, domain failures and resource diagnostics. A
  512-character word passes a 4,096-unit string-work limit after replacing growing
  prefix concatenation with cursor slices. See [workload scope](workloads.md).
* Passed: React Hook Form field-path workload: 13 fixtures covering 25 native-Type
  comparisons, fixed outputs, recursive/nominal/opaque Types, depth truncation,
  generic inference and forwarding, fresh declaration consumers with augmentation,
  optional/indexed-access options and negative/resource cases. See [workload scope](workloads.md).
* Passed: full parser, binder, printer, Go API encoder, checker, DSL harness and
  language-service suites. Checker, harness and language-service suites were rerun
  after native receiver capture, recovery-Type rejection and declaration ABI checks.
  Parser/binder/printer/checker/LS suites also pass at written-recipe emission; optional-provenance assertions pass
  in both exact-optional modes, and the added generic-body dependency assertion passes.
  Full checker, harness and language-service suites rerun after inferred call recipes.
* Passed: 751 upstream compiler configurations selected by `declarationEmit*`, generic
  calls, contextual typing, type-parameter defaults and the type-system selection below,
  including native emit and type/symbol baselines. The runner skipped 23 configurations
  through its existing filters. Rerun after native union/intersection cancellation polling.
* Passed: AST, enum and sync API regeneration. Generated enum values agree with Go.
* Passed: API/test project type checks, ten DSL AST traversal/factory/protocol tests
  and 28 API encoder tests.
* Passed: async and generated sync RPC profile tests against the compiler rebuilt
  at `2b1368aec`.
  Both read parameter, bound, lambda and result-kind nodes from the Go-encoded AST,
  preserve computed native results and reject the disabled profile.
* Passed: CLI rebuilt at `2b1368aec` against the editor example project and the constraint probe
  script (13 fork cases and the native baseline). Unused bounds no longer execute,
  and later-parameter references diagnose. This does not add `where` syntax.
* Passed: decoded declaration mappings for DSL statement starts after the ABI header;
  full printer and integration suites rerun after enabling per-node mapping.
* Passed: ordinary, forced and builder-signature emission for invalid definitions,
  malformed DSL bodies and transitive stripped dependencies. Forced/builder cases
  explicitly enable noEmitOnError; the DSL failure boundary still withholds affected
  output, while independent native output is preserved. Full integration harness rerun.
* Passed: desktop and web extension bundles rebuilt at `2b1368aec`. This is a build
  check, not a graphical Extension Host session.
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
go test ./tsc/internal/testrunner -run '^TestLocal/(declarationEmit.*|generic.*|contextual.*|typeParameterDefault.*|conditionalTypes[12]|mappedTypes[12]|unionTypeReduction2?|templateLiteralTypes[12])\.ts' -count=1
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

## Declaration-only harness

An integration fixture may include a `consumer` fixture. The producer emits actual
declaration bytes; only `.d.ts`, `.d.mts` and `.d.cts` outputs and the consumer's
explicit inputs enter a fresh filesystem, host, program and checker. Producer
sources and source redirects are unavailable. Consumer files cannot overwrite
emitted declarations. Native Type assertions and complete diagnostics use the
same checks as ordinary fixtures. Explicit emit expectations also run after
semantic errors, so emit-on-error and atomic suppression can be tested.

Three harness regressions verify generic/nominal declaration consumption, a failed
import of the absent producer source, and native declaration consumption following
an allowed emit-on-error. They establish the harness; they do not yet demonstrate
DSL declaration recipes; generated-recipe fixtures now use the same isolated consumer path.


Community word-scanner benchmark (same host/toolchain, three iterations):

```sh
go test ./tsc/internal/typefntest -run '^TestIntegration/(community-words|error-community-words)' -bench '^BenchmarkTypefnCommunityWords$' -benchtime=3x -count=1
```

250.0 ms/op, 120,239,165 bytes/op, 835,707 allocations/op. This includes cold compiler
projects, the DSL scanner, its native comparison and all corpus assertions. It is
neither isolated VM throughput nor a native-versus-DSL speed comparison.

Community field-path benchmark (same host/toolchain, three iterations):

```sh
go test ./tsc/internal/typefntest -run '^TestIntegration/(community-paths|error-community-paths)' -bench '^BenchmarkTypefnCommunityPaths$' -benchtime=3x -count=1
```

168.9 ms/op, 48,379,653 bytes/op, 276,568 allocations/op. This includes cold compiler
setup, both path algorithms, assertions and native generic call checking. It is not
an isolated VM or comparative speed measurement.

Native-only polling comparison (Go 1.27.0, Linux/amd64, EPYC 9V74, GOMAXPROCS=4):

| Revision | Median time/op | Median bytes/op | Median allocations/op |
| --- | ---: | ---: | ---: |
| `c45dcc067`, before native polling | 1.445 s | 114,342,888 | 933,908 |
| `a75816d92`, with native polling | 1.434 s | 114,415,512 | 933,914.5 |

Eight alternating pairs, three iterations per sample, use the upstream
`BenchmarkCheck` compiler fixture with profile 0. Each iteration creates a fresh
checker over an already bound program. The median paired time change is -0.87%;
the descriptive paired bootstrap interval is -4.90% to +1.77%. These samples show
no clear regression or improvement on this fixture. They do not establish zero
overhead, enabled-profile performance, or performance across native projects.
[Raw samples and binary/source hashes](probes/native-polling-benchmark.json) are retained.

To repeat, create separate checkouts at the two revisions and compile each test
binary with `go test -c -o /absolute/path/checker.test ./tsc/internal/checker`.
Keep both checkouts available: the binaries read their compiler-fixture sources.
Then run the [paired runner](probes/compare-native-checker.py):

```sh
python3 docs/typefn/probes/compare-native-checker.py /absolute/path/baseline.test /absolute/path/candidate.test > comparison.json
```

The runner executes sequentially, alternates ordering, sets GOMAXPROCS for both
binaries and records raw metrics and hashes. Build before measuring and avoid
concurrent compiler workloads during the comparison.
