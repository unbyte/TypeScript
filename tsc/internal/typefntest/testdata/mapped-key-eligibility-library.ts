typefn Validate<K> {
  if (K is string) {
    if (K == "bad") { fail<"Mapped key failure">; }
  }
  return K;
}
export typefn Key<K> { type checked = Validate<K>; return K; }
export type Remap<T> = { [K in keyof T as Key<K>]: T[K] };
export type Direct<K extends PropertyKey> = { [P in Key<K>]: P };
export typefn Read<T> { return reflect.properties<type(Remap<T>)>; }
export typefn ReadKeys<K> { return reflect.properties<type(Direct<K>)>; }
export declare const tag: unique symbol;
export type Input = { readonly good?: 1; 4: 2; [tag]: 3 };
// The branch makes the generic result opaque; the native intersection supplies
// the key-domain proof while retaining the deferred calculation.
export typefn ComputeKey<K> {
  if (K is string) { type checked = Validate<K>; return K; }
  return K;
}
export type ComputedRemap<T> = { [K in keyof T as ComputeKey<K> & PropertyKey]: T[K] };
export typefn ReadComputed<T> { return reflect.properties<type(ComputedRemap<T>)>; }
