package parser_test

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/parser"
	"github.com/microsoft/TypeScript/tsc/internal/tspath"
)

func TestTypefnDeclarationABI(t *testing.T) {
	for _, test := range []struct{ name, source, message string }{
		{"valid", "// @typefn-abi: 1\n", ""},
		{"BOM and trivia", "\ufeff /* license */\r\n\t// @typefn-abi: 1\r\n", ""},
		{"duplicate", "// @typefn-abi: 1\n// @typefn-abi: 1\n", ""},
		{"missing", "", "requires"},
		{"unsupported", "// @typefn-abi: 2\n", "unsupported"},
		{"empty", "// @typefn-abi: \n", "unsupported"},
		{"conflicting", "// @typefn-abi: 1\n// @typefn-abi: 2\n", "conflicting"},
		{"block text", "/* // @typefn-abi: 1 */\n", "requires"},
		{"after statement", "export type T=1;\n// @typefn-abi: 1\n", "requires"},
		{"string text", "export type T=\"// @typefn-abi: 1\";\n", "requires"},
	} {
		t.Run(test.name, func(t *testing.T) {
			file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/test.d.ts", PathKey: "/test.d.ts", TypefnVersion: 1}, test.source+"export typefn Id<T> {return T;}", core.ScriptKindTS)
			diagnostics := file.Diagnostics()
			if test.message == "" {
				if len(diagnostics) != 0 {
					t.Fatalf("unexpected diagnostics: %v", diagnostics)
				}
				return
			}
			if len(diagnostics) != 1 || !strings.Contains(diagnostics[0].String(), test.message) {
				t.Fatalf("diagnostics: %v", diagnostics)
			}
		})
	}
	for _, name := range []string{"/source.ts", "/source.d.ts"} {
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePath(name), PathKey: tspath.PathKey(name), TypefnVersion: 1}, "export type Plain=1;", core.ScriptKindTS)
		if len(file.Diagnostics()) != 0 {
			t.Fatalf("native-only file requires no header: %v", file.Diagnostics())
		}
	}
}
