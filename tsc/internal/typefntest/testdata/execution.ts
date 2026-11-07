type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
export typefn Double<N> { return N * 2; }
typefn Reverse<T> {
 type result = [];
 for (type item of T) { result = [item, ...result]; }
 return result;
}
typefn Range<N> {
 type result = [];
 type i = 0;
 while (i < N) {
  result = [...result, i];
  i = i + 1;
 }
 return result;
}
typefn Filter<T> {
 type result = never;
 for (type member in T) { if (member is string) { result = result | member; } }
 return result;
}
typefn Make<K, T> { return { readonly [K]?: T; value: [T, T] }; }
typefn Snapshot<T> { type x = T; type y = [x]; x = number; return y; }
typefn Loop<> {
 type result = [];
 type n = 0;
 while (true) {
  n = n + 1;
  if (n == 2) { continue; }
  if (n >= 4) { break; }
  result = [...result, n];
 }
 return result;
}
type Box<T extends string = "default"> = { value: T };
typefn Native<T> { return [Box<T>, Uppercase<T>, Promise<T>]; }
typefn Nested<T> { return Reverse<Range<Double<T>>>; }
typefn Exact<> { return [-7n / 3n, -7n % 3n, 12345678901234567890n * 2n]; }
typefn Safe<> { return false && (1 / 0 == 0); }
typefn Indexed<T> { return T["value"]; }
typefn Scoped<T> { type x = T; { type x = number; x = boolean; } return x; }
type Check = [
 Assert<Equal<Double<21>, 42>>,
 Assert<Equal<Reverse<[1, "two", true]>, [true, "two", 1]>>,
 Assert<Equal<Range<4>, [0, 1, 2, 3]>>,
 Assert<Equal<Filter<"get" | "post" | 42>, "get" | "post">>,
 Assert<Equal<Make<"key", string>, {readonly key?: string; value: [string, string]}>>,
 Assert<Equal<Snapshot<string>, [string]>>,
 Assert<Equal<Loop<>, [1, 3]>>,
 Assert<Equal<Native<"hello">, [{value: "hello"}, "HELLO", Promise<"hello">]>>,
 Assert<Equal<Nested<2>, [3, 2, 1, 0]>>,
 Assert<Equal<Exact<>, [-2n, -1n, 24691357802469135780n]>>,
 Assert<Equal<Safe<>, false>>,
 Assert<Equal<Indexed<{value: 42}>, 42>>,
 Assert<Equal<Scoped<"outer">, "outer">>
];
export const answer: Double<21> = 42;
typefn Factorial<N> {
 if (N <= 1) { return 1; }
 return N * Factorial<N - 1>;
}
type Recursive = Assert<Equal<Factorial<6>, 720>>;
