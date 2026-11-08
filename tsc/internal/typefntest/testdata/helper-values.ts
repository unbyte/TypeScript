export {};
type Assert<T extends true> = T;
type Equal<A,B> = (<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Box<T> {return [T];}
typefn Record<T> {return {value:T};}
typefn Empty<> {return 42;}
typefn Binding<T> {type F=Box; type G=F; return G<T>;}
type Bound=Assert<Equal<Binding<123>,[123]>>;
type Deferred<T>=Binding<T>;
type Replayed=Assert<Equal<Deferred<456>,[456]>>;
typefn Zero<> {type F=Empty; return F<>;}
type ZeroArity=Assert<Equal<Zero<>,42>>;
typefn Reassign<T> {type F=Box; F=Record; return F<T>;}
type Reassigned=Assert<Equal<Reassign<123>,{value:123}>>;
typefn Choose<T> {type F=Box; if(T == "record") {F=Record;} return F<T>;}
type Choice<T>=Choose<T>;
type BoxChoice=Assert<Equal<Choice<"box">,["box"]>>;
type RecordChoice=Assert<Equal<Choice<"record">,{value:"record"}>>;
typefn Snapshot<T> {
 type F=Box;
 typefn Inner<U> {return F<[T,U]>;}
 F=Record;
 type G=Inner;
 return G<1>;
}
type Frozen<T>=Snapshot<T>;
type FrozenReplay=Assert<Equal<Frozen<2>,[[2,1]]>>;
typefn Local<T> {
 type prefix=T;
 typefn First<U> {return [prefix,U];}
 type F=First;
 prefix="changed";
 typefn Second<U> {return {value:U};}
 type before=F<1>;
 F=Second;
 return [before,F<2>];
}
type Locals<T>=Local<T>;
type LocalReplay=Assert<Equal<Locals<3>,[[3,1],{value:2}]>>;
typefn Bounded<T extends string> {return T;}
typefn Good<T> {type F=Box; F=Bounded; return F<T>;}
type BoundOK=Assert<Equal<Good<"yes">,"yes">>;
type Defaulted<T=123> = [T];
typefn NativeDefault<> {type T=Defaulted; return T;}
type NativeDefaultValue=Assert<Equal<NativeDefault<>,[123]>>;
typefn Countdown<T>: type {if(T == 0) {return 0;} type F=Countdown; return F<T-1>;}
type Recursive=Assert<Equal<Countdown<5>,0>>;
