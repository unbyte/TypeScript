export {};
type Assert<T extends true> = T;
type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
typefn Fn<P, R> { return types.function<P, R>; }
typefn Ctor<P, R, A> { return types.constructor<P, R, A>; }
type Params = [name: string, count?: number, ...flags: boolean[]];
declare const fn: Fn<Params, "result">;
const a: "result" = fn("a");
const b: "result" = fn("a", 1, true, false);
type Args = Assert<Equal<Parameters<typeof fn>, Params>>;
type Result = Assert<Equal<ReturnType<typeof fn>, "result">>;
type VoidArgs = Assert<Equal<Parameters<Fn<[], void>>, []>>;
declare const empty: Fn<[], 1>;
const e: 1 = empty();
type Symbolic<T> = Fn<[value: T], T>;
type Twice<T> = Symbolic<T>;
declare const symbolic: Twice<"value">;
const kept: "value" = symbolic("value");
type Tail<T extends unknown[]> = Fn<[first: string, ...tail: T], T>;
declare const tail: Tail<[number, boolean]>;
const tailResult: [number, boolean] = tail("x", 1, true);
type ReadonlyArgs = Fn<readonly [name: string], number>;
declare const readonlyFn: ReadonlyArgs;
const readonlyResult: number = readonlyFn("x");
declare const ctor: Ctor<[value: number], { value: number }, false>;
const built: { value: number } = new ctor(1);
type CtorArgs = Assert<Equal<ConstructorParameters<typeof ctor>, [value: number]>>;
type CtorResult = Assert<Equal<InstanceType<typeof ctor>, {value: number}>>;
// The native host permits construct signatures returning primitive Types.
type NativePrimitiveConstructor = new () => number;
type PrimitiveCtor = Ctor<[], number, false>;
type PrimitiveResult = Assert<Equal<InstanceType<PrimitiveCtor>, InstanceType<NativePrimitiveConstructor>>>;
typefn Observe<T> { if (T is NativeFn) { return true; } return false; }
type NativeFn = (value: string) => string;
type Observed<T> = Observe<Symbolic<T>>;
type Observation = Assert<Equal<Observed<string>, true>>;
type Rejection = Assert<Equal<Observed<number>, false>>;
