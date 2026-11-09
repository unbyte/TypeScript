import {explicit, paired, box, fixed, again, deepDefault, fromValue, chained, Generic, Open} from "./out/library";
declare module "./out/library" { interface Open { second: number } }
type Assert<T extends true> = T;
type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Expected = "first" | "second";
type A = Assert<Equal<typeof explicit, Expected>>;
type B = Assert<Equal<typeof paired, [Expected, Expected]>>;
type C = Assert<Equal<typeof box, {payload: Expected}>>;
type D = Assert<Equal<typeof fromValue, "first">>;
type E = Assert<Equal<typeof chained, Expected>>;
type F = Assert<Equal<Generic<Open>["value"], Expected>>;
type G = Assert<Equal<typeof fixed, {payload: "first"}>>;
type H = Assert<Equal<typeof again, {payload: Expected}>>;
type I = Assert<Equal<typeof deepDefault, {payload: Expected}>>;
