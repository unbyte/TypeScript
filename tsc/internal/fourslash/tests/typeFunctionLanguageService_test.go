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
	const content = `typefn Count<S> { return tuple.count<string.split<S, " ">> * 2; }
declare const count: Count<"a/*input*/ b">;
const four: 4 = count;`
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

func TestTypeFunctionReturnBoundaryAfterEdit(t *testing.T) {
	t.Parallel()
	const content = `typefn Make<N> { return <T>(arg: T) => [T, /*operand*/T + 1]; }
declare const make: Make<3>;
const pair: ["x", 4] = make("x" as const);`
	f, done := fourslash.NewFourslash(t, nil, content)
	defer done()
	f.VerifyNumberOfErrorsInCurrentFile(t, 1)
	f.GoToMarker(t, "operand")
	f.DeleteAtCaret(t, 1)
	f.Insert(t, "N")
	f.VerifyNoErrors(t)
}

func TestTypeFunctionPrimitiveHover(t *testing.T) {
	t.Parallel()
	const source = `typefn Inspect<T> {
    lo/*log*/g<>;
    fa/*fail*/il<"bad input">;
    return str/*namespace*/ing.sp/*split*/lit<"x y", " ">;
}
namespace custom {
    type log<T> = { value: T };
    type fail = "ordinary type";
    export namespace tuple { export type count<T> = T; }
    export typefn Count<T> {
        lo/*statement*/g<T>;
        type Box = lo/*alias*/g<fa/*argument*/il>;
        return tuple.co/*shadow*/unt<T>;
    }
    type Outside = lo/*outside*/g<string>;
}`
	f, done := fourslash.NewFourslash(t, nil, source)
	defer done()
	f.VerifyBaselineHover(t)
	f.VerifyBaselineGoToDefinition(t, true, "statement", "alias")
}

func TestTypeFunctionStatementSemanticTokens(t *testing.T) {
	t.Parallel()
	const source = `type log = "alias";
typefn Check<> { log<>; fail<log>; }`
	f, done := fourslash.NewFourslash(t, nil, source)
	defer done()
	f.VerifySemanticTokens(t, []fourslash.SemanticToken{
		{Type: "type.declaration", Text: "log"},
		{Type: "type.declaration", Text: "Check"},
		{Type: "keyword", Text: "log"},
		{Type: "keyword", Text: "fail"},
		{Type: "type", Text: "log"},
	})
}
