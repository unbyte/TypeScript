declare const key: unique symbol;
typefn F<K> {return {[K]:1,[K]:2};}
type Result=F<typeof key>;
