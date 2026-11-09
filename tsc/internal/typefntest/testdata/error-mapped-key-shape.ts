import type { Key } from "./library";
typefn Bad<K> { type checked = Key<K>; return {value: K}; }
typefn Opaque<K> { if (K is string) { return K; } return never; }
type Invalid<T> = { [K in keyof T as Bad<K>]: T[K] };
type Unknown<T> = { [K in keyof T as Opaque<K>]: T[K] };
