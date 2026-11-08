// Research evidence against installed compilers; does not accept proposed syntax.
import assert from "node:assert/strict";
import { fileURLToPath } from "node:url";
import ts from "typescript";

const file = fileURLToPath(new URL("./where-baseline.ts", import.meta.url));
const program = ts.createProgram([file], {
    strict: true, exactOptionalPropertyTypes: true, noEmit: true, types: [],
    target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext,
});
const source = program.getSourceFile(file);
const expectedDiagnostics = source.text.split("\n").flatMap((line, index) => {
    const match = /@probe-error (\d+)/.exec(line);
    return match ? [{ line: index + 2, code: Number(match[1]) }] : [];
});
const diagnostics = ts.getPreEmitDiagnostics(program);
const actualDiagnostics = diagnostics.map(d => ({
    line: d.file?.getLineAndCharacterOfPosition(d.start).line + 1,
    code: d.code,
}));
assert.deepEqual(actualDiagnostics, expectedDiagnostics,
    ts.formatDiagnosticsWithColorAndContext(diagnostics, {
        getCanonicalFileName: f => f, getCurrentDirectory: () => process.cwd(), getNewLine: () => "\n",
    }));

const expectedTypes = {
    AnyArgument: '"accepted"', NeverArgument: '"accepted"',
    DistributedUnion: "boolean", WholeUnion: "false", DistributedAny: "boolean",
    DistributedNever: "never", WholeNever: "true", MergedAcceptsString: "string",
    inferredPlain: "[string]", inferredBounded: '["hello"]', silentlyNever: "never",
};
const checker = program.getTypeChecker();
const observed = {};
for (const symbol of checker.getExportsOfModule(checker.getSymbolAtLocation(source))) {
    if (!Object.hasOwn(expectedTypes, symbol.name)) continue;
    const type = symbol.flags & ts.SymbolFlags.TypeAlias
        ? checker.getDeclaredTypeOfSymbol(symbol)
        : checker.getTypeOfSymbolAtLocation(symbol, symbol.valueDeclaration);
    // Expand alias names when recording resolved Types.
    observed[symbol.name] = checker.typeToString(type, undefined,
        ts.TypeFormatFlags.NoTruncation | ts.TypeFormatFlags.InTypeAlias);
}
assert.deepEqual(observed, expectedTypes);

console.log(JSON.stringify({
    compiler: ts.version,
    expectedRejections: expectedDiagnostics.length, observed,
    note: "Native TypeScript observations; no where feature or Rust execution is being tested.",
}, null, 2));
