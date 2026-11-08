package checker

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/binder"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/parser"
)

func TestTypefnLowering(t *testing.T) {
	for _, test := range []struct {
		name, body string
		valid      bool
	}{
		{"nested control", `type value = T; while (true) { if (false) { break; } else { continue; } } return value;`, true},
		{"iteration", `for (type item of [T]) { type copy = item; } return T;`, true},
		{"dead forward read", `if (false) { type a = b; type b = T; } return T;`, false},
		{"self initializer", `type value = value; return T;`, false},
		{"immutable parameter", `T = 1; return T;`, false},
		{"break outside loop", `break; return T;`, false},
		{"reserved local", `type numeric = T; return T;`, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/test.ts", PathKey: "/test.ts", TypefnVersion: 1}, "typefn F<T> {"+test.body+"}", core.ScriptKindTS)
			if len(file.Diagnostics()) != 0 {
				t.Fatal(file.Diagnostics())
			}
			binder.BindSourceFile(file)
			c := &Checker{}
			declaration := file.Statements.Nodes[0]
			program := c.compileTypefn(declaration)
			if program.valid != test.valid {
				t.Fatalf("valid=%v, want %v", program.valid, test.valid)
			}
			if program != c.compileTypefn(declaration) {
				t.Fatal("lowered program was not reused")
			}
			if !program.valid {
				return
			}
			for _, instruction := range program.code {
				switch instruction.op {
				case tfJump, tfBranch, tfIterNext:
					if instruction.target < 0 || instruction.target > len(program.code) {
						t.Fatalf("invalid jump target: %d", instruction.target)
					}
				}
			}
			other := &Checker{}
			if program == other.compileTypefn(declaration) {
				t.Fatal("program shared across checker snapshots")
			}
		})
	}
}
