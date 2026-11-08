type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
typefn Header<S> { match (S) { `${method} ${route}` => { return [method, route]; } _ => { return false; } } }
typefn Tail<S> { match (S) { `GET ${tail}` => { return tail; } _ => { return false; } } }
typefn Whole<S> { match (S) { `${value}` => { return value; } _ => { return false; } } }
typefn Units<S> { match (S) { `${a}${b}${c}` => { return [a,b,c]; } _ => { return false; } } }
typefn End<S> { match (S) { `${a}::${b}!` => { return [a,b]; } _ => { return false; } } }
typefn Pinned<S,D> { match (S) { `${a}${=D}${b}` => { return [a,b]; } _ => { return false; } } }
typefn Delimiter<S> { match (S) { `${a}${=":"}x${b}` => { return [a,b]; } _ => { return false; } } }
typefn Shadow<S> { type x = "::"; match (S) { `${x}${=x}${tail}` => { return [x,tail]; } _ => { return false; } } }
typefn Wild<S> { match (S) { `${_}/${tail}` => { return tail; } _ => { return false; } } }
typefn Empty<S> { match (S) { `` => { return true; } _ => { return false; } } }
typefn Escaped<S> { match (S) { `a\n${tail}` => { return tail; } _ => { return false; } } }
type Deferred<T> = Header<T>;
type Symbolic<T extends string> = Tail<`GET /users/${T}`>;
type DeferredPin<T> = Pinned<"left::right",T>;
type DeferredEmpty<T extends string> = Empty<`${T}`>;
type DeferredOpaque<T extends string> = Wild<`x${T}/tail`>;
typefn PairUnits<S> { match (S) { `${a}${b}` => { return [a,b]; } _ => { return false; } } }
type Checks = [Assert<Equal<DeferredOpaque<"a">, "tail">>, Assert<Equal<PairUnits<`x${string}`>,["x",string]>>,
 Assert<Equal<Header<"GET /users/:id">,["GET","/users/:id"]>>, Assert<Equal<Deferred<"POST /">,["POST","/"]>>,
 Assert<Equal<Tail<`GET /users/${string}`>,`/users/${string}`>>, Assert<Equal<Symbolic<"42">,"/users/42">>,
 Assert<Equal<Whole<string>,string>>, Assert<Equal<Whole<"">,"">>, Assert<Equal<Whole<never>,false>>, Assert<Equal<Whole<1>,false>>,
 Assert<Equal<Units<"abcde">,["a","b","cde"]>>, Assert<Equal<Units<"ab">,["a","b",""]>>, Assert<Equal<Units<"a">,false>>,
 Assert<Equal<Units<"😀x">,["\uD83D","\uDE00","x"]>>, Assert<Equal<Units<`${number}`>,false>>,
 Assert<Equal<End<"a::b!">,["a","b"]>>, Assert<Equal<End<"a::b!c!">,false>>,
 Assert<Equal<Pinned<"a::b::c","::">,["a","b::c"]>>, Assert<Equal<Delimiter<"a:bad:xend">,["a:bad","end"]>>,
 Assert<Equal<Shadow<"left::right">,["left","right"]>>, Assert<Equal<DeferredPin<"::">,["left","right"]>>,
 Assert<Equal<Wild<"/tail">,"tail">>, Assert<Equal<Wild<`x${string}/tail`>,false>>,
 Assert<Equal<Empty<"">,true>>, Assert<Equal<Empty<string>,false>>, Assert<Equal<DeferredEmpty<"">,true>>, Assert<Equal<DeferredEmpty<"x">,false>>,
 Assert<Equal<Escaped<"a\nrest">,"rest">>
];
