import {defaults, inferred, again, chosen, other} from "./out/library";
type Assert<T extends true> = T;
type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Expected = "first";
type A = Assert<Equal<typeof defaults, Expected>>;
type B = Assert<Equal<typeof inferred, "first">>;
type C = Assert<Equal<typeof again, Expected>>;
type D = Assert<Equal<typeof chosen, Expected>>;
type E = Assert<Equal<typeof other, "fixed">>;
