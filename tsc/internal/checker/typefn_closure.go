package checker

import (
	"slices"
	"unsafe"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/diagnostics"
)

// A helper closes over values, never mutable caller registers. Self-reference is
// supplied by the invocation frame, so immutable captures have no self-cycle.
type typefnClosure struct {
	id       uint32
	program  *typefnProgram
	captures []typefnValue
}

type typefnClosureKey struct {
	program  *typefnProgram
	captures CacheHashKey
}

func typefnValueHash(values []typefnValue) CacheHashKey {
	var key keyBuilder
	key.writeInt(len(values))
	for _, value := range values {
		switch {
		case value.closure != nil:
			key.writeUint32(2)
			key.writeUint32(value.closure.id)
		case value.native != nil:
			key.writeUint32(1)
			key.writeType(value.native)
		default:
			key.writeUint32(0)
		}
	}
	return key.hash()
}

func (c *Checker) internTypefnClosure(p *typefnProgram, captures []typefnValue) *typefnClosure {
	state := c.getTypefnState()
	key := typefnClosureKey{p, typefnValueHash(captures)}
	if value := state.closures[key]; value != nil {
		return value
	}
	if state.vm != nil {
		if !state.vm.allocate(p.declaration, len(captures)+1) || !state.vm.reserveValues(p.declaration, len(captures)*int(unsafe.Sizeof(typefnValue{}))+int(unsafe.Sizeof(typefnClosure{}))) {
			return nil
		}
	}
	if state.closures == nil {
		state.closures = map[typefnClosureKey]*typefnClosure{}
	}
	state.nextClosureID++
	value := &typefnClosure{id: state.nextClosureID, program: p, captures: slices.Clone(captures)}
	state.closures[key] = value
	return value
}

func (v *typefnVM) captureHelper(p *typefnProgram, registers []typefnValue) *typefnClosure {
	size := len(p.captures) * int(unsafe.Sizeof(typefnValue{}))
	if !v.reserveValues(p.declaration, size) {
		return nil
	}
	defer func() { v.liveBytes -= size }()
	captures := make([]typefnValue, len(p.captures))
	for i, capture := range p.captures {
		if !v.step(p.declaration) {
			return nil
		}
		captures[i] = registers[capture.source]
	}
	return v.c.internTypefnClosure(p, captures)
}

// Capture graphs are immutable DAGs. Traverse constructor edges iteratively so a
// long chain of snapshots cannot consume the host call stack.
func (c *Checker) instantiateTypefnClosure(root *typefnClosure, mapper *TypeMapper) *typefnClosure {
	if root == nil {
		return nil
	}
	type frame struct {
		value *typefnClosure
		exit  bool
	}
	work := []frame{{value: root}}
	mapped := map[*typefnClosure]*typefnClosure{}
	for len(work) != 0 {
		f := work[len(work)-1]
		work = work[:len(work)-1]
		if mapped[f.value] != nil {
			continue
		}
		if c.typefn.vm != nil && !c.typefn.vm.step(f.value.program.declaration) {
			return nil
		}
		if !f.exit {
			work = append(work, frame{value: f.value, exit: true})
			for _, capture := range f.value.captures {
				if capture.closure != nil && mapped[capture.closure] == nil {
					work = append(work, frame{value: capture.closure})
				}
			}
			continue
		}
		values := slices.Clone(f.value.captures)
		changed := false
		for i, value := range values {
			if value.closure != nil {
				values[i].closure = mapped[value.closure]
			} else if value.native != nil {
				values[i].native = c.instantiateType(value.native, mapper)
			}
			changed = changed || values[i] != value
		}
		if changed {
			mapped[f.value] = c.internTypefnClosure(f.value.program, values)
		} else {
			mapped[f.value] = f.value
		}
		if c.typefn.vm != nil && c.typefn.vm.failed {
			return nil
		}
	}
	return mapped[root]
}

func (v *typefnVM) callHelper(e *typefnExpression, slots []typefnValue) typefnValue {
	closure := slots[e.slot].closure
	if closure == nil {
		return typefnValue{native: v.error(e.node, "application requires a constructor value")}
	}
	args := make([]*Type, len(e.args))
	for i, argument := range e.args {
		args[i] = v.typeArgument(argument, slots)
		if v.failed || v.blocked {
			return typefnValue{native: v.c.errorType}
		}
	}
	return v.callValue(e.node, closure.program.declaration, args, closure)
}

func (c *Checker) instantiateTypefnValues(values []typefnValue, mapper *TypeMapper) []typefnValue {
	var result []typefnValue
	for i, value := range values {
		mapped := value
		if value.closure != nil {
			mapped.closure = c.instantiateTypefnClosure(value.closure, mapper)
		} else if value.native != nil {
			mapped.native = c.instantiateType(value.native, mapper)
		}
		if mapped != value && result == nil {
			result = slices.Clone(values)
		}
		if result != nil {
			result[i] = mapped
		}
	}
	if result == nil {
		return values
	}
	return result
}

func (v *typefnVM) inputConstraints(node *ast.Node, p *typefnProgram, registers []typefnValue, args []*Type) bool {
	for i, bound := range p.bounds {
		if bound == nil {
			continue
		}
		constraint := v.expression(bound, registers)
		if v.failed || v.blocked {
			return false
		}
		if !v.step(bound.node) {
			return false
		}
		if v.c.isTypeAssignableTo(args[i], constraint) {
			continue
		}
		if v.unresolvedGraph(bound.node, args[i]) || v.unresolvedGraph(bound.node, constraint) {
			if !v.failed {
				v.blocked = true
			}
			return false
		}
		if p.local {
			v.error(bound.node, "input Type does not satisfy the helper constraint")
		} else {
			site := node
			if v.rootSite != nil {
				site = v.rootSite
			}
			if v.replaySite != nil {
				site = v.replaySite
			}
			var reports []*ast.Diagnostic
			if v.c.checkTypeAssignableToEx(args[i], constraint, site, diagnostics.Type_0_does_not_satisfy_the_constraint_1, &reports) {
				continue
			}
			for _, report := range reports {
				report.AddRelatedInfo(NewDiagnosticForNode(bound.node, diagnostics.NewAdHocMessage("typefn input constraint declared here")))
				v.c.addDiagnostic(report)
			}
			v.failed = true
		}
		return false
	}
	return true
}

func (v *typefnVM) reserveValues(node *ast.Node, bytes int) bool {
	if bytes > v.limits.LiveBytes-v.liveBytes {
		v.error(node, "live VM allocation limit exceeded")
		return false
	}
	v.liveBytes += bytes
	return true
}
