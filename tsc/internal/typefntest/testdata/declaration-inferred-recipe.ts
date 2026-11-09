export interface Open { first: string }
typefn Count<T extends object> { return tuple.length<reflect.properties<T>>; }
declare function read(): Count<Open>;
declare function generic<T extends object>(): Count<T>;
export const count = read();
export const specialized = generic<Open>();
type Assert<T extends true> = T;
type Producer = Assert<typeof count extends 1 ? true : false>;
type GenericProducer = Assert<typeof specialized extends 1 ? true : false>;
export class Box { count = read(); readonly specialized = generic<Open>(); }
export const nested = { count: read() };
export const unrelated = 1;
export const widened: number = read();
export class GenericBox<T extends object> { value = generic<T>(); }
declare function mapped<T extends object>(): Count<{ [K in keyof T]: T[K] }>;
export const mappedCount = mapped<Open>();
