typefn Require<T> { if (T is string) { return 1; } fail<"Expected string">; }
type Native<T> = {value: Require<T>};
typefn Same<A,B> { return A == B; }
type Later<T> = Same<Native<T>,{value:1}>;
type Failure = Later<0>;
