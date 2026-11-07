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
 typefn Empty<> { return []; }`
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
	if !strings.Contains(printed, "return [];") || !strings.Contains(printed, "Empty<>") {
		t.Fatalf("lost required syntax: %s", printed)
	}
}
