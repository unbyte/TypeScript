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
typefn PrivateCount<T, Sep> {
    type Words = string.split<T, Sep>;
    log<"split", Words>;
    return tuple.count<Words>;
}
/** @internal */
type Space = " ";
export typefn Words<T, Sep = Space> { return PrivateCount<T, Sep>; }
/** @internal */
type Box<T> = { value: T };
export typefn Wrapped<T, F<_> = Box<_>> { return F<T>; }
export typefn Twice<N> { if (N is number) { return -(N + N); } return undefined; }
typefn Functions<N> { return [<T>(arg: T) => [T, N + 1]]; }
export declare const withFour: Functions<3>[0];
typefn Empty<> { return 0; }
export type Zero = Empty<>;
typefn Tick<> { log<"tick">; return 4; }
export typefn Select<T> { return T extends true ? { value: Tick<> } : never; }`
	const inferred = `typefn Increment<N> { return N + 1; }
type Hidden<N> = { value: Increment<N> };
typefn Functions<N> { return [<T>(arg: T) => [T, Hidden<N>]]; }
declare const functions: Functions<3>;
export const transform = functions[0];`
	const consumer = `import { transform } from "words/inferred";
const transformed: ["x", { value: 4 }] = transform("x" as const);
// @ts-expect-error the declaration preserves the computed field
const badField: 3 = transform(0)[1].value;
import type { Select, Twice, Words, Wrapped, Zero } from "words";
import { withFour } from "words";
declare const selected: Select<true>;
const tick: 4 = selected.value;
const pair: ["x", 4] = withFour("x" as const);
// @ts-expect-error the emitted generic signature preserves the completed literal
const wrong: ["x", 3] = withFour("x" as const);
type PrivateCount<T> = never;
declare const twice: Twice<4>;
const negativeEight: -8 = twice;
// @ts-expect-error a numeric literal result retains its value in the consumer
const wrongTwice: 8 = twice;
declare const wrapped: Wrapped<7>;
const seven: 7 = wrapped.value;
// @ts-expect-error the default constructor retains the supplied literal
const wrongValue: 8 = wrapped.value;
declare const zero: Zero;
const verifiedZero: 0 = zero;
// @ts-expect-error the published zero-argument application returns literal 0
const wrongZero: 1 = zero;
declare const n: Words<"one two three">;
const three: 3 = n;
// @ts-expect-error
const four: 4 = n;`
	sys := newTestSys(&tscInput{cwd: "/project", files: FileMap{
		"/project/library.ts":                      source,
		"/project/inferred.ts":                     inferred,
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
	run("--declaration", "--emitDeclarationOnly", "--outDir", "node_modules/words", "inferred.ts")
	declaration := sys.readFileNoError("/project/node_modules/words/inferred.d.ts")
	if strings.Contains(declaration, "typefn") || strings.Contains(declaration, "Hidden") {
		t.Fatalf("inferred declaration contains an unfinished calculation\n%s", declaration)
	}
	sys.removeNoError("/project/library.ts")
	sys.removeNoError("/project/inferred.ts")
	output := run("--noEmit", "--skipLibCheck", "--traceTypeFunctions", "consumer.ts")
	if strings.Count(output, `"tick"`) != 1 {
		t.Fatalf("expected the selected field to execute once\n%s", output)
	}
	if strings.Count(output, `"split"`) != 1 || !strings.Contains(output, "three") {
		t.Fatalf("missing type function trace\n%s", output)
	}
}
