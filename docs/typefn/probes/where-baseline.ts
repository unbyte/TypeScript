// Native TypeScript probes, not an implementation of proposed `where` syntax.
export type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends
  (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;

type Upper<T extends string> = "accepted";
export type AnyArgument = Upper<any>;
export type NeverArgument = Upper<never>;
// @probe-error 2344
type InvalidArgument = Upper<number>;
// @probe-error 2344
type InvalidUnion = Upper<"text" | 1>;
// @probe-error 2344
type UnprovedGeneric<T> = Upper<T>;
type ProvenGeneric<T extends string> = Upper<T>;

type Distribute<T> = T extends string ? true : false;
type Whole<T> = [T] extends [string] ? true : false;
export type DistributedUnion = Distribute<"text" | 1>;
export type WholeUnion = Whole<"text" | 1>;
export type DistributedAny = Distribute<any>;
export type DistributedNever = Distribute<never>;
export type WholeNever = Whole<never>;
type ConditionalChecks = [
  Assert<Equal<DistributedUnion, boolean>>,
  Assert<Equal<WholeUnion, false>>,
  Assert<Equal<DistributedAny, boolean>>,
  Assert<Equal<DistributedNever, never>>,
  Assert<Equal<WholeNever, true>>
];

// Assignability through any is not a transitive proof system.
declare let aString: string;
declare let anything: any;
declare let aNumber: number;
anything = aString;
aNumber = anything;
// @probe-error 2322
aNumber = aString;

// Combining separately written bounds as intersections is not always conjunction.
type Merged<T extends any & number> = T;
export type MergedAcceptsString = Merged<string>;
type Numeric<T extends number> = T;
// @probe-error 2344
type NumericRejectsString = Numeric<string>;

declare function plain<T>(x: T): [T];
declare function bounded<T extends string>(x: T): [T];
export const inferredPlain = plain("hello");
export const inferredBounded = bounded("hello");
type InferenceChecks = [
  Assert<Equal<typeof inferredPlain, [string]>>,
  Assert<Equal<typeof inferredBounded, ["hello"]>>
];

// A bound does not say that a type parameter equals the bound.
function manufacture<T extends { value: string }>(): T {
  // @probe-error 2322
  return { value: "hello" };
}
declare function needsString<T extends string>(x: T): void;
function boundedCaller<T extends string>(x: T) { needsString(x); }
function unboundedCaller<T>(x: T) {
  // @probe-error 2345
  needsString(x);
}

// Native TS allows later-parameter references and structural F-bounds.
type Forward<K extends keyof T, T> = T[K];
type Linked<T extends { next?: T }> = T;
type ForwardCheck = Assert<Equal<Forward<"id", { id: 42 }>, 42>>;
type Node = { next?: Node };
type LinkedCheck = Linked<Node>;
// @probe-error 2313
type Immediate<T extends T> = T;

class Animal { name = ""; }
class Unrelated { name = ""; }
type AnimalBound<T extends Animal> = T;
type StructurallyAccepted = AnimalBound<Unrelated>;
class BrandedAnimal { private brand!: void; name = ""; }
type NominalBound<T extends BrandedAnimal> = T;
// @probe-error 2344 native=2741
type MissingPrivateIdentity = NominalBound<Unrelated>;
type ConstructorBound<T extends typeof Animal> = T;
type ConstructorAccepted = ConstructorBound<typeof Unrelated>;
// @probe-error 2344 native=2741
type InstanceIsNotConstructor = ConstructorBound<Animal>;

// Returning never does not itself report a bad call.
declare function conditionalReturn<T>(x: T): T extends string ? string : never;
export const silentlyNever = conditionalReturn(123);
type NeverCheck = Assert<Equal<typeof silentlyNever, never>>;

type Consumer<T> = (x: T) => void;
declare let consumesString: Consumer<string>;
declare let consumesLiteral: Consumer<"only">;
consumesLiteral = consumesString;
// @probe-error 2322
consumesString = consumesLiteral;
type Method<T> = { consume(x: T): void };
declare let stringMethod: Method<string>;
declare let literalMethod: Method<"only">;
stringMethod = literalMethod; // Native method bivariance, even under strict mode.
