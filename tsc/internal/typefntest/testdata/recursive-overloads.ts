export {};
type Assert<T extends true> = T;
type Equal<A,B>=(<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
type Generic = <T>(value:T)=>T;
interface Method { run(value:string):void }
type MethodHandle = Method['run'];
typefn Calls<T> {return reflect.calls<T>;}
typefn Constructs<T> {return reflect.constructs<T>;}
typefn GenericFlag<T> {return reflect.isGeneric<T>;}
typefn Descriptor<T> {return reflect.signature<T>;}
typefn Functions<> {return types.recursive<typefn<Self> {
 return types.overload<[types.function<[],Self>,MethodHandle,Generic,Generic]>;
}>;}
type F = Functions<>;
type Count=Assert<Equal<Calls<F>['length'],4>>;
type Order0=Assert<Equal<Parameters<Calls<F>[0]>,[]>>;
type RecursiveResult=Assert<ReturnType<Calls<F>[0]> extends F?true:false>;
type Order1=Assert<Equal<Parameters<Calls<F>[1]>,[value:string]>>;
type Generic0=Assert<GenericFlag<Calls<F>[2]>>;
type Generic1=Assert<GenericFlag<Calls<F>[3]>>;
type Narrow=(value:"literal")=>void;
type MethodVariance=Assert<Narrow extends Calls<F>[1]?true:false>;
declare const fn:F;
const inferred=fn({literal:"yes"} as const);
type Inferred=Assert<Equal<typeof inferred,{readonly literal:"yes"}>>;
const explicit=fn<42>(42);
type Explicit=Assert<Equal<typeof explicit,42>>;
typefn Factory<T> {return types.recursive<typefn<Self> {
 return types.object<[{key:"tag";value:T;optional:false;readonly:true}],
  [{key:`child-${string}`;value:Self;readonly:false}],
  [types.function<[],Self>],
  [types.constructor<[T],Self,true>]
 >;
}>;}
type A=Factory<1>;
type B=Factory<2>;
type Call=Assert<Equal<ReturnType<A>['tag'],1>>;
type Instance=Assert<Equal<InstanceType<B>['tag'],2>>;
type Index=Assert<Equal<A['child-one']['tag'],1>>;
type Abstract=Assert<Equal<Descriptor<Constructs<A>[0]>['abstract'],true>>;
type ConstructorArgs=Assert<Equal<ConstructorParameters<B>,[2]>>;
typefn Nested<> {return types.recursive<typefn<Self> {
 type single=types.overload<[types.function<[],Self>]>;
 return types.overload<[single,single]>;
}>;}
type NestedCount=Assert<Equal<Calls<Nested<>>['length'],2>>;
// Free native leaves in a signature-only graph must substitute without properties.
typefn Substitute<T> {return types.recursive<typefn<Self> {
 return types.overload<[types.function<[Self],T>]>;
}>;}
type Later<T>=Substitute<T>;
type First=Assert<Equal<ReturnType<Later<1>>,1>>;
type Second=Assert<Equal<ReturnType<Later<2>>,2>>;
type RecursiveParameter=Assert<Equal<ReturnType<Parameters<Later<2>>[0]>,2>>;
