package checker

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
)

func (c *Checker) newTypefnTerm(d *typefnTerm) *Type {
	state := c.getTypefnState()
	var key keyBuilder
	key.writeInt(len(d.args))
	for _, arg := range d.args {
		if arg == nil {
			key.writeUint32(0)
		} else {
			key.writeType(arg)
		}
	}
	k := typefnTermKey{kind: d.kind, declaration: d.declaration, symbol: d.symbol, expression: d.expression, operator: d.operator, operands: key.hash(), shape: d.shape, values: typefnValueHash(d.values), closure: d.closure}
	if state.terms == nil {
		state.terms = make(map[typefnTermKey]*Type)
	}
	if t := state.terms[k]; t != nil {
		return t
	}
	// Charge retained operands as well as the node before allocating the copy.
	if v := state.vm; v != nil && !v.allocate(d.node, len(d.args)+len(d.values)+1) {
		return c.errorType
	}
	d.args = slices.Clone(d.args)
	d.values = slices.Clone(d.values)
	t := c.newType(TypeFlagsTypefn, ObjectFlagsNone, d.AsType(), typeKindTypefn)
	state.terms[k] = t
	return t
}

// splitTypefn extracts checks before native operations can absorb a result. An
// opaque term is itself a requirement; a guard has a usable shape and a journal.
func splitTypefn(t *Type) (*Type, []*Type) {
	if t == nil || t.flags&TypeFlagsTypefn == 0 {
		return t, nil
	}
	d := t.asTypefnTerm()
	if d.kind == typefnGuardTerm {
		return d.shape, d.args
	}
	return t, []*Type{t}
}

func (c *Checker) guardTypefn(shape *Type, checks []*Type) *Type {
	if shape == nil || c.isErrorType(shape) {
		return shape
	}
	if len(checks) == 0 && shape.flags&TypeFlagsTypefn == 0 {
		return shape
	}
	shape, own := splitTypefn(shape)
	if len(checks) == 0 {
		checks = own
	} else {
		checks = append(slices.Clone(checks), own...)
	}
	var pending []*Type
	seen := make(map[*Type]bool)
	for _, check := range checks {
		if check == nil {
			continue
		}
		if c.isErrorType(check) {
			return c.errorType
		}
		_, requirements := splitTypefn(check)
		for _, requirement := range requirements {
			if !seen[requirement] {
				seen[requirement] = true
				pending = append(pending, requirement)
			}
		}
	}
	if len(pending) == 0 || len(pending) == 1 && pending[0] == shape {
		return shape
	}
	// Canonical order makes identity independent of journal concatenation order.
	slices.SortFunc(pending, func(a, b *Type) int { return int(a.id) - int(b.id) })
	return c.newTypefnTerm(&typefnTerm{kind: typefnGuardTerm, node: pending[0].asTypefnTerm().node, shape: shape, args: pending})
}

func (c *Checker) splitTypefnOperands(types []*Type) ([]*Type, []*Type) {
	profileEnabled := c.compilerOptions.GetTypefnVersion() == 1
	// A checker without DSL state cannot own any deferred terms. Preserve the
	// native hot path, but still propagate recovery errors in enabled projects.
	if c.typefn == nil && !profileEnabled {
		return types, nil
	}

	var values []*Type
	var checks []*Type
	for i, t := range types {
		value, requirements := splitTypefn(t)
		if profileEnabled && c.isErrorType(t) {
			requirements = []*Type{t}
		}
		if len(requirements) != 0 {
			if values == nil {
				values = slices.Clone(types)
			}
			values[i] = value
			checks = append(checks, requirements...)
		}
	}
	if values == nil {
		return types, nil
	}
	return values, checks
}

func (c *Checker) instantiateTypefn(t *Type, m *TypeMapper) *Type {
	d := t.asTypefnTerm()
	return c.runTypefn(d.node, func(v *typefnVM) *Type {
		saved := v.replaySite
		v.replaySite = c.currentNode
		defer func() { v.replaySite = saved }()
		if !v.step(d.node) {
			return c.errorType
		}
		args := c.instantiateTypes(d.args, m)
		values := c.instantiateTypefnValues(d.values, m)
		closure := c.instantiateTypefnClosure(d.closure, m)
		if v.failed {
			return c.errorType
		}
		for _, arg := range args {
			if arg != nil && c.isErrorType(arg) {
				return c.errorType
			}
		}
		if d.kind == typefnGuardTerm {
			shape := c.instantiateType(d.shape, m)
			if shape == d.shape && core.Same(args, d.args) {
				return t
			}
			return c.guardTypefn(shape, args)
		}
		if core.Same(args, d.args) && slices.Equal(values, d.values) && closure == d.closure {
			return t
		}
		switch d.kind {
		case typefnBinaryTerm:
			return v.binary(d.node, d.operator, args[0], args[1])
		case typefnUnaryTerm:
			return v.unary(d.node, d.operator, args[0])
		case typefnCallTerm:
			return v.callWithClosure(d.node, d.declaration, values, closure)
		case typefnNativeCallTerm:
			return v.apply(d.node, d.symbol, args)
		case typefnExpressionTerm:
			return v.replayExpression(d.expression, values)
		}
		panic("unknown typefn recipe")
	})
}

// Scalar observation needs only the outer Type, not every latent member of a
// native graph. Constraints are proofs of eligibility, never observed values.
func typefnUnresolved(t *Type) bool {
	t, _ = splitTypefn(t)
	// Native permissive relation probes are not concrete user inputs.
	if t == t.checker.wildcardType {
		return true
	}
	switch {
	case t.flags&(TypeFlagsTypeParameter|TypeFlagsTypefn) != 0:
		return true
	case t.flags&TypeFlagsUnionOrIntersection != 0:
		return core.Some(t.Types(), typefnUnresolved)
	case t.flags&TypeFlagsTemplateLiteral != 0:
		return core.Some(t.AsTemplateLiteralType().types, typefnUnresolved)
	case t.flags&TypeFlagsStringMapping != 0:
		return typefnUnresolved(t.AsStringMappingType().target)
	case t.flags&TypeFlagsIndex != 0:
		return typefnUnresolved(t.AsIndexType().target)
	case t.flags&TypeFlagsIndexedAccess != 0:
		d := t.AsIndexedAccessType()
		return typefnUnresolved(d.objectType) || typefnUnresolved(d.indexType)
	case t.flags&TypeFlagsConditional != 0:
		// Native conditional Types remain symbolic until native instantiation selects a branch.
		return true
	case t.flags&TypeFlagsSubstitution != 0:
		return typefnUnresolved(t.AsSubstitutionType().baseType)
	}
	return false
}

func (v *typefnVM) observe(t *Type) *Type {
	shape, checks := splitTypefn(t)
	for _, check := range checks {
		if v.checkSet == nil {
			v.checkSet = make(map[*Type]bool)
		}
		if !v.checkSet[check] {
			v.checkSet[check] = true
			v.checks = append(v.checks, check)
		}
	}
	return shape
}

func (v *typefnVM) deferExpression(e *typefnExpression, slots []typefnValue) *Type {
	frozen := make([]typefnValue, len(slots))
	var capture func(*typefnExpression)
	capture = func(e *typefnExpression) {
		if e == nil {
			return
		}
		if e.slot >= 0 {
			frozen[e.slot] = slots[e.slot]
		}
		for _, a := range e.args {
			capture(a)
		}
		for _, k := range e.keys {
			capture(k)
		}
	}
	capture(e)
	return v.c.newTypefnTerm(&typefnTerm{kind: typefnExpressionTerm, node: e.node, expression: e, values: frozen})
}

// DSL-created anonymous objects have no source declaration/symbol for the native
// object instantiator. Instantiate their immutable property table explicitly.
func (c *Checker) newTypefnObject(members ast.SymbolTable) *Type {
	return c.newTypefnObjectEx(members, nil, nil, nil)
}

func (c *Checker) newTypefnObjectEx(members ast.SymbolTable, calls, constructs []*Signature, indexInfos []*IndexInfo) *Type {
	state := c.getTypefnState()
	t := c.newAnonymousType(nil, members, calls, constructs, indexInfos)
	if state.objects == nil {
		state.objects = make(map[*Type]bool)
	}
	state.objects[t] = true
	return t
}

func (c *Checker) instantiateTypefnObject(t *Type, m *TypeMapper) *Type {
	members := ast.NewSymbolTable()
	changed := false
	for name, property := range t.AsObjectType().members.All() {
		old := c.getTypeOfSymbol(property)
		value := c.instantiateType(old, m)
		if c.isErrorType(value) {
			return c.errorType
		}
		declared := c.typefn.propertyValues[property]
		mapped := declared
		if declared != nil {
			mapped = c.instantiateType(declared, m)
			if c.isErrorType(mapped) {
				return c.errorType
			}
		}
		if value == old && mapped == declared {
			members.Set(name, property)
			continue
		}
		changed = true
		p := c.newSymbolEx(property.Flags(), name, property.CheckFlags())
		c.valueSymbolLinks.Get(p).resolvedType = value
		if mapped != nil {
			c.rememberTypefnProperty(p, mapped)
		}
		c.valueSymbolLinks.Get(p).setNameType(c, c.valueSymbolLinks.Get(property).getNameType())
		members.Set(name, p)
	}
	oldIndexInfos := c.getIndexInfosOfType(t)
	indexInfos := c.instantiateIndexInfos(oldIndexInfos, m)
	for i, info := range indexInfos {
		if c.isErrorType(info.valueType) {
			return c.errorType
		}
		changed = changed || info != oldIndexInfos[i]
	}
	calls := c.instantiateTypefnSignatures(c.getSignaturesOfType(t, SignatureKindCall), m)
	constructs := c.instantiateTypefnSignatures(c.getSignaturesOfType(t, SignatureKindConstruct), m)
	changed = changed || !slices.Equal(calls, c.getSignaturesOfType(t, SignatureKindCall)) || !slices.Equal(constructs, c.getSignaturesOfType(t, SignatureKindConstruct))
	if !changed {
		return t
	}
	return c.newTypefnObjectEx(members, calls, constructs, indexInfos)
}

// Inspect only eager native Type expressions. Native object members, mapped
// members and signatures remain lazy; do not expand recursive object graphs or
// lift obligations out of their generic signature binders.
func (c *Checker) typefnSyntaxHasCalculations(node *ast.Node) bool {
	if c.compilerOptions.GetTypefnVersion() != 1 {
		return false
	}
	state := c.getTypefnState()
	if result, ok := state.syntaxChecks[node]; ok {
		return result
	}
	seen := make(map[*ast.Symbol]bool)
	var visit func(*ast.Node) bool
	visit = func(n *ast.Node) bool {
		switch n.Kind {
		case ast.KindTypeLiteral, ast.KindMappedType, ast.KindFunctionType, ast.KindConstructorType:
			return false
		case ast.KindTypeReference:
			symbol := c.resolveEntityName(n.AsTypeReferenceNode().TypeName, ast.SymbolFlagsType, true, false, nil)
			if symbol != nil && !seen[symbol] {
				seen[symbol] = true
				for _, declaration := range symbol.Declarations() {
					if !ast.IsTypeAliasDeclaration(declaration) && !ast.IsInterfaceDeclaration(declaration) && !ast.IsClassDeclaration(declaration) {
						continue
					}
					parameters := declaration.TypeParameters()
					for i := len(n.TypeArguments()); i < len(parameters); i++ {
						if value := parameters[i].AsTypeParameterDeclaration().DefaultType; value != nil && visit(value) {
							return true
						}
					}
				}
				if core.Find(symbol.Declarations(), ast.IsTypeFunctionDeclaration) != nil {
					return true
				}
				if decl := core.Find(symbol.Declarations(), ast.IsTypeAliasDeclaration); decl != nil && visit(decl.Type()) {
					return true
				}
			}
		}
		return n.ForEachChild(visit)
	}
	result := visit(node)
	if state.syntaxChecks == nil {
		state.syntaxChecks = make(map[*ast.Node]bool)
	}
	state.syntaxChecks[node] = result
	return result
}

func (c *Checker) typefnChecksFromSyntax(nodes []*ast.Node) []*Type {
	var checks []*Type
	for _, node := range nodes {
		if c.typefnSyntaxHasCalculations(node) {
			value := c.getTypeFromTypeNode(node)
			_, requirements := splitTypefn(value)
			if c.isErrorType(value) {
				requirements = []*Type{value}
			}
			checks = append(checks, requirements...)
		}
	}
	return checks
}

func (c *Checker) typefnNativeArguments(node *ast.Node, symbol *ast.Symbol) ([]*Type, []*Type) {
	if c.compilerOptions.GetTypefnVersion() != 1 || !c.typefnSyntaxHasCalculations(node) {
		return nil, nil
	}
	parameters := c.getLocalTypeParametersOfClassOrInterfaceOrTypeAlias(symbol)
	if len(node.TypeArguments()) < c.getMinTypeArgumentCount(parameters) || len(node.TypeArguments()) > len(parameters) {
		return nil, nil
	}
	args := c.getTypeArgumentsFromNode(node)
	return c.fillMissingTypeArguments(args, parameters, c.getMinTypeArgumentCount(parameters), false), parameters
}

func (c *Checker) pendingTypefnNativeApplication(node *ast.Node, symbol *ast.Symbol) *Type {
	args, parameters := c.typefnNativeArguments(node, symbol)
	if !core.Some(args, func(t *Type) bool { return t.flags&TypeFlagsTypefn != 0 }) {
		return nil
	}
	mapper := newTypeMapper(parameters, args)
	for i, parameter := range parameters {
		if bound := c.getConstraintOfTypeParameter(parameter); bound != nil {
			bound = c.instantiateType(bound, mapper)
			value, _ := splitTypefn(args[i])
			if !c.isTypeAssignableTo(value, bound) && (typefnUnresolved(value) || typefnUnresolved(bound)) {
				return c.newTypefnTerm(&typefnTerm{kind: typefnNativeCallTerm, node: node, symbol: symbol, args: args})
			}
		}
	}
	return nil
}
