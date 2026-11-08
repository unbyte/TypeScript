package checker

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
)

// Descriptor schemas are small and fixed. A value array avoids allocating maps
// and permits source-independent field order without forcing ignored metadata.
func (v *typefnVM) descriptor(node *ast.Node, value *Type, required, extra []string) ([8]*Type, bool) {
	var fields [8]*Type
	if len(required) > len(fields) {
		panic("typefn descriptor schema exceeds fixed field capacity")
	}
	value = v.observe(value)
	if graph := constructionOf(value); graph != nil {
		if graph.kind != ast.KindTypeLiteral {
			v.error(node, "cannot observe an unsealed recursive reference")
			return fields, false
		}
		for i, property := range graph.properties {
			if !v.step(node) {
				return fields, false
			}
			index := slices.Index(required, property.name)
			if index < 0 {
				if !slices.Contains(extra, property.name) {
					v.error(node, "unrecognized descriptor field "+property.name)
					return fields, false
				}
				continue
			}
			if property.optional {
				v.error(node, "descriptor field must be required and public: "+property.name)
				return fields, false
			}
			fields[index] = graph.inputs[i]
		}
		for index, name := range required {
			if fields[index] == nil {
				v.error(node, "missing descriptor field "+name)
				return fields, false
			}
		}
		return fields, true
	}
	if typefnUnresolved(value) || v.c.isGenericObjectType(value) {
		v.blocked = true
		return fields, false
	}
	if value.flags&(TypeFlagsObject|TypeFlagsIntersection) == 0 {
		v.error(node, "descriptor requires one structural object")
		return fields, false
	}
	for _, property := range v.c.getPropertiesOfType(value) {
		if !v.step(node) {
			return fields, false
		}
		index := slices.Index(required, property.Name())
		if index < 0 {
			if !slices.Contains(extra, property.Name()) {
				v.error(node, "unrecognized descriptor field "+property.Name())
				return fields, false
			}
			continue
		}
		if property.Flags()&ast.SymbolFlagsOptional != 0 || getDeclarationModifierFlagsFromSymbol(property)&ast.ModifierFlagsNonPublicAccessibilityModifier != 0 {
			v.error(node, "descriptor field must be required and public: "+property.Name())
			return fields, false
		}
		fields[index] = v.observe(v.c.getTypeOfSymbol(property))
		if v.c.isErrorType(fields[index]) {
			v.failed = true
			return fields, false
		}
	}
	for index, name := range required {
		if fields[index] == nil {
			v.error(node, "missing descriptor field "+name)
			return fields, false
		}
	}
	return fields, true
}

func (v *typefnVM) requiredTuple(node *ast.Node, value *Type) ([]*Type, bool) {
	value = v.observe(value)
	if graph := constructionOf(value); graph != nil {
		if requiredConstructionTuple(graph) {
			return graph.inputs, true
		}
		v.error(node, "descriptor list requires a required finite tuple")
		return nil, false
	}
	if typefnUnresolved(value) || v.c.isGenericTupleType(value) {
		v.blocked = true
		return nil, false
	}
	if !isTupleType(value) || value.TargetTupleType().combinedFlags & ^ElementFlagsRequired != 0 {
		v.error(node, "descriptor list requires a required finite tuple")
		return nil, false
	}
	return v.c.getTypeArguments(value), true
}
