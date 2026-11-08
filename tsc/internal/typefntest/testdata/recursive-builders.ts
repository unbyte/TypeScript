export {};
type Assert<T extends true>=T;
type Equal<A,B>=(<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn ListNode<Item> {
 type Build=typefn<Self> {return {value:Item;next:Self|null};};
 return types.recursive<Build>;
}
type List=ListNode<1>;
type Value=Assert<Equal<List['value'],1>>;
type Link=Assert<Equal<NonNullable<List['next']>['value'],1>>;
type Circle=Assert<Equal<NonNullable<List['next']>['next'],List['next']>>;
type NativeList<T>={value:T;next:NativeList<T>|null};
type NativeCompatible=Assert<List extends NativeList<1>?true:false>;
type NativeCompatibleBack=Assert<NativeList<1> extends List?true:false>;
type Generic<T>=ListNode<T>;
type Mapped=Assert<Equal<NonNullable<Generic<2>['next']>['value'],2>>;
type MappedAgain=Assert<Equal<NonNullable<Generic<3>['next']>['value'],3>>;
typefn Ignore<> {return types.recursive<typefn<Self> {return 42;}>;}
type NoSelf=Assert<Equal<Ignore<>,42>>;
typefn Pair<T> {
 return types.recursiveGroup<2,typefn<Refs> {
  return [{value:T;other:Refs[1]},{other:Refs[0]}];
 }>;
}
type PairValue=Assert<Equal<Pair<4>[1]['other']['value'],4>>;
type PairCycle=Assert<Equal<Pair<4>[0]['other']['other']['value'],4>>;
typefn Nested<T> {return types.recursive<typefn<Outer> {
 return {value:T;child:types.recursive<typefn<Inner> {return {up:Outer;again:Inner};}>};
}>;}
type NestedCycle=Assert<Equal<Nested<5>['child']['again']['up']['value'],5>>;
typefn Tree<T> {return types.recursive<typefn<Self> {return T | [Self,Self];}>;}
type RecursiveTuple=Assert<Extract<Tree<1>,unknown[]>[0] extends Tree<1>?true:false>;
typefn ArrayTree<T> {return types.recursive<typefn<Self> {return {value:T;children:types.array<Self>};}>;}
type ArrayChild=Assert<Equal<ArrayTree<6>['children'][0]['value'],6>>;
typefn Deferred<N> {return types.recursive<typefn<Self> {if(N==1) {return {next:Self;value:1};} return {next:Self;value:2};}>;}
type DeferredOne=Assert<Equal<Deferred<1>['next']['value'],1>>;
type DeferredTwo=Assert<Equal<Deferred<2>['next']['value'],2>>;

type DeferredGeneric<N>=Deferred<N>;
type DeferredReplay=Assert<Equal<DeferredGeneric<1>['next']['value'],1>>;
type DeferredReplayTwo=Assert<Equal<DeferredGeneric<2>['next']['value'],2>>;
typefn PureCycle<> {return types.recursive<typefn<Self> {return {next:Self};}>;}
type NoLeaves=Assert<PureCycle<>['next'] extends PureCycle<>?true:false>;
typefn CopyGroup<N> {return types.recursiveGroup<N,typefn<R> {
 type result=[];
 for(type self of R) {result=[...result,{next:self}];}
 return result;
}>;}
type DynamicGroup<N>=CopyGroup<N>;
type ReplayedCount=Assert<Equal<DynamicGroup<3>['length'],3>>;
type GroupCycle=Assert<DynamicGroup<3>[2]['next'] extends DynamicGroup<3>[2]?true:false>;
// A reached valid nested builder may be discarded, without adding its leaves
// to the returned graph's native generic environment.
typefn Discarded<T> {return types.recursive<typefn<Outer> {
 type unused=types.recursive<typefn<Inner> {return {next:Inner;value:T};}>;
 return {next:Outer};
}>;}
type DiscardedGeneric<T>=Discarded<T>;
type DiscardedValid=Assert<DiscardedGeneric<1>['next'] extends DiscardedGeneric<2>?true:false>;
typefn DiscardedLeaf<T> {return types.recursive<typefn<Outer> {
 type unused=types.recursive<typefn<Inner> {return {next:Inner;value:T};}>;
 return 42;
}>;}
type DiscardedLeafGeneric<T>=DiscardedLeaf<T>;
type DiscardedLeafResult=Assert<Equal<DiscardedLeafGeneric<1>,42>>;
typefn ClassifyDiscarded<T> {return reflect.kind<Discarded<T>>;}
type NoPhantomDependency<T>=Assert<Equal<ClassifyDiscarded<T>,"object">>;
