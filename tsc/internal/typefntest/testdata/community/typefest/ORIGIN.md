# type-fest word-scanner workload

This ports `_Words` for finite, unbranded string literals and resolved boolean options.
It is not full compatibility with the public `Words` API: brand/tag normalization,
option-default normalization and broad/infinite string fallback remain separate work.
Enum-specific behavior is outside the DSL profile.

The scanner body and its two local helpers in `words-native.ts` retain the conditional
algorithm from the inspected upstream file. Its character predicates are specialized
to single unbranded UTF-16 units: numeric template matching and native case intrinsics.
Whitespace exits before numeric classification, so the upstream `Trim` check has no
remaining effect on these units. This restriction does not cover arbitrary strings,
brands or broad templates. The DSL scanner uses a loop and independent state bindings.
The native implementation is not a transpilation of that loop.

Upstream: https://github.com/sindresorhus/type-fest/blob/main/source/words.d.ts
Inspected snapshot SHA-256: `78adb990f3e50a9350a1b65a62ecd18bed02eac7113a2dd3dddd7c5930c5f1ee`.
The branch URL is mutable; this fingerprint identifies the inspected bytes, not a
claimed immutable upstream release. The research snapshot was collected 2026-11-06.

`characters.ts` was fetched on 2026-11-09 from
https://github.com/sindresorhus/type-fest/blob/main/source/internal/characters.d.ts
and is included verbatim. SHA-256: `69363d1204e72a3bf5f21881ab360420bea351be67626a39c2e4fce6732acb6d`.
Upstream's MIT license is included in `LICENSE`.

The fixture compares native Types in direct, generic-specialization and fresh
declaration-only programs. Fixed expected outputs provide checks independent of the
native comparison. Resource failures must diagnose rather than return partial words.

The scanner keeps UTF-16 start/cursor offsets and slices completed words once.
The original concatenating port exceeded a 4,096-unit string-work budget on a
512-character word; the offset-based port passes that budget. This is a checked
resource bound for that case, not a claim about every native operation's complexity.
