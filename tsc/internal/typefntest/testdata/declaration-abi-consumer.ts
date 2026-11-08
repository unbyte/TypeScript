import type {Twice,Saved} from './library';
type Assert<T extends true>=T;
type Equal<A,B>=(<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
type Direct=Assert<Equal<Twice<21>,42>>;
type Generic<T>=Twice<T>;
type Replay=Assert<Equal<Generic<22>,44>>;
type Closure=Assert<Equal<Saved<number>,["saved",number]>>;
