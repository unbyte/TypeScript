export {};
typefn Bad<> {return types.recursiveGroup<2,typefn<R> {return [{next:R[0]}];}>;}
type Result=Bad<>;
