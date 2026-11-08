package checker

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/jsnum"
)

func TestTypefnScalarEvaluationAndTermOwnership(t *testing.T) {
	c, _ := newCompositeTestChecker()
	c.compilerOptions = &core.CompilerOptions{}
	c.numberLiteralTypes = make(map[jsnum.Number]*Type)
	c.regularTrueType = c.newLiteralType(TypeFlagsBooleanLiteral, true, nil)
	c.regularFalseType = c.newLiteralType(TypeFlagsBooleanLiteral, false, nil)
	v := &typefnVM{c: c}
	two, three := c.getNumberLiteralType(2), c.getNumberLiteralType(3)
	if got := v.binary(nil, ast.KindPlusToken, two, three); got != c.getNumberLiteralType(5) {
		t.Fatal("addition did not produce a native literal identity")
	}
	if got := v.binary(nil, ast.KindLessThanToken, two, three); got != c.regularTrueType {
		t.Fatal("comparison did not produce literal true")
	}
	parameter := c.newType(TypeFlagsTypeParameter, ObjectFlagsNone, (&TypeParameter{}).AsType(), typeKindTypeParameter)
	args := []*Type{parameter, two}
	term := c.newTypefnTerm(&typefnTerm{kind: typefnBinaryTerm, operator: ast.KindPlusToken, args: args})
	args[1] = three
	if term.asTypefnTerm().args[1] != two {
		t.Fatal("term borrowed mutable caller operands")
	}
	same := c.newTypefnTerm(&typefnTerm{kind: typefnBinaryTerm, operator: ast.KindPlusToken, args: []*Type{parameter, two}})
	if same != term {
		t.Fatal("equal recipes did not retain identity")
	}
	if got := v.binary(nil, ast.KindPlusToken, parameter, two); got != term {
		t.Fatal("symbolic arithmetic did not produce the interned recipe")
	}
	if !typefnUnresolved(term) {
		t.Fatal("recipe was incorrectly treated as a concrete value")
	}
}
