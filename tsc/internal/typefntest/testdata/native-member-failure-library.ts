typefn Guard<N> {
  if (N == 0) { fail<"Member replay failure">; }
  return N;
}
export type Mapped<T> = { [K in keyof T]: Guard<T[K]> };
export type Input = { value: 0 };
export typefn Read<T> { return type(Mapped<T>["value" & keyof T]); }
