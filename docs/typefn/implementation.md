# Implementation status

The `feat/typefn` branch has an **experimental implementation** of
language version 1. The R2 specification remains the design baseline; this is not
a claim of R2 conformance or production maturity. Native editor integration is
implemented for this slice; see [editor integration](editor-integration.md).
The [implementation review series](review-series.md) records the current upstream
baseline, focused commits, ownership/performance changes and validation. The older
`typefn-redesign` branch remains available for comparison.
The [design/implementation audit](design-implementation-audit.md) records verified
defects, corrections and the remaining specification gaps.
The subsequent [deferred-evaluation implementation](deferred-evaluation.md)
adds native substitution/replay and retained validation checks at tested boundaries.
The [where-clause investigation](where-constraints-investigation.md) records
constraint timing and scope discrepancies found afterward; `where` remains a
design candidate, not implemented syntax.
The examples in the manual describe the target language. Use the integration
fixtures to see what the compiler currently executes.

## Try the implemented slice

Build the fork with `go build -o /tmp/tsc-typefn ./tsc/cmd/tsc`, then run
`/tmp/tsc-typefn --project path/to/tsconfig.json`. Enable the DSL in `tsconfig.json`:

```json
{
  "compilerOptions": {
    "strict": true,
    "typefn": { "version": 1 }
  }
}
```

```ts
export typefn Reverse<T> {
  type result = [];
  for (type item of T) { result = [item, ...result]; }
  return result;
}

typefn Range<N> {
  type result = [];
  type i = 0;
  while (i < N) {
    result = [...result, i];
    i = i + 1;
  }
  return result;
}

type Result = Reverse<Range<4>>; // [3, 2, 1, 0]
```

Native Type assertions, assignments, type-only imports and re-exports consume the
results through the actual checker. JavaScript emit erases the declarations.
Declaration emit currently fails explicitly: source-recipe serialization must be
implemented before executable definitions can be published safely in `.d.ts` files.

## Implemented architecture

* The opt-in profile and parse-affecting limits participate in source parse options.
  Version and limit settings are validated by the native tsconfig parser.
* Dedicated generated AST nodes represent DSL syntax. A separate recursive-descent
  statement parser and Pratt expression parser handle bodies. Angle-list recognition
  uses bounded, memoized speculation. No body falls back to JavaScript parsing.
* Native binding registers declarations, Type parameters, mutable locals and iteration
  bindings in native lexical scopes. Editor queries and VM lowering share symbol
  identities. A checker-owned compiler resolves DSL slots and names, validates all branches,
  and lowers control flow into store, return, branch, jump and iterator instructions.
  Expression trees have resolved slots/symbols and are evaluated by the same VM.
* Registers hold immutable native checker Types. Assignment changes a slot, and
  constructed Types retain their previous contents. Named calls share root budgets.
  Native generic applications use native instantiation and constraint checking.
* Compiled programs belong to one checker instance. No cross-checker Type handles,
  fabricated deferred type parameters, or global VM caches are used. Native checker
  node caches still apply. Pending terms are interned per checker; completed VM
  calls have no general result memoization layer yet.
  A lazily allocated `typefnState` owns these tables and the active root VM.
  Iterator slots are assigned while lowering and stored in each call frame; their
  storage participates in the live-frame estimate. Native-only operations avoid
  allocating DSL state, and disabled-profile checking skips DSL declaration scans.
* Dedicated native deferred Types retain immutable operands and captured locals.
  Native TypeMapper substitution replays calculations when arguments change.
  Per-call journals retain reached checks even when a result is discarded or
  absorbed by the supported native simplifications. See the
  [implemented boundary](deferred-evaluation.md) for details and limitations.
* The source printer preserves the implemented syntax. Round-trip tests include
  zero-argument applications, angle arguments containing predicates, mandatory
  semicolons, objects, tuples and native-island syntax.

## Supported execution

Module-level named functions with ordinary Type parameters; mutable `type` locals;
lexical block scopes; return, if/else, while, break/continue; required finite tuple
iteration; conservative closed-scalar union iteration; tuple construction and finite
tuple spreads; object construction with string/number keys and optional/readonly
properties; native indexing and `keyof`; native union/intersection construction;
number/bigint arithmetic and comparisons; literal boolean short-circuit operators;
native scalar identity/assignability predicates; calls to other named typefns and
ordinary native aliases/interfaces/classes, including native defaults and bounds.

Definitions reject missing names, wrong arity, duplicate/reserved bindings,
reads before local initialization, writes to immutable bindings and misplaced loop
control, including in unreachable branches. Runtime checks reject invalid scalar
families, division/remainder by zero, nonfinite arithmetic and unsupported tuple
shapes. Property keys use decoded names and native numeric normalization; equivalent
duplicate keys and nonfinite keys diagnose. Native recovery Types fail the concrete
calculation even when an intermediate value is discarded. Pending operations retain
checks for later specialization. Identical active applications diagnose cycles; changing-argument recursion
is bounded by call depth. Runtime fallthrough diagnoses.

## Important remaining work

* Native islands parse and print but execution diagnoses as unsupported. The lexical
  bridge must preserve local snapshots, native scopes and constraints before enabling
  dynamic function/mapped/signature construction.
* Strings can be native Type values, but string iteration, text/numeric/tuple
  intrinsics, interpolation and structural pattern matching are not implemented.
  The manual's HTTP-route and calculator examples are not executable yet.
* Constructor slots, closures/local typefns and restricted HKT support are pending.
* Deferred scalar expressions, calls, control flow, captured construction, and
  native substitution/replay are implemented. Full requirement-bearing native
  relation proofs, scoped generic-body assumptions, latent member/signature
  dependencies, and exhaustive inference/replay laws remain pending. Guarded
  relations currently require the same canonical check set. Non-scalar predicates
  and union traversal remain conservative; no constraint is enumerated as a value.
* General tuple spreads, symbol-key construction, reflection, descriptors and dynamic
  callable/constructable builders are pending. Enum-specific support stays deferred.
* Declaration source recipes, advanced refactorings/auto-import completion, complete
  incremental dependency/replay tests and broad API factory compatibility remain pending.
  Native symbols now support definition/references/rename, hover, scoped completion,
  angle-call signature help, semantic coloring, result hints, outline/folding and
  formatting. Real editor-session tests cover imported-definition edits, cross-file
  navigation/rename, malformed producers, cancellation and profile toggling.
  The generated TypeScript API is type-checked. Runtime tests cover DSL serialization,
  transforming visitors and preservation of expression/statement kinds by updates.
* Resource accounting is preliminary: instruction dispatch, DSL call depth, explicit
  tuple/property construction, frame allocation and arithmetic bigint magnitude are
  checked. This is **not** complete accounting for all live buffers, native graph
  visits, native allocations, literal parsing or retained compilation caches.
  Explicit `nativeWork`, `stringUnits` or `cacheBytes` overrides diagnose on DSL
  execution until their accounting is implemented. Full native-operation
  cancellation and budget instrumentation remain required before calling this a hardened service boundary.
* Definition checking still needs the complete kind/readiness analysis and conservative
  return-path analysis. There is no claim of optimized append or memoized VM calls.
  Parameter bounds currently use native Type grammar rather than full DSL expressions.
  Bound checking can evaluate header applications at definition time, and an earlier
  blocked bound delays later checks. Later-parameter references are accepted even
  though the R2 text describes preceding-parameter scope. These behaviors need an
  explicit contract decision; see the [constraint probes](where-constraints-investigation.md#4-current-dsl-behavior-that-matters-to-the-design).
  Explicit `: type` result annotations are accepted but not retained in the AST or
  printer; the result-kind representation needs to be completed.

## Validation and audit

`tsc/internal/typefntest` currently has 65 isolated virtual compiler-project fixtures,
using bundled libraries,
real config parsing, native Type equality assertions, exact diagnostic sets and exact
emit text. Fixtures run concurrently. The harness itself is calibrated with failures
for wrong assignments and accidental `any` results. Successful `Assert<...>` uses
must resolve to the literal Type `true`, so error recovery cannot silently pass
the assertions. Negative tests cover definition
errors, scalar errors, constraints, unsupported features, cycles and resource limits.
The frame-budget cases include iterator storage, in addition to register storage.
Optional-property assertions run with both exact-optional settings. Iteration tests
cover snapshots, nested loop control, readonly/empty tuples, special union Types
and symbolic leaves in a known tuple. A standalone CLI smoke test compiled the
execution fixture and emitted exactly `export const answer = 42;` with no DSL code.
Deferred fixtures additionally cover composed substitution, discarded checks,
native simplification/containers/defaults/constraints, inference, signature
rejection, local snapshots, and replay budgets. Contextual names such as `out`
and `async` are tested as both bindings and reads/assignments.

The parser has AST parent/span/source-order invariants, grammar/recovery tests, limit tests and a
fuzz target. An initial ten-second fuzz campaign executed 74,055 inputs without
failure. That is bounded evidence, not a proof of parser safety or full grammar
coverage; this historical campaign was not repeated for the review series. The printer has a stable parse/print/reparse test. The generated API has a DSL
encode/decode/print round-trip test. A real editor-session test changes an imported
function, checks the unchanged consumer and hover, restores the function, then disables
and reenables the DSL profile to check parse-cache isolation. Native LS tests cover
scope-aware queries, incomplete input, semantic tokens, overlays and formatting;
see [editor integration](editor-integration.md) for their scope and limitations.
The audit adds semantic diagnostics on incomplete buffers, navigation through parser
limit recovery, suppression of failed results, and comparison-argument signature help.

Run from the repository root:

```sh
go test ./tsc/internal/typefntest ./tsc/internal/parser ./tsc/internal/printer ./tsc/internal/ast ./tsc/internal/api/encoder ./tsc/internal/tsoptions ./tsc/internal/checker ./tsc/internal/compiler
go test -race ./tsc/internal/typefntest ./tsc/internal/parser ./tsc/internal/project -run 'TestIntegration|TestTypefn'
go test ./tsc/internal/parser -run '^$' -fuzz FuzzTypefnParser -fuzztime 10s -parallel 2
go test ./tsc/internal/ls ./tsc/internal/format ./tsc/internal/ls/lsutil ./tsc/internal/project ./tsc/internal/lsp
go test ./tsc/internal/astnav -run TestTypefn
go test -race ./tsc/internal/ls ./tsc/internal/project ./tsc/internal/lsp -run TestTypefn
node_modules/.bin/tsc -p packages/typescript/tsconfig.json --noEmit
node_modules/.bin/tsc -b packages/typescript-wasip1-wasm
node_modules/.bin/tsc -b packages/typescript/test --force
node --experimental-strip-types --no-warnings --conditions @typescript/source packages/typescript/test/typefn_ast.test.ts
node --experimental-strip-types --no-warnings --conditions @typescript/source packages/typescript/test/typefn_flags.test.ts
go test ./tsc/internal/typefntest -run '^$' -bench BenchmarkTypefnProject -benchmem
```

The benchmark measures complete cold compiler projects, including libraries and
native checking. It does not isolate VM throughput. A historical three-iteration Linux/amd64 run on an AMD EPYC 9V74 measured
237.8 ms/project and 75.7 MB/project for the native calibration, and 217.6 ms/project
and 76.6 MB/project for the execution fixture. These are different workloads and
a tiny sample; the numbers do not establish DSL overhead or a speed advantage.
Current-series samples and their limitations are recorded in the
[review ledger](review-series.md).

The full 141-case specification
conformance suite and wide community workload ports remain outstanding.

Implementation commits are deliberately separated for review; use
`git log --reverse --oneline f7973164a..feat/typefn`. The integration manifests are the executable
record of this milestone's behavior.
