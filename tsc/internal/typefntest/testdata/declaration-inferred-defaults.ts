export interface Open { first: string }
typefn Keys<T extends object> { return type(keyof T); }
declare function select<T = Keys<Open>>(value?: T): T;
declare const literal: "first";
export const defaults = select();
export const inferred = select(literal);
export const again = select();
declare function overload<T = Keys<Open>>(kind: "keys", value?: T): T;
declare function overload<T = "fixed">(kind: "fixed"): T;
export const chosen = overload("keys");
export const other = overload("fixed");
