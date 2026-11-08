export {};
type Assert<T extends true> = T;
type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
typefn Rebuild<T> { return types.signature<reflect.signature<T>>; }
type Source<T> = (this:{context:T}, value:T, count?:number, ...tail:boolean[]) => T;
type Rebuilt<T> = Rebuild<Source<T>>;
type Sig = Rebuilt<"value">;
declare const fn: Sig;
const result: "value" = fn.call({context:"value"}, "value", 1, true);
type Args = Assert<Equal<Parameters<Sig>, [value:"value", count?:number, ...tail:boolean[]]>>;
type This = Assert<Equal<ThisParameterType<Sig>, {context:"value"}>>;
type Result = Assert<Equal<ReturnType<Sig>, "value">>;
declare const guard: Rebuild<(value: unknown) => value is string>;
declare let maybe: unknown;
if (guard(maybe)) { const narrowed: string = maybe; }
declare const assertion: Rebuild<(value: unknown) => asserts value is number>;
assertion(maybe);
const number: number = maybe;
declare const truthy: Rebuild<(value: unknown) => asserts value>;
declare let truth: string | undefined;
truthy(truth);
const string: string = truth;
type Receiver = {value:unknown; check: Rebuild<() => this is {value:string}>};
declare const receiver: Receiver;
if (receiver.check()) { const narrowed: string = receiver.value; }
type AbstractCtor = Rebuild<abstract new (value:number) => {value:number}>;
type AbstractArgs = Assert<Equal<ConstructorParameters<AbstractCtor>, [value:number]>>;
type AbstractResult = Assert<Equal<InstanceType<AbstractCtor>, {value:number}>>;
declare const ctor: Rebuild<new (value:number) => {value:number}>;
const instance: {value:number} = new ctor(1);
typefn Make<T> {
 return types.signature<{
  kind:"call"; parameters:[{name:"value";value:T;mode:"required"}];
  this:[]; returns:boolean; abstract:false;
  predicate:[{kind:"is";target:0;narrowed:[string]}]
 }>;
}
type Symbolic<T> = Make<T>;
declare const symbolic: Symbolic<unknown>;
declare let x: unknown;
if (symbolic(x)) { const narrowed: string = x; }
typefn Rest<T> {
 return types.signature<{kind:"call";parameters:[{name:"values";value:T;mode:"rest"}];this:[];returns:void;abstract:false;predicate:[]}>;
}
type PendingRest<T> = Rest<T>;
declare const rest: PendingRest<[value:string, count?:number]>;
rest("x"); rest("x", 1);
