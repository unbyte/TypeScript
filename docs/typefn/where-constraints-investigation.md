# Should typefn have a `where` clause?

Investigated 2026-11-06 against repository revision `1c7be289` (compiler code at
`68838672`), installed TypeScript 6.0.3, and the rebuilt fork reporting
`7.1.0-dev`. **Research and design recommendation, not an adopted syntax extension.**
The formal specification is unchanged and the compiler does not parse `where`.

## Recommendation

Keep the existing `extends` bounds for now. Do not add the previously suggested
general `where A extends B` or arbitrary predicate facility in this milestone.
There is no evidence that the keyword itself is incompatible with TypeScript.
However, a general clause that both validates arguments and supplies facts inside
generic bodies would extend the checker substantially. Our deferred evaluator is
not yet a complete implementation of those contracts.

A narrow alternative remains viable: an optional layout for **the same parameter
bounds**, with exactly the same semantics as inline `extends`. Section 6 specifies
that candidate. It does not introduce lower bounds, new inference, or a theorem
prover. It offers readability rather than additional expressive power, so it is
not urgent enough to precede fixing the constraint-contract gaps found below.

This narrows the earlier conversational proposal. In particular, accepting
`where Animal extends T` and using arbitrary clauses as scoped facts cannot be
promised as a small extension of existing bounds.

## 1. What Rust actually supplies

In Rust, `T: Trait` restricts the trait implementations available for T. It is not
class inheritance. Bounds on an item impose obligations at uses, while the item's
body can use its declared bounds. Rust also supports bounds on compound Types and
associated Types; the original [where RFC](https://rust-lang.github.io/rfcs/0135-where.html)
motivated the syntax partly by those needs. The
[Reference](https://doc.rust-lang.org/reference/trait-bounds.html) defines the
current bound forms and checking rules. These are not arbitrary imperative Boolean
programs executed in source-list order.

Under the hood, the compiler passes a parameter environment containing the
assumptions available at that particular declaration. A goal consists of a
predicate and its environment. Solving can succeed, fail, or remain ambiguous;
implementations and existing assumptions supply candidates. See
[parameter environments](https://rustc-dev-guide.rust-lang.org/typing-parameter-envs.html)
and [trait solving](https://rustc-dev-guide.rust-lang.org/solve/trait-solving.html).
This explains why moving a bound out of angle brackets is easy *when the semantic
machinery already exists*. It does not show that we can omit that machinery.

Rust is also evidence that preserving assumptions is difficult, not a proof of
automatic safety. [Rust #25860](https://github.com/rust-lang/rust/issues/25860)
documents a soundness hole involving implied lifetime bounds and variance; its
current description distinguishes a fixed non-higher-ranked case from a remaining
case. The [compiler guide](https://rustc-dev-guide.rust-lang.org/traits/implied-bounds.html)
explains related limitations. Our DSL has no Rust borrow checker, so that exact
bug does not transfer. The relevant lesson is that assumptions must survive
function-type transformations and be discharged at the correct boundaries.

Niko Matsakis's [2014 design article](https://smallcultfollowing.com/babysteps/blog/2014/07/06/implied-bounds/)
explains the motivation for moving repeated obligations from users to the
compiler. The [Chalk model](https://rust-lang.github.io/chalk/book/clauses/implied_bounds.html)
explores how naive global implication rules can become circular. Both help explain
the design space; neither should be mistaken for the full current Rust contract.
No Rust compiler experiments were run in this investigation.

## 2. The TypeScript community has distinguished these features too

| Source | What the evidence establishes |
| --- | --- |
| [#42388: C#/Rust-style where](https://github.com/microsoft/TypeScript/issues/42388) | Open proposal covering readable bounds, local aliases, and partial inference. These are separable requests, not one approved feature. |
| [Maintainer discussion of lower bounds](https://github.com/microsoft/TypeScript/issues/42388#issuecomment-763835984) | Ryan Cavanaugh considered the possibility useful while questioning syntax and introduction of new parameters. |
| [Maintainer discussion of inference](https://github.com/microsoft/TypeScript/issues/42388#issuecomment-763837635) | He suggested clauses that do not influence inference. That would be a semantic choice, not ordinary bound relocation. |
| [#51385: conditional checks](https://github.com/microsoft/TypeScript/issues/51385) and [generic-composition objection](https://github.com/microsoft/TypeScript/issues/51385#issuecomment-1301233321) | A community proposal deliberately checks after inference. Discussion identifies why ordinary bounds cannot stand in for an unresolved predicate when generic callers compose. |
| [#14520: lower bounds](https://github.com/microsoft/TypeScript/issues/14520) | Open, in-discussion request for the opposite direction of an ordinary bound. Lack of implementation is not evidence of mathematical impossibility. |
| [#51011: circular constraints](https://github.com/microsoft/TypeScript/issues/51011) | Distinguishes structural self-reference accepted by TS from immediate cycles rejected by TS. A new syntax cannot safely erase that distinction. |
| [#35899: exact constraints](https://github.com/microsoft/TypeScript/issues/35899) | Middleware authors need more than an upper bound when supplying fields to a potentially narrower generic input. |
| [#51612: reverse mapped inference](https://github.com/microsoft/TypeScript/issues/51612) | A library author's examples show that adding known constraints interacts with inference of complex mapped structures. This is a motivation report, not a conformance theorem. |

Retrieved issue bodies and relevant discussion, including all 13 returned comments
for #42388, all 6 for #51385, and selected comments from the 51 returned for #14520.
The counts matched the retrieved metadata for those discussions. This is not an
exhaustive audit of every related issue or compiler prototype. The older
[#13995](https://github.com/microsoft/TypeScript/issues/13995) was also checked: it
closed in 2021 and must not be cited as proof that its original narrowing example
still fails today.

## 3. Experiments on the installed compilers

Reproduce from the repository root:

```sh
node docs/typefn/probes/check-where-baseline.mjs
python3 docs/typefn/probes/check-where-native.py built/local/tsgo
```

The stock probe checks 11 expected diagnostic positions/codes and 11 resolved
Type observations. The native probe checks the same native source and 13 separate
DSL projects, each in a fresh compiler process. Two native diagnostic codes differ
from TS 6: missing private identity and a missing constructor `prototype` report
2741 instead of 2344. The probes record those explicit differences. These are
observations of existing syntax, not tests of an implemented `where` language.

### Bounds already influence inference

This native example resolves differently in both tested compilers:

```ts
declare function plain<T>(x: T): [T];
declare function bounded<T extends string>(x: T): [T];

const a = plain("hello");   // [string]
const b = bounded("hello"); // ["hello"]
```

Consequently, these promises cannot both hold in general:

1. `where T extends B` is exactly equivalent to inline `T extends B`.
2. `where` never affects inference and runs only after inference is complete.

Typefn itself currently requires explicit arguments. Nevertheless, native runtime
function signatures can contain typefn applications whose arguments are inferred,
so the boundary still matters. The native checker also uses permissive and
restrictive instantiations for speculative relation checks. A VM must not treat
those provisional operands as final user arguments or commit their diagnostics.

### TypeScript assignability is not a general logical implication relation

The baseline accepts `string -> any` and `any -> number`, but rejects
`string -> number`. It also preserves method bivariance under strict checking.
These observations agree with the documented
[compatibility model](https://www.typescriptlang.org/docs/handbook/type-compatibility.html).
We must not add unrestricted transitive proof rules over native assignability,
infer arbitrary variance laws, or treat mutually assignable Types as identical.

The probe also accepts `Merged<string>` here:

```ts
type Merged<T extends any & number> = T;
```

The intersection normalizes to `any`. Separate obligations that T satisfy `any`
and T satisfy `number` would reject `string`. Therefore silently merging repeated
`where` clauses into one intersection is not a faithful conjunction operation.

Native `T extends string` accepts both `any` and `never`. In our DSL, that does
not make either a single numeric/string literal available to an arithmetic or
parsing operation. The native bound and the operation's input domain are separate.
Likewise, `N extends number` does not make `Twice<number>` executable arithmetic.

### Conditional Types are not a drop-in constraint evaluator

For `D<T> = T extends string ? true : false`, the probes observe:

| Argument | Result |
| --- | --- |
| `"text" | 1` | `boolean` |
| `any` | `boolean` |
| `never` | `never` |

Wrapping both sides in tuples stops distribution: the same union then produces
`false`, while `never` produces `true`. See the
[conditional-Type rules](https://www.typescriptlang.org/docs/handbook/2/conditional-types.html).
A bound should validate its whole argument, not filter its union constituents.
Future predicate contracts would need an explicit exact-`true` policy and separate
blocked/error states; `any`, `never`, and recovery Types cannot stand for proof.

Returning `never` on failure is also insufficient: the probe calls a function
whose conditional return becomes `never` without a diagnostic at that call.

### Subtype, supertype, ancestry, and equality are different contracts

`T extends Animal` lets a body use Animal's available structure but cannot safely
manufacture an arbitrary T from an Animal. T might require narrower properties.
`Animal extends T` is a lower-bound requirement: it can accept `unknown`, `object`,
or a structural supertype, so it does not give T all Animal members. The latter
direction could support useful input-acceptance contracts, but not the member-access
motivation for the former.

Unrelated classes with matching public members pass the tested native bound.
Private identity can prevent compatibility. Class instance Types and constructor
Types are different inputs. Neither structural direction is a reliable query for
the declared superclass chain. We should not add a third meaning to `extends`.

## 4. Current DSL behavior that matters to the design

The 13 DSL probes confirm ordinary dependent bounds, valid bounded callers,
specialization of initially unresolved bounds, retained checks from discarded
calls, native `any`/`never` bound behavior, and literal-readiness rejection.
They also expose three boundaries to settle before broadening the contract.

### 4.1 Pending bounds make diagnostic timing depend on parameter order

Both programs below impose an impossible concrete number bound. At the current
revision, they do not report it at the same stage:

```ts
typefn A<T extends string, U extends number> { return 0; }
type PendingA<X> = A<X, string>; // No diagnostic yet: first bound blocks.

typefn B<U extends number, T extends string> { return 0; }
type PendingB<X> = B<string, X>; // Immediate constraint diagnostic.
```

`typefnVM.constraints` visits parameters in order and returns at the first blocked
bound. PendingA retains a call to complete later; this observation does not show
that a finished invalid specialization succeeds. It does show that an arbitrary
list of clauses cannot be described as an order-independent eager validator by
simply reusing the current loop.

An implementation could examine every independently ready clause before deferring,
but that changes diagnostic timing and work. It must be chosen explicitly rather
than introduced accidentally by the order of AST traversal. This investigation
records the existing behavior without changing it.

### 4.2 Header computations can execute while checking an unused declaration

```ts
typefn Twice<N> { return N * 2; }
typefn Unused<T extends Twice<"bad">> { return 0; }
```

The arithmetic already diagnoses, without any call to Unused. A native bound
such as `Uppercase<number>` similarly diagnoses in an unused declaration. In
contrast, a bad DSL application inside an unexecuted body branch is not executed.

The earlier statement that a declaration does not execute its *body* remains true,
but it must not be stretched into a claim that no computation in its header can
run. `compileTypefn` checks bound ASTs through the native checker. A native alias
can hide a typefn application, so merely banning an obvious call in a future
`where` RHS would not settle this behavior.

### 4.3 Bound scope differs between implementation and specification

The fork currently accepts `typefn Select<K extends keyof T, T>`, matching native
TypeScript's later-parameter references. The R2 semantics describe bounds in an
environment of preceding parameters. Native structural F-bounds and immediate
circular bounds also need to be distinguished; the stock/native baseline accepts
the former and rejects the latter. This research does not claim the DSL's complete
recursive-bound behavior matches native TypeScript.

The scope and timing discrepancy must be resolved explicitly before adding an
alternate surface spelling and documenting it as a stable contract.

## 5. Why a general `where` is a larger feature

There are three separate designs:

| Design | What it buys | Main consequence |
| --- | --- | --- |
| Relocate existing upper bounds | Shorter parameter lists | Same native bound behavior, including its inference interactions. |
| Validate arbitrary relations after arguments settle | Lower bounds and compound checks | New obligations; no automatic new facts for the body or inference. |
| Use arbitrary clauses as body assumptions | Generic composition under richer contracts | Scoped proof environments, proof dependencies, binder-safe substitution, cache changes, and rules for callable compatibility. |

The last two cannot be implemented faithfully as a Boolean `if` at the start of
a VM function. A useful design would have to answer all of the following:

* **Caller versus callee:** the caller must prove the callee's requirement using
  its own environment. Installing the callee's clause as an assumption before
  checking that same call would make every contract prove itself.
* **Unknown versus false:** a generic alias can retain a partial calculation.
  A native runtime body checked once cannot silently acquire new hidden caller
  requirements. It needs a written contract or a proof. Different predicate
  programs cannot be equated just because examples produce the same answer.
* **Inference:** a post-inference check must not rewrite inferred Types to satisfy
  itself, search alternate substitutions, or invert VM calculations. Moving a
  general clause into a native TypeParameter constraint would violate that rule.
* **Ordering:** independent clauses are logically conjunctive, but their Type
  expressions can be blocked, invalid, recursively dependent, or resource-limited.
  A bound needed to make `T[K]` well-formed is not a license to inspect unknown T.
  Reordering clauses must not change a completed result with sufficient resources;
  first diagnostic and cost can differ. Blocked-call timing needs its own rule.
* **Cycles:** native recursive structural comparisons can make provisional
  coinductive assumptions. That is not permission to accept a VM contract that
  calls itself while trying to establish its own validity. Termination, incomplete
  knowledge, native recursion, and failed proof require distinct handling.
* **Erasure and composition:** obligations must survive unused results, native
  normalization, callable assignment, constructor adaptation, and declaration
  serialization. No rule may turn a stricter callable into an unconditional one
  merely because its visible result Type is the same.
* **Caching:** a proof valid under one declaration's assumptions is not globally
  valid for the same two Type handles. Cache entries must include the relevant
  environment/mode or carry explicit dependencies. Ordinary errors, blocked
  queries, and canceled work cannot become unconditional cached answers.
* **Resources and editing:** contract evaluation needs shared VM/native budgets,
  cancellation, stable diagnostic provenance, and invalidation when a bound or
  its dependency changes. The current native resource accounting is incomplete.

These are grounded in the implementation, not just hypothetical concerns:

| Code | Observation |
| --- | --- |
| `internal/checker/typefn.go`: `compileTypefn`, `constraints`, `call` | Native bound checking during lowering; ordered early exit during calls; pending calls and journals. |
| `internal/checker/typefn_deferred.go` | Immutable pending terms and retained requirements, but not a general predicate entailment engine. |
| `internal/checker/checker.go`: `getRelationKey` | Relation keys encode Type identities/constraint structure and relation state, not an arbitrary new where-assumption environment. |
| `internal/checker/relater.go`: `recursiveTypeRelatedTo` | Cached and provisional recursive results need careful interpretation; native ternary states are not our full requirement protocol. |
| `internal/checker/inference.go`: `getInferredType` | Native constraints participate in candidate validation/fallback. |
| `internal/checker/checker.go`: `getReturnTypeOfSignature` | The recent fix for cached deferred failures illustrates why reusing native caches requires tests at multiple use sites. |

The existing [integration specification](spec/integration.md) already requires much
of this discipline. The [implemented deferred boundary](deferred-evaluation.md)
explicitly does not claim full scoped proof support. Adding `where` would not fill
that gap by itself.

## 6. A constrained candidate worth keeping

If the layout is useful later, start with this **unimplemented candidate**:

```ts
typefn PickValue<T, K>
where
  T extends object,
  K extends keyof T
{
  return T[K];
}
```

It must mean exactly the same thing as:

```ts
typefn PickValue<T extends object, K extends keyof T> {
  return T[K];
}
```

Candidate rules:

1. `where` is contextual only after a typefn header (after an optional result-kind
   annotation), before the body. It does not extend native function/class syntax.
2. A clause is `ParameterName extends BoundType`. The left side must identify one
   ordinary Type parameter of this declaration. It cannot be `Animal`, `T[K]`,
   a tuple expression, or a constructor slot. No implicit parameters or local
   aliases are introduced; typefn already has local `type` bindings.
3. Each parameter has at most one bound across both inline and where forms.
   Duplicate bounds diagnose. Authors may explicitly write an intersection,
   receiving native intersection semantics rather than a new conjunction rule.
4. The compiler associates clauses with parameter identities and checks them in
   parameter-declaration order. Reordering the written where list cannot change
   behavior. Parameter order still determines positional application. The current
   early-block behavior is inherited unless separately revised for both spellings.
5. Bound grammar, lexical scope, header checking, and replay must be identical
   between spellings. For a first R2-aligned fragment, enforce preceding-parameter
   references in both, with ordinary native Type grammar as the current execution
   boundary. Later/self references need a separately accepted scope decision;
   do not silently claim equivalence while one form admits more than the other.
   Enforcing preceding-only scope would reject some currently accepted files;
   that compatibility change must be recorded separately, not hidden in a parser edit.
6. Arbitrary Boolean predicates, DSL arithmetic as header syntax, lower bounds,
   equality clauses, implication rules, trait implementations, and where-based
   overload selection are outside this candidate. Native aliases in bounds retain
   the same semantics as inline bounds, including any computations they trigger;
   this is not a claim that bounds are computation-free.
7. A native argument that meets a bound still needs to satisfy each executed
   operation's domain. `any`/`never` retain native bound behavior. No new member,
   literal value, or successful-result overlay is invented from a constraint.

This candidate deliberately does **not** satisfy the general lower-bound request.
For the common need to read members of an input, an ordinary upper bound already
does the useful part. If accepting a supertype is an actual algorithm requirement,
we should revisit a dedicated validation-contract design rather than pretend the
layout feature solves it.

## 7. Adoption gates and deferred alternatives

Before even the layout candidate is implemented:

* Settle bound scope and declaration/application timing in the existing spec and
  implementation; retain the probes above as review evidence.
* Require inline/where equivalence tests for direct calls, partial substitution,
  discarded results, native signature specialization, inference, and failure
  diagnostics. Include reversed where-list order and fixed parameter order.
* Give clauses real AST/source spans. Do not move their nodes into the angle-list
  source ranges or duplicate AST children to simulate sugar. Binding, formatting,
  navigation, semantic tokens, serializer/factory/visitor support, and source
  recipes must see the actual written structure.
* Exercise invalid/recovered clauses, duplicate/unknown parameters, cycles,
  `any`/`unknown`/`never`, mixed unions, private identities, constructor Types,
  repeated invalid uses, cancellation, incremental edits, and shared budgets.

A general validation-only extension could eventually support whole-Type
assignability goals such as `Animal extends T`, without inventing new inference
candidates. A stronger assumption-bearing extension would additionally need the
proof and cache architecture in section 5. They deserve separate proposals and
implementation gates. Neither is declared part of version 1 by this document.

**Decision for this investigation:** retain `extends`, defer the new syntax, and
keep the narrow layout candidate documented. The immediate engineering priority
is making the existing constraint contract consistent, not enlarging its syntax.
