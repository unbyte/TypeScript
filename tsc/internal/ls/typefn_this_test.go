package ls

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/lsp/lsproto"
)

func TestTypefnNativeReceiverInEditor(t *testing.T) {
	source := `class Box<T> {
 value!:T;
 self() {typefn Read<> {return type(this);} let result!:Read<>;return result;}
 query() {typefn Read<> {return type(typeof this.value);} let result!:Read<>;return result;}
 }
 class Derived extends Box<1> {extra!:true;}
 declare const box:Derived;
 const result=box.self();
 result./*extra*/extra;
 result./*value*/value;
 const /*queried*/queried=box.query();`
	l, ctx, pos := typefnService(t, source)
	for _, test := range []struct{ marker, want string }{{"/*extra*/", "extra: true"}, {"/*value*/", "value: 1"}, {"/*queried*/", "queried: 1"}} {
		hover, err := l.ProvideHover(ctx, &lsproto.HoverParams{TextDocument: lsproto.TextDocumentIdentifier{Uri: "file:///main.ts"}, Position: pos(test.marker)})
		if err != nil || hover.Hover == nil {
			t.Fatalf("hover %s: %v %#v", test.marker, err, hover.Hover)
		}
		text := hover.Hover.Contents.MarkupContent.Value
		if !strings.Contains(text, test.want) || strings.Contains(text, "typefn this") {
			t.Fatalf("hover %s: %s", test.marker, text)
		}
	}
	if diagnostics := l.program.GetSemanticDiagnostics(ctx, nil); len(diagnostics) != 0 {
		t.Fatalf("receiver diagnostics: %v", diagnostics)
	}
}
