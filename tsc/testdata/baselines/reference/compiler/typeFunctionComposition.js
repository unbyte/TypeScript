//// [tests/cases/compiler/typeFunctionComposition.ts] ////

//// [typeFunctionComposition.ts]
export typefn Flags<Base extends object, Names extends readonly string[]> {
    type Out = Base;
    for (type Name of Names) {
        Out = object.setField<Out, Name, { type: boolean; readonly: true; optional: true }>;
    }
    return Out;
}
declare const flags: Flags<{ id: 1 }, ["debug", "trace"]>;
const id: 1 = flags.id;
const debug: boolean | undefined = flags.debug;
// @ts-expect-error field must be readonly
flags.trace = false;

type Box<T> = { value: T };
typefn Apply<F<_>, X> { return F<X>; }
declare const box: Apply<Box<_>, "boxed">;
const boxed: "boxed" = box.value;

typefn WrapResult<F> {
    type Parts = fn.decompose<F>;
    type Call = Parts["calls"][0];
    return fn.compose<{
        members: {};
        calls: [{ parameters: Call["parameters"]; result: { kind: "return"; type: [Call["result"]["type"]] } }];
        constructs: [];
    }>;
}
export declare const wrapped: WrapResult<(id: number) => string>;
const reply: [string] = wrapped(42);
// @ts-expect-error parameter type must survive recomposition
wrapped("wrong");

typefn ExtractId<T> {
    match T {
        case { id: infer I } { return I; }
        else { return never; }
    }
}
declare const ref: ExtractId<{ id: 42 }>;
const n: 42 = ref;

function capture<Outer>(v: Outer) {
    type Pair<T> = [Outer, T];
    declareValue<Apply<Pair<_>, 1>>();
    return null! as Apply<Pair<_>, 1>;
}
declare function declareValue<T>(): void;
const captured: [string, 1] = capture<string>("a");


//// [typeFunctionComposition.js]
const id = flags.id;
const debug = flags.debug;
// @ts-expect-error field must be readonly
flags.trace = false;
const boxed = box.value;
const reply = wrapped(42);
// @ts-expect-error parameter type must survive recomposition
wrapped("wrong");
const n = ref;
function capture(v) {
    declareValue();
    return null;
}
const captured = capture("a");
export {};


//// [typeFunctionComposition.d.ts]
export typefn Flags<Base extends object, Names extends readonly string[]> {
    type Out = Base;
    for (type Name of Names) {
        Out = object.setField<Out, Name, {
            type: boolean;
            readonly: true;
            optional: true;
        }>;
    }
    return Out;
}
typefn WrapResult<F> {
    type Parts = fn.decompose<F>;
    type Call = Parts["calls"][0];
    return fn.compose<{
        members: {};
        calls: [{
            parameters: Call["parameters"];
            result: {
                kind: "return";
                type: [Call["result"]["type"]];
            };
        }];
        constructs: [];
    }>;
}
export declare const wrapped: WrapResult<(id: number) => string>;
export {};
