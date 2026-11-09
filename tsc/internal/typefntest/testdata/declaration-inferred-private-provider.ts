export interface Open { first: string }
typefn Count<T extends object> { return tuple.length<reflect.properties<T>>; }
export declare function read<T extends object>(): Count<T>;
