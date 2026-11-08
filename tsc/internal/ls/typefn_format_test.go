package ls

import (
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/ls/lsutil"
	"github.com/microsoft/TypeScript/tsc/internal/lsp/lsproto"
	"github.com/microsoft/TypeScript/tsc/internal/parser"
	"github.com/microsoft/TypeScript/tsc/internal/spanmap"
)

func TestTypefnFormattingPreservesSyntax(t *testing.T) {
	for _, source := range []string{
		"typefn F<N>{\n// preserve this comment\ntype value=N+1;\nif(value>2){return value;}\nreturn 0;\n}\n",
		`typefn F<N>{type x=- -N;type y=!(N is never);return [x,y,N>=2];}`,
		`typefn F<N>{type result=[];for(type item of N){if(item is string){result=[item,...result];}else{continue;}}return result;}`,
		`typefn F<N>{return {readonly key?:N;value:F<[N,1]>[0]};}`,
		`typefn F<N>{return type((input:N)=>Promise<N>);}`,
	} {
		t.Run(source, func(t *testing.T) {
			l, ctx, _ := typefnService(t, source)
			l.activeConfig.FormatCodeSettings = lsutil.GetDefaultFormatCodeSettings()
			l.activeConfig.FormatCodeSettings.Semicolons = lsutil.SemicolonPreferenceRemove
			response, err := l.ProvideFormatDocument(ctx, "file:///main.ts", &lsproto.FormattingOptions{TabSize: 2, InsertSpaces: true})
			if err != nil {
				t.Fatal(err)
			}
			formatted := source
			if response.TextEdits != nil {
				changes := []core.TextChange{}
				for _, edit := range *response.TextEdits {
					mapped := l.converters.FromLSPRangeForSourceFile(l.program.GetSourceFile("/main.ts"), edit.Range, spanmap.FeatureFormatting)
					if len(mapped) != 1 {
						t.Fatal("unmapped format edit")
					}
					changes = append(changes, core.TextChange{TextRange: mapped[0].Span, NewText: edit.NewText})
				}
				slices.SortFunc(changes, func(a, b core.TextChange) int { return b.Pos() - a.Pos() })
				for _, change := range changes {
					formatted = formatted[:change.Pos()] + change.NewText + formatted[change.End():]
				}
			}
			file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/main.ts", PathKey: "/main.ts", TypefnVersion: 1}, formatted, core.ScriptKindTS)
			if ds := file.Diagnostics(); len(ds) != 0 {
				t.Fatalf("formatter broke DSL: %s\n%v", formatted, ds)
			}
			if strings.Contains(source, "// preserve") && (!strings.Contains(formatted, "// preserve this comment") || !strings.Contains(formatted, "\n  type value = N + 1;")) {
				t.Fatalf("comment/indent/spacing: %s", formatted)
			}
			l2, ctx2, _ := typefnService(t, formatted)
			l2.activeConfig = l.activeConfig
			again, err := l2.ProvideFormatDocument(ctx2, "file:///main.ts", &lsproto.FormattingOptions{TabSize: 2, InsertSpaces: true})
			if err != nil {
				t.Fatal(err)
			}
			if again.TextEdits != nil && len(*again.TextEdits) > 0 {
				t.Fatalf("formatting is not idempotent: %#v", again)
			}
		})
	}
}
