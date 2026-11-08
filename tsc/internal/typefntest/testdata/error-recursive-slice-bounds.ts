export {};
typefn Bad<> {return types.recursiveGroup<1,typefn<R> {return tuple.slice<R,0,2>;}>;}
type Result=Bad<>;
