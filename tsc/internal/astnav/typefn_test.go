package astnav_test

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/astnav"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/parser"
)

func TestTypefnRecoveryNavigation(t *testing.T) {
	for _, source := range []string{
		"typefn F<T> { return " + strings.Repeat("[", 40) + "T; } const tail = 1;",
		"typefn F<T> { " + strings.Repeat("if (false) {} else ", 40) + "{ return T; } }",
		"typefn F<T> { return " + strings.Repeat("F<", 40) + "T; }",
		"typefn F<T> { type local = ; return local; }",
	} {
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/main.ts", PathKey: "/main.ts", TypefnVersion: 1, TypefnSyntaxDepth: 16}, source, core.ScriptKindTS)
		if len(file.Diagnostics()) == 0 {
			t.Fatal("expected recovery syntax")
		}
		for pos := 0; pos <= len(source); pos++ {
			astnav.GetTokenAtPosition(file, pos)
			astnav.GetTouchingToken(file, pos)
			astnav.GetTouchingPropertyName(file, pos)
			astnav.FindPrecedingToken(file, pos)
		}
	}
}
