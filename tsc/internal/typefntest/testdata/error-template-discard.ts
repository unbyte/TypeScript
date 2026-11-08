typefn Interpolate<T> { return `x${T}`; }
type Discard<T> = any | Interpolate<T>;
type Bad = Discard<symbol>;
