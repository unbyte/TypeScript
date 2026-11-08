type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
typefn Search<S> { return [text.indexOf<S, "aaab">, text.indexOf<S, "", 2>, text.includes<S, "aaab">, text.startsWith<S, "aa">, text.endsWith<S, "b">]; }
type Generic<S> = Search<S>;
type A = Assert<Equal<Generic<"aaaaab">, [2, 2, true, true, true]>>;
typefn Split<S> { return text.split<S, "::">; }
type B = Assert<Equal<Split<"a::b::">, ["a", "b", ""]>>;
type C = Assert<Equal<Split<"">, [""]>>;
typefn Replace<S> { return text.replaceAll<S, "aa", "$&">; }
type D = Assert<Equal<Replace<"aaaaa">, "$&$&a">>;
typefn Join<T> { return text.concat<["a", T, "b"]>; }
type Deferred<T> = Join<T>;
type E = Assert<Equal<Deferred<"x">, "axb">>;
typefn Unicode<> { return [text.indexOf<"😀x", "x">, text.split<"😀", "\uDE00">, text.concat<["\uD83D", "\uDE00"]>, text.replaceAll<"a\uDE00", "a", "\uD83D">]; }
type F = Assert<Equal<Unicode<>, [2, ["\uD83D", ""], "😀", "😀"]>>;
typefn Edges<> { return [text.indexOf<"abc", "z">, text.indexOf<"abc", "a", 1>, text.includes<"", "">, text.startsWith<"", "">, text.endsWith<"x", "xx">, text.concat<[]>]; }
type G = Assert<Equal<Edges<>, [-1, -1, true, true, false, ""]>>;
