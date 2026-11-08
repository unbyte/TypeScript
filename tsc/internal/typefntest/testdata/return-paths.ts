export {};
type Assert<T extends true> = T;
type Equal<A,B> = (<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Choose<T> { if (T) { return 1; } else { return 2; } }
typefn Abort<>: type { fail<"stop">; }
typefn Guard<T> { if (T) { return 1; } else { failAt<"stop",0>; } }
typefn Match<T> { match(T) { 0 => { return 1; } value if (value == 1) => { return 2; } _ => { return 3; } } }
typefn Capture<T> { match(T) { value => { return value; } _ => {} } }
typefn Nested<T> { if (T) { { return 1; } } else { match(T) { _ => { return 2; } } } }
typefn Loop<T> { for(type item of T) { if (item == 0) {continue;} if(item == 1) {break;} return item; } return -1; }
typefn Dead<> {return 1; type unused = 2;}
type C = Assert<Equal<Choose<false>,2>>;
type G = Assert<Equal<Guard<true>,1>>;
type M = Assert<Equal<Match<1>,2>>;
type W = Assert<Equal<Capture<123>,123>>;
type N = Assert<Equal<Nested<false>,2>>;
type L = Assert<Equal<Loop<[0,1,2]>,-1>>;
type D = Assert<Equal<Dead<>,1>>;
