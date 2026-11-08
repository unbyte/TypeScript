export {};
typefn Box<T,U> {return [T,U];}
typefn Apply<F<_>,T> {return F<T>;}
type Result=Apply<Box,1>;
