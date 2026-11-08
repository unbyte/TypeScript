package checker

import "github.com/microsoft/TypeScript/tsc/internal/ast"

func (v *typefnVM) indexDescriptors(node *ast.Node, value *Type) ([]*IndexInfo, bool) {
	entries, ok := v.requiredTuple(node, value)
	if !ok {
		return nil, false
	}
	if !v.allocate(node, len(entries)) {
		return nil, false
	}
	infos := make([]*IndexInfo, 0, len(entries))
	for _, entry := range entries {
		fields, ok := v.descriptor(node, entry, []string{"key", "value", "readonly"}, nil)
		if !ok {
			return nil, false
		}
		key := fields[0]
		if typefnUnresolved(key) {
			v.blocked = true
			return nil, false
		}
		if !v.c.isValidIndexKeyType(key) {
			v.error(node, "index key requires a native string, number, symbol or pattern domain")
			return nil, false
		}
		readonly, ok := v.boolean(node, fields[2])
		if !ok {
			return nil, false
		}
		for _, previous := range infos {
			if !v.step(node) {
				return nil, false
			}
			if v.c.isTypeIdenticalTo(previous.keyType, key) {
				v.error(node, "duplicate index domain")
				return nil, false
			}
		}
		infos = append(infos, v.c.newIndexInfo(key, fields[1], readonly, nil, nil))
	}
	return infos, true
}

// Positive native proofs establish legality even for symbolic field Types. A
// failed symbolic proof is a retained requirement, never a concrete rejection.
func (v *typefnVM) requireAssignable(node *ast.Node, source, target *Type, message string) {
	if !v.step(node) {
		return
	}
	source, target = v.observe(source), v.observe(target)
	if v.c.isTypeAssignableTo(source, target) {
		return
	}
	if v.failed {
		return
	}
	if v.unresolvedGraph(node, source) || v.unresolvedGraph(node, target) {
		if !v.failed {
			v.blocked = true
		}
		return
	}
	v.error(node, message)
}

func (v *typefnVM) validateObjectIndexers(node *ast.Node, result *Type, infos []*IndexInfo) {
	if len(infos) == 0 {
		return
	}
	for _, property := range v.c.getPropertiesOfType(result) {
		key := v.c.getLiteralTypeFromProperty(property, TypeFlagsStringOrNumberLiteralOrUnique, true)
		value := v.c.getNonMissingTypeOfSymbol(property)
		for _, info := range infos {
			if !v.step(node) {
				return
			}
			if v.c.isApplicableIndexType(key, info.keyType) {
				v.requireAssignable(node, value, info.valueType, "property Type is incompatible with an applicable indexer")
				if v.failed {
					return
				}
			}
		}
	}
	for _, source := range infos {
		for _, target := range infos {
			if !v.step(node) {
				return
			}
			if source != target && v.c.isApplicableIndexType(source.keyType, target.keyType) {
				v.requireAssignable(node, source.valueType, target.valueType, "indexer value Type is incompatible with an overlapping index domain")
				if v.failed {
					return
				}
			}
		}
	}
}

func (v *typefnVM) reflectIndexers(node *ast.Node, value *Type) *Type {
	if v.c.isGenericObjectType(value) {
		v.blocked = true
		return v.c.errorType
	}
	if value.flags&(TypeFlagsObject|TypeFlagsIntersection) == 0 {
		return v.error(node, "reflect.indexers requires one object Type")
	}
	infos := v.c.getIndexInfosOfType(value)
	if !v.allocate(node, len(infos)+1) {
		return v.c.errorType
	}
	records := make([]*Type, 0, len(infos))
	for _, info := range infos {
		if !v.step(node) {
			return v.c.errorType
		}
		record := v.record(node, typefnRecordField{"key", info.keyType}, typefnRecordField{"value", info.valueType}, typefnRecordField{"readonly", v.boolType(info.isReadonly)})
		if v.failed {
			return v.c.errorType
		}
		records = append(records, record)
	}
	return v.c.createTupleType(records)
}
