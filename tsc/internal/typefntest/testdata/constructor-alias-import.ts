import {AsyncArray as Imported} from "./library.js";
type Assert<T extends true> = T;
type Equal<A,B> = (<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
type Result=Assert<Equal<Imported<string>,Promise<string[]>>>;
typefn Apply<F<_>,T> {return F<T>;}
type Passed=Assert<Equal<Apply<Imported,1>,Promise<1[]>>>;
typefn DSL<T> {return Imported<T>;}
type FromDSL=Assert<Equal<DSL<2>,Promise<2[]>>>;
