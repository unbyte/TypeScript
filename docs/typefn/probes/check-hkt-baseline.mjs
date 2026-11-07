// Reproducible observations of the installed stock compiler, not DSL validation.
import assert from "node:assert/strict";
import { fileURLToPath } from "node:url";
import ts from "typescript";

const file = fileURLToPath(new URL("./hkt-baseline.ts", import.meta.url));
const program = ts.createProgram([file], {
    strict: true,
    exactOptionalPropertyTypes: true,
    noEmit: true,
    types: [],
    target: ts.ScriptTarget.ES2022,
    module: ts.ModuleKind.ESNext,
});
const diagnostics = ts.getPreEmitDiagnostics(program);
if (diagnostics.length) {
    console.error(ts.formatDiagnosticsWithColorAndContext(diagnostics, {
        getCanonicalFileName: f => f,
        getCurrentDirectory: () => process.cwd(),
        getNewLine: () => "\n",
    }));
    process.exit(1);
}

const expected = {
    observedNamedBox: "string",
    observedLiteralBox: "string",
    observedNamedEmpty: "number",
    observedLiteralEmpty: "unknown",
    ObservedGenericReturn: "Box<unknown>",
    ObservedSpecializedReturn: "Box<string>",
    ObservedOverloadReturn: "string",
    observedDirectCall: "number",
    ObservedDefault: "unknown",
    ObservedDistributed: '"hello"',
    ObservedWholeUnion: "never",
};
const checker = program.getTypeChecker();
const source = program.getSourceFile(file);
const moduleSymbol = checker.getSymbolAtLocation(source);
const observed = {};
for (const symbol of checker.getExportsOfModule(moduleSymbol)) {
    if (!Object.hasOwn(expected, symbol.name)) continue;
    const type = symbol.flags & ts.SymbolFlags.TypeAlias
        ? checker.getDeclaredTypeOfSymbol(symbol)
        : checker.getTypeOfSymbolAtLocation(symbol, symbol.valueDeclaration);
    observed[symbol.name] = checker.typeToString(type, undefined, ts.TypeFormatFlags.NoTruncation);
}
assert.deepEqual(observed, expected);
console.log(JSON.stringify({
    compiler: ts.version,
    note: "Stock compiler observations; printed spellings and inference can change by version.",
    expectedRejections: 6,
    observed,
}, null, 2));
