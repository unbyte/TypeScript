import type {Id,View,Handlers,StringList,Promises,Names,Field,ReadonlyPublic,ListNode,MapTypes,AsyncArray} from "./construction";
import {tag} from "./model";
import type {Secret,Callable,Indexed,Open,OpenView} from "./model";
type Assert<T extends true> = T;
type Equal<A,B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
declare module "./model" { interface Open { added:number; explicit?:string|undefined; } }
type IdType = Assert<Equal<Id,{id:number}>>;
type ViewType = Assert<Equal<View,{readonly name?:string;readonly count:number}>>;
type HandlerTypes = Assert<Equal<Handlers,{
  opened:(event:{kind:"opened";id:string})=>void;
  closed:(event:{kind:"closed";code:number})=>void;
}>>;
interface NativeList { value:string;next:NativeList|null; }
type ListToNative = Assert<StringList extends NativeList ? true : false>;
type NativeToList = Assert<NativeList extends StringList ? true : false>;
type RecursiveValue = Assert<Equal<NonNullable<StringList["next"]>["value"],string>>;
type NominalList = Assert<Equal<ListNode<Secret>["value"],Secret>>;
type PromiseTypes = Assert<Equal<Promises,[Promise<string>,Promise<number>]>>;
type ComposedType = Assert<Equal<Names,Promise<string[]>>>;
type ComposedNominal = Assert<Equal<AsyncArray<Secret>,Promise<Secret[]>>>;
type EmptyMap = Assert<Equal<MapTypes<[],Array>,[]>>;
type SymbolField = Assert<Equal<Field<typeof tag,Secret>,{[tag]:Secret}>>;
type NominalLeaf = Assert<Equal<ReadonlyPublic<{leaf:Secret}>["leaf"],Secret>>;
type PublicProjection = Assert<Equal<ReadonlyPublic<Secret>,{readonly value:string}>>;
type Signatures = Assert<Equal<ReadonlyPublic<Callable>,Callable>>;
type Indexers = Assert<Equal<ReadonlyPublic<Indexed>,Readonly<Indexed>>>;
type Augmented = Assert<Equal<OpenView,Readonly<Open>>>;
