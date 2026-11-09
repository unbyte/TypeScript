// Package typefntest exercises typefn through real compiler projects, not a
// transpiler or an independent model of the DSL's result types.
package typefntest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/tspath"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/bundled"
	"github.com/microsoft/TypeScript/tsc/internal/checker"
	"github.com/microsoft/TypeScript/tsc/internal/compiler"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/tsoptions"
	"github.com/microsoft/TypeScript/tsc/internal/vfs/vfstest"
)

type fixture struct {
	Files       map[string]string    `json:"files"` // virtual filename -> fixture source filename
	Options     map[string]any       `json:"options"`
	Diagnostics []wantDiagnostic     `json:"diagnostics"`
	Outputs     map[string]string    `json:"outputs"` // exact output text, including newlines
	EmitSkipped bool                 `json:"emitSkipped"`
	Consumer    *fixture             `json:"consumer"` // fresh checker with emitted declarations and consumer files only
	emitOptions compiler.EmitOptions // API-only modes exercised by declaration tests
	beforeCheck func(testing.TB, *compiler.Program)
	beforeEmit  func(testing.TB, *compiler.Program)
	sourceText  map[string]string // Go-generated contents for entries in Files
}

type wantDiagnostic struct {
	Phase   string           `json:"phase"`
	Code    int32            `json:"code"`
	File    string           `json:"file"`
	At      string           `json:"at"` // unique source substring beginning at the diagnostic
	Message string           `json:"message"`
	Related []wantDiagnostic `json:"related,omitempty"`
}

type diagnostic struct {
	phase string
	value *ast.Diagnostic
}

func TestIntegration(t *testing.T) {
	entries, err := filepath.Glob("testdata/*.json")
	if err != nil || len(entries) == 0 {
		t.Fatalf("load fixtures: %v (%d found)", err, len(entries))
	}
	for _, entry := range entries {
		t.Run(strings.TrimSuffix(filepath.Base(entry), ".json"), func(t *testing.T) {
			t.Parallel()
			data, err := os.ReadFile(entry)
			if err != nil {
				t.Fatal(err)
			}
			var f fixture
			decoder := json.NewDecoder(strings.NewReader(string(data)))
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&f); err != nil {
				t.Fatal(err)
			}
			runFixture(t, f)
		})
	}
}

func runFixture(t testing.TB, f fixture) {
	t.Helper()
	files, roots := loadFixtureFiles(t, f)
	runFixtureSources(t, f, files, roots)
}

func loadFixtureFiles(t testing.TB, f fixture) (map[string]string, []string) {
	t.Helper()
	files := make(map[string]string, len(f.Files)+1)
	var roots []string
	for name, source := range f.Files {
		if strings.Contains(name, "..") || strings.HasPrefix(name, "/") {
			t.Fatalf("fixture filename must be relative: %q", name)
		}
		text, generated := f.sourceText[name]
		if !generated {
			data, err := os.ReadFile(filepath.Join("testdata", source))
			if err != nil {
				t.Fatal(err)
			}
			text = string(data)
		}
		files["/project/"+name] = text
		roots = append(roots, name)
	}
	return files, roots
}

func runFixtureSources(t testing.TB, f fixture, files map[string]string, roots []string) {
	t.Helper()
	var emitted map[string]string
	slices.Sort(roots)
	options := map[string]any{"strict": true, "target": "es2022", "module": "esnext", "outDir": "out", "ignoreDeprecations": "6.0"}
	for key, value := range f.Options {
		options[key] = value
	}
	config, err := json.Marshal(map[string]any{"compilerOptions": options, "files": roots})
	if err != nil {
		t.Fatal(err)
	}
	files["/project/tsconfig.json"] = string(config)
	fs := bundled.WrapFS(vfstest.FromMap(files, tspath.CaseSensitive))
	host := compiler.NewCompilerHost(fs, bundled.LibPath(), nil, nil, nil)
	parsed, configErrors := tsoptions.GetParsedCommandLineOfConfigFile("/project/tsconfig.json", &core.CompilerOptions{}, nil, tsoptions.NewDefaultParseConfigHost(fs), nil)
	var got []diagnostic
	add := func(phase string, ds []*ast.Diagnostic) {
		for _, d := range ds {
			got = append(got, diagnostic{phase, d})
		}
	}
	add("config", configErrors)
	if parsed != nil {
		add("config", parsed.GetConfigFileParsingDiagnostics())
	}
	if len(got) == 0 {
		p := compiler.NewProgram(compiler.ProgramOptions{ProgramConfig: compiler.ProgramConfig{Config: parsed}, ProgramHosts: compiler.ProgramHosts{Host: host}})
		if f.beforeCheck != nil {
			f.beforeCheck(t, p)
		}
		add("syntax", p.GetSyntacticDiagnostics(t.Context(), nil))
		add("program", p.GetProgramDiagnostics())
		// Malformed source must never be evaluated by the harness.
		if len(got) == 0 {
			add("global", p.GetGlobalDiagnostics(t.Context()))
			add("semantic", p.GetSemanticDiagnostics(t.Context(), nil))
		}
		if len(got) == 0 && len(f.Diagnostics) == 0 {
			c, done := p.GetTypeChecker(t.Context())
			for name := range f.Files {
				file := p.GetSourceFile(tspath.ToRootedFilePath(name, "/project"))
				if file == nil {
					continue
				}
				var visit func(*ast.Node) bool
				visit = func(node *ast.Node) bool {
					if ast.IsTypeReferenceNode(node) && ast.IsIdentifier(node.AsTypeReferenceNode().TypeName) && node.AsTypeReferenceNode().TypeName.Text() == "Assert" {
						actual := c.GetTypeAtLocation(node)
						if actual.Flags() != checker.TypeFlagsBooleanLiteral || actual.AsLiteralType().Value() != true {
							t.Errorf("%s@%d: Assert must resolve to literal true, got %s (%s)", name, node.Pos(), c.TypeToString(actual), actual.Flags())
						}
					}
					node.ForEachChild(visit)
					return false
				}
				visit(file.AsNode())
			}
			done()
		}
		if f.Outputs != nil || f.Consumer != nil {
			if f.beforeEmit != nil {
				f.beforeEmit(t, p)
			}
			outputs := map[string]string{}
			var outputMu sync.Mutex
			emitOptions := f.emitOptions
			emitOptions.WriteFile = func(name tspath.RootedFilePath, text string, _ *compiler.WriteFileData) error {
				outputMu.Lock()
				defer outputMu.Unlock()
				outputs[strings.TrimPrefix(string(name), "/project/")] = text
				return nil
			}
			result := p.Emit(t.Context(), emitOptions)
			add("emit", result.Diagnostics)
			if result.EmitSkipped != f.EmitSkipped {
				t.Errorf("emit skipped: got %v, want %v", result.EmitSkipped, f.EmitSkipped)
			}
			emitted = outputs
			if f.Outputs != nil && !reflect.DeepEqual(outputs, f.Outputs) {
				t.Errorf("outputs:\n got: %#v\nwant: %#v", outputs, f.Outputs)
			}
		}
	}
	if err := compareDiagnostics(got, f.Diagnostics, files); err != nil {
		t.Fatal(err)
	}
	if f.Consumer != nil {
		if len(emitted) == 0 {
			t.Fatal("declaration consumer requires emitted producer output")
		}
		consumerFiles, consumerRoots := loadFixtureFiles(t, *f.Consumer)
		declarations := 0
		for name, text := range emitted {
			if !strings.HasSuffix(name, ".d.ts") && !strings.HasSuffix(name, ".d.mts") && !strings.HasSuffix(name, ".d.cts") {
				continue
			}
			path := "/project/" + name
			if _, exists := consumerFiles[path]; exists {
				t.Fatalf("consumer overwrites emitted declaration %s", name)
			}
			consumerFiles[path] = text
			declarations++
		}
		if declarations == 0 {
			t.Fatal("producer emitted no declaration files")
		}
		// The consumer reuses no producer file map, host, Program, checker, or
		// source redirect.
		// Imports must resolve against the actual declaration bytes emitted above.
		runFixtureSources(t, *f.Consumer, consumerFiles, consumerRoots)
	}
}

// Match every diagnostic, including phase, code, file and source position.
// Unexpected diagnostics and missing expectations both fail the fixture.
func compareDiagnostics(got []diagnostic, want []wantDiagnostic, files map[string]string) error {
	var actual []string
	for _, d := range got {
		name := ""
		if d.value.File() != nil {
			name = strings.TrimPrefix(string(d.value.File().FileName()), "/project/")
		}
		actual = append(actual, fmt.Sprintf("%s TS%d %s@%d: %s", d.phase, d.value.Code(), name, d.value.Pos(), d.value.String()))
	}
	if len(got) != len(want) {
		return fmt.Errorf("diagnostic count: got %d, want %d\n%s", len(got), len(want), strings.Join(actual, "\n"))
	}
	used := make([]bool, len(got))
	for _, w := range want {
		pos := -1
		if w.At != "" {
			source := files["/project/"+w.File]
			if strings.Count(source, w.At) != 1 {
				return fmt.Errorf("diagnostic marker %q must occur once in %s", w.At, w.File)
			}
			pos = strings.Index(source, w.At)
		}
		found := false
		for i, g := range got {
			file := ""
			if g.value.File() != nil {
				file = strings.TrimPrefix(string(g.value.File().FileName()), "/project/")
			}
			if !used[i] && g.phase == w.Phase && g.value.Code() == w.Code && file == w.File &&
				(pos < 0 || g.value.Pos() == pos) && strings.Contains(g.value.String(), w.Message) {
				if w.Related != nil {
					var related []diagnostic
					for _, d := range g.value.RelatedInformation() {
						related = append(related, diagnostic{g.phase, d})
					}
					if err := compareDiagnostics(related, w.Related, files); err != nil {
						return fmt.Errorf("related information: %w", err)
					}
				}
				used[i], found = true, true
				break
			}
		}
		if !found {
			return fmt.Errorf("missing diagnostic %#v\n%s", w, strings.Join(actual, "\n"))
		}
	}
	return nil
}
