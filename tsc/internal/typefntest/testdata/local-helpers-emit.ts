export typefn Outer<T> {
 typefn Pair<U> {return [T,U];}
 return Pair<2>;
}
export const result:Outer<1> = [1,2];
