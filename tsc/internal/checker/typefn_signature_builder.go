package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/scanner"
)

func (v *typefnVM) descriptorString(node *ast.Node, value *Type) (string, bool) {
	if v.rejectConstruction(node, value) {
		return "", false
	}
	value = v.observe(value)
	if typefnUnresolved(value) {
		v.blocked = true
		return "", false
	}
	if value.flags != TypeFlagsStringLiteral {
		v.error(node, "descriptor metadata requires one string literal")
		return "", false
	}
	return value.AsLiteralType().value.(string), true
}

func (v *typefnVM) optionalDescriptorType(node *ast.Node, value *Type) (*Type, bool) {
	entries, ok := v.requiredTuple(node, value)
	if !ok {
		return nil, false
	}
	if len(entries) > 1 {
		v.error(node, "optional descriptor field requires an empty or single-element tuple")
		return nil, false
	}
	if len(entries) == 0 {
		return nil, true
	}
	return v.observe(entries[0]), true
}

func typefnParameterName(name string) bool {
	token := scanner.StringToToken(name)
	return scanner.IsIdentifierText(name, core.LanguageVariantStandard) &&
		!(token >= ast.KindFirstReservedWord && token <= ast.KindLastFutureReservedWord) &&
		token != ast.KindAwaitKeyword && name != "eval" && name != "arguments"
}

// Metadata is validated separately from Type payloads. The same immutable plan
// builds ordinary signatures now and can be checked after recursive sealing.
type typefnSignatureParameter struct {
	name  string
	flags ast.CheckFlags
}
type typefnSignaturePredicate struct {
	kind     TypePredicateKind
	target   int
	narrowed int // input index, -1 for a truthiness assertion
}
type typefnSignaturePlan struct {
	kind          SignatureKind
	flags         SignatureFlags
	parameters    []typefnSignatureParameter
	minimum       int
	this, returns int // input indices; this is -1 when absent
	predicate     *typefnSignaturePredicate
}

func (v *typefnVM) signatureType(node *ast.Node, value *Type) *Type {
	plan, inputs := v.signaturePlan(node, value)
	if plan == nil {
		return v.c.errorType
	}
	v.validateSignaturePlan(node, plan, inputs)
	if v.failed {
		return v.c.errorType
	}
	return v.c.typefnSignatureFromPlan(plan, inputs)
}

func (v *typefnVM) signaturePlan(node *ast.Node, value *Type) (*typefnSignaturePlan, []*Type) {
	fields, ok := v.descriptor(node, value, []string{"kind", "parameters", "this", "returns", "abstract", "predicate"}, nil)
	if !ok {
		return nil, nil
	}
	kindName, ok := v.descriptorString(node, fields[0])
	if !ok {
		return nil, nil
	}
	plan := &typefnSignaturePlan{kind: SignatureKindCall, this: -1}
	switch kindName {
	case "call":
	case "construct":
		plan.kind, plan.flags = SignatureKindConstruct, SignatureFlagsConstruct
	default:
		v.error(node, "signature kind must be call or construct")
		return nil, nil
	}
	abstract, ok := v.boolean(node, fields[4])
	if !ok {
		return nil, nil
	}
	if abstract {
		if plan.kind == SignatureKindCall {
			v.error(node, "call signatures cannot be abstract")
			return nil, nil
		}
		plan.flags |= SignatureFlagsAbstract
	}
	thisType, ok := v.optionalDescriptorType(node, fields[2])
	if !ok {
		return nil, nil
	}
	predicate, ok := v.optionalDescriptorType(node, fields[5])
	if !ok {
		return nil, nil
	}
	if plan.kind == SignatureKindConstruct && (thisType != nil || predicate != nil) {
		v.error(node, "construct signatures require empty this and predicate fields")
		return nil, nil
	}
	entries, ok := v.requiredTuple(node, fields[1])
	if !ok || !v.allocate(node, len(entries)+3) {
		return nil, nil
	}
	inputs := make([]*Type, 0, len(entries)+3)
	names := map[string]bool{}
	optional := false
	for i, entry := range entries {
		if !v.step(node) {
			return nil, nil
		}
		fields, ok := v.descriptor(node, entry, []string{"name", "value", "mode"}, nil)
		if !ok {
			return nil, nil
		}
		name, ok := v.descriptorString(node, fields[0])
		if !ok {
			return nil, nil
		}
		if !typefnParameterName(name) {
			v.error(node, "parameter name must be a native-valid identifier")
			return nil, nil
		}
		if names[name] {
			v.error(node, "duplicate parameter name")
			return nil, nil
		}
		names[name] = true
		mode, ok := v.descriptorString(node, fields[2])
		if !ok {
			return nil, nil
		}
		flags := ast.CheckFlagsNone
		switch mode {
		case "required":
			if optional {
				v.error(node, "required parameter cannot follow an optional parameter")
				return nil, nil
			}
			plan.minimum = i + 1
		case "optional":
			optional, flags = true, ast.CheckFlagsOptionalParameter
		case "rest":
			if i != len(entries)-1 {
				v.error(node, "rest parameter must be last")
				return nil, nil
			}
			plan.flags |= SignatureFlagsHasRestParameter
			flags = ast.CheckFlagsRestParameter
		default:
			v.error(node, "parameter mode must be required, optional or rest")
			return nil, nil
		}
		plan.parameters = append(plan.parameters, typefnSignatureParameter{name, flags})
		inputs = append(inputs, fields[1])
	}
	if thisType != nil {
		plan.this = len(inputs)
		inputs = append(inputs, thisType)
	}
	plan.returns = len(inputs)
	inputs = append(inputs, fields[3])
	if predicate != nil {
		fields, ok := v.descriptor(node, predicate, []string{"kind", "target", "narrowed"}, nil)
		if !ok {
			return nil, nil
		}
		kind, ok := v.descriptorString(node, fields[0])
		if !ok {
			return nil, nil
		}
		pred := &typefnSignaturePredicate{kind: TypePredicateKindIdentifier, narrowed: -1}
		switch kind {
		case "is":
		case "asserts":
			pred.kind = TypePredicateKindAssertsIdentifier
		default:
			v.error(node, "predicate kind must be is or asserts")
			return nil, nil
		}
		narrowed, ok := v.optionalDescriptorType(node, fields[2])
		if !ok {
			return nil, nil
		}
		if kind == "is" && narrowed == nil {
			v.error(node, "is predicate requires one narrowed Type")
			return nil, nil
		}
		if narrowed != nil {
			pred.narrowed = len(inputs)
			inputs = append(inputs, narrowed)
		}
		target := fields[1]
		if v.rejectConstruction(node, target) {
			return nil, nil
		}
		if typefnUnresolved(target) {
			v.blocked = true
			return nil, nil
		}
		if target.flags == TypeFlagsStringLiteral && target.AsLiteralType().value.(string) == "this" {
			if kind == "is" {
				pred.kind = TypePredicateKindThis
			} else {
				pred.kind = TypePredicateKindAssertsThis
			}
		} else {
			index, ok := v.intrinsicOffset(node, target)
			if !ok {
				return nil, nil
			}
			if index >= int64(len(plan.parameters)) {
				v.error(node, "predicate target is outside the parameter list")
				return nil, nil
			}
			if plan.parameters[index].flags&ast.CheckFlagsRestParameter != 0 {
				v.error(node, "predicate cannot target a rest parameter")
				return nil, nil
			}
			pred.target = int(index)
		}
		plan.predicate = pred
	}
	return plan, inputs
}

func (v *typefnVM) validateSignaturePlan(node *ast.Node, plan *typefnSignaturePlan, inputs []*Type) {
	for i, parameter := range plan.parameters {
		if parameter.flags&ast.CheckFlagsRestParameter != 0 {
			v.requireAssignable(node, inputs[i], v.c.anyReadonlyArrayType, "rest parameter requires an array or tuple Type")
			if v.failed {
				return
			}
		}
	}
	if pred := plan.predicate; pred != nil {
		expected := v.c.booleanType
		if pred.kind == TypePredicateKindAssertsIdentifier || pred.kind == TypePredicateKindAssertsThis {
			expected = v.c.voidType
		}
		result := v.observe(inputs[plan.returns])
		if result != expected {
			if v.unresolvedGraph(node, result) {
				v.blocked = true
			} else {
				v.error(node, "predicate return Type must be boolean for is or void for asserts")
				return
			}
		}
		if pred.narrowed != -1 && (pred.kind == TypePredicateKindIdentifier || pred.kind == TypePredicateKindAssertsIdentifier) {
			parameterType := inputs[pred.target]
			if plan.parameters[pred.target].flags&ast.CheckFlagsOptionalParameter != 0 {
				parameterType = v.c.addOptionalityEx(parameterType, false, true)
			}
			v.requireAssignable(node, inputs[pred.narrowed], parameterType, "predicate narrowed Type is not assignable to the parameter Type")
		}
	}
}

func (c *Checker) typefnSignatureFromPlan(plan *typefnSignaturePlan, inputs []*Type) *Type {
	parameters := make([]*ast.Symbol, len(plan.parameters))
	for i, p := range plan.parameters {
		symbol := c.newSymbolEx(ast.SymbolFlagsFunctionScopedVariable, p.name, p.flags)
		value := inputs[i]
		if p.flags&ast.CheckFlagsOptionalParameter != 0 {
			value = c.addOptionalityEx(value, false, true)
		}
		c.valueSymbolLinks.Get(symbol).resolvedType = value
		parameters[i] = symbol
	}
	var this *ast.Symbol
	if plan.this != -1 {
		this = c.newSymbolEx(ast.SymbolFlagsFunctionScopedVariable, ast.InternalSymbolNameThis, ast.CheckFlagsNone)
		c.valueSymbolLinks.Get(this).resolvedType = inputs[plan.this]
	}
	predicate := c.noTypePredicate
	if pred := plan.predicate; pred != nil {
		var narrowed *Type
		if pred.narrowed != -1 {
			narrowed = inputs[pred.narrowed]
		}
		name := ""
		if pred.kind == TypePredicateKindIdentifier || pred.kind == TypePredicateKindAssertsIdentifier {
			name = plan.parameters[pred.target].name
		}
		predicate = c.newTypePredicate(pred.kind, name, int32(pred.target), narrowed)
	}
	sig := c.newSignature(plan.flags, nil, nil, this, parameters, inputs[plan.returns], predicate, plan.minimum)
	return c.typefnSignatureType(sig, plan.kind)
}
