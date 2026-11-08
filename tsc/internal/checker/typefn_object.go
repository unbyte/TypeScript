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
	}
	var indexers []typefnIndexDescriptor
	if len(args) > 1 {
		var ok bool
		indexers, ok = v.indexDescriptors(node, args[1])
		if !ok {
			return v.c.errorType
		}
	}
	var calls, constructs []typefnSignatureEntry
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
	return v.finishObject(node, fieldsPlan, values, indexers, calls, constructs)
}

func (v *typefnVM) finishObject(node *ast.Node, properties []typefnConstructionProperty, values []*Type, indexers []typefnIndexDescriptor, calls, constructs []typefnSignatureEntry) *Type {
	private := false
	for _, value := range values {
		private = private || constructionOf(value) != nil
	}
	for _, info := range indexers {
		private = private || constructionOf(info.value) != nil
	}
	for _, entry := range calls {
		private = private || entry.signature == nil
	}
	for _, entry := range constructs {
		private = private || entry.signature == nil
	}
	if private {
		plan := make([]typefnConstructionIndexer, len(indexers))
		for i, info := range indexers {
			plan[i] = typefnConstructionIndexer{info.key, info.readonly}
			values = append(values, info.value)
		}
		for _, entry := range calls {
			values = append(values, entry.value)
		}
		for _, entry := range constructs {
			values = append(values, entry.value)
		}
		return v.construct(node, &typefnConstruction{kind: ast.KindTypeLiteral, inputs: values, properties: properties, indexers: plan, calls: len(calls), constructs: len(constructs)})
	}
	if !v.allocate(node, len(properties)+1) {
		return v.c.errorType
	}
	indexInfos := make([]*IndexInfo, len(indexers))
	for i, info := range indexers {
		indexInfos[i] = v.c.newIndexInfo(info.key, info.value, info.readonly, nil, nil)
	}
	callSignatures := make([]*Signature, len(calls))
	for i, entry := range calls {
		callSignatures[i] = entry.signature
	}
	constructSignatures := make([]*Signature, len(constructs))
	for i, entry := range constructs {
		constructSignatures[i] = entry.signature
	}
	result := v.c.newTypefnObjectEx(v.c.typefnPropertyMembers(properties, values), callSignatures, constructSignatures, indexInfos)
	v.validateObjectIndexers(node, result, indexInfos)
	if v.failed {
		return v.c.errorType
	}
	return result
}
