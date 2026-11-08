package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
)

// Render deferred calculations as DSL expressions, rather than invented native
// generic names. This is display-only; declaration recipes are still unsupported.
func (b *NodeBuilderImpl) typefnTypeToTypeNode(t *Type) *ast.Node {
	b.typefnDisplayDepth++
	defer func() { b.typefnDisplayDepth-- }()
	b.ctx.approximateLength += 4
	if b.typefnDisplayDepth > 64 || b.checkTruncationLength() {
		return b.createElidedInformationPlaceholder()
	}
	d := t.asTypefnTerm()
	switch d.kind {
	case typefnCallTerm:
		arguments := core.Map(d.values, b.typefnValueToNode)
		if ast.IsTypeFunctionLambda(d.declaration) {
			return b.f.NewTypeReferenceNode(b.f.NewIdentifier("typefn closure"), b.f.NewNodeList(arguments))
		}
		return b.symbolToTypeNode(d.declaration.Symbol(), ast.SymbolFlagsType, b.f.NewNodeList(arguments))
	case typefnNativeCallTerm:
		return b.symbolToTypeNode(d.symbol, ast.SymbolFlagsType, b.f.NewNodeList(core.Map(d.args, b.typeToTypeNode)))
	case typefnGuardTerm:
		// Type display strips comments, so use an explicit diagnostic label.
		return b.f.NewTypeReferenceNode(b.f.NewIdentifier("pending validation"), b.f.NewNodeList([]*ast.Node{b.typeToTypeNode(d.shape)}))
	case typefnExpressionTerm:
		return b.typefnExpressionToNode(d.expression, d.values)
	case typefnBinaryTerm, typefnUnaryTerm:
		kind := ast.KindTypeFunctionBinary
		if d.kind == typefnUnaryTerm {
			kind = ast.KindTypeFunctionUnary
		}
		operands := core.Map(d.args, func(t *Type) *ast.Node { return b.typefnDisplayLiteral(t) })
		return b.f.NewTypeFunctionExpression(kind, nil, b.f.NewNodeList(operands), b.f.NewToken(d.operator), nil, false, false)
	}
	panic("unknown typefn display")
}

func (b *NodeBuilderImpl) typefnDisplayLiteral(t *Type) *ast.Node {
	return b.f.NewTypeFunctionExpression(ast.KindTypeFunctionLiteral, nil, nil, nil, b.typeToTypeNode(t), false, false)
}

func (b *NodeBuilderImpl) typefnExpressionToNode(e *typefnExpression, slots []typefnValue) *ast.Node {
	if ast.IsTypeFunctionLambda(e.node) {
		return e.node
	}
	if e.slot >= 0 {
		value := slots[e.slot]
		if value.closure != nil {
			name := e.node.Name()
			if e.node.Kind == ast.KindTypeFunctionCall {
				args := core.Map(e.args, func(arg *typefnExpression) *ast.Node { return b.typefnExpressionToNode(arg, slots) })
				return b.f.NewTypeFunctionExpression(ast.KindTypeFunctionCall, name, b.f.NewNodeList(args), nil, nil, false, false)
			}
			return b.f.NewTypeFunctionExpression(ast.KindTypeFunctionName, name, nil, nil, nil, false, false)
		}
		return b.typefnDisplayLiteral(value.native)
	}
	if e.node.Kind == ast.KindTypeFunctionTemplate {
		template := e.node.AsTypeFunctionTemplate()
		spans := make([]*ast.Node, len(e.args))
		for i, arg := range e.args {
			spans[i] = b.f.NewTypeFunctionTemplateSpan(b.typefnExpressionToNode(arg, slots), template.Spans.Nodes[i].AsTypeFunctionTemplateSpan().Literal)
		}
		return b.f.NewTypeFunctionTemplate(template.Head, b.f.NewNodeList(spans))
	}
	data := e.node.AsTypeFunctionExpression()
	var operands []*ast.Node
	if e.node.Kind == ast.KindTypeFunctionObject {
		for i, field := range data.Operands.Nodes {
			fd := field.AsTypeFunctionExpression()
			name := field.Name()
			if e.keys[i] != nil {
				name = b.typefnExpressionToNode(e.keys[i], slots)
			}
			value := b.typefnExpressionToNode(e.args[i], slots)
			operands = append(operands, b.f.NewTypeFunctionExpression(ast.KindTypeFunctionField, name, b.f.NewNodeList([]*ast.Node{value}), nil, nil, fd.Readonly, fd.Optional))
		}
	} else {
		operands = core.Map(e.args, func(arg *typefnExpression) *ast.Node { return b.typefnExpressionToNode(arg, slots) })
	}
	return b.f.NewTypeFunctionExpression(e.node.Kind, e.node.Name(), b.f.NewNodeList(operands), data.OperatorToken, data.NativeType, data.Readonly, data.Optional)
}

func (b *NodeBuilderImpl) typefnValueToNode(value typefnValue) *ast.Node {
	if value.native != nil {
		return b.typeToTypeNode(value.native)
	}
	if value.closure != nil {
		if value.closure.native != nil {
			return b.symbolToTypeNode(value.closure.native.symbol, ast.SymbolFlagsType, nil)
		}
		node := value.closure.program.declaration
		if node.Name() != nil {
			return b.symbolToTypeNode(node.Symbol(), ast.SymbolFlagsType, nil)
		}
	}
	return b.f.NewTypeReferenceNode(b.f.NewIdentifier("typefn closure"), nil)
}
