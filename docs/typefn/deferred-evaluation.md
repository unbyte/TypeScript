# Deferred calculations: implemented boundary

This guide describes the deferred-evaluation support in the current fork.
The [evaluation specification](spec/evaluation.md) remains normative. This is a
tested portion of that design, not a claim that its entire native-inference and
requirement-preservation contract is implemented.

## When calculations run

A `typefn` declaration supplies a program. Checking a declaration validates its
names and control constructs; it does not execute the body with invented inputs.
An application evaluates as far as its inputs permit. Native generic substitution
replays pending calculations when arguments change:

```ts
typefn Twice<N> { return N * 2; }

type Deferred<N> = Twice<N>; // Pending arithmetic; no diagnostic.
type Answer = Deferred<21>; // 42.
type Invalid = Deferred<"x">; // Diagnostic at specialization: invalid arithmetic.

declare function double<N>(value: N): Twice<N>;
const answer = double(21 as const); // Native inference supplies 21; result is 42.
```

This happens during type checking, when the checker requests the Type. There is
no JavaScript execution or runtime code for the calculation. A constraint such as
`N extends number` is evidence about N, not a specific numeric value.

Known structures can retain pending components:

```ts
typefn Box<N> { return { value: N * 2 }; }
type DeferredBox<N> = Box<N>;
type Value = DeferredBox<21>["value"]; // 42.
```

When a branch, loop, or iteration needs an unknown value, the current call remains
pending. Replay starts that call with its original arguments and fresh mutable
slots. The VM does not choose a speculative branch or save a partially mutated
frame. Tuple spreads and computed property keys can instead retain an expression
with an immutable snapshot of the local slots it references.

## A result includes checks that must still succeed

An evaluated expression can require validation even if its value is discarded:

```ts
typefn Checked<N> {
  type unused = N * 2;
  return "ok";
}
type Result<N> = Checked<N>;
type Good = Result<21>; // "ok".
type Bad = Result<"x">; // Diagnostic; the unused multiplication still matters.
```

The VM records reached pending operations in a per-call journal. Returning a
known Type attaches those checks to it. Native simplification must not turn the
result into an unconditional success. Tests cover checks surviving discarded
locals, unused native alias parameters, `any` union absorption, `never`
intersection absorption, object projection, `keyof`, conditional selection,
native arrays/tuples, native constraints, and omitted generic defaults.

Only reached operations contribute. A closed branch that is not taken does not
execute its arithmetic. An unknown branch defers the call until its condition can
be evaluated. Native declaration errors such as unknown names remain definition
errors even in unreachable DSL branches.

## Native checker representation

`tsc/internal/checker/typefn_types.go` defines the native term representation;
`tsc/internal/checker/typefn_deferred.go` manages immutable checker Types for
unary/binary operations, named calls, captured expressions, native applications,
and known shapes with pending checks. These use `TypeFlagsTypefn`; they are not
ordinary Type parameters or an `any` placeholder.

The checker interns terms by their operation and operand identities. Canonical
guard sets deduplicate requirements. Native TypeMapper substitution substitutes
operands, reuses an unchanged term, and replays a changed term under the shared
VM instruction, construction, and call-depth budgets. Guard requirements are
substituted before the guarded shape. Failure produces native error recovery and
a diagnostic, not a usable successful Type.
Failed deferred return Types are not cached on shared instantiated signatures;
each invalid call site must receive its own diagnostic. Successful and pending
return Types retain the normal native cache behavior.

Native union/intersection, indexing, conditional selection, and alias/container
construction have explicit integration hooks. DSL-created anonymous objects have
a substitution path for their immutable property tables. A syntax reachability
cache finds eager native arguments and omitted defaults containing calculations;
it is cycle-aware and deliberately avoids expanding latent object members,
mapped members, and generic signatures across binder scopes.

Relation checking is conservative: identical pending terms are reflexive, and
known shapes with the same canonical check set can be compared. A generic
signature cannot silently drop its pending checks; `any` cannot manufacture a
new guarded result in a generic body. This is not yet the full scoped-assumption
and requirement-bearing relation-proof system from the specification.

Programs, terms, snapshots, and Type handles belong to their checker instance.
No mutable VM frame is retained in a term. No Type handle is shared across editor
snapshots. Canceled replay makes its checker unusable; the editor pool discards
that checker before a later request. The interner is not a general memo table for
completed VM calls.

## Editor behavior

Pending arithmetic displays as an expression such as `N * 2`. A known shape with
outstanding checks displays as `pending validation<...>`. That label is
**display-only**, not additional source syntax. A pending application has no
completed-result inlay hint and no hover claiming a completed result. Once a
direct application completes, its ordinary result overlay is available.

Diagnostics during replay use the specialization site. Declaration-file recipes
are still unsupported, so display syntax is not a declaration ABI.

## Regression coverage

The compiler-project harness now checks that every `Assert<...>` in a successful
fixture actually resolves to the literal Type `true`. A native constraint alone
can accidentally accept an error-recovery Type; the harness rejects that false
positive. Negative fixtures verify complete expected diagnostic sets.

The `deferred-*` and `error-deferred-*` fixtures cover composed substitutions,
arithmetic, bigint unary operations, control flow, loop replay, captured locals,
object substitution, native inference (including repeated invalid inferred calls),
constraints/defaults, check retention,
generic signature rejection, and replay budgets. Editor tests verify pending
display, omission of result hints, successful specialization, and unchanged-query
Type-count stability. Project tests deterministically cancel both direct execution
and replay, then verify a successful request with a fresh checker.

See [testing](testing.md) for commands and validation scope.

## Remaining limits

* Full native relation proofs, scoped generic-signature assumptions, latent
  member/signature dependencies, and exhaustive inference/replay laws remain
  unimplemented. The conservative relation rules can reject otherwise valid
  generic programs. The tested native boundaries are not exhaustive coverage of
  every checker simplification or cache.
* Readiness analysis handles the implemented scalar and structural operations;
  the complete kind/readiness analysis and recursive call-graph checks remain work.
* Resource accounting remains partial. Native graph work, all live buffers,
  strings, retained caches, and cancellation inside every native operation still
  need instrumentation. Shared replay budgets alone do not establish a hardened
  service boundary or large-project performance.
* String algorithms, patterns, restricted HKT/closures, executable native islands,
  dynamic signature builders, and declaration recipes remain outside this slice.
  The manual's route and calculator programs are still target-language examples.

See [implementation status](implementation.md) for the wider language status and
[editor integration](editor-integration.md) for running the fork.
