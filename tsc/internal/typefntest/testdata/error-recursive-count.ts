export {};
typefn Bad<> {return types.recursiveGroup<0,typefn<R> {return [];}>;}
type Result=Bad<>;
