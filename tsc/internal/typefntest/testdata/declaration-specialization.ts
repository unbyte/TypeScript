export interface Open { first: string }
typefn Keys<T extends object> { return type(keyof T); }
declare function identity<T>(): T;
declare function partial<A extends object, B = Keys<A>>(): B;
export const specialized = identity<Keys<Open>>;
export const result = specialized();
export const direct = (identity<Keys<Open>>)();
export const defaulted = partial<Open>;
export const defaultResult = defaulted();
export class Generic<T extends object> { factory = identity<Keys<T>>; }
export const sameValue = identity<"first">;
