export {};
typefn Bad<> {return types.recursive<typefn<Self> {return tuple.slice<[1],Self>;}>;}
type Result=Bad<>;
