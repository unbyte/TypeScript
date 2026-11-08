export {};
type Assert<T extends true> = T;
type Equal<A,B> = (<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
// Result-kind checking never executes these intentionally cyclic functions.
typefn First<T>: type {return Second<T>;}
typefn Second<T> {return Third<T>;}
typefn Third<T> {return First<T>;}
typefn Inferred<T> {return Later<T>;}
typefn Later<T> {return [T];}
type Value=Assert<Equal<Inferred<123>,[123]>>;
typefn Branch<T> {if(T == 0) {return 0;} return Branch<T-1>;}
type Recursive=Assert<Equal<Branch<5>,0>>;
typefn NeverCalled<> {
 type F=typefn<T> {fail<"unreached closure body">;};
 return 123;
}
type Uncalled=Assert<Equal<NeverCalled<>,123>>;
typefn Abort<>: type {fail<"unreached helper body">;}
