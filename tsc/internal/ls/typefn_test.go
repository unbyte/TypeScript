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

func TestTypefnCompletionScope(t *testing.T) {
	source := strings.Replace(typefnEditorSource, "Twice<21>", "Twice</*arg*/21>", 1)
	l, ctx, pos := typefnService(t, source)
	completion, err := l.ProvideCompletion(ctx, "file:///main.ts", pos("/*return*/"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if completion.List == nil {
		t.Fatal("no completions")
	}
	names := map[string]bool{}
	for _, item := range completion.List.Items {
		names[item.Label] = true
	}
	for _, name := range []string{"current", "N", "Twice", "return"} {
		if !names[name] {
			t.Errorf("missing completion %s", name)
		}
	}
	if names["saved"] {
		t.Error("completion leaked inner scope")
	}
	if ds := l.program.GetSemanticDiagnostics(ctx, nil); len(ds) != 0 {
		t.Fatalf("queries corrupted diagnostics: %v", ds)
	}
}

func TestTypefnSignature(t *testing.T) {
	source := strings.Replace(typefnEditorSource, "Twice<21>", "Twice</*arg*/21>", 1)
	l, ctx, pos := typefnService(t, source)
	signature, err := l.ProvideSignatureHelp(ctx, "file:///main.ts", pos("/*arg*/"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if signature.SignatureHelp == nil || signature.SignatureHelp.Signatures[0].Label != "typefn Twice<N>: type" {
		t.Fatalf("signature: %#v", signature)
	}
	if ds := l.program.GetSemanticDiagnostics(ctx, nil); len(ds) != 0 {
		t.Fatalf("queries corrupted diagnostics: %v", ds)
	}
}

func TestTypefnSignatureWhileTyping(t *testing.T) {
	for _, test := range []struct {
		source string
		active uint32
	}{
		{`typefn F<A,B> { return A; } typefn G<> { return F</*at*/`, 0},
		{`typefn F<A,B> { return A; } typefn G<> { return F<1, /*at*/`, 1},
		{`typefn F<A,B> { return A; } typefn G<> { return F<[1,2], /*at*/string>; }`, 1},
		{`typefn F<A,B> { return A; } typefn G<> { return F<F<1,2>, /*at*/string>; }`, 1},
		{`typefn F<A,B> { return A; } type R = F<1, /*at*/string>;`, 1},
		{`typefn F<A,B> { return A; } type R = F</*at*/`, 0},
		{`typefn F<A,B> { return A; } typefn G<N> { return F<(N < 2), /*at*/string>; }`, 1},
		{`typefn F<A,B> { return A; } typefn G<N> { return F<(N < 2), [/*at*/1,2]>; }`, 1},
	} {
		t.Run(test.source, func(t *testing.T) {
			l, ctx, pos := typefnService(t, test.source)
			result, err := l.ProvideSignatureHelp(ctx, "file:///main.ts", pos("/*at*/"), nil)
			if err != nil {
				t.Fatal(err)
			}
			help := result.SignatureHelp
			label := "typefn F<A, B>: type"
			if len(l.program.GetSourceFile("/main.ts").Diagnostics()) != 0 {
				label = "typefn F<A, B>"
			}
			if help == nil || help.Signatures[0].Label != label || help.ActiveParameter == nil || help.ActiveParameter.Uinteger == nil || *help.ActiveParameter.Uinteger != test.active {
				t.Fatalf("signature: %#v", help)
			}
		})
	}
}

func TestTypefnNamespaceCompletion(t *testing.T) {
	for _, suffix := range []string{`/*at*/Id<1>; }`, `/*at*/`} {
		source := `namespace Types { export typefn Id<T> { return T; } export type Alias = 1; export const runtime = 1; } typefn F<> { return Types.` + suffix
		l, ctx, pos := typefnService(t, source)
		response, err := l.ProvideCompletion(ctx, "file:///main.ts", pos("/*at*/"), nil)
		if err != nil {
			t.Fatal(err)
		}
		if response.List == nil {
			t.Fatal("missing member completions")
		}
		names := map[string]bool{}
		for _, item := range response.List.Items {
			names[item.Label] = true
		}
		if !names["Id"] || !names["Alias"] || names["runtime"] || names["return"] {
			t.Fatalf("member names: %#v", names)
		}
	}
}

func TestTypefnSemanticTokens(t *testing.T) {
	source := strings.Replace(typefnEditorSource, "Twice<21>", "Twice</*arg*/21>", 1)
	l, ctx, pos := typefnService(t, source)
	tokens, err := l.ProvideSemanticTokens(ctx, "file:///main.ts")
	if err != nil {
		t.Fatal(err)
	}
	if tokens.SemanticTokens == nil {
		t.Fatal("no semantic tokens")
	}
	type classified struct {
		kind tokenType
		mod  uint32
	}
	got := map[lsproto.Position]classified{}
	line, char := uint32(0), uint32(0)
	for i := 0; i < len(tokens.SemanticTokens.Data); i += 5 {
		data := tokens.SemanticTokens.Data[i : i+5]
		line += data[0]
		if data[0] != 0 {
			char = 0
		}
		char += data[1]
		got[lsproto.Position{Line: line, Character: char}] = classified{tokenType(data[3]), data[4]}
	}
	if item := got[lsproto.Position{Line: 0, Character: 7}]; item.kind != tokenTypeKeyword {
		t.Errorf("typefn keyword token: %#v", item)
	}
	if item := got[pos("/*decl*/")]; item.kind != tokenTypeType || item.mod&uint32(tokenModifierDeclaration) == 0 {
		t.Errorf("binding declaration token: %#v", item)
	}
	if item := got[pos("/*write*/")]; item.mod&uint32(tokenModifierModification) == 0 {
		t.Errorf("assignment token: %#v", item)
	}
	if ds := l.program.GetSemanticDiagnostics(ctx, nil); len(ds) != 0 {
		t.Fatalf("queries corrupted diagnostics: %v", ds)
	}
}

func TestTypefnSemanticTokenUTF16Ranges(t *testing.T) {
	source := "// 😀\r\nexport typefn F<N> { type text = \"😀\"; return /*at*/N; }\r\n"
	l, ctx, pos := typefnService(t, source)
	response, err := l.ProvideSemanticTokens(ctx, "file:///main.ts")
	if err != nil {
		t.Fatal(err)
	}
	if response.SemanticTokens == nil {
		t.Fatal("missing tokens")
	}
	line, column, previousEnd := uint32(0), uint32(0), uint32(0)
	found := false
	for i := 0; i < len(response.SemanticTokens.Data); i += 5 {
		token := response.SemanticTokens.Data[i : i+5]
		line += token[0]
		if token[0] != 0 {
			column = 0
			previousEnd = 0
		}
		column += token[1]
		if column < previousEnd {
			t.Fatal("overlapping tokens")
		}
		previousEnd = column + token[2]
		if (lsproto.Position{Line: line, Character: column}) == pos("/*at*/") {
			found = true
			if token[2] != 1 || tokenType(token[3]) != tokenTypeTypeParameter {
				t.Fatalf("wrong parameter token: %v", token)
			}
		}
	}
	if !found {
		t.Fatal("parameter token not at UTF-16 position after emoji")
	}
}

func TestTypefnResultHints(t *testing.T) {
	source := strings.Replace(typefnEditorSource, "Twice<21>", "Twice</*arg*/21>", 1)
	l, ctx, _ := typefnService(t, source)
	l.activeConfig.InlayHintsPreferences.IncludeInlayTypefnResultHints = core.TSTrue
	hints, err := l.ProvideInlayHint(ctx, &lsproto.InlayHintParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: "file:///main.ts"}, Range: lsproto.Range{End: lsproto.Position{Line: 20}}})
	if err != nil {
		t.Fatal(err)
	}
	if hints.InlayHints == nil || len(*hints.InlayHints) != 1 || (*hints.InlayHints)[0].Label.String == nil || *(*hints.InlayHints)[0].Label.String != " = 42" {
		t.Fatalf("result hints: %#v", hints)
	}
	if ds := l.program.GetSemanticDiagnostics(ctx, nil); len(ds) != 0 {
		t.Fatalf("queries corrupted diagnostics: %v", ds)
	}
}

func TestTypefnIncompleteBufferQueries(t *testing.T) {
	for _, source := range []string{
		`typefn F<T> { type local = T; return /*at*/`,
		`typefn F<T> { return T; } type R = F</*at*/`,
		`typefn F<T> { return [T]; } typefn G<> { return F</*at*/`,
		`typefn F<T> { type local = ; return /*at*/local; } type R = F<1>;`,
		"typefn F<T> { return " + strings.Repeat("[", 300) + "/*at*/T",
	} {
		t.Run(source, func(t *testing.T) {
			l, ctx, pos := typefnService(t, source)
			if _, err := l.ProvideHover(ctx, &lsproto.HoverParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: "file:///main.ts"}, Position: pos("/*at*/")}); err != nil {
				t.Fatal(err)
			}
			if _, err := l.ProvideCompletion(ctx, "file:///main.ts", pos("/*at*/"), nil); err != nil {
				t.Fatal(err)
			}
			if _, err := l.ProvideSignatureHelp(ctx, "file:///main.ts", pos("/*at*/"), nil); err != nil {
				t.Fatal(err)
			}
			if _, err := l.ProvideSemanticTokens(ctx, "file:///main.ts"); err != nil {
				t.Fatal(err)
			}
			l.activeConfig.InlayHintsPreferences.IncludeInlayTypefnResultHints = core.TSTrue
			hints, err := l.ProvideInlayHint(ctx, &lsproto.InlayHintParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: "file:///main.ts"}, Range: lsproto.Range{End: lsproto.Position{Line: 20}}})
			if err != nil {
				t.Fatal(err)
			}
			if hints.InlayHints != nil && len(*hints.InlayHints) != 0 {
				t.Fatal("evaluated recovered syntax")
			}
			// Editors also request semantic diagnostics on incomplete snapshots.
			l.program.GetSemanticDiagnostics(ctx, nil)
		})
	}
}

func TestTypefnPatternCaptureScope(t *testing.T) {
	source := "typefn F<T> { type /*outer*/item = \"/\"; match (/*subject*/item) { `${/*capture*/item}${=/*pin*/item}${tail}` if (/*guard*/item == \"a\") => { return /*body*/item; } _ => { return /*fallback*/item; } } return /*after*/item; }"
	l, ctx, pos := typefnService(t, source)
	for _, test := range []struct{ use, decl string }{{"/*subject*/", "/*outer*/"}, {"/*pin*/", "/*outer*/"}, {"/*guard*/", "/*capture*/"}, {"/*body*/", "/*capture*/"}, {"/*fallback*/", "/*outer*/"}, {"/*after*/", "/*outer*/"}} {
		response, err := l.ProvideDefinition(ctx, "file:///main.ts", pos(test.use))
		if err != nil {
			t.Fatal(err)
		}
		if response.Locations == nil || len(*response.Locations) != 1 || (*response.Locations)[0].Range.Start != pos(test.decl) {
			t.Fatalf("definition %s: %#v", test.use, response)
		}
	}
	hover, err := l.ProvideHover(ctx, &lsproto.HoverParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: "file:///main.ts"}, Position: pos("/*body*/")})
	if err != nil {
		t.Fatal(err)
	}
	if hover.Hover == nil || !strings.Contains(hover.Hover.Contents.MarkupContent.Value, "pattern binding") {
		t.Fatalf("capture hover: %#v", hover.Hover)
	}
	edits, err := l.ProvideRename(ctx, &lsproto.RenameParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: "file:///main.ts"}, Position: pos("/*body*/"), NewName: "captured"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if edits.WorkspaceEdit == nil || len((*edits.WorkspaceEdit.Changes)["file:///main.ts"]) != 3 {
		t.Fatalf("capture rename: %#v", edits)
	}
	// The capture 'tail' has no outer namesake: completion must exclude it in pins
	// and expose it in guards/bodies, without needing VM evaluation.
	for _, test := range []struct {
		marker  string
		visible bool
	}{{"/*pin*/", false}, {"/*guard*/", true}, {"/*body*/", true}, {"/*fallback*/", false}} {
		response, err := l.ProvideCompletion(ctx, "file:///main.ts", pos(test.marker), nil)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		if response.List != nil {
			for _, item := range response.List.Items {
				if item.Label == "tail" {
					found = true
				}
			}
		}
		if found != test.visible {
			t.Fatalf("tail visibility at %s: %v", test.marker, found)
		}
	}
}

func TestTypefnArmBodyHasNestedScope(t *testing.T) {
	source := `typefn F<T> { type /*outer*/label = 1; match (T) { /*capture*/item if (/*guard*/label == 1) => { type /*local*/label = 2; type /*shadow*/item = 3; return [/*body*/label, /*read*/item]; } _ => { return never; } } }`
	l, ctx, pos := typefnService(t, source)
	for _, test := range []struct{ use, decl string }{{"/*guard*/", "/*outer*/"}, {"/*body*/", "/*local*/"}, {"/*read*/", "/*shadow*/"}} {
		response, err := l.ProvideDefinition(ctx, "file:///main.ts", pos(test.use))
		if err != nil {
			t.Fatal(err)
		}
		if response.Locations == nil || len(*response.Locations) != 1 || (*response.Locations)[0].Range.Start != pos(test.decl) {
			t.Fatalf("arm body scope %s: %#v", test.use, response)
		}
	}
}

func TestTypefnAnnotatedResultHover(t *testing.T) {
	for _, annotation := range []string{"type", "typefn<>", "typefn<_, _>"} {
		source := "typefn /*name*/Factory<>: " + annotation + " { return string; }"
		l, ctx, pos := typefnService(t, source)
		hover, err := l.ProvideHover(ctx, &lsproto.HoverParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: "file:///main.ts"}, Position: pos("/*name*/")})
		if err != nil {
			t.Fatal(err)
		}
		if hover.Hover == nil || !strings.Contains(hover.Hover.Contents.MarkupContent.Value, "Factory<>: "+annotation) {
			t.Fatalf("lost result kind %s in hover: %#v", annotation, hover.Hover)
		}
	}
}

func TestTypefnLocalHelperSymbols(t *testing.T) {
	source := `typefn Outer<T> {
 type /*value*/saved = T;
 typefn /*helper*/Helper<U>: type {if(U == 0) {return /*capture*/saved;} return /*self*/Helper<0>;}
 return /*call*/Helper<1>;
 }`
	l, ctx, pos := typefnService(t, source)
	for _, test := range []struct{ use, declaration string }{{"/*capture*/", "/*value*/"}, {"/*call*/", "/*helper*/"}, {"/*self*/", "/*helper*/"}} {
		response, err := l.ProvideDefinition(ctx, "file:///main.ts", pos(test.use))
		if err != nil {
			t.Fatal(err)
		}
		if response.Locations == nil || len(*response.Locations) != 1 || (*response.Locations)[0].Range.Start != pos(test.declaration) {
			t.Fatalf("helper definition at %s: %#v", test.use, response)
		}
	}
	hover, err := l.ProvideHover(ctx, &lsproto.HoverParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: "file:///main.ts"}, Position: pos("/*capture*/")})
	if err != nil {
		t.Fatal(err)
	}
	if hover.Hover == nil || !strings.Contains(hover.Hover.Contents.MarkupContent.Value, "capture, readonly") {
		t.Fatalf("capture hover: %#v", hover.Hover)
	}
	edits, err := l.ProvideRename(ctx, &lsproto.RenameParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: "file:///main.ts"}, Position: pos("/*call*/"), NewName: "Renamed"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if edits.WorkspaceEdit == nil || len((*edits.WorkspaceEdit.Changes)["file:///main.ts"]) != 3 {
		t.Fatalf("helper rename: %#v", edits)
	}
	if ds := l.program.GetSemanticDiagnostics(ctx, nil); len(ds) != 0 {
		t.Fatal(ds)
	}
}

func TestTypefnLocalHelperCompletionOrder(t *testing.T) {
	source := `typefn Outer<> {type first = /*before*/0; typefn Later<> {return /*self*/1;} return /*after*/Later<>;}`
	l, ctx, _ := typefnService(t, source)
	file := l.program.GetSourceFile("/main.ts")
	for _, test := range []struct {
		marker string
		want   bool
	}{{"/*before*/", false}, {"/*self*/", true}, {"/*after*/", true}} {
		items := l.typefnCompletions(ctx, file, strings.Index(source, test.marker)+len(test.marker))
		found := false
		if items != nil {
			for _, item := range items.Items {
				found = found || item.Label == "Later"
			}
		}
		if found != test.want {
			t.Fatalf("Later completion at %s: got %v want %v", test.marker, found, test.want)
		}
	}
}

func TestTypefnLambdaScopesAndDisplay(t *testing.T) {
	source := `typefn F<T> {type /*decl*/saved=T; type fn=typefn<U> {if(/*capture*/saved == "yes") {return U;} return false;}; return {value:fn<1>};} type /*pending*/Pending<T>=F<T>; type Value=Pending<"yes">;`
	l, ctx, pos := typefnService(t, source)
	result, err := l.ProvideDefinition(ctx, "file:///main.ts", pos("/*capture*/"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Locations == nil || len(*result.Locations) != 1 || (*result.Locations)[0].Range.Start != pos("/*decl*/") {
		t.Fatalf("wrong captured symbol: %#v", result)
	}
	for _, marker := range []string{"/*capture*/", "/*pending*/"} {
		hover, err := l.ProvideHover(ctx, &lsproto.HoverParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: "file:///main.ts"}, Position: pos(marker)})
		if err != nil || hover.Hover == nil {
			t.Fatalf("hover %s: %#v %v", marker, hover, err)
		}
		if marker == "/*capture*/" && !strings.Contains(hover.Hover.Contents.MarkupContent.Value, "capture, readonly") {
			t.Fatal(hover.Hover.Contents.MarkupContent.Value)
		}
	}
	if ds := l.program.GetSemanticDiagnostics(ctx, nil); len(ds) != 0 {
		t.Fatal(ds)
	}
}

func TestTypefnConstructorBindingHover(t *testing.T) {
	source := `typefn Box<T> {return [T];} typefn F<T> {type /*decl*/Constructor=Box; return /*use*/Constructor<T>;}`
	l, ctx, pos := typefnService(t, source)
	for _, marker := range []string{"/*decl*/", "/*use*/"} {
		result, err := l.ProvideHover(ctx, &lsproto.HoverParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: "file:///main.ts"}, Position: pos(marker)})
		if err != nil || result.Hover == nil || !strings.Contains(result.Hover.Contents.MarkupContent.Value, "type Constructor: typefn<_>") {
			t.Fatalf("constructor binding hover: %#v %v", result, err)
		}
	}
	if ds := l.program.GetSemanticDiagnostics(ctx, nil); len(ds) != 0 {
		t.Fatal(ds)
	}
}

func TestTypefnFactoryKindHover(t *testing.T) {
	source := `typefn /*factory*/Factory<T> {return typefn<U> {return [T,U];};} typefn Use<T> {type /*binding*/F=Factory<T>; return F<1>;} type Result=Use<123>;`
	l, ctx, pos := typefnService(t, source)
	for _, test := range []struct{ marker, want string }{{"/*factory*/", "typefn Factory<T>: typefn<_>"}, {"/*binding*/", "type F: typefn<_>"}} {
		result, err := l.ProvideHover(ctx, &lsproto.HoverParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: "file:///main.ts"}, Position: pos(test.marker)})
		if err != nil || result.Hover == nil || !strings.Contains(result.Hover.Contents.MarkupContent.Value, test.want) {
			t.Fatalf("factory hover: %#v %v", result, err)
		}
	}
	if ds := l.program.GetSemanticDiagnostics(ctx, nil); len(ds) != 0 {
		t.Fatal(ds)
	}
}
