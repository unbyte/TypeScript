import type {Box as Imported} from "./library.js";
type Assert<T extends true> = T;
type Equal<A,B> = (<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Binding<T> {type F=Imported; return F<T>;}
type Pending<T>=Binding<T>;
type Result=Assert<Equal<Pending<123>,[123]>>;
