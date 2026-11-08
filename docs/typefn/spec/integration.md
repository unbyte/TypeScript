# TypeScript compiler, declarations and language server

Normative R2 chapter. These are required fork behaviors, not options available in
the unmodified compiler today.

## 1. Files and configuration

Enable this profile explicitly in the fork's tsconfig:

```json
{
  "compilerOptions": {
    "strict": true,
    "typefn": { "version": 1 }
  }
}
```

Absent typefn configuration means native behavior, with the extension disabled.
Unsupported versions diagnose. The option is accepted by both CLI and language
server project configuration. Limits can be overridden through
`typefn.limits`: `vmInstructions`, `callDepth`, `syntaxDepth`, `parserWork`,
`liveBytes`, `bigintBits`, `stringUnits`, `nativeWork`, `nativeNodes`, `cacheBytes`.
Values are positive safe integers; parserWork, when supplied, is an absolute cap
instead of the token-count-derived default. Limits affect service behavior, not
the Type semantics or declaration ABI. An inferred editor project must explicitly
select the profile through its configuration; file text cannot enable execution.

Source support is `.ts`, `.tsx`, `.mts`, `.cts` and their declaration counterparts.
In TSX the parser enters DSL mode inside typefn; JSX is not parsed in that body.
There is no `.typefn` module/file kind in R2. Example `.typefn` files in this
documentation are source listings to copy into enabled TypeScript files.
JavaScript/checkJs files cannot declare typefns. Native .ts syntax outside DSL
declarations/islands remains unchanged, except constructor-valued nongeneric aliases.

`typefn` is contextual at native declaration positions. Existing uses of the
identifier in native value code remain legal. Valid-looking and malformed typefn
declarations use dedicated AST/recovery nodes; the parser must not swallow the
rest of the file as a JavaScript expression after a DSL error.

## 2. Module and runtime boundary

```ts
// transforms.ts
export typefn Box<T> { return { value: T }; }
// consumer.ts
import type { Box } from "./transforms.js";
type Result = Box<string>;
```

Typefn declarations/constructor aliases occupy the type namespace. Native symbol
collision rules apply, but declaration merging/overloading of typefn itself is
not allowed. Type-only imports/re-exports and qualified imported namespace names
resolve through native module resolution. Imported named helpers are directly
callable; source recipes are compiled in the consuming checker snapshot.

Top-level/local typefn declarations and constructor aliases emit no runtime code.
Runtime reads, `new`, `typeof` as a JS expression, value exports or value calls of
a typefn diagnose. Names referenced only by a DSL body count as type dependencies;
they must not cause a runtime import. Ordinary native code emits normally. The
VM cannot execute runtime functions, inspect parameter values or run a module.

## 3. Native generic checking and inference

Known result structure participates in native property access, contextual typing,
assignability and inference. An opaque pending call supplies no invented fields
or reverse inference candidates. Native NoInfer, variance and declaration identity
must survive construction/forwarding. Constructors stay in DSL signatures and
environments, never native interface/class/signature argument vectors.

The VM observes inference variables only at stable native query/substitution
boundaries. It cannot inspect candidate lists, inference priority or half-finished
solutions. Arbitrary programs are not invertible. A Box<T> field may expose T
structurally; Calc<T> does not infer its input string from a numeric output.

Relation queries distinguish proven, failed and blocked. Positive relations
depending on guards record those guards. A cached proof under one generic
signature's assumptions cannot be reused as unconditional true elsewhere.
Identical deferred terms can be reflexive, but their validity checks remain.
A native use requiring an unresolved member/relation diagnoses at that use;
it is neither silently accepted nor permanently cached as false.

### Generic runtime bodies are checked once

Native generic aliases may describe partial computations. A written native
signature may carry pending checks through its parameter/result Type expressions.
Its generic body can reason under those explicit requirements; each specialization
must discharge them. Requirement matching uses canonical recipe identity after
substitution, not general theorem proving.

```ts
typefn Checked<T> { return Uppercase<T>; }
typefn Discard<T> { type unused = Checked<T>; return string; }

function declared<T>(x: Checked<T>): Checked<T> { return x; }
function bounded<T extends string>(x: Checked<T>): Checked<T> { return x; }

function bad<T>(x: any) {
  const y: Discard<T> = x; // New unproved local check: definition diagnostic.
  return y;
}
```

declared<number> fails during signature specialization even with an any argument
or an ignored result. The body is not rechecked for each call. New local validity
requirements must be proven from known Types/bounds or already be in the written
signature; do not infer hidden runtime-function effects. An unused local generic
type alias merely declares a recipe, rather than executing a runtime operation.

Function assignability must not drop a source signature's unresolved requirements.
A symbolic comparison must prove/match them to the target's explicit requirements;
a concrete comparison must discharge them. Any/casts can still erase an entire
callable Type under native escape rules. An any *argument* to a still-known
signature does not erase that signature's specialization checks.

## 4. Declaration ABI and recipes

Every emitted declaration file containing DSL definitions or constructor aliases
begins, after optional BOM/trivia, with the preserved pragma:

```ts
// @typefn-abi: 1
```

Missing, conflicting or unsupported headers on such declaration files diagnose
before execution, including with skipLibCheck. Ordinary native-only declarations
need no header. This is the first public ABI; archived draft headers were proposals,
not released formats. Source files select their profile through configuration and
need no pragma. ABI covers DSL grammar, intrinsic
contracts and staged behavior, not compatibility across arbitrary native versions.

Emit source-equivalent bodies and transitive dependencies. Do not serialize
frames, bytecode, checker pointers, cache results or quota counters. Prefer the
originating recipe even for currently concrete results: dependencies on native
open interfaces, constraints and captures can matter in the consuming program.

```ts
// Producer source
typefn WithError<E> {
  return typefn<A> {
    return { ok: true; value: A } | { ok: false; error: E };
  };
}
export type Result = WithError<Error>;
```

The declaration retains the private WithError definition and exported alias, with
ABI 1. It does not export WithError automatically or replace Result by a guessed
native generic alias. Constructor creation replays source capture timing:

```ts
export typefn Save<T> {
  type current = T;
  type saved = typefn<X> { return [current, X]; };
  current = number;
  return saved;
}
```

Save<string>'s returned constructor must still use string after declaration-only
consumption. The emitter cannot serialize the final mutable current slot instead.

## 5. Dependencies, lifting and atomic failure

Collect references lexically from signatures, all branch bodies, islands, local
functions and closure recipes, including branches not exercised by the producer.
Preserve needed private helpers with hygienic names and external type-only imports.
Native private classes, unique symbols, enums and generic declarations retain
their module identities; no structural textual clones are permitted.

If an inferred export depends on a local constructor alias capturing a native
generic parameter, lift a private Type-returning helper with explicit ordinary
Type parameters and the source recipe. Do not add constructor slots to the native
signature. Generated names are deterministic for the source dependency graph and
collision-checked. Source/declaration redirects must preserve original identities.
When the recipe/dependencies cannot be made nameable, diagnose the escaping
dependency and require an exportable annotation/helper; never substitute any.

Guards survive emission: `type Hidden<T> = Discard<T>` must not emit as string.
Deferred native operations retain their scoped binder/laziness recipe. Re-exported
typefns keep original code identity; they are not independent text copies. Module
cycles can bind mutually recursive functions before execution, but cannot execute
a partially bound module. Recursive result-kind annotations remain required.

With noEmitOnError, follow the native whole-program policy. Otherwise native
outputs may follow native policy, but withhold any declaration file whose DSL
dependency closure is invalid. Its output is atomic. Cancellation emits no stale
file. skipLibCheck never skips frontend/bytecode safety checks needed for execution.
There is no stock-TypeScript declaration fallback or implicit materialization mode.

## 6. Editor and incremental behavior

CLI and language server share parsing, binding, bytecode validation, evaluation
and ABI rules. Formatting retains mandatory semicolons, template patterns and
native islands. Rename/go-to-definition use lexical binder identities, including
template captures and lifted helper mappings. Completion distinguishes local
Types, constructors and prelude namespaces; it does not suggest host JS methods.

Hover may show published shape and a concise pending computation/validation note.
It must not present a guarded Type as an unconditional completed result. Hover
and completion execute only within request budgets/cancellation, never by ignoring
the selected profile. Malformed buffers may provide syntax services but no recovered
body executes. Consumer errors point to their use, with related producer spans.

An edit produces a new checker snapshot. Module augmentation, native options,
imports, function code and captured definitions invalidate relevant compilation/
results. Native handles/results never cross snapshots. A source declaration map
must not accidentally make a declaration-only test consume producer source.

## 7. Implementation map and compatibility gate

The baseline has entry points in `tsc/internal/parser`, `tsc/internal/ast`,
`tsc/internal/binder`, `tsc/internal/checker`, `tsc/internal/transformers/declarations`,
`tsc/internal/tsoptions`, `tsc/internal/lsp` and `tsc/internal/ls`. Required changes include:

* Dedicated AST kinds, visitors, scanner modes, formatter/emitter support and
  generated protocol/visitor updates where the native architecture requires them.
* Checker substitution of guarded/deferred terms before native fallback paths;
  guard extraction/reapplication around absorbing union/alias simplification.
* Readiness-aware constraint/relation adapters, native inference boundaries,
  incremental dependency tracking and cancellable resource accounting.
* Source-recipe declaration building/tracking before apparent-Type serialization.

This map does not assert those hooks alone are sufficient. Native regression
suites must pass with the feature disabled and enabled on DSL-free projects.
Source and declaration-only projects, editor edits, module cycles, TSX and
project references require dedicated conformance tests before implementation
can be called integrated.
