package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
)

// Native syntax supplies values according to the declared DSL parameter kinds.
// Constructor arguments bypass the native Type cache and never masquerade as Types.
func (c *Checker) evaluateTypefnReference(node, declaration *ast.Node) *Type {
	return c.runTypefn(node, func(v *typefnVM) *Type {
		p := c.compileTypefn(declaration)
		args := v.nativeArguments(node, p)
		if v.failed {
			return c.errorType
		}
		return v.callWithClosure(node, declaration, args, nil)
	})
}

func (v *typefnVM) nativeArguments(node *ast.Node, p *typefnProgram) []typefnValue {
	nodes := node.TypeArguments()
	if !p.valid {
		v.failed = true
		return nil
	}
	if len(nodes) != len(p.parameterKinds) {
		v.error(node, "wrong number of Type arguments")
		return nil
	}
	args := make([]typefnValue, len(nodes))
	for i, argument := range nodes {
		if kind := p.parameterKinds[i].resolved(); kind > 0 {
			args[i] = v.nativeConstructorArgument(argument, kind-1)
		} else {
			args[i].native = v.c.getTypeFromTypeNode(argument)
		}
		if v.failed {
			return nil
		}
	}
	return args
}

func (v *typefnVM) nativeConstructorArgument(node *ast.Node, arity int) typefnValue {
	if !v.step(node) {
		return typefnValue{}
	}
	if !ast.IsTypeReferenceNode(node) {
		v.error(node, "constructor argument requires a named constructor or factory application")
		return typefnValue{}
	}
	symbol := v.c.resolveEntityName(node.AsTypeReferenceNode().TypeName, ast.SymbolFlagsType, false, false, nil)
	if symbol == nil || symbol == v.c.unknownSymbol {
		v.failed = true
		return typefnValue{}
	}
	if alias := v.c.typefnConstructorAlias(symbol); alias != nil {
		if node.TypeArgumentList() != nil || alias.kind.resolved() != arity+1 {
			v.error(node, "constructor argument kind or arity does not match the parameter")
			return typefnValue{}
		}
		return typefnValue{closure: v.constructorAlias(alias)}
	}
	declaration := core.Find(symbol.Declarations(), ast.IsTypeFunctionDeclaration)
	if declaration == nil {
		native := v.c.typefnNativeConstructor(symbol)
		if node.TypeArgumentList() != nil || native == nil || len(native.parameters) != arity {
			v.error(node, "constructor argument kind or arity does not match the parameter")
			return typefnValue{}
		}
		return typefnValue{closure: v.c.internTypefnConstructor(nil, native, typefnTypeValues(native.outer))}
	}
	p := v.c.compileTypefn(declaration)
	if !p.valid {
		v.failed = true
		return typefnValue{}
	}
	var value typefnValue
	if node.TypeArgumentList() == nil {
		if typefnHasConstructorParameters(p) || p.resultKind.resolved() != 0 {
			v.error(node, "a first-class constructor must accept only Types and return a Type")
			return typefnValue{}
		}
		value.closure = v.c.internTypefnClosure(p, nil)
	} else {
		if p.resultKind.resolved() != arity+1 {
			v.error(node, "constructor argument kind or arity does not match the parameter")
			return typefnValue{}
		}
		args := v.nativeArguments(node, p)
		if v.failed {
			return typefnValue{}
		}
		value = v.callValue(node, declaration, args, nil)
		if v.blocked && !v.failed {
			v.error(node, "constructor identity and creation checks must be ready at a native boundary")
		}
	}
	if !v.failed && !value.closure.acceptsArity(arity) {
		v.error(node, "constructor argument kind or arity does not match the parameter")
	}
	return value
}
