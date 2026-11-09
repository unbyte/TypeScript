import { result, type Form, type Paths, type Deferred } from "./out/library";
import type { NativeArrayPaths } from "./out/reference";
type Equal<A,B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
declare module "./out/library" { interface Form { rows: { name:string }[] } }
type Fixed = Assert<Equal<Paths, "main" | "rows">>;
type Native = Assert<Equal<Paths, NativeArrayPaths<Form>>>;
type Inferred = Assert<Equal<typeof result, Paths>>;
type Generic = Assert<Equal<Deferred<Form>, Paths>>;
