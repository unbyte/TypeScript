package checker

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/diagnostics"
)

// Native constructor code retains the real declaration and binders. Captures live
// in the common immutable closure record and are mapped by native substitution.
// Merely passing a constructor does not evaluate its body or prove its bounds.
type typefnNativeConstructor struct {
	symbol      *ast.Symbol
	declaration *ast.Node
	parameters  []*Type
	outer       []*Type
}

func (c *Checker) typefnNativeConstructor(symbol *ast.Symbol) *typefnNativeConstructor {
	if symbol.Flags()&(ast.SymbolFlagsTypeAlias|ast.SymbolFlagsClass|ast.SymbolFlagsInterface) == 0 {
		return nil
	}
	state := c.getTypefnState()
	if state.nativeConstructors == nil {
		state.nativeConstructors = map[*ast.Symbol]*typefnNativeConstructor{}
	}
	if code := state.nativeConstructors[symbol]; code != nil {
		return code
	}
	declaration := core.Find(symbol.Declarations(), func(node *ast.Node) bool {
		return ast.IsTypeAliasDeclaration(node) || ast.IsClassLike(node) || ast.IsInterfaceDeclaration(node)
	})
	if declaration == nil {
		return nil
	}
	code := &typefnNativeConstructor{symbol: symbol, declaration: declaration,
		parameters: c.getLocalTypeParametersOfClassOrInterfaceOrTypeAlias(symbol),
		outer:      c.getOuterTypeParameters(declaration, false)}
	state.nativeConstructors[symbol] = code
	return code
}

func (value *typefnClosure) declaration() *ast.Node {
	if value.native != nil {
		return value.native.declaration
	}
	return value.program.declaration
}
func (value *typefnClosure) acceptsArity(arity int) bool {
	if value == nil {
		return false
	}
	if value.native != nil {
		return len(value.native.parameters) == arity
	}
	return len(value.program.parameterKinds) == arity && !typefnHasConstructorParameters(value.program) && value.program.resultKind.resolved() == 0
}

// Only explicit constructor contexts reinterpret native names. Ordinary bare
// defaulted aliases retain native Type semantics, independent of inferred kinds.
func (b *typefnCompiler) expressionInKind(node *ast.Node, expected *typefnKind) *typefnExpression {
	if expected.resolved() <= 0 || node.Kind != ast.KindTypeFunctionName {
		return b.expression(node)
	}
	name := node.Name()
	if ast.IsIdentifier(name) {
		if _, found := b.lookup(name.Text()); found {
			return b.expression(node)
		}
	}
	symbol := b.c.resolveEntityName(name, ast.SymbolFlagsType, false, false, nil)
	if symbol == nil || symbol == b.c.unknownSymbol {
		b.p.valid = false
		return &typefnExpression{node: node, slot: -1}
	}
	if core.Find(symbol.Declarations(), ast.IsTypeFunctionDeclaration) != nil {
		return b.expression(node)
	}
	if b.c.typefnConstructorAlias(symbol) != nil {
		return b.expression(node)
	}
	code := b.c.typefnNativeConstructor(symbol)
	if code == nil {
		b.error(node, "constructor context requires a native generic or eligible typefn")
		return &typefnExpression{node: node, slot: -1}
	}
	return &typefnExpression{node: node, slot: -1, symbol: symbol, nativeConstructor: code, kind: typefnConstructorKind(len(code.parameters))}
}

func (v *typefnVM) applyNativeConstructor(node *ast.Node, closure *typefnClosure, original []*Type) *Type {
	code := closure.native
	if len(original) != len(code.parameters) {
		return v.error(node, "constructor application must supply every declared slot")
	}
	args := slices.Clone(original)
	for i, t := range args {
		args[i] = v.observe(t)
		if v.c.isErrorType(args[i]) {
			v.failed = true
			return v.c.errorType
		}
	}
	outer := make([]*Type, len(closure.captures))
	for i, capture := range closure.captures {
		outer[i] = v.observe(capture.native)
	}
	parameters := append(slices.Clone(code.outer), code.parameters...)
	values := append(slices.Clone(outer), args...)
	mapper := newTypeMapper(parameters, values)
	for i, parameter := range code.parameters {
		if !v.step(node) {
			return v.c.errorType
		}
		if constraint := v.c.getConstraintOfTypeParameter(parameter); constraint != nil {
			bound := v.c.instantiateType(constraint, mapper)
			if v.c.isErrorType(bound) {
				v.failed = true
				return v.c.errorType
			}
			if v.c.isTypeAssignableTo(args[i], bound) {
				continue
			}
			if v.unresolvedGraph(node, args[i]) || v.unresolvedGraph(node, bound) {
				if v.failed {
					return v.c.errorType
				}
				return v.c.newTypefnTerm(&typefnTerm{kind: typefnNativeCallTerm, node: node, symbol: code.symbol, args: original, closure: closure})
			}
			site := core.Coalesce(v.replaySite, core.Coalesce(v.rootSite, node))
			if !v.c.checkTypeAssignableTo(args[i], bound, site, diagnostics.Type_0_does_not_satisfy_the_constraint_1) {
				v.failed = true
				return v.c.errorType
			}
		}
	}
	declared := v.c.getDeclaredTypeOfSymbol(code.symbol)
	if code.symbol.Flags()&ast.SymbolFlagsTypeAlias != 0 {
		if len(code.outer) == 0 || declared == v.c.intrinsicMarkerType {
			if len(code.parameters) == 0 {
				return declared
			}
			return v.c.getTypeAliasInstantiation(code.symbol, args, nil)
		}
		return v.c.instantiateType(declared, mapper)
	}
	if len(parameters) == 0 {
		return declared
	}
	return v.c.createTypeReference(declared, values)
}
