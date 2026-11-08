type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
typefn Twice<N> { return N * 2; }
typefn Discard<N> { type unused = Twice<N>; return "ok"; }
typefn Envelope<N> { return {tag: "number"; value: Twice<N>}; }
type Erase<X> = "erased";
type Discarded<T> = Discard<T>;
type Erased<T> = Erase<Twice<T>>;
type Absorbed<T> = any | Twice<T>;
type Projected<T> = Envelope<T>["tag"];
type Keys<T> = keyof Envelope<T>;
type Selected<T> = Discard<T> extends "ok" ? 1 : 0;
type Empty<T> = never & Twice<T>;
type Results = [
 Assert<Equal<Discarded<21>, "ok">>,
 Assert<Equal<Erased<21>, "erased">>,
 Assert<Equal<Absorbed<21>, any>>,
 Assert<Equal<Projected<21>, "number">>,
 Assert<Equal<Keys<21>, "tag" | "value">>,
 Assert<Equal<Selected<21>, 1>>,
 Assert<Equal<Empty<21>, never>>
];
typefn Upper<T> { return Uppercase<T>; }
type Capital<T> = Upper<T>;
type UpperResult = Assert<Equal<Capital<"hello">, "HELLO">>;
typefn NativeDiscard<T> { type unused = Uppercase<T>; return 1; }
type NativeChecked<T> = NativeDiscard<T>;
type NativeResult = Assert<Equal<NativeChecked<"hello">, 1>>;
