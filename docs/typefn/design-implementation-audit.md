# Design and implementation audit

Date: 2026-11-06. Baseline reviewed: `a9d52843` on `typefn-redesign`.
The normative baseline is the [R2 specification](spec/README.md); public language
and configuration version remains 1. This review covers the implementation and
its claims, not a new language revision.

**Follow-up:** [Deferred calculations](deferred-evaluation.md) records the later
implementation of native terms, substitution/replay, and retained checks. The
verdict and coverage table below describe the reviewed baseline, not the current
status of that follow-up. Full scoped native proof/requirement integration remains
unfinished; see [current implementation status](implementation.md).

## Verdict

The implemented subset uses native TypeScript syntax trees, symbols, checker
Types, project snapshots and language-service requests. It is integrated into
the compiler rather than translated through a separate generated TypeScript file.
That integration does **not** make it an implementation of the complete language.

It remains an experimental concrete evaluator with some symbolic Type forwarding.
The central symbolic evaluation contract, the string workloads requested during
design, dynamic signature construction, HKT and declaration recipes are missing.
It is not ready to be described as mature, broadly community-compatible, or
performance-hardened.

The earlier editor checks were too narrow. They exercised successful queries and
some incomplete input, but missed semantic checking of recovery nodes and parser
resource recovery. API type checking and serialization also failed to detect
runtime transforming-visitor defects. This audit added regression tests at those
boundaries instead of treating successful compilation as sufficient evidence.

## Confirmed defects and corrections

| Defect | Correction and regression evidence |
| --- | --- |
| Object property keys used source spelling. `0x10` and `"16"` became different keys; escaped identifiers were not decoded correctly. | Use decoded names and native number-to-property-key normalization. Native Type-equality assertions cover hexadecimal, exponent, separator and escaped spellings. Equivalent duplicate keys and nonfinite numeric keys diagnose. |
| A failed native indexed access could be discarded or absorbed by a union, allowing hover to publish a successful result. Unresolved arguments and bounds could also pass through unused parameters. | Propagate native recovery Types as failure through the concrete VM, arguments and constraints. Editor tests assert that diagnostics remain and no successful hover/inlay result is published. This is not the deferred requirement protocol. |
| Reserved names could bind successfully but resolve as primitives or future DSL syntax. | Reject decoded reserved binding/function names; preserve keyword property names. Negative fixtures include escaped names and prelude names. |
| Long `else if` chains bypassed the DSL syntax-depth counter. Native parameter modifiers were silently accepted. | Count recursive statements and reject unsupported `in`, `out` and `const` parameter modifiers. Parser regressions cover both cases. |
| Generated JavaScript AST visitors skipped most DSL expression/statement kinds. Updaters could rebuild a binary expression as a name or a return as a binding. | Model shared nodes with generic kind parameters, generate a visitor entry for every kind, and retain the original kind in factory updates. Runtime tests transform every expression kind, a return statement, and a serialized remote tree. |
| Signature help treated `<` in a parenthesized comparison as an unclosed application and selected the wrong parameter. | Close tentative angle frames at the enclosing group delimiter. Test comparison arguments followed by scalar and tuple arguments. |
| Semantic diagnostics on incomplete DSL bodies could lower a missing name and panic. | Refuse lowering from a source file with syntax diagnostics. Tests request semantic diagnostics after ordinary editor queries on incomplete input. |
| Navigation inside source skipped after a parser limit could panic because an identifier had no corresponding AST child. | Token lookup recognizes skipped text inside DSL recovery nodes. Navigation tests query every source position in several malformed/resource-limited files; editor tests include an over-depth tuple expression. |

The last two fixes deliberately keep malformed files out of evaluation. Syntax
queries remain available. Recovery does not fabricate a usable Type or a bound
identifier for skipped source.

Code corrections are separated into reviewable commits:

- `0f7133d6`: generated DSL node kinds, factories and transforming visitors.
- `00748509`: statement nesting limits and unsupported parameter modifiers.
- `14d02475`: property identity, reserved names and concrete failure propagation.
- `cc4d19a7`: editor recovery crashes and comparison-argument signature help.

## Specification coverage and gaps

| Specification area | Implemented boundary | Remaining work |
| --- | --- | --- |
| [Syntax](spec/syntax.md) | Named Type-only functions, imperative statements, scalar expressions, tuples/objects, angle calls, native-island parsing. | Patterns, interpolated templates, constructor slots and closures. Parameter bounds currently use native Type grammar rather than the full DSL TypeArgument grammar. An explicit `: type` result annotation is consumed but not retained in the AST/printer. |
| Definition checking | Lexical symbols, mutable/immutable bindings, initialization, names, ordinary call arity and loop-control checks, including unreachable branches. | Complete kind/readiness and return-path analysis; recursive call-graph result-kind rules. Runtime fallthrough checking is not the specified static analysis. |
| Scalars and iteration | Number/bigint arithmetic, boolean control, required finite tuples and conservative closed-scalar union traversal. | String iteration, full structural predicates, general tuple shapes, symbolic observation and replay. Constraints are not enumerated as values. |
| [Construction](spec/semantics.md) | Simple tuples/objects and named native alias/interface/class applications use checker Types. Existing native identity can be forwarded. | Executable native islands with lexical snapshots, symbol keys, reflection/descriptors, dynamic call/construct signatures and overloads, recursive builders. Generic-type construction and enum-specific support remain outside the current goal. |
| [Library](spec/library.md) | Native generic applications provide a small amount of reuse. | The specified text, numeric conversion, tuple-state, reflection, builder and failure intrinsics. String literals can be carried as Types; this is not string parsing support. |
| [Symbolic evaluation](spec/evaluation.md) | Known structures can contain symbolic native Type leaves. Concrete failures now survive discarded intermediate values. | `TypeResult = (TypeTerm, Requirements)`, guarded/deferred checker representations, reached-frame requirement journals, dependency tracking and substitution/replay laws. No general requirement preservation across native simplification or generic signature relations. |
| Restricted HKT | Design only. | Constructor kinds, native constructor adapters, closure snapshots, constructor factories and their readiness rules. |
| [Native inference](spec/integration.md) | Concrete VM results enter native assignments and ordinary generic operations. | Deferred instantiation, generic runtime-body checking, requirement-bearing signatures and source/declaration equivalence. Concrete assignment tests do not establish these laws. |
| Declaration ABI | JavaScript erases DSL declarations; declaration emit fails explicitly. | Source recipes, dependency lifting, hidden guard dependencies and reconstruction. No declaration-only consumption claim. |
| Editor/API | Native symbols and the services listed in [editor integration](editor-integration.md); targeted generated-factory/visitor/wire tests. | Broad refactor/code-fix/auto-import compatibility, project references/source maps, execution tracing and general API compatibility. No graphical VS Code session was exercised. |
| Resources | Parser work/depth, VM dispatch/call depth, partial construction/frame/bigint limits. | Complete live-buffer, native operation, native allocation, string and cache accounting; cancellation throughout native relations. Some buffers, including tuple spread accumulation, are allocated before the later construction check. |
| Diagnostics | Native diagnostics and explicit DSL errors. | Consistent diagnostic categories, use-site attribution, bounded execution traces and input-offset errors from string parsers. |

## Design issues requiring implementation evidence

The syntax and semantic documents form a substantial design baseline, but several
contracts have no executable implementation proof. In particular, requirements
that survive union absorption, projection, generic instantiation and native
signature comparison need a prototype in the checker before the design can be
called validated. The same is true of native-island scope bridging and declaration
recipe reconstruction. A raw native Type pointer is insufficient to represent the
specified result and its obligations.

The finite constructor-kind restriction is useful, but does not remove these
integration obligations. HKT should be implemented on top of a working symbolic
result model, not as an arity-only wrapper around the present concrete VM.

The HTTP route and calculator listings remain target-language examples. They are
not executed by the current integration harness. The research's 39 sampled files
across 17 families document use cases, not successful ports of those libraries.
The 141 specification cases are acceptance requirements, not 141 passing tests.

## Next implementation gates

1. Implement the native result/requirement adapter and substitution protocol.
   Test discarded checks, absorbing unions, projection, generic constraints,
   signature assignment and repeated instantiation before extending the surface.
2. Implement readiness and replay with dependency-aware cache/cancellation rules.
   Verify the substitution law, including symbolic inputs nested in known shapes.
3. Add string iteration, text/numeric intrinsics and pattern matching against the
   actual HTTP binding and calculator programs, including invalid input and UTF-16
   edge cases. Keep executable workloads beside their specification requirements.
4. Build the native-island lexical bridge and dynamic Type builders; then add
   restricted constructors and closure snapshots with kind/call-graph analysis.
5. Implement declaration recipes and compare source consumers with declaration-only
   consumers. Complete resource instrumentation and adversarial tests before making
   robustness or large-project latency claims.

Each gate needs meaningful compiler assertions and an auditable implementation
commit series. Editor features must continue to be tested on invalid and canceled
snapshots, not only successful examples.

## Validation record

The executable compiler fixture count increased from 34 to 38. New regressions
also live in the parser, token navigation, language service and JavaScript AST
API suites. See [implementation status](implementation.md) for commands. Passing
these checks establishes the corrections above; it does not close the coverage
gaps in this audit.

The affected Go package test commands passed: `astnav`, `ls`, `project`, `lsp`,
`parser`, `api/encoder`, `typefntest`, `checker`, `compiler`, `format`, `ast` and
`printer`. Native-preview API type checking and test-project compilation passed;
the four new DSL AST runtime tests and 26 existing encoder runtime tests passed.
The compiler and extension rebuilt successfully, and the rebuilt compiler checked
the editor example project. The full upstream repository suite and large community
projects were not run in this audit.

Targeted DSL tests passed under the race detector in `ls`, `project`, `lsp` and
`typefntest`. A ten-second parser fuzz run completed 58,370 executions without
failure. Initial concurrent instrumented builds exhausted the workspace disk;
after removing inactive generated caches, the unfinished checks were rerun
sequentially and passed. This bounded fuzz run does not establish parser safety.
