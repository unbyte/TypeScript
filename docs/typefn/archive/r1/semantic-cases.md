> Historical R1 material, superseded by the [R2 specification](../../spec/README.md). Not the current language contract.

# Evaluation and declaration acceptance cases

Design oracles for [evaluation](evaluation.md) and [declaration emit](declaration-emit.md).
These are not passing compiler tests. “Pending” denotes internal dependency/check
state, not any/unknown/never and not a new surface Type. “Shape” does not erase
associated checks. Each failure case must retain a source/call trace.

Use separate fixture programs for cases with intentional errors. A poisoned
recovery Type from one case must not make another appear successful. Run eventual
tests both from source and declarations where applicable.

## Shared definitions

```ts
typefn Box<T> { return { value: T }; }
typefn Apply<F<_>, T> { return F<T>; }
typefn Checked<T> { return Uppercase<T>; }
type Erase<X> = string;
type NativeStrings<T> = T extends string ? T : never;

typefn Envelope<T> {
  return { tag: "text"; value: Checked<T> };
}

typefn Discard<T> {
  type unused = Checked<T>;
  return string;
}

typefn Read<T> {
  const text = literal.text(Checked<T>);
  return literal(text);
}

typefn Select<T> {
  type F = typefn<X> { return Array<X>; };
  if (T is string) { F = Promise; }
  return F<T>;
}

typefn Choose<T>: typefn<_> {
  if (T is string) { return Promise; }
  return Array;
}

typefn UseChoice<T> {
  type F = Choose<T>;
  return F<T>;
}

typefn Save<T> {
  type current = T;
  type saved = typefn<X> { return [current, X]; };
  current = number;
  return saved;
}
```

T/U below are unresolved native Type parameters unless stated otherwise. A later
substitution means instantiating the same stored symbolic result, not merely running
an unrelated closed call with a similar printed name.

## Publication and observations

| ID | Setup/action | Required observation |
| --- | --- | --- |
| P01 | Box<T>; then keyof Box<T> | `{value:T}` and `"value"`, with no invented concrete argument |
| P02 | Envelope<T> | Known `tag`/`value` fields; value and validity check pending; no placeholder any |
| P03 | `type Tag<T> = Envelope<T>["tag"]`; specialize with number | Visible shape `"text"` does not suppress the Checked<number> error |
| P04 | Checked<T> where T extends string | Native Uppercase<T> after bound validation; no concrete string guessed |
| P05 | Read<T> with that same string bound | Still pending: a valid string Type is not readable literal text |
| P06 | Read<"hello"> | `"HELLO"` |
| P07 | Select<T>, including T extends string | Control observation waits; a constraint is not the observed argument |
| P08 | Select<"x">; Select<number> | Promise<"x">; number[] respectively |
| P09 | Same-handle `T == T` inside a typefn | true; retain any checks carried by that handle |
| P10 | Shallow reflect of `{value:T}` | Field metadata ready; value Type remains T |
| P11 | Required tuple traversal of `[T,U]` | Exactly two symbolic Type elements in order |
| P12 | Union traversal of `"a" | "b"` | Two literal members, no ordering guarantee |
| P13 | Union traversal of `"a" | T`, or objects with generic field Types | Wait under the conservative symbolic-union policy |
| P14 | Traversal of never / any / unknown | Zero / one / one members; none is treated as an unresolved binder |

## Bounds, erased work and control flow

| ID | Setup/action | Required observation |
| --- | --- | --- |
| V01 | Apply<Uppercase,"hello"> | `"HELLO"` |
| V02 | Apply<Uppercase,number> or Apply<Uppercase,unknown> | Native constraint failure with supplied-constructor and call locations |
| V03 | Apply<Uppercase,T>, unconstrained T | Pending; neither success with any nor a guessed negative answer |
| V04 | `type K<T> = Apply<Uppercase,T>`; instantiate K<number> | Failure after substitution, despite legal deferred alias definition |
| V05 | `type Ignore<S extends string> = 0`; Apply<Ignore,number> | Failure before erased native body can produce 0 |
| V06 | Discard<T>, then T→"hello" / T→number | string / failure; the unused local check survives |
| V07 | `type Hidden<T> = Erase<Discard<T>>`; Hidden<number> | Failure; erased native Type arguments retain DSL checks |
| V08 | `type Absorbed<T> = any | Checked<T>`; Absorbed<number> | Failure; normalization to an any shape retains the guard |
| V09 | Reassign the `unused` binding in Discard before return | Same result as V06; checks belong to execution, not just the final slot |
| V10 | Call Checked<number> only inside a known-false DSL branch | No execution-time bound failure from that branch |
| V11 | An unknown name or wrong kind in that same branch | Static definition diagnostic |
| V12 | Unknown condition followed by failing statements in either arm | Suspend before either arm executes |
| V13 | Unused `const text = literal.text(T)` before returning string | Whole call waits; an observation cannot disappear as dead local data |
| V14 | Pass Pick through F<_,_>; use keys inside/outside keyof actual T | Valid native Pick / original dependent-bound failure |
| V15 | Apply<NativeStrings,"x"|1> | `"x"`; preserve native conditional distribution |
| V16 | A native conditional selects a branch without a deferred DSL call | Do not force the unselected branch's pending DSL checks; existing native static errors remain errors |
| V17 | A local data helper calls Checked<T>, discards its Type handle and returns integer 1 | The helper's caller inherits the check; specialization with number fails |

V10 deliberately concerns a DSL if statement, not an assertion that stock
TypeScript skips static checking of native conditional branches. V16 requires
checks to follow the native operator's selected dependency, rather than collecting
both branches' future checks unconditionally.

## Replay, constructors and captures

| ID | Setup/action | Required observation |
| --- | --- | --- |
| R01 | Evaluate Read<T>; substitute T→U, then U→"x" | Same result as direct Read<"x">; no stale pending state |
| R02 | Evaluate Discard<T> once; specialize separately with string and number | First valid, second invalid; neither specialization mutates the shared generic entry |
| R03 | `type S = Save<string>`; S<boolean> | `[string,boolean]`; later reassignment of current does not alter capture |
| R04 | Save<T> captures symbolic T; specialize that environment to string | Applying returned closure matches R03; local shadowing named T is unaffected |
| R05 | UseChoice<T> | One Type-valued deferred call; no abstract constructor escapes |
| R06 | UseChoice<"x"> and UseChoice<number> | Promise<"x"> and number[] |
| R07 | Inside native `function f<T>()`, declare `type C = Choose<T>` | Constructor-not-ready diagnostic; an unchosen factory cannot publish C |
| R08 | Factory returns `typefn<X> { return Uppercase<T>; }`, instantiated with T=number | Creating closure succeeds; applying it fails; body checks occur on invocation |
| R09 | Return Type on one path and constructor on another, even unreachable | Definition-time return-kind mismatch |
| R10 | Forward an existing checked Type through an identity typefn | Preserve its guards and Type identity; forwarding must not complete it by fiat |
| R11 | Two different constructors produce structurally equal objects | Native structural compatibility remains; cache identity does not become output nominality |

## Native generic contexts

Use these signatures independently:

```ts
function identity<T>(x: Checked<T>): Checked<T> { return x; }
function bounded<T extends string>(x: Checked<T>): Checked<T> { return x; }

function local<T>(x: any) {
  const y: Discard<T> = x;
  return y;
}

declare function partial<T>(x: T): Discard<T>;
type Total = <T>(x: T) => string;
```

| ID | Setup/action | Required observation |
| --- | --- | --- |
| N01 | Check identity's body | Identity relation under the explicit signature requirement; no per-call body rechecking |
| N02 | Specialize identity<number>, supplying an any argument | Bound failure during signature specialization; any argument does not suppress the check |
| N03 | Check bounded's body | Native bound establishes validity; ordinary native checking applies |
| N04 | Check local's body | Definition error for the new unproved check, even though any is assignable to its string shape; do not infer an invisible input contract |
| N05 | Assign partial to Total | Reject dropping its unresolved requirement for arbitrary T; visible string shape alone is insufficient |
| N06 | Valid closed specialization of partial compared with native string-returning signature | Ordinary native relation after checks complete |
| N07 | Generic alias definition containing pending DSL work, never specialized | Keep recipe; no invalid input is invented just to execute it |
| N08 | Definite native use asks for a member/relation of unknown result shape | Focused unresolved-computation diagnostic, not missing-members-of-unknown noise |

These rules operate on statically known function Types. Erasing a callable Type
through native any/casts still has native TypeScript's ordinary escape behavior;
the DSL does not insert runtime checks or reconstruct an erased signature.

## Declaration-only consumption

| ID | Producer/consumer scenario | Required observation |
| --- | --- | --- |
| E01 | Emit/import WithError and a constructor alias | Consumer applies the constructor with the same ordinary Types |
| E02 | Emit Save/S from R03 and consume with producer sources removed | Snapshot capture still yields `[string,boolean]` |
| E03 | Emit inferred result using local constructor alias with outer T | Private Type-returning helper binds T hygienically; no native F parameter or chained application |
| E04 | Emit Hidden/Absorbed from V07/V08 | Consumer number specialization still fails; shape-only emit is forbidden |
| E05 | Referenced helper appears only in a branch not used by producer | Helper remains in dependency closure for other consumer arguments |
| E06 | Class-private or unique-symbol Type captured through import | Preserve original symbol identity and module resolution |
| E07 | Captured dependency cannot be named/reconstructed in public declaration | Explicit nameability diagnostic; no structural clone or any fallback |
| E08 | Missing/conflicting/unsupported ABI directive | Diagnostic before executing imported DSL, including with skipLibCheck |
| E09 | Module augmentation changes between two consumer programs | Rebind/recompute per program; do not reuse producer native Type pointers |
| E10 | Compare source/declaration redirects with declaration-only consumption | Same intended source identities; absence of producer source actually enforced in declaration-only fixture |

## Failure, resource and incremental behavior

| ID | Scenario | Required observation |
| --- | --- | --- |
| X01 | Retry an unchanged blocked call repeatedly through editor queries | Reuse pending dependency state without reevaluating the same frame indefinitely |
| X02 | Previously blocked relation becomes valid after substitution | No permanently cached false result |
| X03 | Same active call and arguments recursively reentered | Cycle diagnostic with bounded trace; no success/never cache entry |
| X04 | Native work exhausts a resource limit within one VM instruction | Resource diagnostic; fuel-only accounting is insufficient |
| X05 | Cancel during builder/native validation, then retry | No unsealed graph/cache poison; fresh request may complete |
| X06 | Change captured/native dependency in a new checker snapshot | Old handles, pending checks and diagnostics do not enter new results |
| X07 | Same failure reached from two user applications after cache reuse | Each diagnostic points to its own use, with the original failing operation related |
| X08 | Emit-on-error with invalid DSL dependency | Withhold affected declaration file atomically; no truncated executable body |
| X09 | Repeated warm/cold evaluation with sufficient resources | Same Type and semantic validity; work/resource counters may differ |
| X10 | Forward an opaque recursively expanding native interface Type through an identity typefn | Do not eagerly traverse all latent members or force an infinite native graph |
| X11 | Reuse a shape relation proved under a signature requirement in a context lacking that requirement | Recheck its guards; the cached proof must not become an unconditional true |

These cases define an implementation gate, not a maturity score. Passing their
eventual implementation tests is necessary but does not replace the larger community
workload, native compiler regression, editor and performance suites.
