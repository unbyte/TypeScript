typefn Guard<N> {
  if (N == 0) { fail<"Native cached-container failure">; }
  return N;
}
export type Functions<N> = (() => Guard<N>) | (() => N);
export type Constructors<N> = (new () => [Guard<N>]) | (new () => [N]);
export type Overloaded<N> = ((tag: "fail") => Guard<N>) & ((tag: "ok") => 1);
