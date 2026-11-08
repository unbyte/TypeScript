export {};
typefn Bounded<T extends string> {return [T];}
typefn Apply<F<_>,T> {return F<T>;}
type Invalid=Apply<Bounded,123>;
