//// [tests/cases/compiler/typeFunctionReturnBoundary.ts] ////

//// [typeFunctionReturnBoundary.ts]
typefn Add<A, B> { return A + B; }
type Hidden<T> = { nested: Add<T, 1> };
interface Container<T> { contents: Hidden<T>; }

typefn WithNext<N> {
    type Next = N + 1;
    return <T>(x: T) => [T, Next];
}
typefn Inline<N> { return <T>(x: T) => [T, N + 1]; }
declare const saved: WithNext<3>;
declare const inline: Inline<8>;
const four: ["x", 4] = saved("x" as const);
const nine: [false, 9] = inline(false as const);

typefn Direct<N> { return <T>(value: T) => T + N; }
typefn Named<N> { return <T>(value: T) => Add<T, N>; }
typefn Aliased<> { type Local = <T>() => Container<T>; return Local; }
typefn Builtin<> { return <S extends string>() => string.split<S, " ">; }
declare const direct: Direct<1>;
declare const named: Named<1>;
declare const aliased: Aliased<>;
declare const builtin: Builtin<>;
direct.recovered; named.recovered; aliased.recovered; builtin.recovered;

// A helper may construct an intermediate signature whose result is later specialized.
typefn Finish<N> { return fn.specialize<Direct<N>, [2]>; }
declare const finished: Finish<3>;
const five: 5 = finished(2);

typefn Echo<T> { return T; }
type Tree<T> = { value: T; children: Tree<T[]> };
declare const tree: Echo<Tree<1>>;
const treeValue: 1[] = tree.children.value;
declare const promise: Echo<Promise<string[]>>;
const standard: Promise<string[]> = promise;

typefn Native<N> { return <T>(arg: T) => T extends string ? [T, N + 1] : T; }
declare const native: Native<4>;
const selected: ["x", 5] = native("x" as const);
const other: 0 = native(0 as const);
typefn Conditional<> { return <T>() => T extends string ? Add<T, 1> : never; }
typefn Mapped<> { return <T>() => { [K in keyof T]: Add<T[K], 1> }; }
typefn Guard<> { return <T>(value: unknown) => value is Add<T, 1>; }
typefn Default<> { return <T, U = Add<T, 1>>() => U; }
declare const conditional: Conditional<>;
declare const mapped: Mapped<>;
declare const guard: Guard<>;
declare const defaults: Default<>;
conditional.recovered; mapped.recovered; guard.recovered; defaults.recovered;

typefn RecordResult<N> { return <T>(arg: T) => { value: T; tag: N + 1 }; }
declare const recordResult: RecordResult<3>;
const record: { value: "x"; tag: 4 } = recordResult("x" as const);

typefn MapResult<N> { return <T>(arg: T) => { readonly [K in keyof T]: [T[K], N + 1] }; }
declare const mapResult: MapResult<3>;
const mappedResult: { readonly z: ["hi", 4] } = mapResult({ z: "hi" } as const);

typefn DefaultResult<N> { return <T extends N + 1 = N + 1>(arg?: T) => T; }
declare const defaultResult: DefaultResult<3>;
const defaultFour: 4 = defaultResult();

type Recursive<N> = { value: Add<N, 1>; next: Recursive<N> | null };
declare const recursive: Echo<Recursive<3>>;
const recursiveFour: 4 | undefined = recursive.next?.next?.value;

class Fluent { private brand = 0; self(): this { return this; } }
declare const fluent: Echo<Fluent>;
const ordinaryClass: Fluent = fluent.self();

type Path<N> = readonly [Add<N, 1>, Path<N>?];
declare const path: Echo<Path<3>>;
const step: 4 | undefined = path[1]?.[0];
// @ts-expect-error the recursive element keeps its computed literal
const wrongStep: 5 | undefined = path[1]?.[1]?.[0];

declare const label: "value";
type Phantom<T> = { label: typeof label };
typefn ThroughArgument<> { return <T>() => Phantom<Add<T, 1>>; }
declare const phantom: ThroughArgument<>;
phantom.recovered;

typefn InferResult<N> { return <T>(arg: T) => T extends { id: infer I } ? [I, N + 1] : never; }
declare const inferResult: InferResult<3>;
const inferredPair: ["x", 4] = inferResult({ id: "x" } as const);
// @ts-expect-error the native conditional still infers the id type
const wrongId: [false, 4] = inferResult({ id: "x" } as const);


//// [typeFunctionReturnBoundary.js]
"use strict";
const four = saved("x");
const nine = inline(false);
direct.recovered;
named.recovered;
aliased.recovered;
builtin.recovered;
const five = finished(2);
const treeValue = tree.children.value;
const standard = promise;
const selected = native("x");
const other = native(0);
conditional.recovered;
mapped.recovered;
guard.recovered;
defaults.recovered;
const record = recordResult("x");
const mappedResult = mapResult({ z: "hi" });
const defaultFour = defaultResult();
const recursiveFour = recursive.next?.next?.value;
class Fluent {
    brand = 0;
    self() { return this; }
}
const ordinaryClass = fluent.self();
const step = path[1]?.[0];
// @ts-expect-error the recursive element keeps its computed literal
const wrongStep = path[1]?.[1]?.[0];
phantom.recovered;
const inferredPair = inferResult({ id: "x" });
// @ts-expect-error the native conditional still infers the id type
const wrongId = inferResult({ id: "x" });
