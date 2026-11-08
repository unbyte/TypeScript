package ls

import (
	"github.com/microsoft/TypeScript/tsc/internal/lsp/lsproto"
	"strings"
	"testing"
)

func TestTypefnTemplateEditor(t *testing.T) {
	source := "typefn Wrap<T> { return `if ${text.slice<T, /*arg*/0>} return ${T}`; }\ntype Result = /*call*/Wrap<\"abc\">;"
	l, ctx, pos := typefnService(t, source)
	signature, err := l.ProvideSignatureHelp(ctx, "file:///main.ts", pos("/*arg*/"), nil)
	if err != nil || signature.SignatureHelp == nil || !strings.HasPrefix(signature.SignatureHelp.Signatures[0].Label, "text.slice<") {
		t.Fatalf("template signature: %v %#v", err, signature)
	}
	hover, err := l.ProvideHover(ctx, &lsproto.HoverParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: "file:///main.ts"}, Position: pos("/*call*/")})
	if err != nil || hover.Hover == nil || !strings.Contains(hover.Hover.Contents.MarkupContent.Value, "if abc return abc") {
		t.Fatalf("template hover: %v %#v", err, hover)
	}
	tokens := l.typefnLexicalTokens(ctx, l.program.GetSourceFile("/main.ts"), 0, len(source))
	literals, returns := 0, 0
	for _, token := range tokens {
		if token.tokenType == tokenTypeString {
			literals++
		}
		if token.tokenType == tokenTypeKeyword && source[token.node.Pos():token.node.End()] == "return" {
			returns++
		}
	}
	if literals != 3 || returns != 1 {
		t.Fatalf("template tokens: %d literals, %d returns", literals, returns)
	}
	if ds := l.program.GetSemanticDiagnostics(ctx, nil); len(ds) != 0 {
		t.Fatalf("editor queries altered template checking: %v", ds)
	}
}

func TestTypefnSignatureAfterTemplateArgument(t *testing.T) {
	source := "typefn Pair<A, B> { return [A, B]; } typefn Demo<T> { return Pair<`x${T}`, /*arg*/T>; }"
	l, ctx, pos := typefnService(t, source)
	response, err := l.ProvideSignatureHelp(ctx, "file:///main.ts", pos("/*arg*/"), nil)
	if err != nil || response.SignatureHelp == nil {
		t.Fatalf("signature: %v %#v", err, response)
	}
	help := response.SignatureHelp
	if help.Signatures[0].Label != "typefn Pair<A, B>: type" || *help.ActiveParameter.Uinteger != 1 {
		t.Fatalf("template argument corrupted signature frames: %#v", help)
	}
}
