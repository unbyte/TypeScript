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
type cancelInTypefnNative struct {
	context.Context
	cancel context.CancelFunc
	inside bool
	frame  string
}

func (c *cancelInTypefnNative) Err() error {
	if err := c.Context.Err(); err != nil {
		return err
	}
	var pcs [128]uintptr
	n := runtime.Callers(2, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		if strings.Contains(frame.Function, c.frame) {
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
			testTypefnNativeCancellation(t, source.String(), ".(*Relater).")
		})
	}
}

func newTypefnNativeTestProgram(t testing.TB, source string) *compiler.Program {
	t.Helper()
	fs := bundled.WrapFS(vfstest.FromMap(map[string]string{
		"/main.ts":       source,
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

// Each source ends with a native Result alias applying the DSL and an Expected
// alias containing the completed native Type. Exercise both public query paths.
func testTypefnNativeCancellation(t *testing.T, source, frame string) {
	t.Helper()
	for _, diagnostics := range []bool{false, true} {
		t.Run(fmt.Sprintf("diagnostics=%v", diagnostics), func(t *testing.T) {
			p := newTypefnNativeTestProgram(t, source)
			c, done := p.GetTypeChecker(t.Context())
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			inside := &cancelInTypefnNative{Context: ctx, cancel: cancel, frame: frame}
			file := p.GetSourceFile("/main.ts")
			nodes := file.Statements.Nodes
			if diagnostics {
				if got := c.GetDiagnostics(inside, file); len(got) != 0 {
					t.Fatal("canceled checking published diagnostics", got)
				}
			} else if got := c.GetTypeFunctionApplicationResult(inside, nodes[len(nodes)-2].Type()); got != nil {
				t.Fatal("canceled native computation published a query result")
			}
			if !inside.inside || !c.WasCanceled() {
				t.Fatalf("did not cancel inside %s: inside=%v canceled=%v", frame, inside.inside, c.WasCanceled())
			}
			done()
			// Discard canceled snapshots; complete results come from a fresh checker.
			fresh := newTypefnNativeTestProgram(t, source)
			clean, release := fresh.GetTypeChecker(t.Context())
			defer release()
			nodes = fresh.GetSourceFile("/main.ts").Statements.Nodes
			actual := clean.GetTypeFunctionApplicationResult(t.Context(), nodes[len(nodes)-2].Type())
			expected := clean.GetTypeAtLocation(nodes[len(nodes)-1].Type())
			if actual == nil || actual != expected || clean.WasCanceled() {
				t.Fatal("fresh checker did not produce the expected native Type")
			}
		})
	}
}

func TestTypefnCancellationInsideNativeInstantiation(t *testing.T) {
	t.Run("mapped properties", func(t *testing.T) {
		var source strings.Builder
		source.WriteString("interface Input {\n")
		for i := range 4000 {
			fmt.Fprintf(&source, "p%d: number;\n", i)
		}
		source.WriteString(`}
   typefn Project<T> { return tuple.length<reflect.properties<type({[K in keyof T]: {value: T[K]}})>>; }
   type Result = Project<Input>;
   type Expected = 4000;
  `)
		testTypefnNativeCancellation(t, source.String(), ".(*Checker).getTypeOfMappedSymbol")
	})
	t.Run("distributive conditional", func(t *testing.T) {
		var source strings.Builder
		source.WriteString("type Input = never")
		for i := range 4000 {
			fmt.Fprintf(&source, " | \"v%d\"", i)
		}
		source.WriteString(`;
   type Decode<T> = T extends string ? {value: T} : never;
   typefn Project<T> { return reflect.kind<type(Decode<T>)>; }
   type Result = Project<Input>;
   type Expected = "union";
  `)
		testTypefnNativeCancellation(t, source.String(), ".(*Checker).getConditionalType")
	})
}

func TestTypefnCancellationInsideNativeNormalization(t *testing.T) {
	var union strings.Builder
	union.WriteString("type Input = never")
	for i := range 4000 {
		fmt.Fprintf(&union, " | \"v%d\"", i)
	}
	union.WriteString(`;
	 typefn Normalize<T> { return reflect.kind<type(T | "tail")>; }
	 type Result = Normalize<Input>;
	 type Expected = "union";
	`)
	for _, frame := range []string{".(*Checker).addTypesToUnion", ".(*Checker).sortTypefnUnion"} {
		t.Run(frame, func(t *testing.T) {
			testTypefnNativeCancellation(t, union.String(), frame)
		})
	}
	t.Run("intersection members", func(t *testing.T) {
		var source strings.Builder
		source.WriteString("type Input = unknown")
		for i := range 4000 {
			fmt.Fprintf(&source, " & { p%d: number }", i)
		}
		source.WriteString(`;
		 typefn Normalize<T> { return reflect.kind<type(T & {tail: true})>; }
		 type Result = Normalize<Input>;
		 type Expected = "intersection";
		`)
		testTypefnNativeCancellation(t, source.String(), ".(*Checker).addTypesToIntersection")
	})
}
