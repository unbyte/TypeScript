type Generic = <T>(value:T) => T;
typefn Reflect<T> { return reflect.signature<T>; }
type Result = Reflect<Generic>;
