package checker_test

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/bundled"
	"github.com/microsoft/TypeScript/tsc/internal/compiler"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/tsoptions"
	"github.com/microsoft/TypeScript/tsc/internal/tspath"
	"github.com/microsoft/TypeScript/tsc/internal/vfs/vfstest"
)

// Cancellation must be observed inside the native comparison, not just at the
// next VM opcode or query return. Restricting the cancellation point by native
// stack frames makes this deterministic without sleeps or production test hooks.
type cancelInTypefnRelation struct {
	context.Context
	cancel context.CancelFunc
	inside bool
}

func (c *cancelInTypefnRelation) Err() error {
	if err := c.Context.Err(); err != nil {
		return err
	}
	var pcs [128]uintptr
	n := runtime.Callers(2, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		if strings.Contains(frame.Function, ".(*Relater).") {
			c.inside = true
			c.cancel()
			return c.Context.Err()
		}
		if !more {
			return nil
		}
	}
}

func TestTypefnCancellationInsideNativeRelation(t *testing.T) {
	for _, matches := range []bool{true, false} {
		t.Run(fmt.Sprint(matches), func(t *testing.T) {
			var source strings.Builder
			source.WriteString("interface Source {\n")
			for i := range 4000 {
				value := "0"
				if !matches && i == 3999 {
					value = "string"
				}
				fmt.Fprintf(&source, "p%d: %s;\n", i, value)
			}
			source.WriteString("}\ninterface Target {\n")
			for i := range 4000 {
				fmt.Fprintf(&source, "p%d: number;\n", i)
			}
			fmt.Fprintf(&source, "}\ntypefn Relate<A, B> { return A is B; }\ntype Result = Relate<Source, Target>;\ntype Expected = %t;\n", matches)
			newProgram := func() *compiler.Program {
				fs := bundled.WrapFS(vfstest.FromMap(map[string]string{
					"/main.ts":       source.String(),
					"/tsconfig.json": `{"compilerOptions":{"typefn":{"version":1},"strict":true,"ignoreDeprecations":"6.0"},"files":["main.ts"]}`,
				}, tspath.CaseSensitive))
				host := compiler.NewCompilerHost(fs, bundled.LibPath(), nil, nil, nil)
				config, errors := tsoptions.GetParsedCommandLineOfConfigFile("/tsconfig.json", &core.CompilerOptions{}, nil, tsoptions.NewDefaultParseConfigHost(fs), nil)
				if len(errors) != 0 {
					t.Fatal(errors)
				}
				p := compiler.NewProgram(compiler.ProgramOptions{Config: config, Host: host})
				p.BindSourceFiles()
				return p
			}
			p := newProgram()
			c, done := p.GetTypeChecker(t.Context())
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			inRelation := &cancelInTypefnRelation{Context: ctx, cancel: cancel}
			node := p.GetSourceFile("/main.ts").Statements.Nodes[3].Type()
			result := c.GetTypeFunctionApplicationResult(inRelation, node)
			if !inRelation.inside || result != nil || !c.WasCanceled() {
				t.Fatalf("native cancellation did not suppress publication: inside=%v, result=%v, canceled=%v", inRelation.inside, result, c.WasCanceled())
			}
			done()

			// Canceled snapshots are discarded. A fresh checker must still compute
			// the exact native boolean, in both successful and failed relations.
			fresh := newProgram()
			clean, release := fresh.GetTypeChecker(t.Context())
			defer release()
			nodes := fresh.GetSourceFile("/main.ts").Statements.Nodes
			actual := clean.GetTypeFunctionApplicationResult(t.Context(), nodes[3].Type())
			expected := clean.GetTypeAtLocation(nodes[4].Type())
			if actual == nil || actual != expected || clean.WasCanceled() {
				t.Fatal("fresh native comparison did not produce the expected boolean Type")
			}
		})
	}
}
