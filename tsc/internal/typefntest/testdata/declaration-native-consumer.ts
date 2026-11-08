import type {Item,Pair} from './out/library';
type Assert<T extends true>=T;
type Equal<A,B>=(<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
type Value=Assert<Equal<Pair<2>[0],2>>;
type Instance=Assert<Equal<Pair<2>[1]['value'],1>>;
type Nominal=Assert<{value:1} extends Item?false:true>;
