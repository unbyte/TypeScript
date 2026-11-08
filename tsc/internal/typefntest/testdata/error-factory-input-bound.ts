export {};
typefn Factory<T extends string> {return typefn<U> {return [T,U];};}
typefn Outer<T> {Factory<T>; return 123;}
type Pending<T>=Outer<T>;
type Invalid=Pending<1>;
