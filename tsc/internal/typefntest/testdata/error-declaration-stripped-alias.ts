/** @internal */
typefn Save<T> { return typefn<U> { return [T, U]; }; }
export type Saved = Save<string>;
