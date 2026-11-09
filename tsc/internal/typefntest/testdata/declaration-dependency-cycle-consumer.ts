import type {Pair} from './out/library';
type Assert<T extends true>=T;
type Equal<A,B>=(<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
type Value=Pair<1>;
type First=Assert<Equal<Value[0],1>>;
type Cycle=Assert<Equal<Value[1]["right"]["left"],Value[1]>>;
