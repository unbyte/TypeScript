package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
)

// Constructor aliases have a value cache separate from native declared Types.
// Their finite kind is discovered without evaluating a factory body. Ordinary
// aliases, including bare defaulted generics, keep the native Type interpretation.
type typefnAlias struct {
	declaration  *ast.Node
	kind         *typefnKind
	value        *typefnClosure
	helper       *typefnProgram // Only synthetic island bindings carry a named combinator signature.
	nativeDirect bool           // A captured native name keeps ordinary default-instantiation semantics.
	evaluating   bool
}

func (c *Checker) typefnAlias(symbol *ast.Symbol) *typefnAlias {
	if c.compilerOptions.GetTypefnVersion() != 1 || symbol == nil || symbol.Flags()&ast.SymbolFlagsTypeAlias == 0 {
		return nil
	}
	declaration := core.Find(symbol.Declarations(), ast.IsTypeAliasDeclaration)
	if declaration == nil {
		return nil
	}
	state := c.getTypefnState()
	if state.aliases == nil {
		state.aliases = map[*ast.Symbol]*typefnAlias{}
	}
	if alias := state.aliases[symbol]; alias != nil {
		return alias
	}
	alias := &typefnAlias{declaration: declaration}
	state.aliases[symbol] = alias
	node := ast.SkipTypeParentheses(declaration.Type())
	if !ast.IsTypeReferenceNode(node) {
		return alias
	}
	target := c.resolveEntityName(node.AsTypeReferenceNode().TypeName, ast.SymbolFlagsType, true, false, nil)
	if target == nil || target == c.unknownSymbol {
		return alias
	}
	if fn := core.Find(target.Declarations(), ast.IsTypeFunctionDeclaration); fn != nil {
		if node.TypeArgumentList() == nil {
			// Code identity and arity do not require evaluating or compiling the
			// body. Publish this kind before following recursive uses of the alias.
			alias.kind = typefnConstructorKind(len(fn.TypeParameters()))
		} else if annotation := ast.TypeFunctionResultKind(fn); annotation != nil {
			if slots := annotation.AsTypeFunctionKind().Slots; slots != nil {
				alias.kind = typefnConstructorKind(len(slots.Nodes))
			}
		} else {
			var parent *typefnProgram
			if len(state.kindStack) != 0 {
				parent = state.kindStack[len(state.kindStack)-1]
			}
			known := state.programs[fn] != nil
			program := c.compileTypefn(fn)
			if parent != nil && program.kindOnStack {
				index := program.kindIndex
				if !known {
					index = program.kindLowlink
				}
				parent.kindLowlink = min(parent.kindLowlink, index)
			}
			alias.kind = program.resultKind
		}
	} else if node.TypeArgumentList() == nil {
		if other := c.typefnAlias(target); other != nil {
			alias.kind = other.kind
		}
	}
	return alias
}
func (c *Checker) typefnConstructorAlias(symbol *ast.Symbol) *typefnAlias {
	alias := c.typefnAlias(symbol)
	if alias != nil && alias.kind.resolved() > 0 {
		return alias
	}
	return nil
}

func (v *typefnVM) constructorAlias(alias *typefnAlias) *typefnClosure {
	if len(alias.declaration.TypeParameters()) != 0 {
		v.error(alias.declaration, "a native constructor alias must be nongeneric; use a typefn factory")
		return nil
	}
	if alias.value != nil {
		return alias.value
	}
	if alias.evaluating {
		v.error(alias.declaration, "recursive constructor alias creation")
		return nil
	}
	alias.evaluating = true
	savedMapper := v.nativeMapper
	v.nativeMapper = nil
	defer func() { alias.evaluating = false; v.nativeMapper = savedMapper }()
	node := ast.SkipTypeParentheses(alias.declaration.Type())
	value := v.nativeConstructorArgument(node, alias.kind.resolved()-1)
	if v.failed || v.blocked {
		return nil
	}
	alias.value = value.closure
	return alias.value
}

func (c *Checker) checkTypefnConstructorAlias(alias *typefnAlias) {
	c.runTypefn(alias.declaration, func(v *typefnVM) *Type {
		v.constructorAlias(alias)
		return c.unknownType
	})
}
func (c *Checker) applyTypefnConstructorAlias(node *ast.Node, alias *typefnAlias) *Type {
	return c.runTypefn(node, func(v *typefnVM) *Type {
		closure := v.constructorAlias(alias)
		if v.failed {
			return c.errorType
		}
		if alias.nativeDirect {
			return v.applyNativeDirect(node, closure, c.getTypeArgumentsFromNode(node))
		}
		if node.TypeArgumentList() == nil {
			return v.error(node, "constructor value cannot be used as a Type")
		}
		if alias.helper != nil {
			args := v.nativeArguments(node, alias.helper)
			if v.failed {
				return c.errorType
			}
			return v.callWithClosure(node, alias.helper.declaration, args, closure)
		}
		if !closure.acceptsArity(len(node.TypeArguments())) {
			return v.error(node, "wrong number of constructor Type arguments")
		}
		args := c.getTypeArgumentsFromNode(node)
		return v.applyConstructor(node, closure, args)
	})
}
func (v *typefnVM) applyConstructor(node *ast.Node, closure *typefnClosure, args []*Type) *Type {
	if closure.native != nil {
		return v.applyNativeConstructor(node, closure, args)
	}
	return v.callWithClosure(node, closure.program.declaration, typefnTypeValues(args), closure)
}
