package ast

// Typefn nodes use the native AST, lexical scopes and symbol infrastructure.
// Keep the boundary explicit: DSL operators are not JavaScript expressions.
func IsTypeFunctionNode(node *Node) bool {
	return node != nil && node.Kind >= KindTypeFunctionDeclaration && node.Kind <= KindTypeFunctionPinPattern
}

func IsTypeFunctionIteration(node *Node) bool {
	return node != nil && (node.Kind == KindTypeFunctionForOf || node.Kind == KindTypeFunctionForIn)
}

func IsTypeFunctionLocalDeclaration(node *Node) bool {
	return node != nil && (node.Kind == KindTypeFunctionBinding || node.Kind == KindTypeFunctionCapturePattern || IsTypeFunctionIteration(node))
}

// TypeFunctionReference returns the whole qualified Type name for a DSL read,
// application or assignment. A declaration name is handled by normal binding.
func TypeFunctionReference(node *Node) *Node {
	if node == nil {
		return nil
	}
	name := node
	for name.Parent != nil && name.Parent.Kind == KindQualifiedName {
		name = name.Parent
	}
	if name.Parent == nil {
		return nil
	}
	switch name.Parent.Kind {
	case KindTypeFunctionName, KindTypeFunctionCall, KindTypeFunctionAssignment:
		if name.Parent.Name() == name {
			return name
		}
	}
	return nil
}

// Operators must be visited in source order. Operands is a compact internal
// representation, not a syntactic list enclosing the operator.
func forEachChild_TypeFunctionExpression(node *TypeFunctionExpression, v Visitor) bool {
	if node.Kind == KindTypeFunctionBinary {
		return visit(v, node.Operands.Nodes[0]) || visit(v, node.OperatorToken) || visit(v, node.Operands.Nodes[1])
	}
	return visit(v, node.name) || visit(v, node.OperatorToken) || visitNodeList(v, node.Operands) || visit(v, node.NativeType)
}

func visitEachChild_TypeFunctionExpression(node *TypeFunctionExpression, v *NodeVisitor) *Node {
	name := v.visitNode(node.name)
	var operator *Node
	operands := node.Operands
	if node.Kind == KindTypeFunctionBinary {
		left := v.visitNode(operands.Nodes[0])
		operator = v.visitNode(node.OperatorToken)
		right := v.visitNode(operands.Nodes[1])
		if left != operands.Nodes[0] || right != operands.Nodes[1] {
			operands = v.Factory.NewNodeList([]*Node{left, right})
			operands.Loc = node.Operands.Loc
		}
	} else {
		operator = v.visitNode(node.OperatorToken)
		operands = v.visitNodes(operands)
	}
	return v.Factory.UpdateTypeFunctionExpression(node, name, operands, operator, v.visitNode(node.NativeType), node.Readonly, node.Optional)
}

// Captures are in scope in the guard and body, never in the arm's patterns.
// In particular, pins and Type expressions refer to the enclosing environment.
func TypeFunctionScopeVisible(scope, child *Node) bool {
	return scope.Kind != KindTypeFunctionMatchArm || child != scope.AsTypeFunctionMatchArm().Pattern
}
