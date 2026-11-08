package ls

import (
	"context"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/tspath"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/bundled"
	"github.com/microsoft/TypeScript/tsc/internal/compiler"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/ls/lsconv"
	"github.com/microsoft/TypeScript/tsc/internal/lsp/lsproto"
	"github.com/microsoft/TypeScript/tsc/internal/sourcemap"
	"github.com/microsoft/TypeScript/tsc/internal/tsoptions"
	"github.com/microsoft/TypeScript/tsc/internal/vfs/vfstest"
)

type typefnTestHost struct {
	Host
	files map[string]string
}

func (h typefnTestHost) CaseSensitivity() tspath.CaseSensitivity { return tspath.CaseSensitive }
func (h typefnTestHost) ReadFile(name tspath.RootedFilePath) (string, bool) {
	s, ok := h.files[string(name)]
	return s, ok
}

func (h typefnTestHost) FileExists(name tspath.RootedFilePath) bool {
	_, ok := h.files[string(name)]
	return ok
}

func (h typefnTestHost) GetECMALineInfo(tspath.RootedFilePath) *sourcemap.ECMALineInfo { return nil }

func typefnService(t testing.TB, source string) (*LanguageService, context.Context, func(string) lsproto.Position) {
	t.Helper()
	files := map[string]string{"/main.ts": source, "/tsconfig.json": `{"compilerOptions":{"typefn":{"version":1},"strict":true,"target":"es2022","ignoreDeprecations":"6.0"},"files":["main.ts"]}`}
	fs := bundled.WrapFS(vfstest.FromMap(files, tspath.CaseSensitive))
	host := compiler.NewCompilerHost(fs, bundled.LibPath(), nil, nil, nil)
	config, errors := tsoptions.GetParsedCommandLineOfConfigFile("/tsconfig.json", &core.CompilerOptions{}, nil, tsoptions.NewDefaultParseConfigHost(fs), nil)
	if len(errors) > 0 {
		t.Fatal(errors)
	}
	program := compiler.NewProgram(compiler.ProgramOptions{ProgramConfig: compiler.ProgramConfig{Config: config}, ProgramHosts: compiler.ProgramHosts{Host: host}})
	program.BindSourceFiles()
	converters := lsconv.NewConverters(lsproto.PositionEncodingKindUTF16, func(name tspath.RootedFilePath) *lsconv.LSPLineMap {
		return lsconv.ComputeLSPLineStarts(files[string(name)])
	})
	l := &LanguageService{program: program, converters: converters, host: typefnTestHost{files: files}}
	caps := &lsproto.ResolvedClientCapabilities{}
	caps.TextDocument.Hover.ContentFormat = []lsproto.MarkupKind{lsproto.MarkupKindMarkdown}
	caps.TextDocument.DocumentSymbol.HierarchicalDocumentSymbolSupport = true
	for _, kind := range tokenTypes {
		caps.TextDocument.SemanticTokens.TokenTypes = append(caps.TextDocument.SemanticTokens.TokenTypes, string(kind))
	}
	for _, mod := range tokenModifiers {
		caps.TextDocument.SemanticTokens.TokenModifiers = append(caps.TextDocument.SemanticTokens.TokenModifiers, string(mod))
	}
	ctx := lsproto.WithClientCapabilities(t.Context(), caps)
	position := func(marker string) lsproto.Position {
		t.Helper()
		if strings.Count(source, marker) != 1 {
			t.Fatalf("nonunique marker %s", marker)
		}
		pos, _ := converters.ToLSPPosition(program.GetSourceFile("/main.ts"), core.TextPos(strings.Index(source, marker)+len(marker)))
		return pos
	}
	return l, ctx, position
}

const typefnEditorSource = `export typefn Twice<N> {
 type /*decl*/current = N;
 { type /*shadow*/current = string; type saved = /*inner*/current; }
 /*write*/current = /*read*/current + N;
 return /*return*/current;
}
type Answer = /*call*/Twice<21>;
`

func TestTypefnNativeBindings(t *testing.T) {
	l, ctx, pos := typefnService(t, typefnEditorSource)
	// Query before diagnostics: native binding is independent of VM compilation.
	for _, test := range []struct{ use, decl string }{{"/*return*/", "/*decl*/"}, {"/*write*/", "/*decl*/"}, {"/*inner*/", "/*shadow*/"}} {
		response, err := l.ProvideDefinition(ctx, "file:///main.ts", pos(test.use))
		if err != nil {
			t.Fatal(err)
		}
		if response.Locations == nil || len(*response.Locations) != 1 || (*response.Locations)[0].Range.Start != pos(test.decl) {
			t.Fatalf("definition %s: %#v", test.use, response)
		}
	}
	refs, err := l.ProvideReferences(ctx, &lsproto.ReferenceParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: "file:///main.ts"}, Position: pos("/*return*/"), Context: &lsproto.ReferenceContext{IncludeDeclaration: true}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if refs.Locations == nil || len(*refs.Locations) != 4 {
		t.Fatalf("references: %#v", refs)
	}
	edits, err := l.ProvideRename(ctx, &lsproto.RenameParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: "file:///main.ts"}, Position: pos("/*return*/"), NewName: "total"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if edits.WorkspaceEdit == nil || len((*edits.WorkspaceEdit.Changes)["file:///main.ts"]) != 4 {
		t.Fatalf("rename: %#v", edits)
	}
	if ds := l.program.GetSemanticDiagnostics(ctx, nil); len(ds) != 0 {
		t.Fatalf("checking after queries: %v", ds)
	}
	c, done := l.program.GetTypeCheckerForFile(ctx, l.program.GetSourceFile("/main.ts"))
	defer done()
	function := l.program.GetSourceFile("/main.ts").Statements.Nodes[0]
	binding := function.AsTypeFunctionDeclaration().Body.AsTypeFunctionBlock().Statements.Nodes[0]
	found := false
	for _, symbol := range c.GetSymbolsInScope(binding, ast.SymbolFlagsType) {
		if symbol == binding.Symbol() {
			found = true
		}
	}
	if !found {
		t.Fatal("DSL binding missing from native scope API")
	}
}

func TestTypefnIterationUsesBodyScope(t *testing.T) {
	source := `typefn F<T> { type /*outer*/item = T; type result = []; for (type /*loop*/item of /*subject*/item) { result = [...result, /*inside*/item]; } return /*after*/item; }`
	l, ctx, pos := typefnService(t, source)
	for _, test := range []struct{ use, declaration string }{{"/*subject*/", "/*outer*/"}, {"/*inside*/", "/*loop*/"}, {"/*after*/", "/*outer*/"}} {
		result, err := l.ProvideDefinition(ctx, "file:///main.ts", pos(test.use))
		if err != nil {
			t.Fatal(err)
		}
		if result.Locations == nil || len(*result.Locations) != 1 || (*result.Locations)[0].Range.Start != pos(test.declaration) {
			t.Fatalf("wrong loop scope at %s: %#v", test.use, result)
		}
	}
}

func TestTypefnHover(t *testing.T) {
	source := strings.Replace(typefnEditorSource, "Twice<21>", "Twice</*arg*/21>", 1)
	l, ctx, pos := typefnService(t, source)
	hover, err := l.ProvideHover(ctx, &lsproto.HoverParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: "file:///main.ts"}, Position: pos("/*call*/")})
	if err != nil {
		t.Fatal(err)
	}
	if hover.Hover == nil || !strings.Contains(hover.Hover.Contents.MarkupContent.Value, "typefn Twice<N>: type") || !strings.Contains(hover.Hover.Contents.MarkupContent.Value, "= 42") {
		t.Fatalf("call hover: %#v", hover.Hover)
	}
	local, err := l.ProvideHover(ctx, &lsproto.HoverParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: "file:///main.ts"}, Position: pos("/*return*/")})
	if err != nil {
		t.Fatal(err)
	}
	if local.Hover == nil || !strings.Contains(local.Hover.Contents.MarkupContent.Value, "type current: type (mutable binding)") {
		t.Fatalf("local hover: %#v", local.Hover)
	}
	if ds := l.program.GetSemanticDiagnostics(ctx, nil); len(ds) != 0 {
		t.Fatalf("queries corrupted diagnostics: %v", ds)
	}
}

func TestTypefnOutlineAndFolding(t *testing.T) {
	source := strings.Replace(typefnEditorSource, "Twice<21>", "Twice</*arg*/21>", 1)
	l, ctx, _ := typefnService(t, source)
	symbols, err := l.ProvideDocumentSymbols(ctx, "file:///main.ts")
	if err != nil {
		t.Fatal(err)
	}
	if symbols.DocumentSymbols == nil || len(*symbols.DocumentSymbols) == 0 || (*symbols.DocumentSymbols)[0].Name != "Twice" || (*symbols.DocumentSymbols)[0].Kind != lsproto.SymbolKindFunction {
		t.Fatalf("outline: %#v", symbols)
	}
	folds, err := l.ProvideFoldingRange(ctx, "file:///main.ts")
	if err != nil {
		t.Fatal(err)
	}
	if folds.FoldingRanges == nil || len(*folds.FoldingRanges) == 0 {
		t.Fatal("no DSL folding ranges")
	}
	if ds := l.program.GetSemanticDiagnostics(ctx, nil); len(ds) != 0 {
		t.Fatalf("queries corrupted diagnostics: %v", ds)
	}
}

func BenchmarkTypefnHover(b *testing.B) {
	l, ctx, pos := typefnService(b, typefnEditorSource)
	params := &lsproto.HoverParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: "file:///main.ts"}, Position: pos("/*call*/")}
	if _, err := l.ProvideHover(ctx, params); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		result, err := l.ProvideHover(ctx, params)
		if err != nil || result.Hover == nil {
			b.Fatalf("hover: %v", err)
		}
	}
}
