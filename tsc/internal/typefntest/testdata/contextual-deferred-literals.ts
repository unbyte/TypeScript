type Assert<T extends true> = T;
type Equal<A,B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
typefn Keys<T> { return type(keyof T); }
declare function select<T extends object>(value:T, key:Keys<T>):T;
const selected = select({name:1}, "name");
type Inferred = Assert<Equal<typeof selected, {name:number}>>;
function forward<T extends object>(value:T, key:Keys<T>) {
  return select<T>(value, key);
}
