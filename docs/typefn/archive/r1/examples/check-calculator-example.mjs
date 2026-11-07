// Algorithm check only: this is NOT a typefn compiler or the proposed VM.
// Extract the trusted example's ordinary data-helper body, erase annotations,
// and adapt its int/string operations to JavaScript BigInt and UTF-16 strings.
// Does not verify DSL parsing/kinds, symbolic Types, captures, budgets or emit.
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import vm from "node:vm";
import ts from "typescript";

const source = readFileSync(new URL("./calculator.typefn", import.meta.url), "utf8");
const match = source.match(/^typefn CalcBig<S> \{([\s\S]*?)^\}/m);
assert.ok(match, "Cannot locate CalcBig's data-helper body");

function adapt(context) {
    const visit = node => {
        if (ts.isNumericLiteral(node)) {
            return ts.factory.createBigIntLiteral(`${node.text}n`);
        }
        if (ts.isStringLiteral(node)) {
            return ts.factory.createCallExpression(ts.factory.createIdentifier("str"), undefined, [node]);
        }
        if (ts.isBinaryExpression(node) &&
            [ts.SyntaxKind.EqualsEqualsToken, ts.SyntaxKind.ExclamationEqualsToken].includes(node.operatorToken.kind)) {
            const equal = ts.factory.createCallExpression(ts.factory.createIdentifier("equal"), undefined,
                [ts.visitNode(node.left, visit), ts.visitNode(node.right, visit)]);
            return node.operatorToken.kind === ts.SyntaxKind.EqualsEqualsToken ? equal
                : ts.factory.createPrefixUnaryExpression(ts.SyntaxKind.ExclamationToken, equal);
        }
        return ts.visitEachChild(node, visit, context);
    };
    return node => ts.visitNode(node, visit);
}

const compiled = ts.transpileModule(`function calculate(S: string) {${match[1]}\n}\ncalculate;`, {
    compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext },
    transformers: { before: [adapt] },
    reportDiagnostics: true,
});
assert.equal(compiled.diagnostics?.length ?? 0, 0,
    (compiled.diagnostics ?? []).map(d => ts.flattenDiagnosticMessageText(d.messageText, "\n")).join("\n"));

const unbox = value => value instanceof DataString ? value.value : value;
class DataString {
    constructor(value) { this.value = String(unbox(value)); Object.freeze(this); }
    get length() { return BigInt(this.value.length); }
    slice(start, end) {
        return new DataString(this.value.slice(Number(start), end === undefined ? undefined : Number(end)));
    }
    includes(value) { return this.value.includes(unbox(value)); }
    indexOf(value) { return BigInt(this.value.indexOf(unbox(value))); }
    [Symbol.toPrimitive]() { return this.value; }
}

const context = vm.createContext({
    str: value => new DataString(value),
    equal: (left, right) => unbox(left) === unbox(right),
    literal: { text: value => new DataString(value), bigint: value => value },
    fail: message => { throw new Error(String(message)); },
});
new vm.Script(compiled.outputText).runInContext(context, { timeout: 1000 });
const invoke = new vm.Script("calculate(input)");
const calculate = input => {
    context.input = input;
    return invoke.runInContext(context, { timeout: 1000 });
};

const valid = [
    ["2 + 3 * 4", 14n], ["(2 + 3) * 4", 20n], ["-(10 - 4) / 3", -2n],
    ["8 - 3 - 2", 3n], ["16 / 4 / 2", 2n], ["2 * -3", -6n],
    ["--2", 2n], ["1+-2", -1n], ["1++2", 3n], ["+(3)", 3n],
    ["00012 + 3", 15n], [" \t2 + 3\r\n", 5n], ["0", 0n], ["-0", 0n],
    ["9007199254740992 + 1", 9007199254740993n],
    ["12345678901234567890 * 10", 123456789012345678900n],
    ["2 * (3 + (4 * 5))", 46n], ["3 * (2 - 5)", -9n], ["6 / -3", -2n],
];
for (const [input, expected] of valid) assert.equal(calculate(input), expected, input);

const invalid = [
    ["", "Expected an expression", 0], [" \t ", "Expected an expression", 3],
    ["1+", "Expected an expression", 2], ["()", "Expected an integer or '('", 1],
    ["(1+2", "Expected ')'", 4], ["1+2)", "Unexpected token", 3],
    ["2(3)", "Unexpected token", 1], ["2**3", "Expected an integer or '('", 2],
    ["1.5+2", "Unexpected token", 1], ["1e3", "Unexpected token", 1],
    ["0x10", "Unexpected token", 1], ["1_000", "Unexpected token", 1],
    ["2^3", "Unexpected token", 1], ["6/0", "Division by zero", 1],
    ["7/2", "Non-integral division", 1], ["1+💡", "Expected an integer or '('", 2],
    ["1 + (2 * 3", "Expected ')'", 10], ["2 + 3 garbage", "Unexpected token", 6],
];
for (const [input, message, offset] of invalid) {
    assert.throws(() => calculate(input), { message: `${message} at UTF-16 offset ${offset}` }, input);
}

// Expected values come from generated expression trees, not another parser.
let seed = 0x12345678;
const random = max => {
    seed = (Math.imul(seed, 1664525) + 1013904223) >>> 0;
    return seed % max;
};
function expressionTree(depth) {
    if (depth === 0) {
        const value = BigInt(random(19) - 9);
        return { text: value.toString(), value };
    }
    const left = expressionTree(depth - 1);
    const right = expressionTree(depth - 1);
    switch (random(4)) {
        case 0: return { text: `(${left.text} + ${right.text})`, value: left.value + right.value };
        case 1: return { text: `(${left.text} - ${right.text})`, value: left.value - right.value };
        case 2: return { text: `(${left.text} * ${right.text})`, value: left.value * right.value };
        default: {
            const divisor = BigInt(random(9) + 1);
            return { text: `((${left.text} * ${divisor}) / ${divisor})`, value: left.value };
        }
    }
}
for (let i = 0; i < 200; i++) {
    const tree = expressionTree(1 + random(5));
    assert.equal(calculate(tree.text), tree.value, tree.text);
}

console.log(`${valid.length} fixed results, ${invalid.length} error/offset cases, 200 generated expressions passed.`);
console.log("Host algorithm checks only; no DSL compiler, VM or native-Type behavior was tested.");
