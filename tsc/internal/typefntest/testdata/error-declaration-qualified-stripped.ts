export interface Open { first: string }
typefn Count<T extends object> { return tuple.length<reflect.properties<T>>; }
/** @internal */
namespace Hidden { export declare function read<T extends object>(): Count<T>; }
export const result = Hidden.read<Open>();
