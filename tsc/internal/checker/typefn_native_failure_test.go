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
	testTypefnMemberFailureQueryOrder(t, source, true, []string{"Read<Input /* first */>", "Read<Input /* second */>"})
}

func TestTypefnInstantiatedMemberFailureQueryOrder(t *testing.T) {
	const header = `typefn Guard<N> {
  if (N == 0) { fail<"Member replay failure">; }
  return N;
}
interface Box<N> { value: Guard<N>; }
`
	t.Run("indexed", func(t *testing.T) {
		const uses = `type First = Box<0 /* first */>["value"];
type Second = Box<0 /* second */>["value"];
`
		testTypefnMemberFailureQueryOrder(t, header+uses, false, []string{"Box<0 /* first */>", "Box<0 /* second */>"})
	})
	t.Run("alias", func(t *testing.T) {
		const uses = `type Index<N> = Box<N>["value"];
type First = Index<0 /* first */>;
type Second = Index<0 /* second */>;
`
		testTypefnMemberFailureQueryOrder(t, header+uses, false, []string{"Index<0 /* first */>", "Index<0 /* second */>"})
	})
}

func TestTypefnConditionalFailureQueryOrder(t *testing.T) {
	const header = `typefn Guard<N> {
  if (N == 0) { fail<"Member replay failure">; }
  return N;
}
`
	for _, test := range []struct {
		name, declaration, input string
	}{
		{"tuple", "type Container<N> = N extends number ? [Guard<N>] : [N];", "0"},
		{"absorbed", "type Container<N> = N extends number ? Guard<N> | unknown : N;", "0"},
		{"distributed", "type Container<N> = N extends number ? [Guard<N>] : [N];", "0 | 1"},
	} {
		t.Run(test.name, func(t *testing.T) {
			sites := []string{"Container<" + test.input + " /* first */>", "Container<" + test.input + " /* second */>"}
			source := header + test.declaration + "\ntype First = " + sites[0] + ";\ntype Second = " + sites[1] + ";\n"
			testTypefnMemberFailureQueryOrder(t, source, false, sites)
		})
	}
}

func testTypefnMemberFailureQueryOrder(t *testing.T, source string, typefnQuery bool, sites []string) {
	t.Helper()
	modes := []string{"typefn"}
	if !typefnQuery {
		modes = []string{"type", "name", "declaration", "symbol"}
	}
	for _, mode := range modes {
		for _, reverse := range []bool{false, true} {
			order := "forward"
			if reverse {
				order = "reverse"
			}
			t.Run(mode+"/"+order, func(t *testing.T) {
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
						switch mode {
						case "typefn":
							if c.GetTypeFunctionApplicationResult(t.Context(), applications[i].Type()) != nil {
								t.Fatal("failed member calculation published a result")
							}
						case "type":
							c.GetTypeAtLocation(applications[i].Type())
						case "name":
							c.GetTypeAtLocation(applications[i].Name())
						case "declaration":
							c.GetTypeAtLocation(applications[i])
						case "symbol":
							c.GetDeclaredTypeOfSymbol(applications[i].Symbol())
						}
					}
				}
				diagnostics := c.GetDiagnostics(t.Context(), file)
				if len(diagnostics) != 2 {
					t.Fatalf("got %d diagnostics, want both applications", len(diagnostics))
				}
				for i, site := range sites {
					if diagnostics[i].Code() != -1 || diagnostics[i].Loc().Pos() != strings.Index(source, site) {
						t.Fatalf("diagnostic %d: code %d at %d, want application at %d (%s)", i, diagnostics[i].Code(), diagnostics[i].Loc().Pos(), strings.Index(source, site), site)
					}
				}
			})
		}
	}
}
