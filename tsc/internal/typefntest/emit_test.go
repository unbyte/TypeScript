package typefntest

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/compiler"
)

// Builder signatures and forced declaration output bypass native error gates.
// They must still withhold a file whose executable recipe cannot be published.
func TestTypefnDeclarationAtomicEmit(t *testing.T) {
	for _, name := range []string{
		"error-declaration-invalid-recipe",
		"error-declaration-malformed-recipe",
		"error-declaration-native-wrapper",
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
