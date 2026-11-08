export {};
typefn Unused<> {
 typefn Invoke<F<>> {return F<>;}
 typefn Apply<F<_>,T> {return F<T>;}
 return type(Apply<Invoke,1>);
}
