package parser_test

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/parser"
)

func parseTypefn(t testing.TB, source string, options ast.SourceFileParseOptions) *ast.SourceFile {
	t.Helper()
	options.FileName, options.PathKey = "/test.ts", "/test.ts"
	return parser.ParseSourceFile(options, source, core.ScriptKindTS)
}

func TestTypefnExpressions(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		source   string
		kind     ast.Kind
		operator ast.Kind
	}{
		{"A + B * C", ast.KindTypeFunctionBinary, ast.KindPlusToken},
		{"A < B && C > D", ast.KindTypeFunctionBinary, ast.KindAmpersandAmpersandToken},
		{"A is B | C", ast.KindTypeFunctionBinary, ast.KindIsKeyword},
		{"F<A + 1, G<B>>", ast.KindTypeFunctionCall, ast.KindUnknown},
		{"F < T >", ast.KindTypeFunctionCall, ast.KindUnknown},
		{"Choose<(A < B)>", ast.KindTypeFunctionCall, ast.KindUnknown},
		{"A >= B", ast.KindTypeFunctionBinary, ast.KindGreaterThanEqualsToken},
		{"[A, ...B]", ast.KindTypeFunctionTuple, ast.KindUnknown},
		{"{ readonly: true; readonly value?: T }", ast.KindTypeFunctionObject, ast.KindUnknown},
		{"type((value: T) => Promise<T>)", ast.KindTypeFunctionNativeType, ast.KindUnknown},
		{"F<T>[0]", ast.KindTypeFunctionIndex, ast.KindUnknown},
		{"`head${F<T + 1>}tail${T}`", ast.KindTypeFunctionTemplate, ast.KindUnknown},
		{"`outer${`inner${T}`}end`", ast.KindTypeFunctionTemplate, ast.KindUnknown},
	} {
		t.Run(test.source, func(t *testing.T) {
			file := parseTypefn(t, "typefn Example<A, B, C, D, T> { return "+test.source+"; }", ast.SourceFileParseOptions{TypefnVersion: 1})
			if ds := file.Diagnostics(); len(ds) != 0 {
				t.Fatalf("parse diagnostics: %v", ds)
			}
			decl := file.Statements.Nodes[0]
			if !ast.IsTypeFunctionDeclaration(decl) {
				t.Fatalf("wrong declaration kind: %v", decl.Kind)
			}
			ret := decl.AsTypeFunctionDeclaration().Body.AsTypeFunctionBlock().Statements.Nodes[0]
			expr := ret.AsTypeFunctionStatement().Expression
			if expr.Kind != test.kind {
				t.Fatalf("kind: %v, want %v", expr.Kind, test.kind)
			}
			if test.operator != ast.KindUnknown && expr.AsTypeFunctionExpression().OperatorToken.Kind != test.operator {
				t.Fatalf("wrong root operator: %v", expr.AsTypeFunctionExpression().OperatorToken.Kind)
			}
			checkTypefnTree(t, file.AsNode())
		})
	}
}

func checkTypefnTree(t testing.TB, node *ast.Node) {
	t.Helper()
	end := node.Pos()
	node.ForEachChild(func(child *ast.Node) bool {
		if child.Pos() < end {
			t.Fatalf("children of %v are not in source order at %v", node.Kind, child.Kind)
		}
		end = child.End()
		if child.Parent != node {
			t.Fatalf("%v has incorrect parent", child.Kind)
		}
		if child.Pos() < node.Pos() || child.End() > node.End() {
			t.Fatalf("child %v span outside parent %v", child.Kind, node.Kind)
		}
		checkTypefnTree(t, child)
		return false
	})
}

func TestTypefnStatementsAndErasureBoundary(t *testing.T) {
	t.Parallel()
	source := `export typefn Reverse<T> {
  type result = [];
  for (type item of T) {
    if (item is string) { result = [item, ...result]; } else { continue; }
  }
  while (false) { break; }
  return result;
}
const typefn = 1;
export const value = typefn;
`
	file := parseTypefn(t, source, ast.SourceFileParseOptions{TypefnVersion: 1})
	if ds := file.Diagnostics(); len(ds) != 0 {
		t.Fatal(ds)
	}
	if len(file.Statements.Nodes) != 3 || !ast.IsVariableStatement(file.Statements.Nodes[1]) {
		t.Fatal("DSL consumed following native code")
	}
	checkTypefnTree(t, file.AsNode())
	disabled := parseTypefn(t, source, ast.SourceFileParseOptions{})
	if len(disabled.Diagnostics()) == 0 {
		t.Fatal("disabled profile accepted DSL")
	}
	for _, node := range disabled.Statements.Nodes {
		if ast.IsTypeFunctionDeclaration(node) {
			t.Fatal("disabled profile created DSL AST")
		}
	}
}

func TestTypefnRejectsUnsupportedSyntax(t *testing.T) {
	t.Parallel()
	for _, body := range []string{
		"type x = 1\nreturn x;", "return A < B > C;", "return F<T><U>;",
		"function f() {} return string;", "type f = x => x;", "type [x] = T;",
		"return /abc/;", "return text.length(T);", "if (true) return string;", "return T;;",
	} {
		t.Run(body, func(t *testing.T) {
			file := parseTypefn(t, "typefn F<T> { "+body+" }", ast.SourceFileParseOptions{TypefnVersion: 1})
			if len(file.Diagnostics()) == 0 {
				t.Fatal("accepted invalid DSL")
			}
		})
	}
}

func TestTypefnParserLimits(t *testing.T) {
	t.Parallel()
	for _, options := range []ast.SourceFileParseOptions{
		{TypefnVersion: 1, TypefnSyntaxDepth: 16},
		{TypefnVersion: 1, TypefnParserWork: 16},
	} {
		file := parseTypefn(t, "typefn F<T> { return "+strings.Repeat("F<", 100)+"T"+strings.Repeat(">", 100)+"; }", options)
		found := false
		for _, diagnostic := range file.Diagnostics() {
			if strings.Contains(diagnostic.String(), "limit exceeded") {
				found = true
			}
		}
		if !found {
			t.Fatal("missing resource diagnostic after speculative parsing")
		}
	}
}

func FuzzTypefnParser(f *testing.F) {
	for _, source := range []string{
		"typefn F<T> { return F<T>; }", "typefn F<T> { type x = 0; while (x < 10) { x = x + 1; } return x; }",
		"typefn F<T> { return F<(T < 3)>; }", "typefn F<T> { return { readonly: true }; }", "typefn F<T> { return F<F<F<T>>>; }",
	} {
		f.Add(source)
	}
	f.Fuzz(func(t *testing.T, source string) {
		if len(source) > 16384 {
			t.Skip()
		}
		file := parseTypefn(t, source, ast.SourceFileParseOptions{TypefnVersion: 1, TypefnSyntaxDepth: 64, TypefnParserWork: 20000})
		if len(file.Diagnostics()) == 0 {
			checkTypefnTree(t, file.AsNode())
		}
	})
}

func TestTypefnElseIfNestingLimit(t *testing.T) {
	source := "typefn F<> { " + strings.Repeat("if (false) {} else ", 1000) + "{ return 0; } }"
	file := parseTypefn(t, source, ast.SourceFileParseOptions{TypefnVersion: 1, TypefnSyntaxDepth: 16})
	for _, diagnostic := range file.Diagnostics() {
		if strings.Contains(diagnostic.String(), "limit exceeded") {
			return
		}
	}
	t.Fatal("else-if recursion bypassed syntax depth limit")
}

func TestTypefnRejectsNativeParameterModifiers(t *testing.T) {
	for _, modifier := range []string{"in", "out", "const", "in out"} {
		file := parseTypefn(t, "typefn F<"+modifier+" T> { return T; }", ast.SourceFileParseOptions{TypefnVersion: 1})
		if len(file.Diagnostics()) == 0 {
			t.Errorf("accepted unsupported parameter modifier %s", modifier)
		}
	}
}

func TestTypefnContextualIdentifiers(t *testing.T) {
	for _, name := range []string{"out", "async", "override"} {
		source := "typefn Identity<" + name + "> { return " + name + "; } " +
			"typefn F<> { type " + name + " = 1; " + name + " = Identity<" + name + " + 1>; return " + name + "; }"
		file := parseTypefn(t, source, ast.SourceFileParseOptions{TypefnVersion: 1})
		if ds := file.Diagnostics(); len(ds) != 0 {
			t.Fatalf("contextual identifier %s: %v", name, ds)
		}
		checkTypefnTree(t, file.AsNode())
	}
}

func TestTypefnTemplateRecovery(t *testing.T) {
	for _, source := range []string{"typefn F<T> { return `a${T + }b`; }", "typefn F<T> { return `a${() => T}b`; }", "typefn F<T> { return `a${T"} {
		file := parseTypefn(t, source, ast.SourceFileParseOptions{TypefnVersion: 1})
		if len(file.Diagnostics()) == 0 {
			t.Fatalf("accepted malformed template: %s", source)
		}
	}
}
