/** @internal */
typefn Hidden<T> { return T; }
export typefn Public<T> { return Hidden<T>; }
