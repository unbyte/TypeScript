import type { A, B, C, D, E, F, G } from "./numbers";
type Equal<X, Y> = (<T>() => T extends X ? 1 : 2) extends (<T>() => T extends Y ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
type Checks = [Assert<Equal<A,true>>, Assert<Equal<B,2.25>>, Assert<Equal<C,100>>, Assert<Equal<D,2|8>>, Assert<Equal<E,42>>, Assert<Equal<F,0|1|2|3>>, Assert<Equal<G,true>>];
