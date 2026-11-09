package checker

import (
	"context"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/jsnum"
)

func TestTypefnDeclaredConstraintRelation(t *testing.T) {
	c, _ := newCompositeTestChecker()
	c.ctx = context.Background()
	c.initializeClosures()
	c.compilerOptions = &core.CompilerOptions{Typefn: &core.TypefnOptions{Version: 1}}
	c.noConstraintType = c.newObjectType(ObjectFlagsAnonymous, nil)
	c.circularConstraintType = c.newObjectType(ObjectFlagsAnonymous, nil)
	c.numberLiteralTypes = make(map[jsnum.Number]*Type)
	c.assignableRelation, c.identityRelation = &Relation{}, &Relation{}
	input := c.newTypeParameter(nil)
	input.AsTypeParameter().constraint = c.noConstraintType
	bound := c.newTypefnTerm(&typefnTerm{kind: typefnBinaryTerm, operator: ast.KindPlusToken, args: []*Type{input, c.getNumberLiteralType(1)}})
	other := c.newTypefnTerm(&typefnTerm{kind: typefnBinaryTerm, operator: ast.KindMinusToken, args: []*Type{input, c.getNumberLiteralType(1)}})
	parameter := c.newTypeParameter(nil)
	parameter.AsTypeParameter().constraint = bound
	chained := c.newTypeParameter(nil)
	chained.AsTypeParameter().constraint = parameter
	for _, source := range []*Type{parameter, chained} {
		if !c.isTypeAssignableTo(source, bound) {
			t.Fatal("declared bound was not available to native assignability")
		}
		if c.isTypeIdenticalTo(source, bound) {
			t.Fatal("a declared constraint became native Type identity")
		}
		if c.isTypeAssignableTo(bound, source) {
			t.Fatal("the bound was treated as assignable to every specialization")
		}
		if c.isTypeAssignableTo(source, other) {
			t.Fatal("a different deferred computation acquired an assumption")
		}
	}
	if parameter.AsTypeParameter().constraint != bound {
		t.Fatal("the declared computation was discharged")
	}
	c.getTypefnState().vm = &typefnVM{failed: true}
	if c.isTypeAssignableTo(parameter, bound) {
		t.Fatal("a failed evaluation published an assumption proof")
	}
}
