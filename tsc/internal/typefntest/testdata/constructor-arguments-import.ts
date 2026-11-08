import {Box as Imported} from "./library.js";
type Assert<T extends true> = T;
type Equal<A,B> = (<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Apply<F<_>,T> {return F<T>;}
type Result=Assert<Equal<Apply<Imported,1>,[1]>>;
typefn DSL<> {return Apply<Imported,2>;}
type InBody=Assert<Equal<DSL<>,[2]>>;
