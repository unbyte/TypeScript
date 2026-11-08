export typefn Compose<F<_>,G<_>> {return typefn<T> {return F<G<T>>;};}
export type AsyncArray=Compose<Promise,Array>;
