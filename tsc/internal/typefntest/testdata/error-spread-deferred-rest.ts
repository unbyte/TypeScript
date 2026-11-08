typefn Join<A,B> {return [...A,...B];}
type Pending<T> = Join<T,string[]> | unknown;
type Result = Pending<number[]>;
