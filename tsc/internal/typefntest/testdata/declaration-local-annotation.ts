export interface Open { first:string }
typefn Count<T extends object> { return tuple.length<reflect.properties<T>>; }
declare function read<T extends object>():Count<T>;
declare function maybeRead<T extends object>():Count<T>|undefined;
export function local<T extends object>() {
  const value:Count<T> = read<T>();
  return value;
}
export function optional<T extends object>() {
  const value:Count<T>|undefined = maybeRead<T>();
  return value;
}
export function parameter<T extends object>(value:Count<T>) { return value; }
export class Box<T extends object> {
  method() { const value:Count<T> = read<T>(); return value; }
  get count() { const value:Count<T> = read<T>(); return value; }
}
export const count = local<Open>();
export const optionalCount = optional<Open>();
export const literal = 1;
