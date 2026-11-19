package checker

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/diagnostics"
)

type typeFunctionConstructor struct {
	target          *ast.Symbol
	arguments       []*Type
	outer, captured []*Type
}

func typeFunctionSlots(parameter *ast.Node) int {
	if parameter == nil || !ast.IsTypeParameterDeclaration(parameter) {
		return 0
	}
	expr := parameter.AsTypeParameterDeclaration().Expression
	if expr == nil || !ast.IsTypeFunctionConstructorSlots(expr) {
		return 0
	}
	return len(expr.AsTypeFunctionConstructorSlots().Holes.Nodes)
}

func typeFunctionConstructorParameter(symbol *ast.Symbol) int {
	if symbol == nil || symbol.Flags()&ast.SymbolFlagsTypeParameter == 0 {
		return 0
	}
	for _, d := range symbol.Declarations() {
		if n := typeFunctionSlots(d); n != 0 {
			return n
		}
	}
	return 0
}

func (c *Checker) typeFunctionConstructorArity(node *ast.Node) int {
	call := node.Parent
	if call != nil && ast.IsTypeParameterDeclaration(call) && call.AsTypeParameterDeclaration().DefaultType == node {
		return typeFunctionSlots(call)
	}
	if call == nil || !ast.IsTypeReferenceNode(call) {
		return 0
	}
	i := slices.Index(call.TypeArguments(), node)
	if i < 0 {
		return 0
	}
	if i == 1 && c.typeFunctionBuiltinName(call) == "union.toTuple" {
		return 1
	}
	sym := c.getSymbolFromTypeReference(call)
	if sym == nil {
		return 0
	}
	decl := ast.GetDeclarationOfKind(sym, ast.KindTypeFunctionDeclaration)
	if decl == nil || i >= len(decl.TypeParameters()) {
		return 0
	}
	return typeFunctionSlots(decl.TypeParameters()[i])
}

func typeFunctionHole(n *ast.Node) bool {
	return isSimpleIdentifierTypeReference(n) && n.AsTypeReferenceNode().TypeName.Text() == "_"
}

func (c *Checker) readTypeFunctionConstructor(node *ast.Node, arity int) *Type {
	if !ast.IsTypeReferenceNode(node) {
		c.error(node, diagnostics.Type_function_Colon_0, "expected a type constructor")
		return c.errorType
	}
	sym := c.getSymbolFromTypeReference(node)
	if sym == nil || sym.Flags()&(ast.SymbolFlagsTypeAlias|ast.SymbolFlagsInterface|ast.SymbolFlagsClass|ast.SymbolFlagsTypeParameter) == 0 {
		return c.errorType
	}
	var arguments []*Type
	holes := 0
	for _, a := range node.TypeArguments() {
		if typeFunctionHole(a) {
			arguments = append(arguments, nil)
			holes++
		} else {
			arguments = append(arguments, c.getTypeFromTypeNode(a))
		}
	}
	if holes != arity {
		c.error(node, diagnostics.Type_function_Colon_0, "constructor argument has the wrong number of holes")
		return c.errorType
	}
	if typeFunctionConstructorParameter(sym) == arity && holes == len(arguments) {
		return c.getDeclaredTypeOfSymbol(sym)
	}
	if ast.GetDeclarationOfKind(sym, ast.KindTypeFunctionDeclaration) != nil {
		c.error(node, diagnostics.Type_function_Colon_0, "expected an ordinary generic type constructor")
		return c.errorType
	}
	params := c.getLocalTypeParametersOfClassOrInterfaceOrTypeAlias(sym)
	if len(arguments) < c.getMinTypeArgumentCount(params) || len(arguments) > len(params) {
		c.error(node, diagnostics.Type_function_Colon_0, "incorrect constructor arity")
		return c.errorType
	}
	var outer []*Type
	if decl := core.FirstOrNil(sym.Declarations()); decl != nil {
		outer = c.getOuterTypeParameters(decl.Parent, false)
		outer = core.Filter(outer, func(p *Type) bool { return c.isTypeParameterPossiblyReferenced(p, decl) })
	}
	return c.newTypeFunctionConstructor(sym, arguments, outer, outer)
}

func (c *Checker) newTypeFunctionConstructor(sym *ast.Symbol, args, outer, captured []*Type) *Type {
	t := c.newTypeParameter(nil)
	t.AsTypeParameter().constraint = c.noConstraintType
	if c.typeFunctionConstructors == nil {
		c.typeFunctionConstructors = make(map[*Type]*typeFunctionConstructor)
	}
	c.typeFunctionConstructors[t] = &typeFunctionConstructor{sym, args, outer, captured}
	return t
}

func (c *Checker) applyTypeFunctionConstructor(node *ast.Node, t *Type, values []*Type) *Type {
	ctor := c.typeFunctionConstructors[t]
	if ctor == nil {
		return c.failTypeFunction(node, "expected a constructor input")
	}
	args := slices.Clone(ctor.arguments)
	index := 0
	for i, v := range args {
		if v == nil {
			if index == len(values) {
				return c.failTypeFunction(node, "missing constructor argument")
			}
			args[i] = values[index]
			index++
		}
	}
	if index != len(values) {
		return c.failTypeFunction(node, "too many constructor arguments")
	}
	params := c.getLocalTypeParametersOfClassOrInterfaceOrTypeAlias(ctor.target)
	if len(args) < c.getMinTypeArgumentCount(params) || len(args) > len(params) {
		return c.failTypeFunction(node, "incorrect constructor arity")
	}
	args = c.fillMissingTypeArguments(args, params, c.getMinTypeArgumentCount(params), false)
	m := newTypeMapper(params, args)
	for i, p := range params {
		if constraint := c.getConstraintOfTypeParameter(p); constraint != nil && !c.isTypeAssignableTo(args[i], c.instantiateType(constraint, m)) {
			return c.failTypeFunction(node, "constructor argument violates its constraint")
		}
	}
	if ctor.target.Flags()&ast.SymbolFlagsTypeAlias != 0 {
		value := c.getTypeAliasInstantiation(ctor.target, args, nil)
		if len(ctor.outer) != 0 {
			value = c.instantiateType(value, newTypeMapper(ctor.outer, ctor.captured))
		}
		return value
	}
	target := c.getDeclaredTypeOfSymbol(ctor.target)
	if captured := target.AsInterfaceType().OuterTypeParameters(); len(captured) != 0 {
		values := c.instantiateTypes(captured, newTypeMapper(ctor.outer, ctor.captured))
		args = append(slices.Clone(values), args...)
	}
	return c.createTypeReference(target, args)
}
