package encoder_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/api/encoder"
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/parser"
	"github.com/microsoft/TypeScript/tsc/internal/printer"
)

func TestTypefnEncodeDecode(t *testing.T) {
	t.Parallel()
	source := `export typefn F<T> {
 type x = -1;
 type flag = !(T is never);
 if (T is number) { x = T * 2; } else { x = 0; }
 return { readonly data?: [x, ...T]; ["key"]: F<x> };
 }`
	source += "\ntypefn Lambdas<T> { type F = typefn<U extends T>: type { return U; }; type Z = typefn<> { return 1; }; return F<Z<>>; }"
	source += "\ntypefn Text<T> { return `a${T + 1}b${T}c`; }"
	source += "\ntypefn Match<T> { match (T) { [head, ...rest] if (true) => { return head; } `GET ${path}${=\"/\"}` => { return path; } _ => { return never; } } }"
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/test.ts", PathKey: "/test.ts", TypefnVersion: 1}, source, core.ScriptKindTS)
	if ds := file.Diagnostics(); len(ds) != 0 {
		t.Fatalf("parse: %v", ds)
	}
	buf, table, err := encoder.EncodeSourceFile(file)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := encoder.DecodeSourceFile(buf)
	if err != nil {
		t.Fatal(err)
	}
	if !ast.IsTypeFunctionDeclaration(decoded.Statements.Nodes[0]) {
		t.Fatal("lost declaration kind")
	}
	p := printer.NewPrinter(printer.PrinterOptions{NewLine: core.NewLineKindLF}, printer.PrintHandlers{}, nil)
	if got, want := p.EmitSourceFile(decoded), p.EmitSourceFile(file); got != want {
		t.Fatalf("DSL changed across API serialization:\ngot %s\nwant %s", got, want)
	}
	decodedTable := encoder.BuildNodeIndexTable(decoded)
	if len(decodedTable.Nodes) != len(table.Nodes) {
		t.Fatal("lost AST children")
	}
}
