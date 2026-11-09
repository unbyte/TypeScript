# Executed community workloads

These cases test defined portions of community algorithms against native TypeScript
observations. They do not establish whole-library compatibility. The remaining
families and required boundaries are listed in the [conformance matrix](spec/conformance.md#9-community-workload-traceability).

## type-fest: finite-literal word scanner

[Imperative scanner](../../tsc/internal/typefntest/testdata/community/typefest/words.ts)
and [native comparison](../../tsc/internal/typefntest/testdata/community/typefest/words-native.ts).
The [origin record](../../tsc/internal/typefntest/testdata/community/typefest/ORIGIN.md)
identifies the inspected source bytes, dependency specialization and upstream license.

The supported workload is `_Words` on one unbranded string literal, with resolved
number/punctuation options. `WordsUnion` explicitly distributes over input members.
The scanner handles UTF-16 units, word separators, case transitions, acronym tails,
numeric sequences and optional punctuation boundaries. It uses native casing through
Type islands; it neither invents Unicode casing rules nor treats every non-ASCII
character as a separator.

Evidence:

* 464 comparisons against the original conditional-type scanner, run directly and
  after declaration-only emission into a fresh checker without producer sources.
* Fixed expected tuples for acronyms, punctuation, numeric boundaries, accented
  capitals, Japanese text, emoji and whitespace; the corpus also includes lone
  surrogate units and Unicode titlecase/combining characters.
* Generic specialization, boolean option behavior, explicit unions and never unions.
* Exact tuple descriptors and reconstruction, including requiredness and readonly state.
* Rejection of broad, union, never and branded operands by the scalar entry point.
* Instruction-budget failure in source and declaration-only consumers, including
  skipLibCheck; no partial result is accepted as successful completion.
* A 512-character word succeeds within 4,096 string-work units. A concatenating
  implementation failed this case; storing offsets and slicing final words passes.

The [testing record](testing.md) includes the reproducible cold-project benchmark.
It includes both implementations and assertions, so it cannot establish a VM speedup.

Still missing for public `Words` compatibility: brand/tag normalization, public option
defaulting and broad/infinite string fallback. Enum-specific behavior remains outside
the language profile. MergeDeep, Jsonify and the remaining community families still need
their own ports and evidence; this scanner does not stand in for them.

## React Hook Form: field paths

[Imperative traversal](../../tsc/internal/typefntest/testdata/community/rhf/paths.ts)
and [native comparison](../../tsc/internal/typefntest/testdata/community/rhf/paths-native.ts).
The [origin record](../../tsc/internal/typefntest/testdata/community/rhf/ORIGIN.md)
records the source snapshots, selected shared contracts and MIT license.

`FieldPaths` ports `FieldPath` for object/form inputs. It explicitly distributes
unions, traverses native keys and indexed Types, distinguishes tuples from arrays,
and carries a visited-Type union through recursion. Native equality preserves
recursive and nominal identity. Primitive, browser-native and registered opaque
Types are leaves. The default cutoff is ten segments; `FieldPathsAt` exposes
deliberate truncation for depths 0 through 9. Exhausting a VM budget diagnoses
instead of returning that truncated algorithmic result.

Evidence:

* 25 native comparisons plus fixed expectations for nested and optional properties,
  required/optional never, unknown/any, browser objects, self/mutual recursion,
  readonly/optional/variadic tuples, arrays, string/number/symbol keys and indexers,
  nominal classes, unions, dotted keys and callable properties.
* Direct and fresh declaration-only programs, repeated with exact optional properties
  and unchecked indexed access enabled. Explicit depth and final-segment assertions
  check the cutoff independently of the native oracle.
* Registered nominal leaves and their subclasses stop traversal; unrelated classes
  with distinct private brands remain traversable, including declaration consumers.
* Generic specialization, native input inference, generic forwarding and invalid
  path arguments. This workload exposed and now guards against a compiler stack
  overflow during contextual literal checking of a deferred computation.
* Consumer interface augmentation recomputes aliases and inferred native-call
  results. A non-augmented consumer checks their original output.
* Invalid input/depth diagnostics and call-depth exhaustion in source and fresh
  declaration-only consumers, including skipLibCheck.

ArrayPath and PathValue are not included yet. This port does not claim compatibility
for unrestricted primitive roots or the complete React Hook Form library.
