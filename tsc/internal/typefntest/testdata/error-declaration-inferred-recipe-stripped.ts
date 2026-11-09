export interface Open { first: string }
/** @internal */
typefn Count<T extends object> { return tuple.length<reflect.properties<T>>; }
declare function read(): Count<Open>;
export const count = read();
