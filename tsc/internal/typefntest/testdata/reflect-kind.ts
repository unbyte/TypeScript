type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
typefn Kind<T> { return reflect.kind<T>; }
typefn Tests<T> { return [reflect.isAny<T>,reflect.isNever<T>,reflect.isUnknown<T>]; }
declare const unique: unique symbol;
interface Box<T> {value:T;}
type Late<T> = Kind<T>;
type LateUnion<T> = Kind<Box<T>|Box<string>>;
type Mapped<T> = {[K in keyof T]:T[K]};
type LateMap<T> = Kind<Mapped<T>>;
type Variadic<T extends unknown[]> = Kind<[...T]>;
type Shallow<T> = [Kind<{value:T}>,Kind<T[]>];
type Checks = [
 Assert<Equal<Kind<any>,"any">>, Assert<Equal<Kind<never>,"never">>, Assert<Equal<Kind<unknown>,"unknown">>,
 Assert<Equal<Kind<"s">,"stringLiteral">>, Assert<Equal<Kind<1>,"numberLiteral">>, Assert<Equal<Kind<1n>,"bigintLiteral">>, Assert<Equal<Kind<true>,"booleanLiteral">>,
 Assert<Equal<Kind<string>,"string">>, Assert<Equal<Kind<number>,"number">>, Assert<Equal<Kind<bigint>,"bigint">>, Assert<Equal<Kind<boolean>,"boolean">>,
 Assert<Equal<Kind<null>,"null">>, Assert<Equal<Kind<undefined>,"undefined">>, Assert<Equal<Kind<void>,"void">>, Assert<Equal<Kind<typeof unique>,"uniqueSymbol">>, Assert<Equal<Kind<symbol>,"symbol">>,
 Assert<Equal<Kind<readonly [x:1,y?:2]>,"tuple">>, Assert<Equal<Kind<readonly number[]>,"array">>, Assert<Equal<Kind<1|2>,"union">>,
 Assert<Equal<Kind<{a:1}&{b:2}>,"intersection">>, Assert<Equal<Kind<{}>,"object">>, Assert<Equal<Kind<object>,"object">>, Assert<Equal<Kind<()=>void>,"object">>,
 Assert<Equal<Late<"x">,"stringLiteral">>, Assert<Equal<LateUnion<string>,"object">>, Assert<Equal<LateUnion<number>,"union">>,
 Assert<Equal<LateMap<[1,2]>,"tuple">>, Assert<Equal<LateMap<{a:1}>,"object">>, Assert<Equal<Variadic<string[]>,"array">>, Assert<Equal<Variadic<[1]>,"tuple">>,
 Assert<Equal<Shallow<1>,["object","array"]>>, Assert<Equal<Tests<any>,[true,false,false]>>, Assert<Equal<Tests<never>,[false,true,false]>>, Assert<Equal<Tests<unknown>,[false,false,true]>>
];
