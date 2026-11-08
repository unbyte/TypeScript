type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
typefn Inspect<S> { return [text.length<S>, text.peek<S, 0>, text.peek<S, 1>, text.peek<S, 2>, text.codeUnit<S, 0>, text.slice<S, 1>, text.slice<S, 0, 0>]; }
type Generic<S> = Inspect<S>;
type Check = Assert<Equal<Generic<"\uD83D\uDE00">, [2, "\uD83D", "\uDE00", "", 55357, "\uDE00", ""]>>;
typefn Lone<> { return [text.fromCodeUnit<55296>, text.length<"\uD800">, text.codeUnit<"\uD800", 0>]; }
type LoneCheck = Assert<Equal<Lone<>, ["\uD800", 1, 55296]>>;
typefn Digits<> { return [text.digit<"0">, text.digit<"9">, text.digit<"">, text.digit<"12">, text.digit<"９">]; }
type DigitCheck = Assert<Equal<Digits<>, [0n, 9n, -1n, -1n, -1n]>>;
typefn Units<S> { type result = []; for (type unit of S) { result = [...result, unit]; } return result; }
type UnitCheck = Assert<Equal<Units<"A😀\uD800">, ["A", "\uD83D", "\uDE00", "\uD800"]>>;
type EmptyCheck = Assert<Equal<Units<"">, []>>;
