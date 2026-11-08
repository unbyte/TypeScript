package ls

import (
	"github.com/microsoft/TypeScript/tsc/internal/lsp/lsproto"
	"strings"
	"testing"
)

func TestTypefnIntrinsicEditorServices(t *testing.T) {
	source := `typefn Slice<T> { return text./*hover*/slice<T, /*argument*/0>; }
 type Result = Slice<"abc">;`
	l, ctx, pos := typefnService(t, source)
	hover, err := l.ProvideHover(ctx, &lsproto.HoverParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: "file:///main.ts"}, Position: pos("/*hover*/")})
	if err != nil || hover.Hover == nil || !strings.Contains(hover.Hover.Contents.MarkupContent.Value, "text.slice<Text, Start, End = text.length<Text>>: type") {
		t.Fatalf("intrinsic hover: %v %#v", err, hover)
	}
	signature, err := l.ProvideSignatureHelp(ctx, "file:///main.ts", pos("/*argument*/"), nil)
	if err != nil || signature.SignatureHelp == nil {
		t.Fatalf("signature: %v %#v", err, signature)
	}
	help := signature.SignatureHelp
	if !strings.HasPrefix(help.Signatures[0].Label, "text.slice<") || len(*help.Signatures[0].Parameters) != 3 || *help.ActiveParameter.Uinteger != 1 {
		t.Fatalf("signature: %#v", help)
	}
	tokens := l.typefnLexicalTokens(ctx, l.program.GetSourceFile("/main.ts"), 0, len(source))
	functions, namespaces := 0, 0
	for _, token := range tokens {
		if token.tokenType == tokenTypeFunction {
			functions++
		}
		if token.tokenType == tokenTypeNamespace {
			namespaces++
		}
	}
	if functions != 1 || namespaces != 1 {
		t.Fatalf("intrinsic tokens: %d functions, %d namespaces", functions, namespaces)
	}
	if ds := l.program.GetSemanticDiagnostics(ctx, nil); len(ds) != 0 {
		t.Fatalf("editor queries altered checking: %v", ds)
	}
}

func TestTypefnIntrinsicCompletion(t *testing.T) {
	for _, source := range []string{`typefn F<> { return text./*at*/length<"abc">; }`, `typefn F<> { return text./*at*/`, `namespace text { export type Custom = 1; } typefn F<> { return text./*at*/length<"x">; }`} {
		l, ctx, pos := typefnService(t, source)
		response, err := l.ProvideCompletion(ctx, "file:///main.ts", pos("/*at*/"), nil)
		if err != nil || response.List == nil {
			t.Fatalf("completion: %v %#v", err, response)
		}
		names := map[string]bool{}
		for _, item := range response.List.Items {
			names[item.Label] = true
		}
		if !names["length"] || !names["peek"] || !names["replaceAll"] || names["return"] || names["toNumber"] || names["Custom"] {
			t.Fatalf("intrinsic completions: %v", names)
		}
	}
}

func TestTypefnIntrinsicSignatureWhileTyping(t *testing.T) {
	for _, source := range []string{`typefn F<> { return text.slice</*at*/`, `typefn F<> { return text.slice<"abc", /*at*/`} {
		l, ctx, pos := typefnService(t, source)
		response, err := l.ProvideSignatureHelp(ctx, "file:///main.ts", pos("/*at*/"), nil)
		if err != nil || response.SignatureHelp == nil || !strings.HasPrefix(response.SignatureHelp.Signatures[0].Label, "text.slice<") {
			t.Fatalf("incomplete signature: %v %#v", err, response)
		}
	}
}

func TestTypefnPreludeDoesNotReplaceNativeNamespaces(t *testing.T) {
	source := `namespace text { export type length<T> = T; } type Result = text./*hover*/length<42>;`
	l, ctx, pos := typefnService(t, source)
	hover, err := l.ProvideHover(ctx, &lsproto.HoverParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: "file:///main.ts"}, Position: pos("/*hover*/")})
	if err != nil || hover.Hover == nil || strings.Contains(hover.Hover.Contents.MarkupContent.Value, "text.length<Text>: type") {
		t.Fatalf("native namespace hover: %v %#v", err, hover)
	}
	if ds := l.program.GetSemanticDiagnostics(ctx, nil); len(ds) != 0 {
		t.Fatalf("native namespace checking: %v", ds)
	}
}

func TestTypefnResultKindTokens(t *testing.T) {
	source := "typefn Factory<>: typefn<_, _> { return string; }"
	l, ctx, _ := typefnService(t, source)
	file := l.program.GetSourceFile("/main.ts")
	tokens := l.typefnLexicalTokens(ctx, file, 0, len(source))
	keywords, slots := 0, 0
	for _, token := range tokens {
		spelling := source[token.node.Pos():token.node.End()]
		if spelling == "typefn" && token.tokenType == tokenTypeKeyword {
			keywords++
		}
		if spelling == "_" && token.tokenType == tokenTypeOperator {
			slots++
		}
	}
	if keywords != 2 || slots != 2 {
		t.Fatalf("result kind token counts: typefn=%d slots=%d", keywords, slots)
	}
}
