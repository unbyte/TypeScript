package checker

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
)

// Resolve captures against original native binders. Native mapped, infer and
// signature binders shadow DSL bindings by symbol identity. Constructor contexts
// follow callee parameter kinds; a constructor is never checked as a native Type.
func (b *typefnCompiler) island(e *typefnExpression) {
	plan := &typefnIsland{}
	e.island = plan
	// Charge the entire cloned tree, including names skipped by kind checking.
	var count func(*ast.Node) bool
	count = func(node *ast.Node) bool { plan.nodes++; node.ForEachChild(count); return false }
	count(e.node.AsTypeFunctionExpression().NativeType)
	captured := map[*ast.Symbol]bool{}
	var visit func(*ast.Node, *typefnKind)
	visit = func(node *ast.Node, expected *typefnKind) {
		if node.Kind == ast.KindParenthesizedType {
			visit(node.AsParenthesizedTypeNode().Type, expected)
			return
		}
		if ast.IsTypeQueryNode(node) {
			name := node.AsTypeQueryNode().ExprName
			if !ast.IsThisIdentifier(name) {
				symbol := b.c.resolveEntityName(name, ast.SymbolFlagsValue, false, false, nil)
				if symbol == nil || symbol == b.c.unknownSymbol {
					b.p.valid = false
				}
			}
		}
		if !ast.IsTypeReferenceNode(node) {
			if expected.resolved() > 0 {
				b.error(node, "constructor argument requires a named constructor or factory application")
			}
			node.ForEachChild(func(child *ast.Node) bool { visit(child, nil); return false })
			return
		}
		name := node.AsTypeReferenceNode().TypeName
		symbol := b.c.resolveEntityName(name, ast.SymbolFlagsType, false, false, nil)
		if symbol == nil || symbol == b.c.unknownSymbol {
			b.p.valid = false
			return
		}
		var binding *typefnBinding
		if ast.IsIdentifier(name) && b.lexicalSymbol(name.Text()) == symbol {
			if local, ok := b.lookup(name.Text()); ok {
				if local.slot < 0 {
					if local.unavailableBound {
						b.error(name, "input bound may reference only preceding Type parameters")
					} else {
						b.error(name, "binding read before initialization")
					}
					return
				}
				binding = &local
				if !captured[symbol] {
					captured[symbol] = true
					plan.names = append(plan.names, name.Text())
					plan.helpers = append(plan.helpers, local.helper)
					e.args = append(e.args, &typefnExpression{node: name, slot: local.slot, kind: local.kind})
				}
			}
		}
		var function *typefnProgram
		var constructorKind *typefnKind
		if binding != nil {
			function = binding.helper
			constructorKind = binding.kind
			if function != nil {
				b.kindDependency(function, false)
			}
		} else if declaration := core.Find(symbol.Declarations(), ast.IsTypeFunctionDeclaration); declaration != nil {
			function = b.compileFunction(declaration, nil)
			if !function.valid {
				b.p.valid = false
			}
			constructorKind = typefnConstructorKind(len(declaration.TypeParameters()))
		} else if alias := b.c.typefnConstructorAlias(symbol); alias != nil {
			constructorKind = alias.kind
		}
		applied := node.TypeArgumentList() != nil
		if !applied && (constructorKind.resolved() > 0 || expected.resolved() > 0) {
			if function != nil {
				b.constructorEligible(node, function)
			}
			if constructorKind.resolved() == 0 && binding == nil {
				if code := b.c.typefnNativeConstructor(symbol); code != nil {
					constructorKind = typefnConstructorKind(len(code.parameters))
				}
			}
			if expected.resolved() == 0 && function != nil {
				b.error(node, "typefn requires an explicit application in a native island")
			} else {
				b.sameKind(node, constructorKind, expected, "constructor value kind or arity is invalid in this native context")
			}
			return
		}
		count, minimum := 0, 0
		var parameters []*typefnKind
		var resultKind *typefnKind
		switch {
		case function != nil:
			parameters = function.parameterKinds
			count = len(parameters)
			minimum = count
			resultKind = function.resultKind
		case constructorKind.resolved() > 0:
			count = constructorKind.resolved() - 1
			minimum = count
		case binding != nil:
			if applied {
				b.error(node, "a captured Type is not a constructor")
				return
			}
		default:
			native := b.c.getLocalTypeParametersOfClassOrInterfaceOrTypeAlias(symbol)
			count = len(native)
			minimum = b.c.getMinTypeArgumentCount(native)
		}
		actual := len(node.TypeArguments())
		if actual < minimum || actual > count {
			b.error(node, fmt.Sprintf("native application requires %d..%d Type arguments", minimum, count))
		}
		b.sameKind(node, resultKind, expected, "constructor value cannot be used as a Type or has the wrong argument kind")
		for i, argument := range node.TypeArguments() {
			var kind *typefnKind
			if i < len(parameters) {
				kind = parameters[i]
			}
			visit(argument, kind)
		}
	}
	visit(e.node.AsTypeFunctionExpression().NativeType, nil)
}
