export interface Open { first: string }
typefn Count<T extends object> { return tuple.length<reflect.properties<T>>; }
declare function read<T extends object>(): Count<T>;
function wrap<T extends object>() { return read<T>(); }
const arrow = <T extends object>() => read<T>();
export const result = wrap<Open>();
export const fromArrow = arrow<Open>();
export class Generic<T extends object> { value = wrap<T>(); }
typefn Callable<T extends object> { return types.function<[], Count<T>>; }
declare const generated: Callable<Open>;
export const fromGenerated = generated();
export const widened: number = wrap<Open>();
