import value, {value as named, result, Box} from "./out/library";
type Assert<T extends true> = T;
type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type A = Assert<Equal<typeof named, {value: 1}>>;
type B = Assert<Equal<typeof value, {value: 1}>>;
type C = Assert<Equal<ReturnType<typeof result>, {value: 1}>>;
type D = Assert<Equal<Box["value"], {value: 1}>>;
