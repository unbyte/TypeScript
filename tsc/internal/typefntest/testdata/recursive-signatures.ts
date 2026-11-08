export {};
type Assert<T extends true>=T;
type Equal<A,B>=(<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Chain<T> {return types.recursive<typefn<Self> {
 return types.function<[T],Self>;
}>;}
type NativeChain<T>=(value:T)=>NativeChain<T>;
type Args=Assert<Equal<Parameters<Chain<1>>,[value:1]>>;
type Returns=Assert<Equal<Parameters<ReturnType<Chain<1>>>,[value:1]>>;
type Forward=Assert<Chain<1> extends NativeChain<1>?true:false>;
type Back=Assert<NativeChain<1> extends Chain<1>?true:false>;
type Generic<T>=Chain<T>;
type Replay=Assert<Equal<Parameters<ReturnType<Generic<2>>>,[value:2]>>;
typefn Consumer<T> {return types.recursive<typefn<Self> {
 return types.function<[Self],T>;
}>;}
type Consumed=Assert<Equal<ReturnType<Parameters<Consumer<3>>[0]>,3>>;
typefn Construct<T,A> {return types.recursive<typefn<Self> {
 return types.constructor<[T],{value:T;again:Self},A>;
}>;}
type Constructed=Assert<Equal<InstanceType<Construct<4,false>>['value'],4>>;
type Reconstructed=Assert<Equal<ConstructorParameters<InstanceType<Construct<4,false>>['again']>,[4]>>;
type AbstractConstruct=Assert<Construct<4,true> extends new(...args:any[])=>any?false:true>;
type ConstructGeneric<T>=Construct<T,false>;
type NativeNew=new (value:5)=>{value:5;again:NativeNew};
type NativeNewMatches=Assert<NativeNew extends ConstructGeneric<5>?true:false>;
typefn NativeArray<> {return types.recursive<Array>;}
type ArrayLoops=Assert<NativeArray<>[0] extends NativeArray<>?true:false>;
typefn NativeReadonly<> {return types.recursive<ReadonlyArray>;}
type ReadonlyLoops=Assert<NativeReadonly<>[0] extends NativeReadonly<>?true:false>;
type IsReadonly=Assert<NativeReadonly<> extends unknown[]?false:true>;
typefn ArrayInObject<T> {return types.recursive<typefn<Self> {return {value:T;items:Array<Self>};}>;}
type ArrayMember=Assert<Equal<ArrayInObject<6>['items'][0]['value'],6>>;
typefn FunctionGroup<> {return types.recursiveGroup<2,typefn<R> {
 return [types.function<[],R[1]>,types.function<[R[0]],1>];
}>;}
type GroupReturn=Assert<Equal<ReturnType<ReturnType<FunctionGroup<>[0]>>,1>>;
