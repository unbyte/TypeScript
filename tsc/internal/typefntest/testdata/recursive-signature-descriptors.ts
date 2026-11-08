export {};
type Assert<T extends true> = T;
type Equal<A,B>=(<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Node<T> {return types.recursive<typefn<Self> {
 type update = types.signature<{
  kind:"call"; parameters:[
   {name:"value";value:T;mode:"required"},
   {name:"next";value:Self;mode:"optional"},
   {name:"tail";value:types.array<Self>;mode:"rest"}
  ]; this:[Self];returns:Self;abstract:false;predicate:[]
 }>;
 type check = types.signature<{
  kind:"call";parameters:[{name:"value";value:unknown;mode:"required"}];
  this:[];returns:boolean;abstract:false;predicate:[{kind:"is";target:0;narrowed:[Self]}]
 }>;
 type assert = types.signature<{
  kind:"call";parameters:[{name:"value";value:unknown;mode:"required"}];
  this:[];returns:void;abstract:false;predicate:[{kind:"asserts";target:0;narrowed:[Self]}]
 }>;
 type truthy = types.signature<{
  kind:"call";parameters:[{name:"value";value:Self|undefined;mode:"required"}];
  this:[];returns:void;abstract:false;predicate:[{kind:"asserts";target:0;narrowed:[]}]
 }>;
 type selfCheck = types.signature<{
  kind:"call";parameters:[];this:[Self];returns:boolean;abstract:false;
  predicate:[{kind:"is";target:"this";narrowed:[Self&{ready:true}]}]
 }>;
 type ctor = types.signature<{
  kind:"construct";parameters:[{name:"value";value:T;mode:"required"},{name:"next";value:Self;mode:"optional"}];
  this:[];returns:Self;abstract:true;predicate:[]
 }>;
 return {value:T;update:update;check:check;assert:assert;truthy:truthy;selfCheck:selfCheck;ctor:ctor};
}>;}
type N=Node<1>;
typefn Describe<T>{return reflect.signature<T>;}
type D=Describe<N['update']>;
type Names=Assert<Equal<D['parameters'][number]['name'],"value"|"next"|"tail">>;
type Modes=Assert<Equal<D['parameters'][number]['mode'],"required"|"optional"|"rest">>;
type Optional=Assert<Equal<NonNullable<D['parameters'][1]['value']>['value'],1>>;
type Rest=Assert<Equal<D['parameters'][2]['value'][number]['value'],1>>;
type Receiver=Assert<Equal<ThisParameterType<N['update']>['value'],1>>;
type Result=Assert<Equal<ReturnType<N['update']>['value'],1>>;
type Constructor=Assert<Equal<InstanceType<N['ctor']>['value'],1>>;
type Abstract=Assert<Equal<Describe<N['ctor']>['abstract'],true>>;
type Generic<T>=Node<T>;
type Replay=Assert<Equal<ReturnType<Generic<2>['update']>['value'],2>>;
declare const n:N;
const updated=n.update(1,n,n);
type Updated=Assert<Equal<typeof updated['value'],1>>;
declare let unknownValue:unknown;
if(n.check(unknownValue)){const exact:1=unknownValue.value;}
n.assert(unknownValue);
const exact:1=unknownValue.value;
declare let maybe:N|undefined;
n.truthy(maybe);
const defined:1=maybe.value;
if(n.selfCheck()){const ready:true=n.ready;}

typefn Callback<> {return types.recursive<typefn<Self>{
 return types.signature<{kind:"call";parameters:[{name:"value";value:Self;mode:"required"}];this:[];returns:void;abstract:false;predicate:[]}>;
}>;}
type Narrow=(value:Callback<>&{extra:1})=>void;
type OrdinaryVariance=Assert<Narrow extends Callback<>?false:true>;
// Reached checks may use a reference whose final definition is an ordinary Type.
typefn Discard<T> {return types.recursive<typefn<Self>{
 type unused=types.signature<{kind:"call";parameters:[{name:"x";value:T;mode:"required"}];this:[];returns:boolean;abstract:false;predicate:[{kind:"is";target:0;narrowed:[Self]}]}>;
 return 1;
}>;}
type Later<T>=Discard<T>;
type Good=Assert<Equal<Later<number>,1>>;
