typefn F<T> { return reflect.kind<T>; }
type Result = F<`/users/${string}`>;
