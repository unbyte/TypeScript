# Design review and implementation gates

R0 is a specification candidate. None of its examples have been executed by a new
implementation. This branch contains documentation only on top of the upstream
compiler baseline. The archived experiment's passing tests do not validate R0.

## Review findings resolved in this draft

| Finding | Resolution |
| --- | --- |
| The initial redesign made `type` immutable contrary to the user's preference | `type` is a mutable Type binding inside the DSL; `const type` is immutable |
| Requiring `type(...)` around every result made common calculations noisy | Type binding/assignment/return positions use direct Type syntax |
| Identical braces could mean data or a native Type without a rule | Expression modes are explicit; data-path assignments and lambdas use data mode |
| Plain mutable collections threatened saved type/descriptor snapshots | Value semantics and local-root updates; closure captures are immutable snapshots |
| Ordered Map iteration conflicted with order-independent memo keys | Map equality/cache identity includes insertion order; Set/TypeSet remain unordered |
| Symbolic construction could erase an unverified generic constraint | Pending validation suspends instead of publishing an unchecked result |
| Exposing an unresolved Type category would let programs observe inference timing | Unresolved classification blocks; only tooling may display that state |
| Async signature conversion could copy an invalid type predicate | The worked transformation explicitly rejects predicates and generic signatures |
| Native match helper used code points while JS expectations use UTF-16 | R0 explicitly chooses UTF-16; implementation must adapt rather than inherit the helper |
| Copying an origin-restricted class descriptor could silently lose identity | Unchanged round trip or explicit public projection; edits cannot forge identity |
| Nominal enum identity could depend on evaluation/cache order | Explicit owner, operation and key identity with transactional registration |
| Parser failure, resource exhaustion and generic blocking could all become never | Separate Result data, diagnostics and deferred applications |

These are design-review decisions, not proof that no further issue exists. A new
counterexample must update the specification and examples before being hidden in
an implementation-specific exception.

## Syntax and evaluation oracles

The implementation suite must contain at least these independently meaningful
cases, asserting native assignability or data results, not only displayed text:

| Area | Positive oracle | Negative/edge oracle |
| --- | --- | --- |
| Type binding | Save `{x:T}`, reassign T local, saved field unchanged | Assignment to const type or a parameter diagnoses |
| Data boundary | String data converted with literal gives a literal Type | Returning data record from Type entry diagnoses |
| Type context | Direct object/tuple/function/conditional expressions work | Data projection of kind string cannot silently become a Type |
| Names | Imports, shadowing and separate pattern scopes resolve | Read before initialization, duplicate captures and leaked names diagnose |
| Helpers | Mutual recursion and first-class pipeline work | Wrong argument kind/arity, helper used as native generic diagnose |
| Local data | List/record snapshots survive later path updates | Mutation of a parameter, capture or temporary diagnoses |
| Loops | Iteration snapshots, break and continue compose with match | Mutating source collection does not extend current iteration |
| Strings | Empty captures, pinned delimiters, earliest split and full anchoring | Nonstring pin, absent delimiter and no-progress loop behavior |
| Unicode | Astral input splits into UTF-16 units; explicit codePoints does not | Lone surrogates preserved, no hidden replacement/normalization |
| Match | First successful arm and guard scope | Blocked earlier arm must not fall through to wildcard |
| Data variants | Exhaustive Option/Result/tagged cases accepted | Missing cases without wildcard diagnose |
| Type patterns | Required tuple/rest and required object fields | Open array/optional tuple do not masquerade as required finite tuple |
| Union | Per-member transformation preserves correlations | Never empty, any singleton, absorbed literals not recoverable |
| Ordering | Ordered tuple/data list and Map order preserved | TypeSet order never asserted as a portable tuple order |
| Arithmetic | Exact integer arithmetic beyond number's safe range | Overflow budget, inexact conversion, nonfinite result and zero division |
| Diagnostics | Failures point to operation and current application | A cached failure must not cite an unrelated earlier call |

## Native type semantics and generic integration

Run these under relevant combinations of strictNullChecks,
exactOptionalPropertyTypes, noUncheckedIndexedAccess and strictFunctionTypes:

* Box<T> exposes `.value` in a generic function and supports ordinary native
  inference from a concrete argument.
* Fixed symbolic tuples can be reflected and rebuilt; unknown-length/unknown-key
  observations suspend.
* Branches on unresolved relations do not substitute constraints or convert
  lack of proof into false.
* A constraint-erasing native alias still validates the DSL's computed argument.
* Nested suspended calls substitute correctly, including native outer generic
  captures, defaults and constraints.
* Symbolic Type constructors with universally valid constraints publish usable
  structure; pending validation does not disappear on later instantiation.
* any, unknown, never, open string, branded string, enum and union cases use their
  specified contracts instead of sharing an “unknown” sentinel.
* NoInfer, native variance, unique symbols, private/protected identities and
  recursive types survive preservation operations.
* Optional fields round-trip without confusing read type, declared type and
  absence; inherited fields and numeric/string/unique-symbol keys remain correct.
* Object call/construct/index signatures and overlapping indexer checks survive
  reification. Duplicate effective keys reject.
* Tuple labels, optional/rest/variadic/readonly metadata round-trip. Illegal
  transformations reject instead of degrading to an array.
* Overload order, this, predicates, assertion predicates and abstract constructors
  are preserved; unsupported generic signature edits reject.
* Recursive graphs seal atomically; unguarded recursion and premature inspection
  fail. Failed or canceled graphs do not leak into relations/cache entries.
* Nominal enums repeat identity for equal owners/arguments/keys, differ where
  specified, reject conflicting keys and survive declaration/project-reference
  redirects. Structural enumValues does not acquire nominal identity.

## Community acceptance matrix

| Family | Required workload | Evidence beyond a concrete output |
| --- | --- | --- |
| type-fest MergeDeep | Optional/undefined/indexer/object/tuple policy matrix | Native reference comparisons, generic use and allocation profile |
| type-fest Jsonify | Object/array position, optional length, toJSON once, brands | Explicit representability limits and negative serialization cases |
| type-fest Words | Acronyms, digits, punctuation, brands and broad string | Upstream expected outputs and selected option-union cases |
| React Hook Form | Paths/PathValue with cycles, depth, arrays and optionality | Per-union visited state and native generic form usage |
| ts-pattern | Selective nested union products and deep exclusion | Correlation, negative patterns, unrepresentable complements and explosion limits |
| Kysely/Drizzle | Selection aliases, directional columns and outer joins | Generic callbacks, whole-object nullability and fluent method availability |
| tRPC/Zod | Parser input/output, context and native variance | Argument authority, inferred callback inputs and rejected wrong outputs |
| TypeBox/Effect | Static/decoded/encoded/requirements transformations | Recursive references, opaque generics and NoInfer behavior |
| ArkType | Parser state, precedence, scopes, diagnostics | Full consumption, malformed inputs and completion candidates |
| gql.tada | Tokens, parser, variables/defaults, fragments and selections | Schema-directed semantic phase and nullability distinctions |
| TanStack Router | Required/optional/rest params, escapes, route resolution | Source/destination inference, broad fallback and negative routes |
| XState | Event indexes and correlated state/actor selection | NoInfer/low-inference-sensitive API calls |
| hotscript/fp-ts | Higher-order pipelines and native open registries | Closure identity, arity errors and module augmentation |
| ts-toolbelt Curry | Nongeneric parameter metadata and placeholder policy | Explicit exclusion of newly generated generic signatures |
| Type challenges | All 15 query and 9 restricted JSON cases | Extra malformed cases, Unicode, symbolic inputs and resource behavior |

Investigated, expressible in principle, implemented, fixtures passed, and upstream
suite passed are distinct statuses. The source manifest currently establishes
inspection only. Do not upgrade a row's status because a different library passed.

## Integration and robustness gates

The new compiler implementation must pass its own tests and unchanged upstream
regressions. In addition:

1. CLI/API/LS use the same semantics and ABI. Generated AST visitors, transport,
   clone/print and recovery operate on dedicated nodes.
2. JavaScript erasure, declaration dependency preservation, import/reexport,
   namespaces, project references and declaration-only consumers work.
3. Hover, rename, references, definitions, completions, semantic tokens and
   formatting work in both expression modes and in incomplete patterns.
4. Edits to a helper, native dependency, compiler option and source declaration
   invalidate results; obsolete requests never publish results or diagnostics.
5. Parser fuzzing, semantic small-program generation and resource/cancellation
   stress tests do not panic, hang or publish partial native types.
6. Repeated failure/cancellation/cycle detection leaves cache and transaction
   state clean. Checker concurrency passes race testing.
7. Source inputs cannot reach host I/O or arbitrary Go/JavaScript calls.

## Performance gates

Measure cold and warm execution separately, with allocation and peak-memory
profiles, for tuple reversal, large-string scanning, nested parser trees, wide
objects, correlated products, descriptor round trips and generic applications.
Include adversarial cases: repeated snapshots, long failed delimiters, deep
recursion, repeated canceled evaluations and single-instruction native expansion.

Compare against the same pinned native compiler and relevant stock type encodings,
without treating a failed baseline as a slow successful result. Record machine,
options, corpus, sample counts and variability. “No DSL” projects must not suffer
an unexplained regression. No absolute throughput or maturity claim is made
before these measurements exist.

## Implementation sequence after design review

1. Dedicated grammar/AST, binder and static kind checker; no execution of invalid
   programs. Validate formatting/recovery/transport immediately.
2. Data/Type value VM and native adapter with symbolic publication/suspension,
   constraint validation, budgets and cancellation as core contracts.
3. Reflection/construction, matching and the worked parser/pipeline workloads,
   including generic and negative fixtures alongside each capability.
4. Native inference/relations audit, recursive/nominal types, declaration and
   full language-server integration.
5. Community differential corpus, incremental/race/fuzz testing and measured
   performance work. Only then assess a mature-release claim.

This sequence is not permission to call an early stage complete. It makes the
requirements reviewable before another compiler experiment constrains the design.
