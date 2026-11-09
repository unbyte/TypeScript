import { result, type Form, type Paths, type Deferred } from "./out/library";
import type { NativePaths } from "./out/reference";
type Equal<A,B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
type Fixed = Assert<Equal<Paths, "user" | "user.id">>;
type Native = Assert<Equal<Paths, NativePaths<Form>>>;
type Inferred = Assert<Equal<typeof result, Paths>>;
type Generic = Assert<Equal<Deferred<Form>, Paths>>;
