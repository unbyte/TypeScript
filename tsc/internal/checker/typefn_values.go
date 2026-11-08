package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
)

func (v *typefnVM) boolType(value bool) *Type {
	if value {
		return v.c.regularTrueType
	}
	return v.c.regularFalseType
}

func (v *typefnVM) boolean(node *ast.Node, value *Type) (bool, bool) {
	value = v.observe(value)
	if typefnUnresolved(value) {
		v.blocked = true
		return false, false
	}
	if value.flags == TypeFlagsBooleanLiteral {
		return value.AsLiteralType().value.(bool), true
	}
	v.error(node, "condition requires one boolean literal")
	return false, false
}

func (v *typefnVM) elements(node *ast.Node, value *Type, union bool) []*Type {
	value = v.observe(value)
	if typefnUnresolved(value) {
		v.blocked = true
		return nil
	}
	if union {
		// R2 waits for reachable free dependencies even when the outer partition
		// is already known. The summary avoids forcing recursive native members.
		if v.unresolvedGraph(node, value) {
			if !v.failed {
				v.blocked = true
			}
			return nil
		}
		if value.flags&TypeFlagsNever != 0 {
			return nil
		}
		if value.flags&TypeFlagsUnion != 0 {
			return value.Types()
		}
		return []*Type{value}
	}
	if value.flags == TypeFlagsStringLiteral {
		units, ok := v.stringUnitsOf(node, value)
		if !ok {
			return nil
		}
		if !v.allocate(node, len(units)+1) {
			return nil
		}
		values := make([]*Type, len(units))
		for i := range units {
			values[i] = v.textResult(node, units[i:i+1])
			if v.failed {
				return nil
			}
		}
		return values
	}
	if !isTupleType(value) || value.TargetTupleType().combinedFlags & ^ElementFlagsRequired != 0 {
		v.error(node, "iteration/spread requires a required finite tuple or a string literal")
		return nil
	}
	return v.c.getTypeArguments(value)
}
