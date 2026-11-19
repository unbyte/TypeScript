package checker_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/bundled"
	"github.com/microsoft/TypeScript/tsc/internal/checker"
	"github.com/microsoft/TypeScript/tsc/internal/compiler"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/tsoptions"
	"github.com/microsoft/TypeScript/tsc/internal/tspath"
	"github.com/microsoft/TypeScript/tsc/internal/vfs/vfstest"
	"gotest.tools/v3/assert"
)

func BenchmarkTypeFunctionFields(b *testing.B) {
	for _, size := range []int{8, 32, 100} {
		names := make([]string, size)
		for i := range names {
			names[i] = fmt.Sprintf(`"field%d"`, i)
		}
		use := fmt.Sprintf("declare const result: Fields<[%s]>; const check: true | undefined = result.field%d;", strings.Join(names, ","), size-1)
		for _, form := range []struct{ name, declaration string }{
			{"mapped", "type Fields<T extends string[]> = { readonly [K in T[number]]?: true };"},
			{"typefn", `typefn Fields<T extends string[]> {
				type Out = {};
				for (type K of T) { Out = object.addField<Out, K, {type: true; optional: true; readonly: true}>; }
				return Out;
			}`},
		} {
			b.Run(fmt.Sprintf("%s/%d", form.name, size), func(b *testing.B) {
				fs := bundled.WrapFS(vfstest.FromMap(map[string]string{"/bench.ts": form.declaration + "\n" + use}, tspath.CaseSensitive))
				config := tsoptions.NewParsedCommandLine(&core.CompilerOptions{Strict: core.TSTrue}, []tspath.RootedFilePath{"/bench.ts"}, nil, "/", fs.CaseSensitivity())
				program, err := compiler.NewProgram(b.Context(), compiler.ProgramOptions{Config: config, Host: compiler.NewCompilerHost(fs, bundled.LibPath(), nil, nil, nil)})
				assert.NilError(b, err)
				program.BindSourceFiles()
				file := program.GetSourceFile("/bench.ts")
				b.ReportAllocs()
				for b.Loop() {
					c, _ := checker.NewChecker(program, nil)
					assert.Equal(b, len(c.GetDiagnostics(b.Context(), file)), 0)
				}
			})
		}
	}
}
