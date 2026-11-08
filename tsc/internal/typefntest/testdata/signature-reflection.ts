export {};
type Assert<T extends true> = T;
type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
typefn Reflect<T> { return reflect.signature<T>; }
type Source<T> = (this: {context: T}, name: T, count?: number, ...tail: boolean[]) => T;
type Wrapped<T> = Reflect<Source<T>>;
type R = Wrapped<"value">;
type Shape = Assert<Equal<R, {
 kind: "call";
 parameters: [{name: "name"; value: "value"; mode: "required"}, {name: "count"; value: number | undefined; mode: "optional"}, {name: "tail"; value: boolean[]; mode: "rest"}];
 this: [{context: "value"}]; returns: "value"; abstract: false; predicate: [];
}>>;
type Pred = Reflect<(value: unknown) => value is string>;
type Predicate = Assert<Equal<Pred['predicate'], [{kind:"is";target:0;narrowed:[string]}]>>;
type PredReturn = Assert<Equal<Pred['returns'], boolean>>;
type Assertion = Reflect<(value: unknown) => asserts value>;
type AssertPredicate = Assert<Equal<Assertion['predicate'], [{kind:"asserts";target:0;narrowed:[]}]>>;
type AssertReturn = Assert<Equal<Assertion['returns'], void>>;
type AssertType = Reflect<(value: unknown) => asserts value is number>;
type AssertNarrowed = Assert<Equal<AssertType['predicate'], [{kind:"asserts";target:0;narrowed:[number]}]>>;
type ThisPredicate = Reflect<(this: unknown) => this is string>;
type ThisTarget = Assert<Equal<ThisPredicate['predicate'], [{kind:"is";target:"this";narrowed:[string]}]>>;
type ThisAssertion = Reflect<() => asserts this>;
type ThisAssert = Assert<Equal<ThisAssertion['predicate'], [{kind:"asserts";target:"this";narrowed:[]}]>>;
type Ctor = Reflect<abstract new (name: string) => { name: string }>;
type Construct = Assert<Equal<Ctor['kind'], "construct">>;
type Abstract = Assert<Equal<Ctor['abstract'], true>>;
type ConstructorThis = Assert<Equal<Ctor['this'], []>>;
type VoidParameter = Assert<Equal<Reflect<(value: void) => void>['parameters'][0]['mode'], "required">>;
type RestTuple = [name: string, count?: number];
type Rest = Assert<Equal<Reflect<(...args: RestTuple) => void>['parameters'], [{name:"args";value:RestTuple;mode:"rest"}]>>;
function initialized(value = 1, other: string) { return value; }
type Initial = Assert<Equal<Reflect<typeof initialized>['parameters'][0]['mode'], "required">>;
function optional(value = 1) { return value; }
type Optional = Assert<Equal<Reflect<typeof optional>['parameters'][0]['mode'], "optional">>;
