export interface Open { first:string }
typefn Count<T extends object> { return tuple.length<reflect.properties<T>>; }
declare function read():Count<Open>|undefined;
export function narrowed() {
  const value:Count<Open>|undefined = read();
  if (value === undefined) { throw 0; }
  return value;
}
