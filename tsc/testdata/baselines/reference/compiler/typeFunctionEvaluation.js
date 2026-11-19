//// [tests/cases/compiler/typeFunctionEvaluation.ts] ////

//// [library.ts]
export typefn Echo<T> {
    return T;
}

export typefn CountWords<S extends string> {
    type Words = string.split<S, " ">;
    return tuple.count<Words>;
}

//// [consumer.ts]
import type { Echo, CountWords } from "./library";

type Later<T> = Echo<{ value: T }>;
declare const first: Later<"first">;
declare const second: Later<42>;
const a: "first" = first.value;
const b: 42 = second.value;
const count: CountWords<"a b c"> = 3;
// @ts-expect-error a literal count must survive substitution
const badCount: CountWords<"a b c"> = 4;


//// [library.js]
export {};
//// [consumer.js]
const a = first.value;
const b = second.value;
const count = 3;
// @ts-expect-error a literal count must survive substitution
const badCount = 4;
export {};


//// [library.d.ts]
export typefn Echo<T> {
    return T;
}
export typefn CountWords<S extends string> {
    type Words = string.split<S, " ">;
    return tuple.count<Words>;
}
//// [consumer.d.ts]
export {};
