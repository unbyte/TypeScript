export {};
typefn Apply<F<_>,T> {return F<T>;}
type Bound<T extends string>=[T];
type Invalid=Apply<Bound,123>;
