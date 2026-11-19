package tsctests

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/execute"
	"github.com/microsoft/TypeScript/tsc/internal/execute/tsc"
)

func TestTypeFunctionPublishedPackage(t *testing.T) {
	t.Parallel()
	const source = `typefn PrivateCount<T> {
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
	run("--declaration", "--emitDeclarationOnly", "--isolatedDeclarations", "--outDir", "node_modules/words", "library.ts")
	sys.removeNoError("/project/library.ts")
	output := run("--noEmit", "--traceTypeFunctions", "consumer.ts")
	if !strings.Contains(output, "split") || !strings.Contains(output, "three") {
		t.Fatalf("missing type function trace\n%s", output)
	}
}
