export interface Open { first:string }
type NonNullable<T> = never;
typefn Count<T extends object> { return tuple.length<reflect.properties<T>>; }
declare function read<T extends object>():Count<T>|null|undefined;
export const direct = read<Open>()!;
export const wrapped = { value: (read<Open>())! };
export function local<T extends object>() {
  const value:Count<T>|null|undefined = read<T>();
  return value!;
}
export function parameter<T extends object>(value?:Count<T>) { return value!; }
export function twice<T extends object>() { return ((read<T>())!)!; }
export class Box<T extends object> {
  get value() { return read<T>()!; }
  method() { const value:Count<T>|null|undefined = read<T>(); return value!; }
}
typefn Maybe<T extends object> {
  type count = Count<T>;
  if (count == 1) { return null; }
  return count;
}
declare function readMaybe<T extends object>():Maybe<T>;
export const initiallyNever = readMaybe<Open>()!;
typefn Identity<T> { return T; }
declare function readAny():Identity<any>;
declare function readUnknown():Identity<unknown>;
declare function readNever():Identity<never>;
export const anyValue = readAny()!;
export const unknownValue = readUnknown()!;
export const neverValue = readNever()!;
