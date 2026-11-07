# TypeScript compiler and language-server integration

R0 candidate. Direct declarations in `.ts`, `.tsx`, `.mts`, `.cts` and declaration
files are the selected integration. A string-tag DSL or source preprocessor would
lose ordinary type symbol navigation and cannot by itself participate in native
generic instantiation. This personal fork can add grammar directly.

## Frontend and symbols

`typefn` and `data` are contextual declaration words. Existing legal uses as
ordinary identifiers remain valid. Declaration lookahead distinguishes native
entries and data helpers without treating `<` in runtime expressions as DSL code.
JS/JSX files receive the normal class of diagnostic for TypeScript-only syntax.
The DSL body has its own grammar and lexical modes. TSX angle syntax inside its
native type islands is not parsed as JSX.

Use dedicated AST kinds for TypeFunctionDeclaration, DataKindDeclaration,
TypeBinding, DataBinding, TypeAssignment, DataAssignment, NativeTypeExpression,
TypeCall, DataCall, MatchStatement, MatchArm and patterns. Reuse a native type node
only inside an explicitly typed native island. Do not place type nodes in ordinary
JavaScript expression fields or disguise captures as unrelated alias declarations.

Every syntax node, including capture names, native islands, pinned expressions
and missing recovery tokens, retains original source ranges and parent links.
Generated Go/TypeScript visitors, encoder/decoder, cloning, printing and incremental
parsing must describe the actual node kinds. A public AST client must not need a
raw untyped clone escape hatch to traverse a body safely.

The binder distinguishes native Type symbols, DSL Type bindings, data bindings,
helper/entry callable values and data-kind names. An identifier can resolve to a
Type handle in the DSL without acquiring a runtime JavaScript value. Runtime
names remain accessible only through native type expressions such as `typeof x`.
Native namespace/module aliases retain their ordinary resolution behavior.

DSL return/break/continue statements do not alter runtime control flow. Native
outer generic parameters can be captured by a typefn declared in a native scope;
runtime local values cannot be captured as executable data. DSL closures capture
only compile-time values and native Type handles, by snapshot.

## Checker hooks

Native generic application dispatch checks an entry's arity, defaults and constraints,
constructs its captured Type environment, and invokes the VM adapter. An entry
can return a native symbolic Type, a closed Type, a deferred application or a
diagnostic. Use the same path from CLI, language-server queries, declaration
checking and native API queries.

Required checker integration points include:

1. Native type-reference resolution and generic alias instantiation.
2. Dedicated deferred computation instantiation and its apparent/indexed types.
3. Native inference through published symbolic structure, with NoInfer/variance
   preserved and deferred inverse inference left opaque.
4. Relations, normalization and recursive graph handling involving deferred Types.
5. Source diagnostics with current application provenance and cancellation.
6. Module visibility, type-only import/export, unused-symbol accounting and
   dependency tracking for helper/data declarations.

Type syntax inside an imperative branch must not be eagerly checked as if every
branch executes for every application. Names and grammar are checked statically;
dependent native semantic validation is a VM/native-adapter operation. Every native
expression form must have a validation policy, not just tuple spreads and indexed
accesses. Missing validators block enabling the feature for release.

## Output and ABI

All DSL declarations and bodies erase from JavaScript. They do not emit a runtime
module dependency. Existing runtime imports are not silently executed or converted
into compile-time modules.

Fork-specific declaration output retains public native entries, their bodies and
all transitive private native/DSL dependencies needed for later applications.
Private helpers and data kinds may need internal emitted names to avoid collisions.
Preserving only producer-side concrete results is insufficient for a generic
consumer. Local snapshot captures require a serializable representation; closures
depending on nonexportable native identities receive a declaration diagnostic,
not an any fallback.

Declaration artifacts carry a `// @typefn-language 1` header and stable logical
operation identities for generative enum sites. The header is an ABI declaration,
not an instruction to fetch or execute external code. A compiler that does not
support that ABI rejects the declaration clearly. When native source dependencies
are copied into `.d.ts`, enum ownership must remain the same logical declaration
identity across producer and consumer; physical output line numbers cannot define
it. Project-reference tests must cover source/declaration redirects.

No portable generic fallback is promised. A separate explicit materialization
command can export finite concrete results to stock TypeScript when they are
representable, including named recursive/nominal declarations if supported. It
must reject arbitrary unresolved programs and nonserializable identities. There
is no silent translation of an imperative loop into an unsupported approximation.

## Configuration and incremental work

The personal fork recognizes the grammar by default. Projects can pin
`compilerOptions.typefn.version` and configure execution resource limits. An
unsupported requested version is a configuration diagnostic. Ordinary TypeScript
projects with no DSL syntax must retain ordinary behavior and performance within
the established regression tolerance.

Changed bodies, imported helpers, native referenced types, compiler options or
resolved modules invalidate affected evaluations. A new checker snapshot creates
new native handles. Parsed/verified code can be reused only when its native symbol
bindings are rebound or known to belong to the same snapshot. Unchanged file text
alone does not establish an unchanged semantic environment.

Incomplete edit states produce recoverable AST nodes and static diagnostics. They
do not enter the VM. A source version canceled during checking cannot later publish
its cache values, diagnostics, quick info or generated declarations to the new version.

## Language server

The server must support:

* Quick info distinguishing a computed result, symbolic structure and a blocked
  application, without showing an invented approximation.
* Definition, references and rename for native entries, helpers, data kinds,
  Type/data locals, template captures, assignments and cross-module imports.
* Completion based on expression mode: native types in Type positions, data
  names/operators in data positions, descriptor fields and pattern vocabulary.
* Semantic tokens and formatting for the real grammar, not regex highlighting.
* Diagnostics on both the failing operation and the application that supplied
  its arguments, with bounded traces and distinct resource/cancellation handling.
* Recovery through incomplete patterns, generics, blocks and descriptor edits.

A debugger/inspector may expose VM steps, captures, resource counts and why an
operation suspended. That information is observational tooling and cannot be read
back by a program to choose its semantic result.

The compiler and language server must use one implementation and ABI. Third-party
parsers/transpilers such as stock TypeScript, Babel and esbuild will not understand
the syntax automatically. Users must run the fork for type checking and erase DSL
syntax before handing source to a parser that does not support it. Supporting
such a pipeline needs a source-preserving erasure transform; a documentation claim
is not a substitute for testing it.
