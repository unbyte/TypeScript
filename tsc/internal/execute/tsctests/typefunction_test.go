package tsctests

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/execute"
	"github.com/microsoft/TypeScript/tsc/internal/execute/tsc"
)

func TestTypeFunctionIncrementalDependency(t *testing.T) {
	t.Parallel()
	const library = "/project/lib/index.ts"
	sys := newTestSys(&tscInput{cwd: "/project", files: FileMap{
		"/project/lib/tsconfig.json": `{"compilerOptions":{"composite":true,"outDir":"dist"},"files":["index.ts"]}`,
		library: `typefn Adjust<N> { return number.add<N, 1>; }
export typefn Result<N> { return Adjust<N>; }`,
		"/project/app/tsconfig.json": `{"compilerOptions":{"composite":true},"references":[{"path":"../lib"}],"files":["index.ts"]}`,
		"/project/app/index.ts": `import type { Result } from "../lib/dist/index";
declare const x: Result<1>;
const expected: 2 = x;`,
	}}, false)
	build := func(wantError bool) {
		t.Helper()
		sys.currentWrite.Reset()
		result := execute.CommandLine(t.Context(), sys, []string{"--build", "app", "--pretty", "false"}, sys)
		output := sys.currentWrite.String()
		if wantError {
			if result.Status == tsc.ExitStatusSuccess || !strings.Contains(output, "Type '0' is not assignable to type '2'") {
				t.Fatalf("expected the consumer to see the changed private helper\n%s", output)
			}
		} else if result.Status != tsc.ExitStatusSuccess {
			t.Fatalf("project build failed\n%s", output)
		}
	}
	build(false)
	sys.replaceFileText(library, "number.add", "number.subtract")
	build(true)
	sys.replaceFileText(library, "number.subtract", "number.add")
	build(false)
}

func TestTypeFunctionPublishedPackage(t *testing.T) {
	t.Parallel()
	const source = `/** @internal */
typefn PrivateCount<T> {
    type Words = string.split<T, " ">;
    log<"split", Words>;
    return tuple.count<Words>;
}
export typefn Words<T> { return PrivateCount<T>; }`
	const consumer = `import type { Words } from "words";
type PrivateCount<T> = never;
declare const n: Words<"one two three">;
const three: 3 = n;
// @ts-expect-error
const four: 4 = n;`
	sys := newTestSys(&tscInput{cwd: "/project", files: FileMap{
		"/project/library.ts":                      source,
		"/project/consumer.ts":                     consumer,
		"/project/node_modules/words/package.json": `{"name":"words","types":"library.d.ts"}`,
	}}, false)
	run := func(args ...string) string {
		t.Helper()
		sys.currentWrite.Reset()
		common := []string{"--ignoreConfig", "--strict", "--target", "esnext", "--module", "esnext", "--moduleResolution", "bundler", "--pretty", "false"}
		if result := execute.CommandLine(t.Context(), sys, append(common, args...), sys); result.Status != tsc.ExitStatusSuccess {
			t.Fatalf("tsc %v failed\n%s", args, sys.currentWrite.String())
		}
		return sys.currentWrite.String()
	}
	run("--stripInternal", "--declaration", "--emitDeclarationOnly", "--isolatedDeclarations", "--outDir", "node_modules/words", "library.ts")
	sys.removeNoError("/project/library.ts")
	output := run("--noEmit", "--skipLibCheck", "--traceTypeFunctions", "consumer.ts")
	if strings.Count(output, `"split"`) != 1 || !strings.Contains(output, "three") {
		t.Fatalf("missing type function trace\n%s", output)
	}
}
