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
	capturedValues := map[*ast.Symbol]bool{}
	addCapture := func(symbol *ast.Symbol, info typefnIslandCapture, expression *typefnExpression) {
		seen := captured
		if info.value {
			seen = capturedValues
		}
		if seen[symbol] {
			return
		}
		seen[symbol] = true
		plan.captures = append(plan.captures, info)
		expression.environment = b.p
		e.args = append(e.args, expression)
	}
	addContext := func(node *ast.Node, value bool) {
		state := b.c.getTypefnState()
		previous := state.definitionCheck
		state.definitionCheck = b.p
		defer func() { state.definitionCheck = previous }()
		var native *Type
		if value {
			native = b.c.checkThisExpression(node)
		} else {
			native = b.c.getThisType(node)
		}
		if !b.p.valid || b.c.isErrorType(native) {
			b.p.valid = false
			return
		}
		if plan.contexts == nil {
			plan.contexts = map[*ast.Node]int{}
		}
		index := len(plan.captures)
		plan.contexts[node] = index
		// Parentheses make the synthetic name unspellable by source identifiers.
		plan.captures = append(plan.captures, typefnIslandCapture{name: fmt.Sprintf("(typefn this %d)", index), context: true, value: value})
		e.args = append(e.args, &typefnExpression{node: node, slot: -1, environment: b.p, nativeContext: native})
	}
	var visit func(*ast.Node, *typefnKind)
	visit = func(node *ast.Node, expected *typefnKind) {
		if node.Kind == ast.KindParenthesizedType {
			visit(node.AsParenthesizedTypeNode().Type, expected)
			return
		}
		if node.Kind == ast.KindThisType {
			// A predicate's `this is` target is syntax, not a captured Type operand.
			if node.Parent == nil || node.Parent.Kind != ast.KindTypePredicate || node.Parent.AsTypePredicateNode().ParameterName != node {
				addContext(node, false)
			}
		}
		if ast.IsTypeQueryNode(node) {
			name := ast.GetFirstIdentifier(node.AsTypeQueryNode().ExprName)
			if ast.IsThisIdentifier(name) {
				addContext(name, true)
			} else {
				symbol := b.c.resolveEntityName(name, ast.SymbolFlagsValue, false, false, nil)
				if symbol == nil || symbol == b.c.unknownSymbol {
					b.p.valid = false
				} else if !core.Some(symbol.Declarations(), func(declaration *ast.Node) bool {
					return isNodeDescendantOf(declaration, e.node.AsTypeFunctionExpression().NativeType)
				}) {
					addCapture(symbol, typefnIslandCapture{name: name.Text(), value: true}, &typefnExpression{node: name, slot: -1, valueSymbol: symbol})
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
				addCapture(symbol, typefnIslandCapture{name: name.Text(), helper: local.helper}, &typefnExpression{node: name, slot: local.slot, kind: local.kind})
			}
		}
		if binding == nil && ast.IsIdentifier(name) {
			for i, parameter := range b.p.nativeOuter {
				if parameter.symbol == symbol && symbol.Flags()&ast.SymbolFlagsTypeParameter != 0 {
					binding = &typefnBinding{slot: b.p.nativeSlots[i], symbol: symbol}
					addCapture(symbol, typefnIslandCapture{name: name.Text()}, &typefnExpression{node: name, slot: binding.slot})
					break
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
			if ast.IsIdentifier(name) {
				addCapture(symbol, typefnIslandCapture{name: name.Text()}, &typefnExpression{node: name, slot: -1, constructorAlias: alias, kind: alias.kind})
			}
		} else if ast.IsIdentifier(name) {
			if code := b.c.typefnNativeConstructor(symbol); code != nil && len(code.outer) != 0 {
				addCapture(symbol, typefnIslandCapture{name: name.Text(), nativeDirect: true}, &typefnExpression{node: name, slot: -1, nativeConstructor: code, kind: typefnConstructorKind(len(code.parameters))})
			}
		}
		if binding == nil && function != nil && len(function.nativeOuter) != 0 && ast.IsIdentifier(name) {
			addCapture(symbol, typefnIslandCapture{name: name.Text(), helper: function}, &typefnExpression{node: name, slot: -1, helper: function, kind: constructorKind})
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
