export {};
typefn Box<T> {return [T];}
typefn Choose<T> {if(T==1) {return Box;} return Box;}
typefn Apply<F<_>,T> {return F<T>;}
type Pending<T>=Apply<Choose<T>,1>;
