> Investigation/history, not a current language contract. The [R2 specification](spec/README.md) supersedes earlier design recommendations; the [manual](manual.md) uses the current syntax.

# Why TypeScript has no native higher-kinded types: constraints for R1

Investigated 2026-11-06. This is a review of primary discussions, current retrieved
issue status, the pinned TypeScript-Go checker and reproducible stock TypeScript
behavior. It is not an implementation of the proposed DSL. Findings below separate
maintainer statements from community proposals and our design conclusions.

## What the maintainers actually said

The main request, [TypeScript #1213](https://github.com/microsoft/TypeScript/issues/1213),
was opened in 2014. At retrieval it remained open with Suggestion and Help Wanted
labels. Ryan Cavanaugh's [response](https://github.com/microsoft/TypeScript/issues/1213#issuecomment-96854288)
was **“Quasi-approved.”** He expressed interest and asked for a working implementation
to assess implications and edge cases. A [later response](https://github.com/microsoft/TypeScript/issues/1213#issuecomment-190325390)
said the status and priority had not changed. These historical comments are not
a current roadmap commitment, but they contradict a claim that the team rejected
HKTs as impossible or fundamentally incompatible with TypeScript.

There is no single definitive rejection rationale in the evidence reviewed here.
The defensible conclusion is that an interested but unresolved feature still needs
an acceptable design and implementation. Technical difficulties below explain what
that entails; they are not a claim to know the team's internal prioritization.

| Source | Authority and relevance |
| --- | --- |
| [#1213 maintainer response](https://github.com/microsoft/TypeScript/issues/1213#issuecomment-96854288) | Maintainer position: interested, needs an implementation to evaluate |
| [Artazor's implementation exploration](https://github.com/microsoft/TypeScript/issues/1213#issuecomment-275222845) | Community analysis: arity is insufficient; nested kinds, scopes and bounds matter |
| [GregRos on result constraints](https://github.com/microsoft/TypeScript/issues/1213#issuecomment-370812753) | Community analysis: promise-like constructors need an input/output relationship |
| [Proposal #55280](https://github.com/microsoft/TypeScript/issues/55280) | Community proposal exploring structural constructors, constraints and inference; not an adopted spec |
| [Higher-order inference PR #30215](https://github.com/microsoft/TypeScript/pull/30215) | Official implementation rationale for bounded, directional generic-function inference |
| [Instantiation expressions PR #47607](https://github.com/microsoft/TypeScript/pull/47607) | Official semantics of specializing generic runtime values; distinct from constructor abstraction |
| [Call-types clarification](https://github.com/microsoft/TypeScript/issues/52035#issuecomment-1402795342) | Maintainer distinction between HKTs and repeated inference for hypothetical function calls |
| [TypeScript FAQ](https://github.com/microsoft/TypeScript/wiki/FAQ#structural-vs-instantiation-based-inference) | Official warning that structural versus instantiation-based inference is implementation-dependent |
| [Design goals](https://github.com/microsoft/TypeScript/wiki/TypeScript-Design-Goals) | General compatibility, erasure, structural typing and usability constraints, not an HKT rejection |

Coverage limitation: the #1213 endpoint returned 187 comments while issue metadata
reported 189. Relevant maintainer and technical comments were examined; this is
not an assertion that every comment, prototype or design meeting was audited.
The repository source baseline is `89d5d5b2849a0db0957065889ca58536fa6d2e4a`.

## 1. TypeScript generics are not uniformly invertible constructors

Consider native aliases:

```ts
type Box<T> = { value: T };
type OtherBox<T> = { value: T };
type Constant<T> = string;
type Branch<T> = T extends string ? number : boolean;
```

Box<string> and OtherBox<string> are structurally interchangeable. Constant<number>
and Constant<boolean> both produce string. Knowing an output therefore cannot
uniquely identify either its constructor or its arguments. A TypeScript alias can
compute, discard arguments or disappear into a primitive/union; its name is not
a universal nominal tag. Native private/protected member identities and unique
symbols remain important exceptions to purely structural treatment.

Consequently an unknown equation F<A> = string has many solutions. Trying candidate
declarations, consulting the printed name or reversing arbitrary VM code is not
a sound inference policy. This does **not** make explicit F<A> application hard
in the same way: with F known, forward evaluation is well-defined, subject to the
usual resource limits.

Nor does support for HKTs require unrestricted higher-order unification. Even
[Haskell 2010 section 4.2.2](https://www.haskell.org/onlinereport/haskell2010/haskellch4.html)
restricts partial application of parameterized type synonyms, despite supporting
higher kinds. A language's constructor fragment and inference policy matter.
Our proposal accepts computational native aliases and imperative closures, so
analogies to simpler generative constructors cannot justify stronger inversion.

**R1 decision:** explicit constructor arguments are the baseline. No injectivity,
automatic inverse computation or recovery of constructors from resulting structure.

## 2. Structural inference and preserved generic arguments are different evidence

The FAQ explains that the checker may infer from members or use generic argument
metadata. The second is often faster, but the selected strategy can change. The
Go checker visibly has a same-alias fast path at
[inference.go](../../internal/checker/inference.go), `inferFromTypes`: it infers from
the original alias arguments using variance information when the alias matches.
That fast path does not establish a universal provenance contract.

The included TypeScript 6.0.3 probes observe:

```ts
type Empty<T> = {};
declare function extract<T>(x: Empty<T>): T;
declare const named: Empty<number>;

const a = extract(named); // number in the measured compiler
const b = extract({});    // unknown in the measured compiler
```

Both argument types have the same empty structure. Their different inference
results demonstrate why structural compatibility is not sufficient evidence of
a constructor argument. These observations are version-specific, not a new
language guarantee.

**R1 correction:** the earlier claim that Functor<Array> would reliably infer F
was too strong. Copying a deliberately preserved constructor slot is a plausible
feature, but requires a specified evidence-preservation mechanism. Mapped copies,
unused parameters, aliases, .d.ts round trips and overloads must be covered first.
An annotation is not, by itself, proof that the native fast path will preserve it.

## 3. Arity, input bounds and output guarantees are separate

`F<_>` establishes Type → Type. It does not establish that every Type is accepted
by every unary generic declaration:

```ts
type StringBox<T extends string> = { text: T };
```

StringBox cannot satisfy a contract that may apply F<number>. Dependent bounds
such as `K extends keyof T` add binder-scoping and substitution obligations.
Higher-order constructor arguments also require recursively structured kinds;
counting parameter slots alone loses their categories.

There is a second, previously underdeveloped issue. A generic implementation
receiving a value of Type F<T> cannot call .then without an output guarantee.
The desired guarantee might be:

```text
For every T in F's permitted domain: F<T> is assignable to PromiseLike<T>.
```

This must distinguish Promise<T> from:

```ts
interface Shifted<T> extends Promise<T[]> {}
```

Both have one parameter; the second promises an array of that parameter. Checking
one instantiation, especially any, never or unknown, does not establish the
universal relationship. Imperative type functions can branch on an argument and
invalidate any finite sample-based assumption.

Even that result bound would only justify PromiseLike operations. It would not
prove that .then returns F<U> for the same F: a subclass may inherit a method
returning a base promise. Preserving F requires an explicitly supplied operation
or an additional contract. Our Functor<F> dictionary provides such an operation
for map without claiming every constructor intrinsically supports mapping.

**R1 correction:** output contracts are a real open requirement. Input bounds are
already proposed; a finished output-bound design needs syntax, binder scope,
conservative symbolic verification, closure propagation and declaration emit.
No syntax is added casually in this audit. Useful unbounded composition remains
possible, but general bounded native HKT support is not fully specified yet.

## 4. Subtyping and variance do not follow from kind equality

For Branch above, "hello" is assignable to unknown, but Branch<"hello"> is number
and Branch<unknown> is boolean. Therefore an arbitrary unary type function is
not covariant. Function-parameter positions also introduce contravariance under
strict function checking; method bivariance remains a native compatibility rule.

Three relations must be kept distinct:

* **Kind compatibility:** can this constructor accept the promised arguments?
* **Internal identity:** is this the same declaration or captured closure for caching?
* **Native assignability:** are the Types produced by applications compatible?

Proposed `Keeps<F<_>> = { value: F<string> }` must distinguish Array from Promise
instantiations. Empty `Phantom<F<_>> = {}` must remain structurally compatible
across constructors. Making every enclosing type nominal to preserve F breaks
ordinary structural behavior; treating all same-kind constructors as interchangeable
loses real member differences.

The native `getVariancesWorker` in
[relater.go](../../internal/checker/relater.go) substitutes ordinary marker Types
with known subtype relations. Such a marker cannot stand in for F and then be
applied to string. Supporting constructors changes variance inference and the
relation fast paths, not just parameter storage.

**R1 correction:** conservative treatment of unknown F<T> is necessary but not
a complete relation algorithm. Constructor-aware variance or structural fallbacks,
recursive comparisons and blocked relation caches need explicit design. Existing
Type-kind probes must not be applied to constructor-kind slots.

## 5. Higher-order functions, HKTs and call types are different capabilities

```ts
type Poly = <T>(x: T) => { value: T };
```

Poly is an ordinary Type with a generic call signature. `Poly<string>` is not
legal native application. In the measured compiler ReturnType<Poly> yields
`{value:unknown}` through its Box alias, whereas specializing a known generic
function with `typeof makeBox<string>` preserves Box<string>.

An overloaded function provides another distinction: extracting ReturnType uses
the last signature, while a concrete call resolves an applicable overload. A
request to infer the result of a hypothetical call needs call-resolution semantics,
not merely a first-class Type → Type value. Ryan Cavanaugh explicitly identified
this distinction in #52035; his
[follow-up example](https://github.com/microsoft/TypeScript/issues/52035#issuecomment-1404393905)
explains why repeated inference can be required.

PR #30215 documents deliberately directional higher-order generic-function
inference with restrictions on the resulting call signature. It does not promise
complete unification. Its rationale explicitly values behavior while editing.
PR #47607 specializes generic value signatures and handles unresolved generic
values via their constraints. That operation cannot simply be reused to replace
an unknown constructor with its constraint without losing its identity/dependency.

**R1 decision:** do not automatically turn generic callable Types into constructors,
promise higher-rank inference, or advertise a call-types feature as a consequence
of the VM. Explicit adapters may express specific transformations; general
signature application requires its own semantics and tests.

## 6. Defaults expose an actual compatibility bug in our shorthand

The earlier draft lifted a bare required generic in an initializer into a
constructor, but treated a bare defaulted generic as its native default instance.
Then this upstream library edit would silently change the kind of a client binding:

```ts
// Before: type Source<T> = { value: T };
// After:  type Source<T = unknown> = { value: T };
type Local = Source;
```

**R1 fix:** a missing native argument is a diagnostic, not a signal to infer a
different kind. Native generics are captured unapplied in an expected constructor
slot, such as Twice<Array,string> or assignment to an established constructor
binding. Otherwise native arity/default rules apply. A constructor literal or a
constructor-valued application establishes a new binding's kind explicitly:

```ts
type F = typefn<T> { return Array<T>; };
type AsyncArray = Compose<Promise, Array>;
```

This preserves direct native constructors at call sites without overloading an
existing valid Type expression based on whether a dependency happens to add defaults.

## 7. A VM inside the checker helps evaluation, not every checker obligation

The Go baseline's `getTypeFromTypeAliasReference` and `getTypeAliasInstantiation`
in [checker.go](../../internal/checker/checker.go) instantiate known declarations.
The [TypeMapper](../../internal/checker/mapper.go) maps native Types to Types;
alias/reference argument vectors and caches likewise assume Type arguments.
Constructor handles and typed deferred applications need a coherent representation
across those paths. An opaque fake any/unknown would erase needed information.

For example, `keyof F<T>` must remain dependent while F is unknown. A concrete
substitution may later reveal members. Blocking observation is different from
rejecting a generic body or guessing its output. Native conditional aliases must
also retain their distribution semantics when invoked; imperative branches do
not acquire implicit distribution merely because their inputs are unions.

Performance risks include repeated normalization, branching over unions, closure
capture growth, relation cycles and rechecking partially edited code. Caches need
kind-aware arguments, closure identity, checker-snapshot ownership and cancellation
cleanup. Shared budgets must cover normalization and native operations, not just
VM instruction dispatch. The actual overhead must be measured; this investigation
does not supply benchmark evidence.

The personal fork removes the need to obtain upstream syntax approval. It does
not remove declaration portability, native project regression, language-server
latency or substitution correctness. The HKT portion is a checker extension;
embedding a type interpreter is useful infrastructure rather than a solution to
all these obligations.

## Verification and resulting scope

Run from the repository root:

```sh
node docs/typefn/probes/check-hkt-baseline.mjs
```

With installed stock TypeScript **6.0.3**, strict and exact optional property
checking, the probe validates 11 inferred/reduced Type observations and six
expected diagnostic sites. It also checks structural alias assignments, a matching
shifted-promise result and private-member incompatibility. The first exploratory
run found that named Empty<number> preserves number; the recorded expectation
was corrected to that measured behavior. This is evidence about native behavior,
not a passing DSL suite, a checker benchmark or exhaustive coverage of inference.

Updated [design](design.md), [HKT proposal](higher-kinded-types.md) and
[acceptance cases](acceptance.md) now distinguish:

| Area | Status after audit |
| --- | --- |
| Explicit constructor passing, closures, forward application | Proposed core retained |
| Bare generic/default ambiguity | Rule corrected; failed arity checks never infer a new kind |
| Input domain/kind checking | Proposed; requires conservative symbolic checking |
| Inferring F from native argument metadata | No baseline guarantee; evidence contract required |
| Output bounds and same-constructor operations | Missing design work explicitly identified |
| Native relations/variance involving constructor parameters | Incomplete; checker-wide design required |
| Generic callable application and higher-rank inference | Separate capabilities, not implied |
| Compiler/LS implementation and performance | Not validated by these probes |

The next design work should resolve output contracts and native relations with
small positive/negative examples before adding syntax or implementation. This
audit found substantive gaps in R1; it is not evidence that the language is mature.
