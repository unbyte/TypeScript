package binder_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/binder"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/parser"
)

func TestTypefnLexicalScopes(t *testing.T) {
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/test.ts", PathKey: "/test.ts", TypefnVersion: 1}, `
 export typefn F<T> {
   type value = T;
   if (true) { type value = 1; }
   for (type item of [value]) { type copy = item; }
   return value;
 }`, core.ScriptKindTS)
	if len(file.Diagnostics()) != 0 {
		t.Fatal(file.Diagnostics())
	}
	binder.BindSourceFile(file)
	declaration := file.Statements.Nodes[0]
	body := declaration.AsTypeFunctionDeclaration().Body
	statements := body.AsTypeFunctionBlock().Statements.Nodes
	outer := ast.GetLocals(declaration)
	if declaration.Symbol() == nil || outer.Get("T") == nil || outer.Get("value") == nil {
		t.Fatal("missing declaration or lexical symbols")
	}
	if outer.Get("item") != nil || outer.Get("copy") != nil {
		t.Fatal("iteration locals leaked into enclosing scope")
	}
	inner := statements[1].AsTypeFunctionStatement().Body
	if ast.GetLocals(inner).Get("value") == nil || ast.GetLocals(inner).Get("value") == outer.Get("value") {
		t.Fatal("shadowed local must have its own identity")
	}
	loop := statements[2]
	loopLocals := ast.GetLocals(loop.AsTypeFunctionStatement().Body)
	if loopLocals.Get("item") != loop.Symbol() || loopLocals.Get("copy") == nil {
		t.Fatal("iteration binding must belong to body scope")
	}
	if outer.Get("value") != statements[0].Symbol() {
		t.Fatal("binding identity does not match scope table")
	}
}

func TestTypefnLambdaScope(t *testing.T) {
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/test.ts", PathKey: "/test.ts", TypefnVersion: 1}, `typefn F<T> {type closure=typefn<T> {type value=T; return value;}; return closure<T>;}`, core.ScriptKindTS)
	if len(file.Diagnostics()) != 0 {
		t.Fatal(file.Diagnostics())
	}
	binder.BindSourceFile(file)
	fn := file.Statements.Nodes[0]
	binding := ast.TypeFunctionBody(fn).AsTypeFunctionBlock().Statements.Nodes[0]
	lambda := binding.AsTypeFunctionStatement().Expression
	locals := ast.GetLocals(lambda)
	if locals.Get("T") == nil || locals.Get("value") == nil || locals.Get("T") == ast.GetLocals(fn).Get("T") {
		t.Fatal("lambda needs its own parameter/body scope")
	}
	if ast.GetLocals(fn).Get("value") != nil {
		t.Fatal("lambda local escaped into enclosing function")
	}
}
