export {};
type Assert<T extends true> = T;
type Equal<A,B>=(<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Tree<T> { return types.recursive<typefn<Self> {
 return types.object<[{key:"tag";value:T;optional:false;readonly:false}], [
  {key:`child-${string}`;value:Self;readonly:true},
  {key:symbol;value:Self;readonly:false},
  {key:number;value:Self;readonly:false}
 ]>;
}>; }
declare const key:unique symbol;
type Pattern = Assert<Equal<Tree<1>["child-a"]["child-b"]["tag"],1>>;
type Symbol = Assert<Equal<Tree<1>[typeof key]["tag"],1>>;
type Number = Assert<Equal<Tree<1>[0]["tag"],1>>;
type Generic<T> = Tree<T>;
type Replay = Assert<Equal<Generic<2>["child-a"][typeof key][0]["tag"],2>>;
typefn Indexers<T> { return reflect.indexers<T>; }
type Entries = Indexers<Tree<1>>;
type PatternInfo = Extract<Entries[number],{key:`child-${string}`}>;
type Flag = Assert<Equal<PatternInfo['readonly'],true>>;
type Value = Assert<Equal<PatternInfo['value']['tag'],1>>;
typefn Overlap<> {return types.recursive<typefn<Self> {
 return types.object<[],[
  {key:string;value:Self;readonly:false},
  {key:number;value:Self;readonly:true}
 ]>;
}>;}
interface Expected { [key:string]:Expected; readonly [key:number]:Expected }
type Native = Assert<Overlap<> extends Expected ? true:false>;
type Reverse = Assert<Expected extends Overlap<> ? true:false>;
// A reached but discarded check still executes, without adding phantom leaves.
typefn Discard<T> { return types.recursive<typefn<Self> {
 type unused = types.object<[{key:"x";value:T;optional:false;readonly:false}],[{key:string;value:T|Self;readonly:false}]>;
 return 1;
}>; }
type NoDependency<T> = Assert<Equal<Discard<T>,1>>;
