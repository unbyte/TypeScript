typefn F<T> {return reflect.indexers<T>; }
type Result=F<{a:1}|{b:2}>;
