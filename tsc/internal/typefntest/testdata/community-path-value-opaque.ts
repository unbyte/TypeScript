import type { FieldPathValue, FieldArrayPathValue } from "./values";
import type { NativePathValue } from "./reference";
type Assert<T extends true> = T;
type Equal<A,B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
class Registered { private brand!: void; rows!: {id:1}[]; }
class Unregistered { private brand!: void; rows!: {id:2}[]; }
class Derived extends Registered { extra!: {value:3}[]; }
declare module "./contracts" { interface OpaqueTypes { registered:Registered } }
type Model = {chosen:Registered; other:Unregistered; derived:Derived; list:Registered[]};
type Compared = Assert<Equal<FieldPathValue<Model,"chosen"|"other.rows">, NativePathValue<Model,"chosen"|"other.rows">>>;
type Identity = Assert<Equal<FieldPathValue<Model,"chosen">, Registered>>;
type Subclass = Assert<Equal<FieldPathValue<Model,"derived">, Derived>>;
type ArrayIdentity = Assert<Equal<FieldPathValue<Model,"list.0">, Registered>>;
type UnregisteredArray = Assert<Equal<FieldArrayPathValue<Model,"other.rows">, {id:2}[]>>;
