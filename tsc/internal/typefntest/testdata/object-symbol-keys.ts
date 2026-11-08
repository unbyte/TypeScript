import {key,other} from "./keys";
type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
typefn Field<K,T> { return {readonly [K]?:T}; }
typefn Snapshot<K,J,T> { type name=K; type result={[name]:T}; name=J; return result; }
type Later<T> = Field<typeof key,T>;
type Frozen<T> = Snapshot<typeof key,typeof other,T>;
type Checks = [
 Assert<Equal<Field<typeof key,1>,{readonly [key]?:1}>>,
 Assert<Equal<keyof Later<string>,typeof key>>, Assert<Equal<Later<string>[typeof key],string|undefined>>,
 Assert<Equal<keyof Frozen<1>,typeof key>>, Assert<Equal<Frozen<1>[typeof key],1>>
];
