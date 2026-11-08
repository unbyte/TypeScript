type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
typefn Same<A,B> { return A == B; }
typefn Assign<A,B> { return A is B; }
typefn Collect<U> { type result = never; for (type member in U) { result = result | [member]; } return result; }
typefn IsText<T> { match (T) { is {value: string} => { return true; } _ => { return false; } } }
typefn Built<T> { return Same<{value:T}, {value:string}>; }
interface Box<T> { value: T; }
interface Recursive { readonly next?: Recursive; value: string; }
interface Recursive2 { readonly next?: Recursive2; value: string; }
type Fn<T> = (value: T) => T;
type Generic<T> = <U>(value: U) => [T,U];
type Mapping<T> = {[K in keyof T]: T[K]};
type Late<T> = Same<Box<T>,Box<string>>;
type LateFn<T> = Same<Fn<T>,Fn<string>>;
type LateGeneric<T> = Same<Generic<T>,Generic<string>>;
type LateMapped<T> = Same<Mapping<{value:T}>,{value:string}>;
type LateAnonymous<T> = Same<{value:T},{value:string}>;
type LateUnion<T> = Collect<{kind:"a";value:T}|{kind:"b";value:number}>;
type LateBuilt<T> = Built<T>;
typefn Property<T> { match (T) { {value: item} => { return item; } _ => { return false; } } }
typefn Prototype<T> { match (T) { {toString: item} => { return true; } _ => { return false; } } }
type PropertyLater<T> = Property<Mapping<T>>;
typefn TupleHead<T> { match (T) { [head,..._] => { return head; } _ => { return false; } } }
type TupleLater<T extends unknown[]> = TupleHead<[...T]>;
function inferred<T>(value: T) {
 const object = {value};
 type Check = Same<typeof object, {value:string}>;
 let result!: Check;
 return result;
}
type Checks = [Assert<Equal<ReturnType<typeof inferred<string>>,true>>, Assert<Equal<ReturnType<typeof inferred<number>>,false>>,Assert<Equal<PropertyLater<{value:1}>,1>>, Assert<Equal<PropertyLater<{}>,false>>,
 Assert<Equal<TupleLater<[7,8]>,7>>, Assert<Equal<Prototype<{}>,false>>, Assert<Equal<Prototype<{toString:()=>string}>,true>>,
 Assert<Equal<Same<{a:1},{a:1}>,true>>, Assert<Equal<Same<{a:1},{a:2}>,false>>,
 Assert<Equal<Assign<{a:1;b:2},{a:number}>,true>>, Assert<Equal<Assign<{a:1},{a:number;b:2}>,false>>,
 Assert<Equal<Same<{a:1},never>,false>>, Assert<Equal<Same<Recursive,Recursive2>,true>>,
 Assert<Equal<Same<<T>(value:T)=>T,<U>(value:U)=>U>,true>>,
 Assert<Equal<Late<string>,true>>, Assert<Equal<Late<number>,false>>,
 Assert<Equal<LateFn<string>,true>>, Assert<Equal<LateFn<number>,false>>,
 Assert<Equal<LateGeneric<string>,true>>, Assert<Equal<LateGeneric<number>,false>>,
 Assert<Equal<LateMapped<string>,true>>, Assert<Equal<LateMapped<number>,false>>,
 Assert<Equal<LateAnonymous<string>,true>>, Assert<Equal<LateAnonymous<number>,false>>,
 Assert<Equal<LateBuilt<string>,true>>, Assert<Equal<LateBuilt<number>,false>>,
 Assert<Equal<LateUnion<1>,[{kind:"a";value:1}]|[{kind:"b";value:number}]>>,
 Assert<Equal<Collect<Recursive|{a:1}>,[Recursive]|[{a:1}]>>,
 Assert<Equal<IsText<{value:string}>,true>>, Assert<Equal<IsText<{value:number}>,false>>
];
