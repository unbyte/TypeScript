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
	size := typefnNativeEnvironmentSize(e.environment)
	if !v.reserveValues(e.node, size) {
		return v.c.errorType
	}
	savedMapper := v.nativeMapper
	v.nativeMapper = typefnNativeEnvironment(e.environment, slots)
	defer func() { v.nativeMapper = savedMapper; v.liveBytes -= size }()
	value := v.expression(e, slots)
	if v.blocked && !v.failed {
		v.blocked = false
		return v.deferExpression(e, slots)
	}
	return value
}

// Store and pass a Type result with the checks reached while computing it.
// The caller journal still keeps those checks if the value is later discarded.
// A factory argument must not lose them merely because its native shape is ready.
func (v *typefnVM) resultExpression(e *typefnExpression, slots []typefnValue) typefnValue {
	saved, savedSet := v.checks, v.checkSet
	v.checks, v.checkSet = nil, nil
	value := v.valueExpression(e, slots)
	reached := v.checks
	if value.native != nil {
		value.native = v.c.guardTypefn(value.native, reached)
	}
	v.checks, v.checkSet = saved, savedSet
	for _, check := range reached {
		v.observe(check)
	}
	if !v.failed && value.closure != nil && len(reached) != 0 {
		v.blocked = true
		return typefnValue{}
	}
	return value
}

func (v *typefnVM) typeArgument(e *typefnExpression, slots []typefnValue) *Type {
	value := v.resultExpression(e, slots)
	if v.failed || v.blocked {
		return v.c.errorType
	}
	return v.nativeValue(e.node, value)
}

func typefnTypeValues(types []*Type) []typefnValue {
	values := make([]typefnValue, len(types))
	for i, t := range types {
		values[i].native = t
	}
	return values
}
