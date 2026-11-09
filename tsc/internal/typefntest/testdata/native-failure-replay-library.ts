typefn Guard<N> {
  if (N == 0) { fail<"Native replay failure">; }
  return N;
}
export type Alias<T> = Guard<T>;
export typefn ThroughAlias<T> { return type(Alias<T>); }
export type Nested<T> = ThroughAlias<T>;
