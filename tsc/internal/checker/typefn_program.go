package checker

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/diagnostics"
)

// The initial execution profile lowers control flow to jumps and resolves lexical
// bindings before execution. Registers contain native checker Type identities,
// never JavaScript values, synthetic type parameters, or strings of TypeScript.
type typefnOp uint8

const (
	tfStore typefnOp = iota
	tfReturn
	tfBranch
	tfJump
	tfIterStart
	tfIterNext
	tfDiscard
	tfMatch
	tfHelper
)

type typefnInstruction struct {
	op           typefnOp
	node         *ast.Node
	expr         *typefnExpression
	pattern      *typefnPattern
	slot, target int
	iterator     int
	helper       *typefnProgram
}
type typefnExpression struct {
	node      *ast.Node
	slot      int
	symbol    *ast.Symbol
	intrinsic string
	island    *typefnIsland
	helper    *typefnProgram // Code for a constructor-valued expression.
	kind      *typefnKind
	callee    *typefnProgram
	args      []*typefnExpression
	keys      []*typefnExpression
}
type typefnCapture struct{ source, target int }

type typefnProgram struct {
	declaration            *ast.Node
	code                   []typefnInstruction
	slots                  int
	iterators              int
	valid                  bool
	selfSlot               int
	local                  bool
	captures               []typefnCapture
	bounds                 []*typefnExpression
	parameterKinds         []*typefnKind
	resultKind             *typefnKind
	hasValueReturn         bool
	kindIndex, kindLowlink int
	kindOnStack            bool
	bindings               map[*ast.Symbol]*typefnKind
}
type typefnBinding struct {
	slot             int
	mutable          bool
	symbol           *ast.Symbol
	helper           *typefnProgram // Named function signature, not a constructor binding.
	kind             *typefnKind
	unavailableBound bool
}
type (
	typefnLoop     struct{ breaks, continues []int }
	typefnCompiler struct {
		c              *Checker
		p              *typefnProgram
		scopes         []map[string]typefnBinding
		loops          []*typefnLoop
		parent         *typefnCompiler
		captures       map[string]typefnBinding
		boundParameter int
	}
)

func (c *Checker) typefnError(node *ast.Node, message string) {
	c.error(node, diagnostics.NewAdHocMessage("typefn: "+message))
}

func (b *typefnCompiler) error(node *ast.Node, message string) {
	b.p.valid = false
	b.c.typefnError(node, message)
}

func typefnReserved(name string) bool {
	switch name {
	case "text", "numeric", "tuple", "reflect", "types", "fail", "failAt", "_",
		"type", "typefn", "return", "if", "else", "while", "for", "of", "in",
		"match", "is", "break", "continue", "extends", "readonly", "keyof",
		"true", "false", "null", "undefined", "any", "unknown", "never",
		"string", "number", "bigint", "boolean", "object", "void", "symbol":
		return true
	}
	return false
}

func (b *typefnCompiler) bind(node *ast.Node, mutable bool) int {
	name := node.Text()
	scope := b.scopes[len(b.scopes)-1]
	if existing, exists := scope[name]; exists && existing.slot >= 0 || typefnReserved(name) {
		b.error(node, "duplicate or reserved binding "+name)
	}
	slot := b.p.slots
	b.p.slots++
	scope[name] = typefnBinding{slot: slot, mutable: mutable, symbol: node.Parent.Symbol()}
	return slot
}

func (b *typefnCompiler) lookup(name string) (typefnBinding, bool) {
	for i := len(b.scopes) - 1; i >= 0; i-- {
		if v, ok := b.scopes[i][name]; ok {
			if b.boundParameter >= 0 {
				for index, parameter := range b.p.declaration.TypeParameters() {
					if parameter.Symbol() == v.symbol && index >= b.boundParameter {
						v.slot = -1
						v.unavailableBound = true
						break
					}
				}
			}
			return v, true
		}
	}
	if binding, ok := b.captures[name]; ok {
		return binding, true
	}
	if b.parent != nil {
		if source, ok := b.parent.lookup(name); ok {
			if source.slot < 0 {
				return source, true
			}
			binding := source
			binding.mutable = false
			binding.slot = b.p.slots
			b.p.slots++
			b.p.captures = append(b.p.captures, typefnCapture{source: source.slot, target: binding.slot})
			if b.captures == nil {
				b.captures = map[string]typefnBinding{}
			}
			b.captures[name] = binding
			return binding, true
		}
	}
	return typefnBinding{}, false
}

func (b *typefnCompiler) emit(i typefnInstruction) int {
	pc := len(b.p.code)
	b.p.code = append(b.p.code, i)
	return pc
}

func (b *typefnCompiler) block(node *ast.Node, newScope bool) {
	if newScope {
		b.scopes = append(b.scopes, map[string]typefnBinding{})
		defer func() { b.scopes = b.scopes[:len(b.scopes)-1] }()
	}
	// Reserve the block's locals before resolving initializers, including dead code.
	scope := b.scopes[len(b.scopes)-1]
	for _, statement := range node.AsTypeFunctionBlock().Statements.Nodes {
		if statement.Kind == ast.KindTypeFunctionBinding || statement.Kind == ast.KindTypeFunctionDeclaration {
			name := statement.Name().Text()
			if _, exists := scope[name]; exists {
				b.error(statement.Name(), "duplicate binding "+name)
			}
			scope[name] = typefnBinding{slot: -1, mutable: true, symbol: statement.Symbol()}
		}
	}
	for _, s := range node.AsTypeFunctionBlock().Statements.Nodes {
		b.statement(s)
	}
}

func (b *typefnCompiler) statement(node *ast.Node) {
	if node.Kind == ast.KindTypeFunctionBlock {
		b.block(node, true)
		return
	}
	if node.Kind == ast.KindTypeFunctionMatch {
		b.match(node)
		return
	}
	if node.Kind == ast.KindTypeFunctionDeclaration {
		child := b.compileFunction(node, b)
		if !child.valid {
			b.p.valid = false
		}
		slot := b.bind(node.Name(), false)
		binding := b.scopes[len(b.scopes)-1][node.Name().Text()]
		binding.helper = child
		binding.kind = typefnConstructorKind(len(node.TypeParameters()))
		b.scopes[len(b.scopes)-1][node.Name().Text()] = binding
		b.emit(typefnInstruction{op: tfHelper, node: node, slot: slot, helper: child})
		return
	}
	s := node.AsTypeFunctionStatement()
	switch node.Kind {
	case ast.KindTypeFunctionBinding:
		expr := b.expression(s.Expression)
		slot := b.bind(node.Name(), true)
		binding := b.scopes[len(b.scopes)-1][node.Name().Text()]
		binding.kind = expr.kind
		b.p.bindings[binding.symbol] = expr.kind
		b.scopes[len(b.scopes)-1][node.Name().Text()] = binding
		b.emit(typefnInstruction{op: tfStore, node: node, expr: expr, slot: slot})
	case ast.KindTypeFunctionAssignment:
		value, ok := b.lookup(node.Name().Text())
		if !ok || !value.mutable || value.slot < 0 {
			b.error(node.Name(), "assignment requires a mutable local binding")
		}
		expr := b.expression(s.Expression)
		if ok && value.mutable && value.slot >= 0 {
			message := "assignment cannot change a binding's value kind"
			if value.kind.resolved() > 0 && expr.kind.resolved() > 0 {
				message = "assignment cannot change constructor arity"
			}
			b.sameKind(node, value.kind, expr.kind, message)
		}
		b.emit(typefnInstruction{op: tfStore, node: node, expr: expr, slot: value.slot})
	case ast.KindTypeFunctionReturn:
		expr := b.expression(s.Expression)
		b.sameKind(node, b.p.resultKind, expr.kind, "return kinds must agree with the function result kind")
		b.emit(typefnInstruction{op: tfReturn, node: node, expr: expr})
	case ast.KindTypeFunctionExpressionStatement:
		b.emit(typefnInstruction{op: tfDiscard, node: node, expr: b.expression(s.Expression)})
	case ast.KindTypeFunctionIf:
		branch := b.emit(typefnInstruction{op: tfBranch, node: node, expr: b.typeExpression(s.Expression)})
		b.block(s.Body, true)
		jump := b.emit(typefnInstruction{op: tfJump, node: node})
		b.p.code[branch].target = len(b.p.code)
		if s.ElseBody != nil {
			b.statement(s.ElseBody)
		}
		b.p.code[jump].target = len(b.p.code)
	case ast.KindTypeFunctionWhile, ast.KindTypeFunctionForOf, ast.KindTypeFunctionForIn:
		expr := b.typeExpression(s.Expression)
		loop := &typefnLoop{}
		b.loops = append(b.loops, loop)
		start := len(b.p.code)
		var exit, next int
		if node.Kind == ast.KindTypeFunctionWhile {
			exit = b.emit(typefnInstruction{op: tfBranch, node: node, expr: expr})
			next = start
			b.block(s.Body, true)
		} else {
			b.scopes = append(b.scopes, map[string]typefnBinding{})
			slot := b.bind(node.Name(), false)
			iterator := b.p.iterators
			b.p.iterators++
			b.emit(typefnInstruction{op: tfIterStart, node: node, expr: expr, slot: slot, iterator: iterator})
			next = b.emit(typefnInstruction{op: tfIterNext, node: node, slot: slot, iterator: iterator})
			exit = next
			b.block(s.Body, false)
			b.scopes = b.scopes[:len(b.scopes)-1]
		}
		b.emit(typefnInstruction{op: tfJump, node: node, target: next})
		end := len(b.p.code)
		b.p.code[exit].target = end
		for _, pc := range loop.breaks {
			b.p.code[pc].target = end
		}
		for _, pc := range loop.continues {
			b.p.code[pc].target = next
		}
		b.loops = b.loops[:len(b.loops)-1]
	case ast.KindTypeFunctionBreak, ast.KindTypeFunctionContinue:
		if len(b.loops) == 0 {
			b.error(node, "loop control outside a loop")
			return
		}
		loop := b.loops[len(b.loops)-1]
		pc := b.emit(typefnInstruction{op: tfJump, node: node})
		if node.Kind == ast.KindTypeFunctionBreak {
			loop.breaks = append(loop.breaks, pc)
		} else {
			loop.continues = append(loop.continues, pc)
		}
	default:
		b.error(node, "statement is not implemented yet")
	}
}

func (b *typefnCompiler) expression(node *ast.Node) *typefnExpression {
	e := &typefnExpression{node: node, slot: -1}
	if node.Kind == ast.KindTypeFunctionLambda {
		e.helper = b.compileFunction(node, b)
		e.kind = typefnConstructorKind(len(node.TypeParameters()))
		b.constructorEligible(node, e.helper)
		if !e.helper.valid {
			b.p.valid = false
		}
		return e
	}
	if node.Kind == ast.KindTypeFunctionTemplate {
		for _, span := range node.AsTypeFunctionTemplate().Spans.Nodes {
			e.args = append(e.args, b.typeExpression(span.AsTypeFunctionTemplateSpan().Expression))
		}
		return e
	}
	data := node.AsTypeFunctionExpression()
	switch node.Kind {
	case ast.KindTypeFunctionName, ast.KindTypeFunctionCall:
		name := node.Name()
		if ast.IsIdentifier(name) {
			if binding, ok := b.lookup(name.Text()); ok {
				if binding.slot < 0 {
					if binding.unavailableBound {
						b.error(node, "input bound may reference only preceding Type parameters")
					} else {
						b.error(node, "binding read before initialization")
					}
				}
				if node.Kind == ast.KindTypeFunctionCall {
					if binding.helper != nil {
						b.kindDependency(binding.helper, false)
						e.kind = binding.helper.resultKind
						e.callee = binding.helper
					}
					if binding.slot >= 0 && binding.kind.resolved() == 0 {
						b.error(node, "application requires a constructor value")
					} else if binding.slot >= 0 {
						b.sameKind(node, binding.kind, typefnConstructorKind(len(data.Operands.Nodes)), "wrong number of helper Type arguments")
					}
					for i, argument := range data.Operands.Nodes {
						e.args = append(e.args, b.callArgument(argument, binding.helper, i))
					}
				} else {
					e.kind = binding.kind
					if binding.helper != nil {
						b.kindDependency(binding.helper, false)
						b.constructorEligible(node, binding.helper)
					}
				}
				e.slot = binding.slot
				e.symbol = binding.symbol
				return e
			}
		}
		if b.intrinsic(e) {
			break
		}
		e.symbol = b.c.resolveEntityName(name, ast.SymbolFlagsType, false, false, nil)
		if e.symbol == nil || e.symbol == b.c.unknownSymbol {
			b.p.valid = false
		} else {
			declaration := core.Find(e.symbol.Declarations(), ast.IsTypeFunctionDeclaration)
			parameters := b.c.getLocalTypeParametersOfClassOrInterfaceOrTypeAlias(e.symbol)
			count := len(parameters)
			minimum := b.c.getMinTypeArgumentCount(parameters)
			if declaration != nil {
				count = len(declaration.TypeParameters())
				minimum = count
			}
			actual := 0
			if data.Operands != nil {
				actual = len(data.Operands.Nodes)
			}
			if declaration != nil {
				callee := b.compileFunction(declaration, nil)
				if !callee.valid {
					b.p.valid = false
				}
				if node.Kind == ast.KindTypeFunctionName {
					e.helper = callee
					e.kind = typefnConstructorKind(count)
					b.constructorEligible(node, callee)
				} else {
					e.callee = callee
					e.kind = callee.resultKind
				}
			}
			if (declaration == nil || node.Kind != ast.KindTypeFunctionName) && (actual < minimum || actual > count) {
				b.error(node, fmt.Sprintf("application requires %d..%d Type arguments", minimum, count))
			}
		}
	case ast.KindTypeFunctionNativeType:
		b.island(e)
		return e
	case ast.KindTypeFunctionObject:
		for _, field := range data.Operands.Nodes {
			name := field.Name()
			var key *typefnExpression
			if !ast.IsIdentifier(name) && !ast.IsStringLiteral(name) && !ast.IsNumericLiteral(name) {
				key = b.typeExpression(name)
			}
			e.keys = append(e.keys, key)
			e.args = append(e.args, b.typeExpression(field.AsTypeFunctionExpression().Operands.Nodes[0]))
		}
		return e
	}
	if data.Operands != nil {
		for i, operand := range data.Operands.Nodes {
			e.args = append(e.args, b.callArgument(operand, e.callee, i))
		}
	}
	return e
}

func (c *Checker) compileTypefn(node *ast.Node) *typefnProgram {
	if c.typefn != nil && c.typefn.programs[node] != nil {
		return c.typefn.programs[node]
	}
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if ast.IsTypeFunctionLike(parent) {
			c.compileTypefn(parent)
			if p := c.getTypefnState().programs[node]; p != nil {
				return p
			}
			break
		}
	}
	return c.compileTypefnWithParent(node, nil)
}

func (c *Checker) compileTypefnWithParent(node *ast.Node, parent *typefnCompiler) *typefnProgram {
	state := c.getTypefnState()
	if state.programs == nil {
		state.programs = make(map[*ast.Node]*typefnProgram)
	}
	if p := state.programs[node]; p != nil {
		return p
	}
	p := &typefnProgram{declaration: node, valid: true, selfSlot: -1, local: parent != nil, resultKind: &typefnKind{value: -1}, bindings: map[*ast.Symbol]*typefnKind{}}
	state.programs[node] = p
	state.nextKindIndex++
	p.kindIndex, p.kindLowlink = state.nextKindIndex, state.nextKindIndex
	p.kindOnStack = true
	state.kindStack = append(state.kindStack, p)
	defer func() {
		if p.kindLowlink != p.kindIndex {
			return
		}
		index := len(state.kindStack) - 1
		for state.kindStack[index] != p {
			index--
		}
		component := state.kindStack[index:]
		for _, member := range component {
			member.kindOnStack = false
		}
		c.finishTypefnKinds(component)
		state.kindStack = state.kindStack[:index]
	}()
	// Semantic diagnostics are also requested on incomplete editor snapshots.
	// Recovery nodes are for syntax services; never lower them to executable code.
	if len(ast.GetSourceFileOfNode(node).Diagnostics()) != 0 {
		p.valid = false
		return p
	}
	if ast.IsTypeFunctionLambda(node) {
		p.resultKind.value = 0
	}
	b := typefnCompiler{c: c, p: p, scopes: []map[string]typefnBinding{{}}, parent: parent, boundParameter: -1}
	if annotation := ast.TypeFunctionResultKind(node); annotation != nil {
		kind := annotation.AsTypeFunctionKind()
		p.resultKind.value = 0
		if kind.Slots != nil {
			p.resultKind.value = len(kind.Slots.Nodes) + 1
		}
	}
	if node.Name() != nil && typefnReserved(node.Name().Text()) {
		b.error(node.Name(), "reserved function name "+node.Name().Text())
	}
	p.parameterKinds = make([]*typefnKind, len(node.TypeParameters()))
	for i, parameter := range node.TypeParameters() {
		b.bind(parameter.Name(), false)
		if slots := parameter.AsTypeFunctionParameter().Slots; slots != nil {
			binding := b.scopes[len(b.scopes)-1][parameter.Name().Text()]
			binding.kind = typefnConstructorKind(len(slots.Nodes))
			b.scopes[len(b.scopes)-1][parameter.Name().Text()] = binding
			p.bindings[parameter.Symbol()] = binding.kind
			p.parameterKinds[i] = binding.kind
		}
	}
	if parent != nil {
		if node.Name() != nil {
			// Parameters/body may shadow the helper's self-name; captures cannot mutate it.
			p.selfSlot = p.slots
			p.slots++
			self := map[string]typefnBinding{node.Name().Text(): {slot: p.selfSlot, symbol: node.Symbol(), helper: p, kind: typefnConstructorKind(len(node.TypeParameters()))}}
			b.scopes = append([]map[string]typefnBinding{self}, b.scopes...)
		}
	}
	p.bounds = make([]*typefnExpression, len(node.TypeParameters()))
	for i, parameter := range node.TypeParameters() {
		if constraint := parameter.AsTypeFunctionParameter().Constraint; constraint != nil {
			b.boundParameter = i
			p.bounds[i] = b.typeExpression(constraint)
			b.boundParameter = -1
		}
	}

	b.block(ast.TypeFunctionBody(node), false)
	flow := p.flow()
	p.hasValueReturn = flow.returns
	if p.valid && flow.fallsThrough {
		b.error(node, "a control-flow path reaches the end without returning a value")
	}
	return p
}

func (b *typefnCompiler) typeExpression(node *ast.Node) *typefnExpression {
	expression := b.expression(node)
	b.sameKind(node, expression.kind, nil, "constructor value cannot be used as a Type")
	return expression
}
