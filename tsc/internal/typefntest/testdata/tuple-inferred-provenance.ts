export {};
type Assert<T extends true> = T;
type Equal<A,B> = (<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Reflect<T> { return reflect.tuple<T>; }
typefn Rebuild<T> { return types.tuple<reflect.tuple<T>>; }
type P = Parameters<(first?:string, second?:number|undefined)=>void>;
type Parameter = Assert<Equal<Reflect<P>['elements'][0]['value'],string|undefined>>;
type ParameterMode = Assert<Equal<Reflect<P>['elements'][0]['mode'],'optional'>>;
type ParameterRoundtrip = Assert<Equal<Rebuild<P>,P>>;
type R = Parameters<(...args:[value?:string])=>void>;
type RestParameter = Assert<Equal<Reflect<R>['elements'][0]['value'],string>>;
type Prefix = Parameters<(first:number,...args:[value?:string,explicit?:boolean|undefined])=>void>;
type PrefixRest = Assert<Equal<Reflect<Prefix>['elements'][1]['value'],string>>;
type PrefixExplicit = Assert<Equal<Reflect<Prefix>['elements'][2]['value'],boolean|undefined>>;
type PrefixRoundtrip = Assert<Equal<Rebuild<Prefix>,Prefix>>;
type Constructor = ConstructorParameters<abstract new (value?:string)=>object>;
type ConstructorValue = Assert<Equal<Reflect<Constructor>['elements'][0]['value'],string|undefined>>;
type ConstructorRest = ConstructorParameters<new (...args:[value?:string])=>object>;
type ConstructorRestValue = Assert<Equal<Reflect<ConstructorRest>['elements'][0]['value'],string>>;
type Pending<T> = Reflect<Parameters<(value?:T)=>void>>;
type Replay = Assert<Equal<Pending<123>['elements'][0]['value'],123|undefined>>;
type ReplayNever = Assert<Equal<Pending<never>['elements'][0]['value'],undefined>>;
function binding([value=""]) {return value;}
type Binding = Parameters<typeof binding>[0];
type BindingValue = Assert<Equal<Reflect<Binding>['elements'][0]['value'],string|undefined>>;
type BindingRoundtrip = Assert<Equal<Rebuild<Binding>,Binding>>;
type Wrapped<T> = {[K in keyof T]:{value:T[K]}};
declare function unwrap<T extends readonly unknown[]>(value:Wrapped<T>):T;
declare const wrapped:readonly [first?:{value:string},second?:{value:number|undefined}];
const unwrapped = unwrap(wrapped);
type ReverseValue = Assert<Equal<Reflect<typeof unwrapped>['elements'][0]['value'],string>>;
type ReverseExplicit = Assert<Equal<Reflect<typeof unwrapped>['elements'][1]['value'],number|undefined>>;
type ReverseRoundtrip = Assert<Equal<Rebuild<typeof unwrapped>,typeof unwrapped>>;
