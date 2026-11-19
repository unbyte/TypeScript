// @strict: true
// @target: esnext

typefn Skip<C> {
    match string.match<C, { run: [[" ", " "]] }> {
        case ["hit", infer Spaces, infer Next] { return Next; }
        else { return C; }
    }
}
typefn Expect<C, Text> {
    match string.match<Skip<C>, Text> {
        case ["hit", infer Spelling, infer Next] { return Next; }
        else { fail<`expected ${Text & string}`>; }
    }
}
typefn Word<C> {
    match string.match<Skip<C>, { run: [["a", "z"]] }> {
        case ["hit", infer Text, infer Next] { return [Text, Next]; }
        else { fail<"expected a name">; }
    }
}
typefn Scalar<C> {
    type [Name, Next] = Word<C>;
    match Name {
        case "number" { return [number, Next]; }
        case "string" { return [string, Next]; }
        else { fail<"unknown field type">; }
    }
}
typefn RecordFields<C> {
    type At = C;
    type Out = {};
    while (true) {
        match string.match<Skip<At>, "}"> {
            case ["hit", infer Close, infer End] { return [Out, End]; }
        }
        type [Name, AfterName] = Word<At>;
        type [Value, AfterValue] = Annotation<Expect<AfterName, ":">>;
        Out = object.addField<Out, Name, { type: Value }>;
        match string.match<Skip<AfterValue>, ";"> {
            case ["hit", infer Sep, infer Next] { At = Next; }
            else { return [Out, Expect<AfterValue, "}">]; }
        }
    }
}
typefn Primary<C> {
    match string.match<Skip<C>, "{"> {
        case ["hit", infer Open, infer At] { return RecordFields<At>; }
        else { return Scalar<C>; }
    }
}
typefn Annotation<C> {
    type [Value, At] = Primary<C>;
    while (true) {
        match string.match<Skip<At>, "[]"> {
            case ["hit", infer Suffix, infer Next] { Value = Value[]; At = Next; }
            else { return [Value, At]; }
        }
    }
}
typefn Parent<S> {
    type Start = { source: S; offset: 0 };
    type [Name, AfterName] = Word<Expect<Start, "root(parent ">>;
    type [Value, AfterValue] = Annotation<Expect<AfterName, ":">>;
    type End = Expect<AfterValue, ")">;
    match string.read<Skip<End>> {
        case ["eof"] { return object.addField<{}, Name, { type: Value }>; }
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
