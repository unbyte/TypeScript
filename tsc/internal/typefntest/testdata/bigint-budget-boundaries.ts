type Assert<T extends true> = T;
type Equal<A,B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
typefn Compute<N> { return [numeric.toNumber<N>, numeric.format<N>, numeric.abs<N>, -N, N + 0n, N < 0n]; }
type Positive = Assert<Equal<Compute<255n>, [255, "255", 255n, -255n, 255n, false]>>;
type Negative = Assert<Equal<Compute<-255n>, [-255, "-255", 255n, 255n, -255n, true]>>;
type Zero = Assert<Equal<Compute<0n>, [0, "0", 0n, 0n, 0n, false]>>;
typefn Parse<Text> { return numeric.parseInteger<Text>; }
type Parsed = Assert<Equal<Parse<"+000255">, 255n>>;
type ParsedNegative = Assert<Equal<Parse<"-000255">, -255n>>;
type NegativeZero = Assert<Equal<Parse<"-000">, 0n>>;
