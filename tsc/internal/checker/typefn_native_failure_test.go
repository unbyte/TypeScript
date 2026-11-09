package checker_test

import (
	"strings"
	"testing"
)

func TestTypefnMappedFailureQueryOrder(t *testing.T) {
	const source = `typefn Guard<N> {
  if (N == 0) { fail<"Member replay failure">; }
  return N;
}
type Mapped<T> = { [K in keyof T]: Guard<T[K]> };
type Input = { value: 0 };
typefn Read<T> { return type(Mapped<T>["value" & keyof T]); }
type First = Read<Input /* first */>;
type Second = Read<Input /* second */>;
`
	for _, reverse := range []bool{false, true} {
		t.Run(map[bool]string{false: "forward", true: "reverse"}[reverse], func(t *testing.T) {
			p := newTypefnNativeTestProgram(t, source)
			c, release := p.GetTypeChecker(t.Context())
			defer release()
			file := p.GetSourceFile("/main.ts")
			applications := file.Statements.Nodes[len(file.Statements.Nodes)-2:]
			for range 2 {
				for i := range applications {
					if reverse {
						i = len(applications) - 1 - i
					}
					if c.GetTypeFunctionApplicationResult(t.Context(), applications[i].Type()) != nil {
						t.Fatal("failed member calculation published a result")
					}
				}
			}
			diagnostics := c.GetDiagnostics(t.Context(), file)
			if len(diagnostics) != 2 {
				t.Fatalf("got %d diagnostics, want both applications", len(diagnostics))
			}
			for i, site := range []string{"Read<Input /* first */>", "Read<Input /* second */>"} {
				if diagnostics[i].Code() != -1 || diagnostics[i].Loc().Pos() != strings.Index(source, site) {
					t.Fatalf("diagnostic %d lost its application site", i)
				}
			}
		})
	}
}
