package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/scanner"
)

func (f *typeFunctionFrame) binaryOperand(b *ast.TypeFunctionBinary) *Type {
	c := f.c
	left := f.operand(b.Left)
	if c.typeFunctionExecution.failed {
		return c.anyType
	}
	op := b.OperatorToken.Kind
	if op == ast.KindBarBarToken || op == ast.KindAmpersandAmpersandToken {
		if left != c.regularTrueType && left != c.regularFalseType {
			return c.failTypeFunction(b.Left, "expected a boolean literal")
		}
		if op == ast.KindAmpersandAmpersandToken && left == c.regularFalseType || op == ast.KindBarBarToken && left == c.regularTrueType {
			return left
		}
		right := f.operand(b.Right)
		if right != c.regularTrueType && right != c.regularFalseType {
			return c.failTypeFunction(b.Right, "expected a boolean literal")
		}
		return right
	}
	right := f.operand(b.Right)
	if c.typeFunctionExecution.failed {
		return c.anyType
	}
	args := []*Type{left, right}
	if !c.chargeTypeFunctionOperands(b.OperatorToken, args) {
		return c.anyType
	}
	switch op {
	case ast.KindIsKeyword:
		return f.boolean(c.isTypeAssignableTo(left, right))
	case ast.KindEqualsEqualsEqualsToken:
		return f.boolean(c.isTypeIdenticalTo(left, right))
	case ast.KindExclamationEqualsEqualsToken:
		return f.boolean(!c.isTypeIdenticalTo(left, right))
	case ast.KindPlusToken:
		return c.typeFunctionNumericOperator(b.OperatorToken, "add", args)
	case ast.KindMinusToken:
		return c.typeFunctionNumericOperator(b.OperatorToken, "subtract", args)
	case ast.KindAsteriskToken:
		return c.typeFunctionNumericOperator(b.OperatorToken, "multiply", args)
	case ast.KindSlashToken:
		return c.typeFunctionNumericOperator(b.OperatorToken, "divide", args)
	case ast.KindPercentToken:
		return c.typeFunctionNumericOperator(b.OperatorToken, "remainder", args)
	default:
		panic("unexpected type function operator")
	}
}

// A native type can retain a nested expression across instantiation. Capture its
// lexical inputs rather than computing its operands here: an unselected
// conditional arm or the right side of a short-circuit may never run.
func (c *Checker) getTypeFromTypeFunctionExpression(node *ast.Node) *Type {
	links := c.typeNodeLinks.Get(node)
	if links.resolvedType == nil {
		params := core.Filter(c.getOuterTypeParameters(node, true), func(p *Type) bool {
			return c.isTypeParameterPossiblyReferenced(p, node)
		})
		links.resolvedType = c.newTypeFunctionApplication(typeFunctionApplication{use: node, expression: node, arguments: params})
	}
	return links.resolvedType
}

// Arithmetic uses the same literal representation and operations as the named
// builtins. Report failures at the operator as it appears in the program.
func (c *Checker) typeFunctionNumericOperator(node *ast.Node, operation string, args []*Type) *Type {
	token := node.Kind
	if token == ast.KindTypeFunctionNegate {
		token = ast.KindMinusToken
	}
	name := scanner.TokenToString(token)
	kind := args[0].flags & (TypeFlagsNumberLiteral | TypeFlagsBigIntLiteral)
	if len(args) == 1 && kind == 0 {
		return c.failTypeFunction(node, "operator '-' requires a numeric literal")
	}
	if kind == 0 || len(args) == 2 && args[1].flags&kind == 0 {
		return c.failTypeFunction(node, "operator '"+name+"' requires numeric literals of the same kind")
	}
	group := "number"
	if kind == TypeFlagsBigIntLiteral {
		group = "bigint"
	}
	if result := c.typeFunctionArithmetic(group, operation, args); result != nil {
		return result
	}
	return c.failTypeFunction(node, "operator '"+name+"' has no finite numeric result")
}

// Display an open expression using its instantiated lexical bindings. Walking the
// syntax leaves unevaluated operands alone, including calls on a short-circuit path.
func (b *NodeBuilderImpl) typeFunctionExpressionToNode(app *typeFunctionApplication) *ast.Node {
	params := b.ch.typeNodeLinks.Get(app.expression).resolvedType.typeFunctionApplication().arguments
	var visitor *ast.NodeVisitor
	visitor = ast.NewNodeVisitor(func(node *ast.Node) *ast.Node {
		if isSimpleIdentifierTypeReference(node) {
			symbol := b.ch.getSymbolFromTypeReference(node)
			for i, param := range params {
				if param.symbol == symbol {
					return b.typeToTypeNode(app.arguments[i])
				}
			}
		}
		return visitor.VisitEachChild(node)
	}, b.f, ast.NodeVisitorHooks{})
	return b.f.DeepCloneNode(visitor.VisitNode(app.expression))
}
