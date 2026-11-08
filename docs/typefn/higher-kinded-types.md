> Investigation/history, not a current language contract. The [R2 specification](spec/README.md) supersedes earlier design recommendations; the [manual](manual.md) uses the current syntax.

# Higher-kinded types as a compiler feature

**Broader exploration, superseded in current scope by
[the restricted HKT proposal](restricted-hkt.md).** Native constructor parameters,
nested constructor kinds, bounded constructor signatures and chained application
below are not requirements of the restricted profile. This document preserves
the larger design and its unresolved obligations for comparison.

R1 candidate. Higher-kinded support is a deliberate extension of the fork's
generic system, not something supplied automatically by running an interpreter
inside the checker. The surface should remain small even though the implementation
requires a kind system, constructor substitution and deferred applications.

The [TypeScript constraint audit](typescript-hkt-audit.md) revises this candidate.
Native higher-kinded declarations remain a design goal: output contracts and
relations involving constructor parameters are not yet a complete specification.
Concrete constructor application does not establish those broader capabilities.

## 1. The distinction

```ts
string                 // a Type
Array<string>          // a Type
Array                  // a constructor, when a constructor is expected
```

A normal parameter T takes a Type. `F<_>` takes a constructor with one Type
parameter and a Type result. `F<_, _>` takes a binary constructor. The `_` names
are anonymous signature slots, not any/unknown Types, omitted arguments or
partial-application holes.

```ts
typefn Twice<F<_>, T> {
  return F<F<T>>;
}

type A = Twice<Array, string>;   // string[][]
type B = Twice<Promise, number>; // Promise<Promise<number>>
```

This is type construction, not runtime promise resolution: Promise nesting is not
flattened by the evaluator. A Type such as Promise<string> cannot be supplied as F.
TypeScript's native any also cannot bypass this kind error.

Existing generic aliases, interfaces and classes on their instance-Type side can
be captured as constructors. A class constructor value (`typeof SomeClass`) is a
different native Type. Intrinsic aliases are eligible only with their actual
declared constraints; none receive a special “works for every Type” exemption.
Native typefn declarations and returned typefn closures are also constructors.

## 2. Passing, returning and applying constructors

```ts
typefn Compose<F<_>, G<_>> {
  return typefn<T> { return F<G<T>>; };
}

type AsyncArray = Compose<Promise, Array>;
type Names = AsyncArray<string>; // Promise<string[]>
```

An anonymous typefn has the same body rules as a named one. It captures immutable
snapshots of F and G. Its body runs when applied, not when the closure is created.
Constructors can be passed without inspecting their bodies; mere capture does
not instantiate them with any to discover a result.

The result kind of Compose is inferred statically as a unary constructor. It is
independent of its concrete arguments. Branches cannot return a Type in one arm
and a constructor in another. Such a program fails kind checking even if a
particular call would choose only one arm.

For recursive signatures or documentation, a return-kind annotation is available:

```ts
typefn Compose<F<_>, G<_>>: typefn<_> {
  return typefn<T> { return F<G<T>>; };
}
```

The annotation `typefn<_>` denotes a constructor kind; `: type` denotes a Type
result. Most programs need neither. A recursive return-kind cycle without enough
information must request an annotation, not execute sample arguments or assume any.

Application is uniform and left-associative:

```ts
type Names = Compose<Promise, Array><string>;
```

This first creates the constructor, then applies it. A local `type F = ...` or a
constructor-valued alias can name the intermediate value when it is easier to read.
Native TypeScript does not currently support this general chained application;
the fork's parser/checker must add it explicitly.

## 3. Partial application without special holes

```ts
type Outcome<A, E> =
  | { ok: true; value: A }
  | { ok: false; error: E };

typefn WithError<E> {
  return typefn<A> { return Outcome<A, E>; };
}

type IOResult = WithError<Error>;
type Message = IOResult<string>;
```

A closure provides partial application, argument reordering and additional bounds
without a separate placeholder calculus. For Effect's A/E/R shape, the same form
can fix E/R and expose A. An `_` in an application is an ordinary name if one is
in scope, otherwise a name error: it does not secretly create a closure.

Native generic aliases retain their own semantics when applied. In particular,
an alias's conditional distribution is preserved by invoking native instantiation:

```ts
type NativeStrings<T> = T extends string ? T : never;
typefn Apply<F<_>, T> { return F<T>; }
type S = Apply<NativeStrings, "x" | 1>; // "x"
```

Replacing the alias body with textual substitution could change distribution and
is forbidden. Imperative typefn branches still operate on whole Types unless the
program explicitly iterates members.

## 4. Kinds and bounds

Internally the kinds are recursively:

```text
Type
Constructor(parameters: [kind + optional native bound], result: kind)
```

Examples:

| Parameter | Meaning |
| --- | --- |
| `T` | one Type |
| `F<_>` | Type → Type |
| `F<_, _>` | (Type, Type) → Type |
| `H<F<_>>` | (Type → Type) → Type |
| `C<F<_>, G<_>>: typefn<_>` | two unary constructors → a unary constructor |
| `F<A extends string>` | a constructor accepting all permitted string-subtypes |
| `F<T, K extends keyof T>` | a binary constructor with a dependent second bound |

Named signature slots in F's declaration are scoped to that constructor signature,
not introduced into the enclosing function body. Bounds can refer to preceding
signature slots and outer parameters. Anonymous slots cannot be referenced. F
itself is the only new enclosing parameter.

Kinds describe legal argument categories, not termination or guaranteed success
of an arbitrary VM body. A well-kinded program can still report an explicit error
or exhaust its budget. That is different from ignoring a declared native bound.

### Domain compatibility

```ts
type StringBox<T extends string> = { text: T };

typefn StringsOnly<F<T extends string>> { return F<"hello">; }
type Good = StringsOnly<StringBox>;

typefn ForAnything<F<_>> { return F<number>; }
type Bad = ForAnything<StringBox>; // Rejected at the constructor argument.
```

A supplied constructor must accept every argument promised by the receiving
signature. Its domain can be wider, not narrower. Array can serve as F<T extends
string>; StringBox cannot serve as F<_>. Matching is not based only on arity.

Dependent bounds are checked left to right with alpha-renamed signature parameters.
The checker proves that the expected domain implies the supplied domain under
those same symbolic parameters. Result kinds must agree recursively; higher-order
argument contracts use the same function-domain rule. If a bound implication
depends on unresolved outer arguments, retain an obligation until substitution.
An obligation with no such unresolved dependency that cannot be verified is a
diagnostic. It is never accepted by replacing symbolic arguments with any.

### Output contracts: unresolved design requirement

`F<_>` guarantees a Type result, not properties of that result. A native generic
body cannot access `value.then` merely because value has Type F<T>. To do that,
it needs a contract meaning: for each permitted T, F<T> is assignable to
PromiseLike<T>. This is different from checking F<unknown> once, or checking
that F has one argument. `interface Shifted<T> extends Promise<T[]> {}` accepts
one argument but does not meet the PromiseLike<T> relationship.

The current syntax does not express this output contract. Its binder scope,
symbolic checking, propagation through returned closures, diagnostics and
declaration serialization must be designed before claiming bounded native HKT
support is complete. Arbitrary imperative constructors need not admit a proof;
a future rule must be conservative and must not sample any/never/unknown as a
substitute for checking all permitted arguments. No result-bound syntax is
silently implied by `F<T extends string>`, which constrains inputs only.

A dictionary such as Functor<F> below supplies callable operations explicitly.
It supports useful programs without granting arbitrary members to F<T>. Even a
PromiseLike<T> output bound would not prove that calling .then returns F<U>;
preserving a particular constructor requires a stronger operation contract.

### Defaults and arity

Constructor matching uses declared parameter slots. It does not silently drop
trailing defaulted parameters or auto-curry a binary constructor into a unary one.
A wrapper closure states the intended adaptation explicitly. Defaults remain
available when applying a concrete constructor under its own declaration, but
an abstract F<_> contract promises one supplied argument, not zero.

Ordinary TypeScript default instantiation remains intact:

```ts
type DefaultBox<T = unknown> = { value: T };
type Existing = DefaultBox;              // {value:unknown}, as before
type Wrapped = Twice<DefaultBox, string>; // Constructor expected in this slot
```

In an expected constructor position, a generic identifier supplies its unapplied
declaration; no defaults are instantiated. In an ordinary Type position, normal
TypeScript default/arity rules apply. A bare native generic in an unannotated
local/alias initializer uses those normal rules, even if they produce a missing
argument diagnostic. A failed arity check never changes an expression's kind.
Otherwise adding an upstream default could silently turn a constructor binding
into a concrete Type binding.

An initializer can infer constructor kind from a typefn literal, an already
constructor-valued binding or a constructor-valued computation. Use a closure to
alias a native generic: `type F = typefn<T> { return Array<T>; };`. Once F's kind
is known, `F = Promise` has an expected constructor context and is valid. The
same interpretation applies regardless of whether the supplied native generic
has defaults. Expected result kinds from explicit return annotations also provide
constructor context; inferred return kinds must not use failed native arity checks
to decide that a bare native generic was intended as a constructor.

## 5. Native higher-kinded generic code

Constructor parameters are meaningful beyond executing concrete typefn calls.
The target is to permit them in native generic parameter lists: aliases,
interfaces, classes, functions, methods and signatures, with the same kind rules.
The following is a proposed acceptance example, not an implemented capability
or evidence that native relations and output bounds are fully specified.

```ts
interface Functor<F<_>> {
  map<A, B>(value: F<A>, f: (value: A) => B): F<B>;
}

const arrayFunctor: Functor<Array> = {
  map: (values, f) => values.map(f)
};

function mapWith<F<_>, A, B>(
  functor: Functor<F>,
  value: F<A>,
  f: (value: A) => B
): F<B> {
  return functor.map(value, f);
}

const strings = mapWith<Array, number, string>(arrayFunctor, [1, 2], n => n.toFixed());
```

The runtime implementation comes from the dictionary's map operation. A constructor
parameter by itself does not manufacture runtime mapping, a class instance or a
Promise constructor. This example requires generic F<A> to remain well-formed
while F and A are unknown; a VM that only accepts concrete functions is insufficient.

Allowing these native declarations is distinct from dynamically fabricating
parameter lists through a reflection descriptor. Existing/native-written generic
signatures retain their native binders. Arbitrary runtime-sized generic signature
construction remains outside this design.

## 6. Inference and relations: the limits matter

A type constructor need not be injective or covariant:

```ts
type Constant<T> = string;
type Consume<T> = (value: T) => void;
type Branch<T> = T extends string ? number : boolean;
```

From `F<A>` and `F<B>` producing the same Type, the checker cannot conclude A=B.
From A being assignable to B, it cannot generally conclude F<A> is assignable to
F<B>. From a string result, it cannot recover F or A. A compiler-integrated VM
does not remove any of these ambiguities.

R1's inference policy is deliberately bounded:

* Infer an ordinary Type argument through native symbolic structure that a known
  constructor publishes, using the native checker with NoInfer and variance intact.
* A possible inference extension copies a constructor parameter from a preserved
  constructor argument slot, such as Functor<Array> matched against Functor<F>.
  This requires a dedicated evidence-preservation contract; incidental native
  alias metadata does not establish one. It is not a baseline inference promise.
* Do not infer F by enumerating declarations, executing candidate constructors,
  matching a displayed type name, or solving arbitrary equations F<A>=B.
* Do not infer A by decomposing an application of unknown F. F may discard A.
  If another input later determines F, retry the postponed ordinary inference.
* If no constructor evidence remains, request an explicit argument; do not default
  to Array, an identity constructor, any, or a synthetic inverse lambda.

Explicit constructor arguments are the normative baseline. Annotation alone is
not yet a guarantee of constructor inference: native inference can switch between
structural and instantiation-based strategies. Before enabling inferred constructor
arguments, specify what evidence survives mapped copies, aliases, declaration emit,
unused parameters and overload/contextual typing. No compiler optimization may
silently serve as the language's only evidence-preservation rule.

Unknown applications are invariant by default in relations. The same application
is reflexive, but unrelated applications are not equated by syntactic optimism.
After F becomes known, reduce it and use native relations on the result. A blocked
relation must be replayed rather than permanently cached as false. Existing
variance information of a known native constructor can be used normally; R1 does
not introduce an unproved covariance annotation for arbitrary imperative functions.

This conservative application rule does not settle relations between native types
that contain constructor parameters. For example, proposed `Keeps<F<_>>` with a
`value: F<string>` member must distinguish Keeps<Array> from Keeps<Promise>, while
empty `Phantom<F<_>>` must retain structural compatibility across constructors.
Kind compatibility, constructor identity for caching, and assignability of applied
native Types are separate relations. Making all enclosing types nominal or all
same-kind constructors interchangeable would both be wrong. Native variance
inference currently substitutes ordinary marker Types; those markers cannot be
applied as F. Kind-correct variance probing or a conservative structural fallback
is a required design decision, with recursive relation/cycle handling still to
specify. This is a blocker to claiming general native HKT relations are settled.

NoInfer must also work at constructor kind. In an expected constructor slot,
`NoInfer<F>` preserves F's kind and application behavior but does not contribute
a constructor-inference candidate. This is a kind-preserving compiler intrinsic,
not a constructor converted to an ordinary Type. Existing Type-kind NoInfer keeps
its native behavior. Other native aliases do not acquire kind polymorphism merely
because NoInfer has this explicitly specified special contract.

No extensional equality test for constructors is provided. Internal identity uses
the declaration or closure site and captured values. Two different functions may
produce identical results without having the same constructor identity. Budgeted
beta reduction of applications is allowed; speculative eta/extensional proofs
are not required for ordinary kind checking.

## 7. Required compiler representation

The inspected baseline's `getTypeFromTypeAliasReference` resolves a declared
generic, checks arity and instantiates it with native Type arguments. Its
instantiation key and TypeMapper are built around Type identities. That is not
already a first-class constructor abstraction.

The fork needs distinct semantic entities:

```text
TypeLevelArgument = NativeType | ConstructorValue
ConstructorValue = NativeDeclarationRef(symbol, captured environment)
                 | DSLClosure(code, captured environment, kind)
                 | ConstructorParameter(symbol, kind)
                 | DeferredConstructorApplication(callee, arguments, result kind)
TypeApplication  = Apply(constructor, kind-checked arguments) // Type result
```

A constructor is not represented by a native generic instantiated with any. It
cannot appear as an ordinary object property Type, union constituent or tuple
element until applied. An ordinary data collection may carry opaque constructor
handles, but that does not turn a native tuple into a heterogeneous kind container.

Substitution maps both Type and constructor parameters, traverses captures and
applications, and respects locally bound parameters. Applying a native reference
uses original native instantiation, preserving conditional distribution, defaults,
constraints, variance, identity and module augmentation. Applying a closure
invokes the VM; a symbolic constructor parameter retains a typed application node.
An application returning a constructor retains its result kind even when blocked.

Native operators must retain their dependency on unknown applications. For example,
`keyof F<T>` remains a symbolic key computation; it must not become never by treating
F<T> as unknown. After substituting Box for F it can become `"value"`. A generic
runtime body still cannot access an unproved member of F<T>; preserving symbolic
type operators is different from inventing member guarantees.

Generic declaration metadata and all argument/cache paths must become kind-aware,
including native interface/class references, defaults, signatures, apparent types,
inference candidates, relation caches, alias preservation and declaration emit.
A tagged argument vector is required; merely adding one parser production and a
VM opcode would leave most of the type system inconsistent.

Inference-barrier wrappers likewise retain the wrapped argument's kind. They
are substituted and applied transparently but honored when collecting candidates.

Closure memoization includes code identity, captured Types/constructors/data,
argument kinds and compiler snapshot/options. Captured native pointers never
cross checkers. Substitution/evaluation that fails or is canceled publishes no
partial closure/application result. Recursion and constructor normalization share
the root evaluation budget; normalization is not an unbounded theorem prover.

## 8. Declaration and editor implications

Native declarations with F<_> retain that syntax in fork-specific .d.ts output.
Constructor aliases retain their kind, and returned closures retain accessible
definitions/captured type dependencies. Source/declaration redirects must not
invent new constructor identity or drop constraints. Constructor values have no
JavaScript emission.

Hover distinguishes a constructor from its applied Type. Completion in an F<_>
argument slot offers compatible generic constructors; a normal Type slot offers
Types. Rename follows F references in application heads and captured closures.
Kind/constraint errors identify both the supplied declaration and the required
constructor signature. Native TypeScript consumers cannot parse these declarations
without an explicit compatibility/export strategy.

## 9. Evidence and scope

The inspected [fp-ts HKT](https://github.com/gcanti/fp-ts/blob/master/src/HKT.ts)
uses URItoKind registries and separate arities to recover constructor application.
[Hotscript Core](https://github.com/gvergnaud/hotscript/blob/main/src/internals/core/Core.ts)
uses this/rawArgs plus Apply/Call to encode callable type programs and pipelines.
[Effect HKT](https://github.com/Effect-TS/effect/blob/main/packages/effect/src/HKT.ts)
uses TypeLambda slots and Kind, including distinct input/output/target roles and
variance wrappers in its abstract fallback.

R1 replaces the application encodings with actual constructor values and symbolic
applications. It must preserve the libraries' underlying relationships; replacing
four Effect slots with one undifferentiated argument would not preserve its API.
Existing open native registries remain usable. Their source snapshots are recorded
in the research manifest; no upstream HKT suites have been run for R1.

This design proposes higher-order, higher-kinded constructor polymorphism with
bounded inference. Concrete application has a narrower contract than general
native HKT integration. Open output-contract, relation and inference-evidence
decisions are recorded in the audit. It does not claim arbitrary higher-order
unification, generic runtime code synthesis, extensional function equality or a
mature implementation. Applying generic runtime call signatures (including overload
resolution) is a separate capability; a Type with a generic call signature is not
automatically a constructor value.
