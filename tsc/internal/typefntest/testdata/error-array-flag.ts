typefn F<T> { return types.array<string,T>; }
type Later<T> = F<T>;
type Result = Later<boolean>;
