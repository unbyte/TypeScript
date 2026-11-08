/**
 * Hand-written visitor implementations for nodes with runtime-dependent
 * child ordering. Generated code in visitor.generated.ts and factory.generated.ts
 * delegates to these functions.
 */

import { SyntaxKind } from "../../enums/syntaxKind.js";
import type {
    JSDocComment,
    JSDocParameterOrPropertyTag,
    JSDocParameterTag,
    JSDocPropertyTag,
    Node,
    NodeArray,
    TypeFunctionExpression,
} from "./ast.ts";
import {
    createNodeArray,
    updateJSDocParameterTag,
    updateJSDocPropertyTag,
    updateTypeFunctionExpression,
} from "./factory.generated.ts";
import {
    isEntityName,
    isIdentifier,
    isToken,
    isTypeNode,
} from "./is.ts";
import type { Visitor } from "./visitor.generated.ts";
import {
    visitNode,
    visitNodes,
} from "./visitor.generated.ts";

export type { Visitor };
export { visitEachChild, visitNode, visitNodes, visitNodesArray } from "./visitor.generated.ts";

// ── forEachChild helpers (same signature as forEachChildTable entries) ──

function visitNodeForEachChild<T>(cbNode: (node: Node) => T, node: Node | undefined): T | undefined {
    return node ? cbNode(node) : undefined;
}

function visitNodesForEachChild<T>(cbNode: (node: Node) => T, cbNodes: ((nodes: NodeArray<Node>) => T) | undefined, nodes: NodeArray<Node> | undefined): T | undefined {
    if (!nodes) return undefined;
    if (cbNodes) return cbNodes(nodes);
    for (const node of nodes) {
        const result = cbNode(node);
        if (result) return result;
    }
    return undefined;
}

// ── forEachChild implementations ──

function forEachChildOfJSDocParameterOrPropertyTag<T>(data: any, cbNode: (node: Node) => T, cbNodes: ((nodes: NodeArray<Node>) => T) | undefined): T | undefined {
    return visitNodeForEachChild(cbNode, data.tagName) ||
        (data.isNameFirst
            ? visitNodeForEachChild(cbNode, data.name) || visitNodeForEachChild(cbNode, data.typeExpression)
            : visitNodeForEachChild(cbNode, data.typeExpression) || visitNodeForEachChild(cbNode, data.name)) ||
        visitNodesForEachChild(cbNode, cbNodes, data.comment);
}

export { forEachChildOfJSDocParameterOrPropertyTag as forEachChildOfJSDocParameterTag, forEachChildOfJSDocParameterOrPropertyTag as forEachChildOfJSDocPropertyTag };

// ── yieldEachChild implementations ──

function* yieldEachChildOfJSDocParameterOrPropertyTag<T>(data: any): Generator<Node, T | undefined, T> {
    if (data.tagName) {
        const res = yield data.tagName;
        if (res) return res;
    }
    if (data.isNameFirst) {
        if (data.name) {
            const res = yield data.name;
            if (res) return res;
        }
        if (data.typeExpression) {
            const res = yield data.typeExpression;
            if (res) return res;
        }
    }
    else {
        if (data.typeExpression) {
            const res = yield data.typeExpression;
            if (res) return res;
        }
        if (data.name) {
            const res = yield data.name;
            if (res) return res;
        }
    }
    if (data.comment) {
        for (const node of data.comment) {
            const res = yield node;
            if (res) return res;
        }
    }
}

export { yieldEachChildOfJSDocParameterOrPropertyTag as yieldEachChildOfJSDocParameterTag, yieldEachChildOfJSDocParameterOrPropertyTag as yieldEachChildOfJSDocPropertyTag };

// ── visitEachChild implementations ──

function visitEachChildOfJSDocParameterOrPropertyTag(node: JSDocParameterOrPropertyTag, visitor: Visitor): JSDocParameterOrPropertyTag {
    const _tagName = visitNode(node.tagName, visitor, isIdentifier);
    const _name = visitNode(node.name, visitor, isEntityName);
    const _typeExpression = visitNode(node.typeExpression, visitor, isTypeNode);
    const _comment = visitNodes(node.comment, visitor);
    return node.kind === SyntaxKind.JSDocParameterTag
        ? updateJSDocParameterTag(node, _tagName, _name, _typeExpression, _comment)
        : updateJSDocPropertyTag(node, _tagName, _name, _typeExpression, _comment);
}

export { visitEachChildOfJSDocParameterOrPropertyTag as visitEachChildOfJSDocParameterTag, visitEachChildOfJSDocParameterOrPropertyTag as visitEachChildOfJSDocPropertyTag };

// The serialized shape stores operands together; visit the operator between
// them for binary expressions and before them for unary expressions.
export function forEachChildOfTypeFunctionExpression<T>(data: any, cbNode: (node: Node) => T, cbNodes: ((nodes: NodeArray<Node>) => T) | undefined): T | undefined {
    if (data.operatorToken && data.operands?.length === 2) {
        return cbNode(data.operands[0]) || cbNode(data.operatorToken) || cbNode(data.operands[1]);
    }
    return visitNodeForEachChild(cbNode, data.name) || visitNodeForEachChild(cbNode, data.operatorToken) ||
        visitNodesForEachChild(cbNode, cbNodes, data.operands) || visitNodeForEachChild(cbNode, data.nativeType);
}

export function* yieldEachChildOfTypeFunctionExpression<T>(data: any): Generator<Node, T | undefined, T> {
    if (data.operatorToken && data.operands?.length === 2) {
        for (const child of [data.operands.at(0), data.operatorToken, data.operands.at(1)]) {
            const result = yield child;
            if (result) return result;
        }
        return undefined;
    }
    if (data.name) {
        const result = yield data.name;
        if (result) return result;
    }
    if (data.operatorToken) {
        const result = yield data.operatorToken;
        if (result) return result;
    }
    if (data.operands) {
        for (const child of data.operands) {
            const result = yield child;
            if (result) return result;
        }
    }
    if (data.nativeType) {
        const result = yield data.nativeType;
        if (result) return result;
    }
    return undefined;
}

export function visitEachChildOfTypeFunctionExpression(node: TypeFunctionExpression, visitor: Visitor): TypeFunctionExpression {
    const name = visitNode(node.name, visitor);
    let operands = node.operands;
    let operator;
    if (node.operatorToken && operands?.length === 2) {
        const left = visitNode(operands[0], visitor)!;
        operator = visitNode(node.operatorToken, visitor, isToken);
        const right = visitNode(operands[1], visitor)!;
        if (left !== operands[0] || right !== operands[1]) {
            operands = createNodeArray([left, right], operands.pos, operands.end);
        }
    }
    else {
        operator = visitNode(node.operatorToken, visitor, isToken);
        operands = visitNodes(operands, visitor);
    }
    return updateTypeFunctionExpression(node, name, operands, operator, visitNode(node.nativeType, visitor, isTypeNode));
}
