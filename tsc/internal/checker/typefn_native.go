package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"unsafe"
)

func (v *typefnVM) mapNative(t *Type) *Type {
	if v.nativeMapper == nil {
		return t
	}
	return v.c.instantiateType(t, v.nativeMapper)
}

func (v *typefnVM) captureNativeConstructor(code *typefnNativeConstructor) *typefnClosure {
	size := len(code.outer) * int(unsafe.Sizeof(typefnValue{}))
	if !v.reserveValues(code.declaration, size) {
		return nil
	}
	defer func() { v.liveBytes -= size }()
	captures := make([]typefnValue, len(code.outer))
	for i, parameter := range code.outer {
		if !v.step(code.declaration) {
			return nil
		}
		captures[i].native = v.mapNative(parameter)
	}
	return v.c.internTypefnConstructor(nil, code, captures)
}

// Direct native applications retain native default behavior, but share bound
// validation and lexical substitution with first-class constructor application.
func (v *typefnVM) apply(node *ast.Node, symbol *ast.Symbol, args []*Type) *Type {
	if declaration := core.Find(symbol.Declarations(), ast.IsTypeFunctionDeclaration); declaration != nil {
		return v.call(node, declaration, args)
	}
	if code := v.c.typefnNativeConstructor(symbol); code != nil {
		args = v.c.fillMissingTypeArguments(args, code.parameters, v.c.getMinTypeArgumentCount(code.parameters), false)
		args = core.Map(args, v.mapNative)
		closure := v.captureNativeConstructor(code)
		if v.failed {
			return v.c.errorType
		}
		return v.applyNativeConstructor(node, closure, args)
	}
	return v.mapNative(v.c.getDeclaredTypeOfSymbol(symbol))
}
