package fourslash_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/fourslash"
	"github.com/microsoft/TypeScript/tsc/internal/testutil"
)

func TestTypeFunctionLanguageService(t *testing.T) {
	t.Parallel()
	defer testutil.RecoverAndFail(t, "type function language service")
	const content = `export typefn Count<T> {
    type Items = T;
    return tuple./*members*/count</*local*/Items>;
}
type Two = /*call*/Count<[1, 2]>;`
	f, done := fourslash.NewFourslash(t, nil, content)
	defer done()
	f.VerifyCompletions(t, "members", &fourslash.CompletionsExpectedList{
		Items: &fourslash.CompletionsExpectedItems{Includes: []fourslash.CompletionsExpectedItem{"count", "decompose", "compose"}},
	})
	f.VerifyBaselineGoToDefinition(t, true, "local", "call")
	f.VerifyBaselineRename(t, nil, "local", "call")
	f.VerifyBaselineHover(t)
}

func TestTypeFunctionRecheckAfterEdit(t *testing.T) {
	t.Parallel()
	defer testutil.RecoverAndFail(t, "update a type function input")
	const content = `typefn Count<S> { return tuple.count<string.split<S, " ">>; }
declare const count: Count<"a/*input*/ b">;
const two: 2 = count;`
	f, done := fourslash.NewFourslash(t, nil, content)
	defer done()
	f.VerifyNoErrors(t)
	f.GoToMarker(t, "input")
	f.Insert(t, " c")
	f.VerifyNumberOfErrorsInCurrentFile(t, 1)
	f.GoToMarker(t, "input")
	f.DeleteAtCaret(t, 2)
	f.VerifyNoErrors(t)
}

func TestTypeFunctionBuiltinSignatureHelp(t *testing.T) {
	t.Parallel()
	const source = `typefn Parse<S> { return number./*methods*/parse</*input*/S>; }
typefn Match<S> { return string.matchPrefix<S, /*index*/0, /*prefix*/" ">; }`
	f, done := fourslash.NewFourslash(t, nil, source)
	defer done()
	f.VerifyCompletions(t, "methods", &fourslash.CompletionsExpectedList{
		Items: &fourslash.CompletionsExpectedItems{Includes: []fourslash.CompletionsExpectedItem{"parse"}},
	})
	f.GoToMarker(t, "input")
	f.VerifySignatureHelp(t, fourslash.VerifySignatureHelpOptions{Text: "number.parse<S>", ParameterCount: 1, ParameterName: "S"})
	const signature = "string.matchPrefix<S, Index, Prefix>"
	f.GoToMarker(t, "index")
	f.VerifySignatureHelp(t, fourslash.VerifySignatureHelpOptions{Text: signature, ParameterCount: 3, ParameterName: "Index"})
	f.GoToMarker(t, "prefix")
	f.VerifySignatureHelp(t, fourslash.VerifySignatureHelpOptions{Text: signature, ParameterCount: 3, ParameterName: "Prefix"})
}
