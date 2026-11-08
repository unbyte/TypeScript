export {};
type Assert<T extends true> = T;
type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
typefn Calls<T> { return reflect.calls<T>; }
typefn Constructs<T> { return reflect.constructs<T>; }
typefn Generic<T> { return reflect.isGeneric<T>; }
typefn Forward<T> { return types.overload<reflect.calls<T>, reflect.constructs<T>>; }
typefn ObjectForward<T> { return types.object<[], [], reflect.calls<T>, reflect.constructs<T>>; }
interface Source {
 (x: "first"): 1;
 (x: string): 2;
 new (x: number): { value: number };
}
type Project = Forward<Source>;
declare const project: Project;
const first: 1 = project("first");
const second: 2 = project("second");
const created: { value: number } = new project(1);
type Order = Assert<Equal<Parameters<Calls<Source>[0]>, [x: "first"]>>;
type Last = Assert<Equal<Parameters<Calls<Source>[1]>, [x: string]>>;
type NewResult = Assert<Equal<InstanceType<Constructs<Source>[0]>, { value: number }>>;
type Rebuilt = Assert<Equal<Calls<ObjectForward<Source>>['length'], 2>>;
type Id = <U>(value: U) => U;
type IsGeneric = Assert<Equal<Generic<Calls<Id>[0]>, true>>;
type IsMono = Assert<Equal<Generic<Calls<Source>[0]>, false>>;
declare const identity: Forward<Id>;
const literal: "kept" = identity("kept");
type Outer<T> = (value: T) => T;
type OuterGeneric<T> = <U>(value: U) => [T, U];
type Wrapped<T> = Forward<Outer<T>>;
type WrappedAgain<T> = ObjectForward<Wrapped<T>>;
type WrappedGeneric<T> = Forward<OuterGeneric<T>>;
declare const specialized: WrappedAgain<"outer">;
const specializedValue: "outer" = specialized("outer");
declare const specializedGeneric: WrappedGeneric<"outer">;
const specializedPair: ["outer", 123] = specializedGeneric<123>(123);
type NoCalls = Assert<Equal<Calls<{}>, []>>;
type NoConstructs = Assert<Equal<Constructs<Id>, []>>;
abstract class Abstract { abstract value: number; }
class Hidden { private constructor() {} }
class Protected { protected constructor() {} }
type HiddenPublicView = Assert<Equal<Constructs<typeof Hidden>, []>>;
type ProtectedPublicView = Assert<Equal<Constructs<typeof Protected>, []>>;
type AbstractInstance = Assert<Equal<InstanceType<Forward<typeof Abstract>>, Abstract>>;
type StringResult = () => string;
typefn IsStringReturn<T> { if (T is StringResult) { return true; } return false; }
type Observation<T> = IsStringReturn<Forward<() => T>>;
type Yes = Assert<Equal<Observation<string>, true>>;
type No = Assert<Equal<Observation<number>, false>>;
interface Animal { animal: true; }
interface Dog extends Animal { dog: true; }
type Method<T> = { method(x: T): void }['method'];
declare const dogMethod: Forward<Method<Dog>>;
const animalMethod: Forward<Method<Animal>> = dogMethod;

class Box<T> { value!: T; }
type GenericClass = Assert<Equal<Generic<Constructs<typeof Box>[0]>, true>>;
type EmptyConstruct = new () => unknown;
typefn IsConstructable<T> { if (T is EmptyConstruct) { return true; } return false; }
type GenericClassObservation = Assert<Equal<IsConstructable<Forward<typeof Box>>, true>>;
declare const box: Forward<typeof Box>;
const boxValue: string = new box<string>().value;
typefn NeedString<T> { if (T is string) { return T; } fail<"not a string">; }
type Latent = <T>(value: T) => NeedString<T>;
type LatentGeneric = Assert<Equal<Generic<Calls<Latent>[0]>, true>>;
declare const latent: Forward<Latent>;
const latentResult: "ok" = latent<"ok">("ok");
