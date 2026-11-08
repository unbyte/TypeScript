export {};
typefn Factory<T> {return typefn<U> {return [T,U];};}
type Invalid<T>=Factory<T>;
