import {count, specialized, Box, nested, unrelated, widened, GenericBox, mappedCount, Open} from "./out/library";
declare module "./out/library" { interface Open { second: number } }
type Assert<T extends true> = T;
type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type A = Assert<Equal<typeof count, 2>>;
type B = Assert<Equal<typeof specialized, 2>>;
type C = Assert<Equal<Box["count"], 2>>;
type D = Assert<Equal<Box["specialized"], 2>>;
type E = Assert<Equal<typeof nested.count, 2>>;
type F = Assert<Equal<typeof unrelated, 1>>;
type G = Assert<Equal<typeof widened, number>>;
type H = Assert<Equal<GenericBox<Open>["value"], 2>>;
type I = Assert<Equal<typeof mappedCount, 2>>;
