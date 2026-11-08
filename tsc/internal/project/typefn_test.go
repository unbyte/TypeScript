package project

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/bundled"
	"github.com/microsoft/TypeScript/tsc/internal/compiler"
	"github.com/microsoft/TypeScript/tsc/internal/lsp/lsproto"
	"github.com/microsoft/TypeScript/tsc/internal/tspath"
	"github.com/microsoft/TypeScript/tsc/internal/vfs/vfstest"
)

// Exercise the real editor session, source parse cache and checker pool: editing
// an imported definition must invalidate computations in an unchanged consumer.
func TestTypefnProjectEditsAndProfileCache(t *testing.T) {
	t.Parallel()
	caps := &lsproto.ResolvedClientCapabilities{}
	caps.TextDocument.Hover.ContentFormat = []lsproto.MarkupKind{lsproto.MarkupKindMarkdown}
	ctx := lsproto.WithClientCapabilities(t.Context(), caps)
	const mainURI = "file:///p/main.ts"
	const helperURI = "file:///p/helper.ts"
	const enabled = `{"compilerOptions":{"strict":true,"noEmit":true,"target":"es2022","ignoreDeprecations":"6.0","typefn":{"version":1}},"files":["main.ts","helper.ts"]}`
	const disabled = `{"compilerOptions":{"strict":true,"noEmit":true,"target":"es2022","ignoreDeprecations":"6.0"},"files":["main.ts","helper.ts"]}`
	const main = `import type { Double } from "./helper"; export const result: Double<21> = 42;`
	const helper = `export typefn Double<T> { return T * 2; }`
	fs := bundled.WrapFS(vfstest.FromMap(map[string]any{"/p/tsconfig.json": enabled, "/p/main.ts": main, "/p/helper.ts": helper}, tspath.CaseSensitive))
	session := NewSession(&SessionInit{Client: noopClient{}, BackgroundCtx: ctx, Options: &SessionOptions{CurrentDirectory: "/", DefaultLibraryPath: bundled.LibPath(), PositionEncoding: lsproto.PositionEncodingKindUTF8}, FS: fs})
	session.DidOpenFile(ctx, mainURI, 1, main, lsproto.LanguageKindTypeScript)
	session.DidOpenFile(ctx, helperURI, 1, helper, lsproto.LanguageKindTypeScript)
	program := func() *compiler.Program {
		t.Helper()
		if _, err := session.GetLanguageService(ctx, mainURI); err != nil {
			t.Fatal(err)
		}
		return session.Snapshot().ProjectCollection.ConfiguredProject(tspath.PathKey("/p/tsconfig.json")).GetProgram()
	}
	check := func(wantError bool) {
		t.Helper()
		// Request hover before diagnostics, as an editor can do on every edit.
		service, err := session.GetLanguageService(ctx, mainURI)
		if err != nil {
			t.Fatal(err)
		}
		position := lsproto.Position{Character: uint32(strings.LastIndex(main, "Double<"))}
		hover, err := service.ProvideHover(ctx, &lsproto.HoverParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: mainURI}, Position: position})
		if err != nil {
			t.Fatal(err)
		}
		expected := "Double<21> = 42"
		if wantError {
			expected = "Double<21> = 63"
		}
		if hover.Hover == nil || !strings.Contains(hover.Hover.Contents.MarkupContent.Value, expected) {
			t.Fatalf("stale or missing hover, expected %s: %#v", expected, hover.Hover)
		}
		definition, err := service.ProvideDefinition(ctx, mainURI, position)
		if err != nil {
			t.Fatal(err)
		}
		if definition.Locations == nil || len(*definition.Locations) != 1 || (*definition.Locations)[0].Uri != helperURI {
			t.Fatalf("import definition: %#v", definition)
		}
		refs, err := service.ProvideReferences(ctx, &lsproto.ReferenceParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: mainURI}, Position: position, Context: &lsproto.ReferenceContext{IncludeDeclaration: true}}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if refs.Locations == nil || len(*refs.Locations) != 3 {
			t.Fatalf("import references: %#v", refs)
		}
		rename, err := service.ProvideRename(ctx, &lsproto.RenameParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: helperURI}, Position: lsproto.Position{Character: uint32(strings.Index(helper, "Double"))}, NewName: "Scale"}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if rename.WorkspaceEdit == nil || rename.WorkspaceEdit.Changes == nil {
			t.Fatalf("import rename: %#v", rename)
		}
		changes := *rename.WorkspaceEdit.Changes
		if len(changes[mainURI]) != 2 || len(changes[helperURI]) != 1 {
			t.Fatalf("rename edits: %#v", changes)
		}
		p := program()
		if ds := p.GetSyntacticDiagnostics(ctx, nil); len(ds) != 0 {
			t.Fatalf("syntax: %v", ds)
		}
		ds := p.GetSemanticDiagnostics(ctx, nil)
		if !wantError && len(ds) != 0 {
			t.Fatalf("unexpected errors: %v", ds)
		}
		if wantError && (len(ds) != 1 || ds[0].Code() != 2322 || !strings.Contains(ds[0].String(), "63")) {
			t.Fatalf("expected updated literal assignment error, got %v", ds)
		}
	}
	check(false)
	change := func(version int32, text string) {
		session.DidChangeFile(ctx, helperURI, version, []lsproto.TextDocumentContentChangePartialOrWholeDocument{{WholeDocument: &lsproto.TextDocumentContentChangeWholeDocument{Text: text}}})
	}
	change(2, strings.Replace(helper, "* 2", "* 3", 1))
	check(true)
	change(3, helper)
	check(false)
	change(4, "export typefn Double<T> { return T * ; }")
	service, err := session.GetLanguageService(ctx, mainURI)
	if err != nil {
		t.Fatal(err)
	}
	malformedHover, err := service.ProvideHover(ctx, &lsproto.HoverParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: mainURI}, Position: lsproto.Position{Character: uint32(strings.LastIndex(main, "Double<"))}})
	if err != nil {
		t.Fatal(err)
	}
	if malformedHover.Hover != nil && strings.Contains(malformedHover.Hover.Contents.MarkupContent.Value, " = ") {
		t.Fatal("evaluated recovered imported body")
	}
	change(5, helper)
	check(false)
	setConfig := func(text string) {
		t.Helper()
		if err := fs.WriteFile("/p/tsconfig.json", text); err != nil {
			t.Fatal(err)
		}
		session.DidChangeWatchedFiles(ctx, []*lsproto.FileEvent{{Uri: "file:///p/tsconfig.json", Type: lsproto.FileChangeTypeChanged}})
	}
	setConfig(disabled)
	p := program()
	file := p.GetSourceFile("/p/helper.ts")
	if file.ParseOptions().TypefnVersion != 0 || ast.IsTypeFunctionDeclaration(file.Statements.Nodes[0]) || len(p.GetSyntacticDiagnostics(ctx, file)) == 0 {
		t.Fatal("disabled profile reused DSL parse")
	}
	setConfig(enabled)
	check(false)
	if file := program().GetSourceFile("/p/helper.ts"); file.ParseOptions().TypefnVersion != 1 || !ast.IsTypeFunctionDeclaration(file.Statements.Nodes[0]) {
		t.Fatal("enabled profile reused native parse")
	}
	originalFile := program().GetSourceFile("/p/helper.ts")
	setConfig(strings.Replace(enabled, `"version":1`, `"version":1,"limits":{"syntaxDepth":1}`, 1))
	limitedFile := program().GetSourceFile("/p/helper.ts")
	if limitedFile == originalFile || limitedFile.ParseOptions().TypefnSyntaxDepth != 1 {
		t.Fatal("changed syntax limit reused a source file parsed under another profile")
	}
	setConfig(enabled)
	check(false)
	session.WaitForBackgroundTasks()
}

// Deterministic cancellation during VM execution, independent of machine speed.
type typefnCancelContext struct {
	context.Context
	cancel    context.CancelFunc
	remaining atomic.Int32
}

func (c *typefnCancelContext) Err() error {
	if c.remaining.Add(-1) <= 0 {
		c.cancel()
	}
	return c.Context.Err()
}

func TestTypefnCanceledResultDoesNotPoisonCheckerPool(t *testing.T) {
	for _, deferred := range []bool{false, true} {
		name := "direct"
		if deferred {
			name = "replay"
		}
		t.Run(name, func(t *testing.T) {
			source := `typefn Count<N> { type i = 0; while (i < N) { i = i + 1; } return i; } `
			resultName := "Count"
			if deferred {
				source += `type Deferred<N>=Count<N>; typefn Run<N>{return Deferred<N>;} `
				resultName = "Run"
			}
			source += "type R = /*at*/" + resultName + "<100>;"
			caps := &lsproto.ResolvedClientCapabilities{}
			caps.TextDocument.Hover.ContentFormat = []lsproto.MarkupKind{lsproto.MarkupKindMarkdown}
			ctx := lsproto.WithClientCapabilities(t.Context(), caps)
			fs := bundled.WrapFS(vfstest.FromMap(map[string]any{
				"/main.ts":       source,
				"/tsconfig.json": `{"compilerOptions":{"typefn":{"version":1},"strict":true,"target":"es2022","ignoreDeprecations":"6.0"},"files":["main.ts"]}`,
			}, tspath.CaseSensitive))
			session := NewSession(&SessionInit{Client: noopClient{}, BackgroundCtx: ctx, Options: &SessionOptions{CurrentDirectory: "/", DefaultLibraryPath: bundled.LibPath(), PositionEncoding: lsproto.PositionEncodingKindUTF8}, FS: fs})
			session.DidOpenFile(ctx, "file:///main.ts", 1, source, lsproto.LanguageKindTypeScript)
			l, err := session.GetLanguageService(ctx, "file:///main.ts")
			if err != nil {
				t.Fatal(err)
			}
			program := l.GetProgram()
			file := program.GetSourceFile("/main.ts")
			c, done := program.GetTypeCheckerForFile(ctx, file)
			canceled, cancel := context.WithCancel(ctx)
			defer cancel()
			queryCtx := &typefnCancelContext{Context: canceled, cancel: cancel}
			queryCtx.remaining.Store(80)
			application := file.Statements.Nodes[len(file.Statements.Nodes)-1].AsTypeAliasDeclaration().Type
			result := c.GetTypeFunctionApplicationResult(queryCtx, application)
			if result != nil || !c.WasCanceled() {
				done()
				t.Fatal("canceled result was retained")
			}
			done()
			hover, err := l.ProvideHover(ctx, &lsproto.HoverParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: "file:///main.ts"}, Position: lsproto.Position{Character: uint32(strings.Index(source, "/*at*/") + len("/*at*/"))}})
			if err != nil {
				t.Fatal(err)
			}
			if hover.Hover == nil || !strings.Contains(hover.Hover.Contents.MarkupContent.Value, resultName+"<100> = 100") {
				t.Fatalf("cancellation poisoned later query: %#v", hover.Hover)
			}
			if ds := program.GetSemanticDiagnostics(ctx, nil); len(ds) != 0 {
				t.Fatalf("cancellation leaked diagnostics: %v", ds)
			}
		})
	}
}
