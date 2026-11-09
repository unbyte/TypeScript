import { result, type Values, type Form, type Deferred } from "./out/library";
import type { NativePathValue } from "./out/reference";
import type { Path } from "./out/paths-native";
type Assert<T extends true> = T;
type Equal<A,B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Fixed = Assert<Equal<Values, {id:number}|number>>;
type Compared = Assert<Equal<Values, NativePathValue<Form, Path<Form>>>>;
type Inferred = Assert<Equal<typeof result, Values>>;
type Generic = Assert<Equal<Deferred<Form, "user.id">, number>>;
