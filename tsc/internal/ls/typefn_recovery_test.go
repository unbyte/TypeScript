package ls

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/lsp/lsproto"
)

func TestTypefnRecoveryTypesDoNotPublishResults(t *testing.T) {
	for _, source := range []string{
		`typefn F<T> { type unused = T["missing"]; return 42; } type R = /*at*/F<{key: 1}>;`,
		`typefn F<T> { return any | T["missing"]; } type R = /*at*/F<{key: 1}>;`,
		`typefn F<T> { return 42; } type R = /*at*/F<MissingType>;`,
		`typefn F<T extends MissingType> { return 42; } type R = /*at*/F<1>;`,
	} {
		t.Run(source, func(t *testing.T) {
			l, ctx, pos := typefnService(t, source)
			hover, err := l.ProvideHover(ctx, &lsproto.HoverParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: "file:///main.ts"}, Position: pos("/*at*/")})
			if err != nil {
				t.Fatal(err)
			}
			if hover.Hover != nil && strings.Contains(hover.Hover.Contents.MarkupContent.Value, " = ") {
				t.Fatalf("published a successful result from recovery: %s", hover.Hover.Contents.MarkupContent.Value)
			}
			l.activeConfig.InlayHintsPreferences.IncludeInlayTypefnResultHints = core.TSTrue
			hints, err := l.ProvideInlayHint(ctx, &lsproto.InlayHintParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: "file:///main.ts"}, Range: lsproto.Range{End: lsproto.Position{Line: 5}}})
			if err != nil {
				t.Fatal(err)
			}
			if hints.InlayHints != nil && len(*hints.InlayHints) != 0 {
				t.Fatal("published recovery result hint")
			}
			if ds := l.program.GetSemanticDiagnostics(ctx, nil); len(ds) == 0 {
				t.Fatal("native diagnostic was lost")
			}
		})
	}
}
