export {};
type Assert<T extends true>=T;
type Equal<A,B>=(<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Tuple<T> {return types.recursive<typefn<Self> {
 return types.tuple<{elements:[
  {value:T;label:["value"];mode:"required"},
  {value:Self;label:["next"];mode:"optional"}
 ];readonly:true}>;
}>;}
type Native<T>=readonly [value:T,next?:Native<T>];
type Value=Assert<Equal<Tuple<1>[0],1>>;
type Next=Assert<Equal<NonNullable<Tuple<1>[1]>[0],1>>;
type Compatible=Assert<Tuple<1> extends Native<1>?true:false>;
type CompatibleBack=Assert<Native<1> extends Tuple<1>?true:false>;
type Readonly=Assert<Tuple<1> extends unknown[]?false:true>;
type Generic<T>=Tuple<T>;
type Mapped=Assert<Equal<NonNullable<Generic<2>[1]>[0],2>>;
typefn Describe<T> {return reflect.tuple<T>;}
type Label=Assert<Equal<Describe<Tuple<1>>['elements'][1]['label'],["next"]>>;
type Mode=Assert<Equal<Describe<Tuple<1>>['elements'][1]['mode'],"optional">>;
typefn Rest<T> {return types.recursive<typefn<Self> {
 return types.tuple<{elements:[
  {value:T;label:["head"];mode:"required"},
  {value:types.array<Self>;label:["rest"];mode:"rest"}
 ];readonly:false}>;
}>;}
type RestElement=Assert<Equal<Rest<3>[1][0],3>>;
typefn Variadic<T> {return types.recursive<typefn<Self> {
 return types.tuple<{elements:[{value:[T,Self];label:[];mode:"variadic"}];readonly:false}>;
}>;}
type VariadicElement=Assert<Equal<Variadic<4>[1][1][0],4>>;
typefn Counted<N> {return types.recursiveGroup<N,typefn<R> {
 type result=[];
 type count=tuple.length<R>;
 type index=0;
 while(index<count) {
  type one=tuple.slice<R,index,index+1>;
  result=[...result,{next:one[0]}];
  index=index+1;
 }
 return result;
}>;}
type CountedGeneric<N>=Counted<N>;
type GroupLength=Assert<Equal<CountedGeneric<3>['length'],3>>;
type GroupRoot=Assert<CountedGeneric<3>[2]['next'] extends CountedGeneric<3>[2]?true:false>;
type DeclaredOptional=Assert<Equal<Describe<Tuple<1>>['elements'][1]['value'],Tuple<1>>>;
