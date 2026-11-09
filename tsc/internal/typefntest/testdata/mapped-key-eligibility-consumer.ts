import { tag } from "./out/library";
import type { Remap, ComputedRemap, Direct, Input } from "./out/library";
type Assert<T extends true> = T;
type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Mapped = Assert<Equal<Remap<Input>, Input>>;
type StringKey = Assert<Equal<Direct<"good">, {good: "good"}>>;
type NumberKey = Assert<Equal<Direct<4>, {4: 4}>>;
type SymbolKey = Assert<Equal<Direct<typeof tag>, {[tag]: typeof tag}>>;
type Computed = Assert<Equal<ComputedRemap<Input>, Input>>;
