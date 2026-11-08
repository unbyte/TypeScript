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
)

type typefnInstruction struct {
	op           typefnOp
	node         *ast.Node
	expr         *typefnExpression
	slot, target int
}
type typefnExpression struct {
	node   *ast.Node
	slot   int
	symbol *ast.Symbol
	args   []*typefnExpression
	keys   []*typefnExpression
}
type typefnProgram struct {
	declaration *ast.Node
	code        []typefnInstruction
	slots       int
	valid       bool
}
type typefnBinding struct {
	slot    int
	mutable bool
	symbol  *ast.Symbol
}
type (
	typefnLoop     struct{ breaks, continues []int }
	typefnCompiler struct {
		c      *Checker
		p      *typefnProgram
		scopes []map[string]typefnBinding
		loops  []*typefnLoop
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
			return v, true
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
		if statement.Kind == ast.KindTypeFunctionBinding {
			name := statement.Name().Text()
			if _, exists := scope[name]; exists {
				b.error(statement.Name(), "duplicate binding "+name)
			}
			scope[name] = typefnBinding{slot: -1, mutable: true}
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
	if node.Kind == ast.KindTypeFunctionDeclaration {
		b.error(node, "local typefns are not implemented yet")
		return
	}
	s := node.AsTypeFunctionStatement()
	switch node.Kind {
	case ast.KindTypeFunctionBinding:
		expr := b.expression(s.Expression)
		slot := b.bind(node.Name(), true)
		b.emit(typefnInstruction{op: tfStore, node: node, expr: expr, slot: slot})
	case ast.KindTypeFunctionAssignment:
		value, ok := b.lookup(node.Name().Text())
		if !ok || !value.mutable || value.slot < 0 {
			b.error(node.Name(), "assignment requires a mutable local binding")
		}
		b.emit(typefnInstruction{op: tfStore, node: node, expr: b.expression(s.Expression), slot: value.slot})
	case ast.KindTypeFunctionReturn:
		b.emit(typefnInstruction{op: tfReturn, node: node, expr: b.expression(s.Expression)})
	case ast.KindTypeFunctionExpressionStatement:
		b.emit(typefnInstruction{op: tfDiscard, node: node, expr: b.expression(s.Expression)})
	case ast.KindTypeFunctionIf:
		branch := b.emit(typefnInstruction{op: tfBranch, node: node, expr: b.expression(s.Expression)})
		b.block(s.Body, true)
		jump := b.emit(typefnInstruction{op: tfJump, node: node})
		b.p.code[branch].target = len(b.p.code)
		if s.ElseBody != nil {
			b.statement(s.ElseBody)
		}
		b.p.code[jump].target = len(b.p.code)
	case ast.KindTypeFunctionWhile, ast.KindTypeFunctionForOf, ast.KindTypeFunctionForIn:
		expr := b.expression(s.Expression)
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
			b.emit(typefnInstruction{op: tfIterStart, node: node, expr: expr, slot: slot})
			next = b.emit(typefnInstruction{op: tfIterNext, node: node, slot: slot})
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
	data := node.AsTypeFunctionExpression()
	switch node.Kind {
	case ast.KindTypeFunctionName, ast.KindTypeFunctionCall:
		name := node.Name()
		if ast.IsIdentifier(name) {
			if binding, ok := b.lookup(name.Text()); ok {
				if binding.slot < 0 {
					b.error(node, "binding read before initialization")
				}
				if node.Kind == ast.KindTypeFunctionCall {
					b.error(node, "constructor-valued locals are not implemented yet")
				}
				e.slot = binding.slot
				e.symbol = binding.symbol
				return e
			}
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
			if actual < minimum || actual > count || declaration != nil && node.Kind != ast.KindTypeFunctionCall {
				b.error(node, fmt.Sprintf("application requires %d..%d Type arguments", minimum, count))
			}
		}
	case ast.KindTypeFunctionNativeType:
		b.error(node, "native islands are not implemented yet")
	case ast.KindTypeFunctionObject:
		for _, field := range data.Operands.Nodes {
			name := field.Name()
			var key *typefnExpression
			if !ast.IsIdentifier(name) && !ast.IsStringLiteral(name) && !ast.IsNumericLiteral(name) {
				key = b.expression(name)
			}
			e.keys = append(e.keys, key)
			e.args = append(e.args, b.expression(field.AsTypeFunctionExpression().Operands.Nodes[0]))
		}
		return e
	}
	if data.Operands != nil {
		for _, operand := range data.Operands.Nodes {
			e.args = append(e.args, b.expression(operand))
		}
	}
	return e
}

func (c *Checker) compileTypefn(node *ast.Node) *typefnProgram {
	state := c.getTypefnState()
	if state.programs == nil {
		state.programs = make(map[*ast.Node]*typefnProgram)
	}
	if p := state.programs[node]; p != nil {
		return p
	}
	p := &typefnProgram{declaration: node, valid: true}
	state.programs[node] = p
	// Semantic diagnostics are also requested on incomplete editor snapshots.
	// Recovery nodes are for syntax services; never lower them to executable code.
	if len(ast.GetSourceFileOfNode(node).Diagnostics()) != 0 {
		p.valid = false
		return p
	}
	b := typefnCompiler{c: c, p: p, scopes: []map[string]typefnBinding{{}}}
	if typefnReserved(node.Name().Text()) {
		b.error(node.Name(), "reserved function name "+node.Name().Text())
	}
	for _, parameter := range node.TypeParameters() {
		b.bind(parameter.Name(), false)
		if parameter.AsTypeParameterDeclaration().Constraint != nil {
			c.checkSourceElement(parameter.AsTypeParameterDeclaration().Constraint)
		}
	}
	b.block(node.AsTypeFunctionDeclaration().Body, false)
	return p
}
