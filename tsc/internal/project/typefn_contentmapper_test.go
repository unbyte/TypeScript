package project_test

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/bundled"
	"github.com/microsoft/TypeScript/tsc/internal/lsp/lsproto"
	"github.com/microsoft/TypeScript/tsc/internal/project"
	"github.com/microsoft/TypeScript/tsc/internal/testutil/contentmappertest"
	"github.com/microsoft/TypeScript/tsc/internal/testutil/projecttestutil"
	"gotest.tools/v3/assert"
)

// The canonical file may contain no TypeScript from the user's document. Queries
// must find the DSL in supplemental projections and return original coordinates.
func TestTypefnContentMappedProjections(t *testing.T) {
	t.Parallel()
	for _, mapper := range []string{contentmappertest.SupplementalMapper, contentmappertest.DuplicateProjectionMapper} {
		t.Run(mapper, func(t *testing.T) {
			t.Parallel()
			const uri = "file:///home/project/main.box"
			const source = "export typefn Twice<N> { return N * 2; } export type Answer = Twice<21>;"
			files := map[string]any{
				"/home/project/tsconfig.json":                    `{"compilerOptions":{"typefn":{"version":1},"strict":true,"noLib":true},"contentMappers":[{"package":"mapper","extensions":[".box"]}],"files":["main.box"]}`,
				"/home/project/node_modules/mapper/package.json": contentmappertest.PackageJSON(mapper),
				"/home/project/main.box":                         source,
			}
			init, _ := projecttestutil.GetSessionInitOptions(files, &project.SessionOptions{
				CurrentDirectory: "/home/project", DefaultLibraryPath: bundled.LibPath(),
				PositionEncoding: lsproto.PositionEncodingKindUTF8, RunExternalCode: true,
			}, nil)
			init.Spawner = contentmappertest.NewSpawner()
			session := project.NewSession(init)
			defer session.Close()
			caps := &lsproto.ResolvedClientCapabilities{}
			caps.TextDocument.Hover.ContentFormat = []lsproto.MarkupKind{lsproto.MarkupKindMarkdown}
			ctx := lsproto.WithClientCapabilities(t.Context(), caps)
			session.DidOpenFile(ctx, uri, 1, source, lsproto.LanguageKind("box"))
			service, err := session.GetLanguageService(ctx, uri)
			assert.NilError(t, err)
			position := lsproto.Position{Character: uint32(strings.LastIndex(source, "Twice<"))}
			hover, err := service.ProvideHover(ctx, &lsproto.HoverParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: uri}, Position: position})
			assert.NilError(t, err)
			assert.Assert(t, hover.Hover != nil)
			assert.Assert(t, strings.Contains(hover.Hover.Contents.MarkupContent.Value, "Twice<21> = 42"))
			assert.Assert(t, hover.Hover.Range != nil)
			assert.Equal(t, hover.Hover.Range.Start, position)
			position.Character += uint32(len("Twice<"))
			signature, err := service.ProvideSignatureHelp(ctx, uri, position, nil)
			assert.NilError(t, err)
			assert.Assert(t, signature.SignatureHelp != nil)
			assert.Equal(t, signature.SignatureHelp.Signatures[0].Label, "typefn Twice<N>: type")
		})
	}
}
