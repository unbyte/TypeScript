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

func (v *typefnVM) signatureType(node *ast.Node, value *Type) *Type {
	fields, ok := v.descriptor(node, value, []string{"kind", "parameters", "this", "returns", "abstract", "predicate"}, nil)
	if !ok {
		return v.c.errorType
	}
	kindName, ok := v.descriptorString(node, fields[0])
	if !ok {
		return v.c.errorType
	}
	kind, flags := SignatureKindCall, SignatureFlagsNone
	switch kindName {
	case "call":
	case "construct":
		kind, flags = SignatureKindConstruct, SignatureFlagsConstruct
	default:
		return v.error(node, "signature kind must be call or construct")
	}
	abstract, ok := v.boolean(node, fields[4])
	if !ok {
		return v.c.errorType
	}
	if abstract {
		if kind == SignatureKindCall {
			return v.error(node, "call signatures cannot be abstract")
		}
		flags |= SignatureFlagsAbstract
	}
	thisType, ok := v.optionalDescriptorType(node, fields[2])
	if !ok {
		return v.c.errorType
	}
	predicateDescriptor, ok := v.optionalDescriptorType(node, fields[5])
	if !ok {
		return v.c.errorType
	}
	if kind == SignatureKindConstruct && (thisType != nil || predicateDescriptor != nil) {
		return v.error(node, "construct signatures require empty this and predicate fields")
	}
	entries, ok := v.requiredTuple(node, fields[1])
	if !ok {
		return v.c.errorType
	}
	if !v.allocate(node, len(entries)+3) {
		return v.c.errorType
	}
	parameters := make([]*ast.Symbol, 0, len(entries))
	names := map[string]bool{}
	minimum, optional := 0, false
	for i, entry := range entries {
		if !v.step(node) {
			return v.c.errorType
		}
		parameter, ok := v.descriptor(node, entry, []string{"name", "value", "mode"}, nil)
		if !ok {
			return v.c.errorType
		}
		name, ok := v.descriptorString(node, parameter[0])
		if !ok {
			return v.c.errorType
		}
		if !typefnParameterName(name) {
			return v.error(node, "parameter name must be a native-valid identifier")
		}
		if names[name] {
			return v.error(node, "duplicate parameter name")
		}
		names[name] = true
		mode, ok := v.descriptorString(node, parameter[2])
		if !ok {
			return v.c.errorType
		}
		checkFlags := ast.CheckFlagsNone
		parameterType := parameter[1]
		switch mode {
		case "required":
			if optional {
				return v.error(node, "required parameter cannot follow an optional parameter")
			}
			minimum = i + 1
		case "optional":
			optional = true
			checkFlags = ast.CheckFlagsOptionalParameter
			parameterType = v.c.addOptionalityEx(parameterType, false, true)
		case "rest":
			if i != len(entries)-1 {
				return v.error(node, "rest parameter must be last")
			}
			flags |= SignatureFlagsHasRestParameter
			checkFlags = ast.CheckFlagsRestParameter
			v.requireAssignable(node, parameterType, v.c.anyReadonlyArrayType, "rest parameter requires an array or tuple Type")
			if v.failed {
				return v.c.errorType
			}
		default:
			return v.error(node, "parameter mode must be required, optional or rest")
		}
		symbol := v.c.newSymbolEx(ast.SymbolFlagsFunctionScopedVariable, name, checkFlags)
		v.c.valueSymbolLinks.Get(symbol).resolvedType = parameterType
		parameters = append(parameters, symbol)
	}
	var thisParameter *ast.Symbol
	if thisType != nil {
		thisParameter = v.c.newSymbolEx(ast.SymbolFlagsFunctionScopedVariable, ast.InternalSymbolNameThis, ast.CheckFlagsNone)
		v.c.valueSymbolLinks.Get(thisParameter).resolvedType = thisType
	}
	predicate := v.c.noTypePredicate
	if predicateDescriptor != nil {
		predicate = v.signaturePredicate(node, predicateDescriptor, parameters, flags, fields[3])
		if v.failed || predicate == nil {
			return v.c.errorType
		}
	}
	signature := v.c.newSignature(flags, nil, nil, thisParameter, parameters, fields[3], predicate, minimum)
	return v.c.typefnSignatureType(signature, kind)
}

func (v *typefnVM) signaturePredicate(node *ast.Node, value *Type, parameters []*ast.Symbol, flags SignatureFlags, result *Type) *TypePredicate {
	fields, ok := v.descriptor(node, value, []string{"kind", "target", "narrowed"}, nil)
	if !ok {
		return nil
	}
	kind, ok := v.descriptorString(node, fields[0])
	if !ok {
		return nil
	}
	predicateKind := TypePredicateKindIdentifier
	expectedResult := v.c.booleanType
	switch kind {
	case "is":
	case "asserts":
		predicateKind, expectedResult = TypePredicateKindAssertsIdentifier, v.c.voidType
	default:
		v.error(node, "predicate kind must be is or asserts")
		return nil
	}
	narrowed, ok := v.optionalDescriptorType(node, fields[2])
	if !ok {
		return nil
	}
	if kind == "is" && narrowed == nil {
		v.error(node, "is predicate requires one narrowed Type")
		return nil
	}
	result = v.observe(result)
	if result != expectedResult {
		if v.unresolvedGraph(node, result) {
			v.blocked = true
		} else {
			v.error(node, "predicate return Type must be boolean for is or void for asserts")
			return nil
		}
	}
	target := fields[1]
	if typefnUnresolved(target) {
		v.blocked = true
		return nil
	}
	if target.flags == TypeFlagsStringLiteral && target.AsLiteralType().value.(string) == "this" {
		if kind == "is" {
			predicateKind = TypePredicateKindThis
		} else {
			predicateKind = TypePredicateKindAssertsThis
		}
		// The native host permits this predicates without an explicit this parameter
		// and does not require narrowed to be assignable to an explicit receiver Type.
		return v.c.newTypePredicate(predicateKind, "", 0, narrowed)
	}
	index, ok := v.intrinsicOffset(node, target)
	if !ok {
		return nil
	}
	if index >= int64(len(parameters)) {
		v.error(node, "predicate target is outside the parameter list")
		return nil
	}
	if flags&SignatureFlagsHasRestParameter != 0 && int(index) == len(parameters)-1 {
		v.error(node, "predicate cannot target a rest parameter")
		return nil
	}
	parameter := parameters[index]
	if narrowed != nil {
		v.requireAssignable(node, narrowed, v.c.getTypeOfSymbol(parameter), "predicate narrowed Type is not assignable to the parameter Type")
		if v.failed {
			return nil
		}
	}
	return v.c.newTypePredicate(predicateKind, parameter.Name(), int32(index), narrowed)
}
