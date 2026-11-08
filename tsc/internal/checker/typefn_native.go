package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/diagnostics"
	"slices"
)

func (v *typefnVM) constraints(node *ast.Node, parameters, args []*Type) bool {
	mapper := newTypeMapper(parameters, args)
	for i, parameter := range parameters {
		if constraint := v.c.getConstraintOfTypeParameter(parameter); constraint != nil {
			if v.c.isErrorType(constraint) {
				v.failed = true
				return false
			}
			bound := v.c.instantiateType(constraint, mapper)
			if !v.c.isTypeAssignableTo(args[i], bound) && (typefnUnresolved(args[i]) || typefnUnresolved(bound)) {
				v.blocked = true
				return false
			}
			site := node
			if v.replaySite != nil {
				site = v.replaySite
			}
			if !v.c.checkTypeAssignableTo(args[i], bound, site, diagnostics.Type_0_does_not_satisfy_the_constraint_1) {
				v.failed = true
				return false
			}
		}
	}
	return true
}

func (v *typefnVM) apply(node *ast.Node, symbol *ast.Symbol, args []*Type) *Type {
	args = slices.Clone(args)
	for i, arg := range args {
		args[i] = v.observe(arg)
	}
	if declaration := core.Find(symbol.Declarations(), ast.IsTypeFunctionDeclaration); declaration != nil {
		return v.call(node, declaration, args)
	}
	t := v.c.getDeclaredTypeOfSymbol(symbol)
	parameters := v.c.getLocalTypeParametersOfClassOrInterfaceOrTypeAlias(symbol)
	args = v.c.fillMissingTypeArguments(args, parameters, v.c.getMinTypeArgumentCount(parameters), false)
	for i, arg := range args {
		args[i] = v.observe(arg)
	}
	if !v.constraints(node, parameters, args) {
		if v.blocked && !v.failed {
			v.blocked = false
			return v.c.newTypefnTerm(&typefnTerm{kind: typefnNativeCallTerm, node: node, symbol: symbol, args: args})
		}
		return v.c.errorType
	}
	if len(parameters) == 0 {
		return t
	}
	if symbol.Flags()&ast.SymbolFlagsTypeAlias != 0 {
		return v.c.getTypeAliasInstantiation(symbol, args, nil)
	}
	if symbol.Flags()&(ast.SymbolFlagsClass|ast.SymbolFlagsInterface) != 0 {
		all := append(append([]*Type(nil), t.AsInterfaceType().OuterTypeParameters()...), args...)
		return v.c.createTypeReference(t, all)
	}
	return v.error(node, "unsupported native constructor")
}
