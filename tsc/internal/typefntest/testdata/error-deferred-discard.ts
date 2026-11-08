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
type Invalid = Discarded<"bad">;
