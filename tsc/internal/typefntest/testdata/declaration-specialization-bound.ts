export interface Open { first: string }
typefn Keys<T extends object> { return type(keyof T); }
declare function bounded<T extends "first">(): T;
const specialized = bounded<Keys<Open>>;
export const result = specialized();
