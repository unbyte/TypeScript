export {};
type Pair<T,U=string>=[T,U];
typefn Apply<F<_>,T> {return F<T>;}
type Invalid=Apply<Pair,1>;
