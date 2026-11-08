typefn Join<T> { return text.concat<["a", T]>; }
type Generic<T> = Join<T>;
type Result = Generic<number>;
