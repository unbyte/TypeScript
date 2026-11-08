typefn Len<T> { return tuple.length<T>; }
type Result = Len<[1, ...string[]]>;
