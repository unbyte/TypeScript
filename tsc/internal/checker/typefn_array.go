package checker

import "github.com/microsoft/TypeScript/tsc/internal/ast"

func (v *typefnVM) arrayType(node *ast.Node, args []*Type) *Type {
	readonly := false
	if len(args) == 2 {
		var ok bool
		readonly, ok = v.boolean(node, args[1])
		if !ok {
			return v.c.errorType
		}
	}
	if !v.allocate(node, 1) {
		return v.c.errorType
	}
	if constructionOf(args[0]) != nil {
		return v.construct(node, &typefnConstruction{kind: ast.KindArrayType, inputs: []*Type{args[0]}, readonly: readonly})
	}
	return v.c.createArrayTypeEx(args[0], readonly)
}

func (v *typefnVM) reflectArray(node *ast.Node, value *Type) *Type {
	if v.c.isGenericObjectType(value) {
		v.blocked = true
		return v.c.errorType
	}
	if !v.c.isArrayType(value) {
		return v.error(node, "reflect.array requires a non-tuple native array")
	}
	return v.record(node,
		typefnRecordField{"element", v.c.getTypeArguments(value)[0]},
		typefnRecordField{"readonly", v.boolType(v.c.isReadonlyArrayType(value))})
}
