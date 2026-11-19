package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/internal/typefunction"
	"slices"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/diagnostics"
)

// The result of an open application is an opaque type variable. Its arguments
// participate in substitution, but never in inference from its result.
type typeFunctionApplication struct {
	use         *ast.Node
	declaration *ast.Node
	builtin     string
	arguments   []*Type
}

func (t *Type) typeFunctionApplication() *typeFunctionApplication {
	if t != nil && t.flags&TypeFlagsTypeParameter != 0 {
		return t.AsTypeParameter().application
	}
	return nil
}

func (c *Checker) newTypeFunctionApplication(use, declaration *ast.Node, builtin string, arguments []*Type) *Type {
	var symbol *ast.Symbol
	if declaration != nil {
		symbol = c.getSymbolOfDeclaration(declaration)
	}
	t := c.newTypeParameter(symbol)
	tp := t.AsTypeParameter()
	tp.constraint = c.noConstraintType
	tp.application = &typeFunctionApplication{use, declaration, builtin, arguments}
	if symbol != nil {
		t.alias = &TypeAlias{symbol: symbol, typeArguments: arguments}
	}
	return t
}

func (c *Checker) typeFunctionReference(node *ast.Node, symbol *ast.Symbol, decl *ast.Node) *Type {
	if node.TypeArgumentList() == nil || len(node.TypeArguments()) != len(decl.TypeParameters()) {
		c.error(node, diagnostics.Type_function_Colon_0, fmt.Sprintf("%s requires %d explicit type arguments", symbol.Name(), len(decl.TypeParameters())))
		return c.errorType
	}
	arguments := make([]*Type, len(node.TypeArguments()))
	for i, a := range node.TypeArguments() {
		if arity := typeFunctionSlots(decl.TypeParameters()[i]); arity != 0 {
			arguments[i] = c.readTypeFunctionConstructor(a, arity)
		} else {
			arguments[i] = c.getTypeFromTypeNode(a)
		}
	}
	return c.newTypeFunctionApplication(node, decl, "", arguments)
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
	if t == nil || bound[t] || path[t] {
		return true
	}
	if t.flags&(TypeFlagsStructuredOrInstantiable) == 0 {
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
	switch {
	case t.typeFunctionApplication() != nil:
		return all(t.typeFunctionApplication().arguments)
	case t.flags&TypeFlagsTypeParameter != 0:
		if ctor := t.AsTypeParameter().constructor; ctor != nil {
			return all(ctor.arguments)
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
		return closed(a.checkType) && closed(a.extendsType)
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
	trace  []string
	refs   map[*Type]*Type
	site   *ast.Node
	stack  []*typeFunctionApplication
	steps  int
	failed bool
}

func (c *Checker) failTypeFunction(node *ast.Node, message string) *Type {
	root := c.typeFunctionExecution
	if root != nil && !root.failed {
		root.failed = true
		d := c.error(root.site, diagnostics.Type_function_Colon_0, message)
		if node != nil && node != root.site && d != nil {
			d.AddRelatedInfo(createDiagnosticForNode(node, diagnostics.Type_function_Colon_0, message))
		}
	}
	return c.anyType
}

func (c *Checker) typeFunctionStep(node *ast.Node) bool {
	root := c.typeFunctionExecution
	root.steps++
	if root.steps > 100000 {
		c.failTypeFunction(node, "work limit exceeded")
	}
	if c.ctx != nil && c.ctx.Err() != nil {
		root.failed = true
	}
	return !root.failed
}

func (c *Checker) forceTypeFunction(t *Type) *Type {
	app := t.typeFunctionApplication()
	if app == nil {
		return t
	}
	for _, arg := range app.arguments {
		if !(app.builtin != "" && c.typeFunctionExecution != nil) && !c.typeFunctionClosed(arg, map[*Type]bool{}, map[*Type]bool{}) {
			return t
		}
	}
	if c.typeFunctionExecution == nil {
		c.typeFunctionExecution = &typeFunctionExecution{site: app.use}
		defer func() {
			if traceHost, ok := c.program.(interface{ TraceTypeFunction(string) }); ok && (c.ctx == nil || c.ctx.Err() == nil) {
				for _, event := range c.typeFunctionExecution.trace {
					traceHost.TraceTypeFunction(event)
				}
			}
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
		if active.declaration == app.declaration && active.builtin == app.builtin && slices.Equal(active.arguments, app.arguments) {
			return c.failTypeFunction(app.use, "recursive call with unchanged arguments")
		}
	}
	root.stack = append(root.stack, app)
	defer func() { root.stack = root.stack[:len(root.stack)-1] }()
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
	value, flow := frame.block(app.declaration.Type())
	if root.failed {
		return c.anyType
	}
	if flow != typeFunctionReturned {
		return c.failTypeFunction(app.declaration, "execution reached the end without a return")
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
	case ast.KindTypeFunctionBinary:
		b := node.AsTypeFunctionBinary()
		left := f.operand(b.Left)
		switch b.OperatorToken.Kind {
		case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken:
			if left != c.regularTrueType && left != c.regularFalseType {
				return c.failTypeFunction(b.Left, "expected a boolean literal")
			}
			if (left == c.regularTrueType) == (b.OperatorToken.Kind == ast.KindBarBarToken) {
				return left
			}
			right := f.operand(b.Right)
			if right != c.regularTrueType && right != c.regularFalseType {
				return c.failTypeFunction(b.Right, "expected a boolean literal")
			}
			return right
		default:
			right := f.operand(b.Right)
			const literals = TypeFlagsStringLiteral | TypeFlagsNumberLiteral | TypeFlagsBigIntLiteral | TypeFlagsBooleanLiteral | TypeFlagsNull | TypeFlagsUndefined
			if left.flags&literals == 0 || right.flags&literals == 0 {
				return c.failTypeFunction(node, "literal equality requires two primitive literals")
			}
			return f.boolean(c.isTypeIdenticalTo(left, right) == (b.OperatorToken.Kind == ast.KindEqualsEqualsEqualsToken))
		}
	}
	sources := slices.Clone(f.sources)
	values := slices.Clone(f.values)
	node.ForEachChild(func(n *ast.Node) bool { return f.forceNestedOperand(n, &sources, &values) })
	if c.typeFunctionExecution.failed {
		return c.anyType
	}
	t := c.getTypeFromTypeNode(node)
	if len(sources) > 0 {
		t = c.instantiateType(t, newTypeMapper(sources, values))
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
		cond := node.AsConditionalTypeNode()
		v := f.operand(cond.CheckType)
		target := f.operand(cond.ExtendsType)
		if c.typeFunctionClosed(v, map[*Type]bool{}, map[*Type]bool{}) && c.typeFunctionClosed(target, map[*Type]bool{}, map[*Type]bool{}) {
			branch := cond.FalseType
			if c.isTypeAssignableTo(v, target) {
				branch = cond.TrueType
			}
			return f.forceNestedOperand(branch, sources, values)
		}
		return false
	case ast.KindTypeReference:
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
		return f.operand(node.AsTypeFunctionOperandStatement().Operand), typeFunctionReturned
	case ast.KindTypeFunctionBinding:
		b := node.AsTypeFunctionBinding()
		value := f.operand(b.Initializer)
		if b.Destructured {
			if !value.IsTupleType() || len(c.getTypeArguments(value)) != len(b.Bindings.Nodes) {
				c.failTypeFunction(node, "tuple does not match the binding list")
				break
			}
			for i, name := range b.Bindings.Nodes {
				f.bind(name, c.getTypeArguments(value)[i])
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
		operand := node.AsTypeFunctionOperandStatement().Operand
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
					c.failTypeFunction(node, message.AsLiteralType().value.(string))
				}
				return nil, typeFunctionNext
			case "log":
				f.log(operand)
				return nil, typeFunctionNext
			}
		}
		f.operand(operand)
	case ast.KindTypeFunctionIf, ast.KindTypeFunctionWhile:
		b := node.AsTypeFunctionBranch()
		for c.typeFunctionStep(node) {
			v := f.operand(b.Condition)
			if v == c.regularFalseType {
				if b.ElseBranch != nil {
					if ast.IsTypeFunctionBlock(b.ElseBranch) {
						return f.block(b.ElseBranch)
					}
					return f.statement(b.ElseBranch)
				}
				break
			}
			if v != c.regularTrueType {
				c.failTypeFunction(node, "expected a boolean literal")
				break
			}
			value, flow := f.block(b.Block)
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
		}
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
	return c.newTypeFunctionApplication(app.use, app.declaration, app.builtin, args)
}

func (f *typeFunctionFrame) match(value *Type, pattern *ast.Node) bool {
	c := f.c
	target := f.operand(pattern)
	var names []*ast.Node
	var parameters []*Type
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindInferType {
			p := n.AsInferTypeNode().TypeParameter
			names = append(names, p)
			parameters = append(parameters, c.getDeclaredTypeOfSymbol(c.getSymbolOfDeclaration(p)))
			return false
		}
		if n.Kind == ast.KindConditionalType {
			return false
		}
		return n.ForEachChild(visit)
	}
	visit(pattern)
	if len(names) == 0 {
		return c.isTypeAssignableTo(value, target)
	}
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
	return x != nil && y != nil && x.declaration == y.declaration && x.builtin == y.builtin && slices.Equal(x.arguments, y.arguments)
}
