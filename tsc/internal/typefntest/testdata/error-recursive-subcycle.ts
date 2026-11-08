export {};
typefn Bad<> {return types.recursiveGroup<2,typefn<R> {return [{next:R[1]},R[1] | R[0]];}>;}
type Result=Bad<>;
