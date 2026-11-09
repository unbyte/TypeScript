import { specialized, result, direct, defaulted, defaultResult, Generic, sameValue, type Open } from "./out/library";
declare module "./out/library" { interface Open { second: number } }
type Assert<T extends true> = T;
type Equal<A,B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Check0 = Assert<Equal<ReturnType<typeof specialized>, "first" | "second">>;
type Check1 = Assert<Equal<typeof result, "first" | "second">>;
type Check2 = Assert<Equal<typeof direct, "first" | "second">>;
type Check3 = Assert<Equal<ReturnType<typeof defaulted>, "first" | "second">>;
type Check4 = Assert<Equal<typeof defaultResult, "first" | "second">>;
type Check5 = Assert<Equal<ReturnType<Generic<Open>["factory"]>, "first" | "second">>;
type Fixed = Assert<Equal<ReturnType<typeof sameValue>, "first">>;
