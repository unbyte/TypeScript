typefn Bad<T> { return types.function<T, string>; }
type Result = Bad<string[]>;
