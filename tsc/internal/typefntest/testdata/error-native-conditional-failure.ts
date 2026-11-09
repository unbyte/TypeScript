typefn Guard<N> {
  if (N == 0) { fail<"Conditional replay failure">; }
  return N;
}
type Conditional<T> = T extends number ? Guard<T> : "other";
typefn ThroughConditional<T> { return type(Conditional<T>); }
type First = Conditional<0 /* first */>;
type Second = Conditional<0 /* second */>;
type WrappedFirst = ThroughConditional<0 /* wrapped first */>;
type WrappedSecond = ThroughConditional<0 /* wrapped second */>;
type Assert<T extends true> = T;
type SuccessfulSibling = Assert<ThroughConditional<1> extends 1 ? true : false>;
type Unselected = Assert<ThroughConditional<"x"> extends "other" ? true : false>;
