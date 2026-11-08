typefn Outer<T> {typefn Inner<U extends string> {return U;} Inner<T>; return unknown;} type Pending<T> = Outer<T>; type Invalid = Pending<123>;
