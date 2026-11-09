export interface Open { first: string }
typefn Callable<T extends object> {
  type count = tuple.length<reflect.properties<T>>;
  if (count != 1) { fail<"Expected one callable input property">; }
  return types.function<[], count>;
}
declare const generated: Callable<Open>;
export const result = generated();
