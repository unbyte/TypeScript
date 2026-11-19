// @strict: true
// @target: esnext

type Letter = "a" | "b" | "c" | "d" | "e" | "f" | "g" | "h" | "i" | "j" | "k" | "l" | "m"
    | "n" | "o" | "p" | "q" | "r" | "s" | "t" | "u" | "v" | "w" | "x" | "y" | "z";

typefn Scan<S, Index, Allowed> {
    type At = Index;
    while (true) {
        match string.read<S, At> {
            case undefined: { return At; }
            case [Allowed, infer Next]: { At = Next; }
            else: { return At; }
        }
    }
}
typefn Skip<S, Index> { return Scan<S, Index, " ">; }
typefn Expect<S, Index, Text> {
    match string.match<S, Skip<S, Index>, Text> {
        case undefined { fail<`expected ${Text & string}`>; }
        case [infer Spelling, infer Next] { return Next; }
    }
}
typefn Word<S, Index> {
    type Start = Skip<S, Index>;
    type End = Scan<S, Start, Letter>;
    if (Start === End) { fail<"expected a name">; }
    return [string.slice<S, Start, End>, End];
}
typefn Scalar<S, Index> {
    type [Name, Next] = Word<S, Index>;
    match Name {
        case "number" { return [number, Next]; }
        case "string" { return [string, Next]; }
        else { fail<"unknown field type">; }
    }
}
typefn RecordFields<S, Index> {
    type At = Index;
    type Out = {};
    while (true) {
        match string.match<S, Skip<S, At>, "}"> {
            case undefined {} // Read the next field.
            case [infer Close, infer End] { return [Out, End]; }
        }
        type [Name, AfterName] = Word<S, At>;
        type [Value, AfterValue] = Annotation<S, Expect<S, AfterName, ":">>;
        Out = object.addField<Out, Name, { type: Value }>;
        match string.match<S, Skip<S, AfterValue>, ";"> {
            case undefined { return [Out, Expect<S, AfterValue, "}">]; }
            case [infer Sep, infer Next] { At = Next; }
        }
    }
}
typefn Primary<S, Index> {
    match string.match<S, Skip<S, Index>, "{"> {
        case undefined { return Scalar<S, Index>; }
        case [infer Open, infer At] { return RecordFields<S, At>; }
    }
}
typefn Annotation<S, Index> {
    type [Value, At] = Primary<S, Index>;
    while (true) {
        match string.match<S, Skip<S, At>, "[]"> {
            case undefined { return [Value, At]; }
            case [infer Suffix, infer Next] { Value = Value[]; At = Next; }
        }
    }
}
typefn Parent<S> {
    type [Name, AfterName] = Word<S, Expect<S, 0, "root(parent ">>;
    type [Value, AfterValue] = Annotation<S, Expect<S, AfterName, ":">>;
    type End = Expect<S, AfterValue, ")">;
    match string.read<S, Skip<S, End>> {
        case undefined { return object.addField<{}, Name, { type: Value }>; }
        else { fail<"trailing input">; }
    }
}
declare const a: Parent<"root(parent account: { id: number; roles: string[] })">;
declare const b: Parent<"root(parent owner: string)">;
const id: number = a.account.id;
const role: string = a.account.roles[0];
const owner: string = b.owner;
// @ts-expect-error independent calculations have different named fields
b.account;
