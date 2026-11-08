import {
    isIdentifier,
    isTypeFunctionExpression,
    SyntaxKind,
    TokenFlags,
    visitEachChild,
} from "@typescript/typescript/ast";
import {
    createIdentifier,
    createKeywordTypeNode,
    createSourceFile,
    createTemplateHead,
    createTemplateTail,
    createToken,
    createTypeFunctionBlock,
    createTypeFunctionDeclaration,
    createTypeFunctionExpression,
    createTypeFunctionKind,
    createTypeFunctionMatch,
    createTypeFunctionMatchArm,
    createTypeFunctionPattern,
    createTypeFunctionStatement,
    createTypeFunctionTemplate,
    createTypeFunctionTemplateSpan,
    updateTypeFunctionExpression,
} from "@typescript/typescript/ast/factory";
import {
    CaseSensitivity,
    pathKey,
    toRootedFilePath,
} from "@typescript/typescript/path";
import assert from "node:assert/strict";
import { test } from "node:test";
import { encodeSourceFile } from "../src/api/node/encoder.ts";
import { RemoteSourceFile } from "../src/api/node/node.ts";

function binary() {
    const node = createTypeFunctionExpression(SyntaxKind.TypeFunctionBinary, undefined, [createIdentifier("left"), createIdentifier("right")], createToken(SyntaxKind.PlusToken));
    return node;
}

test("typefn expression transforms visit operators in order and retain kind", () => {
    const expression = binary();
    const visited: SyntaxKind[] = [];
    const transformed = visitEachChild(expression, child => {
        visited.push(child.kind);
        return child === expression.operands![1] ? createIdentifier("changed") : child;
    });
    assert.deepEqual(visited, [SyntaxKind.Identifier, SyntaxKind.PlusToken, SyntaxKind.Identifier]);
    assert.equal(transformed.kind, SyntaxKind.TypeFunctionBinary);
    assert.notEqual(transformed, expression);
    const updated = updateTypeFunctionExpression(expression, undefined, [createIdentifier("other"), expression.operands![1]], expression.operatorToken);
    assert.equal(updated.kind, SyntaxKind.TypeFunctionBinary);
});

test("typefn return transforms retain statement kind", () => {
    const expression = binary();
    const statement = createTypeFunctionStatement(SyntaxKind.TypeFunctionReturn, undefined, expression);
    const transformed = visitEachChild(statement, child => child === expression ? createIdentifier("changed") : child);
    assert.notEqual(transformed, statement);
    assert.equal(transformed.kind, SyntaxKind.TypeFunctionReturn);
});

test("all DSL expression kinds participate in transforming visitors", () => {
    const value = createIdentifier("value");
    const name = createIdentifier("name");
    const native = createKeywordTypeNode(SyntaxKind.StringKeyword);
    const expressions = [
        createTypeFunctionExpression(SyntaxKind.TypeFunctionName, name),
        createTypeFunctionExpression(SyntaxKind.TypeFunctionLiteral, undefined, undefined, undefined, native),
        createTypeFunctionExpression(SyntaxKind.TypeFunctionCall, name, [value]),
        createTypeFunctionExpression(SyntaxKind.TypeFunctionUnary, undefined, [value], createToken(SyntaxKind.MinusToken)),
        binary(),
        createTypeFunctionExpression(SyntaxKind.TypeFunctionTuple, undefined, [value]),
        createTypeFunctionExpression(SyntaxKind.TypeFunctionObject, undefined, [createTypeFunctionExpression(SyntaxKind.TypeFunctionField, name, [value])]),
        createTypeFunctionExpression(SyntaxKind.TypeFunctionField, name, [value]),
        createTypeFunctionExpression(SyntaxKind.TypeFunctionIndex, undefined, [name, value]),
        createTypeFunctionExpression(SyntaxKind.TypeFunctionNativeType, undefined, undefined, undefined, native),
        createTypeFunctionExpression(SyntaxKind.TypeFunctionSpread, undefined, [value]),
    ];
    for (const expression of expressions) {
        const transformed = visitEachChild(expression, child =>
            child === native
                ? createKeywordTypeNode(SyntaxKind.NumberKeyword)
                : child === expression.operatorToken ? createToken(SyntaxKind.PlusToken) : createIdentifier("changed"));
        assert.notEqual(transformed, expression, `visitor skipped ${expression.kind}`);
        assert.equal(transformed.kind, expression.kind);
    }
});

test("remote DSL trees transform and re-encode without changing kinds", () => {
    const ret = createTypeFunctionStatement(SyntaxKind.TypeFunctionReturn, undefined, binary());
    const fn = createTypeFunctionDeclaration(undefined, createIdentifier("F"), [], undefined, createTypeFunctionBlock([ret]));
    const source = createSourceFile([fn], createToken(SyntaxKind.EndOfFile), "", toRootedFilePath("/test.ts", undefined), pathKey(toRootedFilePath("/test.ts", undefined), CaseSensitivity.Sensitive));
    const remote = new RemoteSourceFile(encodeSourceFile(source), new TextDecoder());
    const expression = remote.statements!.at(0)!.body!.statements!.at(0)!.expression!;
    assert.ok(isTypeFunctionExpression(expression));
    const order: SyntaxKind[] = [];
    expression.forEachChild(child => {
        order.push(child.kind);
    });
    assert.deepEqual(order, [SyntaxKind.Identifier, SyntaxKind.PlusToken, SyntaxKind.Identifier]);
    const transformed = visitEachChild(expression, child => child.kind === SyntaxKind.Identifier ? createIdentifier("changed") : child);
    assert.equal(transformed.kind, SyntaxKind.TypeFunctionBinary);
    const newRet = createTypeFunctionStatement(SyntaxKind.TypeFunctionReturn, undefined, transformed);
    const next = createSourceFile([createTypeFunctionDeclaration(undefined, createIdentifier("F"), [], undefined, createTypeFunctionBlock([newRet]))], createToken(SyntaxKind.EndOfFile), "", toRootedFilePath("/test.ts", undefined), pathKey(toRootedFilePath("/test.ts", undefined), CaseSensitivity.Sensitive));
    const decoded = new RemoteSourceFile(encodeSourceFile(next), new TextDecoder());
    const result = decoded.statements!.at(0)!.body!.statements!.at(0)!;
    assert.equal(result.kind, SyntaxKind.TypeFunctionReturn);
    assert.equal(result.expression!.kind, SyntaxKind.TypeFunctionBinary);
    assert.equal(result.expression!.operands!.at(1)!.text, "changed");
});

test("DSL child iterators preserve source order and early termination locally and over the wire", () => {
    const expressions = [
        binary(),
        createTypeFunctionExpression(SyntaxKind.TypeFunctionUnary, undefined, [createIdentifier("right")], createToken(SyntaxKind.MinusToken)),
        createTypeFunctionExpression(SyntaxKind.TypeFunctionCall, createIdentifier("F"), [createIdentifier("arg")]),
        createTypeFunctionExpression(SyntaxKind.TypeFunctionNativeType, undefined, undefined, undefined, createKeywordTypeNode(SyntaxKind.StringKeyword)),
    ];
    for (const expression of expressions) {
        const source = createSourceFile(
            [createTypeFunctionDeclaration(
                undefined,
                createIdentifier("F"),
                [],
                undefined,
                createTypeFunctionBlock([
                    createTypeFunctionStatement(SyntaxKind.TypeFunctionReturn, undefined, expression),
                ]),
            )],
            createToken(SyntaxKind.EndOfFile),
            "",
            toRootedFilePath("/test.ts", undefined),
            pathKey(toRootedFilePath("/test.ts", undefined), CaseSensitivity.Sensitive),
        );
        const remote = new RemoteSourceFile(encodeSourceFile(source), new TextDecoder());
        const decoded = remote.statements!.at(0)!.body!.statements!.at(0)!.expression!;
        for (const node of [expression, decoded]) {
            const expected: SyntaxKind[] = [];
            node.forEachChild(child => {
                expected.push(child.kind);
            });
            assert.deepEqual([...node.childrenIter()].map(child => child.kind), expected);
            const iterator = node.childrenIter<string>();
            assert.equal(iterator.next().done, false);
            assert.deepEqual(iterator.next("stop"), { done: true, value: "stop" });
        }
    }
});

test("DSL template children preserve interpolation and literal order", () => {
    const value = binary();
    const span = createTypeFunctionTemplateSpan(value, createTemplateTail("tail", "tail", TokenFlags.None));
    const template = createTypeFunctionTemplate(createTemplateHead("head", "head", TokenFlags.None), [span]);
    const order: SyntaxKind[] = [];
    template.forEachChild(child => {
        order.push(child.kind);
    });
    assert.deepEqual(order, [SyntaxKind.TemplateHead, SyntaxKind.TypeFunctionTemplateSpan]);
    const changed = visitEachChild(span, child => child === value ? binary() : child);
    assert.notEqual(changed, span);
    assert.equal(changed.literal, span.literal);
    assert.equal(changed.expression.kind, SyntaxKind.TypeFunctionBinary);
});

test("match arm and pattern visitors retain native nodes across serialization", () => {
    const capture = createTypeFunctionPattern(SyntaxKind.TypeFunctionCapturePattern, createIdentifier("item"));
    const rest = createTypeFunctionPattern(SyntaxKind.TypeFunctionRestPattern, undefined, capture);
    const tuple = createTypeFunctionPattern(SyntaxKind.TypeFunctionTuplePattern, undefined, undefined, [rest]);
    const arm = createTypeFunctionMatchArm(tuple, undefined, createTypeFunctionBlock([]));
    const match = createTypeFunctionMatch(binary(), [arm]);
    const transformed = visitEachChild(capture, () => createIdentifier("changed"));
    assert.equal(transformed.kind, SyntaxKind.TypeFunctionCapturePattern);
    assert.ok(isIdentifier(transformed.name!));
    assert.equal(transformed.name.text, "changed");
    const source = createSourceFile([createTypeFunctionDeclaration(undefined, createIdentifier("F"), [], undefined, createTypeFunctionBlock([match]))], createToken(SyntaxKind.EndOfFile), "", toRootedFilePath("/test.ts", undefined), pathKey(toRootedFilePath("/test.ts", undefined), CaseSensitivity.Sensitive));
    const remote = new RemoteSourceFile(encodeSourceFile(source), new TextDecoder());
    const decoded = remote.statements!.at(0)!.body!.statements!.at(0)!;
    assert.equal(decoded.kind, SyntaxKind.TypeFunctionMatch);
    assert.equal(decoded.arms!.at(0)!.pattern!.elements!.at(0)!.expression!.name!.text, "item");
});

test("DSL result kinds retain source order and zero constructor arity over the wire", () => {
    for (const slots of [undefined, [], [createIdentifier("_"), createIdentifier("_")]]) {
        const keyword = slots === undefined ? SyntaxKind.TypeKeyword : SyntaxKind.Identifier;
        const kind = createTypeFunctionKind(slots === undefined ? createToken(SyntaxKind.TypeKeyword) : createIdentifier("typefn"), slots);
        const fn = createTypeFunctionDeclaration(undefined, createIdentifier("F"), [], kind, createTypeFunctionBlock([]));
        const source = createSourceFile([fn], createToken(SyntaxKind.EndOfFile), "", toRootedFilePath("/test.ts", undefined), pathKey(toRootedFilePath("/test.ts", undefined), CaseSensitivity.Sensitive));
        const remote = new RemoteSourceFile(encodeSourceFile(source), new TextDecoder());
        for (const value of [fn, remote.statements!.at(0)!]) {
            const order: SyntaxKind[] = [];
            value.forEachChild(child => {
                order.push(child.kind);
            });
            assert.deepEqual(order, [SyntaxKind.Identifier, SyntaxKind.TypeFunctionKind, SyntaxKind.TypeFunctionBlock]);
            assert.equal(value.resultKind!.kindToken!.kind, keyword);
            assert.equal(value.resultKind!.slots?.length, slots?.length);
            const children: SyntaxKind[] = [];
            value.resultKind!.forEachChild(child => {
                children.push(child.kind);
            });
            assert.deepEqual(children, [keyword, ...(slots ?? []).map(() => SyntaxKind.Identifier)]);
        }
        const updated = visitEachChild(kind, child => child === kind.kindToken ? createToken(SyntaxKind.TypeKeyword) : child);
        assert.equal(updated.kindToken!.kind, SyntaxKind.TypeKeyword);
    }
});
