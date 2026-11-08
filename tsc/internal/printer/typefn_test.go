package printer_test

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/parser"
	"github.com/microsoft/TypeScript/tsc/internal/printer"
)

func TestTypefnRoundTrip(t *testing.T) {
	source := `export typefn F<T extends string> {
 type x = {readonly value?: T, ["key"]: [1, ...T]};
 while (true) { if (x == T) { break; } else if (false) { continue; } }
 for (type item of T) { x = F<(item < 2)>; }
 for (type item in T) { F<item>; }
 return [-(1 + 2) * 3, !false, type((v: T) => T), x["value"]];
 }
 typefn Empty<>: type { return []; }
 typefn Factory<>: typefn<_, _> { return []; }`
	source += "\ntypefn Apply<F<_, _>, A, B extends types.array<A>> { return F<A,B>; }"
	source += "\ntypefn Lambdas<T> { type F = typefn<U extends T>: type { return U; }; type Z = typefn<> { return 1; }; return F<Z<>>; }"
	source += "\ntypefn Text<T> { return `head${T + 1} middle${`nested${T}`} tail\\`${T}`; }"
	source += "\ntypefn Patterns<T> { match (T) { [head, {value: item}, ...rest] if (true) => { return head; } is F<T> => { return 1; } -2n => { return 2; } type(string) => { return 3; } `GET ${path}${=text.concat<[\"/\", \":id\"]>}${_}` => { return path; } _ => { return never; } } }"
	opts := ast.SourceFileParseOptions{FileName: "/test.ts", PathKey: "/test.ts", TypefnVersion: 1}
	file := parser.ParseSourceFile(opts, source, core.ScriptKindTS)
	if ds := file.Diagnostics(); len(ds) > 0 {
		t.Fatalf("input: %v", ds)
	}
	p := printer.NewPrinter(printer.PrinterOptions{NewLine: core.NewLineKindLF, OmitTrailingSemicolon: true}, printer.PrintHandlers{}, nil)
	printed := p.EmitSourceFile(file)
	again := parser.ParseSourceFile(opts, printed, core.ScriptKindTS)
	if ds := again.Diagnostics(); len(ds) > 0 {
		t.Fatalf("printed source: %s\n%v", printed, ds)
	}
	if roundTrip := p.EmitSourceFile(again); roundTrip != printed {
		t.Fatalf("unstable printer:\n%s\n%s", printed, roundTrip)
	}
	if !strings.Contains(printed, "return [];") || !strings.Contains(printed, "Empty<>: type") || !strings.Contains(printed, "Factory<>: typefn<_, _>") {
		t.Fatalf("lost required syntax: %s", printed)
	}
}

func TestTypefnPatternRecoveryPrint(t *testing.T) {
	source := "typefn F<T> { match (T) { " + strings.Repeat("[", 100) + "item" + strings.Repeat("]", 100) + " => {} _ => {} } }"
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/test.ts", PathKey: "/test.ts", TypefnVersion: 1, TypefnSyntaxDepth: 16}, source, core.ScriptKindTS)
	if len(file.Diagnostics()) == 0 {
		t.Fatal("expected recovery nodes")
	}
	p := printer.NewPrinter(printer.PrinterOptions{}, printer.PrintHandlers{}, nil)
	p.EmitSourceFile(file)
}
