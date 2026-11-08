// @typefn-abi: 1
export typefn Twice<N> {return N*2;}
export typefn Save<T> {return typefn<U> {return [T,U];};}
export type Saved=Save<"saved">;
