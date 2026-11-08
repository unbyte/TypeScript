package checker

import "github.com/microsoft/TypeScript/tsc/internal/ast"

// Sealed descriptors become ordinary native signature syntax. In particular,
// predicate and explicit-this behavior is handled by the native checker and
// generated signatures never acquire method-origin bivariance.
func typefnSignatureNode(f *ast.NodeFactory, plan *typefnSignaturePlan, inputs []*ast.Node) *ast.Node {
	var parameters []*ast.Node
	if plan.this != -1 {
		parameters = append(parameters, f.NewParameterDeclaration(nil, nil, f.NewIdentifier("this"), nil, inputs[plan.this], nil))
	}
	for i, parameter := range plan.parameters {
		var optional, rest *ast.Node
		if parameter.flags&ast.CheckFlagsOptionalParameter != 0 {
			optional = f.NewToken(ast.KindQuestionToken)
		}
		if parameter.flags&ast.CheckFlagsRestParameter != 0 {
			rest = f.NewToken(ast.KindDotDotDotToken)
		}
		parameters = append(parameters, f.NewParameterDeclaration(nil, rest, f.NewIdentifier(parameter.name), optional, inputs[i], nil))
	}
	result := inputs[plan.returns]
	if pred := plan.predicate; pred != nil {
		var asserts, narrowed *ast.Node
		if pred.kind == TypePredicateKindAssertsThis || pred.kind == TypePredicateKindAssertsIdentifier {
			asserts = f.NewToken(ast.KindAssertsKeyword)
		}
		var target *ast.Node
		if pred.kind == TypePredicateKindThis || pred.kind == TypePredicateKindAssertsThis {
			target = f.NewThisTypeNode()
		} else {
			target = f.NewIdentifier(plan.parameters[pred.target].name)
		}
		if pred.narrowed != -1 {
			narrowed = inputs[pred.narrowed]
		}
		result = f.NewTypePredicateNode(asserts, target, narrowed)
	}
	if plan.kind == SignatureKindCall {
		return f.NewFunctionTypeNode(nil, f.NewNodeList(parameters), result)
	}
	var modifiers *ast.ModifierList
	if plan.flags&SignatureFlagsAbstract != 0 {
		modifiers = f.NewModifierList([]*ast.Node{f.NewToken(ast.KindAbstractKeyword)})
	}
	return f.NewConstructorTypeNode(modifiers, nil, f.NewNodeList(parameters), result)
}
