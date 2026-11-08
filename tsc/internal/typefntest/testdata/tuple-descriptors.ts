export {};
type Assert<T extends true> = T;
type Equal<A,B> = (<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Reflect<T> { return reflect.tuple<T>; }
typefn Rebuild<T> { return types.tuple<reflect.tuple<T>>; }
type Source<T> = readonly [name:T, optional?:number, explicit?:string|undefined, ...tail:boolean[]];
type Descriptor = Reflect<Source<"x">>;
type Shape = Assert<Equal<Descriptor,{readonly:true;elements:[
 {value:"x";label:["name"];mode:"required"},
 {value:number;label:["optional"];mode:"optional"},
 {value:string|undefined;label:["explicit"];mode:"optional"},
 {value:boolean[];label:["tail"];mode:"rest"}
]}>>;
type Rebuilt = Assert<Equal<Rebuild<Source<"x">>,Source<"x">>>;
type Empty = Assert<Equal<Rebuild<readonly []>,readonly []>>;
type Unlabeled = Assert<Equal<Reflect<[string,number?]>['elements'][1],{value:number;label:[];mode:"optional"}>>;
type Pending<T> = Reflect<Source<T>>;
type Replay = Assert<Equal<Pending<123>['elements'][0]['value'],123>>;
typefn Copy<T> { return [...T]; }
type Copied = Assert<Equal<Reflect<Copy<[implicit?:string,explicit?:string|undefined]>>['elements'][0]['value'],string>>;
type CopiedExplicit = Assert<Equal<Reflect<Copy<[implicit?:string,explicit?:string|undefined]>>['elements'][1]['value'],string|undefined>>;
type Direct = {value:[optional?:string]}['value'];
type DirectValue = Assert<Equal<Reflect<Direct>['elements'][0]['value'],string>>;
typefn Build<T> {
 return types.tuple<{readonly:false;elements:[
  {value:T;label:["first"];mode:"required"},
  {value:number;label:["optional"];mode:"optional"},
  {value:string|undefined;label:["explicit"];mode:"optional"}
 ]}>;
}
type Built = Assert<Equal<Build<"x">,[first:"x",optional?:number,explicit?:string|undefined]>>;
type BuiltMetadata = Assert<Equal<Reflect<Build<"x">>['elements'][1]['value'],number>>;
type BuiltExplicit = Assert<Equal<Reflect<Build<"x">>['elements'][2]['value'],string|undefined>>;
type BuiltPending<T> = Build<T>;
type BuiltReplay = Assert<Equal<Reflect<BuiltPending<123>>['elements'][0]['value'],123>>;
typefn BuildOptional<T> {return types.tuple<{readonly:false;elements:[{value:T;label:[];mode:"optional"}]}>;}
type OptionalPending<T> = BuildOptional<T>;
type OptionalReplay = Assert<Equal<Reflect<OptionalPending<string>>['elements'][0]['value'],string>>;
type OptionalExplicitReplay = Assert<Equal<Reflect<OptionalPending<string|undefined>>['elements'][0]['value'],string|undefined>>;
type OptionalNever = Assert<Equal<Reflect<OptionalPending<never>>['elements'][0]['value'],never>>;
typefn Variadic<T> { return types.tuple<{readonly:true;elements:[{value:0;label:[];mode:"required"},{value:T;label:[];mode:"variadic"}]}>; }
type VariadicPending<T> = Variadic<T>;
type VariadicReplay = Assert<Equal<VariadicPending<[a:string,b?:number]>,readonly [0,a:string,b?:number]>>;
type VariadicMetadata = Assert<Equal<Reflect<VariadicPending<[a:string,b?:number]>>['elements'][2]['value'],number>>;
typefn Rest<T> {return types.tuple<{readonly:false;elements:[{value:0;label:[];mode:"required"},{value:T;label:["tail"];mode:"rest"}]}>;}
type RestArray = Assert<Equal<Rest<string[]>,[0,...tail:string[]]>>;
typefn AllRest<T> {return types.tuple<{readonly:true;elements:[{value:T;label:[];mode:"rest"}]}>;}
type AllArray = Assert<Equal<AllRest<number[]>,readonly number[]>>;
