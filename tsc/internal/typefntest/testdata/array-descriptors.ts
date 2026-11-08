type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
typefn Build<T,R> { return types.array<T,R>; }
typefn Default<T> { return types.array<T>; }
typefn Describe<T> { return reflect.array<T>; }
typefn Rebuild<T> { type description=reflect.array<T>; return types.array<description["element"],description["readonly"]>; }
typefn Length<T> { return types.array<T>["length"]; }
typefn ArrayElement<T> { return reflect.array<Array<T>>["element"]; }
type LengthProof<T> = Assert<Equal<Length<T>,number>>;
type ElementProof<T> = Assert<Equal<ArrayElement<T>,T>>;
type Later<R> = Build<string,R>;
type LaterDescription<T> = Describe<T>;
type Checks = [
 Assert<Equal<Default<1>,1[]>>, Assert<Equal<Build<string,true>,readonly string[]>>,
 Assert<Equal<Build<never,false>,never[]>>, Assert<Equal<Build<any,true>,readonly any[]>>,
 Assert<Equal<Describe<readonly [1,2][]>,{element:[1,2];readonly:true}>>,
 Assert<Equal<Rebuild<readonly (1|2)[]>,readonly (1|2)[]>>,
 Assert<Equal<Later<true>,readonly string[]>>, Assert<Equal<Later<false>,string[]>>,
 Assert<Equal<LaterDescription<number[]>,{element:number;readonly:false}>>
];

