type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
typefn NumericRange<N> { type result = []; type i = 0; while (i < N) { result = [...result, i]; i = i + 1; } return result; }
type R<T> = NumericRange<T>;
type RangeResult = Assert<Equal<R<4>, [0,1,2,3]>>;
typefn Reverse<T> { type result = []; for (type item of T) { result = [item, ...result]; } return result; }
type Rev<T> = Reverse<T>;
type Reversed = Assert<Equal<Rev<[1,"a",true]>, [true,"a",1]>>;
typefn Wrap<T> { return {value: [0,...T]}; }
type W<T> = Wrap<T>;
type Spread = Assert<Equal<W<[1,2]>, {value:[0,1,2]}>>;
typefn Dynamic<K,V> { return {[K]:V}; }
type D<K,V> = Dynamic<K,V>;
type DynamicResult = Assert<Equal<D<"answer",42>, {answer:42}>>;
typefn Capture<T> { type n=T; type saved=[...n]; n=[99]; return saved; }
type C<T> = Capture<T>;
type Captured = Assert<Equal<C<[1,2]>,[1,2]>>;
typefn Bounded<T extends string> { return T; }
type B<T> = Bounded<T>;
type BoundResult = Assert<Equal<B<"ok">,"ok">>;
typefn Short<T> { return false && (T / 0 > 0); }
type S<T> = Short<T>;
type ShortResult = Assert<Equal<S<"not a number">,false>>;
typefn Safe<T> { if (T == 0) { return 0; } return 10 / T; }
type SafeAlias<T> = Safe<T>;
type SafeZero = Assert<Equal<SafeAlias<0>,0>>;
type SafeTwo = Assert<Equal<SafeAlias<2>,5>>;
typefn Combine<T> { return (T < 3) && (T > 0); }
type Both<T> = Combine<T>;
type BothResult = Assert<Equal<Both<2>,true>>;
