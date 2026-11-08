package checker

import (
	"math"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/jsnum"
)

func (v *typefnVM) objectType(node *ast.Node, args []*Type) *Type {
	properties, ok := v.requiredTuple(node, args[0])
	if !ok {
		return v.c.errorType
	}
	fieldsPlan := []typefnConstructionProperty{}
	values := []*Type{}
	names := map[string]bool{}
	private := false
	for _, property := range properties {
		fields, ok := v.descriptor(node, property, []string{"key", "value", "optional", "readonly"}, []string{"read"})
		if !ok {
			return v.c.errorType
		}
		key, value := fields[0], fields[1]
		if v.rejectConstruction(node, key) {
			return v.c.errorType
		}
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
		if names[name] {
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
		var nameType *Type
		if key.flags == TypeFlagsUniqueESSymbol {
			nameType = key
		}
		fieldsPlan = append(fieldsPlan, typefnConstructionProperty{name: name, key: nameType, optional: optional, readonly: readonly})
		values = append(values, value)
		names[name] = true
		private = private || constructionOf(value) != nil
	}
	// Indexer/signature construction needs its own sealed member adapter. Do
	// not let a private payload reach the existing native builder meanwhile.
	if v.rejectConstruction(node, args[1:]...) {
		return v.c.errorType
	}

	var indexInfos []*IndexInfo
	if len(args) > 1 {
		var ok bool
		indexInfos, ok = v.indexDescriptors(node, args[1])
		if !ok {
			return v.c.errorType
		}
	}
	var calls, constructs []*Signature
	if len(args) > 2 {
		var ok bool
		calls, ok = v.signatureList(node, args[2], SignatureKindCall)
		if !ok {
			return v.c.errorType
		}
	}
	if len(args) > 3 {
		var ok bool
		constructs, ok = v.signatureList(node, args[3], SignatureKindConstruct)
		if !ok {
			return v.c.errorType
		}
	}
	if private {
		if len(indexInfos)+len(calls)+len(constructs) != 0 {
			return v.error(node, "recursive object indexer and signature adapters are not implemented yet")
		}
		return v.construct(node, &typefnConstruction{kind: ast.KindTypeLiteral, inputs: values, properties: fieldsPlan})
	}
	if !v.allocate(node, len(fieldsPlan)+1) {
		return v.c.errorType
	}
	result := v.c.newTypefnObjectEx(v.c.typefnPropertyMembers(fieldsPlan, values), calls, constructs, indexInfos)
	v.validateObjectIndexers(node, result, indexInfos)
	if v.failed {
		return v.c.errorType
	}
	return result
}
