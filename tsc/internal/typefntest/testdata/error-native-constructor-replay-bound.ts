export {};
typefn Apply2<F<_,_>,T,U> {return F<T,U>;}
type Pending<T>=Apply2<Pick,{a:1},T>;
type Invalid=Pending<"b">;
