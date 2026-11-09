package typefntest

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/checker"
	"github.com/microsoft/TypeScript/tsc/internal/compiler"
)

// Builder signatures and forced declaration output bypass native error gates.
// They must still withhold a file whose executable recipe cannot be published.
func TestTypefnDeclarationAtomicEmit(t *testing.T) {
	for _, name := range []string{
		"error-declaration-invalid-recipe",
		"error-declaration-malformed-recipe",
		"error-declaration-native-wrapper",
		"error-declaration-inferred-dependency",
		"error-declaration-inferred-recipe-stripped",
		"error-declaration-inferred-local-recipe",
		"error-declaration-local-annotation",
		"error-declaration-value-stripped",
		"error-declaration-qualified-stripped",
		"error-declaration-authored-bridge",
		"error-declaration-malformed-dependency",
		"error-declaration-malformed-authored-dependency",
	} {
		for _, mode := range []struct {
			name    string
			options compiler.EmitOptions
		}{
			{"ordinary", compiler.EmitOptions{}},
			{"forced", compiler.EmitOptions{ForceEmit: true, EmitOnly: compiler.EmitOnlyDts}},
			{"builder", compiler.EmitOptions{EmitOnly: compiler.EmitOnlyBuilderSignature}},
		} {
			t.Run(name+"/"+mode.name, func(t *testing.T) {
				data, err := os.ReadFile("testdata/" + name + ".json")
				if err != nil {
					t.Fatal(err)
				}
				var f fixture
				if err := json.Unmarshal(data, &f); err != nil {
					t.Fatal(err)
				}
				f.emitOptions = mode.options
				if mode.name != "ordinary" {
					f.Options["noEmitOnError"] = true
				}
				runFixture(t, f)
			})
		}
	}
}

func TestTypefnInferredDefaultQueryOrder(t *testing.T) {
	for _, beforeCheck := range []bool{true, false} {
		name := "after-semantic-check"
		if beforeCheck {
			name = "before-semantic-check"
		}
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile("testdata/declaration-inferred-defaults-augmented.json")
			if err != nil {
				t.Fatal(err)
			}
			var f fixture
			if err := json.Unmarshal(data, &f); err != nil {
				t.Fatal(err)
			}
			query := func(t testing.TB, program *compiler.Program) {
				c, done := program.GetTypeChecker(t.Context())
				defer done()
				var inferred, defaulted *ast.Node
				var visit func(*ast.Node) bool
				visit = func(node *ast.Node) bool {
					if ast.IsVariableDeclaration(node) {
						switch node.Name().Text() {
						case "inferred":
							inferred = node.Initializer()
						case "defaults":
							defaulted = node.Initializer()
						}
					}
					return node.ForEachChild(visit)
				}
				visit(program.GetSourceFile("/project/library.ts").AsNode())
				if inferred == nil || defaulted == nil {
					t.Fatal("missing call probes")
				}
				original := c.GetResolvedSignature(inferred)
				if original != c.GetResolvedSignature(defaulted) {
					t.Fatal("probe must share a native instantiated signature despite distinct argument origins")
				}
				for range 2 {
					if len(c.GetCandidateSignaturesForStringLiteralCompletions(inferred, inferred.Arguments()[0])) == 0 {
						t.Fatal("missing completion candidates")
					}
					if _, candidates := checker.GetResolvedSignatureForSignatureHelp(inferred, 1, c); len(candidates) == 0 {
						t.Fatal("missing signature-help candidates")
					}
				}
				if c.GetResolvedSignature(inferred) != original {
					t.Fatal("editor query changed the cached native signature")
				}
			}
			if beforeCheck {
				f.beforeCheck = query
			} else {
				f.beforeEmit = query
			}
			runFixture(t, f)
		})
	}
}
