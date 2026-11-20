//// [tests/cases/compiler/typeFunctionWholeUnion.ts] ////

//// [typeFunctionWholeUnion.ts]
type Expect<T extends true> = T;
type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Native<T> = T extends string ? [T] : false;
type NativeMap<T> = { [K in keyof T]: [T[K]] };

typefn Classify<T> { return T extends string ? [T] : false; }
typefn Local<T> { type P = T; return { tag: P extends string ? "all" : "other" }; }
typefn Default<T, R = T extends string ? 1 : 0> { return R; }
typefn Capture<T> { return T extends { value: infer V } ? [V] : false; }
typefn Wrap<T> { return T extends unknown ? [T] : never; }
typefn Relay<T> { return Classify<T>; }
typefn CallNative<T> { return Native<T>; }
typefn MapFields<T> { return { [K in keyof T]: [T[K]] }; }
typefn CallNativeMap<T> { return NativeMap<T>; }
typefn Each<T> {
    type Result = never;
    for (type Part of union.toTuple<T>) { Result = Result | Classify<Part>; }
    return Result;
}

type Words = "a" | "b";
type Mixed = "a" | 1;
type Objects = { id: 1; left: true } | { id: 2; right: true };
type Whole = Expect<Equal<Classify<Words>, [Words]>>;
type PartlyMatches = Expect<Equal<Classify<Mixed>, false>>;
type Empty = Expect<Equal<Classify<never>, [never]>>;
type Dynamic = Expect<Equal<Classify<any>, [any]>>;
type Nested = Expect<Equal<Local<Mixed>, { tag: "other" }>>;
type Defaulted = Expect<Equal<Default<Mixed>, 0>>;
type Inferred = Expect<Equal<Capture<{ value: 1 } | { value: 2 }>, [1 | 2]>>;
type Unconditional = Expect<Equal<Wrap<Mixed>, [Mixed]>>;
type NestedCall = Expect<Equal<Relay<Mixed>, false>>;
type Ordinary = Expect<Equal<CallNative<Mixed>, ["a"] | false>>;
type Explicit = Expect<Equal<Each<Words>, ["a"] | ["b"]>>;
type KeysTogether = Expect<Equal<MapFields<Objects>, { id: [1 | 2] }>>;
type KeysApart = Expect<Equal<CallNativeMap<Objects>, NativeMap<Objects>>>;
type PreservesModifiers = Expect<Equal<MapFields<{ readonly id?: 1 } | { readonly id?: 2 }>, { readonly id?: [1 | 2 | undefined] }>>;
type SingleTuple = Expect<Equal<MapFields<readonly [1, 2]>, readonly [[1], [2]]>>;

// A concrete union arrives through a deferred call just as it does through a direct call.
type Later<T> = Relay<T>;
type Instantiated = Expect<Equal<Later<Mixed>, false>>;

typefn Abort<> { fail<"unselected arm">; }
typefn Select<T> { return T extends string ? Abort<> : 0; }
type OneBranch = Expect<Equal<Select<Mixed>, 0>>;

// The returned signature binds U when its ordinary TypeScript caller applies it.
typefn Signature<T> {
    return <U>(value: U) => [T extends string ? T : false, U extends string ? [U] : false];
}
declare const signature: Signature<Mixed>;
declare const input: Mixed;
export const signatureResult = signature(input);
type Applied = Expect<Equal<typeof signatureResult, [false, ["a"] | false]>>;


//// [typeFunctionWholeUnion.js]
export const signatureResult = signature(input);


//// [typeFunctionWholeUnion.d.ts]
export declare const signatureResult: [false, false | ["a"]];
