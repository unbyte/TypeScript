package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
)

// Optional read Types can erase whether undefined was explicitly declared when
// exactOptionalPropertyTypes is disabled. Keep the declared value for VM-created
// optional members, rather than trying to subtract undefined during reflection.
func (c *Checker) rememberTypefnProperty(symbol *ast.Symbol, value *Type) {
	state := c.getTypefnState()
	if state.propertyValues == nil {
		state.propertyValues = map[*ast.Symbol]*Type{}
	}
	state.propertyValues[symbol] = value
}

func (v *typefnVM) propertyValue(node *ast.Node, symbol *ast.Symbol) *Type {
	if !v.step(node) {
		return v.c.errorType
	}
	c := v.c
	if symbol.Flags()&ast.SymbolFlagsOptional == 0 {
		return c.getTypeOfSymbol(symbol)
	}
	if c.typefn != nil {
		if value := c.typefn.propertyValues[symbol]; value != nil {
			return value
		}
	}
	links := c.valueSymbolLinks.Get(symbol)
	if symbol.CheckFlags()&ast.CheckFlagsInstantiated != 0 {
		return c.instantiateType(v.propertyValue(node, links.target), links.mapper)
	}
	if symbol.CheckFlags()&ast.CheckFlagsMapped != 0 {
		mapped := links.getContainingType()
		base := core.OrElse(mapped.AsMappedType().target, mapped)
		declaration := base.AsMappedType().declaration
		key := c.mappedSymbolLinks.Get(symbol).keyType
		mapper := appendTypeMapping(mapped.AsMappedType().mapper, c.getTypeParameterFromMappedType(mapped), key)
		template := ast.SkipParentheses(declaration.Type)
		if template != nil && ast.IsIndexedAccessTypeNode(template) {
			data := template.AsIndexedAccessTypeNode()
			source := c.instantiateType(c.getTypeFromTypeNode(data.ObjectType), mapper)
			index := c.instantiateType(c.getTypeFromTypeNode(data.IndexType), mapper)
			// Identity projections (Pick/Partial/Readonly and key remapping) retain
			// the source member's declared value, not its optional-read undefined.
			if origin := c.mappedSymbolLinks.Get(symbol).syntheticOrigin; origin != nil && c.isTypeIdenticalTo(index, key) && c.isTypeIdenticalTo(source, c.getModifiersTypeFromMappedType(mapped)) {
				return v.propertyValue(node, origin)
			}
		}
		if template == nil {
			return v.error(node, "mapped property has no declared value Type")
		}
		return c.removeMissingType(c.instantiateType(c.getTypeFromTypeNode(template), mapper), true)
	}
	if container := links.getContainingType(); container != nil && container.flags&TypeFlagsUnionOrIntersection != 0 {
		var values []*Type
		for _, part := range container.Types() {
			if property := c.getPropertyOfTypeEx(part, symbol.Name(), true, false); property != nil && property != symbol {
				values = append(values, v.propertyValue(node, property))
				if v.failed {
					return c.errorType
				}
			}
		}
		if len(values) != 0 {
			return c.getUnionOrIntersectionType(values, container.flags&TypeFlagsUnion != 0, UnionReductionLiteral)
		}
	}
	if spread := c.spreadLinks.TryGet(symbol); spread != nil {
		left := v.propertyValue(node, spread.leftSpread)
		right := v.propertyValue(node, spread.rightSpread)
		if v.failed {
			return c.errorType
		}
		// Mirror the host's optional-right spread rule, including its treatment of
		// undefined under the current exact-optional option.
		leftRead := c.removeMissingOrUndefinedType(c.getTypeOfSymbol(spread.leftSpread))
		rightRead := c.removeMissingOrUndefinedType(c.getTypeOfSymbol(spread.rightSpread))
		if leftRead == rightRead {
			return left
		}
		return c.getUnionTypeEx([]*Type{left, c.removeMissingOrUndefinedType(right)}, UnionReductionSubtype, nil, nil)
	}
	if origin := c.mappedSymbolLinks.TryGet(symbol); origin != nil && origin.syntheticOrigin != nil {
		source := origin.syntheticOrigin
		if source.Flags()&ast.SymbolFlagsSetAccessor != 0 && source.Flags()&ast.SymbolFlagsGetAccessor == 0 {
			return c.undefinedType
		}
		return v.propertyValue(node, source)
	}
	for _, declaration := range symbol.Declarations() {
		if ast.IsPropertyDeclaration(declaration) || ast.IsPropertySignatureDeclaration(declaration) || ast.IsParameterDeclaration(declaration) {
			return c.widenTypeForVariableLikeDeclaration(c.getTypeForVariableLikeDeclaration(declaration, false, CheckModeNormal), declaration, false)
		}
	}
	if symbol.Flags()&ast.SymbolFlagsMethod != 0 {
		// A method declaration cannot explicitly declare undefined as its *function*
		// Type; undefined in its return Type is nested and remains untouched.
		return c.removeMissingOrUndefinedType(c.getTypeOfSymbol(symbol))
	}
	return v.error(node, "optional property declared-value provenance is unavailable")
}

func (v *typefnVM) reflectProperties(node *ast.Node, value *Type) *Type {
	if v.c.isGenericObjectType(value) {
		v.blocked = true
		return v.c.errorType
	}
	if value.flags&(TypeFlagsObject|TypeFlagsIntersection) == 0 {
		return v.error(node, "reflect.properties requires one object Type")
	}
	properties := v.c.getPropertiesOfType(value)
	if !v.allocate(node, len(properties)+1) {
		return v.c.errorType
	}
	records := make([]*Type, 0, len(properties))
	for _, property := range properties {
		if !v.step(node) {
			return v.c.errorType
		}
		if getDeclarationModifierFlagsFromSymbol(property)&ast.ModifierFlagsNonPublicAccessibilityModifier != 0 {
			continue
		}
		if declaration := property.ValueDeclaration(); declaration != nil && declaration.Name() != nil && ast.IsPrivateIdentifier(declaration.Name()) {
			continue
		}
		optional := property.Flags()&ast.SymbolFlagsOptional != 0
		read := v.observe(v.c.addOptionalityEx(v.c.getNonMissingTypeOfSymbol(property), false, optional))
		declared := v.observe(v.propertyValue(node, property))
		if v.failed {
			return v.c.errorType
		}
		if property.Flags()&ast.SymbolFlagsAccessor != 0 {
			write := v.observe(v.c.getWriteTypeOfSymbol(property))
			if !v.c.isTypeIdenticalTo(read, write) {
				if v.unresolvedGraph(node, read) || v.unresolvedGraph(node, write) {
					v.blocked = true
					return v.c.errorType
				}
				return v.error(node, "differing accessor read/write Types require forwarding the original Type")
			}
		}
		key := v.c.getLiteralTypeFromProperty(property, TypeFlagsStringOrNumberLiteralOrUnique, true)
		if key.flags != TypeFlagsUniqueESSymbol {
			key = v.c.getStringLiteralType(property.Name())
		}
		record := v.record(node,
			typefnRecordField{"key", key}, typefnRecordField{"value", declared}, typefnRecordField{"read", read},
			typefnRecordField{"optional", v.boolType(optional)}, typefnRecordField{"readonly", v.boolType(v.c.isReadonlySymbol(property))})
		if v.failed {
			return v.c.errorType
		}
		records = append(records, record)
	}
	return v.c.createTupleType(records)
}
