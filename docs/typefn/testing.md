# Testing and validation

Latest verification recorded on 2026-11-09, including mixed constructor arguments, native
constructors, native island capture bridges and native lexical invocation environments. Toolchain: Go 1.27.0 and Node 24.22.1 on Linux/amd64;
the rebuilt fork reports `Version 7.1.0-dev`. The implementation remains incomplete
against the language specification; these checks establish the boundaries below.

* Passed: 545 compiler fixtures (full harness, including written-recipe producer/consumer cases). The harness checks complete
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
  Mapped-property failures are also replayed on each read, including ordinary
  native indexed access and fresh declaration consumers. Query-before-diagnostics
  tests in both application orders reject result publication and retain both
  source spans; conditional aliases supply a separate passing replay control.
  Native relation-cache regressions retain failures on repeated ordinary TypeScript
  assignments, whether the remaining members are compatible or incompatible.
  Source and declaration consumers preserve use spans and still report an unrelated
  native error; successful sibling comparisons remain valid. Full harness/checker/LS
  suites pass after the relation publication checks.
  Instantiated-member regressions cover repeated generic indexed reads and
  setter writes, with successful siblings and fresh declaration consumers in
  both skipLibCheck modes. Full integration, checker and language-service suites
  pass for this scope. Native queries before semantic checking now preserve the
  outer application span: tests query indexed-access and generic-alias Types by
  expression, name, declaration and symbol, in both use orders, then check exact
  diagnostic positions. Full harness/checker/LS suites pass with the scoped query helper.
  Conditional-container cases retain failures through tuple branches, distribution
  and unknown absorption. Direct tuple aliases, valid specializations and unselected
  branches are controls. Source/declaration consumers and native query-order tests
  pass, together with the full harness/checker/LS suites, after guarding substitution,
  conditional-root and generic-alias cache publication.
  Combined-return cache tests cover repeated calls, optional calls, constructor
  unions and overload selection, with successful siblings and fresh declaration
  consumers in both skipLibCheck modes. Full integration, checker and language-service
  suites pass for this scope. Repeated initializer/name queries now deduplicate
  failures after attaching operation provenance. Tests retain distinct messages
  and distinct originating operations even when their error text is identical;
  operation spans remain singular. Full AST/harness/checker/LS suites pass.
  Mapped-key eligibility uses known guarded shapes without discharging deferred
  checks. Source and declaration consumers cover string, numeric and unique-symbol
  keys, optional/readonly modifiers and both exact-optional modes. Opaque results
  require a native key-domain proof; non-key shapes still diagnose. Repeated failing
  specializations retain their application spans, including with skipLibCheck.
  Full harness/checker/LS suites pass for this scope.
  Shared mapped-member tables are retried after a failed key calculation, including
  after some properties have already been built. Repeated reflection, a successful
  sibling, fresh declaration consumers and application/native API queries in both
  orders pass. Full harness/checker/LS suites pass after the publication guard.
  Namespace-qualified recipes cover imports, nested/private local namespaces,
  value forwarding, stored/direct specializations and consumer augmentation with
  skipLibCheck both off and on. Computed native bounds retain their declaration
  checking policy. Stripped qualified helpers suppress ordinary, forced and builder
  output. Full harness/checker/LS suites pass for this scope.
  Declared computed bounds support exact-bound forwarding, chained parameters and
  native assignments. Source/declaration consumers retain invalid-argument errors
  and repeated bound-execution failures, including with skipLibCheck. Checker tests
  distinguish assignability from identity/reverse assignability, reject unrelated
  terms and withhold proofs during failed evaluation. Full harness/checker/LS pass.
  Local written annotations supply inferred return recipes for variables, required
  parameters, methods/getters and explicit undefined unions. Plain/augmented
  declaration consumers pass in both skipLibCheck modes. A closed local helper
  whose current result is literal still suppresses ordinary/forced/builder output
  when its annotation cannot be named. Full harness/checker/LS suites pass.
  Optional parameter recipes preserve implicit undefined while defaulted parameters
  remain defined. Strict/non-strict producers and consumers, including a non-strict
  producer with a strict augmented consumer, pass. Full harness/checker/LS pass.
  A closed computed count narrowed by runtime flow no longer emits its producer
  literal: ordinary, forced and builder output diagnose and withhold the file.
  Full harness/checker/LS suites pass after this declaration failure-boundary fix.
* Passed: generated calculator conformance (S20). A deterministic corpus contains
  144 expression trees, each rendered with minimal parentheses and again with
  redundant groups and ASCII whitespace. An independent Go `math/big` tree evaluator
  supplies exact results or the first division failure; rendering records the
  expected UTF-16 operator offset. Tests cover signed operands/divisors, leading
  zeroes, large integers, precedence and associativity, plus `Calc` conversion for
  safe integers. The 288 expressions run from source and from freshly emitted
  declarations with skipLibCheck both off and on. Generated success assertions must
  resolve to literal true; failure cases require exact counts and application spans.
  This supplements the fixed malformed-input and Unicode string tests; it is not
  an exhaustive parser proof or a performance measurement.
* Passed: type-fest literal word-scanner workload: 464 native-Type comparisons in
  direct and declaration-only programs, fixed expected results, generic options,
  explicit unions, tuple metadata, domain failures and resource diagnostics. A
  512-character word passes a 4,096-unit string-work limit after replacing growing
  prefix concatenation with cursor slices. See [workload scope](workloads.md).
* Passed: React Hook Form field-path workload: 13 fixtures covering 25 native-Type
  comparisons, fixed outputs, recursive/nominal/opaque Types, depth truncation,
  generic inference and forwarding, fresh declaration consumers with augmentation,
  optional/indexed-access options and negative/resource cases. See [workload scope](workloads.md).
* Passed: React Hook Form field-array paths: 13 additional fixtures, a 35-model
  native comparison corpus, fixed outputs, independent depth checks, opaque nominal
  leaves, generic inference and consumer augmentation. Source/declaration programs
  cover optional/indexed-access settings and invalid input/path/depth/resource cases.
  The full integration suite passes with both traversal workloads.
* Passed: React Hook Form path-value lookup: 13 fixtures and a 37-model native
  comparison corpus, including optional reads, symbolic string paths, any/never,
  tuple/array indices, class identity, computed input bounds and generic forwarding.
  Source/declaration programs cover compiler-option variants, consumer augmentation,
  invalid inputs/paths and lookup recursion exhaustion through any. Native `const`
  path parameters preserve literal call inference. The full integration suite passes.
* Passed: full parser, binder, printer, Go API encoder, checker, DSL harness and
  language-service suites. Checker, harness and language-service suites were rerun
  after native receiver capture, recovery-Type rejection and declaration ABI checks.
  Parser/binder/printer/checker/LS suites also pass at written-recipe emission; optional-provenance assertions pass
  in both exact-optional modes, and the added generic-body dependency assertion passes.
  Full checker, harness and language-service suites rerun after inferred call recipes.
* Passed: the full native compiler baseline suite (`TestLocal`): 13,811 configurations,
  including diagnostics, emit, type/symbol and source-map checks. The runner skipped
  1,721 configurations through its existing filters. This full run follows optional
  parameter recipe support. Earlier focused selections passed 751 configurations
  with 23 skipped; their narrower command remains useful for individual changes.
* Passed: AST, enum and sync API regeneration. Generated enum values agree with Go.
* Passed: API/test project type checks, ten DSL AST traversal/factory/protocol tests
  and 28 API encoder tests. The API/test project type check, AST tests, unsigned
  flag test and encoder tests were rerun at `e1fad7459`.
* Passed: async and generated sync RPC profile tests against the compiler rebuilt
  at `e1fad7459`.
  Both read parameter, bound, lambda and result-kind nodes from the Go-encoded AST,
  preserve computed native results and reject the disabled profile.
* Passed: CLI rebuilt at `e1fad7459` against the editor example project and the constraint probe
  script (13 fork cases and the native baseline). Unused bounds no longer execute,
  and later-parameter references diagnose. This does not add `where` syntax.
* Passed: decoded declaration mappings for DSL statement starts after the ABI header;
  full printer and integration suites rerun after enabling per-node mapping.
* Passed: ordinary, forced and builder-signature emission for invalid definitions,
  malformed DSL bodies and transitive stripped dependencies. Forced/builder cases
  explicitly enable noEmitOnError; the DSL failure boundary still withholds affected
  output, while independent native output is preserved. Full integration harness rerun.
* Passed: desktop and web extension bundles rebuilt at `e1fad7459`. This is a build
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
go test ./tsc/internal/testrunner -run '^TestLocal$' -count=1 -timeout=20m
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

Community field-array-path benchmark (same host/toolchain, three iterations):

```sh
go test ./tsc/internal/typefntest -run '^$' -bench '^BenchmarkTypefnCommunityArrayPaths$' -benchtime=3x -count=1
```

191.9 ms/op, 53,189,368 bytes/op, 320,689 allocations/op. The verified cold project
includes both the imperative and native traversals and their assertions. This is
not isolated VM throughput or a speedup comparison.

Community path-value benchmark (same host/toolchain, three iterations):

```sh
go test ./tsc/internal/typefntest -run '^$' -bench '^BenchmarkTypefnCommunityPathValue$' -benchtime=3x -count=1
```

230.5 ms/op, 56,412,354 bytes/op, 349,597 allocations/op. This includes native path
bounds, both lookup algorithms, assertions, generic calls and cold compiler setup;
it is not isolated VM throughput or a speedup comparison.

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

Native-only failure-replay comparison, using the same protocol:

| Revision | Median time/op | Median bytes/op | Median allocations/op |
| --- | ---: | ---: | ---: |
| `a75816d92`, before failure-replay fixes | 1.380 s | 114,407,597 | 933,913.5 |
| `0f61c1dcb`, with failure replay and query sites | 1.397 s | 114,410,095.5 | 933,914 |

The median paired change is +0.90%, with a descriptive bootstrap interval of
+0.26% to +3.03%. This indicates a small regression on this native fixture;
it is not evidence of zero overhead. [Raw samples](probes/native-replay-benchmark.json)
retain the separate binary hashes and source revisions. Native query-scope cleanup
at `0f61c1dcb` added a defer to the general type-node resolver. The subsequent
refactor confines that cleanup to a helper used by profile-1 queries; the earlier
measurement remains attached to its source revision.

Two further eight-pair comparisons measured the helper refactor at `5a2c07abd`
with the same fixture and protocol:

| Baseline → helper refactor | Median paired time change | Descriptive bootstrap interval |
| --- | ---: | ---: |
| `0f61c1dcb` → `5a2c07abd` | −3.07% | −5.26% to +2.50% |
| `a75816d92` → `5a2c07abd` | −1.62% | −4.50% to +3.51% |

These samples do not distinguish a clear improvement or regression. The first
comparison isolates the refactor; the second measures all intervening replay
fixes against the earlier baseline. Both [refactor samples](probes/query-scope-fast-benchmark.json)
and [combined samples](probes/native-replay-fast-benchmark.json) are retained,
including the candidate's pre-commit build provenance. Full harness/checker/LS
and the 751 selected native configurations pass for this refactor. Enabled-profile
and broader project performance remain unverified by these measurements.

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
