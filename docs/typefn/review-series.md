# Typefn implementation review series

## Baseline and review contract

`feat/typefn` starts at Microsoft TypeScript main
`f7973164a5f079a2d8b88d0fdb39347222549124`, fetched on 2026-11-08.
Upstream and the fork's main both resolved to this commit. The previous implementation
is preserved on `typefn-redesign` at `7e650c6ce9204c5511fe1cc5e4fd97102bd6b2ef`.

The first commit contains documentation only. Earlier audit and validation records
in this directory describe their named historical branches. They do not establish
that an intermediate commit in this series implements or passes the whole language.
The implementation status and target specification remain distinct.

Each subsequent commit should answer one review question, include its corresponding
regression coverage, and compile against its parent. Generated outputs belong with
the source change that requires them. An unsupported feature must diagnose rather
than silently produce a successful recovery Type. This is a personal fork; preparing
a reviewable series does not imply that an upstream PR will be opened.

## Architectural review

The existing native AST, bounded parser, lexical binder, native Type handles, and
checker-local deferred terms are appropriate foundations. Keep their ownership and
cancellation boundaries. Do not turn the VM into an independent structural type
checker, parse JavaScript in bodies, or lower the language to conditional-type text.

The existing change bundles obscure independent responsibilities:

* Configuration/profile validation and editor preferences can be reviewed separately.
* Syntax representation, parsing, printing, binding, and protocol traversal have
  different invariants and deserve their own commits and tests.
* The thousand-line evaluator combines the lowered program representation, definition
  checking, frame execution, native applications, construction, and scalar operations.
  Separate these responsibilities when introducing them, keeping private checker access
  inside the checker package rather than adding a large exported host abstraction.
* Checker-owned program, term, syntax-dependency and constructed-object tables need
  one explicit ownership boundary. Keep allocation lazy, and verify ordinary native
  operations retain their allocation behavior.
* Iterator state is allocated as a map for every call, with separate iterator objects.
  Review a dense program-assigned iterator layout, allocate only for actual iterator
  programs, and include frame storage in the existing budget accounting.
* Deferred calculation representation and replay are distinct from retaining reached
  checks through native simplification and instantiated signatures. Review both with
  executable positive and negative cases, including repeated failures.
* Language-service symbol queries, computed results, signature help, semantic tokens,
  formatting, preferences, and the public AST/RPC API can be layered independently.

Do not combine this restructuring with speculative new language semantics. In
particular, full HKT, strings/pattern matching, declaration recipes, and `where`
remain separate work. Preserve the documented constraint probes until their scope
and timing questions receive an explicit contract decision. No general result cache
is added without dependency and diagnostic-replay laws.

## Planned dependency order

1. Preserve documentation and record this review.
2. Establish the native compiler-project harness and native calibration fixtures.
3. Register and validate the generated language profile and API configuration.
4. Introduce native syntax nodes and generated representation.
5. Add bounded parsing and profile-aware source caching.
6. Add source printing and type-only emit behavior.
7. Bind DSL declarations and lexical references in native scopes.
8. Introduce lowered program representation and definition validation.
9. Execute concrete calculations with native Types and shared root budgets.
10. Represent and substitute immutable deferred calculations.
11. Retain deferred obligations through native operations and repeated use sites.
12. Improve state ownership and iterator/frame allocation with focused tests.
13. Integrate native editor symbol/navigation services.
14. Add computed display, completion, signature help and semantic classifications.
15. Add native formatting and generated opt-in result hints.
16. Complete generated public AST traversal and serialized API behavior.
17. Verify project invalidation, cancellation, mapped projections, and LSP/RPC boundaries.
18. Record final validation and the actual commit map.

The exact cuts may change when a dependency requires an atomic change. Document the
reason instead of introducing temporary bypasses, skipped assertions, or unbuildable
commits merely to increase the count.

## Completed architecture review

The reconstruction preserves the previous implemented language slice and adds these
changes rather than merely distributing the old diff across commits:

* Lowering, frame execution, native applications, construction, scalar observation,
  arithmetic, deferred representation, and editor queries have separate private
  modules inside the native checker. There is no second structural type checker or
  exported VM host facade.
* One lazy `typefnState` owns the program/term/object/dependency tables and active
  root VM. Ordinary native checking does not allocate this state. Snapshot isolation
  has a direct regression test.
* Lowering assigns dense iterator indices. Each invocation owns its iterator slice;
  non-iterating programs allocate no iterator backing storage. Re-entry, snapshot
  iteration, and nested break/continue retain their existing tested behavior.
* The live-frame estimate includes native-sized register/argument pointers and
  iterator slots, plus the existing bookkeeping allowance. Very tight `liveBytes`
  limits can now reject calls whose iterator storage was previously omitted. This
  is an intentional accounting correction, tested with an explicit rejection case;
  it is not a claim that all checker or Go heap memory is budgeted.
* Disabled-profile checking skips DSL declaration scans and the extra constraint
  type-node query. Native declaration diagnostics also skip whole-AST DSL scans
  when the profile is disabled. Native operand processing skips obligation scans when no DSL
  state exists, and unguarded results bypass journal deduplication. Existing native
  composite-cache tests still require zero allocations.
* Direct binder/lowering/Type ownership tests supplement the compiler harness.
  Editor tests are split by service. The API regression now inspects a binary DSL
  expression received from the Go encoder, including source-order child iteration.

Some dependency cuts differ from the initial plan. AST metadata, generated factories,
kind values, and binary traversal are atomic because generated consumers must agree.
The parser was first tested through explicit parse options; normal project activation
waited for checking and emit support. The immutable Type representation precedes the
VM. The VM and its private replay operations form one mutually dependent execution
commit; native simplification/substitution hooks are reviewed separately. The broad
compiler fixture set arrives with the small project-activation change, rather than
introducing unsupported fixtures into earlier checkpoints.

## Actual commit map

Each implementation checkpoint was tested before committing; the commit bodies state
its validation. The first commit is documentation only. Generated files accompany
the inputs that require them. The final documentation update records results without
changing the tested implementation checkpoint, `ec295e1ef`.

| Commit | Review boundary |
| --- | --- |
| `57613a1cc` | docs(typefn): preserve language design and define implementation review boundaries |
| `54420511b` | test(typefn): establish compiler-project fixture harness with native calibration |
| `62f253e8f` | feat(typefn): register validated opt-in compiler profile and resource limits |
| `ca1c60295` | feat(ast): represent typefn syntax in native and public ASTs |
| `3959d376b` | feat(parser): parse bounded typefn bodies with native recovery nodes |
| `c7a0aa9dc` | feat(printer): round-trip typefn source and erase it from JavaScript |
| `fd5fba9ea` | feat(binder): give typefn declarations and locals native lexical scopes |
| `103dfb4f7` | feat(typefn): lower lexical bindings and control flow into checker-local programs |
| `3a67f76bd` | feat(checker): represent deferred typefn terms as tagged native Types |
| `da6c0fee1` | feat(typefn): execute lowered programs and retain immutable calculation recipes |
| `24382ea6f` | feat(checker): preserve and replay typefn obligations through native operations |
| `e65aead45` | feat(compiler): activate typefn project checking with complete integration fixtures |
| `0095af7d7` | refactor(checker): centralize lazy typefn state within each checker snapshot |
| `b82eeb805` | perf(typefn): use dense frame-local iterator slots and charge their storage |
| `f62da6730` | perf(checker): preserve native fast paths outside the typefn profile |
| `59228d970` | feat(ls): resolve typefn symbols through native navigation and rename |
| `ebe58d947` | feat(ls): display typefn signatures and completed results in native hover |
| `283e4ba40` | feat(ls): complete lexical typefn names and track constructor arguments while typing |
| `36f324f12` | feat(ls): classify typefn syntax and bound symbols with native semantic tokens |
| `72b0416fd` | feat(format): format typefn bodies without erasing required terminators |
| `502ba7a58` | feat(ls): expose opt-in completed typefn result hints through generated preferences |
| `c181b4361` | test(project): cover typefn invalidation, profile changes and canceled checker reuse |
| `b1dd313df` | test(project): exercise typefn through primary and supplemental source projections |
| `43aea24d8` | test(lsp): cover typefn editing through real protocol requests |
| `6e82a04e8` | test(api): verify typefn options, native results and Go-encoded syntax over RPC |
| `02344ddc8` | style(typefn): apply repository import grouping to the extracted modules and tests |
| `ec295e1ef` | perf(compiler): skip typefn declaration scans when the language profile is disabled |

## Validation ledger

Validation date: 2026-11-08. Toolchain: Go 1.27.0 and Node 24.22.1 on Linux/amd64.
The implementation remains experimental; a green integration suite is not full
language-spec conformance or an exhaustive native compiler regression run.

* Passed: 65 isolated compiler-project fixtures, including iterator-frame accounting
  and native declaration output with the profile both off and on. The harness checks complete diagnostics, exact emit, and literal-true
  assertions, with native positive/negative calibration.
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
  The first commit was verified to contain documentation only; `typefn-redesign`
  still points to its original preserved revision.

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

### Performance evidence and limits

Cold-project probes used `BenchmarkTypefnProject`, with three iterations in each of
three runs, before the ownership/iterator changes and after the iterator/operand
fast-path changes. Native, execution and iteration projects fell roughly in the
149–174 ms/project range and 43–45 MB/project range. Setup, bundled libraries,
config parsing, native checking and allocation noise dominate these measurements.
They do not demonstrate a reliable throughput gain or isolate VM cost. No speedup
claim is made. A dedicated VM throughput study and representative large-project
profiling remain future work. Reproduce the current workloads with:

```sh
go test ./tsc/internal/typefntest -run '^$' -bench BenchmarkTypefnProject -benchtime=3x -count=3
```

### Remaining scope

This series does not implement the outstanding full string library/parser workloads,
HKT, `where`, native islands, declaration recipes, or full graph-readiness protocol.
The specification's 141-case target conformance suite and broad community workload
ports remain outstanding. The existing constraint scope/timing observations are
preserved; this restructuring is not a resolution of those design questions.
See [implementation status](implementation.md), the [audit](design-implementation-audit.md),
and the [constraint investigation](where-constraints-investigation.md). No interactive
VS Code session or complete upstream test corpus is claimed by this ledger.
