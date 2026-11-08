export {};
typefn Bad<> {return types.recursiveGroup<1000,typefn<R> {return R;}>;}
type Result=Bad<>;
