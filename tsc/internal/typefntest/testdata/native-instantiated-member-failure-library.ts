typefn Guard<N> {
  if (N == 0) { fail<"Instantiated member failure">; }
  return N;
}
export interface Box<N> { value: Guard<N>; }
export typefn Read<N> { return type(Box<N>["value"]); }
export declare class WriteBox<N> {
  get value(): number;
  set value(input: Guard<N>);
}
