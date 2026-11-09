package checker

import (
	"context"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/jsnum"
)

func TestTypefnNativeAbsorptionAndSubstitution(t *testing.T) {
	c, _ := newCompositeTestChecker()
	c.ctx = context.Background()
	c.initializeClosures()
	c.compilerOptions = &core.CompilerOptions{Typefn: &core.TypefnOptions{Version: 1}}
	c.numberLiteralTypes = make(map[jsnum.Number]*Type)
	parameter := c.newType(TypeFlagsTypeParameter, ObjectFlagsNone, (&TypeParameter{}).AsType(), typeKindTypeParameter)
	term := c.newTypefnTerm(&typefnTerm{kind: typefnBinaryTerm, operator: ast.KindPlusToken, args: []*Type{parameter, c.getNumberLiteralType(2)}})
	for _, test := range []struct {
		name          string
		result, shape *Type
	}{
		{"unknown absorbs union", c.getUnionType([]*Type{c.unknownType, term}), c.unknownType},
		{"never absorbs intersection", c.getIntersectionType([]*Type{c.neverType, term}), c.neverType},
	} {
		t.Run(test.name, func(t *testing.T) {
			shape, checks := splitTypefn(test.result)
			if shape != test.shape || len(checks) != 1 || checks[0] != term {
				t.Fatal("native simplification discarded a reached calculation")
			}
			mapper := newTypeMapper([]*Type{parameter}, []*Type{c.getNumberLiteralType(3)})
			if got := c.instantiateType(term, mapper); got != c.getNumberLiteralType(5) {
				t.Fatal("native substitution did not replay arithmetic")
			}
			if got := c.instantiateType(test.result, mapper); got != test.shape {
				t.Fatal("completed validation did not release result shape")
			}
		})
	}
}

func TestTypefnContextualBaseConstraints(t *testing.T) {
	c, _ := newCompositeTestChecker()
	c.numberType = c.newIntrinsicType(TypeFlagsNumber, "number")
	c.noConstraintType = c.newObjectType(ObjectFlagsAnonymous, nil)
	c.circularConstraintType = c.newObjectType(ObjectFlagsAnonymous, nil)
	c.ctx = context.Background()
	c.initializeClosures()
	c.compilerOptions = &core.CompilerOptions{Typefn: &core.TypefnOptions{Version: 1}}
	c.numberLiteralTypes = make(map[jsnum.Number]*Type)
	parameter := c.newType(TypeFlagsTypeParameter, ObjectFlagsNone, (&TypeParameter{}).AsType(), typeKindTypeParameter)
	literal := c.getNumberLiteralType(2)
	term := c.newTypefnTerm(&typefnTerm{kind: typefnBinaryTerm, operator: ast.KindPlusToken, args: []*Type{parameter, literal}})
	if got := c.getBaseConstraintOfType(term); got != nil {
		t.Fatal("opaque computation invented a result constraint")
	}
	if c.isLiteralOfContextualType(literal, term) {
		t.Fatal("opaque computation supplied a literal context")
	}
	guard := c.guardTypefn(c.numberType, []*Type{term})
	if got := c.getBaseConstraintOfType(guard); got != c.numberType {
		t.Fatal("guard did not expose its known shape for contextual typing")
	}
	if !c.isLiteralOfContextualType(literal, guard) {
		t.Fatal("known numeric shape did not provide a numeric literal context")
	}
	shape, checks := splitTypefn(guard)
	if shape != c.numberType || len(checks) != 1 || checks[0] != term {
		t.Fatal("constraint resolution discarded the validation requirement")
	}
	if c.isTypeAssignableTo(literal, guard) {
		t.Fatal("contextual typing discharged an unresolved requirement")
	}
	mapper := newTypeMapper([]*Type{parameter}, []*Type{literal})
	if got := c.instantiateType(term, mapper); got != c.getNumberLiteralType(4) {
		t.Fatal("cached absent constraint prevented later specialization")
	}
	if got := c.instantiateType(guard, mapper); got != c.numberType {
		t.Fatal("specialization did not discharge the guarded requirement")
	}
}
