package checker

import (
	"slices"
	"unsafe"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/diagnostics"
)

// A root owns the budget; nested DSL/native calls cannot reset it. Failed roots
// are not cached. Programs and all Type handles belong to this checker snapshot.
type typefnActiveCall struct {
	declaration *ast.Node
	args        []*Type
	closure     *typefnClosure
}

// Iterator slots belong to the lowered program, while snapshots belong to each
// invocation. A nested/re-entered loop replaces only its own frame-local slot.
type typefnIterator struct {
	values []*Type
	index  int
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
	stringUnits         int
	textUnits           map[string][]uint16
	rootSite            *ast.Node
	replaySite          *ast.Node
	nativeIslandSite    *ast.Node
}

func (c *Checker) evaluateTypefn(node, declaration *ast.Node, args []*Type) *Type {
	return c.runTypefn(node, func(v *typefnVM) *Type { return v.call(node, declaration, args) })
}

func (c *Checker) runTypefn(node *ast.Node, run func(*typefnVM) *Type) *Type {
	state := c.getTypefnState()
	root := state.vm == nil
	if root {
		limits := core.TypefnLimits{VMInstructions: 1000000, CallDepth: 256, NativeNodes: 250000, BigintBits: 1048576, LiveBytes: 64 << 20, StringUnits: 16777216}
		if c.compilerOptions.Typefn != nil {
			configured := c.compilerOptions.Typefn.Limits
			if configured.NativeWork != 0 || configured.CacheBytes != 0 {
				c.typefnError(node, "requested nativeWork or cacheBytes limit is not implemented yet")
				return c.errorType
			}
			if configured.StringUnits != 0 {
				limits.StringUnits = configured.StringUnits
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
		state.vm = &typefnVM{c: c, limits: limits, rootSite: node}
		defer func() { state.vm = nil }()
	}
	v := state.vm
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
		if v.replaySite == nil && v.rootSite != nil {
			// Attribute each evaluation to its application, even when two calls in the
			// same file fail at the same operation. Per-file checking also requires
			// consumer attribution for failures inside imported helpers.
			site = v.rootSite
		}
		if site == nil {
			site = v.c.currentNode
		}
		diagnostic := v.c.error(site, diagnostics.NewAdHocMessage("typefn: "+message))
		if node != nil && node != site {
			diagnostic.AddRelatedInfo(NewDiagnosticForNode(node, diagnostics.NewAdHocMessage("typefn operation failed here")))
		}
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

func (v *typefnVM) call(node, declaration *ast.Node, args []*Type) *Type {
	return v.callWithClosure(node, declaration, args, nil)
}

// Native Type consumers and deferred Type recipes use this boundary. Factories
// stay in the value evaluator; they never enter a native Type operand vector.
func (v *typefnVM) callWithClosure(node, declaration *ast.Node, args []*Type, closure *typefnClosure) *Type {
	p := v.c.compileTypefn(declaration)
	if closure != nil {
		p = closure.program
	}
	if p.valid && p.resultKind.resolved() > 0 {
		return v.error(node, "constructor value cannot be used as a Type")
	}
	value := v.callValue(node, declaration, args, closure)
	if v.failed || v.blocked {
		return v.c.errorType
	}
	return v.nativeValue(node, value)
}

func (v *typefnVM) callValue(node, declaration *ast.Node, args []*Type, closure *typefnClosure) (result typefnValue) {
	// Module helpers have no captured environment. Canonicalize direct and
	// constructor-valued applications to the same call recipe and cycle identity.
	if closure != nil && !closure.program.local {
		closure = nil
	}
	var p *typefnProgram
	if closure != nil {
		p = closure.program
	} else {
		p = v.c.compileTypefn(declaration)
	}
	originalArgs := args
	savedChecks, savedSet, savedBlocked := v.checks, v.checkSet, v.blocked
	v.checks, v.checkSet, v.blocked = nil, nil, false
	defer func() {
		// A factory may publish only a known constructor whose creation checks
		// completed. It cannot hide pending checks in a fake native Type. Block
		// the enclosing Type computation so its immutable call recipe can replay.
		blockConstructor := !v.failed && p.resultKind.resolved() > 0 && (v.blocked || len(v.checks) != 0)
		if blockConstructor {
			result = typefnValue{}
		} else if v.blocked && !v.failed {
			result = typefnValue{native: v.c.newTypefnTerm(&typefnTerm{kind: typefnCallTerm, node: node, declaration: declaration, args: originalArgs, closure: closure})}
		} else if result.native != nil {
			result.native = v.c.guardTypefn(result.native, v.checks)
		}
		v.checks, v.checkSet, v.blocked = savedChecks, savedSet, savedBlocked || blockConstructor
		if result.native != nil {
			v.observe(result.native)
		}
	}()
	args = slices.Clone(args)
	for i, arg := range args {
		args[i] = v.observe(arg)
	}
	// Recovery Types are not successful Type values, even when a parameter is unused.
	for _, arg := range args {
		if v.c.isErrorType(arg) {
			v.failed = true
			return typefnValue{native: v.c.errorType}
		}
	}
	if len(ast.GetSourceFileOfNode(declaration).Diagnostics()) != 0 {
		v.failed = true
		return typefnValue{native: v.c.errorType}
	}
	if v.depth >= v.limits.CallDepth {
		return typefnValue{native: v.error(node, "call depth limit exceeded")}
	}
	for _, active := range v.active {
		if active.declaration != declaration || active.closure != closure || len(active.args) != len(args) {
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
			return typefnValue{native: v.error(node, "recursive evaluation cycle")}
		}
	}
	v.active = append(v.active, typefnActiveCall{declaration: declaration, args: args, closure: closure})
	v.depth++
	defer func() { v.depth--; v.active = v.active[:len(v.active)-1] }()
	if !p.valid {
		v.failed = true
		return typefnValue{native: v.c.errorType}
	}
	if len(args) != len(declaration.TypeParameters()) {
		return typefnValue{native: v.error(node, "wrong number of Type arguments")}
	}
	if closure == nil {
		if p.local {
			return typefnValue{native: v.error(node, "local helper requires its captured environment")}
		}
		parameters := v.c.appendTypeParameters(nil, declaration.TypeParameters())
		if !v.constraints(node, parameters, args) {
			return typefnValue{native: v.c.errorType}
		}
	}
	// Charge register/argument backing storage and iterator slots, in addition to
	// the existing fixed bookkeeping allowance. No iterator storage is allocated
	// for programs that do not iterate.
	frameBytes := p.slots*int(unsafe.Sizeof(typefnValue{})) + len(args)*int(unsafe.Sizeof((*Type)(nil))) + p.iterators*int(unsafe.Sizeof(typefnIterator{})) + 128
	if frameBytes > v.limits.LiveBytes-v.liveBytes {
		return typefnValue{native: v.error(node, "live VM allocation limit exceeded")}
	}
	v.liveBytes += frameBytes
	defer func() { v.liveBytes -= frameBytes }()
	registers := make([]typefnValue, p.slots)
	for i, arg := range args {
		registers[i] = typefnValue{native: arg}
	}
	if closure != nil {
		if p.selfSlot >= 0 {
			registers[p.selfSlot] = typefnValue{closure: closure}
		}
		for i, capture := range p.captures {
			registers[capture.target] = closure.captures[i]
		}
		if !v.helperConstraints(node, p, registers, args) {
			return typefnValue{native: v.c.errorType}
		}
	}
	iterators := make([]typefnIterator, p.iterators)
	for pc := 0; pc < len(p.code); {
		i := &p.code[pc]
		pc++
		if !v.step(i.node) {
			return typefnValue{native: v.c.errorType}
		}
		var value typefnValue
		if i.expr != nil {
			if i.op == tfStore || i.op == tfReturn || i.op == tfDiscard {
				value = v.resultExpression(i.expr, registers)
			} else {
				value = typefnValue{native: v.expression(i.expr, registers)}
			}
			if v.failed || v.blocked {
				return typefnValue{native: v.c.errorType}
			}
		}
		switch i.op {
		case tfHelper:
			registers[i.slot] = typefnValue{closure: v.captureHelper(i.helper, registers)}
			if v.failed {
				return typefnValue{native: v.c.errorType}
			}
		case tfStore:
			registers[i.slot] = value
		case tfReturn:
			return value
		case tfBranch:
			truth, ok := v.boolean(i.node, value.native)
			if !ok {
				return typefnValue{native: v.c.errorType}
			}
			if !truth {
				pc = i.target
			}
		case tfMatch:
			matched := v.pattern(i.pattern, v.nativeValue(i.node, registers[i.slot]), registers)
			if v.failed || v.blocked {
				return typefnValue{native: v.c.errorType}
			}
			if !matched {
				pc = i.target
			}
		case tfJump:
			pc = i.target
		case tfIterStart:
			values := v.elements(i.node, value.native, i.node.Kind == ast.KindTypeFunctionForIn)
			if v.failed || v.blocked {
				return typefnValue{native: v.c.errorType}
			}
			iterators[i.iterator] = typefnIterator{values: values}
		case tfIterNext:
			iter := &iterators[i.iterator]
			if iter.index == len(iter.values) {
				*iter = typefnIterator{}
				pc = i.target
			} else {
				registers[i.slot] = typefnValue{native: iter.values[iter.index]}
				iter.index++
			}
		}
	}
	return typefnValue{native: v.error(node, "execution reached the end without returning a Type")}
}
