package lsp_test

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/ls/lsutil"
	"github.com/microsoft/TypeScript/tsc/internal/lsp/lsproto"
	"gotest.tools/v3/assert"
)

// Exercise serialized LSP requests and preference routing, in addition to the
// language-service and project-snapshot tests.
func TestTypefnLSPRequests(t *testing.T) {
	const uri = "file:///home/projects/main.ts"
	const source = "typefn Twice<N> { return N * 2; }\ntype R = Twice<21>;\n"
	client := initCompletionClient(t, map[string]string{
		"/home/projects/tsconfig.json": `{"compilerOptions":{"typefn":{"version":1},"strict":true,"target":"es2022","ignoreDeprecations":"6.0"},"files":["main.ts"]}`,
		"/home/projects/main.ts":       source,
	}, &lsutil.UserPreferences{InlayHintsPreferences: lsutil.InlayHintsPreferences{IncludeInlayTypefnResultHints: core.TSTrue}})
	client.SendNotification(t, lsproto.TextDocumentDidOpenInfo, &lsproto.DidOpenTextDocumentParams{
		TextDocument: &lsproto.TextDocumentItem{Uri: uri, LanguageId: "typescript", Version: 1, Text: source},
	})
	hoverMsg, hover, ok := client.SendRequest(t, lsproto.TextDocumentHoverInfo, &lsproto.HoverParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uri}, Position: lsproto.Position{Line: 1, Character: 9},
	})
	assert.Assert(t, ok && hoverMsg.AsResponse().Error == nil)
	assert.Assert(t, hover.Hover != nil)
	assert.Assert(t, strings.Contains(hover.Hover.Contents.MarkupContent.Value, "Twice<21> = 42"))
	definitionMsg, definition, ok := client.SendRequest(t, lsproto.TextDocumentDefinitionInfo, &lsproto.DefinitionParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uri}, Position: lsproto.Position{Line: 1, Character: 9},
	})
	assert.Assert(t, ok && definitionMsg.AsResponse().Error == nil)
	assert.Assert(t, definition.Locations != nil && len(*definition.Locations) == 1)
	assert.Equal(t, (*definition.Locations)[0].Range.Start.Line, uint32(0))
	hintsMsg, hints, ok := client.SendRequest(t, lsproto.TextDocumentInlayHintInfo, &lsproto.InlayHintParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uri}, Range: lsproto.Range{End: lsproto.Position{Line: 2}},
	})
	assert.Assert(t, ok && hintsMsg.AsResponse().Error == nil)
	assert.Assert(t, hints.InlayHints != nil && len(*hints.InlayHints) == 1, "typefn preference did not reach inlay provider")
	assert.Equal(t, *(*hints.InlayHints)[0].Label.String, " = 42")
}
