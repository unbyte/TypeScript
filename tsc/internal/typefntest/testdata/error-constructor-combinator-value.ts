export {};
typefn Apply<F<_>,T> {return F<T>;}
typefn Unused<> {type F=Apply; return 1;}
