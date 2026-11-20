//// [tests/cases/compiler/typeFunctionParent.ts] ////

//// [typeFunctionParent.ts]
type Letter = "a" | "b" | "c" | "d" | "e" | "f" | "g" | "h" | "i" | "j" | "k" | "l" | "m"
    | "n" | "o" | "p" | "q" | "r" | "s" | "t" | "u" | "v" | "w" | "x" | "y" | "z";

typefn Scan<S, Index, Allowed> {
    type At = Index;
    while (true) {
        match string.read<S, At> {
            case [Allowed, infer Next]: { At = Next; }
            else: { return At; }
        }
    }
}

typefn Skip<S, Index> { return Scan<S, Index, " ">; }

typefn Expect<S, Index, Text> {
    match string.matchPrefix<S, Skip<S, Index>, Text> {
        case [string, infer Next]: { return Next; }
        else: { fail<`expected ${Text & string}`>; }
    }
}

typefn Word<S, Index> {
    type Start = Skip<S, Index>;
    type End = Scan<S, Start, Letter>;
    if (Start === End) {
        fail<"expected a name">;
    }
    return [string.slice<S, Start, End>, End];
}

typefn RecordFields<S, Index> {
    type At = Skip<S, Index>;
    type Fields = {};
    while (string.matchPrefix<S, At, "}"> === undefined) {
        type [Name, Colon] = Word<S, At>;
        type [Value, End] = Annotation<S, Expect<S, Colon, ":">>;
        Fields = object.addField<Fields, Name, { type: Value }>;
        At = Skip<S, End>;
        if (string.matchPrefix<S, At, ";"> === undefined) {
            break;
        }
        At = Skip<S, At + 1>;
    }
    return [Fields, Expect<S, At, "}">];
}

typefn Primary<S, Index> {
    match string.matchPrefix<S, Skip<S, Index>, "{"> {
        case [string, infer Start]: { return RecordFields<S, Start>; }
        else: {}
    }
    type [Name, End] = Word<S, Index>;
    match Name {
        case "number": { return [number, End]; }
        case "string": { return [string, End]; }
        else: { fail<"unknown field type">; }
    }
}

typefn Annotation<S, Index> {
    type [Value, At] = Primary<S, Index>;
    while (true) {
        match string.matchPrefix<S, Skip<S, At>, "[]"> {
            case [string, infer Next]: {
                Value = Value[];
                At = Next;
            }
            else: { return [Value, At]; }
        }
    }
}

typefn Parent<S> {
    type [Name, Colon] = Word<S, Expect<S, 0, "root(parent ">>;
    type [Value, Close] = Annotation<S, Expect<S, Colon, ":">>;
    type End = Expect<S, Close, ")">;
    if (Skip<S, End> !== string.length<S>) {
        fail<"trailing input">;
    }
    return object.addField<{}, Name, { type: Value }>;
}

declare const a: Parent<"root(parent account: { id: number; roles: string[] })">;
declare const b: Parent<"root(parent owner: string)">;
const id: number = a.account.id;
const role: string = a.account.roles[0];
const owner: string = b.owner;
// @ts-expect-error independent calculations have different named fields
b.account;

declare const nested: Parent<"root(parent data: { empty: {}; entries: { id: number; }[][]; })">;
const nestedId: number = nested.data.entries[0][0].id;
// @ts-expect-error a missing field separator must not end the record early
declare const missingSeparator: Parent<"root(parent data: { a: number b: string })">;
const invalidParent = missingSeparator;


//// [typeFunctionParent.js]
"use strict";
const id = a.account.id;
const role = a.account.roles[0];
const owner = b.owner;
// @ts-expect-error independent calculations have different named fields
b.account;
const nestedId = nested.data.entries[0][0].id;
const invalidParent = missingSeparator;
