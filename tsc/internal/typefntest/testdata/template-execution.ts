type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
typefn Interpolate<T> { return `head${T}tail`; }
type Later<T> = Interpolate<T>;
type Checks = [
 Assert<Equal<Later<42>, "head42tail">>,
 Assert<Equal<Interpolate<"a" | "b">, "headatail" | "headbtail">>,
 Assert<Equal<Interpolate<string>, `head${string}tail`>>,
 Assert<Equal<Interpolate<never>, never>>,
 Assert<Equal<Interpolate<null>, "headnulltail">>
];
typefn Nested<N> { return `x${N+1}:${`y${N}`}\`${true}:${-1n}`; }
type NestLater<N> = Nested<N>;
type NestedCheck = Assert<Equal<NestLater<2>, "x3:y2`true:-1">>;
typefn Snapshot<T> { type value = T; type saved = `${value}`; value = 0; return saved; }
type Snap<T> = Snapshot<T>;
type SnapshotCheck = Assert<Equal<Snap<"old">, "old">>;
typefn Surrogates<> { return `\uD83D${"\uDE00"}`; }
type UnicodeCheck = Assert<Equal<Surrogates<>, "😀">>;
