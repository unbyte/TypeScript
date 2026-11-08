export {};
typefn Apply2<F<_,_>,T,U> {return F<T,U>;}
type Invalid=Apply2<Pick,{a:1},"b">;
