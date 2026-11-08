package checker

import "github.com/microsoft/TypeScript/tsc/internal/ast"

// VM registers contain either a native Type or a helper closure. Closures never
// become native Types or enter native operator operand lists.
type typefnValue struct {
	native  *Type
	closure *typefnClosure
}

func (v *typefnVM) nativeValue(node *ast.Node, value typefnValue) *Type {
	if value.closure != nil {
		return v.error(node, "constructor value cannot be used as a Type")
	}
	if value.native == nil {
		return v.error(node, "uninitialized VM value")
	}
	return value.native
}

// Deferred expression terms already own immutable value snapshots. Execution
// reads them without copying or permitting assignment to their captured registers.
func (v *typefnVM) replayExpression(e *typefnExpression, slots []typefnValue) *Type {
	value := v.expression(e, slots)
	if v.blocked && !v.failed {
		v.blocked = false
		return v.deferExpression(e, slots)
	}
	return value
}
