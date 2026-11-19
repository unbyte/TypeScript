package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"slices"
)

func (c *Checker) typeFunctionCallable(op string, args []*Type) *Type {
	t := args[0]
	field := c.getTypeOfPropertyOfType
	switch op {
	case "ref":
		if !isTypeUsableAsPropertyName(t) {
			return nil
		}
		return c.typeFunctionBinderRef(t)
	case "rebind":
		pairs, ok := typeFunctionTuple(c, args[1])
		if !ok {
			return nil
		}
		var from, to []*Type
		for _, pair := range pairs {
			a, b := field(pair, "from"), field(pair, "to")
			if a == nil || b == nil || a.flags&TypeFlagsTypeParameter == 0 || slices.Contains(from, a) {
				return nil
			}
			from = append(from, a)
			to = append(to, b)
		}
		return c.instantiateType(t, newTypeMapper(from, to))
	case "specialize":
		input, ok := typeFunctionTuple(c, args[1])
		if !ok {
			return nil
		}
		calls := c.getSignaturesOfType(t, SignatureKindCall)
		ctors := c.getSignaturesOfType(t, SignatureKindConstruct)
		if len(calls)+len(ctors) != 1 {
			return nil
		}
		sig := core.FirstOrNil(calls)
		if sig == nil {
			sig = ctors[0]
		}
		params := sig.TypeParameters()
		if len(params) == 0 || len(input) < c.getMinTypeArgumentCount(params) || len(input) > len(params) {
			return nil
		}
		values := c.fillMissingTypeArguments(input, params, c.getMinTypeArgumentCount(params), false)
		mapper := newTypeMapper(params, values)
		for i, p := range params {
			if constraint := c.getConstraintOfTypeParameter(p); constraint != nil && !c.isTypeAssignableTo(values[i], c.instantiateType(constraint, mapper)) {
				return nil
			}
		}
		sig = c.getSignatureInstantiation(sig, values, false, nil)
		if len(calls) == 1 {
			calls = []*Signature{sig}
		} else {
			ctors = []*Signature{sig}
		}
		return c.newAnonymousType(nil, nil, calls, ctors, nil)
	case "decompose":
		calls, ctors := c.getSignaturesOfType(t, SignatureKindCall), c.getSignaturesOfType(t, SignatureKindConstruct)
		if t.flags&TypeFlagsUnion != 0 || len(calls)+len(ctors) == 0 {
			return nil
		}
		members := ast.NewSymbolTable()
		for _, p := range c.getPropertiesOfType(t) {
			members.Set(p.Name(), p)
		}
		return c.typeFunctionRecord(typeFunctionField{"members", c.newAnonymousType(nil, members, nil, nil, c.getIndexInfosOfType(t))}, typeFunctionField{"calls", c.createTupleType(core.Map(calls, c.describeTypeFunctionSignature))}, typeFunctionField{"constructs", c.createTupleType(core.Map(ctors, c.describeTypeFunctionSignature))})
	case "compose":
		parts := field(t, "members")
		if parts == nil || parts.flags&TypeFlagsObject == 0 {
			return nil
		}
		members := ast.NewSymbolTable()
		for _, p := range c.getPropertiesOfType(parts) {
			members.Set(p.Name(), p)
		}
		var groups [2][]*Signature
		for i, list := range []string{"calls", "constructs"} {
			if len(c.getSignaturesOfType(parts, SignatureKind(i))) != 0 {
				return nil
			}
			records, ok := typeFunctionTuple(c, field(t, list))
			if !ok {
				return nil
			}
			for _, r := range records {
				sig := c.composeTypeFunctionSignature(r, i == 1)
				if sig == nil {
					return nil
				}
				groups[i] = append(groups[i], sig)
			}
		}
		return c.newAnonymousType(nil, members, groups[0], groups[1], c.getIndexInfosOfType(parts))
	}
	return nil
}

func (c *Checker) typeFunctionBinderRef(key *Type) *Type {
	root := c.typeFunctionExecution
	if root.refs == nil {
		root.refs = make(map[*Type]*Type)
	}
	if ref := root.refs[key]; ref != nil {
		return ref
	}
	ref := c.newTypeParameter(nil)
	ref.AsTypeParameter().constraint = c.noConstraintType
	root.refs[key] = ref
	return ref
}

func (c *Checker) describeTypeFunctionSignature(s *Signature) *Type {
	word := c.getStringLiteralType
	record := c.typeFunctionRecord
	var parameters, binders []*Type
	for _, p := range s.TypeParameters() {
		key := word(p.symbol.Name())
		f := []typeFunctionField{{"name", key}, {"key", key}, {"ref", p}}
		if constraint := c.getConstraintOfTypeParameter(p); constraint != nil {
			f = append(f, typeFunctionField{"constraint", constraint})
		}
		if def := c.getDefaultFromTypeParameter(p); def != nil {
			f = append(f, typeFunctionField{"default", def})
		}
		if c.isConstTypeVariable(p, 0) {
			f = append(f, typeFunctionField{"const", c.regularTrueType})
		}
		binders = append(binders, record(f...))
	}
	for i, p := range s.parameters {
		kind := "required"
		if s.HasRestParameter() && i == len(s.parameters)-1 {
			kind = "rest"
		} else if i >= int(s.minArgumentCount) {
			kind = "optional"
		}
		parameters = append(parameters, record(typeFunctionField{"name", word(p.Name())}, typeFunctionField{"key", c.typeFunctionInt(i)}, typeFunctionField{"kind", word(kind)}, typeFunctionField{"type", c.getTypeOfSymbol(p)}))
	}
	var result *Type
	if pred := c.getTypePredicateOfSignature(s); pred != nil {
		kind := "predicate"
		if pred.kind == TypePredicateKindAssertsIdentifier || pred.kind == TypePredicateKindAssertsThis {
			kind = "asserts"
		}
		target := []*Type{word("this")}
		if pred.kind == TypePredicateKindIdentifier || pred.kind == TypePredicateKindAssertsIdentifier {
			target = []*Type{word("parameter"), c.typeFunctionInt(int(pred.parameterIndex))}
		}
		f := []typeFunctionField{{"kind", word(kind)}, {"target", c.createTupleType(target)}}
		if pred.t != nil {
			f = append(f, typeFunctionField{"type", pred.t})
		}
		result = record(f...)
	} else {
		result = record(typeFunctionField{"kind", word("return")}, typeFunctionField{"type", c.getReturnTypeOfSignature(s)})
	}
	origin := c.newAnonymousType(nil, nil, nil, nil, nil)
	if c.typeFunctionOrigins == nil {
		c.typeFunctionOrigins = make(map[*Type]*Signature)
	}
	c.typeFunctionOrigins[origin] = s
	fields := []typeFunctionField{{"typeParameters", c.createTupleType(binders)}, {"parameters", c.createTupleType(parameters)}, {"result", result}, {"origin", origin}}
	if p := s.ThisParameter(); p != nil {
		fields = append(fields, typeFunctionField{"thisParameter", record(typeFunctionField{"type", c.getTypeOfSymbol(p)})})
	}
	if s.flags&SignatureFlagsAbstract != 0 {
		fields = append(fields, typeFunctionField{"abstract", c.regularTrueType})
	}
	return record(fields...)
}

func (c *Checker) composeTypeFunctionSignature(r *Type, construct bool) *Signature {
	field := c.getTypeOfPropertyOfType
	var old, params []*Type
	var declarations []*Type
	if list := field(r, "typeParameters"); list != nil {
		var ok bool
		declarations, ok = typeFunctionTuple(c, list)
		if !ok {
			return nil
		}
	}
	keys := make(map[*Type]bool)
	for _, d := range declarations {
		key := field(d, "key")
		name, ok := typeFunctionString(field(d, "name"))
		if !ok || key == nil || !isTypeUsableAsPropertyName(key) || keys[key] {
			return nil
		}
		keys[key] = true
		ref := field(d, "ref")
		if ref == nil {
			ref = c.typeFunctionBinderRef(key)
		}
		if ref.flags&TypeFlagsTypeParameter == 0 {
			return nil
		}
		p := c.newTypeParameter(c.newSymbol(ast.SymbolFlagsTypeParameter, name))
		p.AsTypeParameter().constraint = c.noConstraintType
		p.AsTypeParameter().resolvedDefaultType = c.noConstraintType
		old = append(old, ref)
		params = append(params, p)
	}
	mapper := newTypeMapper(old, params)
	mapped := func(t *Type) *Type { return c.instantiateType(t, mapper) }
	for i, d := range declarations {
		if con := field(d, "constraint"); con != nil {
			params[i].AsTypeParameter().constraint = mapped(con)
		}
		if def := field(d, "default"); def != nil {
			params[i].AsTypeParameter().resolvedDefaultType = mapped(def)
		}
	}
	var recv *ast.Symbol
	if d := field(r, "thisParameter"); d != nil {
		v := field(d, "type")
		if v == nil {
			return nil
		}
		recv = c.newParameter("this", mapped(v))
	}
	values, ok := typeFunctionTuple(c, field(r, "parameters"))
	if !ok {
		return nil
	}
	var parameters []*ast.Symbol
	indices := make(map[*Type]int)
	flags := SignatureFlagsNone
	if construct {
		flags |= SignatureFlagsConstruct
	}
	abstract, ok := c.typeFunctionFlag(r, "abstract")
	if !ok || abstract && !construct {
		return nil
	}
	if abstract {
		flags |= SignatureFlagsAbstract
	}
	minCount := 0
	for i, d := range values {
		v := field(d, "type")
		if v == nil {
			return nil
		}
		name := fmt.Sprintf("arg%d", i)
		named := field(d, "name")
		if named != nil {
			var ok bool
			name, ok = typeFunctionString(named)
			if !ok {
				return nil
			}
		}
		key := field(d, "key")
		if key == nil {
			key = named
		}
		if key != nil {
			if _, exists := indices[key]; exists {
				return nil
			}
			indices[key] = i
		}
		kind := "required"
		if k := field(d, "kind"); k != nil {
			kind, ok = typeFunctionString(k)
			if !ok {
				return nil
			}
		}
		v = mapped(v)
		switch kind {
		case "required":
			if minCount != i {
				return nil
			}
			minCount++
		case "optional":
			v = c.getOptionalType(v, false)
		case "rest":
			if i != len(values)-1 || !c.isArrayLikeType(v) {
				return nil
			}
			flags |= SignatureFlagsHasRestParameter
		default:
			return nil
		}
		parameters = append(parameters, c.newParameter(name, v))
	}
	result := field(r, "result")
	if result == nil {
		return nil
	}
	kind, ok := typeFunctionString(field(result, "kind"))
	if !ok {
		return nil
	}
	ret := mapped(field(result, "type"))
	predicate := c.noTypePredicate
	switch kind {
	case "return":
		if ret == nil {
			return nil
		}
	case "predicate", "asserts":
		if construct || kind == "predicate" && ret == nil {
			return nil
		}
		target, ok := typeFunctionTuple(c, field(result, "target"))
		if !ok || len(target) == 0 {
			return nil
		}
		tag, _ := typeFunctionString(target[0])
		pk := TypePredicateKindThis
		index := -1
		name := ""
		switch tag {
		case "this":
			if len(target) != 1 || recv == nil {
				return nil
			}
		case "parameter":
			if len(target) != 2 {
				return nil
			}
			index, ok = indices[target[1]]
			if !ok {
				return nil
			}
			pk = TypePredicateKindIdentifier
			name = parameters[index].Name()
			if ret != nil && !c.isTypeAssignableTo(ret, c.getTypeOfSymbol(parameters[index])) {
				return nil
			}
		default:
			return nil
		}
		if kind == "asserts" {
			pk += 2
		}
		predicate = c.newTypePredicate(pk, name, int32(index), ret)
		ret = core.IfElse(kind == "asserts", c.voidType, c.booleanType)
	default:
		return nil
	}
	var declaration *ast.Node
	if origin := field(r, "origin"); origin != nil {
		s := c.typeFunctionOrigins[origin]
		if s == nil {
			return nil
		}
		declaration = s.declaration
	}
	sig := c.newSignature(flags, declaration, params, recv, parameters, ret, predicate, minCount)
	sig.typeFunctionComposed = true
	return sig
}
