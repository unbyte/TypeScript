export {};
typefn Apply<F<_>,T> {return F<T>;}
type Bad=Apply<string,1>;
