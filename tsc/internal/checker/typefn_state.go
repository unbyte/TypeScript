package checker

import "github.com/microsoft/TypeScript/tsc/internal/ast"

// All DSL-owned state belongs to one checker snapshot. Native-only checking does
// not allocate it. The VM is scoped to a root evaluation; programs, immutable
// terms and syntax-dependency tables live for the owning checker lifetime.
type typefnState struct {
	programs       map[*ast.Node]*typefnProgram
	vm             *typefnVM
	terms          map[typefnTermKey]*Type
	syntaxChecks   map[*ast.Node]bool
	objects        map[*Type]bool
	propertyValues map[*ast.Symbol]*Type
}

func (c *Checker) getTypefnState() *typefnState {
	if c.typefn == nil {
		c.typefn = &typefnState{}
	}
	return c.typefn
}
