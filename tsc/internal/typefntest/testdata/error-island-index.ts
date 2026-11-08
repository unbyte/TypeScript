typefn Bad<T> { return type(T["missing"]); }
type Result = Bad<{value:1}>;
