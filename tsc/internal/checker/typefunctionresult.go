package checker

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
)

// Normalize only the result that crosses the outer call. An intermediate
// generic signature can still be specialized or discarded by the caller.
type typeFunctionResult struct {
	c            *Checker
	site         *ast.Node
	path         map[*Type]*Type
	stack, bound []*Type
	// During inspection the edges point from a component to its containers.
	// This lets an unchanged recursive class retain its native identity, even
	// when another part of the result needs rebuilding.
	parents map[*Type][]*Type
	replace map[*Type]bool
	values  map[*Type]*Type
}

func (c *Checker) finishTypeFunctionResult(t *Type, site *ast.Node) *Type {
	r := typeFunctionResult{c: c, site: site, path: make(map[*Type]*Type), parents: make(map[*Type][]*Type)}
	r.visit(t)
	if c.typeFunctionExecution.failed {
		return c.anyType
	}
	if len(r.values) == 0 {
		return t
	}
	var pending []*Type
	for app := range r.values {
		pending = append(pending, app)
	}
	r.replace = make(map[*Type]bool)
	for len(pending) > 0 {
		i := len(pending) - 1
		part := pending[i]
		pending = pending[:i]
		if !r.replace[part] {
			r.replace[part] = true
			pending = append(pending, r.parents[part]...)
		}
	}
	r.parents = nil
	result := r.visit(t)
	if c.typeFunctionExecution.failed {
		return c.anyType
	}
	return result
}

func (r *typeFunctionResult) visit(t *Type) *Type {
	c := r.c
	if t == nil || t.flags&TypeFlagsStructuredOrInstantiable == 0 || c.typeFunctionExecution.failed {
		return t
	}
	if r.parents != nil {
		if n := len(r.stack); n > 0 {
			r.parents[t] = append(r.parents[t], r.stack[n-1])
		}
		if _, active := r.path[t]; active {
			return t
		}
	} else if !r.replace[t] {
		return t
	}
	if !c.typeFunctionStep(r.site) {
		return c.anyType
	}
	if shell, active := r.path[t]; active {
		if shell != nil {
			return shell
		}
		if t.flags&TypeFlagsObject != 0 {
			if t.objectFlags&ObjectFlagsReference != 0 && (t.IsTupleType() || r.nativeDeclaration(t.symbol)) {
				shell = c.cloneTypeReference(t)
			} else {
				shell = c.newObjectType(ObjectFlagsAnonymous, nil)
			}
			r.path[t] = shell
			return shell
		}
		// A recursive union can lead back to an object under construction.
		if t.flags&TypeFlagsUnionOrIntersection != 0 {
			for _, part := range t.Types() {
				if _, active := r.path[part]; active && part.flags&TypeFlagsObject != 0 {
					return r.unionOrIntersection(t)
				}
			}
		}
		return c.failTypeFunction(r.site, "cannot complete a recursive typefn result")
	}
	if c.isDeeplyNestedType(t, r.stack, 100) {
		return c.failTypeFunction(r.site, "return type exceeds the calculation depth limit")
	}
	r.path[t] = nil
	r.stack = append(r.stack, t)
	defer func() {
		delete(r.path, t)
		r.stack = r.stack[:len(r.stack)-1]
	}()
	if app := t.typeFunctionApplication(); app != nil {
		if r.parents == nil {
			return r.visit(r.values[t])
		}
		if _, found := r.values[t]; !found {
			for _, arg := range app.arguments {
				if !c.typeFunctionClosed(arg, map[*Type]bool{}, map[*Type]bool{}) {
					return c.failTypeFunction(app.use, "the returned type contains a typefn calculation that still needs a concrete argument")
				}
			}
			v := c.typeFunctionExecution.selectedValues[t]
			if v == nil {
				v = c.forceTypeFunction(t)
			}
			if r.values == nil {
				r.values = make(map[*Type]*Type)
			}
			r.values[t] = v
			r.visit(v)
		}
		return t
	}
	if t.flags&TypeFlagsTypeParameter != 0 {
		if c.typeFunctionConstructors[t] != nil {
			return c.failTypeFunction(r.site, "a constructor cannot be returned as a type")
		}
		if !slices.Contains(r.bound, getNonDistributedTypeParameter(t)) {
			return c.failTypeFunction(r.site, "an unbound signature reference escaped the calculation")
		}
		return t
	}
	var args []*Type
	if t.alias != nil {
		args = core.SameMap(t.alias.typeArguments, r.visit)
	}
	result := t
	if t.alias == nil || !r.nativeDeclaration(t.alias.symbol) {
		result = r.structure(t)
	}
	if result == t && t.alias != nil && !core.Same(args, t.alias.typeArguments) {
		return c.getTypeAliasInstantiation(t.alias.symbol, args, nil)
	}
	return result
}

func (r *typeFunctionResult) structure(t *Type) *Type {
	c := r.c
	switch {
	case t.flags&TypeFlagsUnionOrIntersection != 0:
		return r.unionOrIntersection(t)
	case t.flags&TypeFlagsIndex != 0:
		if v := r.visit(t.Target()); v != t.Target() {
			return c.getIndexType(v)
		}
	case t.flags&TypeFlagsIndexedAccess != 0:
		a := t.AsIndexedAccessType()
		o, i := r.visit(a.objectType), r.visit(a.indexType)
		if o != a.objectType || i != a.indexType {
			return c.getIndexedAccessTypeEx(o, i, a.accessFlags, nil, nil)
		}
	case t.flags&TypeFlagsTemplateLiteral != 0:
		d := t.AsTemplateLiteralType()
		if args := core.SameMap(d.types, r.visit); !core.Same(args, d.types) {
			return c.getTemplateLiteralType(d.texts, args)
		}
	case t.flags&TypeFlagsStringMapping != 0:
		if v := r.visit(t.Target()); v != t.Target() {
			return c.getStringMappingType(t.symbol, v)
		}
	case t.flags&TypeFlagsSubstitution != 0:
		a := t.AsSubstitutionType()
		base, constraint := r.visit(a.baseType), r.visit(a.constraint)
		if base != a.baseType || constraint != a.constraint {
			return c.getSubstitutionType(base, constraint)
		}
	case t.flags&TypeFlagsConditional != 0:
		return r.conditional(t)
	case t.flags&TypeFlagsObject != 0:
		return r.object(t)
	}
	return t
}

func (r *typeFunctionResult) unionOrIntersection(t *Type) *Type {
	parts := core.SameMap(t.Types(), r.visit)
	if core.Same(parts, t.Types()) {
		return t
	}
	if t.flags&TypeFlagsUnion != 0 {
		return r.c.getUnionType(parts)
	}
	return r.c.getIntersectionType(parts)
}

func (r *typeFunctionResult) bind(params []*Type) func() {
	prev := r.bound
	r.bound = append(r.bound, params...)
	return func() { r.bound = prev }
}

func (r *typeFunctionResult) object(t *Type) *Type {
	c := r.c
	reference := t
	if t.objectFlags&ObjectFlagsReference != 0 {
		args := c.getTypeArguments(t)
		values := core.SameMap(args, r.visit)
		if !core.Same(args, values) {
			reference = c.createNormalizedTypeReference(t.Target(), values)
		}
		if t.IsTupleType() || r.nativeDeclaration(t.symbol) {
			if shell := r.path[t]; shell != nil {
				if reference.objectFlags&ObjectFlagsReference == 0 {
					return c.failTypeFunction(r.site, "cannot complete a recursive rest element")
				}
				shell.AsTypeReference().target = reference.Target()
				shell.AsTypeReference().resolvedTypeArguments = c.getTypeArguments(reference)
				return shell
			}
			return reference
		}
	}
	if t.objectFlags&ObjectFlagsMapped != 0 && c.isGenericMappedType(t) {
		return r.mapped(t)
	}
	properties := c.getPropertiesOfType(t)
	members := core.SameMap(properties, r.member)
	indexes := c.getIndexInfosOfType(t)
	infos := core.SameMap(indexes, func(i *IndexInfo) *IndexInfo {
		key, value := r.visit(i.keyType), r.visit(i.valueType)
		if key == i.keyType && value == i.valueType {
			return i
		}
		return c.newIndexInfo(key, value, i.isReadonly, i.declaration, i.components)
	})
	calls := c.getSignaturesOfType(t, SignatureKindCall)
	ctors := c.getSignaturesOfType(t, SignatureKindConstruct)
	callSignatures := core.SameMap(calls, r.signature)
	constructSignatures := core.SameMap(ctors, r.signature)
	result := r.path[t]
	if result == nil {
		if core.Same(properties, members) && core.Same(indexes, infos) && core.Same(calls, callSignatures) && core.Same(ctors, constructSignatures) {
			return reference
		}
		result = c.newObjectType(ObjectFlagsAnonymous, nil)
	}
	table := ast.NewSymbolTableWithCapacity(len(members))
	for _, p := range members {
		table.Set(p.Name(), p)
	}
	c.setStructuredTypeMembers(result, table, callSignatures, constructSignatures, infos)
	r.capture(result)
	return result
}

func (r *typeFunctionResult) member(p *ast.Symbol) *ast.Symbol {
	if p == nil {
		return nil
	}
	c := r.c
	read := c.getTypeOfSymbol(p)
	value := r.visit(read)
	var write, updated *Type
	if p.Flags()&ast.SymbolFlagsSetAccessor != 0 {
		write = c.getWriteTypeOfSymbol(p)
		updated = r.visit(write)
	}
	if read == value && write == updated {
		return p
	}
	member := c.createSymbolWithType(p, value)
	member.SetCheckFlags(member.CheckFlags() | p.CheckFlags()&(ast.CheckFlagsOptionalParameter|ast.CheckFlagsRestParameter|ast.CheckFlagsLate))
	if updated != nil {
		c.valueSymbolLinks.Get(member).setWriteType(c, updated)
	}
	return member
}

func (r *typeFunctionResult) signature(s *Signature) *Signature {
	c := r.c
	params := s.TypeParameters()
	defer r.bind(params)()
	binders, mapper := r.binders(params)
	receiver := r.member(s.ThisParameter())
	args := core.SameMap(s.Parameters(), r.member)
	ret := c.getReturnTypeOfSignature(s)
	result := r.visit(ret)
	predicate := c.getTypePredicateOfSignature(s)
	updatedPredicate := predicate
	if predicate != nil {
		if v := c.instantiateType(r.visit(predicate.t), mapper); v != predicate.t {
			updatedPredicate = c.newTypePredicate(predicate.kind, predicate.parameterName, predicate.parameterIndex, v)
		}
	}
	if mapper == nil && receiver == s.ThisParameter() && core.Same(args, s.Parameters()) && result == ret && updatedPredicate == predicate {
		return s
	}
	if mapper != nil {
		receiver = c.instantiateSymbol(receiver, mapper)
		args = c.instantiateSymbols(args, mapper)
		result = c.instantiateType(result, mapper)
	}
	// Supplying the predicate (including the negative answer) keeps the
	// checker from recovering a typefn expression from the original signature.
	return c.newSignature(s.flags&SignatureFlagsPropagatingFlags, s.declaration, binders, receiver, args, result, core.OrElse(updatedPredicate, c.noTypePredicate), int(s.minArgumentCount))
}

func (r *typeFunctionResult) binders(params []*Type) ([]*Type, *TypeMapper) {
	c := r.c
	constraints := core.Map(params, c.getConstraintOfTypeParameter)
	defaults := core.Map(params, c.getDefaultFromTypeParameter)
	cons := core.SameMap(constraints, r.visit)
	defs := core.SameMap(defaults, r.visit)
	if core.Same(cons, constraints) && core.Same(defs, defaults) {
		return params, nil
	}
	fresh := core.Map(params, c.cloneTypeParameter)
	m := newTypeMapper(params, fresh)
	for i, p := range fresh {
		p.AsTypeParameter().constraint = core.OrElse(c.instantiateType(cons[i], m), c.noConstraintType)
		p.AsTypeParameter().resolvedDefaultType = core.OrElse(c.instantiateType(defs[i], m), c.noConstraintType)
	}
	return fresh, m
}

func (r *typeFunctionResult) capture(t *Type) {
	c := r.c
	if c.typeFunctionResultParameters == nil {
		c.typeFunctionResultParameters = make(map[*Type][]*Type)
	}
	c.typeFunctionResultParameters[t] = slices.Clone(r.bound)
	t.objectFlags |= ObjectFlagsCouldContainTypeVariablesComputed
	if len(r.bound) > 0 {
		t.objectFlags |= ObjectFlagsCouldContainTypeVariables
	}
}

func (r *typeFunctionResult) conditional(t *Type) *Type {
	c := r.c
	d := t.AsConditionalType()
	outer := slices.Clone(r.bound)
	params := d.root.inferTypeParameters
	defer r.bind(params)()
	binders, m := r.binders(params)
	check, constraint := r.visit(d.checkType), r.visit(d.extendsType)
	a, b := c.getTrueTypeFromConditionalType(t), c.getFalseTypeFromConditionalType(t)
	yes, no := r.visit(a), r.visit(b)
	if m == nil && check == d.checkType && constraint == d.extendsType && yes == a && no == b {
		return t
	}
	root := &ConditionalRoot{
		node: d.root.node, checkType: check, extendsType: c.instantiateType(constraint, m),
		isDistributive: d.root.isDistributive, wholeOperand: d.root.wholeOperand, inferTypeParameters: binders,
		outerTypeParameters: outer, instantiations: make(map[CacheHashKey]*Type),
	}
	if c.typeFunctionResultBranches == nil {
		c.typeFunctionResultBranches = make(map[*ConditionalRoot][2]*Type)
	}
	c.typeFunctionResultBranches[root] = [2]*Type{c.instantiateType(yes, m), no}
	result := c.getConditionalType(root, nil, false, nil)
	root.instantiations[getConditionalTypeKey(outer, nil, false)] = result
	return result
}

func (r *typeFunctionResult) mapped(t *Type) *Type {
	c := r.c
	param := c.getTypeParameterFromMappedType(t)
	constraint := c.getConstraintTypeFromMappedType(t)
	modifiers := c.getModifiersTypeFromMappedType(t)
	name := c.getNameTypeFromMappedType(t)
	template := c.getTemplateTypeFromMappedType(t)
	con, source := r.visit(constraint), r.visit(modifiers)
	pop := r.bind([]*Type{param})
	key, value := r.visit(name), r.visit(template)
	pop()
	if con == constraint && source == modifiers && key == name && value == template {
		return t
	}
	out := c.newObjectType(ObjectFlagsMapped, t.symbol)
	d := out.AsMappedType()
	d.declaration = t.AsMappedType().declaration
	if con != constraint {
		fresh := c.cloneTypeParameter(param)
		fresh.AsTypeParameter().constraint = con
		m := newSimpleTypeMapper(param, fresh)
		key, value = c.instantiateType(key, m), c.instantiateType(value, m)
		param = fresh
	}
	d.typeParameter = param
	d.constraintType, d.modifiersType, d.nameType, d.templateType = con, source, key, value
	r.capture(out)
	return out
}

// A module-level type alias or interface whose declarations refer only to
// native types needs only its actual arguments checked. Prove this transitively,
// including merged declarations. Publish a recursive group only if every edge
// was accounted for. For value queries and imported types use the semantic walk.
func (r *typeFunctionResult) nativeDeclaration(sym *ast.Symbol) bool {
	if sym == nil {
		return false
	}
	if native, found := r.c.typeFunctionNativeDeclarations[sym]; found {
		return native
	}
	group := make(map[*ast.Symbol]bool)
	var declaration func(*ast.Symbol) bool
	var hasWork func(*ast.Node) bool
	hasWork = func(n *ast.Node) bool {
		switch n.Kind {
		case ast.KindTypeFunctionNot, ast.KindTypeFunctionNegate, ast.KindTypeFunctionBinary, ast.KindTypeQuery, ast.KindImportType:
			return true
		case ast.KindComputedPropertyName:
			// Well-known symbol names in the standard library are fixed keys.
			return !r.c.program.IsSourceFileDefaultLibrary(ast.GetSourceFileOfNode(n).PathKey())
		case ast.KindArrayType, ast.KindTupleType:
			if !declaration(r.c.globalArrayType.symbol) || !declaration(r.c.globalReadonlyArrayType.symbol) {
				return true
			}
		case ast.KindTypeReference, ast.KindExpressionWithTypeArguments:
			if !declaration(r.c.resolveTypeReferenceName(n, ast.SymbolFlagsType, true)) {
				return true
			}
		}
		return n.ForEachChild(hasWork)
	}
	declaration = func(s *ast.Symbol) bool {
		if s == nil || len(s.Declarations()) == 0 {
			return false
		}
		if s.Flags()&ast.SymbolFlagsTypeParameter != 0 || group[s] {
			return true
		}
		if native, found := r.c.typeFunctionNativeDeclarations[s]; found {
			return native
		}
		if s.Flags()&(ast.SymbolFlagsTypeAlias|ast.SymbolFlagsInterface) == 0 || s.Flags()&ast.SymbolFlagsClass != 0 {
			return false
		}
		group[s] = true
		foundType := false
		for _, d := range s.Declarations() {
			// A merged value (for example the Array constructor) is independent of the instance type.
			if ast.IsTypeAliasDeclaration(d) || ast.IsInterfaceDeclaration(d) {
				foundType = true
				if !(ast.IsSourceFile(d.Parent) || ast.IsModuleBlock(d.Parent)) || hasWork(d) {
					return false
				}
			} else if ast.IsTypeFunctionDeclaration(d) {
				return false
			}
		}
		return foundType
	}
	native := declaration(sym)
	if r.c.typeFunctionNativeDeclarations == nil {
		r.c.typeFunctionNativeDeclarations = make(map[*ast.Symbol]bool)
	}
	if native {
		for s := range group {
			r.c.typeFunctionNativeDeclarations[s] = true
		}
	}
	r.c.typeFunctionNativeDeclarations[sym] = native
	return native
}
