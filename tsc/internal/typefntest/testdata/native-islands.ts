export {};
type Assert<T extends true> = T;
type Equal<A,B> = (<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2) ? true:false;
typefn ArrayOf<T> { return type(T[]); }
type ArrayCheck = Assert<Equal<ArrayOf<string>, string[]>>;
typefn Snap<T> {
 type local = T;
 type saved = type({value:local});
 local = number;
 return [saved, type({value:local})];
}
type Snapshot = Assert<Equal<Snap<string>, [{value:string},{value:number}]>>;
type Pending<T> = Snap<T>;
type Replay = Assert<Equal<Pending<"literal">, [{value:"literal"},{value:number}]>>;
typefn Map<T> { return type({readonly [K in keyof T]?: T[K]}); }
type Mapped = Assert<Equal<Map<{a:1;b:2}>, {readonly a?:1;readonly b?:2}>>;
typefn Unwrap<T> { return type(T extends Promise<infer R> ? R : never); }
type Inferred = Assert<Equal<Unwrap<Promise<1>>, 1>>;
typefn Shadow<T> { return type(<T>(value:T)=>T); }
declare const identity: Shadow<"outer">;
const literal: 123 = identity(123);
typefn Generic<T> { return type(<U>(value:U)=>[T,U]); }
type GenericAlias<T> = Generic<T>;
declare const generic: GenericAlias<"outer">;
const pair: ["outer",123] = generic<123>(123);
typefn GenericConditional<T> { return type(<U>()=> U extends string ? T : never); }
declare const conditional: GenericConditional<"matched">;
const conditionalResult: "matched" = conditional<string>();
const neverResult: never = conditional<number>();
declare const runtime: {readonly value:123};
typefn Query<> { return type(typeof runtime); }
type Queried = Assert<Equal<Query<>, typeof runtime>>;
class Nominal { private brand!: void; value!: number; }
typefn Preserve<T> { return type({value:T}); }
type Nominality = Assert<Equal<Preserve<Nominal>['value'], Nominal>>;
typefn Pattern<T> { match(T) { type({value:string}) => {return true;} _=>{return false;} } }
type Match = Assert<Equal<Pattern<{value:string}>,true>>;
type NoMatch = Assert<Equal<Pattern<{value:number}>,false>>;
typefn Independent<T> { return type(string); }
type Stable<T> = Assert<Equal<Independent<T>, string>>;

typefn BeforeShadow<> { type first = type(<later>(value:later)=>later); type later = 1; return first; }
declare const beforeShadow: BeforeShadow<>;
const shadowedValue: "ok" = beforeShadow("ok");
