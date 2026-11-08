package checker

import (
	"unsafe"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
)

// VM registers are values, rather than native Type handles. The current execution
// alternatives contain only Types; constructor values will extend this boundary
// without becoming native Types or entering native operator operand lists.
type typefnValue struct {
	native *Type
}

func (v *typefnVM) nativeValue(node *ast.Node, value typefnValue) *Type {
	if value.native == nil {
		return v.error(node, "uninitialized VM value")
	}
	return value.native
}

func typefnValues(types []*Type) []typefnValue {
	values := make([]typefnValue, len(types))
	for i, value := range types {
		values[i] = typefnValue{native: value}
	}
	return values
}

// Deferred native terms retain Type operands. Replaying their expression crosses
// into a fresh VM value frame, charged to the same root as ordinary call frames.
func (v *typefnVM) replayExpression(e *typefnExpression, types []*Type) *Type {
	size := len(types) * int(unsafe.Sizeof(typefnValue{}))
	if size > v.limits.LiveBytes-v.liveBytes {
		return v.error(e.node, "live VM allocation limit exceeded")
	}
	v.liveBytes += size
	defer func() { v.liveBytes -= size }()
	slots := typefnValues(types)
	value := v.expression(e, slots)
	if v.blocked && !v.failed {
		v.blocked = false
		return v.deferExpression(e, slots)
	}
	return value
}
