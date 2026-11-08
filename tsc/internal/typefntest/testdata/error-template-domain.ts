typefn Interpolate<T> { return `x${T}`; }
type Bad = Interpolate<{value:1}>;
