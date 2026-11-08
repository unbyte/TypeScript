type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
typefn Classify<T> { match (T) { 1 => { return "one"; } -2n => { return "negative"; } is string => { return "string"; } _ => { return "other"; } } }
typefn Head<T> { match (T) { [head, ...rest] => { return [head, rest]; } _ => { return never; } } }
typefn Pair<T> { match (T) { [a, b] => { return [b, a]; } _ => { return false; } } }
typefn Fields<T> { match (T) { {value: [head, ...rest], extra: item} => { return [head, rest, item]; } _ => { return false; } } }
typefn Guard<T> { match (T) { [item] if (item == 1) => { return "first"; } [item] => { return item; } _ => { return never; } } }
typefn Mutate<T> { type result = 0; match (T) { 1 => { result = 2; } _ => { result = 3; } } return result; }
typefn Loop<T> { type result = []; for (type item of T) { match (item) { 0 => { continue; } 1 => { break; } capture => { result = [...result, capture]; } _ => {} } } return result; }
typefn Whole<T> { match (T) { item => { return item; } _ => { return never; } } }
type Deferred<T> = Classify<T>;
type DeferredGuard<T> = Guard<[T]>;
type Symbolic<T> = Fields<{value: [T, 2]; extra: T}>;
class Hidden { private value!: [1]; extra!: 2; }
typefn Scoped<T> { type label = 1; match (T) { item if (label == 1) => { type label = 2; type item = 3; return [label, item]; } _ => { return never; } } }
type Checks = [Assert<Equal<Scoped<7>, [2, 3]>>,
 Assert<Equal<Classify<1>, "one">>, Assert<Equal<Classify<-2n>, "negative">>, Assert<Equal<Classify<"a" | "b">, "string">>,
 Assert<Equal<Deferred<2>, "other">>, Assert<Equal<Deferred<1>, "one">>,
 Assert<Equal<Head<readonly [a: 1, b: 2, c: 3]>, [1, readonly [b: 2, c: 3]]>>,
 Assert<Equal<Head<[1?]>, never>>, Assert<Equal<Head<string[]>, never>>, Assert<Equal<Head<1>, never>>,
 Assert<Equal<Pair<[1, 2]>, [2, 1]>>, Assert<Equal<Pair<[1, 2, 3]>, false>>,
 Assert<Equal<Fields<{value: [1, 2]; extra: 3; ignored: 4}>, [1, [2], 3]>>,
 Assert<Equal<Fields<{value?: [1]; extra: 2}>, false>>, Assert<Equal<Fields<{[key: string]: [1]}>, false>>, Assert<Equal<Fields<Hidden>, false>>,
 Assert<Equal<Symbolic<"x">, ["x", [2], "x"]>>, Assert<Equal<Guard<[2]>, 2>>, Assert<Equal<DeferredGuard<1>, "first">>, Assert<Equal<DeferredGuard<3>, 3>>,
 Assert<Equal<Mutate<1>, 2>>, Assert<Equal<Mutate<0>, 3>>, Assert<Equal<Loop<[2, 0, 3, 1, 4]>, [2, 3]>>,
 Assert<Equal<Whole<never>, never>>, Assert<Equal<Whole<{x: 1}|{y: 2}>, {x: 1}|{y: 2}>>
];
