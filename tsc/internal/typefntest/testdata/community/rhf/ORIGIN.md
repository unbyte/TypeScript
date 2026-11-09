# React Hook Form field-path workload

The imperative port targets FieldPath traversal for object/form inputs. It preserves
union distribution, native keyof/indexing, tuple/array dispatch, visited-type equality,
browser/registered opaque leaves and the default maximum of ten path segments.
FieldPathsAt exposes deliberate truncation for integer depths 0..9, matching the
native internal traversal on that range. FieldArrayPaths and FieldArrayPathsAt
separately port ArrayPath with the same depth/domain contract: primitive-only arrays
are excluded, element unions distribute, and any/never retain upstream behavior.
FieldPathValue and FieldArrayPathValue port PathValueImpl lookup with their
respective path bounds. They explicitly traverse input/path unions, retain
optional undefined reads and native nominal results, and use native conditional
helpers for symbolic template inference and generic indexed-access narrowing.

Native algorithm: https://github.com/react-hook-form/react-hook-form/blob/master/src/types/path/eager.ts
Inspected 2026-11-06 snapshot SHA-256: `3f45ff3c90ee307dce876423430404ad1daef7a9cf8ee7633816a175a25b90d2`.
The conditional Path implementation is retained in paths-native.ts and ArrayPath
in array-paths-native.ts, with PathValueImpl in path-value-native.ts (same eager.ts
snapshot); their scalar/key
helpers are selected from upstream common.ts. Native leaf/equality/opaque registry
contracts are selected from upstream utils.ts, fetched on 2026-11-09. Both algorithms
share those declarations so nominal and module-augmentation identities agree.
Fetched helper snapshots (SHA-256 of UTF-8 contents with a trailing newline):

* [utils.ts](https://github.com/react-hook-form/react-hook-form/blob/master/src/types/utils.ts):
  `cd3cf7c0d2e4ad42649a6af799911491221bcd0b96e45ec6730b20b2f65ab093`.
* [path/common.ts](https://github.com/react-hook-form/react-hook-form/blob/master/src/types/path/common.ts):
  `5de90887eccc5eed5bb20a22988d21008c012caa78006c1f6bd06349d8713d7c`.

Branch URLs are mutable; this is a source-byte snapshot, not release compatibility.
The upstream MIT license is included in LICENSE.
