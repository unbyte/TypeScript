import type { FieldArrayPaths } from "./paths";
import type { NativeArrayPaths } from "./reference";
type Assert<T extends true> = T;
type Equal<A,B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
class Registered { private brand!: void; rows!: {id:number}[]; }
class Unregistered { private brand!: void; rows!: {id:number}[]; }
class Derived extends Registered { extra!: {value:string}[]; }
declare module "./contracts" { interface OpaqueTypes { registered: Registered } }
type Model = {chosen:Registered[]; other:Unregistered[]; derived:Derived[]; direct:Registered};
type Compared = Assert<Equal<FieldArrayPaths<Model>, NativeArrayPaths<Model>>>;
type Fixed = Assert<Equal<FieldArrayPaths<Model>, "other" | `other.${number}.rows`>>;
type Deferred<T extends object> = FieldArrayPaths<{payload:T[]}>;
type GenericOpaque = Assert<Equal<Deferred<Registered>, never>>;
