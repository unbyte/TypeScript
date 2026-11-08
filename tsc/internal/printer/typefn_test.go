package printer_test

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/parser"
	"github.com/microsoft/TypeScript/tsc/internal/printer"
	"github.com/microsoft/TypeScript/tsc/internal/sourcemap"
	"github.com/microsoft/TypeScript/tsc/internal/tspath"
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

func TestTypefnDeclarationABIPrinting(t *testing.T) {
	for _, removeComments := range []bool{false, true} {
		for _, newline := range []core.NewLineKind{core.NewLineKindLF, core.NewLineKindCRLF} {
			file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/test.ts", PathKey: "/test.ts", TypefnVersion: 1}, `// ordinary comment
export typefn Identity<T> { return T; }`, core.ScriptKindTS)
			file.IsDeclarationFile = true
			context := printer.NewEmitContext()
			context.SetTypefnDeclarationABI(file)
			p := printer.NewPrinter(printer.PrinterOptions{NewLine: newline, RemoveComments: removeComments, OnlyPrintJSDocStyle: true}, printer.PrintHandlers{}, context)
			printed := p.EmitSourceFile(file)
			if !strings.HasPrefix(printed, "// @typefn-abi: 1") || strings.Count(printed, "@typefn-abi:") != 1 {
				t.Fatalf("missing ABI: %q", printed)
			}
			again := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/test.d.ts", PathKey: "/test.d.ts", TypefnVersion: 1}, printed, core.ScriptKindTS)
			if ds := again.Diagnostics(); len(ds) != 0 {
				t.Fatalf("invalid emitted declaration: %v\n%s", ds, printed)
			}
			// Reusing an emit context/printer must not leak a file's ABI to other outputs.
			native := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/native.d.ts", PathKey: "/native.d.ts"}, "export type Native = string;", core.ScriptKindTS)
			if output := p.EmitSourceFile(native); strings.Contains(output, "@typefn-abi:") {
				t.Fatalf("ABI leaked to native file: %s", output)
			}
			file.IsDeclarationFile = false
			if output := p.EmitSourceFile(file); strings.Contains(output, "@typefn-abi:") {
				t.Fatalf("ABI leaked to source output: %s", output)
			}
		}
	}
}

func TestTypefnDeclarationSourceMaps(t *testing.T) {
	source := "export typefn Identity<T> {\n  type copy = T;\n  return copy;\n}"
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/test.ts", PathKey: "/test.ts", TypefnVersion: 1}, source, core.ScriptKindTS)
	file.IsDeclarationFile = true
	context := printer.NewEmitContext()
	context.SetTypefnDeclarationABI(file)
	p := printer.NewPrinter(printer.PrinterOptions{NewLine: core.NewLineKindLF, SourceMap: true, OnlyPrintJSDocStyle: true}, printer.PrintHandlers{}, context)
	writer := printer.NewTextWriter("\n", 4)
	generator := sourcemap.NewGenerator("test.d.ts", "", "/", tspath.CaseSensitive)
	p.Write(file.AsNode(), file, writer, generator)
	decoder := sourcemap.DecodeMappings(generator.RawSourceMap().Mappings)
	// The semantic ABI line shifts all generated positions by one. Check statement
	// starts inside the DSL body, not just the identifier mappings native printing
	// already supplied. These positions drive declaration-map navigation.
	found := map[int]bool{}
	for mapping := range decoder.Values() {
		if !mapping.IsSourceMapping() {
			continue
		}
		if mapping.GeneratedLine == 0 {
			t.Fatal("ABI pragma acquired a false source mapping")
		}
		if mapping.GeneratedCharacter == 4 && mapping.SourceCharacter == 2 && mapping.GeneratedLine == mapping.SourceLine+1 {
			found[mapping.SourceLine] = true
		}
	}
	if err := decoder.Error(); err != nil {
		t.Fatal(err)
	}
	for _, line := range []int{1, 2} {
		if !found[line] {
			t.Fatalf("missing statement mapping for source line %d: %s\n%s", line, generator.RawSourceMap().Mappings, writer.String())
		}
	}
}
