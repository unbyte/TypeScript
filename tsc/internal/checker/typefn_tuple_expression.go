package checker

import "github.com/microsoft/TypeScript/tsc/internal/ast"

// Use native variadic normalization instead of iterating a spread's members.
// Spreads preserve optional/label/rest metadata and may expand native unions.
// Unknown spread shapes retain the whole expression so legality is rechecked
// after substitution, including the host's multiple-rest restriction.
func (v *typefnVM) tupleExpression(e *typefnExpression, slots []typefnValue) *Type {
	if !v.allocate(e.node, len(e.args)+1) {
		return v.c.errorType
	}
	values := make([]*Type, 0, len(e.args))
	infos := make([]TupleElementInfo, 0, len(e.args))
	seenRest := false
	for _, element := range e.args {
		item := element
		spread := element.node.Kind == ast.KindTypeFunctionSpread
		if spread {
			item = element.args[0]
		}
		value := v.expression(item, slots)
		if v.failed || v.blocked {
			return v.c.errorType
		}
		flags := ElementFlagsRequired
		if spread {
			if typefnUnresolved(value) || v.c.isGenericTupleType(value) || v.c.isGenericObjectType(value) {
				return v.deferExpression(e, slots)
			}
			if !v.c.isArrayLikeType(value) {
				return v.error(element.node, "tuple spread requires an array or tuple Type")
			}
			flags = ElementFlagsVariadic
			hasRest := v.c.isArrayType(value) || isTupleType(value) && value.TargetTupleType().combinedFlags&ElementFlagsRest != 0
			if hasRest {
				if seenRest {
					return v.error(element.node, "a rest element cannot follow another rest element")
				}
				seenRest = true
			}
			if isTupleType(value) && !v.allocate(element.node, len(v.c.getElementTypes(value))) {
				return v.c.errorType
			}
		}
		values = append(values, value)
		infos = append(infos, TupleElementInfo{flags: flags})
	}
	return v.c.createTupleTypeEx(values, infos, false)
}
