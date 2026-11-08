export {};
type Assert<T extends true> = T;
type Equal<A,B> = (<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Fields<T> { return reflect.properties<T>; }
type Field<T,K> = Extract<Fields<T>[number], {key:K}>;
type Source<T> = {readonly first:T; optional?:string; explicit?:string|undefined; required:undefined; empty?:never; 42: true};
type Optional = Assert<Equal<Field<Source<1>,"optional">,{key:"optional";value:string;read:string|undefined;optional:true;readonly:false}>>;
type Explicit = Assert<Equal<Field<Source<1>,"explicit">['value'],string|undefined>>;
type Required = Assert<Equal<Field<Source<1>,"required">['value'],undefined>>;
type Empty = Assert<Equal<Field<Source<1>,"empty">['value'],never>>;
type EmptyRead = Assert<Equal<Field<Source<1>,"empty">['read'],undefined>>;
type Numeric = Assert<Equal<Field<Source<1>,"42">['value'],true>>;
type Order = Assert<Equal<Fields<Source<1>>[0]['key'],"first">>;
type Pending<T> = Fields<Source<T>>;
type Replay = Assert<Equal<Pending<123>[0]['value'],123>>;
declare const unique: unique symbol;
type Unique = Assert<Equal<Field<{[unique]?:1},typeof unique>['value'],1>>;
typefn Built<T> { return {readonly first:T; optional?:string; explicit?:string|undefined; empty?:never}; }
type BuiltOptional = Assert<Equal<Field<Built<1>,"optional">['value'],string>>;
type BuiltExplicit = Assert<Equal<Field<Built<1>,"explicit">['value'],string|undefined>>;
typefn OptionalValue<T> { return {value?:T}; }
type Captured<T> = Fields<OptionalValue<T>>;
type CapturedValue = Assert<Equal<Captured<"x">[0]['value'],"x">>;
type CapturedUndefined = Assert<Equal<Captured<undefined>[0]['value'],undefined>>;
type CapturedNever = Assert<Equal<Captured<never>[0]['value'],never>>;
type Picked = Assert<Equal<Field<Pick<Source<1>,"optional"|"explicit">,"optional">['value'],string>>;
type PartialRequired = Assert<Equal<Field<Partial<{value:string}>,"value">['value'],string>>;
type PartialOptional = Assert<Equal<Field<Partial<{value?:string}>,"value">['value'],string>>;
type ReadonlyOptional = Assert<Equal<Field<Readonly<{value?:string|undefined}>,"value">['value'],string|undefined>>;
type Rename<T> = {[K in keyof T as `new_${K & string}`]:T[K]};
type Remapped = Assert<Equal<Field<Rename<{value?:string}>,"new_value">['value'],string>>;
type Synthesized = Assert<Equal<Field<{[K in "a"]?: string},"a">['value'],string>>;
type Intersected = Assert<Equal<Field<{value?:string|number}&{value?:string|boolean},"value">['value'],string>>;
interface Base<T> { inherited?:T; }
class Derived implements Base<number> {
 inherited?:number;
 public visible=1;
 protected protectedField=2;
 private privateField=3;
 #hidden=4;
 get fixed():string {return "x";}
 get same():number {return 1;}
 set same(value:number) {}
 method?():void;
}
type PublicKeys = Assert<Equal<Fields<Derived>[number]['key'],"inherited"|"visible"|"fixed"|"same"|"method">>;
type Getter = Assert<Equal<Field<Derived,"fixed">['readonly'],true>>;
type Accessor = Assert<Equal<Field<Derived,"same">['readonly'],false>>;
type Method = Assert<Equal<Field<Derived,"method">['value'],()=>void>>;
interface Extended<T> extends Base<T> { own:true; }
type Inherited = Assert<Equal<Field<Extended<123>,"inherited">['value'],123>>;
typefn ReadonlyPublic<T> {
 type output=[];
 for(type field of reflect.properties<T>) {
  output=[...output,{key:field["key"];value:field["value"];optional:field["optional"];readonly:true}];
 }
 return types.object<output>;
}
type Projection=ReadonlyPublic<{a:string;b?:number;c?:boolean|undefined}>;
type Rebuild=Assert<Equal<Projection,{readonly a:string;readonly b?:number;readonly c?:boolean|undefined}>>;

declare const left: {value?:string};
declare const right: {value?:number};
const spread = {...left,...right};
type Spread = Assert<Equal<Field<typeof spread,"value">['value'],string|number>>;
declare const maybeObject: {value:string}|undefined;
const optionalSpread = {...maybeObject};
type OptionalSpread = Assert<Equal<Field<typeof optionalSpread,"value">['value'],string>>;
class OptionalInitializer { value?=1; readonly literal?=2; }
type InitialValue = Assert<Equal<Field<OptionalInitializer,"value">['value'],number>>;
type InitialLiteral = Assert<Equal<Field<OptionalInitializer,"literal">['value'],2>>;
