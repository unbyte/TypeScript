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
	// Signature lists still require their separate native member adapter.
	if len(args) > 2 && v.rejectConstruction(node, args[2:]...) {
		return v.c.errorType
	}
	var indexers []typefnIndexDescriptor
	if len(args) > 1 {
		var ok bool
		indexers, ok = v.indexDescriptors(node, args[1])
		if !ok {
			return v.c.errorType
		}
		for _, info := range indexers {
			private = private || constructionOf(info.value) != nil
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
		if len(calls)+len(constructs) != 0 {
			return v.error(node, "recursive object signature adapters are not implemented yet")
		}
		plan := make([]typefnConstructionIndexer, len(indexers))
		for i, info := range indexers {
			plan[i] = typefnConstructionIndexer{info.key, info.readonly}
			values = append(values, info.value)
		}
		return v.construct(node, &typefnConstruction{kind: ast.KindTypeLiteral, inputs: values, properties: fieldsPlan, indexers: plan})
	}
	if !v.allocate(node, len(fieldsPlan)+1) {
		return v.c.errorType
	}
	indexInfos := make([]*IndexInfo, len(indexers))
	for i, info := range indexers {
		indexInfos[i] = v.c.newIndexInfo(info.key, info.value, info.readonly, nil, nil)
	}
	result := v.c.newTypefnObjectEx(v.c.typefnPropertyMembers(fieldsPlan, values), calls, constructs, indexInfos)
	v.validateObjectIndexers(node, result, indexInfos)
	if v.failed {
		return v.c.errorType
	}
	return result
}
