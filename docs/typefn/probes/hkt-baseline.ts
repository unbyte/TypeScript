// Stock TypeScript observations, NOT tests of the proposed DSL.
// Run: node docs/typefn/probes/check-hkt-baseline.mjs
export {};

type Box<T> = { value: T };
type OtherBox<T> = { value: T };
declare const box: Box<string>;
declare const otherBox: OtherBox<string>;
const structuralForward: OtherBox<string> = box;
const structuralBackward: Box<string> = otherBox;

declare function unbox<T>(value: Box<T>): T;
export const observedNamedBox = unbox(box);
export const observedLiteralBox = unbox({ value: "hello" });

// Same structure, different inference evidence in the installed compiler.
type Empty<T> = {};
declare function inferEmpty<T>(value: Empty<T>): T;
declare const emptyNumber: Empty<number>;
export const observedNamedEmpty = inferEmpty(emptyNumber);
export const observedLiteralEmpty = inferEmpty({});

type Constant<T> = string;
declare const constantNumber: Constant<number>;
const sameOutput: Constant<boolean> = constantNumber;

type Branch<T> = T extends string ? number : boolean;
declare const branchNarrow: Branch<"hello">;
// "hello" extends unknown, but Branch<"hello"> does not extend Branch<unknown>.
// @ts-expect-error Non-covariant constructor: number is not boolean.
const branchWider: Branch<unknown> = branchNarrow;

type Poly = <T>(value: T) => Box<T>;
// @ts-expect-error A generic call signature is not a generic type alias.
type ApplyPoly = Poly<string>;
export type ObservedGenericReturn = ReturnType<Poly>;
declare function makeBox<T>(value: T): Box<T>;
export type ObservedSpecializedReturn = ReturnType<typeof makeBox<string>>;

declare function overload(value: string): number;
declare function overload(value: number): string;
export type ObservedOverloadReturn = ReturnType<typeof overload>;
export const observedDirectCall = overload("hello");

type StringBox<T extends string> = { text: T };
// @ts-expect-error A unary constructor need not accept every Type.
type WrongDomain = StringBox<number>;

interface ShiftedPromise<T> extends Promise<T[]> {}
declare const shifted: ShiftedPromise<string>;
const matchingOutput: PromiseLike<string[]> = shifted;
// @ts-expect-error Same arity does not imply the same input/output relationship.
const wrongOutput: PromiseLike<string> = shifted;

type RequiredBox<T> = { value: T };
// @ts-expect-error Missing type arguments do not mean a constructor value today.
type RequiredAlias = RequiredBox;
type DefaultBox<T = unknown> = { value: T };
export type ObservedDefault = DefaultBox["value"];

type NativeStrings<T> = T extends string ? T : never;
export type ObservedDistributed = NativeStrings<"hello" | 1>;
export type ObservedWholeUnion = ("hello" | 1) extends string ? "hello" | 1 : never;

// Structural typing still has native identity-sensitive exceptions.
declare class PrivateA { private identity: void; }
declare class PrivateB { private identity: void; }
declare const privateA: PrivateA;
// @ts-expect-error Distinct private declarations remain incompatible.
const privateB: PrivateB = privateA;
