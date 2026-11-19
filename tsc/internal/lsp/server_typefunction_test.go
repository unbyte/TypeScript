package lsp_test

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/bundled"
	"github.com/microsoft/TypeScript/tsc/internal/lsp/lsproto"
	"gotest.tools/v3/assert"
)

func TestTypeFunctionLogNotification(t *testing.T) {
	t.Parallel()
	if !bundled.Embedded {
		t.Skip("requires bundled libraries")
	}
	const source = `export typefn Echo<T> { log<"input", T>; return T; }
export const value: Echo<42> = 42;`
	var mu sync.Mutex
	var events []string
	client := initProjectClientWithNotifications(t, map[string]string{
		"/home/projects/tsconfig.json": `{"compilerOptions":{"traceTypeFunctions":true}}`,
		"/home/projects/index.ts":      source,
	}, func(_ context.Context, req *lsproto.RequestMessage) {
		if req.Method != lsproto.MethodWindowLogMessage {
			return
		}
		params, err := req.UnmarshalParams[*lsproto.LogMessageParams]()
		if err == nil && strings.Contains(params.Message, `"input" 42`) {
			mu.Lock()
			events = append(events, params.Message)
			mu.Unlock()
		}
	})
	uri := lsproto.DocumentUri("file:///home/projects/index.ts")
	client.SendNotification(t, lsproto.TextDocumentDidOpenInfo, &lsproto.DidOpenTextDocumentParams{
		TextDocument: &lsproto.TextDocumentItem{Uri: uri, LanguageId: "typescript", Text: source},
	})
	msg, diagnostics, ok := client.SendRequest(t, lsproto.TextDocumentDiagnosticInfo, &lsproto.DocumentDiagnosticParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: uri}})
	assert.Assert(t, ok && msg.AsResponse().Error == nil)
	assert.Equal(t, len(diagnostics.FullDocumentDiagnosticReport.Items), 0)
	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, len(events), 1)
	assert.Assert(t, strings.Contains(events[0], "index.ts:1 (called at /home/projects/index.ts:2)"))
}
