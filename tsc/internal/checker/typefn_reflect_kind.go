package checker

import "github.com/microsoft/TypeScript/tsc/internal/ast"

func (v *typefnVM) reflectKind(node *ast.Node, name string, value *Type) *Type {
	value = v.observe(value)
	// Substitution can collapse a union/intersection, change a homomorphic mapped
	// Type's category, or normalize a variadic tuple to an array. Ordinary object
	// fields and array elements do not need to be observed to classify the shape.
	if typefnUnresolved(value) || v.c.isGenericObjectType(value) ||
		value.flags&TypeFlagsUnionOrIntersection != 0 && v.unresolvedGraph(node, value) {
		if !v.failed {
			v.blocked = true
		}
		return v.c.errorType
	}
	value = v.c.getReducedType(value)
	switch name {
	case "reflect.isAny":
		return v.boolType(value.flags&TypeFlagsAny != 0)
	case "reflect.isNever":
		return v.boolType(value.flags&TypeFlagsNever != 0)
	case "reflect.isUnknown":
		return v.boolType(value.flags&TypeFlagsUnknown != 0)
	}
	if value.flags&TypeFlagsEnumLike != 0 {
		return v.error(node, "reflect.kind does not classify enum Types or members")
	}
	kind := ""
	switch {
	case value.flags&TypeFlagsAny != 0:
		kind = "any"
	case value.flags&TypeFlagsNever != 0:
		kind = "never"
	case value.flags&TypeFlagsUnknown != 0:
		kind = "unknown"
	case value.flags&TypeFlagsStringLiteral != 0:
		kind = "stringLiteral"
	case value.flags&TypeFlagsNumberLiteral != 0:
		kind = "numberLiteral"
	case value.flags&TypeFlagsBigIntLiteral != 0:
		kind = "bigintLiteral"
	case value.flags&TypeFlagsBooleanLiteral != 0:
		kind = "booleanLiteral"
	case value.flags&TypeFlagsString != 0:
		kind = "string"
	case value.flags&TypeFlagsNumber != 0:
		kind = "number"
	case value.flags&TypeFlagsBigInt != 0:
		kind = "bigint"
	case value.flags&TypeFlagsBoolean != 0:
		kind = "boolean"
	case value.flags&TypeFlagsNull != 0:
		kind = "null"
	case value.flags&TypeFlagsUndefined != 0:
		kind = "undefined"
	case value.flags&TypeFlagsVoid != 0:
		kind = "void"
	case value.flags&TypeFlagsUniqueESSymbol != 0:
		kind = "uniqueSymbol"
	case value.flags&TypeFlagsESSymbol != 0:
		kind = "symbol"
	case isTupleType(value):
		kind = "tuple"
	case v.c.isArrayType(value):
		kind = "array"
	case value.flags&TypeFlagsUnion != 0:
		kind = "union"
	case value.flags&TypeFlagsIntersection != 0:
		kind = "intersection"
	case value.flags&(TypeFlagsObject|TypeFlagsNonPrimitive) != 0:
		kind = "object"
	default:
		return v.error(node, "reflect.kind does not support this closed native category")
	}
	return v.c.getStringLiteralType(kind)
}
