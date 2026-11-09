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
the language profile. MergeDeep, Jsonify and the other community families still need
their own ports and evidence; this scanner does not stand in for them.
