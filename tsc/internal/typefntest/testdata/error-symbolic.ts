typefn F<T> { return T + 1; } type R<U> = F<U>;
type Invalid = R<string>;
