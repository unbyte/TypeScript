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

## Validation ledger

Pending: this documentation commit contains no implementation. Validation entries
will identify the capability added at each checkpoint, followed by final full native,
race, generator/API, editor build, and selected upstream baseline checks.
