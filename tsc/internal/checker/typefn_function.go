package checker

import "github.com/microsoft/TypeScript/tsc/internal/ast"

// A native rest-tuple signature retains labels, optionality and open/variadic
// tails. The native checker expands it for calls, inference and presentation.
// No synthetic method declaration is attached: ordinary function variance applies.
func (v *typefnVM) functionType(node *ast.Node, args []*Type, kind SignatureKind) *Type {
	parameters := v.observe(args[0])
	graph := constructionOf(parameters)
	if graph != nil && graph.kind != ast.KindTupleType {
		return v.error(node, "function parameters require one tuple Type")
	}
	if graph == nil && typefnUnresolved(parameters) {
		v.blocked = true
		return v.c.errorType
	}
	if graph == nil && !isTupleType(parameters) {
		return v.error(node, "function parameters require one tuple Type")
	}
	flags := SignatureFlagsHasRestParameter
	if kind == SignatureKindConstruct {
		flags |= SignatureFlagsConstruct
	}
	if len(args) == 3 {
		abstract, ok := v.boolean(node, args[2])
		if !ok {
			return v.c.errorType
		}
		if abstract {
			flags |= SignatureFlagsAbstract
		}
	}
	if graph != nil || constructionOf(args[1]) != nil {
		shape := ast.KindFunctionType
		if kind == SignatureKindConstruct {
			shape = ast.KindConstructorType
		}
		return v.construct(node, &typefnConstruction{kind: shape, inputs: []*Type{parameters, args[1]}, abstract: flags&SignatureFlagsAbstract != 0})
	}
	if !v.allocate(node, 3) {
		return v.c.errorType
	}
	parameter := v.c.newSymbolEx(ast.SymbolFlagsFunctionScopedVariable, "args", ast.CheckFlagsRestParameter)
	v.c.valueSymbolLinks.Get(parameter).resolvedType = parameters
	signature := v.c.newSignature(flags, nil, nil, nil, []*ast.Symbol{parameter}, args[1], v.c.noTypePredicate, 0)
	return v.c.typefnSignatureType(signature, kind)
}
