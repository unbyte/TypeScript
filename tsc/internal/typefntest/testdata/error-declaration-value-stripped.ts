export interface Open { first: string }
typefn Count<T extends object> { return tuple.length<reflect.properties<T>>; }
declare function read<T extends object>(): Count<T>;
/** @internal */
const hidden = read<Open>();
export const result = hidden;
