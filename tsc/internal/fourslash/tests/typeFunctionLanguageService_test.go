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
