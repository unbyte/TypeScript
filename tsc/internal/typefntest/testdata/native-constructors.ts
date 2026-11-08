export {};
type Assert<T extends true> = T;
type Equal<A,B> = (<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Apply<F<_>,T> {return F<T>;}
typefn Apply2<F<_,_>,T,U> {return F<T,U>;}
typefn Compose<F<_>,G<_>> {return typefn<T> {return F<G<T>>;};}
type Composed=Assert<Equal<Apply<Compose<Promise,Array>,string>,Promise<string[]>>>;
type Picked=Assert<Equal<Apply2<Pick,{a:1,b:2},"a">,{a:1}>>;
typefn MapTypes<Types,F<_>> {type result=[]; for(type T of Types) {result=[...result,F<T>];} return result;}
type Mapped=Assert<Equal<MapTypes<[string,number],Promise>,[Promise<string>,Promise<number>]>>;
type Defaulted<T=string>=[T];
typefn Defaults<> {type original=Defaulted; type F=typefn<T> {return T;}; F=Defaulted; return [original,F<number>];}
type DefaultsResult=Assert<Equal<Defaults<>,[[string],[number]]>>;
typefn ReturnArray<>: typefn<_> {return Array;}
type Returned=Assert<Equal<Apply<ReturnArray<>,1>,1[]>>;
interface Node<T> {readonly value:T; next?:Node<T>}
type Interface=Assert<Equal<Apply<Node,string>,Node<string>>>;
class Nominal<T> {private brand!:void; value!:T;}
type NominalResult=Assert<Equal<Apply<Nominal,number>,Nominal<number>>>;
type Remap<T>={[K in keyof T as `get${Capitalize<string & K>}`]:T[K]};
type Alias=Assert<Equal<Apply<Remap,{name:string}>,{getName:string}>>;
type Intrinsic=Assert<Equal<Apply<Uppercase,"hello">,"HELLO">>;
type Pending<T>=Apply<Array,T>;
type Specialized=Assert<Equal<Pending<1>,1[]>>;
typefn Select<F<_>,T> {if(T==0) {return F<T>;} return F<T+1>;}
type Blocked<T>=Select<Array,T>;
type Replayed=Assert<Equal<Blocked<1>,2[]>>;
typefn Capture<T> {type F=Compose<Array,Promise>; return typefn<U> {return [T,F<U>];};}
type Captured<T>=Apply<Capture<T>,1>;
type CapturedResult=Assert<Equal<Captured<2>,[2,Promise<1>[]]>>;
type Bound<T extends string>=[T];
typefn Ignore<F<_>> {return 1;}
type Ignored=Assert<Equal<Ignore<Bound>,1>>;
type Good=Assert<Equal<Apply<Bound,"ok">,["ok"]>>;
type ZeroAlias=123;
typefn Invoke<F<>> {return F<>;}
type Zero=Assert<Equal<Invoke<ZeroAlias>,123>>;
function outer<T>() {
 type Local<U extends T>=[T,U];
 type Result=Apply<Local,T>;
 return null as unknown as Result;
}
type Outer=Assert<Equal<ReturnType<typeof outer<1>>,[1,1]>>;
type Pair<T,U=string>=[T,U];
type FullArity=Assert<Equal<Apply2<Pair,1,2>,[1,2]>>;
function capturedBound<T>() {
 type Local<U extends T>=[T,U];
 let result!:Apply<Local,1>;
 return result;
}
type MappedBound=Assert<Equal<ReturnType<typeof capturedBound<1>>,[1,1]>>;
