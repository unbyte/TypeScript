// @target: esnext
// @strict: true

typefn Arithmetic<> {
    return [
        number.add<9007199254740992, 1>, number.subtract<0, 3>, number.multiply<-2, 4>,
        number.divide<3, 2>, number.remainder<-5, 3>, number.negate<-1>, number.trunc<-1.75>,
        number.lessThan<1, 2>, number.fromDecimal<"+007.5e1">, number.fromBigInt<9007199254740991n>,
        bigint.fromDecimal<"00012345678901234567890">, bigint.fromNumber<-42>,
        bigint.add<3n, 4n>, bigint.subtract<2n, 5n>, bigint.multiply<-3n, 4n>, bigint.negate<7n>,
        bigint.divide<-7n, 3n>, bigint.remainder<-7n, 3n>, bigint.lessThan<0n, 1n>
    ];
}
declare const math: Arithmetic<>;
const numbers: [9007199254740992, -3, -8, 1.5, -2, 1, -1, true, 75, 9007199254740991,
    12345678901234567890n, -42n, 7n, -3n, -12n, -7n, -2n, -1n, true] = math;

typefn LexicalFacts<S> {
    type Cursor = { source: S; offset: 0 };
    return [
        string.length<S>, string.read<Cursor>, string.slice<S, 0, 2>, string.find<Cursor, "a">,
        string.match<Cursor, { run: [["😀", "😀"]]; max: 1 }>, string.split<"a,,b,", ",">,
        string.codePoint<"\ud800">, string.fromCodePoint<55296>,
        string.match<Cursor, "b">, string.find<Cursor, "b">
    ];
}
declare const text: LexicalFacts<"😀a">;
const lexical: [3, ["hit", "😀", {source: "😀a"; offset: 2}], "😀",
    ["hit", {source: "😀a"; start: 2; end: 3}], ["hit", "😀", {source: "😀a"; offset: 2}],
    ["a", "", "b", ""], 55296, "\ud800", ["miss"], ["miss"]] = text;

typefn Forms<T> {
    type Copy = intersection.fromTuple<intersection.parts<T>>;
    type Arr = array.compose<array.decompose<readonly T[]>>;
    type Template = template.compose<template.decompose<`id-${number}`>>;
    return [Copy, Arr, Template, type.kind<T>, type.kind<any>, type.kind<unknown>, type.kind<never>,
        type.kind<[T]>, type.kind<T[]>, type.baseOfLiteral<"ok">,
        type.isIdentical<1, number>, type.isAssignableTo<1, number>, tuple.count<[1, 2?]>];
}
declare const forms: Forms<{a: 1} & {b: 2}>;
const ordinary: [{a: 1; b: 2}, readonly {a: 1; b: 2}[], `id-${number}`, "intersection",
    "any", "unknown", "never", "tuple", "array", string, false, true, 1 | 2] = forms;

enum Choice { First = "x", Second = 5 }
typefn Meta<E> { return enumType.members<E>; }
declare const enumInfo: Meta<typeof Choice>;
const enumValue: ["constant", 5] = enumInfo.Second.value;

typefn Yield<T> { return iter.protocol<T, "sync">; }
declare const protocol: Yield<Generator<1, 2, 3>>;
const iterator: [1, 2, 3] = protocol;

type KindKey<T> = T extends { kind: infer K } ? K : never;
typefn Ordered<T> { return union.toTuple<T, KindKey<_>>; }
declare const ordered: Ordered<{kind:"z"} | {kind:"a"}>;
const sort: [{kind:"a"}, {kind:"z"}] = ordered;

typefn Edit<Base, Name> {
    type Old = object.getField<Base, Name>;
    type Without = object.removeField<Base, Name>;
    type Absent = object.hasField<Without, Name>;
    type Restored = object.addField<Without, Name, Old>;
    type WithIndex = object.withIndexes<Restored, {key: string; type: number; readonly: true}>;
    return [Absent, Restored, object.indexes<WithIndex>];
}
declare const edits: Edit<{n: 1}, "n">;
const edited: [false, {n: 1}, {key: string; type: number; readonly: true}] = edits;
