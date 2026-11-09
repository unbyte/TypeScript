export interface Open { first: string }
typefn Keys<T extends object> { return type(keyof T); }
typefn Wrap<T> { return { payload: T }; }
declare function identity<T>(): T;
declare function pair<T>(): [T, T];
declare function boxed<T>(): Wrap<T>;
declare function defaulted<T = Keys<Open>>(): T;
declare function dependent<A extends object, B = Keys<A>>(): B;
export const explicit = identity<Keys<Open>>();
export const paired = pair<Keys<Open>>();
export const fixed = boxed<"first">();
export const box = boxed<Keys<Open>>();
export const again = boxed<Keys<Open>>();
export const chained = dependent<Open>();
export class Generic<T extends object> { value = identity<Keys<T>>(); }
declare function inferred<T = Keys<Open>>(x: T): T;
export const fromValue = inferred("first" as const);
declare function chain<A extends object, B = Keys<A>, C = Wrap<B>>(): C;
export const deepDefault = chain<Open>();
