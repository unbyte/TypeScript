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
	// Charge register/argument backing storage and iterator slots, in addition to
	// the existing fixed bookkeeping allowance. No iterator storage is allocated
	// for programs that do not iterate.
	frameBytes := (p.slots+len(args))*int(unsafe.Sizeof((*Type)(nil))) + p.iterators*int(unsafe.Sizeof(typefnIterator{})) + 128
	if frameBytes > v.limits.LiveBytes-v.liveBytes {
		return v.error(node, "live VM allocation limit exceeded")
	}
	v.liveBytes += frameBytes
	defer func() { v.liveBytes -= frameBytes }()
	registers := make([]*Type, p.slots)
	copy(registers, args)
	iterators := make([]typefnIterator, p.iterators)
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
		case tfMatch:
			matched := v.pattern(i.pattern, registers[i.slot], registers)
			if v.failed || v.blocked {
				return v.c.errorType
			}
			if !matched {
				pc = i.target
			}
		case tfJump:
			pc = i.target
		case tfIterStart:
			values := v.elements(i.node, value, i.node.Kind == ast.KindTypeFunctionForIn)
			if v.failed || v.blocked {
				return v.c.errorType
			}
			iterators[i.iterator] = typefnIterator{values: values}
		case tfIterNext:
			iter := &iterators[i.iterator]
			if iter.index == len(iter.values) {
				*iter = typefnIterator{}
				pc = i.target
			} else {
				registers[i.slot] = iter.values[iter.index]
				iter.index++
			}
		}
	}
	return v.error(node, "execution reached the end without returning a Type")
}
