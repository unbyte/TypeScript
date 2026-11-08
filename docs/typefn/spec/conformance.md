# Conformance and implementation gates

Normative R2 acceptance requirements. **None of these rows is reported as a passing
R2 compiler test.** Expected Types are compared using native semantic observations
and diagnostics, not display text alone. Use independent fixtures for intentional
errors so recovery Types cannot contaminate successful cases.

Run semantic cases both directly and through symbolic specialization where relevant,
then through declaration-only consumption in a separate checker. Keep producer
sources absent for declaration-only fixtures; test source redirects separately.
Use strict mode, plus a matrix for exactOptionalPropertyTypes, strictFunctionTypes,
noUncheckedIndexedAccess, module resolution and supported native targets.

## 1. Grammar and binding

Fragments occur inside a typefn unless stated otherwise.

| ID | Case | Required outcome |
| --- | --- | --- |
| G01 | `return F<A + 1, G<B>>;` | Two applications, with arithmetic in the outer call's first argument |
| G02 | `return A < B && C > D;` | Two comparisons, logical conjunction |
| G03 | `return Choose<(A < B)>;` | Grouped predicate argument |
| G04 | `return F<T>;` versus `return F < T >;` | Identical application parse |
| G05 | `return A < B > C;` | Comparison-chain diagnostic, not partial call acceptance |
| G06 | `return A is B \| C;` | Assignability to union |
| G07 | `return type((x: T) => Promise<T>);` | One native Type island |
| G08 | Missing semicolon before newline/return | Syntax diagnostic, no ASI |
| G09 | function, arrow expression, regex, destructuring, method call | Unsupported syntax diagnostics |
| G10 | Rename/remove the bindings in G01–G06 | Parsing unchanged; name/kind diagnostics occur later |
| G11 | Deep nested/near-matching unclosed angle lists | Bounded recognition work and recoverable diagnostic |
| G12 | `type X = F<A + 1>;` outside DSL | Native Type syntax diagnostic |
| G13 | Braces/angles in escaped strings/templates and native islands | Correct mode termination and source spans |
| G14 | Assignment to a parameter/capture; duplicate local; self-initializer | Definition diagnostics |
| G15 | Unknown name/wrong kind inside known-false branch | Definition diagnostic without executing branch |
| G16 | Recursive return kind without sufficient annotation | Definition diagnostic; no evaluation to infer kind |
| G17 | Local self-recursion versus reference to a later local helper | Self-reference legal; later helper rejected |
| G18 | `F<T><U>` / binding result then applying it | First rejected; second accepted with matching kind |
| G19 | `is F<T> => { ... }` and a guarded `is F<T>` pattern | Application recognized before pattern arrow/guard |
| G20 | `type x = { readonly: true; readonly value: T };` | First readonly is a property name, second is a modifier |

## 2. Scalar Types and state

Use small wrappers around operators/intrinsics, since scalar expressions outside
a body remain native Type grammar.

| ID | Case | Required outcome |
| --- | --- | --- |
| A01 | 17 + 25; 2 + 3 * 4 | 42; 14 number Types |
| A02 | 9007199254740992n + 1n | Exact 9007199254740993n |
| A03 | Number 9007199254740992 + 1 | Native binary64 result 9007199254740992 |
| A04 | Bigint -7n / 3n and -7n % 3n | -2n and -1n |
| A05 | Zero divisor, nonfinite number result | Domain diagnostic |
| A06 | Number/bigint mixed operation | Domain diagnostic, no coercion |
| A07 | Broad/union/any/unknown/never/branded/enum arithmetic | Domain diagnostic; no distribution/unwrapping |
| A08 | toNumber of 9007199254740992n / 9007199254740993n | Exact first result; reject second |
| A09 | Integral/fractional toBigint | Exact represented integer / diagnostic |
| A10 | `-0 == 0`; arithmetic producing -0 | true; native zero canonicalization |
| A11 | Short-circuit false && failing right call | Right call not reached |
| A12 | Condition 0, string, broad boolean or never | Domain diagnostic, no truthiness |
| A13 | Snapshot an object/closure, then reassign captured local | Snapshot retains old value |
| A14 | Reassign a Type local from number literal to string literal | Valid same-kind assignment |
| A15 | Reassign a Type local to constructor | Kind diagnostic |
| A16 | `0 * T`, `T < T` with unresolved T | Deferred eligibility; no unsafe constant folding |

## 3. Strings, patterns and calculator

| ID | Case | Required outcome |
| --- | --- | --- |
| S01 | Characters<"12+3"> | Four literal-unit Types in order |
| S02 | Characters of astral character / lone surrogate | Two units / one retained unit |
| S03 | peek at length / beyond length / fractional index | Empty literal / diagnostic / diagnostic |
| S04 | slice with negative/reversed/oversized bounds | Diagnostic, no JS clamping |
| S05 | split<"a::b::","::"> | ["a","b",""] |
| S06 | Empty delimiter in split or pinned pattern | Diagnostic |
| S07 | `${a}:${b}` over "a:b:c" | a="a", b="b:c" |
| S08 | `${a}:end` over "x:end:end" | No-match; no last-delimiter backtracking |
| S09 | `${a}${b}` over "ab" / "" | ["a","b"] / no-match |
| S10 | Sole capture over empty string / broad string | Empty literal / broad string Type captured |
| S11 | SplitBinding of `GET /${string}` / broad string | Preserved route template / never |
| S12 | Delimiter search through opaque template hole | No-match, not existential guesses |
| S13 | Unresolved earlier pattern followed by wildcard | Wait; do not select wildcard |
| S14 | Duplicate capture, pin referencing same-arm capture | Definition diagnostic |
| S15 | Closed string union in a template pattern | Domain diagnostic; explicit union loop required |
| S16 | Calc<"2 + 3 * 4">, Calc<"(2 + 3) * 4"> | 14, 20 |
| S17 | Calc<"8 - 3 - 2">, Calc<"-(10 - 4) / 3"> | 3, -2 |
| S18 | Calc<"1 / 0">, Calc<"1 / 2"> | Zero-divisor / nonintegral-division diagnostics |
| S19 | Empty/trailing-token/missing-parenthesis calculator input | Diagnostic with correct UTF-16 offset |
| S20 | Generated expression trees vs independent exact interpreter | Same values or selected exact-division failure; no transpiled-DSL proof substitute |

## 4. HTTP binding workload

Use [the full source](../examples/http-routes.typefn).

| ID | Case | Required outcome |
| --- | --- | --- |
| H01 | GET / and HEAD /health | Correct method/path and empty params |
| H02 | GET /users/:id<number>/posts/:slug | params {id:number;slug:string} |
| H03 | PATCH /flags/:enabled<boolean> | Boolean parameter |
| H04 | Duplicate names, empty name, 9id | Identifier/duplicate diagnostics |
| H05 | Lowercase method, doubled separator, relative path | Validation diagnostics |
| H06 | Trailing/empty segment, query/fragment, path whitespace | Validation diagnostics |
| H07 | Unknown parameter kind, wildcard, optional parameter | Unsupported grammar diagnostics |
| H08 | Broad/infinite-template HttpBinding input | Literal requirement diagnostic |
| H09 | Unresolved S, then specialize valid/invalid literal | Pending, then endpoint/failure |
| H10 | GET/POST union | Correlated method/route/params alternatives |
| H11 | Union containing one invalid binding | Retained failure, not filtering to never |
| H12 | HttpBindings<never> versus HttpBinding<never> | never versus scalar-domain diagnostic |
| H13 | Handler<"GET /users/:id<number>"> | Native callable Type with correct params and Promise<void> |

## 5. Reflection and construction

| ID | Case | Required outcome |
| --- | --- | --- |
| B01 | Required undefined vs optional value with exactOptional on/off | Declared/read/optional metadata retained distinctly |
| B02 | Numeric/string equivalent property keys | Duplicate-key diagnostic |
| B03 | Unique-symbol key construction/import/emit | Original identity retained |
| B04 | Conflicting property/string/number/template indexers | Native compatibility diagnostics, no overwrite |
| B05 | ReadonlyPublic on object with optional property/indexer | Correct readonly public structural projection |
| B06 | Public projection of private class | Deliberate identity loss, never claimed a nominal clone |
| B07 | Required tuple iteration with symbolic elements | Ready finite traversal |
| B08 | Optional/rest tuple iteration | Domain diagnostic; reflection remains available |
| B09 | Reflect/rebuild tuple labels/readonly/optional/rest | Metadata and native behavior retained |
| B10 | Multiple overloads | Order preserved; contextual inference/calls agree |
| B11 | Preserve generic signature / attempt to edit descriptor | Original handle works / explicit unsupported diagnostic |
| B12 | New function signature versus forwarded method signature | Ordinary variance versus preserved native origin |
| B13 | Invalid predicate target/result or constructor fields | Descriptor/native validation diagnostic |
| B16 | Guarded ListNode and mutually recursive group | Sealed recursive graphs, no extra nominal identity |
| B17 | Bare recursive Self, observe Self before sealing, cancel sealing | Diagnostic / diagnostic / discard entire group |
| B18 | Differing accessor getter/setter Types | Explicit reflection limitation, never silent collapse |

B14 and B15 are retired following deferral of enum-specific APIs. Other case IDs
remain unchanged.

## 6. Constructors and symbolic evaluation

Checked<T> returns Uppercase<T>; Discard<T> calls Checked<T> then returns string.
Envelope<T> returns `{tag:"text"; value:Checked<T>}`. Save captures before reassignment.

| ID | Case | Required outcome |
| --- | --- | --- |
| K01 | MapTypes<[string,number],Promise> | [Promise<string>,Promise<number>] |
| K02 | Compose<Promise,Array>, then apply | Promise<T[]> |
| K03 | Bounded native constructor applied to valid/invalid input | Original native bound checked at application |
| K04 | Dependent Pick bound through binary constructor slot | Correct actual-argument validation |
| K05 | Wrong arity, defaulted extra slot, constructor as native Type | Kind/domain diagnostic |
| K06 | Bare Array initializer; bare defaulted native alias | Missing argument diagnostic; ordinary default instantiation |
| K07 | Native interface/runtime-function F<_> parameter | Unsupported syntax/kind diagnostic |
| K08 | Pass Compose itself as F<_> | Wrong input/output kinds |
| K09 | Constructor factory capturing symbolic T without observing it | Ready known closure |
| K10 | Constructor factory branching on unknown T at native alias boundary | Constructor-not-ready diagnostic |
| K11 | Same selection followed by application inside Type-returning function | Deferred Type-valued call |
| K12 | Constructor creation calls invalid computation before return | Creation check retained; closure cannot hide it |
| K13 | `type F = Box;` for eligible named typefn, then F<T> | Known constructor without native generic-default ambiguity |
| V01 | Box<T>, keyof Box<T> | Known value field, "value" key |
| V02 | Envelope<T> then read tag and specialize T=number | Known tag cannot erase Checked failure |
| V03 | Discard<T>, erased native alias argument, union absorbed by any | Guard-only dependency survives each path |
| V04 | Measured<S> with text.length<S> field | Known object structure plus deferred scalar Type |
| V05 | Discarded text.length<S>, specialize S=number | Retained scalar-domain failure |
| V06 | Unknown condition | No arm executes; replay whole observing call |
| V07 | Known-false branch contains Checked<number> | No execution failure; static name/kind checks still apply |
| V08 | Native conditional unselected branch | Preserve native lazy checks; no eager branch evaluation |
| V09 | T→U→literal versus direct T→literal | Same native behavior and validation |
| V10 | Snapshot capture plus substitution and shadowing | Original binder identities and captured values retained |
| V11 | Reached helper check discarded by its caller | Check survives; no second helper ABI |
| V12 | any/unknown/never union traversal | Singleton/singleton/empty; none is an unresolved binder |
| V13 | Union with generic field or unresolved member | Wait; no unstable normalization order assumed |
| V14 | Native bound proof versus observation of constrained T | Validation may pass; value observation still waits |
| V15 | Opaque recursive native Type forwarded by identity function | No eager infinite graph traversal |

## 7. Native checking, emission and editor

| ID | Case | Required outcome |
| --- | --- | --- |
| I01 | Generic unwrap using Box<T>.value | Native structural access/inference preserved |
| I02 | Ask inference to invert arbitrary Calc<S> | No invented inverse candidates |
| I03 | NoInfer-bearing native Type forwarded/constructed | Barrier preserved |
| I04 | Generic function written with Checked<T> signature | Body checked under its explicit requirement |
| I05 | Specialize that signature with number and any argument | Check still fails |
| I06 | Generic body introduces new unproved Discard<T> local use | Definition diagnostic; no hidden caller effect |
| I07 | Assign partial generic signature to total string-returning signature | Reject dropped requirement |
| I08 | Reuse relation cached under different signature assumptions | Guard proof revalidated/matched |
| I09 | Source/declaration-only closure/capture/constructor consumption | Equivalent Types and validation in separate checkers |
| I10 | Hidden guard-only dependence emitted | Recipe preserved, not only visible string shape |
| I11 | Private helper used only in unchosen branch | Retained in emitted dependency closure |
| I12 | Unnameable export, generated-name conflict, nominal capture | Diagnostic / hygienic naming / original identity |
| I13 | Missing/unsupported/conflicting ABI with skipLibCheck | Diagnose before executing imported body |
| I14 | Module augmentation or imported code changes | New snapshot recomputes; producer cache not authoritative |
| I15 | Inferred local constructor captures native T | Lift source recipe with ordinary Type parameter |
| I16 | Type-only imports, re-exports, module cycles, TSX | Correct binding/erasure; no runtime imports or JSX in body |
| I17 | Hover pending Type; rename template capture; malformed edit | Honest pending display; correct scope; no recovered body execution |
| I18 | Cancel/change buffer during calculation or emit | No stale result/diagnostic/declaration publication |
| I19 | Emit-on-error with invalid DSL dependency closure | Withhold affected declaration atomically |
| I20 | Different use sites share failing recipe | Diagnostics point to each consumer and original operation |

## 8. Resource and performance acceptance

| ID | Workload | Required evidence |
| --- | --- | --- |
| R01 | Same active code/captures/arguments recur | Bounded cycle diagnostic, no successful cache entry |
| R02 | Many nested native operations in one VM call | Native work is metered/cancellable internally |
| R03 | Oversized string/bigint/native graph and infinite loop | Correct named resource diagnostic, never approximate Type |
| R04 | Unchanged blocked query repeated | No repeated replay without changed dependencies |
| R05 | Warm/cold cache and eviction with sufficient limits | Same successful semantics and guards |
| R06 | Failed/cancelled recursive builder | No unsealed handle/cache leak |
| R07 | Feature disabled/enabled on native-only projects | Native regression results and baseline cost measured separately |
| R08 | Long parser strings, deep tuples, union products, schema graph | Cold/warm wall time, allocations, peak memory, native/VM work and cache size recorded |

Before calling the implementation usable, all core cases above must have executable
fixtures; no skipped case counts as support. Before calling it robust, run malformed
input fuzzing, editor cancellation/incremental suites, native regression suites and
source/declaration equivalence. Before calling it performant, publish reproducible
measurements against pinned native encodings on the same machine/configuration.
Performance comparisons need equivalent behavior, including error/metadata contracts.
An algorithm-only JS harness is not a DSL parser/VM/compiler benchmark.

## 9. Community workload traceability

The [preserved investigation](https://github.com/unbyte/TypeScript/blob/7fe54fd975b95531092485468ac5677ed3a453fc/docs/typefn/research.md) sampled 39 files from 17 project families;
it did not port these libraries or run their full suites. The following maps
observed requirements to R2 facilities and deliberate boundaries. These are future
port/fixture obligations, not compatibility claims.

| Family / workload | R2 facility and required fixture boundary |
| --- | --- |
| type-fest MergeDeep, Jsonify, Words, Tagged | Property/read/optional/indexer metadata, explicit unions, string scans; serialization context and brand-loss policies must be implemented explicitly |
| ts-toolbelt merge/list/curry | Tuple modes, signatures, recursion, constructors; arbitrary new generic signature synthesis remains excluded |
| React Hook Form Path | Explicit visited state/depth, string construction, tuple/object reflection; algorithm truncation must differ from VM exhaustion |
| ts-pattern deep exclusion | Explicit correlated union products and structural construction; nonrepresentable set subtraction cannot become never |
| ArkType parsers | Cursor arithmetic, tagged result Types, strings, recursion and source offsets; completion/error policies remain parser-library work |
| gql.tada | Token/AST object Types, schema reflection, requiredness/nullability; parsing alone is not GraphQL selection semantics |
| type-challenges query string / JSON | Duplicate/order policy, nested parser state; JSON number/escape coverage must exceed the sampled intentionally restricted challenge before claiming JSON support |
| TanStack Router | String matching/scanning and explicit route policies; the included colon-route sample is not framework grammar parity |
| Kysely | Symbolic structure, property renaming, separate select/insert/update roles; generic callback fixtures required |
| Drizzle | Correlated join nullability and preserved generic signatures; nullable object versus nullable fields must remain distinct |
| tRPC | Separate parser input/output and context state; native generic-body requirements and inference direction tested |
| Zod | Input/output and variance-preserving native handles; scalar computation does not replace schema variance contracts |
| TypeBox | Recursive schemas and function/constructor reflection; existing native enum Types can be preserved, but enum-specific reflection/construction is deferred; dispatch order and generic/static parameters retained |
| Effect | Fixed-arity constructors, closures, separate Type/Encoded/context; full native TypeLambda variance/inference APIs are not replaced |
| XState | Correlated discriminants, key remapping, NoInfer preservation; no access to mutable inference sessions |
| hotscript | Explicit constructor application/composition replaces many encodings; passing higher-order combinators as first-class constructors remains excluded |
| fp-ts | Existing HKT registry encodings remain usable; direct runtime Functor<F<_>> and open-registry replacement are not R2 features |

For every port, pair successful concrete examples with negative, symbolic,
metadata/identity, declaration-only and resource cases. The aim is a precise,
readable computation language with demonstrated coverage, not a completeness
claim inferred from the number of features or pages in this specification.
