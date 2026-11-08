type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
typefn Numbers<N> { return [numeric.floor<N>, numeric.ceil<N>, numeric.trunc<N>, numeric.abs<N>, numeric.isInteger<N>, numeric.isSafeInteger<N>]; }
type Later<N> = Numbers<N>;
type A = Assert<Equal<Later<-1.5>, [-2, -1, -1, 1.5, false, false]>>;
typefn Exact<> { return [numeric.toNumber<9007199254740992n>, numeric.toBigint<9007199254740993>, numeric.parseInteger<"+000123">, numeric.parseInteger<"-000">, numeric.abs<-123n>, numeric.format<-123n>, numeric.format<1e21>, numeric.isSafeInteger<9007199254740992>]; }
type B = Assert<Equal<Exact<>, [9007199254740992, 9007199254740992n, 123n, 0n, 123n, "-123", "1e+21", false]>>;
typefn Parse<S> { return numeric.parseInteger<S>; }
type Generic<S> = Parse<S>;
type C = Assert<Equal<Generic<"9007199254740993">, 9007199254740993n>>;
