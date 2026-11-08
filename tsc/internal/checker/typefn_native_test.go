package checker

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/jsnum"
	"testing"
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
