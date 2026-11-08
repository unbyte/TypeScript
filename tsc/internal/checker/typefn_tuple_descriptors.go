package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/scanner"
)

type typefnTupleLabelKey struct {
	site  *ast.Node
	name  string
	flags ElementFlags
}

func (v *typefnVM) tupleLabel(node *ast.Node, name string, flags ElementFlags) *ast.Node {
	token := scanner.StringToToken(name)
	if !scanner.IsIdentifierText(name, core.LanguageVariantStandard) || token >= ast.KindFirstReservedWord && token <= ast.KindLastFutureReservedWord || token == ast.KindAwaitKeyword {
		v.error(node, "tuple label must be a native-valid identifier")
		return nil
	}
	state := v.c.getTypefnState()
	key := typefnTupleLabelKey{node, name, flags}
	if label := state.tupleLabels[key]; label != nil {
		return label
	}
	if !v.allocate(node, 4) {
		return nil
	}
	factory := ast.NewNodeFactory(ast.NodeFactoryHooks{})
	var question, rest *ast.Node
	if flags&ElementFlagsOptional != 0 {
		question = factory.NewToken(ast.KindQuestionToken)
	}
	if flags&ElementFlagsVariable != 0 {
		rest = factory.NewToken(ast.KindDotDotDotToken)
	}
	label := factory.NewNamedTupleMember(rest, factory.NewIdentifier(name), question, factory.NewKeywordTypeNode(ast.KindUnknownKeyword))
	label.Loc = node.Loc
	label.Parent = node
	ast.SetParentInChildren(label)
	if state.tupleLabels == nil {
		state.tupleLabels = map[typefnTupleLabelKey]*ast.Node{}
	}
	state.tupleLabels[key] = label
	return label
}

func (v *typefnVM) tupleType(node *ast.Node, descriptor *Type) *Type {
	fields, ok := v.descriptor(node, descriptor, []string{"elements", "readonly"}, nil)
	if !ok {
		return v.c.errorType
	}
	readonly, ok := v.boolean(node, fields[1])
	if !ok {
		return v.c.errorType
	}
	entries, ok := v.requiredTuple(node, fields[0])
	if !ok || !v.allocate(node, len(entries)+1) {
		return v.c.errorType
	}
	values := make([]*Type, 0, len(entries))
	infos := make([]TupleElementInfo, 0, len(entries))
	seenOptional, seenRest := false, false
	for _, entry := range entries {
		if !v.step(node) {
			return v.c.errorType
		}
		fields, ok := v.descriptor(node, entry, []string{"value", "label", "mode"}, nil)
		if !ok {
			return v.c.errorType
		}
		mode, ok := v.descriptorString(node, fields[2])
		if !ok {
			return v.c.errorType
		}
		value := fields[0]
		flags := ElementFlagsRequired
		switch mode {
		case "required":
			if seenOptional {
				return v.error(node, "required tuple element cannot follow an optional element")
			}
		case "optional":
			if seenRest {
				return v.error(node, "optional tuple element cannot follow a rest element")
			}
			seenOptional = true
			flags = ElementFlagsOptional
		case "rest", "variadic":
			hasRest := false
			if graph := constructionOf(value); graph != nil {
				if mode == "rest" {
					if graph.kind != ast.KindArrayType {
						return v.error(node, "rest tuple descriptor requires an array Type")
					}
					value = graph.inputs[0]
					flags, hasRest = ElementFlagsRest, true
				} else {
					if graph.kind != ast.KindTupleType {
						return v.error(node, "variadic tuple descriptor requires a tuple Type")
					}
					flags = ElementFlagsVariadic
					for _, info := range graph.elements {
						hasRest = hasRest || info.flags&ElementFlagsVariable != 0
					}
				}
			} else {
				if typefnUnresolved(value) || v.c.isGenericTupleType(value) || v.c.isGenericObjectType(value) {
					v.blocked = true
					return v.c.errorType
				}
				if mode == "rest" {
					if isTupleType(value) || !v.c.isArrayLikeType(value) {
						return v.error(node, "rest tuple descriptor requires an array Type")
					}
					value = v.c.getIndexTypeOfType(value, v.c.numberType)
					if value == nil {
						return v.error(node, "rest array has no numeric element Type")
					}
					flags, hasRest = ElementFlagsRest, true
				} else {
					if !isTupleType(value) {
						return v.error(node, "variadic tuple descriptor requires a tuple Type")
					}
					flags = ElementFlagsVariadic
					hasRest = value.TargetTupleType().combinedFlags&ElementFlagsRest != 0
				}
			}
			if hasRest {
				if seenRest {
					return v.error(node, "a rest tuple element cannot follow another rest element")
				}
				seenRest = true
			}
		default:
			return v.error(node, "tuple mode must be required, optional, rest or variadic")
		}
		var label *ast.Node
		labelValue, ok := v.optionalDescriptorType(node, fields[1])
		if !ok {
			return v.c.errorType
		}
		if labelValue != nil {
			name, ok := v.descriptorString(node, labelValue)
			if !ok {
				return v.c.errorType
			}
			label = v.tupleLabel(node, name, flags)
			if v.failed {
				return v.c.errorType
			}
		}
		values = append(values, value)
		infos = append(infos, TupleElementInfo{flags: flags, labeledDeclaration: label})
	}
	for _, value := range values {
		if constructionOf(value) != nil {
			return v.construct(node, &typefnConstruction{kind: ast.KindTupleType, inputs: values, elements: infos, readonly: readonly})
		}
	}
	target := v.c.getTupleTargetType(infos, readonly)
	if len(values) == 0 {
		return target
	}
	if target.objectFlags&ObjectFlagsTuple == 0 {
		return v.c.createTypeReference(target, values)
	}
	return v.c.createNormalizedTupleTypeWithDeclared(target, values, values, ObjectFlagsNone)
}

func (v *typefnVM) reflectTuple(node *ast.Node, value *Type) *Type {
	if v.c.isGenericTupleType(value) || v.c.isGenericObjectType(value) {
		v.blocked = true
		return v.c.errorType
	}
	if !isTupleType(value) {
		return v.error(node, "reflect.tuple requires one native tuple Type")
	}
	elements := v.c.getElementTypes(value)
	infos := value.TargetTupleType().elementInfos
	declared := v.c.typefnTupleDeclared(value)
	if !v.allocate(node, 2*len(elements)+2) {
		return v.c.errorType
	}
	records := make([]*Type, 0, len(elements))
	for i, element := range elements {
		if !v.step(node) {
			return v.c.errorType
		}
		info := infos[i]
		mode := "required"
		switch info.flags {
		case ElementFlagsOptional:
			mode = "optional"
			if i >= len(declared) || declared[i] == nil {
				return v.error(node, "optional tuple declared-value provenance is unavailable")
			}
			element = declared[i]
		case ElementFlagsRest:
			mode = "rest"
			element = v.c.createArrayType(element)
		case ElementFlagsVariadic:
			mode = "variadic"
		}
		var label []*Type
		if info.labeledDeclaration != nil {
			label = []*Type{v.c.getStringLiteralType(v.c.getTupleElementLabel(info, nil, i))}
		}
		records = append(records, v.record(node,
			typefnRecordField{"value", element}, typefnRecordField{"label", v.c.createTupleType(label)}, typefnRecordField{"mode", v.c.getStringLiteralType(mode)}))
		if v.failed {
			return v.c.errorType
		}
	}
	return v.record(node, typefnRecordField{"elements", v.c.createTupleType(records)}, typefnRecordField{"readonly", v.boolType(value.TargetTupleType().readonly)})
}
