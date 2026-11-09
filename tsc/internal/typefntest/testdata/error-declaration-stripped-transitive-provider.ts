/** @internal */
typefn Hidden<T> { return T; }
export typefn Bridge<T> { return Hidden<T>; }
