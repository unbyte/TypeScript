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
		args = core.Map(args, v.mapNative)
		closure := v.captureNativeConstructor(code)
		if v.failed {
			return v.c.errorType
		}
		return v.applyNativeDirect(node, closure, args)
	}
	return v.mapNative(v.c.getDeclaredTypeOfSymbol(symbol))
}

// Calls and expression recipes rebuild the same native mapper from their frozen
// register snapshot. No replay depends on an incidental caller environment.
func typefnNativeEnvironment(p *typefnProgram, slots []typefnValue) *TypeMapper {
	if p == nil || len(p.nativeOuter) == 0 {
		return nil
	}
	values := make([]*Type, len(p.nativeOuter))
	for i, slot := range p.nativeSlots {
		values[i] = slots[slot].native
	}
	return newTypeMapper(p.nativeOuter, values)
}
func typefnNativeEnvironmentSize(p *typefnProgram) int {
	if p == nil || len(p.nativeOuter) == 0 {
		return 0
	}
	size := len(p.nativeOuter) * int(unsafe.Sizeof((*Type)(nil)))
	if len(p.nativeOuter) == 1 {
		return size + int(unsafe.Sizeof(SimpleTypeMapper{}))
	}
	return size + int(unsafe.Sizeof(ArrayTypeMapper{}))
}

func (v *typefnVM) applyNativeDirect(node *ast.Node, closure *typefnClosure, args []*Type) *Type {
	code := closure.native
	args = v.c.fillMissingTypeArguments(args, code.parameters, v.c.getMinTypeArgumentCount(code.parameters), false)
	if len(code.outer) != 0 {
		outer := make([]*Type, len(code.outer))
		for i, value := range closure.captures {
			outer[i] = value.native
		}
		mapper := newTypeMapper(code.outer, outer)
		args = core.Map(args, func(t *Type) *Type { return v.c.instantiateType(t, mapper) })
	}
	return v.applyNativeConstructor(node, closure, args)
}
