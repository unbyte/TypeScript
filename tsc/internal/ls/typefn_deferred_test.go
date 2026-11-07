package ls

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/lsp/lsproto"
)

func TestTypefnDeferredResultsInEditor(t *testing.T) {
	source := `typefn Twice<N> {return N*2;}
 typefn Discard<N> {type unused=N*2;return string;}
 type Pending<N>=/*pending*/Twice<N>;
 type Checked<N>=/*guarded*/Discard<N>;
 type Ready=/*ready*/Twice<21>;
 type Final=Pending<21>;
 const correct:Final=42;`
	l, ctx, pos := typefnService(t, source)
	for _, marker := range []string{"/*pending*/", "/*guarded*/", "/*ready*/"} {
		hover, err := l.ProvideHover(ctx, &lsproto.HoverParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: "file:///main.ts"}, Position: pos(marker)})
		if err != nil || hover.Hover == nil {
			t.Fatalf("hover %s: %v %#v", marker, err, hover.Hover)
		}
		hasResult := strings.Contains(hover.Hover.Contents.MarkupContent.Value, " = ")
		if hasResult != (marker == "/*ready*/") {
			t.Fatalf("misleading result %s: %s", marker, hover.Hover.Contents.MarkupContent.Value)
		}
	}
	l.activeConfig.InlayHintsPreferences.IncludeInlayTypefnResultHints = core.TSTrue
	hints, err := l.ProvideInlayHint(ctx, &lsproto.InlayHintParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: "file:///main.ts"}, Range: lsproto.Range{End: lsproto.Position{Line: 10}}})
	if err != nil || hints.InlayHints == nil || len(*hints.InlayHints) != 1 {
		t.Fatalf("pending results leaked into hints: %#v %v", hints, err)
	}
	if *(*hints.InlayHints)[0].Label.String != " = 42" {
		t.Fatal("wrong completed result")
	}
	if ds := l.program.GetSemanticDiagnostics(ctx, nil); len(ds) != 0 {
		t.Fatalf("generic definition failed: %v", ds)
	}
	c, done := l.program.GetTypeChecker(ctx)
	defer done()
	file := l.program.GetSourceFile("/main.ts")
	pendingType := c.GetTypeAtLocation(file.Statements.Nodes[2].Name())
	if display := c.TypeToString(pendingType); !strings.Contains(display, "N * 2") || strings.Contains(display, "__typefn") {
		t.Fatalf("opaque pending display: %s", display)
	}
	guardedType := c.GetTypeAtLocation(file.Statements.Nodes[3].Name())
	if display := c.TypeToString(guardedType); !strings.Contains(display, "pending validation") {
		t.Fatalf("hidden validation: %s", display)
	}
	application := file.Statements.Nodes[2].AsTypeAliasDeclaration().Type
	if c.GetTypeFunctionApplicationResult(ctx, application) != nil {
		t.Fatal("pending application completed")
	}
	count := c.TypeCount
	for i := 0; i < 20; i++ {
		if c.GetTypeFunctionApplicationResult(ctx, application) != nil {
			t.Fatal("pending application completed")
		}
	}
	if c.TypeCount != count {
		t.Fatal("unchanged query created new deferred terms")
	}
}
