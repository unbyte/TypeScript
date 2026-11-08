enum E {A,B}
typefn F<T> { return reflect.kind<T>; }
type Result = F<E>;
