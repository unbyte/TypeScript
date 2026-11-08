typefn Bad<S> { return numeric.parseInteger<S>; }
type Result = Bad<"1_2">;
