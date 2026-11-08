import type {Saved} from './out/library';
type Assert<T extends true>=T;
type Equal<A,B>=(<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
type Check=Assert<Equal<Saved<number>,["saved",number]>>;
