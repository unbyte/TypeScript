package checker

import (
	"fmt"
	"slices"

	"github.com/microsoft/TypeScript/tsc/internal/typefunction"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/diagnostics"
)

// The result of an open application is an opaque type variable. Its arguments
// participate in substitution, but never in inference from its result.
type typeFunctionApplication struct {
	completed   *Type
	use         *ast.Node
	declaration *ast.Node
	expression  *ast.Node // An operator nested in a native type captures its lexical bindings.
	builtin     string
	arguments   []*Type
}

func (t *Type) typeFunctionApplication() *typeFunctionApplication {
	if t != nil && t.flags&TypeFlagsTypeParameter != 0 {
		return t.checker.typeFunctionApplications[t]
	}
	return nil
}

func (c *Checker) newTypeFunctionApplication(app typeFunctionApplication) *Type {
	var symbol *ast.Symbol
	if app.declaration != nil {
		symbol = c.getSymbolOfDeclaration(app.declaration)
	}
	t := c.newTypeParameter(symbol)
	tp := t.AsTypeParameter()
	tp.constraint = c.noConstraintType
	if c.typeFunctionApplications == nil {
		c.typeFunctionApplications = make(map[*Type]*typeFunctionApplication)
	}
	app.completed = nil
	c.typeFunctionApplications[t] = &app
	if symbol != nil {
		t.alias = &TypeAlias{symbol: symbol, typeArguments: app.arguments}
	}
	return t
}

func (c *Checker) checkTypeFunctionArgumentCount(node, decl *ast.Node) bool {
	name := decl.Name().Text()
	if node.TypeArgumentList() == nil {
		c.error(node, diagnostics.Type_function_Colon_0, name+" requires an explicit type argument list")
		return false
	}
	parameters := decl.TypeParameters()
	count := len(node.TypeArguments())
	if count == len(parameters) {
		return true
	}
	minimum := len(parameters)
	for minimum > 0 && parameters[minimum-1].AsTypeParameterDeclaration().DefaultType != nil {
		minimum--
	}
	if count >= minimum && count <= len(parameters) {
		return true
	}
	if minimum == len(parameters) {
		c.error(node, diagnostics.Type_function_Colon_0, fmt.Sprintf("%s requires %d explicit type arguments", name, minimum))
	} else {
		c.error(node, diagnostics.Generic_type_0_requires_between_1_and_2_type_arguments, name, minimum, len(parameters))
	}
	return false
}

func (c *Checker) typeFunctionReference(node, decl *ast.Node) *Type {
	if !c.checkTypeFunctionArgumentCount(node, decl) {
		return c.errorType
	}
	arguments := make([]*Type, len(node.TypeArguments()))
	for i, a := range node.TypeArguments() {
		arguments[i] = c.getTypeArgumentForParameter(a, decl.TypeParameters()[i])
	}
	if len(arguments) < len(decl.TypeParameters()) {
		parameters := core.Map(decl.TypeParameters(), func(p *ast.Node) *Type { return c.getDeclaredTypeOfTypeParameter(c.getSymbolOfDeclaration(p)) })
		arguments = c.fillMissingTypeArguments(arguments, parameters, c.getMinTypeArgumentCount(parameters), false)
	}
	return c.newTypeFunctionApplication(typeFunctionApplication{use: node, declaration: decl, arguments: arguments})
}

func (c *Checker) getTypeArgumentForParameter(argument, parameter *ast.Node) *Type {
	if arity := typeFunctionSlots(parameter); arity != 0 {
		return c.readTypeFunctionConstructor(argument, arity)
	}
	return c.getTypeFromTypeNode(argument)
}

func (c *Checker) typeFunctionBuiltinName(node *ast.Node) string {
	if ns := c.TypeFunctionBuiltinNamespace(node); ns != "" {
		return ns + "." + node.AsTypeReferenceNode().TypeName.AsQualifiedName().Right.Text()
	}
	return ""
}

func (c *Checker) TypeFunctionBuiltinNamespace(node *ast.Node) string {
	if !ast.IsTypeReferenceNode(node) {
		return ""
	}
	name := node.AsTypeReferenceNode().TypeName
	if !ast.IsQualifiedName(name) {
		return ""
	}
	q := name.AsQualifiedName()
	if !ast.IsIdentifier(q.Left) {
		return ""
	}
	prefix := q.Left.Text()
	if !typefunction.IsNamespace(prefix) {
		return ""
	}
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if ast.IsTypeFunctionBlock(parent) {
			if c.resolveName(q.Left, prefix, ast.SymbolFlagsNamespace|ast.SymbolFlagsType|ast.SymbolFlagsAlias, nil, false, false) != nil {
				return ""
			}
			return prefix
		}
		if ast.IsSourceFile(parent) {
			break
		}
	}
	return ""
}

func (c *Checker) typeFunctionClosed(t *Type, path, bound map[*Type]bool) bool {
	if t == nil || bound[getNonDistributedTypeParameter(t)] || path[t] {
		return true
	}
	if t.flags&TypeFlagsStructuredOrInstantiable == 0 {
		return true
	}
	path[t] = true
	defer delete(path, t)
	closed := func(t *Type) bool { return c.typeFunctionClosed(t, path, bound) }
	all := func(types []*Type) bool {
		for _, part := range types {
			if !closed(part) {
				return false
			}
		}
		return true
	}
	// A module-level alias can depend only on its explicit parameters. This
	// shortcut also bounds the inspection of recursively expanding aliases.
	if t.typeFunctionApplication() == nil && t.alias != nil && t.alias.symbol != nil {
		if decl := ast.GetDeclarationOfKind(t.alias.symbol, ast.KindTypeAliasDeclaration); decl != nil && (ast.IsSourceFile(decl.Parent) || ast.IsModuleBlock(decl.Parent)) {
			return all(t.alias.typeArguments)
		}
	}
	switch {
	case t.typeFunctionApplication() != nil:
		app := t.typeFunctionApplication()
		if app.builtin == "fn.ref" && c.typeFunctionExecution != nil && c.typeFunctionExecution.fragments {
			return false
		}
		return all(app.arguments)
	case t.flags&TypeFlagsTypeParameter != 0:
		if ctor := c.typeFunctionConstructors[t]; ctor != nil {
			return all(ctor.arguments) && all(ctor.captured)
		}
		return false
	case t.flags&TypeFlagsUnionOrIntersection != 0:
		return all(t.Types())
	case t.flags&TypeFlagsIndexedAccess != 0:
		a := t.AsIndexedAccessType()
		return closed(a.objectType) && closed(a.indexType)
	case t.flags&TypeFlagsIndex != 0:
		return closed(t.AsIndexType().target)
	case t.flags&TypeFlagsTemplateLiteral != 0:
		return all(t.AsTemplateLiteralType().types)
	case t.flags&TypeFlagsStringMapping != 0:
		return closed(t.AsStringMappingType().target)
	case t.flags&TypeFlagsSubstitution != 0:
		a := t.AsSubstitutionType()
		return closed(a.baseType) && closed(a.constraint)
	case t.flags&TypeFlagsConditional != 0:
		a := t.AsConditionalType()
		for _, dep := range a.root.outerTypeParameters {
			if !closed(c.instantiateType(dep, a.mapper)) {
				return false
			}
		}
		return closed(a.checkType)
	case t.objectFlags&ObjectFlagsReference != 0:
		return all(c.getTypeArguments(t))
	case t.flags&TypeFlagsObject != 0:
		for _, member := range c.getPropertiesOfType(t) {
			if !closed(c.getTypeOfSymbol(member)) {
				return false
			}
		}
		for _, info := range c.getIndexInfosOfType(t) {
			if !closed(info.keyType) || !closed(info.valueType) {
				return false
			}
		}
		for _, kind := range []SignatureKind{SignatureKindCall, SignatureKindConstruct} {
			for _, sig := range c.getSignaturesOfType(t, kind) {
				params := sig.TypeParameters()
				var added []*Type
				for _, p := range params {
					if !bound[p] {
						bound[p] = true
						added = append(added, p)
					}
				}
				ok := closed(c.getReturnTypeOfSignature(sig))
				for _, p := range params {
					ok = ok && closed(c.getConstraintOfTypeParameter(p)) && closed(c.getDefaultFromTypeParameter(p))
				}
				for _, p := range sig.Parameters() {
					ok = ok && closed(c.getTypeOfSymbol(p))
				}
				if sig.ThisParameter() != nil {
					ok = ok && closed(c.getTypeOfSymbol(sig.ThisParameter()))
				}
				for _, p := range added {
					delete(bound, p)
				}
				if !ok {
					return false
				}
			}
		}
	}
	return true
}

type typeFunctionExecution struct {
	records        map[*Type]bool
	selectedValues map[*Type]*Type // Explicit computations reached in selected conditional arms.
	fragments      bool
	trace          []string
	refs           map[*Type]*Type
	site           *ast.Node
	stack          []*typeFunctionApplication
	steps          int
	failed         bool
}

func (c *Checker) failTypeFunction(node *ast.Node, message string) *Type {
	root := c.typeFunctionExecution
	if root != nil && c.ctx != nil && c.ctx.Err() != nil {
		root.failed = true
	}
	if root != nil && !root.failed {
		root.failed = true
		d := c.error(root.site, diagnostics.Type_function_Colon_0, message)
		if d != nil {
			root.addFrames(d, node, message)
		}
	}
	return c.anyType
}

func (c *Checker) typeFunctionStep(node *ast.Node) bool {
	return c.typeFunctionCharge(node, 1)
}

func (c *Checker) typeFunctionCharge(node *ast.Node, units int) bool {
	root := c.typeFunctionExecution
	root.steps += units
	if root.steps > 100000 {
		c.failTypeFunction(node, "work limit exceeded")
	}
	if c.ctx != nil && c.ctx.Err() != nil {
		root.failed = true
	}
	return !root.failed
}

func (c *Checker) typeFunctionReady(app *typeFunctionApplication) bool {
	if app.builtin != "" && c.typeFunctionExecution != nil {
		// Reflection may manipulate fragments bound by the signature under construction.
		return true
	}
	for _, arg := range app.arguments {
		if !c.typeFunctionClosed(arg, map[*Type]bool{}, map[*Type]bool{}) {
			return false
		}
	}
	return true
}

func (c *Checker) forceTypeFunction(t *Type) (result *Type) {
	app := t.typeFunctionApplication()
	if app == nil {
		return t
	}
	if c.typeFunctionExecution == nil && app.completed != nil {
		return app.completed
	}
	if !c.typeFunctionReady(app) {
		return t
	}
	if c.typeFunctionExecution == nil {
		c.typeFunctionExecution = &typeFunctionExecution{site: app.use}
		defer func() {
			root := c.typeFunctionExecution
			if !root.failed && (c.ctx == nil || c.ctx.Err() == nil) {
				// Only complete roots can be reused. Nested calls always share and
				// charge the active root's budget, including on a warm checker.
				app.completed = result
			}
			c.publishTypeFunctionTrace(app)
			c.typeFunctionExecution = nil
		}()
	}
	root := c.typeFunctionExecution
	if !c.typeFunctionStep(app.use) {
		return c.anyType
	}
	if len(root.stack) >= 100 {
		return c.failTypeFunction(app.use, "call depth limit exceeded")
	}
	for _, active := range root.stack {
		if active.declaration == app.declaration && active.expression == app.expression && active.builtin == app.builtin && slices.Equal(active.arguments, app.arguments) {
			return c.failTypeFunction(app.use, "recursive call with unchanged arguments")
		}
	}
	root.stack = append(root.stack, app)
	defer func() { root.stack = root.stack[:len(root.stack)-1] }()
	if app.expression != nil {
		template := c.typeNodeLinks.Get(app.expression).resolvedType.typeFunctionApplication()
		frame := typeFunctionFrame{c: c, sources: template.arguments, values: app.arguments}
		return frame.operand(app.expression)
	}
	args := make([]*Type, len(app.arguments))
	for i, arg := range app.arguments {
		args[i] = c.forceTypeFunction(arg)
		if root.failed {
			return c.anyType
		}
	}
	if app.builtin == "@apply" {
		return c.applyTypeFunctionConstructor(app.use, args[0], args[1:])
	}
	if app.builtin != "" {
		return c.runTypeFunctionBuiltin(app.use, app.builtin, args)
	}
	frame := typeFunctionFrame{c: c}
	for i, p := range app.declaration.TypeParameters() {
		frame.bind(p, args[i])
	}
	value, flow := frame.block(app.declaration.Body())
	if root.failed {
		return c.anyType
	}
	if flow != typeFunctionReturned {
		return c.failTypeFunction(app.declaration, "execution reached the end without a return")
	}
	if len(root.stack) == 1 {
		return c.finishTypeFunctionResult(value, app.use)
	}
	if c.typeFunctionConstructors[value] != nil {
		return c.failTypeFunction(app.use, "a constructor cannot be returned as a type")
	}
	return value
}

type typeFunctionFrame struct {
	c               *Checker
	sources, values []*Type
}

func (f *typeFunctionFrame) bind(node *ast.Node, value *Type) {
	f.sources = append(f.sources, f.c.getDeclaredTypeOfSymbol(f.c.getSymbolOfDeclaration(node)))
	f.values = append(f.values, value)
}

func (f *typeFunctionFrame) operand(node *ast.Node) *Type {
	c := f.c
	if !c.typeFunctionStep(node) {
		return c.anyType
	}
	switch node.Kind {
	case ast.KindParenthesizedType:
		return f.operand(node.AsParenthesizedTypeNode().Type)
	case ast.KindTypeFunctionNot:
		v := f.operand(node.AsTypeFunctionNot().Operand)
		if v != c.regularTrueType && v != c.regularFalseType {
			return c.failTypeFunction(node, "expected a boolean literal")
		}
		return f.boolean(v == c.regularFalseType)
	case ast.KindTypeFunctionNegate:
		args := []*Type{f.operand(node.AsTypeFunctionNegate().Operand)}
		if !c.chargeTypeFunctionOperands(node, args) {
			return c.anyType
		}
		return c.typeFunctionNumericOperator(node, "negate", args)
	case ast.KindTypeFunctionBinary:
		return f.binaryOperand(node.AsTypeFunctionBinary())
	}
	sources := slices.Clone(f.sources)
	values := slices.Clone(f.values)
	switch node.Kind {
	case ast.KindConditionalType:
		return f.selectedOperand(node, sources, values)
	case ast.KindFunctionType, ast.KindConstructorType, ast.KindMappedType:
		// Native instantiation introduces binders.
	case ast.KindTypeReference:
		template := c.getTypeFromTypeNode(node)
		value := template
		if len(sources) > 0 {
			value = c.instantiateType(template, newTypeMapper(sources, values))
		}
		if app := value.typeFunctionApplication(); app != nil && !c.typeFunctionReady(app) {
			return value
		}
		fallthrough
	default:
		node.ForEachChild(func(n *ast.Node) bool { return f.forceNestedOperand(n, &sources, &values) })
	}
	if c.typeFunctionExecution.failed {
		return c.anyType
	}
	t := c.getTypeFromTypeNode(node)
	if len(sources) > 0 {
		mapper := newTypeMapper(sources, values)
		if ast.IsTypeLiteralNode(node) && len(sources) > len(f.sources) && t.objectFlags&ObjectFlagsAnonymous != 0 {
			// A closed literal has no lexical type parameters. It still needs a
			// mapper when an explicit member computation has just run.
			t = c.instantiateAnonymousType(t, mapper, nil)
		} else {
			t = c.instantiateType(t, mapper)
		}
	}
	return c.getRegularTypeOfLiteralType(c.forceTypeFunction(t))
}

// Explicit operands of a newly constructed type execute in the owning frame.
// A reference to an existing type keeps the host checker's ordinary member laziness.
func (f *typeFunctionFrame) forceNestedOperand(node *ast.Node, sources, values *[]*Type) bool {
	c := f.c
	switch node.Kind {
	case ast.KindFunctionType, ast.KindConstructorType, ast.KindMethodSignature, ast.KindCallSignature, ast.KindConstructSignature, ast.KindMappedType:
		return false
	case ast.KindConditionalType:
		f.selectedOperand(node, *sources, *values)
		return c.typeFunctionExecution.failed
	case ast.KindTypeReference:
		if core.Some(node.TypeArguments(), typeFunctionHole) && c.typeFunctionConstructorArity(node) > 0 {
			return false
		}
		fallthrough
	case ast.KindTypeFunctionNot, ast.KindTypeFunctionNegate, ast.KindTypeFunctionBinary:
		template := c.getTypeFromTypeNode(node)
		if template.typeFunctionApplication() != nil {
			t := template
			if len(*sources) > 0 {
				t = c.instantiateType(template, newTypeMapper(slices.Clone(*sources), slices.Clone(*values)))
			}
			value := c.forceTypeFunction(t)
			if value != t {
				*sources = append(*sources, template)
				*values = append(*values, value)
			}
			return c.typeFunctionExecution.failed
		}
	}
	return node.ForEachChild(func(child *ast.Node) bool { return f.forceNestedOperand(child, sources, values) })
}

// The checker selects native conditional arms, including distributive and infer
// cases. Inspect only selected shapes written in that operand. Existing named
// objects retain their ordinary lazy member resolution.
func (f *typeFunctionFrame) selectedOperand(node *ast.Node, sources, values []*Type) *Type {
	c := f.c
	t := c.getTypeFromTypeNode(node)
	if len(sources) != 0 {
		t = c.instantiateType(t, newTypeMapper(slices.Clone(sources), slices.Clone(values)))
	}
	visited := make(map[*Type]bool)
	var force func(*Type) *Type
	force = func(v *Type) *Type {
		if app := v.typeFunctionApplication(); app != nil {
			value := c.forceTypeFunction(v)
			root := c.typeFunctionExecution
			if value != v && !root.failed {
				if root.selectedValues == nil {
					root.selectedValues = make(map[*Type]*Type)
				}
				root.selectedValues[v] = value
			}
			return value
		}
		if visited[v] || c.typeFunctionExecution.failed {
			return v
		}
		visited[v] = true
		switch {
		case v.flags&TypeFlagsUnion != 0:
			return c.getUnionType(core.Map(v.Types(), force))
		case v.IsTupleType():
			args := c.getTypeArguments(v)
			mapped := core.Map(args, force)
			if !slices.Equal(args, mapped) {
				target := v.TargetTupleType()
				return c.createTupleTypeEx(mapped, target.elementInfos, target.readonly)
			}
		case v.flags&TypeFlagsObject != 0 && v.symbol != nil:
			if d := ast.GetDeclarationOfKind(v.symbol, ast.KindTypeLiteral); d != nil && node.Pos() <= d.Pos() && d.End() <= node.End() && ast.GetSourceFileOfNode(node) == ast.GetSourceFileOfNode(d) {
				for _, p := range c.getPropertiesOfType(v) {
					force(c.getTypeOfSymbol(p))
					if c.typeFunctionExecution.failed {
						break
					}
				}
			}
		}
		return v
	}
	return force(t)
}

func (f *typeFunctionFrame) boolean(value bool) *Type {
	if value {
		return f.c.regularTrueType
	}
	return f.c.regularFalseType
}

type typeFunctionFlow uint8

const (
	typeFunctionNext typeFunctionFlow = iota
	typeFunctionReturned
	typeFunctionBreak
	typeFunctionContinue
)

func (f *typeFunctionFrame) block(node *ast.Node) (*Type, typeFunctionFlow) {
	start := len(f.sources)
	defer func() { f.sources = f.sources[:start]; f.values = f.values[:start] }()
	for _, s := range node.AsTypeFunctionBlock().Statements.Nodes {
		if !f.c.typeFunctionStep(s) {
			break
		}
		value, flow := f.statement(s)
		if flow != typeFunctionNext {
			return value, flow
		}
	}
	return nil, typeFunctionNext
}

func (f *typeFunctionFrame) statement(node *ast.Node) (*Type, typeFunctionFlow) {
	c := f.c
	switch node.Kind {
	case ast.KindTypeFunctionReturn:
		return f.operand(node.AsTypeFunctionReturn().Operand), typeFunctionReturned
	case ast.KindTypeFunctionBinding:
		b := node.AsTypeFunctionBinding()
		value := f.operand(b.Initializer)
		if b.Destructured {
			items, ok := typeFunctionTuple(c, value)
			if !ok || len(items) != len(b.Bindings.Nodes) {
				c.failTypeFunction(node, "tuple does not match the binding list")
				break
			}
			for i, name := range b.Bindings.Nodes {
				f.bind(name, items[i])
			}
		} else {
			f.bind(b.Bindings.Nodes[0], value)
		}
	case ast.KindTypeFunctionAssignment:
		a := node.AsTypeFunctionAssignment()
		symbol := c.resolveName(a.Target, a.Target.Text(), ast.SymbolFlagsTypeParameter, nil, false, false)
		if symbol != nil {
			i := slices.Index(f.sources, c.getDeclaredTypeOfSymbol(symbol))
			if i >= 0 {
				f.values[i] = f.operand(a.Value)
				break
			}
		}
		c.failTypeFunction(node, "assignment requires a bound local")
	case ast.KindTypeFunctionCallStatement:
		operand := node.AsTypeFunctionCallStatement().Operand
		if ast.IsTypeReferenceNode(operand) && ast.IsIdentifier(operand.AsTypeReferenceNode().TypeName) {
			switch operand.AsTypeReferenceNode().TypeName.Text() {
			case "fail":
				args := operand.TypeArguments()
				if len(args) == 0 {
					c.failTypeFunction(node, "fail requires a diagnostic message")
					break
				}
				message := f.operand(args[0])
				if message.flags&TypeFlagsStringLiteral == 0 {
					c.failTypeFunction(node, "fail requires a string literal")
				} else {
					text := message.AsLiteralType().value.(string)
					if len(args) == 2 {
						c.failTypeFunctionAtSpan(node, text, f.operand(args[1]))
					} else {
						c.failTypeFunction(node, text)
					}
				}
				return nil, typeFunctionNext
			case "log":
				f.log(operand)
				return nil, typeFunctionNext
			}
		}
		f.operand(operand)
	case ast.KindTypeFunctionIf, ast.KindTypeFunctionWhile:
		condition, block, alternative := ast.TypeFunctionBranchParts(node)
		for c.typeFunctionStep(node) {
			v := f.operand(condition)
			if v == c.regularFalseType {
				if alternative != nil {
					if ast.IsTypeFunctionBlock(alternative) {
						return f.block(alternative)
					}
					return f.statement(alternative)
				}
				break
			}
			if v != c.regularTrueType {
				c.failTypeFunction(node, "expected a boolean literal")
				break
			}
			value, flow := f.block(block)
			if node.Kind == ast.KindTypeFunctionIf || flow == typeFunctionReturned {
				return value, flow
			}
			if flow == typeFunctionBreak {
				break
			}
		}
	case ast.KindTypeFunctionFor:
		loop := node.AsTypeFunctionFor()
		seq := f.operand(loop.Sequence)
		if !seq.IsTupleType() || seq.TargetTupleType().combinedFlags & ^ElementFlagsRequired != 0 {
			c.failTypeFunction(node, "for requires a tuple of present elements")
			break
		}
		for _, item := range c.getTypeArguments(seq) {
			if !c.typeFunctionStep(node) {
				break
			}
			f.bind(loop.Binding, item)
			value, flow := f.block(loop.Block)
			f.sources = f.sources[:len(f.sources)-1]
			f.values = f.values[:len(f.values)-1]
			if flow == typeFunctionReturned {
				return value, flow
			}
			if flow == typeFunctionBreak || c.typeFunctionExecution.failed {
				break
			}
		}
	case ast.KindTypeFunctionMatch:
		m := node.AsTypeFunctionMatch()
		v := f.operand(m.Operand)
		for _, a := range m.Arms.Nodes {
			arm := a.AsTypeFunctionArm()
			start := len(f.sources)
			if arm.Pattern == nil || f.match(v, arm.Pattern) {
				value, flow := f.block(arm.Block)
				f.sources = f.sources[:start]
				f.values = f.values[:start]
				return value, flow
			}
			if c.typeFunctionExecution.failed {
				return nil, typeFunctionNext
			}
		}
		c.failTypeFunction(node, "no match arm accepted the input")
	case ast.KindBreakStatement:
		return nil, typeFunctionBreak
	case ast.KindContinueStatement:
		return nil, typeFunctionContinue
	}
	return nil, typeFunctionNext
}

func (c *Checker) instantiateTypeFunction(t *Type, app *typeFunctionApplication, m *TypeMapper) *Type {
	args := c.instantiateTypes(app.arguments, m)
	if core.Same(args, app.arguments) {
		return t
	}
	instantiated := *app
	instantiated.arguments = args
	return c.newTypeFunctionApplication(instantiated)
}

// Inference binders inside a native conditional belong to that conditional.
func typeFunctionPatternCaptures(pattern *ast.Node) []*ast.Node {
	var names []*ast.Node
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindInferType {
			p := n.AsInferTypeNode().TypeParameter
			names = append(names, p)
			return false
		}
		if n.Kind == ast.KindConditionalType {
			return false
		}
		return n.ForEachChild(visit)
	}
	visit(pattern)
	return names
}

func (f *typeFunctionFrame) match(value *Type, pattern *ast.Node) bool {
	c := f.c
	target := f.operand(pattern)
	names := typeFunctionPatternCaptures(pattern)
	if len(names) == 0 {
		return c.isTypeAssignableTo(value, target)
	}
	if pattern.Kind != ast.KindInferType && value.flags&(TypeFlagsAny|TypeFlagsNever) != 0 {
		c.failTypeFunction(pattern, "cannot capture structure from any or never")
		return false
	}
	parameters := core.Map(names, func(p *ast.Node) *Type { return c.getDeclaredTypeOfSymbol(c.getSymbolOfDeclaration(p)) })
	inference := c.newInferenceContext(parameters, nil, InferenceFlagsNoDefault, nil)
	c.inferTypes(inference.inferences, value, target, InferencePriorityNone, false)
	values := c.getInferredTypes(inference)
	if !c.isTypeAssignableTo(value, c.instantiateType(target, newTypeMapper(parameters, values))) {
		return false
	}
	for i, n := range names {
		f.bind(n, values[i])
	}
	return true
}

func (c *Checker) TypeFunctionResultForDisplay(node *ast.Node) *Type {
	return c.forceTypeFunction(c.getTypeFromTypeNode(node))
}

func sameOpenTypeFunction(a, b *Type) bool {
	x := a.typeFunctionApplication()
	y := b.typeFunctionApplication()
	return x != nil && y != nil && x.declaration == y.declaration && x.expression == y.expression && x.builtin == y.builtin && slices.Equal(x.arguments, y.arguments)
}
