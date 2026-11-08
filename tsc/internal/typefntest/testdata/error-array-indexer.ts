typefn F<T> { return reflect.array<T>; }
type Result = F<{[index:number]:string;length:number}>;
