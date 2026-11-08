export {};
type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
type I<K,V,R=false>={key:K;value:V;readonly:R};
type P<K,V>={key:K;value:V;optional:false;readonly:false};
typefn Build<P,I> {return types.object<P,I>;}
typefn Describe<T> {return reflect.indexers<T>;}
typefn Copy<T> {return types.object<[],reflect.indexers<T>>;}
typefn Constrained<T extends string> {return types.object<[P<"x",T>],[I<string,string>]>;}
typefn Checked<T> {return types.object<[P<"x",T>],[I<string,string>]>;}
type Later<T>=Build<[],[I<string,T,true>]>;
type Projection<T>=Checked<T>["x"];
type BoundProof<T extends string>=Assert<Equal<Constrained<T>["x"],T>>;
declare const unique: unique symbol;
typefn Same<A,B> {return A == B;}
type CompareLater<T>=Same<Build<[],[I<string,T>]>,{[key:string]:string}>;
type Checks=[Assert<Equal<CompareLater<string>,true>>,Assert<Equal<CompareLater<number>,false>>,
 Assert<Equal<Build<[],[I<string,number>]>,{[key:string]:number}>>,
 Assert<Equal<Build<[P<"x",1>],[I<string,number,true>]>["x"],1>>,
 Assert<Equal<Build<[P<typeof unique,1>],[I<symbol,number>]>[typeof unique],1>>,
 Assert<Equal<Build<[],[I<`data-${string}`,boolean>]>["data-x"],boolean>>,
 Assert<Equal<Build<[],[I<string,string|number>,I<number,number>]>[123],number>>,
 Assert<Equal<Describe<{readonly [key:string]:number}>,[{key:string;value:number;readonly:true}]>>,
 Assert<Equal<Describe<{a:1}>,[]>>, Assert<Equal<Copy<{readonly [key:symbol]:string}>,{readonly [key:symbol]:string}>>,
 Assert<Equal<Later<1>["anything"],1>>, Assert<Equal<Describe<Later<1>>,[{key:string;value:1;readonly:true}]>>,
 Assert<Equal<Projection<"ok">,"ok">>, Assert<Equal<Describe<Build<[],[I<string,number>,I<symbol,boolean>]>>,[I<string,number>,I<symbol,boolean>]>>
];
