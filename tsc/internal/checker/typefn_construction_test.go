package checker

import (
	"context"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
)

func TestTypefnConstructionDoesNotPublishScratchState(t *testing.T) {
	c, _ := newCompositeTestChecker()
	c.ctx = context.Background()
	c.initializeClosures()
	c.compilerOptions = &core.CompilerOptions{Typefn: &core.TypefnOptions{Version: 1}}
	v := &typefnVM{c: c, limits: core.TypefnLimits{NativeNodes: 1000, LiveBytes: 1 << 20}, construction: &typefnConstructionArena{}}
	c.getTypefnState().vm = v
	self := v.construct(nil, &typefnConstruction{kind: ast.KindTypeReference})
	if self == c.errorType || len(c.typefn.terms) != 0 {
		t.Fatal("private reference entered recipe cache")
	}
	if shape, checks := splitTypefn(self); shape != self || len(checks) != 0 {
		t.Fatal("private reference became a pending requirement")
	}
	p := &typefnProgram{}
	first := c.internTypefnClosure(p, []typefnValue{{native: self}})
	second := c.internTypefnClosure(p, []typefnValue{{native: self}})
	if first == second || len(c.typefn.closures) != 0 {
		t.Fatal("closure capturing Self entered ordinary cache")
	}
	if got := c.newTypefnTerm(&typefnTerm{kind: typefnBinaryTerm, args: []*Type{self}}); got != c.errorType || !v.blocked || !v.construction.blocked || len(c.typefn.terms) != 0 {
		t.Fatal("blocked construction retained a scratch recipe")
	}
}
