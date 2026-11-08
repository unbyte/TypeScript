export {};
type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
declare const key: unique symbol;
typefn Field<K,V,O,R> { return types.object<[{key:K;value:V;optional:O;readonly:R}]>; }
typefn Build<P> { return types.object<P>; }
typefn Empty<> { return types.object<[],[],[],[]>; }
typefn Extra<V> { return types.object<[{key:"x";value:V;read:number;optional:false;readonly:false}]>; }
typefn Keys<V> { return keyof Field<"x",V,false,false>; }
type KeyProof<V> = Assert<Equal<Keys<V>,"x">>;
type Later<K,V> = Field<K,V,false,true>;
type LaterFlags<O,R> = Field<"x",string,O,R>;
type Descriptor = {key:"x";value:1}&{optional:false;readonly:false};
type Checks = [
 Assert<Equal<Field<"x",string,false,false>,{x:string}>>,
 Assert<Equal<Field<"x",string,true,true>,{readonly x?:string}>>,
 Assert<Equal<Field<"x",string|undefined,true,false>,{x?:string|undefined}>>,
 Assert<Equal<Field<1,never,false,false>,{1:never}>>,
 Assert<Equal<Later<typeof key,1>,{readonly [key]:1}>>, Assert<Equal<keyof Later<typeof key,1>,typeof key>>,
 Assert<Equal<Later<"y",2>,{readonly y:2}>>, Assert<Equal<LaterFlags<true,false>,{x?:string}>>,
 Assert<Equal<Build<[Descriptor]>,{x:1}>>, Assert<Equal<Extra<1>,{x:1}>>, Assert<Equal<Empty<>,{}>>
];
