import type { FieldPaths } from "./paths";
import type { NativePaths } from "./reference";
type Assert<T extends true> = T;
type Equal<A,B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
class Registered { private brand!: void; nested!: {value:number}; }
class Unregistered { private brand!: void; nested!: {value:number}; }
class Derived extends Registered { extra!: {value:string}; }
declare module "./contracts" { interface OpaqueTypes { registered: Registered } }
type Model = {chosen:Registered; other:Unregistered; derived:Derived};
type Compared = Assert<Equal<FieldPaths<Model>, NativePaths<Model>>>;
type Fixed = Assert<Equal<FieldPaths<Model>, "chosen" | "derived" | "other" | "other.nested" | "other.nested.value">>;
type Deferred<T extends object> = FieldPaths<{payload:T}>;
type GenericOpaque = Assert<Equal<Deferred<Registered>, "payload">>;
