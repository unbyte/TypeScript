export interface Open { first: string }
typefn One<T extends object> {
  if (tuple.length<reflect.properties<T>> != 1) { fail<"Expected one property">; }
  return string;
}
declare function consume<T extends object>(value: One<T>): string;
export const result = consume<Open>("ok");
