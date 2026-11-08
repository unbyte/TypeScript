package checker

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
)

// Zero denotes Type; a positive value denotes Constructor(value-1). Nil is the
// fixed Type kind. Variables belong to one checker and never contain native Types.
type typefnKind struct {
	parent *typefnKind
	value  int
	rank   uint8
}

func (k *typefnKind) root() *typefnKind {
	if k == nil {
		return nil
	}
	root := k
	for root.parent != nil {
		root = root.parent
	}
	for k.parent != nil {
		next := k.parent
		k.parent = root
		k = next
	}
	return root
}
func (k *typefnKind) resolved() int {
	if root := k.root(); root != nil {
		return root.value
	}
	return 0
}
func typefnConstructorKind(arity int) *typefnKind { return &typefnKind{value: arity + 1} }

// Unification only equates finite kinds. It does not evaluate a body, compare
// native bounds or infer constructor subtyping. Unknown result variables let
// mutually referring definitions be checked as a single static constraint set.
func unifyTypefnKinds(a, b *typefnKind) bool {
	a, b = a.root(), b.root()
	if a == b {
		return true
	}
	if a == nil || b == nil {
		if a == nil {
			a = b
		}
		if a.value > 0 {
			return false
		}
		a.value = 0
		return true
	}
	if a.value >= 0 && b.value >= 0 && a.value != b.value {
		return false
	}
	if a.rank < b.rank {
		a, b = b, a
	}
	if a.value < 0 {
		a.value = b.value
	}
	b.parent = a
	if a.rank == b.rank {
		a.rank++
	}
	return true
}
func (b *typefnCompiler) sameKind(node *ast.Node, a, c *typefnKind, message string) {
	if !unifyTypefnKinds(a, c) {
		b.error(node, message)
	}
}
func typefnKindDisplay(kind *typefnKind) string {
	switch n := kind.resolved(); {
	case n < 0:
		return "unresolved kind"
	case n == 0:
		return "type"
	default:
		if n == 1 {
			return "typefn<>"
		}
		return "typefn<" + strings.Repeat("_, ", n-2) + "_>"
	}
}

func (c *Checker) finishTypefnKinds(programs []*typefnProgram) {
	for _, p := range programs {
		if !p.valid {
			continue
		}
		if !p.hasValueReturn && ast.TypeFunctionResultKind(p.declaration) == nil && !ast.IsTypeFunctionLambda(p.declaration) {
			c.typefnError(p.declaration, "a function without a value-returning path requires a result-kind annotation")
			p.valid = false
		} else if p.resultKind.resolved() < 0 {
			c.typefnError(p.declaration, "result kind cannot be inferred; add an explicit result-kind annotation")
			p.valid = false
		}
	}
}

// Finalize a dependency component before its callers constrain application
// results. Otherwise a later caller could accidentally supply a missing recursive
// annotation, making definition errors depend on source or editor query order.
func (b *typefnCompiler) compileFunction(node *ast.Node, parent *typefnCompiler) *typefnProgram {
	known := b.c.getTypefnState().programs[node] != nil
	var p *typefnProgram
	if parent != nil {
		p = b.c.compileTypefnWithParent(node, parent)
	} else {
		p = b.c.compileTypefn(node)
	}
	b.kindDependency(p, !known)
	return p
}
func (b *typefnCompiler) kindDependency(p *typefnProgram, treeEdge bool) {
	if !p.kindOnStack {
		return
	}
	index := p.kindIndex
	if treeEdge {
		index = p.kindLowlink
	}
	b.p.kindLowlink = min(b.p.kindLowlink, index)
}
