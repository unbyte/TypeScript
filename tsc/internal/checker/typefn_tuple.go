package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/jsnum"
)

func (v *typefnVM) tupleIntrinsic(node *ast.Node, name string, args []*Type) *Type {
	t := args[0]
	if !isTupleType(t) || t.TargetTupleType().combinedFlags & ^ElementFlagsRequired != 0 {
		return v.error(node, "tuple operation requires a required finite tuple")
	}
	elements := v.c.getTypeArguments(t)
	if name == "tuple.length" {
		return v.c.getNumberLiteralType(jsnum.Number(len(elements)))
	}
	start, ok := v.intrinsicOffset(node, args[1])
	if !ok {
		return v.c.errorType
	}
	end := int64(len(elements))
	if len(args) == 3 {
		end, ok = v.intrinsicOffset(node, args[2])
		if !ok {
			return v.c.errorType
		}
	}
	if start > end || end > int64(len(elements)) {
		return v.error(node, "tuple slice requires start <= end <= length")
	}
	if !v.allocate(node, int(end-start)+1) {
		return v.c.errorType
	}
	target := t.TargetTupleType()
	return v.c.createTupleTypeEx(elements[start:end], target.elementInfos[start:end], target.readonly)
}
