package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/jsnum"
	"math"
)

func (v *typefnVM) objectType(node *ast.Node, args []*Type) *Type {
	properties, ok := v.requiredTuple(node, args[0])
	if !ok {
		return v.c.errorType
	}
	for _, argument := range args[1:] {
		entries, ok := v.requiredTuple(node, argument)
		if !ok {
			return v.c.errorType
		}
		if len(entries) != 0 {
			return v.error(node, "nonempty object indexers/signatures are not implemented yet")
		}
	}
	members := ast.NewSymbolTable()
	for _, property := range properties {
		fields, ok := v.descriptor(node, property, []string{"key", "value", "optional", "readonly"}, []string{"read"})
		if !ok {
			return v.c.errorType
		}
		key, value := fields[0], fields[1]
		if typefnUnresolved(key) {
			v.blocked = true
			return v.c.errorType
		}
		if key.flags != TypeFlagsStringLiteral && key.flags != TypeFlagsNumberLiteral && key.flags != TypeFlagsUniqueESSymbol {
			return v.error(node, "property key requires one string/number literal or unique symbol")
		}
		if key.flags == TypeFlagsNumberLiteral {
			n := float64(key.AsLiteralType().value.(jsnum.Number))
			if math.IsInf(n, 0) || math.IsNaN(n) {
				return v.error(node, "number property key must be finite")
			}
		}
		name := getPropertyNameFromType(key)
		if members.Get(name) != nil {
			return v.error(node, "duplicate effective property key")
		}
		optional, ok := v.boolean(node, fields[2])
		if !ok {
			return v.c.errorType
		}
		readonly, ok := v.boolean(node, fields[3])
		if !ok {
			return v.c.errorType
		}
		if !v.allocate(node, 1) {
			return v.c.errorType
		}
		flags := ast.SymbolFlagsProperty
		if optional {
			flags |= ast.SymbolFlagsOptional
		}
		checkFlags := ast.CheckFlagsNone
		if readonly {
			checkFlags |= ast.CheckFlagsReadonly
		}
		symbol := v.c.newSymbolEx(flags, name, checkFlags)
		links := v.c.valueSymbolLinks.Get(symbol)
		links.resolvedType = v.c.addOptionalityEx(value, true, optional)
		if key.flags == TypeFlagsUniqueESSymbol {
			links.setNameType(v.c, key)
		}
		members.Set(name, symbol)
	}
	if !v.allocate(node, 1) {
		return v.c.errorType
	}
	return v.c.newTypefnObject(members)
}
