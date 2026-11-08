export {};
type Assert<T extends true> = T;
type Equal<A,B> = (<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Next<N,Value extends N+1> {return Value;}
type NextValue=Assert<Equal<Next<1,2>,2>>;
type Pending<N,V>=Next<N,V>;
type Replay=Assert<Equal<Pending<2,3>,3>>;
typefn Suffix<Prefix,Value extends text.concat<[Prefix,"!"]>> {return Value;}
type SuffixValue=Assert<Equal<Suffix<"hello","hello!">,"hello!">>;
typefn Field<T,K extends keyof T> {return T[K];}
type FieldValue=Assert<Equal<Field<{x:123},"x">,123>>;
typefn Native<T,U extends type({value:T})> {return U;}
type NativeValue=Assert<Equal<Native<123,{value:123;extra:true}>,{value:123;extra:true}>>;
typefn Local<T> {
 typefn Inner<N,Value extends N+1> {return [T,Value];}
 type F=typefn<N,Value extends N*2> {return Inner<Value,Value+1>;};
 return F<2,4>;
}
type LocalValue=Assert<Equal<Local<1>,[1,5]>>;
// These bounds are well-bound Type expressions, but invalid if evaluated.
// An unused definition must not evaluate its input bound.
typefn Unused<T extends Uppercase<number>> {return T;}
typefn AlsoUnused<T extends numeric.toNumber<"bad">> {return T;}
typefn Fail<T>: type {fail<"bound executed">;}
typefn NeverApplied<T extends Fail<1>> {return T;}
