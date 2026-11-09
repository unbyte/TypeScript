export interface Open { first: string }
typefn Keys<T extends object> { return type(keyof T); }
declare function bounded<T extends "first">(): T;
export const result = bounded<Keys<Open>>();
