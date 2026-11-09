export typefn Keys<T> { return type(keyof T); }
export declare function read<T, P extends Keys<T>>(value:T, path:P): P;
export function forward<T, P extends Keys<T>>(value:T, path:P) { return read<T,P>(value,path); }
export function chained<T, P extends Keys<T>, Q extends P>(value:T,path:Q) { return read<T,Q>(value,path); }
export function assign<T, P extends Keys<T>>(path:P): Keys<T> { const bound:Keys<T> = path; return bound; }
export typefn Validated<T> {
  if (type(keyof T) is "bad") { fail<"Input bound failure">; }
  return type(keyof T);
}
export declare function checked<T, P extends Validated<T>>(value:T,path:P): P;
export function checkedForward<T, P extends Validated<T>>(value:T,path:P) { return checked<T,P>(value,path); }
