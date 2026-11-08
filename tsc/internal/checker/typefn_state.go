package checker

import "github.com/microsoft/TypeScript/tsc/internal/ast"

// All DSL-owned state belongs to one checker snapshot. Disabled-profile checking does
// not allocate it. The VM is scoped to a root evaluation; programs, immutable
// terms and syntax-dependency tables live for the owning checker lifetime.
type typefnState struct {
	programs       map[*ast.Node]*typefnProgram
	closures       map[typefnClosureKey]*typefnClosure
	nextClosureID  uint32
	vm             *typefnVM
	terms          map[typefnTermKey]*Type
	syntaxChecks   map[*ast.Node]bool
	objects        map[*Type]bool
	propertyValues map[*ast.Symbol]*Type
	tupleDeclared  map[*Type][]*Type
	tupleViews     map[typefnTupleKey]*Type
	tupleLabels    map[typefnTupleLabelKey]*ast.Node
}

func (c *Checker) getTypefnState() *typefnState {
	if c.typefn == nil {
		c.typefn = &typefnState{}
	}
	return c.typefn
}
