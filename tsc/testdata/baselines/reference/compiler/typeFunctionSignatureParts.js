//// [tests/cases/compiler/typeFunctionSignatureParts.ts] ////

//// [typeFunctionSignatureParts.ts]
typefn Rebuild<F> { return fn.compose<fn.decompose<F>>; }

interface Factory {
    tag: "factory";
    <T>(x: T | undefined): x is T;
    (x: unknown, kind: "text"): x is string;
    new (name: string): { name: string };
}

declare const factory: Rebuild<Factory>;
declare const maybe: number | undefined;
declare const input: unknown;
if (factory(maybe)) { const n: number = maybe; }
if (factory(input, "text")) { const s: string = input; }
const tag: "factory" = factory.tag;
const made: string = new factory("example").name;

type Shadow = <T>(x: T) => <T>(y: T) => [typeof x, T extends string ? 1 : 2];
declare const nested: Rebuild<Shadow>;
const nestedText: [7, 1] = nested<7>(7)<"x">("x");
const nestedBool: [7, 2] = nested<7>(7)<boolean>(false);

typefn Fresh<> {
    type Item = fn.ref<"item">;
    return fn.compose<{
        members: {};
        constructs: [];
        calls: [{
            typeParameters: [{ name: "T"; key: "item"; const: true }];
            parameters: [{ name: "value"; type: Item }];
            result: { kind: "return"; type: Item };
        }];
    }>;
}
declare const fresh: Fresh<>;
const freshObject: { readonly a: 1 } = fresh({a: 1});

typefn Specialized<F, Args> { return fn.specialize<F, Args>; }
declare const specific: Specialized<<T extends string, R = T>(v: T) => R, ["a"]>;
const result: "a" = specific("a");
// @ts-expect-error rejects a different argument after specialization
specific("b");

typefn Transfer<F> {
    type P = fn.decompose<F>;
    type S = P["calls"][0];
    type Original = S["typeParameters"][0]["ref"];
    type New = fn.ref<"destination">;
    type Result = fn.rebind<S["result"], [{ from: Original; to: New }]>;
    return fn.compose<{
        members: {}; constructs: []; calls: [{
            typeParameters: [{key: "destination"; name: "New"}];
            parameters: [{type: New}]; result: Result;
        }];
    }>;
}
declare const moved: Transfer<<T>(x: T) => { list: [T] }>;
const transferred: { list: [7] } = moved<7>(7);

typefn AddContext<F> {
    type Parts = fn.decompose<F>;
    type Calls = [];
    for (type Call of Parts["calls"]) {
        type Params = [{name: "context"; key: "ctx"; type: object}, ...Call["parameters"]];
        type Edited = object.setField<Call, "parameters", { type: Params }>;
        Calls = [...Calls, Edited];
    }
    return fn.compose<object.setField<Parts, "calls", { type: Calls }>>;
}
declare const contextual: AddContext<Factory>;
if (contextual({}, input, "text")) { const narrowed: string = input; }
// @ts-expect-error the added context is required
contextual(maybe);
const preserved: string = new contextual("a").name;


//// [typeFunctionSignatureParts.js]
"use strict";
if (factory(maybe)) {
    const n = maybe;
}
if (factory(input, "text")) {
    const s = input;
}
const tag = factory.tag;
const made = new factory("example").name;
const nestedText = nested(7)("x");
const nestedBool = nested(7)(false);
const freshObject = fresh({ a: 1 });
const result = specific("a");
// @ts-expect-error rejects a different argument after specialization
specific("b");
const transferred = moved(7);
if (contextual({}, input, "text")) {
    const narrowed = input;
}
// @ts-expect-error the added context is required
contextual(maybe);
const preserved = new contextual("a").name;


//// [typeFunctionSignatureParts.d.ts]
typefn Rebuild<F> {
    return fn.compose<fn.decompose<F>>;
}
interface Factory {
    tag: "factory";
    <T>(x: T | undefined): x is T;
    (x: unknown, kind: "text"): x is string;
    new (name: string): {
        name: string;
    };
}
declare const factory: Rebuild<Factory>;
declare const maybe: number | undefined;
declare const input: unknown;
declare const tag: "factory";
declare const made: string;
type Shadow = <T>(x: T) => <T>(y: T) => [typeof x, T extends string ? 1 : 2];
declare const nested: Rebuild<Shadow>;
declare const nestedText: [7, 1];
declare const nestedBool: [7, 2];
typefn Fresh<> {
    type Item = fn.ref<"item">;
    return fn.compose<{
        members: {};
        constructs: [];
        calls: [
            {
                typeParameters: [{
                    name: "T";
                    key: "item";
                    const: true;
                }];
                parameters: [{
                    name: "value";
                    type: Item;
                }];
                result: {
                    kind: "return";
                    type: Item;
                };
            }
        ];
    }>;
}
declare const fresh: Fresh<>;
declare const freshObject: {
    readonly a: 1;
};
typefn Specialized<F, Args> {
    return fn.specialize<F, Args>;
}
declare const specific: Specialized<<T extends string, R = T>(v: T) => R, ["a"]>;
declare const result: "a";
typefn Transfer<F> {
    type P = fn.decompose<F>;
    type S = P["calls"][0];
    type Original = S["typeParameters"][0]["ref"];
    type New = fn.ref<"destination">;
    type Result = fn.rebind<S["result"], [{
        from: Original;
        to: New;
    }]>;
    return fn.compose<{
        members: {};
        constructs: [];
        calls: [
            {
                typeParameters: [{
                    key: "destination";
                    name: "New";
                }];
                parameters: [{
                    type: New;
                }];
                result: Result;
            }
        ];
    }>;
}
declare const moved: Transfer<<T>(x: T) => {
    list: [T];
}>;
declare const transferred: {
    list: [7];
};
typefn AddContext<F> {
    type Parts = fn.decompose<F>;
    type Calls = [];
    for (type Call of Parts["calls"]) {
        type Params = [{
            name: "context";
            key: "ctx";
            type: object;
        }, ...Call["parameters"]];
        type Edited = object.setField<Call, "parameters", {
            type: Params;
        }>;
        Calls = [...Calls, Edited];
    }
    return fn.compose<object.setField<Parts, "calls", {
        type: Calls;
    }>>;
}
declare const contextual: AddContext<Factory>;
declare const preserved: string;
