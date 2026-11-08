export {};
type Assert<T extends true>=T;
type Equal<A,B>=(<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
declare const key:unique symbol;
declare const other:unique symbol;
typefn Object<T> {return types.recursive<typefn<Self> {
 return types.object<[
 {key:"value";value:T;optional:false;readonly:true},
 {key:type(typeof key);value:Self;optional:true;readonly:false}
 ],[],[],[]>;
}>;}
type Value=Assert<Equal<Object<1>['value'],1>>;
type Next=Assert<Equal<NonNullable<Object<1>[typeof key]>['value'],1>>;
type Keys=Assert<Equal<keyof Object<1>,"value"|typeof key>>;
type Different=Assert<typeof other extends keyof Object<1>?false:true>;
type Generic<T>=Object<T>;
type Replay=Assert<Equal<NonNullable<Generic<2>[typeof key]>['value'],2>>;
typefn Properties<T> {return reflect.properties<T>;}
type Entries=Properties<Object<1>>;
type Entry=Extract<Entries[number],{key:typeof key}>;
type DeclaredOptional=Assert<Equal<Entry['value'],Object<1>>>;
type ReadOptional=Assert<Equal<Entry['read'],Object<1>|undefined>>;
type Readonly=Assert<Equal<Extract<Entries[number],{key:"value"}>['readonly'],true>>;
typefn Literal<T> {return types.recursive<typefn<Self> {return {[type(typeof key)]:Self;value:T};}>;}
type LiteralKey=Assert<Equal<Literal<3>[typeof key]['value'],3>>;
// Derived reflection output may carry Self but is ignored as an input value.
typefn Ignored<> {return types.recursive<typefn<Self> {
 return types.object<[{key:"value";value:1;optional:false;readonly:false;read:Self}]>;
}>;}
type IgnoredRead=Assert<Equal<Ignored<>,{value:1}>>;
