export {};
type Assert<T extends true> = T;
type Equal<A,B> = (<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Box<T> {return [T];}
typefn Wrap<T> {return {value:T};}
typefn Apply<F<_>,T> {return F<T>;}
typefn Compose<F<_>,G<_>> {return typefn<T> {return F<G<T>>;};}
typefn Pair<T> {return typefn<U> {return [T,U];};}
type Direct=Assert<Equal<Apply<Box,1>,[1]>>;
type Composed=Assert<Equal<Apply<Compose<Wrap,Box>,1>,{value:[1]}>>;
type Pending<T>=Apply<Pair<T>,2>;
type Captured=Assert<Equal<Pending<1>,[1,2]>>;
typefn Nested<T> {
 typefn Local<F<_>,U> {return [T,F<U>];}
 type F=Compose<Box,Wrap>;
 return Local<F,2>;
}
type Nest=Assert<Equal<Nested<1>,[1,[{value:2}]]>>;
typefn Choose<T> {if(T==1) {return Box;} return Wrap;}
typefn Deferred<T> {return Apply<Choose<T>,123>;}
type Choice<T>=Deferred<T>;
type Left=Assert<Equal<Choice<1>,[123]>>;
type Right=Assert<Equal<Choice<2>,{value:123}>>;
typefn Increment<T> {return T+1;}
typefn Replay<F<_>,T> {if(T==0) {return F<T>;} return F<T+1>;}
type ReplayPending<T>=Replay<Increment,T>;
type Replayed=Assert<Equal<ReplayPending<2>,4>>;
typefn WithBound<T extends string> {return [T];}
type Bound=Assert<Equal<Apply<WithBound,"ok">,["ok"]>>;
typefn Ignore<F<_>> {return 1;}
type UnusedBound=Assert<Equal<Ignore<WithBound>,1>>;
typefn Constant<> {return 123;}
typefn Invoke<F<>> {return F<>;}
type Zero=Assert<Equal<Invoke<Constant>,123>>;
typefn Recursive<F<_>,T>: type {if(T==0) {return F<T>;} return Recursive<F,T-1>;}
type Recursed=Assert<Equal<Recursive<Box,5>,[0]>>;
