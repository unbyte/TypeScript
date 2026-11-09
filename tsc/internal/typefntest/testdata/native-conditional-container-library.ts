typefn Guard<N> {
  if (N == 0) { fail<"Native cached-container failure">; }
  return N;
}
export type Tuple<N> = [Guard<N>];
export type Conditional<N> = N extends number ? [Guard<N>] : [N];
export type Absorbed<N> = N extends number ? Guard<N> | unknown : N;
