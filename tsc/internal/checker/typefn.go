package checker

import (
	"fmt"
	"math"
	"math/big"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/diagnostics"
	"github.com/microsoft/TypeScript/tsc/internal/jsnum"
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
	if c.typefnPrograms == nil {
		c.typefnPrograms = make(map[*ast.Node]*typefnProgram)
	}
	if p := c.typefnPrograms[node]; p != nil {
		return p
	}
	p := &typefnProgram{declaration: node, valid: true}
	c.typefnPrograms[node] = p
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

// A root owns the budget; nested DSL/native calls cannot reset it. Failed roots
// are not cached. Programs and all Type handles belong to this checker snapshot.
type typefnActiveCall struct {
	declaration *ast.Node
	args        []*Type
}

type typefnVM struct {
	active              []typefnActiveCall
	c                   *Checker
	limits              core.TypefnLimits
	steps, depth, nodes int
	liveBytes           int
	failed              bool
	blocked             bool
	checks              []*Type
	checkSet            map[*Type]bool
	replaySite          *ast.Node
}

func (c *Checker) evaluateTypefn(node, declaration *ast.Node, args []*Type) *Type {
	return c.runTypefn(node, func(v *typefnVM) *Type { return v.call(node, declaration, args) })
}

func (c *Checker) runTypefn(node *ast.Node, run func(*typefnVM) *Type) *Type {
	root := c.typefnVM == nil
	if root {
		limits := core.TypefnLimits{VMInstructions: 1000000, CallDepth: 256, NativeNodes: 250000, BigintBits: 1048576, LiveBytes: 64 << 20}
		if c.compilerOptions.Typefn != nil {
			configured := c.compilerOptions.Typefn.Limits
			if configured.NativeWork != 0 || configured.StringUnits != 0 || configured.CacheBytes != 0 {
				c.typefnError(node, "requested nativeWork, stringUnits or cacheBytes limit is not implemented yet")
				return c.errorType
			}
			if configured.BigintBits != 0 {
				limits.BigintBits = configured.BigintBits
			}
			if configured.LiveBytes != 0 {
				limits.LiveBytes = configured.LiveBytes
			}
			if configured.VMInstructions != 0 {
				limits.VMInstructions = configured.VMInstructions
			}
			if configured.CallDepth != 0 {
				limits.CallDepth = configured.CallDepth
			}
			if configured.NativeNodes != 0 {
				limits.NativeNodes = configured.NativeNodes
			}
		}
		c.typefnVM = &typefnVM{c: c, limits: limits}
		defer func() { c.typefnVM = nil }()
	}
	v := c.typefnVM
	saved, savedSet := v.checks, v.checkSet
	v.checks, v.checkSet = nil, nil
	defer func() { v.checks, v.checkSet = saved, savedSet }()
	result := run(v)
	if v.failed {
		return c.errorType
	}
	return c.guardTypefn(result, v.checks)
}

func (v *typefnVM) error(node *ast.Node, message string) *Type {
	if !v.failed {
		site := node
		if v.replaySite != nil {
			site = v.replaySite
		}
		if site == nil {
			site = v.c.currentNode
		}
		v.c.typefnError(site, message)
	}
	v.failed = true
	return v.c.errorType
}

func (v *typefnVM) step(node *ast.Node) bool {
	if v.failed {
		return false
	}
	if v.c.isCanceled() {
		v.failed = true
		return false
	}
	v.steps++
	if v.steps > v.limits.VMInstructions {
		v.error(node, "VM instruction limit exceeded")
		return false
	}
	return true
}

func (v *typefnVM) allocate(node *ast.Node, count int) bool {
	if count > v.limits.NativeNodes-v.nodes {
		v.error(node, "native construction limit exceeded")
		return false
	}
	v.nodes += count
	return true
}

func (v *typefnVM) call(node, declaration *ast.Node, args []*Type) (result *Type) {
	originalArgs := args
	savedChecks, savedSet, savedBlocked := v.checks, v.checkSet, v.blocked
	v.checks, v.checkSet, v.blocked = nil, nil, false
	defer func() {
		if v.blocked && !v.failed {
			result = v.c.newTypefnTerm(&typefnTerm{kind: typefnCallTerm, node: node, declaration: declaration, args: originalArgs})
		} else {
			result = v.c.guardTypefn(result, v.checks)
		}
		v.checks, v.checkSet, v.blocked = savedChecks, savedSet, savedBlocked
		v.observe(result)
	}()
	args = slices.Clone(args)
	for i, arg := range args {
		args[i] = v.observe(arg)
	}
	// Recovery Types are not successful Type values, even when a parameter is unused.
	for _, arg := range args {
		if v.c.isErrorType(arg) {
			v.failed = true
			return v.c.errorType
		}
	}
	if len(ast.GetSourceFileOfNode(declaration).Diagnostics()) != 0 {
		v.failed = true
		return v.c.errorType
	}
	if v.depth >= v.limits.CallDepth {
		return v.error(node, "call depth limit exceeded")
	}
	for _, active := range v.active {
		if active.declaration != declaration || len(active.args) != len(args) {
			continue
		}
		same := true
		for i, arg := range args {
			if active.args[i] != arg {
				same = false
				break
			}
		}
		if same {
			return v.error(node, "recursive evaluation cycle")
		}
	}
	v.active = append(v.active, typefnActiveCall{declaration, args})
	v.depth++
	defer func() { v.depth--; v.active = v.active[:len(v.active)-1] }()
	p := v.c.compileTypefn(declaration)
	if !p.valid {
		v.failed = true
		return v.c.errorType
	}
	if len(args) != len(declaration.TypeParameters()) {
		return v.error(node, "wrong number of Type arguments")
	}
	parameters := v.c.appendTypeParameters(nil, declaration.TypeParameters())
	if !v.constraints(node, parameters, args) {
		return v.c.errorType
	}
	frameBytes := p.slots*8 + 128
	if frameBytes > v.limits.LiveBytes-v.liveBytes {
		return v.error(node, "live VM allocation limit exceeded")
	}
	v.liveBytes += frameBytes
	defer func() { v.liveBytes -= frameBytes }()
	registers := make([]*Type, p.slots)
	copy(registers, args)
	type iterator struct {
		values []*Type
		index  int
	}
	iterators := make(map[int]*iterator)
	for pc := 0; pc < len(p.code); {
		i := &p.code[pc]
		pc++
		if !v.step(i.node) {
			return v.c.errorType
		}
		var value *Type
		if i.expr != nil {
			value = v.expression(i.expr, registers)
			if v.failed || v.blocked {
				return v.c.errorType
			}
		}
		switch i.op {
		case tfStore:
			registers[i.slot] = value
		case tfReturn:
			return value
		case tfBranch:
			truth, ok := v.boolean(i.node, value)
			if !ok {
				return v.c.errorType
			}
			if !truth {
				pc = i.target
			}
		case tfJump:
			pc = i.target
		case tfIterStart:
			values := v.elements(i.node, value, i.node.Kind == ast.KindTypeFunctionForIn)
			if v.failed || v.blocked {
				return v.c.errorType
			}
			iterators[i.slot] = &iterator{values: values}
		case tfIterNext:
			iter := iterators[i.slot]
			if iter.index == len(iter.values) {
				delete(iterators, i.slot)
				pc = i.target
			} else {
				registers[i.slot] = iter.values[iter.index]
				iter.index++
			}
		}
	}
	return v.error(node, "execution reached the end without returning a Type")
}

func (v *typefnVM) constraints(node *ast.Node, parameters, args []*Type) bool {
	mapper := newTypeMapper(parameters, args)
	for i, parameter := range parameters {
		if constraint := v.c.getConstraintOfTypeParameter(parameter); constraint != nil {
			if v.c.isErrorType(constraint) {
				v.failed = true
				return false
			}
			bound := v.c.instantiateType(constraint, mapper)
			if !v.c.isTypeAssignableTo(args[i], bound) && (typefnUnresolved(args[i]) || typefnUnresolved(bound)) {
				v.blocked = true
				return false
			}
			site := node
			if v.replaySite != nil {
				site = v.replaySite
			}
			if !v.c.checkTypeAssignableTo(args[i], bound, site, diagnostics.Type_0_does_not_satisfy_the_constraint_1) {
				v.failed = true
				return false
			}
		}
	}
	return true
}

func (v *typefnVM) apply(node *ast.Node, symbol *ast.Symbol, args []*Type) *Type {
	args = slices.Clone(args)
	for i, arg := range args {
		args[i] = v.observe(arg)
	}
	if declaration := core.Find(symbol.Declarations(), ast.IsTypeFunctionDeclaration); declaration != nil {
		return v.call(node, declaration, args)
	}
	t := v.c.getDeclaredTypeOfSymbol(symbol)
	parameters := v.c.getLocalTypeParametersOfClassOrInterfaceOrTypeAlias(symbol)
	args = v.c.fillMissingTypeArguments(args, parameters, v.c.getMinTypeArgumentCount(parameters), false)
	for i, arg := range args {
		args[i] = v.observe(arg)
	}
	if !v.constraints(node, parameters, args) {
		if v.blocked && !v.failed {
			v.blocked = false
			return v.c.newTypefnTerm(&typefnTerm{kind: typefnNativeCallTerm, node: node, symbol: symbol, args: args})
		}
		return v.c.errorType
	}
	if len(parameters) == 0 {
		return t
	}
	if symbol.Flags()&ast.SymbolFlagsTypeAlias != 0 {
		return v.c.getTypeAliasInstantiation(symbol, args, nil)
	}
	if symbol.Flags()&(ast.SymbolFlagsClass|ast.SymbolFlagsInterface) != 0 {
		all := append(append([]*Type(nil), t.AsInterfaceType().OuterTypeParameters()...), args...)
		return v.c.createTypeReference(t, all)
	}
	return v.error(node, "unsupported native constructor")
}

func (v *typefnVM) expression(e *typefnExpression, slots []*Type) *Type {
	value := v.expressionWorker(e, slots)
	// Preserve native operation failure even if this value would be discarded,
	// absorbed by a union, or projected away by the surrounding DSL expression.
	if v.c.isErrorType(value) && !v.blocked {
		v.failed = true
	}
	return v.observe(value)
}

func (v *typefnVM) expressionWorker(e *typefnExpression, slots []*Type) *Type {
	node := e.node
	if !v.step(node) {
		return v.c.errorType
	}
	if e.slot >= 0 {
		return slots[e.slot]
	}
	data := node.AsTypeFunctionExpression()
	switch node.Kind {
	case ast.KindTypeFunctionLiteral:
		value := v.c.getTypeFromTypeNode(data.NativeType)
		if value.flags == TypeFlagsNumberLiteral {
			n := float64(value.AsLiteralType().value.(jsnum.Number))
			if math.IsInf(n, 0) || math.IsNaN(n) {
				return v.error(node, "number literal must be finite")
			}
		}
		if value.flags == TypeFlagsBigIntLiteral {
			n, _ := new(big.Int).SetString(value.AsLiteralType().value.(jsnum.PseudoBigInt).String(), 10)
			if n.BitLen() > v.limits.BigintBits {
				return v.error(node, "bigint magnitude limit exceeded")
			}
		}
		return value
	case ast.KindTypeFunctionName:
		return v.apply(node, e.symbol, nil)
	case ast.KindTypeFunctionBinary:
		left := v.expression(e.args[0], slots)
		if v.failed || v.blocked {
			return v.c.errorType
		}
		op := data.OperatorToken.Kind
		if op == ast.KindAmpersandAmpersandToken || op == ast.KindBarBarToken {
			truth, ok := v.boolean(node, left)
			if !ok {
				return v.c.errorType
			}
			if truth == (op == ast.KindBarBarToken) {
				return left
			}
			right := v.expression(e.args[1], slots)
			if v.failed || v.blocked {
				return v.c.errorType
			}
			if _, ok := v.boolean(node, right); !ok {
				return v.c.errorType
			}
			return right
		}
		right := v.expression(e.args[1], slots)
		if v.failed || v.blocked {
			return v.c.errorType
		}
		return v.binary(node, op, left, right)
	case ast.KindTypeFunctionUnary:
		value := v.expression(e.args[0], slots)
		if v.failed || v.blocked {
			return v.c.errorType
		}
		return v.unary(node, data.OperatorToken.Kind, value)
	case ast.KindTypeFunctionTuple:
		var values []*Type
		for _, element := range e.args {
			item := element
			spread := element.node.Kind == ast.KindTypeFunctionSpread
			if spread {
				item = element.args[0]
			}
			value := v.expression(item, slots)
			if v.failed || v.blocked {
				return v.c.errorType
			}
			if spread {
				elements := v.elements(node, value, false)
				if v.blocked && !v.failed {
					v.blocked = false
					return v.deferExpression(e, slots)
				}
				if v.failed {
					return v.c.errorType
				}
				values = append(values, elements...)
			} else {
				values = append(values, value)
			}
			if !v.allocate(node, 1) {
				return v.c.errorType
			}
		}
		if !v.allocate(node, len(values)+1) {
			return v.c.errorType
		}
		return v.c.createTupleType(values)
	case ast.KindTypeFunctionObject:
		members := ast.NewSymbolTable()
		for index, field := range data.Operands.Nodes {
			fieldData := field.AsTypeFunctionExpression()
			name := ""
			if e.keys[index] == nil {
				// Text is the scanner's decoded/canonical value, not source spelling.
				name = field.Name().Text()
				if ast.IsNumericLiteral(field.Name()) {
					number := jsnum.FromString(name)
					if math.IsInf(float64(number), 0) || math.IsNaN(float64(number)) {
						return v.error(field, "number property key must be finite")
					}
					name = number.String()
				}
			}
			if e.keys[index] != nil {
				key := v.expression(e.keys[index], slots)
				if v.failed {
					return v.c.errorType
				}
				if typefnUnresolved(key) {
					return v.deferExpression(e, slots)
				}
				if key.flags == TypeFlagsStringLiteral {
					name = key.AsLiteralType().value.(string)
				} else if key.flags == TypeFlagsNumberLiteral {
					number := key.AsLiteralType().value.(jsnum.Number)
					if math.IsInf(float64(number), 0) || math.IsNaN(float64(number)) {
						return v.error(field, "number property key must be finite")
					}
					name = number.String()
				} else {
					return v.error(field, "computed key requires one string or number literal; unique symbols are not implemented yet")
				}
			}
			if members.Get(name) != nil {
				return v.error(field, "duplicate object field "+name)
			}
			value := v.expression(e.args[index], slots)
			if v.failed || v.blocked {
				return v.c.errorType
			}
			if !v.allocate(field, 1) {
				return v.c.errorType
			}
			flags := ast.SymbolFlagsProperty
			if fieldData.Optional {
				flags |= ast.SymbolFlagsOptional
			}
			checkFlags := ast.CheckFlagsNone
			if fieldData.Readonly {
				checkFlags |= ast.CheckFlagsReadonly
			}
			symbol := v.c.newSymbolEx(flags, name, checkFlags)
			v.c.valueSymbolLinks.Get(symbol).resolvedType = v.c.addOptionalityEx(value, true, fieldData.Optional)
			members.Set(name, symbol)
		}
		if !v.allocate(node, 1) {
			return v.c.errorType
		}
		return v.c.newTypefnObject(members)
	case ast.KindTypeFunctionCall, ast.KindTypeFunctionIndex:
		args := make([]*Type, len(e.args))
		for i, arg := range e.args {
			args[i] = v.expression(arg, slots)
			if v.failed || v.blocked {
				return v.c.errorType
			}
		}
		if node.Kind == ast.KindTypeFunctionIndex {
			return v.c.getIndexedAccessTypeEx(args[0], args[1], AccessFlagsNone, node, nil)
		}
		return v.apply(node, e.symbol, args)
	}
	return v.error(node, "expression is not implemented yet")
}

func (v *typefnVM) boolType(value bool) *Type {
	if value {
		return v.c.regularTrueType
	}
	return v.c.regularFalseType
}

func (v *typefnVM) boolean(node *ast.Node, value *Type) (bool, bool) {
	value = v.observe(value)
	if typefnUnresolved(value) {
		v.blocked = true
		return false, false
	}
	if value.flags == TypeFlagsBooleanLiteral {
		return value.AsLiteralType().value.(bool), true
	}
	v.error(node, "condition requires one boolean literal; deferred control flow is not implemented yet")
	return false, false
}

func (v *typefnVM) elements(node *ast.Node, value *Type, union bool) []*Type {
	value = v.observe(value)
	if typefnUnresolved(value) {
		v.blocked = true
		return nil
	}
	if union {
		// Conservative until the full graph-readiness protocol is implemented.
		if !typefnClosedScalar(value) {
			v.error(node, "union iteration currently requires closed scalar members")
			return nil
		}
		if value.flags&TypeFlagsNever != 0 {
			return nil
		}
		if value.flags&TypeFlagsUnion != 0 {
			return value.Types()
		}
		return []*Type{value}
	}
	if !isTupleType(value) || value.TargetTupleType().combinedFlags & ^ElementFlagsRequired != 0 {
		v.error(node, "iteration/spread requires a required finite tuple; string iteration is not implemented yet")
		return nil
	}
	return v.c.getTypeArguments(value)
}

func typefnClosedScalar(t *Type) bool {
	if t.flags&TypeFlagsUnion != 0 {
		for _, part := range t.Types() {
			if !typefnClosedScalar(part) {
				return false
			}
		}
		return true
	}
	return t.flags & ^(TypeFlagsAny|TypeFlagsUnknown|TypeFlagsNever|TypeFlagsString|TypeFlagsNumber|TypeFlagsBigInt|TypeFlagsBoolean|TypeFlagsStringLiteral|TypeFlagsNumberLiteral|TypeFlagsBigIntLiteral|TypeFlagsBooleanLiteral|TypeFlagsNull|TypeFlagsUndefined|TypeFlagsVoid) == 0
}

func (v *typefnVM) binary(node *ast.Node, op ast.Kind, a, b *Type) *Type {
	a, b = v.observe(a), v.observe(b)
	if op != ast.KindBarToken && op != ast.KindAmpersandToken &&
		!(a == b && (op == ast.KindEqualsEqualsToken || op == ast.KindExclamationEqualsToken)) &&
		(typefnUnresolved(a) || typefnUnresolved(b)) {
		return v.c.newTypefnTerm(&typefnTerm{kind: typefnBinaryTerm, node: node, operator: op, args: []*Type{a, b}})
	}
	switch op {
	case ast.KindBarToken:
		return v.c.getUnionType([]*Type{a, b})
	case ast.KindAmpersandToken:
		return v.c.getIntersectionType([]*Type{a, b})
	case ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken, ast.KindIsKeyword:
		if a != b && (!typefnClosedScalar(a) || !typefnClosedScalar(b)) {
			return v.error(node, "non-scalar predicates require the deferred-readiness implementation")
		}
		result := a == b
		if !result {
			if op == ast.KindIsKeyword {
				result = v.c.isTypeAssignableTo(a, b)
			} else {
				result = v.c.isTypeIdenticalTo(a, b)
			}
		}
		if op == ast.KindExclamationEqualsToken {
			result = !result
		}
		return v.boolType(result)
	}
	compare := func(result int) *Type {
		switch op {
		case ast.KindLessThanToken:
			return v.boolType(result < 0)
		case ast.KindLessThanEqualsToken:
			return v.boolType(result <= 0)
		case ast.KindGreaterThanToken:
			return v.boolType(result > 0)
		case ast.KindGreaterThanEqualsToken:
			return v.boolType(result >= 0)
		}
		return nil
	}
	if a.flags == TypeFlagsNumberLiteral && b.flags == TypeFlagsNumberLiteral {
		left, right := float64(a.AsLiteralType().value.(jsnum.Number)), float64(b.AsLiteralType().value.(jsnum.Number))
		if math.IsInf(left, 0) || math.IsInf(right, 0) || math.IsNaN(left) || math.IsNaN(right) {
			return v.error(node, "number operands must be finite")
		}
		cmp := 0
		if left < right {
			cmp = -1
		} else if left > right {
			cmp = 1
		}
		if result := compare(cmp); result != nil {
			return result
		}
		var result float64
		switch op {
		case ast.KindPlusToken:
			result = left + right
		case ast.KindMinusToken:
			result = left - right
		case ast.KindAsteriskToken:
			result = left * right
		case ast.KindSlashToken:
			if right == 0 {
				return v.error(node, "division by zero")
			}
			result = left / right
		case ast.KindPercentToken:
			if right == 0 {
				return v.error(node, "remainder by zero")
			}
			result = math.Mod(left, right)
		default:
			return v.error(node, "unsupported scalar operator")
		}
		if math.IsNaN(result) || math.IsInf(result, 0) {
			return v.error(node, "number arithmetic must produce a finite value")
		}
		if result == 0 {
			result = 0
		}
		return v.c.getNumberLiteralType(jsnum.Number(result))
	}
	if a.flags == TypeFlagsBigIntLiteral && b.flags == TypeFlagsBigIntLiteral {
		left, _ := new(big.Int).SetString(a.AsLiteralType().value.(jsnum.PseudoBigInt).String(), 10)
		right, _ := new(big.Int).SetString(b.AsLiteralType().value.(jsnum.PseudoBigInt).String(), 10)
		if left.BitLen() > v.limits.BigintBits || right.BitLen() > v.limits.BigintBits {
			return v.error(node, "bigint magnitude limit exceeded")
		}
		if op == ast.KindAsteriskToken && left.Sign() != 0 && right.Sign() != 0 && left.BitLen()-1 > v.limits.BigintBits-right.BitLen() {
			return v.error(node, "bigint magnitude limit exceeded")
		}
		if result := compare(left.Cmp(right)); result != nil {
			return result
		}
		result := new(big.Int)
		switch op {
		case ast.KindPlusToken:
			result.Add(left, right)
		case ast.KindMinusToken:
			result.Sub(left, right)
		case ast.KindAsteriskToken:
			result.Mul(left, right)
		case ast.KindSlashToken:
			if right.Sign() == 0 {
				return v.error(node, "division by zero")
			}
			result.Quo(left, right)
		case ast.KindPercentToken:
			if right.Sign() == 0 {
				return v.error(node, "remainder by zero")
			}
			result.Rem(left, right)
		default:
			return v.error(node, "unsupported bigint operator")
		}
		if result.BitLen() > v.limits.BigintBits {
			return v.error(node, "bigint magnitude limit exceeded")
		}
		text := result.String()
		return v.c.getBigIntLiteralType(jsnum.NewPseudoBigInt(strings.TrimPrefix(text, "-"), result.Sign() < 0))
	}
	return v.error(node, "scalar operator requires single unbranded literals of the same numeric family")
}

func (v *typefnVM) unary(node *ast.Node, op ast.Kind, value *Type) *Type {
	value = v.observe(value)
	if op != ast.KindKeyOfKeyword && typefnUnresolved(value) {
		return v.c.newTypefnTerm(&typefnTerm{kind: typefnUnaryTerm, node: node, operator: op, args: []*Type{value}})
	}
	switch op {
	case ast.KindKeyOfKeyword:
		return v.c.getIndexType(value)
	case ast.KindExclamationToken:
		b, ok := v.boolean(node, value)
		if !ok {
			return v.c.errorType
		}
		return v.boolType(!b)
	case ast.KindPlusToken, ast.KindMinusToken:
		zero := v.c.getNumberLiteralType(0)
		if value.flags == TypeFlagsBigIntLiteral {
			zero = v.c.getBigIntLiteralType(jsnum.PseudoBigInt{})
		}
		return v.binary(node, op, zero, value)
	}
	return v.error(node, "unsupported unary operator")
}
