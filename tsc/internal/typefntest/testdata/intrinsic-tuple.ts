type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
typefn Slice<T, N> { return tuple.slice<T, N>; }
typefn Length<T> { return tuple.length<T>; }
type Later<T> = Slice<readonly [first: string, value: T], 1>;
type Checks = [Assert<Equal<Later<42>, readonly [value: 42]>>, Assert<Equal<Length<[unknown, never]>,2>>, Assert<Equal<Slice<[],0>,[]>>];
typefn Known<T> { return tuple.length<[T, T]>; }
type Stable<T> = Known<T>;
type StableCheck<T> = Assert<Equal<Stable<T>, 2>>;
