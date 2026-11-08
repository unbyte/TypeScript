export {};
type Assert<T extends true> = T;
type Equal<A,B> = (<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Box<T> {return [T];}
typefn Apply<F<_>,T> {return F<T>;}
typefn Compose<F<_>,G<_>> {return typefn<T> {return F<G<T>>;};}
typefn Use<F<_>,T> {return type([F<T>,Apply<F,T>]);}
type Direct=Assert<Equal<Use<Array,1>,[1[],1[]]>>;
type Pending<T>=Use<Promise,T>;
type Replayed=Assert<Equal<Pending<1>,[Promise<1>,Promise<1>]>>;
typefn Known<T> {return type(Apply<Compose<Promise,Array>,T>);}
type Named=Assert<Equal<Known<1>,Promise<1[]>>>;
typefn Local<T> {
 type captured=T;
 typefn Make<U> {return typefn<V> {return [captured,U,V];};}
 typefn Invoke<F<_>,U> {return F<U>;}
 type F=Make<2>;
 type before=type([F<3>,Invoke<F,4>,Apply<Make<5>,6>]);
 captured="changed";
 return before;
}
type Captures=Assert<Equal<Local<1>,[[1,2,3],[1,2,4],[1,5,6]]>>;
type LocalPending<T>=Local<T>;
type ReplayedCaptures=Assert<Equal<LocalPending<7>,[[7,2,3],[7,2,4],[7,5,6]]>>;
typefn Mutable<T> {type F=Box; type result=type(F<T>); F=Array; return [result,type(F<T>)];}
type Snapshot=Assert<Equal<Mutable<1>,[[1],1[]]>>;
typefn Shadow<F<_>> {return type(<F>()=>F);}
type Shadowed=Assert<Equal<Shadow<Box>,<T>()=>T>>;
typefn Recursive<N> {
 typefn Count<T>: type {if(T==0) {return T;} return type(Count<0>);}
 return type(Count<N>);
}
type Self=Assert<Equal<Recursive<1>,0>>;
typefn Empty<T> {typefn F<> {return T;} return type(F<>);}
type Zero=Assert<Equal<Empty<1>,1>>;
typefn Bounded<T> {typefn F<U extends T> {return U;} return type(F<T>);}
type Bounds=Assert<Equal<Bounded<1>,1>>;
typefn Parentheses<T> {return type(Apply<(Array),T>);}
type Parenthesized=Assert<Equal<Parentheses<1>,1[]>>;
typefn ShadowFactory<T> {return typefn<> {return type(<T>()=>T);};}
typefn UseShadow<T> {type F=ShadowFactory<T>;return type(F<>);}
// No phantom capture of the outer T may defer this ready native signature.
type ShadowReady<T>=Assert<Equal<UseShadow<T>,<U>()=>U>>;
