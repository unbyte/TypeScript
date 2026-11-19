package checker

import (
	"cmp"
	"slices"
	"unicode/utf16"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/jsnum"
	"github.com/microsoft/TypeScript/tsc/internal/scanner"
	"github.com/microsoft/TypeScript/tsc/internal/stringutil"
)

func typeFunctionTuple(c *Checker, t *Type) ([]*Type, bool) {
	if t == nil || !t.IsTupleType() || t.TargetTupleType().combinedFlags & ^ElementFlagsRequired != 0 {
		return nil, false
	}
	return c.getTypeArguments(t), true
}

func (c *Checker) typeFunctionFlag(t *Type, name string) (bool, bool) {
	value := c.getTypeOfPropertyOfType(t, name)
	if value == nil {
		return false, true
	}
	if value.flags&TypeFlagsBooleanLiteral == 0 {
		return false, false
	}
	return value.AsLiteralType().value.(bool), true
}

func (c *Checker) typeFunctionStructure(node *ast.Node, name string, args []*Type) *Type {
	t := args[0]
	field := c.getTypeOfPropertyOfType
	word := c.getStringLiteralType
	switch name {
	case "union.toTuple":
		if t.flags&TypeFlagsNever != 0 {
			return c.createTupleType(nil)
		}
		values := slices.Clone(t.Distributed())
		if len(values) <= 1 && len(args) == 1 {
			return c.createTupleType(values)
		}
		type alternative struct {
			value, key *Type
			units      []uint16
		}
		entries := make([]alternative, len(values))
		var kind TypeFlags
		for i, value := range values {
			key := value
			if len(args) == 2 {
				key = c.applyTypeFunctionConstructor(node, args[1], []*Type{value})
				if c.typeFunctionExecution.failed {
					return c.anyType
				}
			}
			mask := key.flags & (TypeFlagsStringLiteral | TypeFlagsNumberLiteral | TypeFlagsBigIntLiteral | TypeFlagsBooleanLiteral)
			if mask == 0 || i > 0 && mask != kind {
				return nil
			}
			kind = mask
			entries[i] = alternative{value: value, key: key}
			if kind == TypeFlagsStringLiteral {
				s, _ := typeFunctionString(key)
				if !c.typeFunctionCharge(node, len(s)) {
					return c.anyType
				}
				for len(s) > 0 {
					ch, n := stringutil.DecodeJSStringRune(s)
					s = s[n:]
					if ch > 0xffff {
						a, b := utf16.EncodeRune(ch)
						entries[i].units = append(entries[i].units, uint16(a), uint16(b))
					} else {
						entries[i].units = append(entries[i].units, uint16(ch))
					}
				}
			}
		}
		compare := func(a, b alternative) int {
			switch kind {
			case TypeFlagsStringLiteral:
				return slices.Compare(a.units, b.units)
			case TypeFlagsNumberLiteral:
				x, _ := typeFunctionNumber(a.key)
				y, _ := typeFunctionNumber(b.key)
				return cmp.Compare(x, y)
			case TypeFlagsBigIntLiteral:
				x, _ := typeFunctionInteger(a.key)
				y, _ := typeFunctionInteger(b.key)
				return x.Cmp(y)
			default:
				if a.key.AsLiteralType().value == b.key.AsLiteralType().value {
					return 0
				}
				if a.key.AsLiteralType().value == false {
					return -1
				}
				return 1
			}
		}
		slices.SortFunc(entries, compare)
		for i, e := range entries {
			if i != 0 && compare(entries[i-1], e) == 0 {
				return c.failTypeFunction(node, "union ordering keys must be unique")
			}
			values[i] = e.value
		}
		return c.createTupleType(values)
	case "enumType.members":
		if t.flags&TypeFlagsObject == 0 || t.symbol == nil || t.symbol.Flags()&ast.SymbolFlagsEnum == 0 {
			return nil
		}
		members := ast.NewSymbolTable()
		properties := c.getPropertiesOfType(t)
		if !c.typeFunctionCharge(node, len(properties)) {
			return c.anyType
		}
		for _, p := range properties {
			d := ast.GetDeclarationOfKind(p, ast.KindEnumMember)
			if d == nil {
				continue
			}
			memberType := c.getDeclaredTypeOfSymbol(p)
			parts := []*Type{word("computed")}
			switch v := c.getEnumMemberValue(d).Value.(type) {
			case string:
				parts = []*Type{word("constant"), word(v)}
			case jsnum.Number:
				parts = []*Type{word("constant"), c.getNumberLiteralType(v)}
			}
			members.Set(p.Name(), c.newProperty(p.Name(), c.typeFunctionRecord(typeFunctionField{"type", memberType}, typeFunctionField{"value", c.createTupleType(parts)})))
		}
		return c.newAnonymousType(nil, members, nil, nil, nil)
	case "array.decompose":
		if !c.isArrayType(t) {
			return nil
		}
		return c.typeFunctionRecord(typeFunctionField{"element", c.getTypeArguments(t)[0]}, typeFunctionField{"readonly", c.typeFunctionBoolean(t.Target() == c.globalReadonlyArrayType)})
	case "array.compose":
		elem := field(t, "element")
		ro, ok := c.typeFunctionFlag(t, "readonly")
		if elem == nil || !ok {
			return nil
		}
		return c.createArrayTypeEx(elem, ro)
	case "tuple.decompose":
		if !t.IsTupleType() {
			return nil
		}
		var slots []*Type
		for i, element := range c.getTypeArguments(t) {
			info := t.TargetTupleType().elementInfos[i]
			kind := map[ElementFlags]string{ElementFlagsRequired: "required", ElementFlagsOptional: "optional", ElementFlagsRest: "rest", ElementFlagsVariadic: "variadic"}[info.flags]
			if info.flags == ElementFlagsRest {
				element = c.createArrayType(element)
			}
			if info.flags == ElementFlagsOptional {
				element = c.removeMissingType(element, true)
			}
			f := []typeFunctionField{{"kind", word(kind)}, {"type", element}}
			if n := info.labeledDeclaration; n != nil && ast.IsIdentifier(n.Name()) {
				f = append(f, typeFunctionField{"label", word(n.Name().Text())})
			}
			slots = append(slots, c.typeFunctionRecord(f...))
		}
		return c.typeFunctionRecord(typeFunctionField{"slots", c.createTupleType(slots)}, typeFunctionField{"readonly", c.typeFunctionBoolean(t.TargetTupleType().readonly)})
	case "tuple.compose":
		slots, ok := typeFunctionTuple(c, field(t, "slots"))
		if !ok || !c.typeFunctionCharge(node, len(slots)) {
			return nil
		}
		ro, ok := c.typeFunctionFlag(t, "readonly")
		if !ok {
			return nil
		}
		var values []*Type
		var infos []TupleElementInfo
		var seenOptional, seenRest bool
		for _, slot := range slots {
			kind, ok := typeFunctionString(field(slot, "kind"))
			value := field(slot, "type")
			if !ok || value == nil {
				return nil
			}
			flags := map[string]ElementFlags{"required": ElementFlagsRequired, "optional": ElementFlagsOptional, "rest": ElementFlagsRest, "variadic": ElementFlagsVariadic}[kind]
			if flags == 0 || flags == ElementFlagsRequired && seenOptional || flags == ElementFlagsOptional && seenRest || flags == ElementFlagsRest && seenRest {
				return nil
			}
			if flags == ElementFlagsVariadic && !c.isArrayLikeType(value) {
				return nil
			}
			if flags == ElementFlagsRest {
				seenRest = true
				if !c.isArrayType(value) {
					return nil
				}
				value = c.getTypeArguments(value)[0]
			}
			if flags == ElementFlagsOptional {
				seenOptional = true
				value = c.getOptionalType(value, true)
			}
			values = append(values, value)
			info := TupleElementInfo{flags: flags}
			if label := field(slot, "label"); label != nil {
				name, ok := typeFunctionString(label)
				if !ok || !scanner.IsIdentifierText(name, core.LanguageVariantStandard) {
					return nil
				}
				factory := ast.NewNodeFactory(ast.NodeFactoryHooks{})
				info.labeledDeclaration = factory.NewNamedTupleMember(nil, factory.NewIdentifier(name), nil, nil)
			}
			infos = append(infos, info)
		}
		return c.createTupleTypeEx(values, infos, ro)
	case "template.decompose":
		var texts []string
		var holes []*Type
		if t.flags&TypeFlagsTemplateLiteral != 0 {
			a := t.AsTemplateLiteralType()
			texts = a.texts
			holes = a.types
		} else if s, ok := typeFunctionString(t); ok {
			texts = []string{s}
		} else {
			return nil
		}
		return c.typeFunctionRecord(typeFunctionField{"texts", c.createTupleType(core.Map(texts, word))}, typeFunctionField{"types", c.createTupleType(holes)})
	case "template.compose":
		values, ok := typeFunctionTuple(c, field(t, "texts"))
		holes, ok2 := typeFunctionTuple(c, field(t, "types"))
		if !ok || !ok2 || len(values) != len(holes)+1 || !c.typeFunctionCharge(node, len(values)+len(holes)) {
			return nil
		}
		var texts []string
		for _, v := range values {
			s, ok := typeFunctionString(v)
			if !ok {
				return nil
			}
			texts = append(texts, s)
		}
		for _, hole := range holes {
			if !c.isTypeAssignableTo(hole, c.templateConstraintType) {
				return nil
			}
		}
		return c.getTemplateLiteralType(texts, holes)
	case "iter.protocol":
		mode, ok := typeFunctionString(args[1])
		if !ok || mode != "sync" && mode != "async" || t.flags&(TypeFlagsAnyOrUnknown|TypeFlagsNever|TypeFlagsUnion) != 0 {
			return nil
		}
		use := core.IfElse(mode == "sync", IterationUseForOf, IterationUseForAwaitOf)
		info := c.getIterationTypesOfIterable(t, use, nil)
		if info.yieldType == nil {
			return nil
		}
		return c.createTupleType([]*Type{info.yieldType, info.returnType, info.nextType})
	}
	if len(name) > 7 && name[:7] == "object." {
		return c.typeFunctionObject(node, name[7:], args)
	}
	return nil
}
