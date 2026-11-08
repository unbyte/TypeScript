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
	v.error(node, "condition requires one boolean literal; deferred control flow is not implemented yet")
	return false, false
}

func (v *typefnVM) elements(node *ast.Node, value *Type, union bool) []*Type {
	value = v.observe(value)
	if typefnUnresolved(value) {
		v.blocked = true
		return nil
	}
	if union {
		// Conservative until the full graph-readiness protocol is implemented.
		if !typefnClosedScalar(value) {
			v.error(node, "union iteration currently requires closed scalar members")
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
	if !isTupleType(value) || value.TargetTupleType().combinedFlags & ^ElementFlagsRequired != 0 {
		v.error(node, "iteration/spread requires a required finite tuple; string iteration is not implemented yet")
		return nil
	}
	return v.c.getTypeArguments(value)
}

func typefnClosedScalar(t *Type) bool {
	if t.flags&TypeFlagsUnion != 0 {
		for _, part := range t.Types() {
			if !typefnClosedScalar(part) {
				return false
			}
		}
		return true
	}
	return t.flags & ^(TypeFlagsAny|TypeFlagsUnknown|TypeFlagsNever|TypeFlagsString|TypeFlagsNumber|TypeFlagsBigInt|TypeFlagsBoolean|TypeFlagsStringLiteral|TypeFlagsNumberLiteral|TypeFlagsBigIntLiteral|TypeFlagsBooleanLiteral|TypeFlagsNull|TypeFlagsUndefined|TypeFlagsVoid) == 0
}
