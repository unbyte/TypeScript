typefn F<T> { return reflect.array<T>; }
type Result = F<readonly [1,2]>;
