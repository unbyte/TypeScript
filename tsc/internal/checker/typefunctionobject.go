package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
)

func typeFunctionPublicField(p *ast.Symbol) bool {
	return getDeclarationModifierFlagsFromSymbol(p)&ast.ModifierFlagsNonPublicAccessibilityModifier == 0 &&
		(p.ValueDeclaration() == nil || !ast.IsPrivateIdentifierClassElementDeclaration(p.ValueDeclaration()))
}

func (c *Checker) describeTypeFunctionField(p *ast.Symbol) *Type {
	read := c.getTypeOfSymbol(p)
	readonly := c.isReadonlySymbol(p)
	write := c.createTupleType([]*Type{c.getStringLiteralType("readonly")})
	if !readonly {
		write = c.createTupleType([]*Type{c.getStringLiteralType("accepts"), c.getWriteTypeOfSymbol(p)})
	}
	kind := "data"
	parts := []typeFunctionField{{"read", read}, {"write", write}, {"optional", c.typeFunctionBoolean(p.Flags()&ast.SymbolFlagsOptional != 0)}}
	if p.Flags()&ast.SymbolFlagsAccessor != 0 {
		kind = "accessor"
		var get, set []*Type
		if p.Flags()&ast.SymbolFlagsGetAccessor != 0 {
			get = []*Type{read}
		}
		if p.Flags()&ast.SymbolFlagsSetAccessor != 0 {
			set = []*Type{c.getWriteTypeOfSymbol(p)}
		}
		parts = append(parts, typeFunctionField{"getter", c.createTupleType(get)}, typeFunctionField{"setter", c.createTupleType(set)})
	} else {
		if p.Flags()&ast.SymbolFlagsMethod != 0 {
			kind = "method"
		}
		parts = append(parts, typeFunctionField{"type", c.getNonMissingTypeOfSymbol(p)}, typeFunctionField{"readonly", c.typeFunctionBoolean(readonly)})
	}
	return c.typeFunctionRecord(append(parts, typeFunctionField{"kind", c.getStringLiteralType(kind)})...)
}

func (c *Checker) installTypeFunctionField(fields ast.SymbolTable, key, d *Type) bool {
	kind := "data"
	field := c.getTypeOfPropertyOfType
	if tag := field(d, "kind"); tag != nil {
		var ok bool
		kind, ok = typeFunctionString(tag)
		if !ok {
			return false
		}
	}
	readonly, ok := c.typeFunctionFlag(d, "readonly")
	optional, ok2 := c.typeFunctionFlag(d, "optional")
	if !ok || !ok2 {
		return false
	}
	var read, write *Type
	flags := ast.SymbolFlagsProperty
	switch kind {
	case "data", "method":
		read = field(d, "type")
		if read == nil {
			return false
		}
		if kind == "method" {
			if len(c.getSignaturesOfType(c.GetNonNullableType(read), SignatureKindCall)) == 0 {
				return false
			}
			flags = ast.SymbolFlagsMethod
		}
		if optional && c.strictNullChecks {
			read = c.getOptionalType(read, true)
		}
	case "accessor":
		get, getOK := typeFunctionTuple(c, field(d, "getter"))
		set, setOK := typeFunctionTuple(c, field(d, "setter"))
		if !getOK || !setOK || len(get) > 1 || len(set) > 1 || len(get)+len(set) == 0 || optional {
			return false
		}
		flags = 0
		if len(set) > 0 {
			read, write = set[0], set[0]
			flags |= ast.SymbolFlagsSetAccessor
		}
		if len(get) > 0 {
			read = get[0]
			flags |= ast.SymbolFlagsGetAccessor
		}
		readonly = len(set) == 0
	default:
		return false
	}
	if optional {
		flags |= ast.SymbolFlagsOptional
	}
	name := getPropertyNameFromType(key)
	p := c.newSymbolEx(flags, name, core.IfElse(readonly, ast.CheckFlagsReadonly, 0))
	links := c.valueSymbolLinks.Get(p)
	links.resolvedType = read
	links.setNameType(c, key)
	if write != nil {
		links.setWriteType(c, write)
	}
	if expected := field(d, "read"); expected != nil && !c.isTypeIdenticalTo(expected, read) {
		return false
	}
	if expected := field(d, "write"); expected != nil {
		actual := []*Type{c.getStringLiteralType("readonly")}
		if !readonly {
			actual = []*Type{c.getStringLiteralType("accepts"), c.getWriteTypeOfSymbol(p)}
		}
		if !c.isTypeIdenticalTo(expected, c.createTupleType(actual)) {
			return false
		}
	}
	fields.Set(name, p)
	return true
}

func (c *Checker) typeFunctionObject(node *ast.Node, op string, args []*Type) *Type {
	t := args[0]
	if t.flags&(TypeFlagsObject|TypeFlagsIntersection) == 0 || t.IsTupleType() || c.isArrayType(t) {
		return nil
	}
	properties := c.getPropertiesOfType(t)
	indexes := c.getIndexInfosOfType(t)
	if !c.typeFunctionCharge(node, len(properties)+len(indexes)) {
		return c.anyType
	}
	fields := ast.NewSymbolTableWithCapacity(len(properties))
	for _, p := range properties {
		fields.Set(p.Name(), p)
	}
	switch op {
	case "fields":
		desc := ast.NewSymbolTable()
		for name, p := range fields.All() {
			if typeFunctionPublicField(p) {
				entry := c.newProperty(name, c.describeTypeFunctionField(p))
				c.valueSymbolLinks.Get(entry).setNameType(c, c.getLiteralTypeFromProperty(p, TypeFlagsStringOrNumberLiteralOrUnique, false))
				desc.Set(name, entry)
			}
		}
		return c.newAnonymousType(nil, desc, nil, nil, nil)
	case "indexes":
		var records []*Type
		for _, i := range indexes {
			records = append(records, c.typeFunctionRecord(typeFunctionField{"key", i.keyType}, typeFunctionField{"type", i.valueType}, typeFunctionField{"readonly", c.typeFunctionBoolean(i.isReadonly)}))
		}
		return c.getUnionType(records)
	case "withIndexes":
		indexes = nil
		if args[1] != c.neverType {
			for _, d := range args[1].Distributed() {
				key := c.getTypeOfPropertyOfType(d, "key")
				value := c.getTypeOfPropertyOfType(d, "type")
				ro, ok := c.typeFunctionFlag(d, "readonly")
				if !ok || value == nil || key == nil {
					return nil
				}
				for _, domain := range key.Distributed() {
					if !c.isValidIndexKeyType(domain) {
						return nil
					}
					if existing := c.findIndexInfo(indexes, domain); existing != nil {
						if existing.isReadonly != ro || !c.isTypeIdenticalTo(existing.valueType, value) {
							return nil
						}
						continue
					}
					indexes = append(indexes, &IndexInfo{keyType: domain, valueType: value, isReadonly: ro})
				}
			}
		}
	case "withFields":
		if args[1].flags&TypeFlagsObject == 0 || len(c.getIndexInfosOfType(args[1])) != 0 {
			return nil
		}
		surviving := ast.NewSymbolTable()
		for name, p := range fields.All() {
			if !typeFunctionPublicField(p) {
				surviving.Set(name, p)
			}
		}
		fields = surviving
		replacements := c.getPropertiesOfType(args[1])
		if !c.typeFunctionCharge(node, len(replacements)) {
			return c.anyType
		}
		for _, p := range replacements {
			key := c.getLiteralTypeFromProperty(p, TypeFlagsStringOrNumberLiteralOrUnique, false)
			if key == c.neverType || fields.Get(p.Name()) != nil || !c.installTypeFunctionField(fields, key, c.getTypeOfSymbol(p)) {
				return nil
			}
		}
	default:
		if !isTypeUsableAsPropertyName(args[1]) {
			return nil
		}
		key := getPropertyNameFromType(args[1])
		p := fields.Get(key)
		if p != nil && !typeFunctionPublicField(p) {
			if op == "hasField" {
				return c.regularFalseType
			}
			return nil
		}
		switch op {
		case "hasField":
			return c.typeFunctionBoolean(p != nil)
		case "getField":
			if p == nil {
				return nil
			}
			return c.describeTypeFunctionField(p)
		case "removeField":
			if p == nil {
				return t
			}
			fields.Delete(key)
		case "addField", "setField":
			if op == "addField" && p != nil || !c.installTypeFunctionField(fields, args[1], args[2]) {
				return nil
			}
		default:
			return nil
		}
	}
	for _, index := range indexes {
		if !c.typeFunctionCharge(node, fields.Len()+len(indexes)) {
			return c.anyType
		}
		for _, p := range fields.All() {
			key := c.getLiteralTypeFromProperty(p, TypeFlagsStringOrNumberLiteralOrUnique, false)
			if key != c.neverType && c.isApplicableIndexType(key, index.keyType) && !c.isTypeAssignableTo(c.getNonMissingTypeOfSymbol(p), index.valueType) {
				return nil
			}
		}
		for _, other := range indexes {
			if other != index && c.isApplicableIndexType(other.keyType, index.keyType) && !c.isTypeAssignableTo(other.valueType, index.valueType) {
				return nil
			}
		}
	}
	return c.newAnonymousType(nil, fields, c.getSignaturesOfType(t, SignatureKindCall), c.getSignaturesOfType(t, SignatureKindConstruct), indexes)
}
