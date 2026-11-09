export interface Open { first:string }
export function outer() {
  typefn Hidden<T extends object> { return tuple.length<reflect.properties<T>>; }
  function read():Hidden<Open> { throw 0; }
  const value:Hidden<Open> = read();
  return value;
}
