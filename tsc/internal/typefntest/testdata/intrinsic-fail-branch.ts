type Assert<T extends true> = T;
typefn Check<T> { if (T == 1) { return true; } fail<"Expected one">; }
type Deferred<T> = Check<T>;
type Result = Assert<Deferred<1>>;
