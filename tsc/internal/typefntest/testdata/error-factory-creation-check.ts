export {};
typefn Box<T> {return [T];}
typefn Factory<T> {type check=T+1; return Box;}
typefn Outer<T> {Factory<T>; return 123;}
type Pending<T>=Outer<T>;
type Bad=Pending<"wrong">;
