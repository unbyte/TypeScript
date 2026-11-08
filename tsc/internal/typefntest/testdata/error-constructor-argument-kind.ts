export {};
typefn Apply<F<_>,T> {return F<T>;}
typefn Unused<> {return Apply<1,2>;}
