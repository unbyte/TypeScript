export {};
type Assert<T extends true> = T;
type Equal<A,B> = (<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Make<T> {return typefn<U> {return [T,U];};}
typefn Apply<T,U> {type F=Make<T>; return F<U>;}
type Pending<T,U>=Apply<T,U>;
type Partial<T>=Pending<T,2>;
type Applied=Assert<Equal<Partial<1>,[1,2]>>;
typefn Box<T> {return [T];}
typefn Named<>: typefn<_> {return Box;}
typefn ApplyNamed<T> {type F=Named<>; return F<T>;}
type NamedResult=Assert<Equal<ApplyNamed<1>,[1]>>;
typefn Local<T> {
 typefn Factory<U> {type captured=U; type F=typefn<V> {return [T,captured,V];}; captured="changed"; return F;}
 type F=Factory<2>;
 return F<3>;
}
type LocalPending<T>=Local<T>;
type LocalResult=Assert<Equal<LocalPending<1>,[1,2,3]>>;
typefn Choose<T> {if(T == "yes") {return typefn<U> {return [U];};} return typefn<U> {return {value:U};};}
typefn UseChoice<T> {type F=Choose<T>; return F<1>;}
type Choice<T>=UseChoice<T>;
type ChoiceYes=Assert<Equal<Choice<"yes">,[1]>>;
type ChoiceNo=Assert<Equal<Choice<"no">,{value:1}>>;
typefn Countdown<T>: typefn<_> {if(T == 0) {return Box;} return Countdown<T-1>;}
typefn Recursive<> {type F=Countdown<5>; return F<123>;}
type Recursed=Assert<Equal<Recursive<>,[123]>>;
typefn Zero<T>: typefn<> {return typefn<> {return T;};}
typefn UseZero<T> {type F=Zero<T>; return F<>;}
type ZeroResult=Assert<Equal<UseZero<123>,123>>;
typefn SafeCapture<T> {return typefn<> {return T+1;};}
typefn DoNotCall<> {type F=SafeCapture<"invalid if called">; return 1;}
type Uncalled=Assert<Equal<DoNotCall<>,1>>;
typefn Validated<T> {type check=T+1; return Box;}
typefn Discard<T> {Validated<T>; return 123;}
type Discarded<T>=Discard<T>;
type DiscardedValid=Assert<Equal<Discarded<1>,123>>;
typefn Reassign<T> {type F=Make<1>; F=Make<T>; return F<2>;}
type Reassigned=Assert<Equal<Reassign<3>,[3,2]>>;

typefn ReturnHelper<T> {
 typefn Recursive<N>: type {if(N == 0) {return T;} return Recursive<N-1>;}
 return Recursive;
}
typefn UseReturnedHelper<T> {type F=ReturnHelper<T>; return F<4>;}
type ReturnedHelperPending<T>=UseReturnedHelper<T>;
type ReturnedHelper=Assert<Equal<ReturnedHelperPending<123>,123>>;
// Constructor creation waits for the input's checks, including through a local.
// The invalid closure application must not execute in this generic definition.
typefn Check<T> {type check=T+1; return 1;}
typefn NeedsReady<T> {return typefn<U> {return U+1;};}
typefn Waiting<T> {type checked=Check<T>; type F=NeedsReady<checked>; return F<"invalid after creation">;}
type StillWaiting<T>=Waiting<T>;
