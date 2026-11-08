export {};
typefn Apply<F<_>,T> {return F<T>;}
typefn Unused<> {return type(Apply<string,1>);}
