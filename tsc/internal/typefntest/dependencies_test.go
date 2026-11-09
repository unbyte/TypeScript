package typefntest

import (
	"slices"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/bundled"
	"github.com/microsoft/TypeScript/tsc/internal/compiler"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/printer"
	"github.com/microsoft/TypeScript/tsc/internal/tsoptions"
	"github.com/microsoft/TypeScript/tsc/internal/tspath"
	"github.com/microsoft/TypeScript/tsc/internal/vfs/vfstest"
)

func TestTypefnDeclarationDependencies(t *testing.T) {
	files := map[string]string{
		"/main.ts": `import type { Imported } from "./other";
import type * as External from "./other";
type Dead = { dead: true };
declare const key: unique symbol;
declare const value: { value: true };
typefn Private<T> { return T; }
type Shadow = never;
function outer<Capture>() {
 type Hidden = Capture;
 typefn Escapes<T> { return [Hidden, Capture, T]; }
}
export typefn Public<T> {
 type Shadow = T;
 typefn Local<U> { return [Shadow, U]; }
 if (false) { return Private<[Dead, Imported, External.Imported]>; }
 type result = type({ [key]: typeof value; local: Shadow; method: <U>(x: U) => U });
 match (T) { [head, ...tail] => { return Local<[head, tail, result]>; } _ => { return text.length<"x">; } }
}`,
		"/other.ts":      `export interface Imported { imported: true }`,
		"/tsconfig.json": `{"compilerOptions":{"typefn":{"version":1},"strict":true,"target":"es2022","ignoreDeprecations":"6.0"},"files":["main.ts","other.ts"]}`,
	}
	fs := bundled.WrapFS(vfstest.FromMap(files, tspath.CaseSensitive))
	host := compiler.NewCompilerHost(fs, bundled.LibPath(), nil, nil, nil)
	config, errors := tsoptions.GetParsedCommandLineOfConfigFile("/tsconfig.json", &core.CompilerOptions{}, nil, tsoptions.NewDefaultParseConfigHost(fs), nil)
	if len(errors) != 0 {
		t.Fatal(errors)
	}
	program := compiler.NewProgram(compiler.ProgramOptions{ProgramConfig: compiler.ProgramConfig{Config: config}, ProgramHosts: compiler.ProgramHosts{Host: host}})
	if ds := program.GetSyntacticDiagnostics(t.Context(), nil); len(ds) != 0 {
		t.Fatal(ds)
	}
	if ds := program.GetSemanticDiagnostics(t.Context(), nil); len(ds) != 0 {
		t.Fatal(ds)
	}
	c, done := program.GetTypeChecker(t.Context())
	defer done()
	resolver := c.NewEmitResolver(printer.NewEmitContext())
	file := program.GetSourceFile("/main.ts")
	declaration := file.Statements.Nodes[len(file.Statements.Nodes)-1]
	// Inspecting another module's dependency graph must not consume or schedule
	// its native visibility work. The later emitting walk must still retain it.
	for _, dependency := range resolver.GetTypefnDeclarationDependencies(declaration, false) {
		if len(dependency.AliasesToMakeVisible) != 0 {
			t.Fatalf("inspection queued visibility: %+v", dependency)
		}
	}
	for _, statement := range file.Statements.Nodes {
		if statement.Name() != nil && (statement.Name().Text() == "Dead" || statement.Name().Text() == "Private") && resolver.IsDeclarationVisible(statement) {
			t.Fatalf("inspection published private declaration %s", statement.Name().Text())
		}
	}
	dependencies := resolver.GetTypefnDeclarationDependencies(declaration, true)
	var names []string
	for _, dependency := range dependencies {
		if dependency.Accessibility != printer.SymbolAccessibilityAccessible {
			t.Fatalf("unresolved dependency: %+v", dependency)
		}
		for _, statement := range dependency.AliasesToMakeVisible {
			switch statement.Kind {
			case ast.KindImportDeclaration:
				names = append(names, statement.AsImportDeclaration().ImportClause.AsImportClause().NamedBindings.Kind.String())
			case ast.KindVariableStatement:
				names = append(names, statement.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0].Name().Text())
			default:
				names = append(names, statement.Name().Text())
			}
		}
	}
	slices.Sort(names)
	want := []string{"Dead", "KindNamedImports", "KindNamespaceImport", "Private", "key", "value"}
	if !slices.Equal(names, want) {
		t.Fatalf("retained dependencies: %v, want %v", names, want)
	}
	// A second walk still resolves all lexical references, but native visibility
	// already owns the retained declarations and must not queue them a second time.
	for _, dependency := range resolver.GetTypefnDeclarationDependencies(declaration, true) {
		if dependency.Accessibility != printer.SymbolAccessibilityAccessible || len(dependency.AliasesToMakeVisible) != 0 {
			t.Fatalf("unstable dependency closure: %+v", dependency)
		}
	}
	outer := file.Statements.Nodes[len(file.Statements.Nodes)-2]
	escaped := outer.AsFunctionDeclaration().Body.AsBlock().Statements.Nodes[1]
	dependencies = resolver.GetTypefnDeclarationDependencies(escaped, true)
	if len(dependencies) != 2 {
		t.Fatalf("escaped dependencies: %+v", dependencies)
	}
	for _, dependency := range dependencies {
		if dependency.Accessibility != printer.SymbolAccessibilityNotAccessible {
			t.Fatalf("local dependency incorrectly nameable: %+v", dependency)
		}
	}

}
