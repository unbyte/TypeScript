export {};
type Assert<T extends true> = T;
type Equal<A,B> = (<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Snapshot<T> {
 type saved=T;
 type F=typefn<U>: type {return [saved,U];};
 saved="changed";
 return F<123>;
}
type Pending<T>=Snapshot<T>;
type Result=Assert<Equal<Pending<456>,[456,123]>>;
typefn Nested<T> {
 type F=typefn<U> {type G=typefn<V> {return [T,U,V];}; return G<1>;};
 return F<2>;
}
type NestedPending<T>=Nested<T>;
type NestedResult=Assert<Equal<NestedPending<3>,[3,2,1]>>;
typefn Shadow<T> {type F=typefn<T> {return T;}; return F<123>;}
type Shadowed=Assert<Equal<Shadow<0>,123>>;
typefn Reassign<T> {type F=typefn<U> {return [U];}; F=typefn<U> {return {value:U};}; return F<T>;}
type Reassigned=Assert<Equal<Reassign<123>,{value:123}>>;
typefn Zero<T> {type F=typefn<> {return T;}; return F<>;}
type ZeroArity=Assert<Equal<Zero<123>,123>>;
typefn Choices<T> {
 type F=typefn<> {if(T == "yes") {return 123;} return false;};
 return {value:F<>};
}
type ChoicePending<T>=Choices<T>;
type ChoiceTrue=Assert<Equal<ChoicePending<"yes">,{value:123}>>;
type ChoiceFalse=Assert<Equal<ChoicePending<"no">,{value:false}>>;
typefn Bounds<T> {type F=typefn<A,B extends A> {return [T,B];}; return F<string,"x">;}
type Bounded=Assert<Equal<Bounds<1>,[1,"x"]>>;
typefn CapturedBound<T> {type F=typefn<U extends T> {return U;}; return F<"x">;}
type BoundPending<T>=CapturedBound<T>;
type CapturedBoundResult=Assert<Equal<BoundPending<string>,"x">>;
typefn Islands<T> {type F=typefn<U> {return type({value:T; own:U});}; return F<1>;}
type IslandPending<T>=Islands<T>;
type IslandResult=Assert<Equal<IslandPending<2>,{value:2; own:1}>>;
typefn Loop<T> {type result=[]; for(type item of T) {type F=typefn<> {return item;}; result=[...result,F<>];} return result;}
type LoopResult=Assert<Equal<Loop<[1,2,3]>,[1,2,3]>>;
typefn Expr<T> {type F=typefn<U> {return U;}; return text.concat<[F<T>,"!"]>;}
type ExprPending<T>=Expr<T>;
type ExprResult=Assert<Equal<ExprPending<"ok">,"ok!">>;
typefn Named<T> {type F=typefn<U> {typefn Inner<V> {return [T,U,V];} return Inner<1>;}; return F<2>;}
type NamedResult=Assert<Equal<Named<3>,[3,2,1]>>;
