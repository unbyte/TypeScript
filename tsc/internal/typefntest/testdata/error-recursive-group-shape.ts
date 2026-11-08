export {};
typefn Bad<> {return types.recursiveGroup<2,typefn<R> {return R[0];}>;}
type Result=Bad<>;
