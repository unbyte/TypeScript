import type { External } from './external';
type Dead = { dead: true };
declare const unique: unique symbol;
declare class Nominal { private brand; }
typefn Helper<T> { return [T, T]; }
export typefn Branch<T> {
 if (T is string) { return Helper<T>; }
 return type({ dead: Dead; external: External; nominal: Nominal; [unique]: T });
}
export typefn Save<T> {
 type current = T;
 type saved = typefn<X> { return [current, X]; };
 current = number;
 return saved;
}
typefn WithError<E> { return typefn<A> { return { ok: true, value: A } | { ok: false, error: E }; }; }
export type Result = WithError<Error>;
export type Frozen = Save<string>;
export namespace API {
 typefn Private<T> { return [T]; }
 export typefn Wrap<T> { return Private<T>; }
}
export type Unique = typeof unique;
export type Native = Nominal;
