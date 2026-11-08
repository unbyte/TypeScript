type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
typefn Identity<out> { return out; }
typefn Calculate<N> {
  type out = N;
  out = Identity<out + 1>;
  type async = out;
  async = async * 2;
  return async;
}
type Deferred<T> = Calculate<T>;
type Answer = Assert<Equal<Deferred<20>, 42>>;
