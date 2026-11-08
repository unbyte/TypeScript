export {};
type Assert<T extends true> = T;
type Equal<A,B> = (<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Reflect<T> { return reflect.tuple<T>; }
typefn Rebuild<T> { return types.tuple<reflect.tuple<T>>; }
type Input = [a:string,b?:number,c?:boolean|undefined];
type P = Reflect<Partial<Input>>;
type RequiredToOptional = Assert<Equal<P['elements'][0]['value'],string>>;
type Implicit = Assert<Equal<P['elements'][1]['value'],number>>;
type Explicit = Assert<Equal<P['elements'][2]['value'],boolean|undefined>>;
type Labels = Assert<Equal<P['elements'][1]['label'],['b']>>;
type Mode = Assert<Equal<P['elements'][0]['mode'],'optional'>>;
type Roundtrip = Assert<Equal<Rebuild<Partial<Input>>,Partial<Input>>>;
type ReadonlyImplicit = Assert<Equal<Reflect<Readonly<Input>>['elements'][1]['value'],number>>;
type ReadonlyExplicit = Assert<Equal<Reflect<Readonly<Input>>['elements'][2]['value'],boolean|undefined>>;
type ReadonlyFlag = Assert<Equal<Reflect<Readonly<Input>>['readonly'],true>>;
type RequiredMode = Assert<Equal<Reflect<Required<Input>>['elements'][1]['mode'],'required'>>;
type RequiredValue = Assert<Equal<Reflect<Required<Input>>['elements'][2]['value'],Required<Input>[2]>>;
type Wrap<T> = {[K in keyof T]?: {wrapped:T[K]}};
type Wrapped = Assert<Equal<Reflect<Wrap<Input>>['elements'][1]['value'],{wrapped:number|undefined}>>;
type Constant<T> = {[K in keyof T]?: string|undefined};
type ExplicitTemplate = Assert<Equal<Reflect<Constant<Input>>['elements'][0]['value'],string|undefined>>;
type Pending<T> = Reflect<Readonly<Partial<[value:T,optional?:T]>>>;
type Replay = Assert<Equal<Pending<123>['elements'][0]['value'],123>>;
type ReplayOptional = Assert<Equal<Pending<123>['elements'][1]['value'],123>>;
type ReplayExplicit = Assert<Equal<Pending<123|undefined>['elements'][1]['value'],123|undefined>>;
type ReplayNever = Assert<Equal<Pending<never>['elements'][1]['value'],never>>;
type Nested<T> = Readonly<Partial<T>>;
type NestedIdentity = Assert<Equal<Reflect<Nested<Input>>['elements'][1]['value'],number>>;
type Empty = Assert<Equal<Rebuild<Partial<[]>>,[]>>;

// Native inference and literal tuple copies retain the originating declaration.
type Tail<T extends unknown[]> = T extends [unknown,...infer R] ? R : never;
type TailValue = Assert<Equal<Reflect<Tail<[0,a?:string,b?:number|undefined]>>['elements'][0]['value'],string>>;
type TailExplicit = Assert<Equal<Reflect<Tail<[0,a?:string,b?:number|undefined]>>['elements'][1]['value'],number|undefined>>;
declare const optional: readonly [a?:string,b?:number|undefined];
const copied = [...optional] as const;
type LiteralCopy = Assert<Equal<Reflect<typeof copied>['elements'][0]['value'],string>>;
type LiteralExplicit = Assert<Equal<Reflect<typeof copied>['elements'][1]['value'],number|undefined>>;
declare function collect<T extends unknown[]>(...args:T):T;
const collected = collect(...optional);
// Native call-spread inference makes these positional arguments required.
type MutableCopy = Assert<Equal<Reflect<typeof collected>['elements'][0]['value'],string|undefined>>;
type MutableMode = Assert<Equal<Reflect<typeof collected>['elements'][0]['mode'],'required'>>;
type MutableExplicit = Assert<Equal<Reflect<typeof collected>['elements'][1]['value'],number|undefined>>;
