# Typefn integration harness

From the repository root, run `go test ./tsc/internal/typefntest`. Each JSON manifest creates an isolated virtual
TypeScript project with the real compiler, bundled native libraries and tsconfig
parser. Files and options are explicit. Native Type assertions establish results;
negative fixtures calibrate that assertions reject wrong Types, including any.
In successful fixtures, each native `Assert<...>` application is also queried
through the checker and must resolve to the boolean literal `true`. This catches
silent recovery Types that can otherwise make constraint-based assertions vacuous.

Diagnostics match phase, code, source file, optional unique source-position marker
and message substring. The entire diagnostic set must match; extra errors cannot
hide behind an expected one. Optional emit expectations compare every output's
exact text. The harness does not evaluate malformed source or transpile DSL code
to JavaScript. Fixtures run concurrently with independent programs/checkers.

Native fixtures calibrate the harness; DSL fixtures use the same pipeline.
`emitSkipped` supports explicit blocked-emit expectations. Benchmarks use fresh
compiler projects and include library setup, rather than claiming VM-only speed. Specification cases remain
unimplemented until a real fixture establishes their behavior; no automatic skips
or baseline acceptance mode is provided here.
