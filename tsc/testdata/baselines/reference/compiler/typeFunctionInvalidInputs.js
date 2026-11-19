//// [tests/cases/compiler/typeFunctionInvalidInputs.ts] ////

//// [typeFunctionInvalidInputs.ts]
typefn Domain<T> { return tuple.count<T>; }
typefn Slice<S> { return string.slice<S, 1, 2>; }
typefn Div<N> { return number.divide<1, N>; }
typefn BadIndex<T> { return object.withIndexes<T, {key: string; type: number}>; }
typefn Recursive<T> { return Recursive<T>; }
typefn Malformed<T> { return number.fromDecimal<T>; }
typefn Duplicate<T> { return object.addField<T, "name", {type: string}>; }
declare const array: Domain<string[]>;
declare const splitPair: Slice<"😀">;
declare const zero: Div<0>;
declare const index: BadIndex<{name: string}>;
declare const cycle: Recursive<1>;
declare const malformed: Malformed<"1e">;
declare const duplicate: Duplicate<{name: string}>;
array.a; splitPair.a; zero.a; index.a; cycle.a; malformed.a; duplicate.a;

typefn Control<T> {
    break;
    T = number;
    fail<"too", "many", "args">;
}
type Outside = tuple.count<[1]>;

typefn Escape<> { return {ref: fn.ref<"missing">}; }
declare const escaped: Escape<>;
escaped.anyProperty;

typefn WrongSlots<F<X>> { return 1; }
typefn NestedConstructor<F<_>> { return {nested: [F]}; }
typefn Parameter<T extends string> { return T; }
type BadConstraint = Parameter<5>;
typefn BadUse<T> {
    type A = Parameter<T, T>;
    42;
    match T { else { return 0; } case string { return 1; } }
}
typefn Missing<T> { if (false) { return T; } }
declare const end: Missing<1>;
end.recovered;

typefn InvalidDefault<> {
    return fn.compose<{ members: {}; constructs: []; calls: [{
        typeParameters: [{name: "T"; key: "t"; constraint: string; default: 1}];
        parameters: []; result: {kind: "return"; type: 0};
    }] }>;
}
declare const invalidDefault: InvalidDefault<>;
invalidDefault.recovered;

typefn MatchOne<T> {
    match T { case [infer X] { return X; } }
    return "fell through";
}
typefn DuplicateCapture<T> { match T { case [infer X, infer X] { return X; } } }
declare const noArm: MatchOne<0>;
declare const ambiguous: MatchOne<any>;
declare const empty: MatchOne<never>;
noArm.recovered; ambiguous.recovered; empty.recovered;


//// [typeFunctionInvalidInputs.js]
"use strict";
array.a;
splitPair.a;
zero.a;
index.a;
cycle.a;
malformed.a;
duplicate.a;
escaped.anyProperty;
end.recovered;
invalidDefault.recovered;
noArm.recovered;
ambiguous.recovered;
empty.recovered;
