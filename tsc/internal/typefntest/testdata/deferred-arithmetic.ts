type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
typefn Twice<N> { return N * 2; }
type Deferred<N> = Twice<N>;
type Answer = Assert<Equal<Deferred<21>, 42>>;
type Again<T> = Deferred<T>;
type Composed = Assert<Equal<Again<3>, 6>>;
typefn Envelope<N> { return {tag: "number"; value: N * 2}; }
type Wrapped<N> = Envelope<N>;
type Value = Assert<Equal<Wrapped<21>["value"], 42>>;
typefn Box<T> { return {value: T}; }
type Boxed<T> = Box<T>;
type Field = Assert<Equal<Boxed<string>["value"], string>>;
typefn Negative<N> { return -N; }
type Signed<N> = Negative<N>;
type Big = Assert<Equal<Signed<12n>, -12n>>;
typefn Choose<T> { if (T < 3) { return "small"; } return "large"; }
type Choice<T> = Choose<T>;
type Small = Assert<Equal<Choice<1>, "small">>;
type Large = Assert<Equal<Choice<4>, "large">>;
